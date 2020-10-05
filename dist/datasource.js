'use strict';

System.register(['lodash', './constants', './util/retry', './query_ctrl', './util/utilFunctions'], function (_export, _context) {
  "use strict";

  var _, aggregations, dimensionKeysQueryRegex, namespacesQueryRegex, resourcegroupsQueryRegex, metricsQueryRegex, regionsQueryRegex, compartmentsQueryRegex, dimensionValuesQueryRegex, removeQuotes, AUTO, retryOrThrow, SELECT_PLACEHOLDERS, resolveAutoWinRes, _createClass, DEFAULT_RESOURCE_GROUP, OCIDatasource;

  function _classCallCheck(instance, Constructor) {
    if (!(instance instanceof Constructor)) {
      throw new TypeError("Cannot call a class as a function");
    }
  }

  return {
    setters: [function (_lodash) {
      _ = _lodash.default;
    }, function (_constants) {
      aggregations = _constants.aggregations;
      dimensionKeysQueryRegex = _constants.dimensionKeysQueryRegex;
      namespacesQueryRegex = _constants.namespacesQueryRegex;
      resourcegroupsQueryRegex = _constants.resourcegroupsQueryRegex;
      metricsQueryRegex = _constants.metricsQueryRegex;
      regionsQueryRegex = _constants.regionsQueryRegex;
      compartmentsQueryRegex = _constants.compartmentsQueryRegex;
      dimensionValuesQueryRegex = _constants.dimensionValuesQueryRegex;
      removeQuotes = _constants.removeQuotes;
      AUTO = _constants.AUTO;
    }, function (_utilRetry) {
      retryOrThrow = _utilRetry.default;
    }, function (_query_ctrl) {
      SELECT_PLACEHOLDERS = _query_ctrl.SELECT_PLACEHOLDERS;
    }, function (_utilUtilFunctions) {
      resolveAutoWinRes = _utilUtilFunctions.resolveAutoWinRes;
    }],
    execute: function () {
      _createClass = function () {
        function defineProperties(target, props) {
          for (var i = 0; i < props.length; i++) {
            var descriptor = props[i];
            descriptor.enumerable = descriptor.enumerable || false;
            descriptor.configurable = true;
            if ("value" in descriptor) descriptor.writable = true;
            Object.defineProperty(target, descriptor.key, descriptor);
          }
        }

        return function (Constructor, protoProps, staticProps) {
          if (protoProps) defineProperties(Constructor.prototype, protoProps);
          if (staticProps) defineProperties(Constructor, staticProps);
          return Constructor;
        };
      }();

      DEFAULT_RESOURCE_GROUP = 'NoResourceGroup';

      OCIDatasource = function () {
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
              _.forEach(result.data.results, function (r) {
                _.forEach(r.series, function (s) {
                  res.push({ target: s.name, datapoints: s.points });
                });
                _.forEach(r.tables, function (t) {
                  t.type = 'table';
                  t.refId = r.refId;
                  res.push(t);
                });
              });

              result.data = res;
              return result;
            });
          }
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
        }, {
          key: 'metricFindQuery',
          value: async function metricFindQuery(target) {
            var _this2 = this;

            if (typeof target === 'string') {
              // used in template editor for creating variables
              return this.templateMetricQuery(target);
            }
            var region = target.region === SELECT_PLACEHOLDERS.REGION ? '' : this.getVariableValue(target.region);
            var compartment = target.compartment === SELECT_PLACEHOLDERS.COMPARTMENT ? '' : this.getVariableValue(target.compartment);
            var namespace = target.namespace === SELECT_PLACEHOLDERS.NAMESPACE ? '' : this.getVariableValue(target.namespace);
            var resourcegroup = target.resourcegroup === SELECT_PLACEHOLDERS.RESOURCEGROUP ? DEFAULT_RESOURCE_GROUP : this.getVariableValue(target.resourcegroup);

            if (_.isEmpty(compartment) || _.isEmpty(namespace)) {
              return this.q.when([]);
            }

            var compartmentId = await this.getCompartmentId(compartment);
            return this.doRequest({
              targets: [{
                environment: this.environment,
                datasourceId: this.id,
                tenancyOCID: this.tenancyOCID,
                queryType: 'search',
                region: _.isEmpty(region) ? this.defaultRegion : region,
                compartment: compartmentId,
                namespace: namespace,
                resourcegroup: resourcegroup
              }],
              range: this.timeSrv.timeRange()
            }).then(function (res) {
              return _this2.mapToTextValue(res, 'search');
            });
          }
        }, {
          key: 'buildQueryParameters',
          value: async function buildQueryParameters(options) {
            var _this3 = this;

            var queries = options.targets.filter(function (t) {
              return !t.hide;
            }).filter(function (t) {
              return !_.isEmpty(_this3.getVariableValue(t.compartment, options.scopedVars)) && t.compartment !== SELECT_PLACEHOLDERS.COMPARTMENT;
            }).filter(function (t) {
              return !_.isEmpty(_this3.getVariableValue(t.namespace, options.scopedVars)) && t.namespace !== SELECT_PLACEHOLDERS.NAMESPACE;
            }).filter(function (t) {
              return !_.isEmpty(_this3.getVariableValue(t.resourcegroup, options.scopedVars));
            }).filter(function (t) {
              return !_.isEmpty(_this3.getVariableValue(t.metric, options.scopedVars)) && t.metric !== SELECT_PLACEHOLDERS.METRIC || !_.isEmpty(_this3.getVariableValue(t.target));
            });

            queries.forEach(function (t) {
              t.dimensions = (t.dimensions || []).filter(function (dim) {
                return !_.isEmpty(dim.key) && dim.key !== SELECT_PLACEHOLDERS.DIMENSION_KEY;
              }).filter(function (dim) {
                return !_.isEmpty(dim.value) && dim.value !== SELECT_PLACEHOLDERS.DIMENSION_VALUE;
              });

              t.resourcegroup = t.resourcegroup === SELECT_PLACEHOLDERS.RESOURCEGROUP ? DEFAULT_RESOURCE_GROUP : t.resourcegroup;
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

                var region = t.region === SELECT_PLACEHOLDERS.REGION ? '' : this.getVariableValue(t.region, options.scopedVars);
                var query = this.getVariableValue(t.target, options.scopedVars);
                var numberOfDaysDiff = this.timeSrv.timeRange().to.diff(this.timeSrv.timeRange().from, 'days');
                // The following replaces 'auto' in window portion of the query and replaces it with an appropriate value.
                // If there is a functionality to access the window variable instead of matching [auto] in the query, it will be
                // better
                if (query) query = query.replace('[auto]', '[' + resolveAutoWinRes(AUTO, '', numberOfDaysDiff).window + ']');
                var resolution = this.getVariableValue(t.resolution, options.scopedVars);
                var window = t.window === SELECT_PLACEHOLDERS.WINDOW ? '' : this.getVariableValue(t.window, options.scopedVars);
                // p.s : timeSrv.timeRange() results in a moment object
                var resolvedWinResolObj = resolveAutoWinRes(window, resolution, numberOfDaysDiff);
                window = resolvedWinResolObj.window;
                resolution = resolvedWinResolObj.resolution;
                if (_.isEmpty(query)) {
                  // construct query
                  var dimensions = (t.dimensions || []).reduce(function (result, dim) {
                    var d = _this3.getVariableValue(dim.key, options.scopedVars) + ' ' + dim.operator + ' "' + _this3.getVariableValue(dim.value, options.scopedVars) + '"';
                    if (result.indexOf(d) < 0) {
                      result.push(d);
                    }
                    return result;
                  }, []);
                  var dimension = _.isEmpty(dimensions) ? '' : '{' + dimensions.join(',') + '}';
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
                  region: _.isEmpty(region) ? this.defaultRegion : region,
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
        }, {
          key: 'splitMultiValueDimensionsIntoQuieries',
          value: function splitMultiValueDimensionsIntoQuieries(queries, options) {
            var _this4 = this;

            return queries.reduce(function (data, t) {

              if (_.isEmpty(t.dimensions) || !_.isEmpty(t.target)) {
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

              if (_.isEmpty(Object.keys(multipleValueDims))) {
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

                    var _newDim = _.cloneDeep(dims[0]);
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
                        var newDim = _.cloneDeep(dims[i]);
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
                var newQuery = _.cloneDeep(t);
                newQuery.dimensions = dims;
                if (i !== 0) {
                  newQuery.refId = '' + newQuery.refId + i;
                }
                newQueries.push(newQuery);
              }
              return data.concat(newQueries);
            }, []);
          }
        }, {
          key: 'templateMetricQuery',
          value: function templateMetricQuery(varString) {

            var regionQuery = varString.match(regionsQueryRegex);
            if (regionQuery) {
              return this.getRegions().catch(function (err) {
                throw new Error('Unable to get regions: ' + err);
              });
            }

            var compartmentQuery = varString.match(compartmentsQueryRegex);
            if (compartmentQuery) {
              return this.getCompartments().then(function (compartments) {
                return compartments.map(function (c) {
                  return { text: c.text, value: c.text };
                });
              }).catch(function (err) {
                throw new Error('Unable to get compartments: ' + err);
              });
            }

            var namespaceQuery = varString.match(namespacesQueryRegex);
            if (namespaceQuery) {
              var target = {
                region: removeQuotes(this.getVariableValue(namespaceQuery[1])),
                compartment: removeQuotes(this.getVariableValue(namespaceQuery[2]))
              };
              return this.getNamespaces(target).catch(function (err) {
                throw new Error('Unable to get namespaces: ' + err);
              });
            }

            var resourcegroupQuery = varString.match(resourcegroupsQueryRegex);
            if (resourcegroupQuery) {
              var _target = {
                region: removeQuotes(this.getVariableValue(resourcegroupQuery[1])),
                compartment: removeQuotes(this.getVariableValue(resourcegroupQuery[2])),
                namespace: removeQuotes(this.getVariableValue(resourcegroupQuery[3]))
              };
              return this.getResourceGroups(_target).catch(function (err) {
                throw new Error('Unable to get resourcegroups: ' + err);
              });
            }

            var metricQuery = varString.match(metricsQueryRegex);
            if (metricQuery) {
              var _target2 = {
                region: removeQuotes(this.getVariableValue(metricQuery[1])),
                compartment: removeQuotes(this.getVariableValue(metricQuery[2])),
                namespace: removeQuotes(this.getVariableValue(metricQuery[3])),
                resourcegroup: removeQuotes(this.getVariableValue(metricQuery[4]))
              };
              return this.metricFindQuery(_target2).catch(function (err) {
                throw new Error('Unable to get metrics: ' + err);
              });
            }

            var dimensionsQuery = varString.match(dimensionKeysQueryRegex);
            if (dimensionsQuery) {
              var _target3 = {
                region: removeQuotes(this.getVariableValue(dimensionsQuery[1])),
                compartment: removeQuotes(this.getVariableValue(dimensionsQuery[2])),
                namespace: removeQuotes(this.getVariableValue(dimensionsQuery[3])),
                metric: removeQuotes(this.getVariableValue(dimensionsQuery[4])),
                resourcegroup: removeQuotes(this.getVariableValue(dimensionsQuery[5]))
              };
              return this.getDimensionKeys(_target3).catch(function (err) {
                throw new Error('Unable to get dimensions: ' + err);
              });
            }

            var dimensionOptionsQuery = varString.match(dimensionValuesQueryRegex);
            if (dimensionOptionsQuery) {
              var _target4 = {
                region: removeQuotes(this.getVariableValue(dimensionOptionsQuery[1])),
                compartment: removeQuotes(this.getVariableValue(dimensionOptionsQuery[2])),
                namespace: removeQuotes(this.getVariableValue(dimensionOptionsQuery[3])),
                metric: removeQuotes(this.getVariableValue(dimensionOptionsQuery[4])),
                resourcegroup: removeQuotes(this.getVariableValue(dimensionOptionsQuery[6]))
              };
              var dimensionKey = removeQuotes(this.getVariableValue(dimensionOptionsQuery[5]));
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

            var region = target.region === SELECT_PLACEHOLDERS.REGION ? '' : this.getVariableValue(target.region);
            var compartment = target.compartment === SELECT_PLACEHOLDERS.COMPARTMENT ? '' : this.getVariableValue(target.compartment);
            if (_.isEmpty(compartment)) {
              return this.q.when([]);
            }

            var compartmentId = await this.getCompartmentId(compartment);
            return this.doRequest({
              targets: [{
                environment: this.environment,
                datasourceId: this.id,
                tenancyOCID: this.tenancyOCID,
                queryType: 'namespaces',
                region: _.isEmpty(region) ? this.defaultRegion : region,
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

            var region = target.region === SELECT_PLACEHOLDERS.REGION ? '' : this.getVariableValue(target.region);
            var compartment = target.compartment === SELECT_PLACEHOLDERS.COMPARTMENT ? '' : this.getVariableValue(target.compartment);
            var namespace = target.namespace === SELECT_PLACEHOLDERS.NAMESPACE ? '' : this.getVariableValue(target.namespace);
            if (_.isEmpty(compartment)) {
              return this.q.when([]);
            }

            var compartmentId = await this.getCompartmentId(compartment);
            return this.doRequest({
              targets: [{
                environment: this.environment,
                datasourceId: this.id,
                tenancyOCID: this.tenancyOCID,
                queryType: 'resourcegroups',
                region: _.isEmpty(region) ? this.defaultRegion : region,
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

            var region = target.region === SELECT_PLACEHOLDERS.REGION ? '' : this.getVariableValue(target.region);
            var compartment = target.compartment === SELECT_PLACEHOLDERS.COMPARTMENT ? '' : this.getVariableValue(target.compartment);
            var namespace = target.namespace === SELECT_PLACEHOLDERS.NAMESPACE ? '' : this.getVariableValue(target.namespace);
            var resourcegroup = target.resourcegroup === SELECT_PLACEHOLDERS.RESOURCEGROUP ? DEFAULT_RESOURCE_GROUP : this.getVariableValue(target.resourcegroup);
            var metric = target.metric === SELECT_PLACEHOLDERS.METRIC ? '' : this.getVariableValue(target.metric);
            var metrics = metric.startsWith("{") && metric.endsWith("}") ? metric.slice(1, metric.length - 1).split(',') : [metric];

            if (_.isEmpty(compartment) || _.isEmpty(namespace) || _.isEmpty(metrics)) {
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
                    region: _.isEmpty(region) ? _this9.defaultRegion : region,
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
              if (_.isEmpty(result)) {
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
            return this.q.when(aggregations);
          }
        }, {
          key: 'doRequest',
          value: function doRequest(options) {
            var _this = this;
            return retryOrThrow(function () {
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
        }, {
          key: 'mapToTextValue',
          value: function mapToTextValue(result, searchField) {
            if (_.isEmpty(result) || _.isEmpty(searchField)) {
              return [];
            }

            var table = result.data.results[searchField].tables[0];
            if (!table) {
              return [];
            }

            var map = _.map(table.rows, function (row, i) {
              if (row.length > 1) {
                return { text: row[0], value: row[1] };
              } else if (_.isObject(row[0])) {
                return { text: row[0], value: i };
              }
              return { text: row[0], value: row[0] };
            });
            return map;
          }
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
        }, {
          key: 'getVariables',
          value: function getVariables(regex, includeCustom) {
            var varDescriptors = this.getVariableDescriptors(regex, includeCustom) || [];
            return varDescriptors.map(function (item) {
              return '$' + item.name;
            });
          }
        }, {
          key: 'getVariableValue',
          value: function getVariableValue(varName) {
            var scopedVars = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : {};

            return this.templateSrv.replace(varName, scopedVars) || varName;
          }
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

      _export('default', OCIDatasource);
    }
  };
});
//# sourceMappingURL=datasource.js.map
