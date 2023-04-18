/*
 ** Copyright © 2022 Oracle and/or its affiliates. All rights reserved.
 ** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
 */
import _ from "lodash";
import {
  DataQueryResponse,
  DataSourceInstanceSettings,
} from '@grafana/data';
import { DataQuery } from '@grafana/schema';
import { QueryEditorProps } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { OCIQuery, OCIConfig } from './types';
import {
  aggregations,
  dimensionKeysQueryRegex,
  namespacesQueryRegex,
  resourcegroupsQueryRegex,
  metricsQueryRegex,
  regionsQueryRegex,
  tenanciesQueryRegex,
  compartmentsQueryRegex,
  dimensionValuesQueryRegex,
  removeQuotes,
  AUTO,
} from "./constants";
import retryOrThrow from "./util/retry";
// import { SELECT_PLACEHOLDERS } from "./query_ctrl";
import { resolveAutoWinRes } from "./util/utilFunctions";
import { toDataQueryResponse } from "@grafana/runtime";

const DEFAULT_RESOURCE_GROUP = "NoResourceGroup";
const DEFAULT_TENANCY = "NoTenancy";


export class OCIDataSource extends DataSourceWithBackend<OCIQuery, OCIConfig> {
  annotations = {};
  settings: DataSourceInstanceSettings<OCIConfig>;
  q: any;
  defaultRegion: any;
  environment: any;
  tenancymode: any;
  timeSrv: any;
  regionsCache: any;
  tenanciesCache: boolean;
  compartmentsCache: any;
  backendSrv: any;
  templateSrv: any;

  constructor(instanceSettings: DataSourceInstanceSettings<OCIConfig>) {
    super(instanceSettings);
    this.settings = instanceSettings;
  }

// export default class OCIDatasource {
//   constructor(instanceSettings: {[key: string]: any}, $q: {[key: string]: any}, backendSrv: {[key: string]: any}, templateSrv: {[key: string]: any}, timeSrv: {[key: string]: any}) {
//     this.type = instanceSettings.type;
//     this.url = instanceSettings.url;
//     this.name = instanceSettings.name;
//     this.id = instanceSettings.id;
//     this.defaultRegion = instanceSettings.jsonData.defaultRegion;
//     this.environment = instanceSettings.jsonData.environment;
//     this.tenancymode = instanceSettings.jsonData.tenancymode;
//     this.q = $q;
//     this.backendSrv = backendSrv;
//     this.templateSrv = templateSrv;
//     this.timeSrv = timeSrv;

//     this.compartmentsCache = [];
//     this.regionsCache = [];
//     this.tenanciesCache = [];

//     // this.getRegions();
//     // this.getCompartments();
//   }

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
  async query(options: any) {
    var query = await this.buildQueryParameters(options);
    if (query.targets.length <= 0) {
      return this.q.when({ data: [] });
    }

    return this.doRequest(query).then((result) => {
      var res: any[] = [];
      _.forEach(result.data, (r) => {
        const name = r.fields[1].config.displayNameFromDS;
        const timeArr = r.fields[0].values.toArray();
        const values = r.fields[1].values.toArray();
        const points = timeArr.map((t: any, i: string | number) => [values[i], t]);
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
  testDatasource(): Promise<{ status: string; message: string; title: string } | undefined> {
    return this.doRequest({
      targets: [
        {
          queryType: "test",
          region: this.defaultRegion,
          compartment: "",
          environment: this.environment,
          tenancymode: this.tenancymode,
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
  async metricFindQuery(target: { tenancy: any; region: any; compartment: any; namespace: any; resourcegroup: any; }) {
    if (typeof target === "string") {
      // used in template editor for creating variables
      return this.templateMetricQuery(target);
    }
    const region =
      target.region === SELECT_PLACEHOLDERS.REGION
        ? ""
        : this.getVariableValue(target.region);
    const tenancy =
      target.tenancy === SELECT_PLACEHOLDERS.TENANCY
        ? DEFAULT_TENANCY
        : this.getVariableValue(target.tenancy);        
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

    const compartmentId = await this.getCompartmentId(compartment, target);
    return this.doRequest({
      targets: [
        {
          environment: this.environment,
          tenancymode: this.tenancymode,
          datasourceId: this.id,
          queryType: "search",
          region: _.isEmpty(region) ? this.defaultRegion : region,
          compartment: compartmentId,
          namespace: namespace,
          resourcegroup: resourcegroup,
          tenancy: tenancy,
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
  async buildQueryParameters(options: { targets: any[]; scopedVars: {} | undefined; }) {
    let queries = options.targets
      .filter((t: { hide: any; }) => !t.hide)
      .filter(
        (t: { compartment: any; }) =>
          !_.isEmpty(
            this.getVariableValue(t.compartment, options.scopedVars)
          ) && t.compartment !== SELECT_PLACEHOLDERS.COMPARTMENT
      )    
      .filter(
        (t: { namespace: any; }) =>
          !_.isEmpty(this.getVariableValue(t.namespace, options.scopedVars)) &&
          t.namespace !== SELECT_PLACEHOLDERS.NAMESPACE
      )
      .filter(
        (t: { resourcegroup: any; }) =>
          !_.isEmpty(this.getVariableValue(t.resourcegroup, options.scopedVars))
      )
      .filter(
        (t: { metric: any; target: any; }) =>
          (!_.isEmpty(this.getVariableValue(t.metric, options.scopedVars)) &&
            t.metric !== SELECT_PLACEHOLDERS.METRIC) ||
          !_.isEmpty(this.getVariableValue(t.target))
      );

    queries.forEach((t: { dimensions: any; resourcegroup: any; }) => {
      t.dimensions = (t.dimensions || [])
        .filter(
          (dim: { key: any; }) =>
            !_.isEmpty(dim.key) && dim.key !== SELECT_PLACEHOLDERS.DIMENSION_KEY
        )
        .filter(
          (dim: { value: any; }) =>
            !_.isEmpty(dim.value) &&
            dim.value !== SELECT_PLACEHOLDERS.DIMENSION_VALUE
        );

        t.resourcegroup =
        t.resourcegroup === SELECT_PLACEHOLDERS.RESOURCEGROUP
          ? DEFAULT_RESOURCE_GROUP
          : t.resourcegroup;          
    });

    // we support multiselect for dimension values, so we need to parse 1 query into multiple queries
    queries = this.splitMultiValueDimensionsIntoQueries(queries, options);

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
        const dimensions = (t.dimensions || []).reduce((result: string[], dim: { key: any; operator: any; value: any; }) => {
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

      const tenancy =
      this.getVariableValue(t.tenancy, options.scopedVars) === SELECT_PLACEHOLDERS.TENANCY
        ? DEFAULT_TENANCY
        : this.getVariableValue(t.tenancy, options.scopedVars);        
      let target = {
        tenancy: tenancy,
        region: _.isEmpty(region) ? this.defaultRegion : region,
      }; 
      const compartmentId = await this.getCompartmentId(
        this.getVariableValue(t.compartment, options.scopedVars), target
      );

      const result = {
        resolution,
        environment: this.environment,
        tenancymode: this.tenancymode,
        datasourceId: this.id,
        queryType: "query",
        refId: t.refId,
        hide: t.hide,
        type: t.type || "timeserie",
        region: _.isEmpty(region) ? this.defaultRegion : region,
        compartment: compartmentId,
        tenancy: tenancy,
        namespace: this.getVariableValue(t.namespace, options.scopedVars),
        resourcegroup: this.getVariableValue(
          t.resourcegroup,
          options.scopedVars
        ),
        query: query,
        legendFormat: t.legendFormat,
      };
      results.push(result);
    }

    options.targets = results;

    return options;
  }

  /**
   * Splits queries with multi valued dimensions into several queries.
   * Example:
   * "DeliverySucceedEvents[1m]{resourceDisplayName = ["ResouceName_1","ResouceName_1"], eventType = ["Create","Delete"]}.mean()" ->
   *  [
   *    "DeliverySucceedEvents[1m]{resourceDisplayName = "ResouceName_1", eventType = "Create"}.mean()",
   *    "DeliverySucceedEvents[1m]{resourceDisplayName = "ResouceName_2", eventType = "Create"}.mean()",
   *    "DeliverySucceedEvents[1m]{resourceDisplayName = "ResouceName_1", eventType = "Delete"}.mean()",
   *    "DeliverySucceedEvents[1m]{resourceDisplayName = "ResouceName_2", eventType = "Delete"}.mean()",
   *  ]
   */
  splitMultiValueDimensionsIntoQueries(queries: any[], options: { scopedVars: {} | undefined; }) {
    return queries.reduce((data: any[], t: {
      refId: string; dimensions: any[]; target: any; 
}) => {
      if (_.isEmpty(t.dimensions) || !_.isEmpty(t.target)) {
        // nothing to split or dimensions won't be used, query is set manually
        return data.concat(t);
      }

      // create a map key : [values] for multiple values
      const multipleValueDims = t.dimensions.reduce((data: { [x: string]: any; }, dim: { key: any; value: any; }) => {
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

      const splitDimensions = (dims: string | any[], multiDims: { [x: string]: any[]; }) => {
        let prev = [];
        let next: any[] = [];

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
  templateMetricQuery(varString: string) {

    let tenancyQuery = varString.match(tenanciesQueryRegex);
    if (tenancyQuery) {
      return this.getTenancies().catch((err: string) => {
        throw new Error("Unable to get tenancies: " + err);
      });    
    }    

    let regionQuery = varString.match(regionsQueryRegex);
    if (regionQuery) {
      if (this.tenancymode === "multitenancy") {
        let target = {
          tenancy: removeQuotes(this.getVariableValue(regionQuery[1])),
        };
        return this.getRegions(target).catch((err) => {
          throw new Error("Unable to get regions: " + err);
        });
      } else {
        let target = {
          tenancy: DEFAULT_TENANCY,
        };        
        return this.getRegions(target).catch((err) => {
          throw new Error("Unable to get regions: " + err);
        });        
      }
    }

    let compartmentQuery = varString.match(compartmentsQueryRegex);
    if (compartmentQuery){
      if (this.tenancymode === "multitenancy") {
        let target = {
          tenancy: removeQuotes(this.getVariableValue(compartmentQuery[1])),
        };       
        return this.getCompartments(target)
          .then((compartments) => {
            return compartments.map((c: { text: any; }) => ({ text: c.text, value: c.text }));
          })
          .catch((err) => {
            throw new Error("Unable to get compartments: " + err);
          });
      } else {
          let target = {
            tenancy: DEFAULT_TENANCY,
          };        
          return this.getCompartments(target)
            .then((compartments) => {
              return compartments.map((c: { text: any; }) => ({ text: c.text, value: c.text }));
            })
            .catch((err) => {
              throw new Error("Unable to get compartments: " + err);
            });  
      }   
    }


    let namespaceQuery = varString.match(namespacesQueryRegex);
    if (namespaceQuery) {
      if (this.tenancymode === "multitenancy") {
        let target = {
          tenancy: removeQuotes(this.getVariableValue(namespaceQuery[1])),
          region: removeQuotes(this.getVariableValue(namespaceQuery[2])),
          compartment: removeQuotes(this.getVariableValue(namespaceQuery[3])),
        };
        return this.getNamespaces(target).catch((err) => {
          throw new Error("Unable to get namespaces: " + err);
        });
      } else {
        let target = {
          tenancy: DEFAULT_TENANCY,
          region: removeQuotes(this.getVariableValue(namespaceQuery[1])),
          compartment: removeQuotes(this.getVariableValue(namespaceQuery[2])),
        };
        return this.getNamespaces(target).catch((err) => {
          throw new Error("Unable to get namespaces: " + err);
        });        
      }
    }

    let resourcegroupQuery = varString.match(resourcegroupsQueryRegex);
    if (resourcegroupQuery) {
      if (this.tenancymode === "multitenancy") {
        let target = {
          tenancy: removeQuotes(this.getVariableValue(resourcegroupQuery[1])),
          region: removeQuotes(this.getVariableValue(resourcegroupQuery[2])),
          compartment: removeQuotes(this.getVariableValue(resourcegroupQuery[3])),
          namespace: removeQuotes(this.getVariableValue(resourcegroupQuery[4])),
        };
        return this.getResourceGroups(target).catch((err) => {
          throw new Error("Unable to get resourcegroups: " + err);
        });
      } else {
        let target = {
          tenancy: DEFAULT_TENANCY,
          region: removeQuotes(this.getVariableValue(resourcegroupQuery[1])),
          compartment: removeQuotes(this.getVariableValue(resourcegroupQuery[2])),
          namespace: removeQuotes(this.getVariableValue(resourcegroupQuery[3])),
        };
        return this.getResourceGroups(target).catch((err) => {
          throw new Error("Unable to get resourcegroups: " + err);
        });        
      }
    }

    let metricQuery = varString.match(metricsQueryRegex);
    if (metricQuery) {
      if (this.tenancymode === "multitenancy") {
        let target = {
          tenancy: removeQuotes(this.getVariableValue(metricQuery[1])),
          region: removeQuotes(this.getVariableValue(metricQuery[2])),
          compartment: removeQuotes(this.getVariableValue(metricQuery[3])),
          namespace: removeQuotes(this.getVariableValue(metricQuery[4])),
          resourcegroup: removeQuotes(this.getVariableValue(metricQuery[5])),
        };
        return this.metricFindQuery(target).catch((err: string) => {
          throw new Error("Unable to get metrics: " + err);
        });
      } else {
        let target = {
          tenancy: DEFAULT_TENANCY,
          region: removeQuotes(this.getVariableValue(metricQuery[1])),
          compartment: removeQuotes(this.getVariableValue(metricQuery[2])),
          namespace: removeQuotes(this.getVariableValue(metricQuery[3])),
          resourcegroup: removeQuotes(this.getVariableValue(metricQuery[4])),
        };
        return this.metricFindQuery(target).catch((err: string) => {
          throw new Error("Unable to get metrics: " + err);
        });        
      }  
    }

    let dimensionsQuery = varString.match(dimensionKeysQueryRegex);
    if (dimensionsQuery) {
      if (this.tenancymode === "multitenancy") {
        let target = {
          tenancy: removeQuotes(this.getVariableValue(dimensionsQuery[1])),
          region: removeQuotes(this.getVariableValue(dimensionsQuery[2])),
          compartment: removeQuotes(this.getVariableValue(dimensionsQuery[3])),
          namespace: removeQuotes(this.getVariableValue(dimensionsQuery[4])),
          metric: removeQuotes(this.getVariableValue(dimensionsQuery[5])),
          resourcegroup: removeQuotes(this.getVariableValue(dimensionsQuery[6])),
        };
        return this.getDimensionKeys(target).catch((err) => {
          throw new Error("Unable to get dimensions: " + err);
        });
      } else {
        let target = {
          tenancy: DEFAULT_TENANCY,
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
    }

    let dimensionOptionsQuery = varString.match(dimensionValuesQueryRegex);
    if (dimensionOptionsQuery) {
      if (this.tenancymode === "multitenancy") {
        let target = {
          tenancy: removeQuotes(this.getVariableValue(dimensionOptionsQuery[1])),
          region: removeQuotes(this.getVariableValue(dimensionOptionsQuery[2])),
          compartment: removeQuotes(this.getVariableValue(dimensionOptionsQuery[3])),
          namespace: removeQuotes(this.getVariableValue(dimensionOptionsQuery[4])),
          metric: removeQuotes(this.getVariableValue(dimensionOptionsQuery[5])),
          resourcegroup: removeQuotes(this.getVariableValue(dimensionOptionsQuery[7])),
        };
        let dimensionKey = removeQuotes(this.getVariableValue(dimensionOptionsQuery[6]));
        return this.getDimensionValues(target, dimensionKey).catch((err) => {
          throw new Error("Unable to get dimension options: " + err);
        });
      } else {
        let target = {
          tenancy: DEFAULT_TENANCY,
          region: removeQuotes(this.getVariableValue(dimensionOptionsQuery[1])),
          compartment: removeQuotes(this.getVariableValue(dimensionOptionsQuery[2])),
          namespace: removeQuotes(this.getVariableValue(dimensionOptionsQuery[3])),
          metric: removeQuotes(this.getVariableValue(dimensionOptionsQuery[4])),
          resourcegroup: removeQuotes(this.getVariableValue(dimensionOptionsQuery[6])),
        };
        let dimensionKey = removeQuotes(this.getVariableValue(dimensionOptionsQuery[5]));        
        return this.getDimensionValues(target, dimensionKey).catch((err) => {
          throw new Error("Unable to get dimension options: " + err);
        });        
      }
    }

    throw new Error("Unable to parse templating string");
  }

  async getRegions(target: { tenancy: any; }) {
    const tenancy =
        target.tenancy === SELECT_PLACEHOLDERS.TENANCY
          ? DEFAULT_TENANCY
          : this.getVariableValue(target.tenancy);

    return this.doRequest({
      targets: [
        {
          environment: this.environment,
          tenancymode: this.tenancymode,
          datasourceId: this.id,
          tenancy: _.isEmpty(tenancy) ? "" : tenancy,
          queryType: "regions",
        },
      ],
      range: this.timeSrv.timeRange(),
    }).then((items) => {
      this.regionsCache = this.mapToTextValue(items, "regions");
      return this.regionsCache;
    });
  }

  getTenancies() {
    if (this.tenanciesCache && this.tenanciesCache.length > 0) {
      return this.q.when(this.tenanciesCache);
    }

    return this.doRequest({
      targets: [
        {
          environment: this.environment,
          tenancymode: this.tenancymode,
          datasourceId: this.id,
          queryType: "tenancies",
        },
      ],
      range: this.timeSrv.timeRange(),
    }).then((items) => {
      this.tenanciesCache = this.mapToTextValue(items, "tenancies");
      return this.tenanciesCache;
    });
  }

  async getCompartments(target: { tenancy: any; }) {
    const tenancy =
        target.tenancy === SELECT_PLACEHOLDERS.TENANCY
          ? DEFAULT_TENANCY
          : this.getVariableValue(target.tenancy);

    return this.doRequest({
      targets: [
        {
          environment: this.environment,
          tenancymode: this.tenancymode,
          datasourceId: this.id,
          tenancy: _.isEmpty(tenancy) ? "" : tenancy,
          queryType: "compartments",
        },
      ],
      range: this.timeSrv.timeRange(),
    }).then((items) => {
      this.compartmentsCache = this.mapToTextValue(items, "compartments");  
      return this.compartmentsCache;
    });
  }

  getCompartmentId(compartment: any, target: { tenancy: any; region: any; }) {   
    return this.getCompartments(target).then((compartments) => {
      const compartmentFound = compartments.find(
        (c: { text: any; value: any; }) => c.text === compartment || c.value === compartment
      );
      return compartmentFound ? compartmentFound.value : compartment;
    });
  }

  async getNamespaces(target: { tenancy: any; region: any; compartment: any; }) {
    const region =
      target.region === SELECT_PLACEHOLDERS.REGION
        ? ""
        : this.getVariableValue(target.region);
    const compartment =
      target.compartment === SELECT_PLACEHOLDERS.COMPARTMENT
        ? ""
        : this.getVariableValue(target.compartment);
    const tenancy =
      target.tenancy === SELECT_PLACEHOLDERS.TENANCY
        ? DEFAULT_TENANCY
        : this.getVariableValue(target.tenancy);         
    if (_.isEmpty(compartment)) {
      return this.q.when([]);
    }

    const compartmentId = await this.getCompartmentId(compartment, target);
    return this.doRequest({
      targets: [
        {
          environment: this.environment,
          tenancymode: this.tenancymode,
          datasourceId: this.id,
          queryType: "namespaces",
          region: _.isEmpty(region) ? this.defaultRegion : region,
          compartment: compartmentId,
          tenancy: tenancy,
        },
      ],
      range: this.timeSrv.timeRange(),
    }).then((items) => {
      return this.mapToTextValue(items, "namespaces");
    });
  }

  async getResourceGroups(target: { tenancy: any; region: any; compartment: any; namespace: any; }) {
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
    const tenancy =
      target.tenancy === SELECT_PLACEHOLDERS.TENANCY
        ? DEFAULT_TENANCY
        : this.getVariableValue(target.tenancy);        
    if (_.isEmpty(compartment)) {
      return this.q.when([]);
    }

    const compartmentId = await this.getCompartmentId(compartment, target);
    return this.doRequest({
      targets: [
        {
          environment: this.environment,
          tenancymode: this.tenancymode,
          datasourceId: this.id,
          queryType: "resourcegroups",
          region: _.isEmpty(region) ? this.defaultRegion : region,
          compartment: compartmentId,
          tenancy: tenancy,
          namespace: namespace,
        },
      ],
      range: this.timeSrv.timeRange(),
    }).then((items) => {
      return this.mapToTextValue(items, "resourcegroups");
    });
  }

  async getDimensions(target: { region: any; compartment: any; namespace: any; resourcegroup: any; metric: any; tenancy: any; }) {
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
    const tenancy =
      target.tenancy === SELECT_PLACEHOLDERS.TENANCY
        ? DEFAULT_TENANCY
        : this.getVariableValue(target.tenancy);         
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

      const compartmentId = await this.getCompartmentId(compartment, target);
      await this.doRequest({
        targets: [
          {
            environment: this.environment,
            tenancymode: this.tenancymode,
            datasourceId: this.id,
            queryType: "dimensions",
            region: _.isEmpty(region) ? this.defaultRegion : region,
            compartment: compartmentId,
            namespace: namespace,
            resourcegroup: resourcegroup,
            tenancy: tenancy,
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

    let result: unknown = [];
    Object.values(dimensionsMap).forEach((dims) => {
      if (_.isEmpty(result)) {
        result = dims;
      } else {
        const newResult: any[] = [];
        dims.forEach((dim: { value: any; }) => {
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

  getDimensionKeys(target: { tenancy: string; region: string; compartment: string; namespace: string; metric: string; resourcegroup: string; }) {
    return this.getDimensions(target)
      .then((dims) => {
        const dimCache = dims.reduce((data: { [x: string]: any[]; }, item: { value: { split: (arg0: string) => never[]; }; }) => {
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

  getDimensionValues(target: { tenancy: string; region: string; compartment: string; namespace: string; metric: string; resourcegroup: string; }, dimKey: string) {
    return this.getDimensions(target)
      .then((dims) => {
        const dimCache = dims.reduce((data: { [x: string]: any[]; }, item: { value: { split: (arg0: string) => never[]; }; }) => {
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
        return items.map((item: any) => ({ text: item, value: item }));
      });
  }

  getAggregations() {
    return this.q.when(aggregations);
  }

  /**
   * Calls grafana backend.
   * Retries 10 times before failure.
   */
  doRequest(options: DataQuery[] | undefined) {
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
  mapToTextValue(result: DataQueryResponse, searchField: string) {
    if (_.isEmpty(result)) return [];

    // All drop-downs send a request to the backend and based on the query type, the backend sends a response
    // Depending on the data available , options are shaped
    // Values in fields are of type vectors (Based on the info from Grafana)

    switch (searchField) {
      case "compartments":
        return result.data[0].fields[0].values.toArray().map((name: any, i: string | number) => ({
          text: name,
          value: result.data[0].fields[1].values.toArray()[i],
        }));
      case "regions":
      case "tenancies":       
      case "namespaces":
      case "resourcegroups":
      case "search":
      case "dimensions":
        return result.data[0].fields[0].values.toArray().map((name: any) => ({
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
  getVariableDescriptors(regex: any, includeCustom = true) {
    const vars = this.templateSrv.variables || [];

    if (regex) {
      let regexVars = vars.filter((item: { query: { match: (arg0: any) => null; }; }) => _.isString(item.query) && item.query.match(regex) !== null);
      if (includeCustom) {
        const custom = vars.filter(
          (item: { type: string; }) => item.type === "custom" || item.type === "constant"
        );
        regexVars = regexVars.concat(custom);
      }
      const uniqueRegexVarsMap = new Map();
      regexVars.forEach((varObj: { name: any; }) =>
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
  getVariables(regex: undefined, includeCustom: boolean | undefined) {
    const varDescriptors =
      this.getVariableDescriptors(regex, includeCustom) || [];
    return varDescriptors.map((item: { name: any; }) => `$${item.name}`);
  }

  /**
   * @param varName valid varName contains '$'. Example: '$dimensionKey'
   * Returns an array with variable values or empty array
   */
  getVariableValue(varName: any, scopedVars = {}) {
    return this.templateSrv.replace(varName, scopedVars) || varName;
  }

  /**
   * @param varName valid varName contains '$'. Example: '$dimensionKey'
   * Returns true if variable with the given name is found
   */
  isVariable(varName: any) {
    const varNames = this.getVariables() || [];
    return !!varNames.find((item: any) => item === varName);
  }
}
