'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }(); /*
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     ** Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     ** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     */


var _lodash = require('lodash');

var _lodash2 = _interopRequireDefault(_lodash);

var _constants = require('./constants');

var _retry = require('./util/retry');

var _retry2 = _interopRequireDefault(_retry);

var _query_ctrl = require('./query_ctrl');

var _utilFunctions = require('./util/utilFunctions');

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var DEFAULT_RESOURCE_GROUP = 'NoResourceGroup';

var OCIDatasource = function () {
  function OCIDatasource(instanceSettings, $q, backendSrv, templateSrv, timeSrv) {
    _classCallCheck(this, OCIDatasource);

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


  _createClass(OCIDatasource, [{
    key: 'query',
    value: async function query(options) {
      var query = await this.buildQueryParameters(options);
      if (query.targets.length <= 0) {
        return this.q.when({ data: [] });
      }

      return this.doRequest(query).then(function (result) {
        var res = [];
        _lodash2.default.forEach(result.data.results, function (r) {
          _lodash2.default.forEach(r.series, function (s) {
            res.push({ target: s.name, datapoints: s.points });
          });
          _lodash2.default.forEach(r.tables, function (t) {
            t.type = 'table';
            t.refId = r.refId;
            res.push(t);
          });
        });

        result.data = res;
        return result;
      });
    }

    /**
     * Required method
     * Used by data source configuration page to make sure the connection is working
     */

  }, {
    key: 'testDatasource',
    value: function testDatasource() {
      return this.doRequest({
        targets: [{
          queryType: 'test',
          region: this.defaultRegion,
          tenancyOCID: this.tenancyOCID,
          compartment: '',
          environment: this.environment,
          datasourceId: this.id
        }],
        range: this.timeSrv.timeRange()
      }).then(function (response) {
        if (response.status === 200) {
          return { status: 'success', message: 'Data source is working', title: 'Success' };
        }
      }).catch(function () {
        return { status: 'error', message: 'Data source is not working', title: 'Failure' };
      });
    }

    /**
     * Required method
     * Used by query editor to get metric suggestions
     */

  }, {
    key: 'metricFindQuery',
    value: async function metricFindQuery(target) {
      var _this2 = this;

      if (typeof target === 'string') {
        // used in template editor for creating variables
        return this.templateMetricQuery(target);
      }
      var region = target.region === _query_ctrl.SELECT_PLACEHOLDERS.REGION ? '' : this.getVariableValue(target.region);
      var compartment = target.compartment === _query_ctrl.SELECT_PLACEHOLDERS.COMPARTMENT ? '' : this.getVariableValue(target.compartment);
      var namespace = target.namespace === _query_ctrl.SELECT_PLACEHOLDERS.NAMESPACE ? '' : this.getVariableValue(target.namespace);
      var resourcegroup = target.resourcegroup === _query_ctrl.SELECT_PLACEHOLDERS.RESOURCEGROUP ? DEFAULT_RESOURCE_GROUP : this.getVariableValue(target.resourcegroup);

      if (_lodash2.default.isEmpty(compartment) || _lodash2.default.isEmpty(namespace)) {
        return this.q.when([]);
      }

      var compartmentId = await this.getCompartmentId(compartment);
      return this.doRequest({
        targets: [{
          environment: this.environment,
          datasourceId: this.id,
          tenancyOCID: this.tenancyOCID,
          queryType: 'search',
          region: _lodash2.default.isEmpty(region) ? this.defaultRegion : region,
          compartment: compartmentId,
          namespace: namespace,
          resourcegroup: resourcegroup
        }],
        range: this.timeSrv.timeRange()
      }).then(function (res) {
        return _this2.mapToTextValue(res, 'search');
      });
    }

    /**
     * Build and validate query parameters.
     */

  }, {
    key: 'buildQueryParameters',
    value: async function buildQueryParameters(options) {
      var _this3 = this;

      var queries = options.targets.filter(function (t) {
        return !t.hide;
      }).filter(function (t) {
        return !_lodash2.default.isEmpty(_this3.getVariableValue(t.compartment, options.scopedVars)) && t.compartment !== _query_ctrl.SELECT_PLACEHOLDERS.COMPARTMENT;
      }).filter(function (t) {
        return !_lodash2.default.isEmpty(_this3.getVariableValue(t.namespace, options.scopedVars)) && t.namespace !== _query_ctrl.SELECT_PLACEHOLDERS.NAMESPACE;
      }).filter(function (t) {
        return !_lodash2.default.isEmpty(_this3.getVariableValue(t.resourcegroup, options.scopedVars));
      }).filter(function (t) {
        return !_lodash2.default.isEmpty(_this3.getVariableValue(t.metric, options.scopedVars)) && t.metric !== _query_ctrl.SELECT_PLACEHOLDERS.METRIC || !_lodash2.default.isEmpty(_this3.getVariableValue(t.target));
      });

      queries.forEach(function (t) {
        t.dimensions = (t.dimensions || []).filter(function (dim) {
          return !_lodash2.default.isEmpty(dim.key) && dim.key !== _query_ctrl.SELECT_PLACEHOLDERS.DIMENSION_KEY;
        }).filter(function (dim) {
          return !_lodash2.default.isEmpty(dim.value) && dim.value !== _query_ctrl.SELECT_PLACEHOLDERS.DIMENSION_VALUE;
        });

        t.resourcegroup = t.resourcegroup === _query_ctrl.SELECT_PLACEHOLDERS.RESOURCEGROUP ? DEFAULT_RESOURCE_GROUP : t.resourcegroup;
      });

      // we support multiselect for dimension values, so we need to parse 1 query into multiple queries
      queries = this.splitMultiValueDimensionsIntoQuieries(queries, options);

      var results = [];
      var _iteratorNormalCompletion = true;
      var _didIteratorError = false;
      var _iteratorError = undefined;

      try {
        for (var _iterator = queries[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
          var t = _step.value;

          var region = t.region === _query_ctrl.SELECT_PLACEHOLDERS.REGION ? '' : this.getVariableValue(t.region, options.scopedVars);
          var query = this.getVariableValue(t.target, options.scopedVars);
          var numberOfDaysDiff = this.timeSrv.timeRange().to.diff(this.timeSrv.timeRange().from, 'days');
          // The following replaces 'auto' in window portion of the query and replaces it with an appropriate value.
          // If there is a functionality to access the window variable instead of matching [auto] in the query, it will be
          // better
          if (query) query = query.replace('[auto]', '[' + (0, _utilFunctions.resolveAutoWinRes)(_constants.AUTO, '', numberOfDaysDiff).window + ']');
          var resolution = this.getVariableValue(t.resolution, options.scopedVars);
          var window = t.window === _query_ctrl.SELECT_PLACEHOLDERS.WINDOW ? '' : this.getVariableValue(t.window, options.scopedVars);
          // p.s : timeSrv.timeRange() results in a moment object
          var resolvedWinResolObj = (0, _utilFunctions.resolveAutoWinRes)(window, resolution, numberOfDaysDiff);
          window = resolvedWinResolObj.window;
          resolution = resolvedWinResolObj.resolution;
          if (_lodash2.default.isEmpty(query)) {
            // construct query
            var dimensions = (t.dimensions || []).reduce(function (result, dim) {
              var d = _this3.getVariableValue(dim.key, options.scopedVars) + ' ' + dim.operator + ' "' + _this3.getVariableValue(dim.value, options.scopedVars) + '"';
              if (result.indexOf(d) < 0) {
                result.push(d);
              }
              return result;
            }, []);
            var dimension = _lodash2.default.isEmpty(dimensions) ? '' : '{' + dimensions.join(',') + '}';
            query = this.getVariableValue(t.metric, options.scopedVars) + '[' + window + ']' + dimension + '.' + t.aggregation;
          }

          var compartmentId = await this.getCompartmentId(this.getVariableValue(t.compartment, options.scopedVars));
          var result = {
            resolution: resolution,
            environment: this.environment,
            datasourceId: this.id,
            tenancyOCID: this.tenancyOCID,
            queryType: 'query',
            refId: t.refId,
            hide: t.hide,
            type: t.type || 'timeserie',
            region: _lodash2.default.isEmpty(region) ? this.defaultRegion : region,
            compartment: compartmentId,
            namespace: this.getVariableValue(t.namespace, options.scopedVars),
            resourcegroup: this.getVariableValue(t.resourcegroup, options.scopedVars),
            query: query
          };
          results.push(result);
        }
      } catch (err) {
        _didIteratorError = true;
        _iteratorError = err;
      } finally {
        try {
          if (!_iteratorNormalCompletion && _iterator.return) {
            _iterator.return();
          }
        } finally {
          if (_didIteratorError) {
            throw _iteratorError;
          }
        }
      }

      ;

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

  }, {
    key: 'splitMultiValueDimensionsIntoQuieries',
    value: function splitMultiValueDimensionsIntoQuieries(queries, options) {
      var _this4 = this;

      return queries.reduce(function (data, t) {

        if (_lodash2.default.isEmpty(t.dimensions) || !_lodash2.default.isEmpty(t.target)) {
          // nothing to split or dimensions won't be used, query is set manually
          return data.concat(t);
        }

        // create a map key : [values] for multiple values
        var multipleValueDims = t.dimensions.reduce(function (data, dim) {
          var key = dim.key;
          var value = _this4.getVariableValue(dim.value, options.scopedVars);
          if (value.startsWith("{") && value.endsWith("}")) {
            var values = value.slice(1, value.length - 1).split(',') || [];
            data[key] = (data[key] || []).concat(values);
          }
          return data;
        }, {});

        if (_lodash2.default.isEmpty(Object.keys(multipleValueDims))) {
          // no multiple values used, only single values
          return data.concat(t);
        }

        var splitDimensions = function splitDimensions(dims, multiDims) {
          var prev = [];
          var next = [];

          var firstDimKey = dims[0].key;
          var firstDimValues = multiDims[firstDimKey] || [dims[0].value];
          var _iteratorNormalCompletion2 = true;
          var _didIteratorError2 = false;
          var _iteratorError2 = undefined;

          try {
            for (var _iterator2 = firstDimValues[Symbol.iterator](), _step2; !(_iteratorNormalCompletion2 = (_step2 = _iterator2.next()).done); _iteratorNormalCompletion2 = true) {
              var _v = _step2.value;

              var _newDim = _lodash2.default.cloneDeep(dims[0]);
              _newDim.value = _v;
              prev.push([_newDim]);
            }
          } catch (err) {
            _didIteratorError2 = true;
            _iteratorError2 = err;
          } finally {
            try {
              if (!_iteratorNormalCompletion2 && _iterator2.return) {
                _iterator2.return();
              }
            } finally {
              if (_didIteratorError2) {
                throw _iteratorError2;
              }
            }
          }

          for (var i = 1; i < dims.length; i++) {
            var values = multiDims[dims[i].key] || [dims[i].value];
            var _iteratorNormalCompletion3 = true;
            var _didIteratorError3 = false;
            var _iteratorError3 = undefined;

            try {
              for (var _iterator3 = values[Symbol.iterator](), _step3; !(_iteratorNormalCompletion3 = (_step3 = _iterator3.next()).done); _iteratorNormalCompletion3 = true) {
                var v = _step3.value;

                for (var j = 0; j < prev.length; j++) {
                  if (next.length >= 20) {
                    // this algorithm of collecting multi valued dimensions is computantionally VERY expensive
                    // set the upper limit for quiries number
                    return next;
                  }
                  var newDim = _lodash2.default.cloneDeep(dims[i]);
                  newDim.value = v;
                  next.push(prev[j].concat(newDim));
                }
              }
            } catch (err) {
              _didIteratorError3 = true;
              _iteratorError3 = err;
            } finally {
              try {
                if (!_iteratorNormalCompletion3 && _iterator3.return) {
                  _iterator3.return();
                }
              } finally {
                if (_didIteratorError3) {
                  throw _iteratorError3;
                }
              }
            }

            prev = next;
            next = [];
          }

          return prev;
        };

        var newDimsArray = splitDimensions(t.dimensions, multipleValueDims);

        var newQueries = [];
        for (var i = 0; i < newDimsArray.length; i++) {
          var dims = newDimsArray[i];
          var newQuery = _lodash2.default.cloneDeep(t);
          newQuery.dimensions = dims;
          if (i !== 0) {
            newQuery.refId = '' + newQuery.refId + i;
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

  }, {
    key: 'templateMetricQuery',
    value: function templateMetricQuery(varString) {

      var regionQuery = varString.match(_constants.regionsQueryRegex);
      if (regionQuery) {
        return this.getRegions().catch(function (err) {
          throw new Error('Unable to get regions: ' + err);
        });
      }

      var compartmentQuery = varString.match(_constants.compartmentsQueryRegex);
      if (compartmentQuery) {
        return this.getCompartments().then(function (compartments) {
          return compartments.map(function (c) {
            return { text: c.text, value: c.text };
          });
        }).catch(function (err) {
          throw new Error('Unable to get compartments: ' + err);
        });
      }

      var namespaceQuery = varString.match(_constants.namespacesQueryRegex);
      if (namespaceQuery) {
        var target = {
          region: (0, _constants.removeQuotes)(this.getVariableValue(namespaceQuery[1])),
          compartment: (0, _constants.removeQuotes)(this.getVariableValue(namespaceQuery[2]))
        };
        return this.getNamespaces(target).catch(function (err) {
          throw new Error('Unable to get namespaces: ' + err);
        });
      }

      var resourcegroupQuery = varString.match(_constants.resourcegroupsQueryRegex);
      if (resourcegroupQuery) {
        var _target = {
          region: (0, _constants.removeQuotes)(this.getVariableValue(resourcegroupQuery[1])),
          compartment: (0, _constants.removeQuotes)(this.getVariableValue(resourcegroupQuery[2])),
          namespace: (0, _constants.removeQuotes)(this.getVariableValue(resourcegroupQuery[3]))
        };
        return this.getResourceGroups(_target).catch(function (err) {
          throw new Error('Unable to get resourcegroups: ' + err);
        });
      }

      var metricQuery = varString.match(_constants.metricsQueryRegex);
      if (metricQuery) {
        var _target2 = {
          region: (0, _constants.removeQuotes)(this.getVariableValue(metricQuery[1])),
          compartment: (0, _constants.removeQuotes)(this.getVariableValue(metricQuery[2])),
          namespace: (0, _constants.removeQuotes)(this.getVariableValue(metricQuery[3])),
          resourcegroup: (0, _constants.removeQuotes)(this.getVariableValue(metricQuery[4]))
        };
        return this.metricFindQuery(_target2).catch(function (err) {
          throw new Error('Unable to get metrics: ' + err);
        });
      }

      var dimensionsQuery = varString.match(_constants.dimensionKeysQueryRegex);
      if (dimensionsQuery) {
        var _target3 = {
          region: (0, _constants.removeQuotes)(this.getVariableValue(dimensionsQuery[1])),
          compartment: (0, _constants.removeQuotes)(this.getVariableValue(dimensionsQuery[2])),
          namespace: (0, _constants.removeQuotes)(this.getVariableValue(dimensionsQuery[3])),
          metric: (0, _constants.removeQuotes)(this.getVariableValue(dimensionsQuery[4])),
          resourcegroup: (0, _constants.removeQuotes)(this.getVariableValue(dimensionsQuery[5]))
        };
        return this.getDimensionKeys(_target3).catch(function (err) {
          throw new Error('Unable to get dimensions: ' + err);
        });
      }

      var dimensionOptionsQuery = varString.match(_constants.dimensionValuesQueryRegex);
      if (dimensionOptionsQuery) {
        var _target4 = {
          region: (0, _constants.removeQuotes)(this.getVariableValue(dimensionOptionsQuery[1])),
          compartment: (0, _constants.removeQuotes)(this.getVariableValue(dimensionOptionsQuery[2])),
          namespace: (0, _constants.removeQuotes)(this.getVariableValue(dimensionOptionsQuery[3])),
          metric: (0, _constants.removeQuotes)(this.getVariableValue(dimensionOptionsQuery[4])),
          resourcegroup: (0, _constants.removeQuotes)(this.getVariableValue(dimensionOptionsQuery[6]))
        };
        var dimensionKey = (0, _constants.removeQuotes)(this.getVariableValue(dimensionOptionsQuery[5]));
        return this.getDimensionValues(_target4, dimensionKey).catch(function (err) {
          throw new Error('Unable to get dimension options: ' + err);
        });
      }

      throw new Error('Unable to parse templating string');
    }
  }, {
    key: 'getRegions',
    value: function getRegions() {
      var _this5 = this;

      if (this.regionsCache && this.regionsCache.length > 0) {
        return this.q.when(this.regionsCache);
      }

      return this.doRequest({
        targets: [{
          environment: this.environment,
          datasourceId: this.id,
          tenancyOCID: this.tenancyOCID,
          queryType: 'regions'
        }],
        range: this.timeSrv.timeRange()
      }).then(function (items) {
        _this5.regionsCache = _this5.mapToTextValue(items, 'regions');
        return _this5.regionsCache;
      });
    }
  }, {
    key: 'getCompartments',
    value: function getCompartments() {
      var _this6 = this;

      if (this.compartmentsCache && this.compartmentsCache.length > 0) {
        return this.q.when(this.compartmentsCache);
      }

      return this.doRequest({
        targets: [{
          environment: this.environment,
          datasourceId: this.id,
          tenancyOCID: this.tenancyOCID,
          queryType: 'compartments',
          region: this.defaultRegion // compartments are registered for the all regions, so no difference which region to use here
        }],
        range: this.timeSrv.timeRange()
      }).then(function (items) {
        _this6.compartmentsCache = _this6.mapToTextValue(items, 'compartments');
        return _this6.compartmentsCache;
      });
    }
  }, {
    key: 'getCompartmentId',
    value: function getCompartmentId(compartment) {
      return this.getCompartments().then(function (compartments) {
        var compartmentFound = compartments.find(function (c) {
          return c.text === compartment || c.value === compartment;
        });
        return compartmentFound ? compartmentFound.value : compartment;
      });
    }
  }, {
    key: 'getNamespaces',
    value: async function getNamespaces(target) {
      var _this7 = this;

      var region = target.region === _query_ctrl.SELECT_PLACEHOLDERS.REGION ? '' : this.getVariableValue(target.region);
      var compartment = target.compartment === _query_ctrl.SELECT_PLACEHOLDERS.COMPARTMENT ? '' : this.getVariableValue(target.compartment);
      if (_lodash2.default.isEmpty(compartment)) {
        return this.q.when([]);
      }

      var compartmentId = await this.getCompartmentId(compartment);
      return this.doRequest({
        targets: [{
          environment: this.environment,
          datasourceId: this.id,
          tenancyOCID: this.tenancyOCID,
          queryType: 'namespaces',
          region: _lodash2.default.isEmpty(region) ? this.defaultRegion : region,
          compartment: compartmentId
        }],
        range: this.timeSrv.timeRange()
      }).then(function (items) {
        return _this7.mapToTextValue(items, 'namespaces');
      });
    }
  }, {
    key: 'getResourceGroups',
    value: async function getResourceGroups(target) {
      var _this8 = this;

      var region = target.region === _query_ctrl.SELECT_PLACEHOLDERS.REGION ? '' : this.getVariableValue(target.region);
      var compartment = target.compartment === _query_ctrl.SELECT_PLACEHOLDERS.COMPARTMENT ? '' : this.getVariableValue(target.compartment);
      var namespace = target.namespace === _query_ctrl.SELECT_PLACEHOLDERS.NAMESPACE ? '' : this.getVariableValue(target.namespace);
      if (_lodash2.default.isEmpty(compartment)) {
        return this.q.when([]);
      }

      var compartmentId = await this.getCompartmentId(compartment);
      return this.doRequest({
        targets: [{
          environment: this.environment,
          datasourceId: this.id,
          tenancyOCID: this.tenancyOCID,
          queryType: 'resourcegroups',
          region: _lodash2.default.isEmpty(region) ? this.defaultRegion : region,
          compartment: compartmentId,
          namespace: namespace
        }],
        range: this.timeSrv.timeRange()
      }).then(function (items) {
        return _this8.mapToTextValue(items, 'resourcegroups');
      });
    }
  }, {
    key: 'getDimensions',
    value: async function getDimensions(target) {
      var _this9 = this;

      var region = target.region === _query_ctrl.SELECT_PLACEHOLDERS.REGION ? '' : this.getVariableValue(target.region);
      var compartment = target.compartment === _query_ctrl.SELECT_PLACEHOLDERS.COMPARTMENT ? '' : this.getVariableValue(target.compartment);
      var namespace = target.namespace === _query_ctrl.SELECT_PLACEHOLDERS.NAMESPACE ? '' : this.getVariableValue(target.namespace);
      var resourcegroup = target.resourcegroup === _query_ctrl.SELECT_PLACEHOLDERS.RESOURCEGROUP ? DEFAULT_RESOURCE_GROUP : this.getVariableValue(target.resourcegroup);
      var metric = target.metric === _query_ctrl.SELECT_PLACEHOLDERS.METRIC ? '' : this.getVariableValue(target.metric);
      var metrics = metric.startsWith("{") && metric.endsWith("}") ? metric.slice(1, metric.length - 1).split(',') : [metric];

      if (_lodash2.default.isEmpty(compartment) || _lodash2.default.isEmpty(namespace) || _lodash2.default.isEmpty(metrics)) {
        return this.q.when([]);
      }

      var dimensionsMap = {};
      var _iteratorNormalCompletion4 = true;
      var _didIteratorError4 = false;
      var _iteratorError4 = undefined;

      try {
        var _loop = async function _loop() {
          var m = _step4.value;

          if (dimensionsMap[m] !== undefined) {
            return 'continue';
          }
          dimensionsMap[m] = null;

          var compartmentId = await _this9.getCompartmentId(compartment);
          await _this9.doRequest({
            targets: [{
              environment: _this9.environment,
              datasourceId: _this9.id,
              tenancyOCID: _this9.tenancyOCID,
              queryType: 'dimensions',
              region: _lodash2.default.isEmpty(region) ? _this9.defaultRegion : region,
              compartment: compartmentId,
              namespace: namespace,
              resourcegroup: resourcegroup,
              metric: m
            }],
            range: _this9.timeSrv.timeRange()
          }).then(function (result) {
            var items = _this9.mapToTextValue(result, 'dimensions');
            dimensionsMap[m] = [].concat(items);
          }).finally(function () {
            if (!dimensionsMap[m]) {
              dimensionsMap[m] = [];
            }
          });
        };

        for (var _iterator4 = metrics[Symbol.iterator](), _step4; !(_iteratorNormalCompletion4 = (_step4 = _iterator4.next()).done); _iteratorNormalCompletion4 = true) {
          var _ret = await _loop();

          if (_ret === 'continue') continue;
        }
      } catch (err) {
        _didIteratorError4 = true;
        _iteratorError4 = err;
      } finally {
        try {
          if (!_iteratorNormalCompletion4 && _iterator4.return) {
            _iterator4.return();
          }
        } finally {
          if (_didIteratorError4) {
            throw _iteratorError4;
          }
        }
      }

      var result = [];
      Object.values(dimensionsMap).forEach(function (dims) {
        if (_lodash2.default.isEmpty(result)) {
          result = dims;
        } else {
          var newResult = [];
          dims.forEach(function (dim) {
            if (!!result.find(function (d) {
              return d.value === dim.value;
            }) && !newResult.find(function (d) {
              return d.value === dim.value;
            })) {
              newResult.push(dim);
            }
          });
          result = newResult;
        }
      });

      return result;
    }
  }, {
    key: 'getDimensionKeys',
    value: function getDimensionKeys(target) {
      return this.getDimensions(target).then(function (dims) {
        var dimCache = dims.reduce(function (data, item) {
          var values = item.value.split('=') || [];
          var key = values[0] || item.value;
          var value = values[1];

          if (!data[key]) {
            data[key] = [];
          }
          data[key].push(value);
          return data;
        }, {});
        return Object.keys(dimCache);
      }).then(function (items) {
        return items.map(function (item) {
          return { text: item, value: item };
        });
      });
    }
  }, {
    key: 'getDimensionValues',
    value: function getDimensionValues(target, dimKey) {
      var _this10 = this;

      return this.getDimensions(target).then(function (dims) {
        var dimCache = dims.reduce(function (data, item) {
          var values = item.value.split('=') || [];
          var key = values[0] || item.value;
          var value = values[1];

          if (!data[key]) {
            data[key] = [];
          }
          data[key].push(value);
          return data;
        }, {});
        return dimCache[_this10.getVariableValue(dimKey)] || [];
      }).then(function (items) {
        return items.map(function (item) {
          return { text: item, value: item };
        });
      });
    }
  }, {
    key: 'getAggregations',
    value: function getAggregations() {
      return this.q.when(_constants.aggregations);
    }

    /**
     * Calls grafana backend.
     * Retries 10 times before failure.
     */

  }, {
    key: 'doRequest',
    value: function doRequest(options) {
      var _this = this;
      return (0, _retry2.default)(function () {
        return _this.backendSrv.datasourceRequest({
          url: '/api/tsdb/query',
          method: 'POST',
          data: {
            from: options.range.from.valueOf().toString(),
            to: options.range.to.valueOf().toString(),
            queries: options.targets
          }
        });
      }, 10);
    }

    /**
     * Converts data from grafana backend to UI format
     */

  }, {
    key: 'mapToTextValue',
    value: function mapToTextValue(result, searchField) {
      if (_lodash2.default.isEmpty(result) || _lodash2.default.isEmpty(searchField)) {
        return [];
      }

      var table = result.data.results[searchField].tables[0];
      if (!table) {
        return [];
      }

      var map = _lodash2.default.map(table.rows, function (row, i) {
        if (row.length > 1) {
          return { text: row[0], value: row[1] };
        } else if (_lodash2.default.isObject(row[0])) {
          return { text: row[0], value: i };
        }
        return { text: row[0], value: row[0] };
      });
      return map;
    }

    // **************************** Template variables helpers ****************************

    /**
     * Get all template variable descriptors
     */

  }, {
    key: 'getVariableDescriptors',
    value: function getVariableDescriptors(regex) {
      var includeCustom = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : true;

      var vars = this.templateSrv.variables || [];

      if (regex) {
        var regexVars = vars.filter(function (item) {
          return item.query.match(regex) !== null;
        });
        if (includeCustom) {
          var custom = vars.filter(function (item) {
            return item.type === 'custom' || item.type === 'constant';
          });
          regexVars = regexVars.concat(custom);
        }
        var uniqueRegexVarsMap = new Map();
        regexVars.forEach(function (varObj) {
          return uniqueRegexVarsMap.set(varObj.name, varObj);
        });
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

  }, {
    key: 'getVariables',
    value: function getVariables(regex, includeCustom) {
      var varDescriptors = this.getVariableDescriptors(regex, includeCustom) || [];
      return varDescriptors.map(function (item) {
        return '$' + item.name;
      });
    }

    /**
     * @param varName valid varName contains '$'. Example: '$dimensionKey'
     * Returns an array with variable values or empty array
    */

  }, {
    key: 'getVariableValue',
    value: function getVariableValue(varName) {
      var scopedVars = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : {};

      return this.templateSrv.replace(varName, scopedVars) || varName;
    }

    /**
     * @param varName valid varName contains '$'. Example: '$dimensionKey'
     * Returns true if variable with the given name is found
    */

  }, {
    key: 'isVariable',
    value: function isVariable(varName) {
      var varNames = this.getVariables() || [];
      return !!varNames.find(function (item) {
        return item === varName;
      });
    }
  }]);

  return OCIDatasource;
}();

exports.default = OCIDatasource;
//# sourceMappingURL=datasource.js.map
