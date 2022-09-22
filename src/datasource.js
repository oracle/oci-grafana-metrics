/*
 ** Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
 ** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
 */
import _ from "lodash";
import {
  aggregations,
  dimensionKeysQueryRegex,
  namespacesQueryRegex,
  resourcegroupsQueryRegex,
  metricsQueryRegex,
  regionsQueryRegex,
  tenancyconfigsQueryRegex,
  compartmentsQueryRegex,
  dimensionValuesQueryRegex,
  removeQuotes,
  AUTO,
} from "./constants";
import retryOrThrow from "./util/retry";
import { SELECT_PLACEHOLDERS } from "./query_ctrl";
import { resolveAutoWinRes } from "./util/utilFunctions";
import { toDataQueryResponse } from "@grafana/runtime";

const DEFAULT_RESOURCE_GROUP = "NoResourceGroup";

export default class OCIDatasource {
  constructor(instanceSettings, $q, backendSrv, templateSrv, timeSrv) {
    this.type = instanceSettings.type;
    this.url = instanceSettings.url;
    this.name = instanceSettings.name;
    this.id = instanceSettings.id;
    this.tenancyOCID = instanceSettings.jsonData.tenancyOCID;
    this.defaultRegion = instanceSettings.jsonData.defaultRegion;
    this.environment = instanceSettings.jsonData.environment;
    this.q = $q;
    this.backendSrv = backendSrv;
    this.templateSrv = templateSrv;
    this.timeSrv = timeSrv;

    this.compartmentsCache = [];
    this.regionsCache = [];
    
    this.getTenancyConfig();

    this.getRegions();
    this.getCompartments();
  }

  /**
   * Each Grafana Data source should contain the following functions:
   *  - query(options) //used by panels to get data
   *  - testDatasource() //used by data source configuration page to make sure the connection is working
   *  - annotationQuery(options) // used by dashboards to get annotations
   *  - metricFindQuery(options) // used by query editor to get metric suggestions.
   * More information: https://grafana.com/docs/plugins/developing/datasources/
   */

  /**
   * Required method
   * Used by panels to get data
   */
  async query(options) {
    var query = await this.buildQueryParameters(options);
    if (query.targets.length <= 0) {
      return this.q.when({ data: [] });
    }

    return this.doRequest(query).then((result) => {
      var res = [];
      _.forEach(result.data, (r) => {
        const name = r.fields[1].config.displayNameFromDS;
        const timeArr = r.fields[0].values.toArray();
        const values = r.fields[1].values.toArray();
        const points = timeArr.map((t, i) => [values[i], t]);
        res.push({ target: name, datapoints: points });
      });

      result.data = res;
      return result;
    });
  }

  /**
   * Required method
   * Used by data source configuration page to make sure the connection is working
   */
  testDatasource() {
    return this.doRequest({
      targets: [
        {
          queryType: "test",
          region: this.defaultRegion,
          tenancyOCID: this.tenancyOCID,
          compartment: "",
          environment: this.environment,
          datasourceId: this.id,
        },
      ],
      range: this.timeSrv.timeRange(),
    })
      .then((response) => {
        if (response.status === 200) {
          return {
            status: "success",
            message: "Data source is working",
            title: "Success",
          };
        }
      })
      .catch(() => {
        return {
          status: "error",
          message: "Data source is not working",
          title: "Failure",
        };
      });
  }

  /**
   * Required method
   * Used by query editor to get metric suggestions
   */
  async metricFindQuery(target) {
    if (typeof target === "string") {
      // used in template editor for creating variables
      return this.templateMetricQuery(target);
    }
    const region =
      target.region === SELECT_PLACEHOLDERS.REGION
        ? ""
        : this.getVariableValue(target.region);
    const compartment =
      target.compartment === SELECT_PLACEHOLDERS.COMPARTMENT
        ? ""
        : this.getVariableValue(target.compartment);
    const namespace =
      target.namespace === SELECT_PLACEHOLDERS.NAMESPACE
        ? ""
        : this.getVariableValue(target.namespace);
    const resourcegroup =
      target.resourcegroup === SELECT_PLACEHOLDERS.RESOURCEGROUP
        ? DEFAULT_RESOURCE_GROUP
        : this.getVariableValue(target.resourcegroup);

    if (_.isEmpty(compartment) || _.isEmpty(namespace)) {
      return this.q.when([]);
    }

    const compartmentId = await this.getCompartmentId(compartment);
    return this.doRequest({
      targets: [
        {
          environment: this.environment,
          datasourceId: this.id,
          tenancyOCID: this.tenancyOCID,
          queryType: "search",
          region: _.isEmpty(region) ? this.defaultRegion : region,
          compartment: compartmentId,
          namespace: namespace,
          resourcegroup: resourcegroup,
          tenancyconfig: tenancyconfig,
        },
      ],
      range: this.timeSrv.timeRange(),
    }).then((res) => {
      return this.mapToTextValue(res, "search");
    });
  }

  /**
   * Build and validate query parameters.
   */
  async buildQueryParameters(options) {
    let queries = options.targets
      .filter((t) => !t.hide)
      .filter(
        (t) =>
          !_.isEmpty(
            this.getVariableValue(t.compartment, options.scopedVars)
          ) && t.compartment !== SELECT_PLACEHOLDERS.COMPARTMENT
      )
      .filter(
        (t) =>
          !_.isEmpty(this.getVariableValue(t.namespace, options.scopedVars)) &&
          t.namespace !== SELECT_PLACEHOLDERS.NAMESPACE
      )
      .filter(
        (t) =>
          !_.isEmpty(this.getVariableValue(t.resourcegroup, options.scopedVars))
      )
      .filter(
        (t) =>
          (!_.isEmpty(this.getVariableValue(t.metric, options.scopedVars)) &&
            t.metric !== SELECT_PLACEHOLDERS.METRIC) ||
          !_.isEmpty(this.getVariableValue(t.target))
      );

    queries.forEach((t) => {
      t.dimensions = (t.dimensions || [])
        .filter(
          (dim) =>
            !_.isEmpty(dim.key) && dim.key !== SELECT_PLACEHOLDERS.DIMENSION_KEY
        )
        .filter(
          (dim) =>
            !_.isEmpty(dim.value) &&
            dim.value !== SELECT_PLACEHOLDERS.DIMENSION_VALUE
        );

      t.resourcegroup =
        t.resourcegroup === SELECT_PLACEHOLDERS.RESOURCEGROUP
          ? DEFAULT_RESOURCE_GROUP
          : t.resourcegroup;
    });

    // we support multiselect for dimension values, so we need to parse 1 query into multiple queries
    queries = this.splitMultiValueDimensionsIntoQuieries(queries, options);

    const results = [];
    for (let t of queries) {
      const region =
        t.region === SELECT_PLACEHOLDERS.REGION
          ? ""
          : this.getVariableValue(t.region, options.scopedVars);
      let query = this.getVariableValue(t.target, options.scopedVars);
      const numberOfDaysDiff = this.timeSrv
        .timeRange()
        .to.diff(this.timeSrv.timeRange().from, "days");
      // The following replaces 'auto' in window portion of the query and replaces it with an appropriate value.
      // If there is a functionality to access the window variable instead of matching [auto] in the query, it will be
      // better
      if (query)
        query = query.replace(
          "[auto]",
          `[${resolveAutoWinRes(AUTO, "", numberOfDaysDiff).window}]`
        );
      let resolution = this.getVariableValue(t.resolution, options.scopedVars);
      let window =
        t.window === SELECT_PLACEHOLDERS.WINDOW
          ? ""
          : this.getVariableValue(t.window, options.scopedVars);
      // p.s : timeSrv.timeRange() results in a moment object
      const resolvedWinResolObj = resolveAutoWinRes(
        window,
        resolution,
        numberOfDaysDiff
      );
      window = resolvedWinResolObj.window;
      resolution = resolvedWinResolObj.resolution;
      if (_.isEmpty(query)) {
        // construct query
        const dimensions = (t.dimensions || []).reduce((result, dim) => {
          const d = `${this.getVariableValue(dim.key, options.scopedVars)} ${
            dim.operator
          } "${this.getVariableValue(dim.value, options.scopedVars)}"`;
          if (result.indexOf(d) < 0) {
            result.push(d);
          }
          return result;
        }, []);
        const dimension = _.isEmpty(dimensions)
          ? ""
          : `{${dimensions.join(",")}}`;
        query = `${this.getVariableValue(
          t.metric,
          options.scopedVars
        )}[${window}]${dimension}.${t.aggregation}`;
      }

      const compartmentId = await this.getCompartmentId(
        this.getVariableValue(t.compartment, options.scopedVars)
      );
      const result = {
        resolution,
        environment: this.environment,
        datasourceId: this.id,
        tenancyOCID: this.tenancyOCID,
        queryType: "query",
        refId: t.refId,
        hide: t.hide,
        type: t.type || "timeserie",
        region: _.isEmpty(region) ? this.defaultRegion : region,
        compartment: compartmentId,
        namespace: this.getVariableValue(t.namespace, options.scopedVars),
        resourcegroup: this.getVariableValue(
          t.resourcegroup,
          options.scopedVars
        ),
        query: query,
      };
      results.push(result);
    }

    options.targets = results;

    return options;
  }

  /**
   * Splits queries with multi valued dimensions into several quiries.
   * Example:
   * "DeliverySucceedEvents[1m]{resourceDisplayName = ["ResouceName_1","ResouceName_1"], eventType = ["Create","Delete"]}.mean()" ->
   *  [
   *    "DeliverySucceedEvents[1m]{resourceDisplayName = "ResouceName_1", eventType = "Create"}.mean()",
   *    "DeliverySucceedEvents[1m]{resourceDisplayName = "ResouceName_2", eventType = "Create"}.mean()",
   *    "DeliverySucceedEvents[1m]{resourceDisplayName = "ResouceName_1", eventType = "Delete"}.mean()",
   *    "DeliverySucceedEvents[1m]{resourceDisplayName = "ResouceName_2", eventType = "Delete"}.mean()",
   *  ]
   */
  splitMultiValueDimensionsIntoQuieries(queries, options) {
    return queries.reduce((data, t) => {
      if (_.isEmpty(t.dimensions) || !_.isEmpty(t.target)) {
        // nothing to split or dimensions won't be used, query is set manually
        return data.concat(t);
      }

      // create a map key : [values] for multiple values
      const multipleValueDims = t.dimensions.reduce((data, dim) => {
        const key = dim.key;
        const value = this.getVariableValue(dim.value, options.scopedVars);
        if (value.startsWith("{") && value.endsWith("}")) {
          const values = value.slice(1, value.length - 1).split(",") || [];
          data[key] = (data[key] || []).concat(values);
        }
        return data;
      }, {});

      if (_.isEmpty(Object.keys(multipleValueDims))) {
        // no multiple values used, only single values
        return data.concat(t);
      }

      const splitDimensions = (dims, multiDims) => {
        let prev = [];
        let next = [];

        const firstDimKey = dims[0].key;
        const firstDimValues = multiDims[firstDimKey] || [dims[0].value];
        for (let v of firstDimValues) {
          const newDim = _.cloneDeep(dims[0]);
          newDim.value = v;
          prev.push([newDim]);
        }

        for (let i = 1; i < dims.length; i++) {
          const values = multiDims[dims[i].key] || [dims[i].value];
          for (let v of values) {
            for (let j = 0; j < prev.length; j++) {
              if (next.length >= 20) {
                // this algorithm of collecting multi valued dimensions is computantionally VERY expensive
                // set the upper limit for quiries number
                return next;
              }
              const newDim = _.cloneDeep(dims[i]);
              newDim.value = v;
              next.push(prev[j].concat(newDim));
            }
          }
          prev = next;
          next = [];
        }

        return prev;
      };

      const newDimsArray = splitDimensions(t.dimensions, multipleValueDims);

      const newQueries = [];
      for (let i = 0; i < newDimsArray.length; i++) {
        const dims = newDimsArray[i];
        const newQuery = _.cloneDeep(t);
        newQuery.dimensions = dims;
        if (i !== 0) {
          newQuery.refId = `${newQuery.refId}${i}`;
        }
        newQueries.push(newQuery);
      }
      return data.concat(newQueries);
    }, []);
  }

  // **************************** Template variable helpers ****************************

  /**
   * Matches the regex from creating template variables and returns options for the corresponding variable.
   * Example:
   * template variable with the query "regions()" will be matched with the regionsQueryRegex and list of available regions will be returned.
   */
  templateMetricQuery(varString) {
    let regionQuery = varString.match(regionsQueryRegex);
    if (regionQuery) {
      return this.getRegions().catch((err) => {
        throw new Error("Unable to get regions: " + err);
      });
    }

    let tenancyconfigQuery = varString.match(tenancyconfigsQueryRegex);
    if (tenancyconfigQuery) {
      return this.getTenancyConfig().catch((err) => {
        throw new Error("Unable to get tenancyconfigs: " + err);
      });
    }    

    let compartmentQuery = varString.match(compartmentsQueryRegex);
    if (compartmentQuery) {
      return this.getCompartments()
        .then((compartments) => {
          return compartments.map((c) => ({ text: c.text, value: c.text }));
        })
        .catch((err) => {
          throw new Error("Unable to get compartments: " + err);
        });
    }

    let namespaceQuery = varString.match(namespacesQueryRegex);
    if (namespaceQuery) {
      let target = {
        region: removeQuotes(this.getVariableValue(namespaceQuery[1])),
        compartment: removeQuotes(this.getVariableValue(namespaceQuery[2])),
      };
      return this.getNamespaces(target).catch((err) => {
        throw new Error("Unable to get namespaces: " + err);
      });
    }

    let resourcegroupQuery = varString.match(resourcegroupsQueryRegex);
    if (resourcegroupQuery) {
      let target = {
        region: removeQuotes(this.getVariableValue(resourcegroupQuery[1])),
        compartment: removeQuotes(this.getVariableValue(resourcegroupQuery[2])),
        namespace: removeQuotes(this.getVariableValue(resourcegroupQuery[3])),
      };
      return this.getResourceGroups(target).catch((err) => {
        throw new Error("Unable to get resourcegroups: " + err);
      });
    }

    let metricQuery = varString.match(metricsQueryRegex);
    if (metricQuery) {
      let target = {
        region: removeQuotes(this.getVariableValue(metricQuery[1])),
        compartment: removeQuotes(this.getVariableValue(metricQuery[2])),
        namespace: removeQuotes(this.getVariableValue(metricQuery[3])),
        resourcegroup: removeQuotes(this.getVariableValue(metricQuery[4])),
      };
      return this.metricFindQuery(target).catch((err) => {
        throw new Error("Unable to get metrics: " + err);
      });
    }

    let dimensionsQuery = varString.match(dimensionKeysQueryRegex);
    if (dimensionsQuery) {
      let target = {
        region: removeQuotes(this.getVariableValue(dimensionsQuery[1])),
        compartment: removeQuotes(this.getVariableValue(dimensionsQuery[2])),
        namespace: removeQuotes(this.getVariableValue(dimensionsQuery[3])),
        metric: removeQuotes(this.getVariableValue(dimensionsQuery[4])),
        resourcegroup: removeQuotes(this.getVariableValue(dimensionsQuery[5])),
      };
      return this.getDimensionKeys(target).catch((err) => {
        throw new Error("Unable to get dimensions: " + err);
      });
    }

    let dimensionOptionsQuery = varString.match(dimensionValuesQueryRegex);
    if (dimensionOptionsQuery) {
      let target = {
        region: removeQuotes(this.getVariableValue(dimensionOptionsQuery[1])),
        compartment: removeQuotes(
          this.getVariableValue(dimensionOptionsQuery[2])
        ),
        namespace: removeQuotes(
          this.getVariableValue(dimensionOptionsQuery[3])
        ),
        metric: removeQuotes(this.getVariableValue(dimensionOptionsQuery[4])),
        resourcegroup: removeQuotes(
          this.getVariableValue(dimensionOptionsQuery[6])
        ),
      };
      const dimensionKey = removeQuotes(
        this.getVariableValue(dimensionOptionsQuery[5])
      );
      return this.getDimensionValues(target, dimensionKey).catch((err) => {
        throw new Error("Unable to get dimension options: " + err);
      });
    }

    throw new Error("Unable to parse templating string");
  }

  getRegions() {
    if (this.regionsCache && this.regionsCache.length > 0) {
      return this.q.when(this.regionsCache);
    }

    return this.doRequest({
      targets: [
        {
          environment: this.environment,
          datasourceId: this.id,
          tenancyOCID: this.tenancyOCID,
          queryType: "regions",
        },
      ],
      range: this.timeSrv.timeRange(),
    }).then((items) => {
      this.regionsCache = this.mapToTextValue(items, "regions");
      return this.regionsCache;
    });
  }

  getTenancyConfig() {
    if (this.tenancyconfigCache && this.tenancyconfigCache.length > 0) {
      return this.q.when(this.tenancyconfigCache);
    }

    return this.doRequest({
      targets: [
        {
          environment: this.environment,
          datasourceId: this.id,
          queryType: "tenancyconfig",
        },
      ],
      range: this.timeSrv.timeRange(),
    }).then((items) => {
      this.tenancyconfigCache = this.mapToTextValue(items, "tenancyconfig");
      return this.tenancyconfigCache;
    });
  }

  getCompartments() {
    if (this.compartmentsCache && this.compartmentsCache.length > 0) {
      return this.q.when(this.compartmentsCache);
    }

    return this.doRequest({
      targets: [
        {
          environment: this.environment,
          datasourceId: this.id,
          tenancyOCID: this.tenancyOCID,
          queryType: "compartments",
          region: this.defaultRegion, // compartments are registered for the all regions, so no difference which region to use here
        },
      ],
      range: this.timeSrv.timeRange(),
    }).then((items) => {
      this.compartmentsCache = this.mapToTextValue(items, "compartments");
      return this.compartmentsCache;
    });
  }

  getCompartmentId(compartment) {
    return this.getCompartments().then((compartments) => {
      const compartmentFound = compartments.find(
        (c) => c.text === compartment || c.value === compartment
      );
      return compartmentFound ? compartmentFound.value : compartment;
    });
  }

  async getNamespaces(target) {
    const region =
      target.region === SELECT_PLACEHOLDERS.REGION
        ? ""
        : this.getVariableValue(target.region);
    const compartment =
      target.compartment === SELECT_PLACEHOLDERS.COMPARTMENT
        ? ""
        : this.getVariableValue(target.compartment);
    if (_.isEmpty(compartment)) {
      return this.q.when([]);
    }

    const compartmentId = await this.getCompartmentId(compartment);
    return this.doRequest({
      targets: [
        {
          environment: this.environment,
          datasourceId: this.id,
          tenancyOCID: this.tenancyOCID,
          queryType: "namespaces",
          region: _.isEmpty(region) ? this.defaultRegion : region,
          compartment: compartmentId,
        },
      ],
      range: this.timeSrv.timeRange(),
    }).then((items) => {
      return this.mapToTextValue(items, "namespaces");
    });
  }

  async getResourceGroups(target) {
    const region =
      target.region === SELECT_PLACEHOLDERS.REGION
        ? ""
        : this.getVariableValue(target.region);
    const compartment =
      target.compartment === SELECT_PLACEHOLDERS.COMPARTMENT
        ? ""
        : this.getVariableValue(target.compartment);
    const namespace =
      target.namespace === SELECT_PLACEHOLDERS.NAMESPACE
        ? ""
        : this.getVariableValue(target.namespace);
    if (_.isEmpty(compartment)) {
      return this.q.when([]);
    }

    const compartmentId = await this.getCompartmentId(compartment);
    return this.doRequest({
      targets: [
        {
          environment: this.environment,
          datasourceId: this.id,
          tenancyOCID: this.tenancyOCID,
          queryType: "resourcegroups",
          region: _.isEmpty(region) ? this.defaultRegion : region,
          compartment: compartmentId,
          namespace: namespace,
        },
      ],
      range: this.timeSrv.timeRange(),
    }).then((items) => {
      return this.mapToTextValue(items, "resourcegroups");
    });
  }

  async getDimensions(target) {
    const region =
      target.region === SELECT_PLACEHOLDERS.REGION
        ? ""
        : this.getVariableValue(target.region);
    const compartment =
      target.compartment === SELECT_PLACEHOLDERS.COMPARTMENT
        ? ""
        : this.getVariableValue(target.compartment);
    const namespace =
      target.namespace === SELECT_PLACEHOLDERS.NAMESPACE
        ? ""
        : this.getVariableValue(target.namespace);
    const resourcegroup =
      target.resourcegroup === SELECT_PLACEHOLDERS.RESOURCEGROUP
        ? DEFAULT_RESOURCE_GROUP
        : this.getVariableValue(target.resourcegroup);
    const metric =
      target.metric === SELECT_PLACEHOLDERS.METRIC
        ? ""
        : this.getVariableValue(target.metric);
    const metrics =
      metric.startsWith("{") && metric.endsWith("}")
        ? metric.slice(1, metric.length - 1).split(",")
        : [metric];

    if (_.isEmpty(compartment) || _.isEmpty(namespace) || _.isEmpty(metrics)) {
      return this.q.when([]);
    }

    const dimensionsMap = {};
    for (let m of metrics) {
      if (dimensionsMap[m] !== undefined) {
        continue;
      }
      dimensionsMap[m] = null;

      const compartmentId = await this.getCompartmentId(compartment);
      await this.doRequest({
        targets: [
          {
            environment: this.environment,
            datasourceId: this.id,
            tenancyOCID: this.tenancyOCID,
            queryType: "dimensions",
            region: _.isEmpty(region) ? this.defaultRegion : region,
            compartment: compartmentId,
            namespace: namespace,
            resourcegroup: resourcegroup,
            metric: m,
          },
        ],
        range: this.timeSrv.timeRange(),
      })
        .then((result) => {
          const items = this.mapToTextValue(result, "dimensions");
          dimensionsMap[m] = [].concat(items);
        })
        .finally(() => {
          if (!dimensionsMap[m]) {
            dimensionsMap[m] = [];
          }
        });
    }

    let result = [];
    Object.values(dimensionsMap).forEach((dims) => {
      if (_.isEmpty(result)) {
        result = dims;
      } else {
        const newResult = [];
        dims.forEach((dim) => {
          if (
            !!result.find((d) => d.value === dim.value) &&
            !newResult.find((d) => d.value === dim.value)
          ) {
            newResult.push(dim);
          }
        });
        result = newResult;
      }
    });

    return result;
  }

  getDimensionKeys(target) {
    return this.getDimensions(target)
      .then((dims) => {
        const dimCache = dims.reduce((data, item) => {
          const values = item.value.split("=") || [];
          const key = values[0] || item.value;
          const value = values[1];

          if (!data[key]) {
            data[key] = [];
          }
          data[key].push(value);
          return data;
        }, {});
        return Object.keys(dimCache);
      })
      .then((items) => {
        return items.map((item) => ({ text: item, value: item }));
      });
  }

  getDimensionValues(target, dimKey) {
    return this.getDimensions(target)
      .then((dims) => {
        const dimCache = dims.reduce((data, item) => {
          const values = item.value.split("=") || [];
          const key = values[0] || item.value;
          const value = values[1];

          if (!data[key]) {
            data[key] = [];
          }
          data[key].push(value);
          return data;
        }, {});
        return dimCache[this.getVariableValue(dimKey)] || [];
      })
      .then((items) => {
        return items.map((item) => ({ text: item, value: item }));
      });
  }

  getAggregations() {
    return this.q.when(aggregations);
  }

  /**
   * Calls grafana backend.
   * Retries 10 times before failure.
   */
  doRequest(options) {
    let _this = this;
    return retryOrThrow(() => {
      return _this.backendSrv.datasourceRequest({
        url: "/api/ds/query",
        method: "POST",
        data: {
          from: options.range.from.valueOf().toString(),
          to: options.range.to.valueOf().toString(),
          queries: options.targets,
        },
      });
    }, 10).then((res) => toDataQueryResponse(res, options));
  }

  /**
   * Converts data from grafana backend to UI format
   */
  mapToTextValue(result, searchField) {
    if (_.isEmpty(result)) return [];

    // All drop-downs send a request to the backend and based on the query type, the backend sends a response
    // Depending on the data available , options are shaped
    // Values in fields are of type vectors (Based on the info from Grafana)

    switch (searchField) {
      case "compartments":
        return result.data[0].fields[0].values.toArray().map((name, i) => ({
          text: name,
          value: result.data[0].fields[1].values.toArray()[i],
        }));
      case "regions":
      case "tenancyconfig":
      case "namespaces":
      case "resourcegroups":
      case "search":
      case "dimensions":
        return result.data[0].fields[0].values.toArray().map((name) => ({
          text: name,
          value: name,
        }));
      // remaining  cases will be completed once the fix works for the above two
      default:
        return {};
    }
  }

  // **************************** Template variables helpers ****************************

  /**
   * Get all template variable descriptors
   */
  getVariableDescriptors(regex, includeCustom = true) {
    const vars = this.templateSrv.variables || [];

    if (regex) {
      let regexVars = vars.filter((item) => _.isString(item.query) && item.query.match(regex) !== null);
      if (includeCustom) {
        const custom = vars.filter(
          (item) => item.type === "custom" || item.type === "constant"
        );
        regexVars = regexVars.concat(custom);
      }
      const uniqueRegexVarsMap = new Map();
      regexVars.forEach((varObj) =>
        uniqueRegexVarsMap.set(varObj.name, varObj)
      );
      return Array.from(uniqueRegexVarsMap.values());
    }
    return vars;
  }

  /**
   * List all variable names optionally filtered by regex or/and type
   * Returns list of names with '$' at the beginning. Example: ['$dimensionKey', '$dimensionValue']
   *
   * Updates:
   * Notes on implementation :
   * If a custom or constant is in  variables and  includeCustom, default is false.
   * Hence,the varDescriptors list is filtered for a unique set of var names
   */
  getVariables(regex, includeCustom) {
    const varDescriptors =
      this.getVariableDescriptors(regex, includeCustom) || [];
    return varDescriptors.map((item) => `$${item.name}`);
  }

  /**
   * @param varName valid varName contains '$'. Example: '$dimensionKey'
   * Returns an array with variable values or empty array
   */
  getVariableValue(varName, scopedVars = {}) {
    return this.templateSrv.replace(varName, scopedVars) || varName;
  }

  /**
   * @param varName valid varName contains '$'. Example: '$dimensionKey'
   * Returns true if variable with the given name is found
   */
  isVariable(varName) {
    const varNames = this.getVariables() || [];
    return !!varNames.find((item) => item === varName);
  }
}
