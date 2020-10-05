'use strict';

System.register(['app/plugins/sdk', './css/query-editor.css!', './constants'], function (_export, _context) {
  "use strict";

  var QueryCtrl, windows, namespacesQueryRegex, resourcegroupsQueryRegex, metricsQueryRegex, regionsQueryRegex, compartmentsQueryRegex, dimensionKeysQueryRegex, dimensionValuesQueryRegex, windowsAndResolutionRegex, resolutions, AUTO, _createClass, SELECT_PLACEHOLDERS, OCIDatasourceQueryCtrl;

  function _toConsumableArray(arr) {
    if (Array.isArray(arr)) {
      for (var i = 0, arr2 = Array(arr.length); i < arr.length; i++) {
        arr2[i] = arr[i];
      }

      return arr2;
    } else {
      return Array.from(arr);
    }
  }

  function _classCallCheck(instance, Constructor) {
    if (!(instance instanceof Constructor)) {
      throw new TypeError("Cannot call a class as a function");
    }
  }

  function _possibleConstructorReturn(self, call) {
    if (!self) {
      throw new ReferenceError("this hasn't been initialised - super() hasn't been called");
    }

    return call && (typeof call === "object" || typeof call === "function") ? call : self;
  }

  function _inherits(subClass, superClass) {
    if (typeof superClass !== "function" && superClass !== null) {
      throw new TypeError("Super expression must either be null or a function, not " + typeof superClass);
    }

    subClass.prototype = Object.create(superClass && superClass.prototype, {
      constructor: {
        value: subClass,
        enumerable: false,
        writable: true,
        configurable: true
      }
    });
    if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass;
  }

  return {
    setters: [function (_appPluginsSdk) {
      QueryCtrl = _appPluginsSdk.QueryCtrl;
    }, function (_cssQueryEditorCss) {}, function (_constants) {
      windows = _constants.windows;
      namespacesQueryRegex = _constants.namespacesQueryRegex;
      resourcegroupsQueryRegex = _constants.resourcegroupsQueryRegex;
      metricsQueryRegex = _constants.metricsQueryRegex;
      regionsQueryRegex = _constants.regionsQueryRegex;
      compartmentsQueryRegex = _constants.compartmentsQueryRegex;
      dimensionKeysQueryRegex = _constants.dimensionKeysQueryRegex;
      dimensionValuesQueryRegex = _constants.dimensionValuesQueryRegex;
      windowsAndResolutionRegex = _constants.windowsAndResolutionRegex;
      resolutions = _constants.resolutions;
      AUTO = _constants.AUTO;
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

      _export('SELECT_PLACEHOLDERS', SELECT_PLACEHOLDERS = {
        DIMENSION_KEY: 'select dimension',
        DIMENSION_VALUE: 'select value',
        COMPARTMENT: 'select compartment',
        REGION: 'select region',
        NAMESPACE: 'select namespace',
        RESOURCEGROUP: 'select resource group',
        METRIC: 'select metric',
        WINDOW: 'select window'
      });

      _export('SELECT_PLACEHOLDERS', SELECT_PLACEHOLDERS);

      _export('OCIDatasourceQueryCtrl', OCIDatasourceQueryCtrl = function (_QueryCtrl) {
        _inherits(OCIDatasourceQueryCtrl, _QueryCtrl);

        function OCIDatasourceQueryCtrl($scope, $injector, $q, uiSegmentSrv) {
          _classCallCheck(this, OCIDatasourceQueryCtrl);

          var _this = _possibleConstructorReturn(this, (OCIDatasourceQueryCtrl.__proto__ || Object.getPrototypeOf(OCIDatasourceQueryCtrl)).call(this, $scope, $injector));

          _this.q = $q;
          _this.uiSegmentSrv = uiSegmentSrv;

          _this.target.region = _this.target.region || SELECT_PLACEHOLDERS.REGION;
          _this.target.compartment = _this.target.compartment || SELECT_PLACEHOLDERS.COMPARTMENT;
          _this.target.namespace = _this.target.namespace || SELECT_PLACEHOLDERS.NAMESPACE;
          _this.target.resourcegroup = _this.target.resourcegroup || SELECT_PLACEHOLDERS.RESOURCEGROUP;
          _this.target.metric = _this.target.metric || SELECT_PLACEHOLDERS.METRIC;
          _this.target.resolution = _this.target.resolution || AUTO;
          _this.target.window = _this.target.window || AUTO;
          _this.target.aggregation = _this.target.aggregation || 'mean()';
          _this.target.dimensions = _this.target.dimensions || [];

          _this.dimensionSegments = [];
          _this.removeDimensionSegment = uiSegmentSrv.newSegment({ fake: true, value: '-- remove dimension --' });
          _this.getSelectDimensionKeySegment = function () {
            return uiSegmentSrv.newSegment({ value: SELECT_PLACEHOLDERS.DIMENSION_KEY, type: 'key' });
          };
          _this.getDimensionOperatorSegment = function () {
            return _this.uiSegmentSrv.newOperator('=');
          };
          _this.getSelectDimensionValueSegment = function () {
            return uiSegmentSrv.newSegment({ value: SELECT_PLACEHOLDERS.DIMENSION_VALUE, type: 'value' });
          };

          _this.dimensionsCache = {};

          // rebuild dimensionSegments on query editor load
          for (var i = 0; i < _this.target.dimensions.length; i++) {
            var dim = _this.target.dimensions[i];
            if (i > 0) {
              _this.dimensionSegments.push(_this.uiSegmentSrv.newCondition(','));
            }
            _this.dimensionSegments.push(_this.uiSegmentSrv.newSegment({ value: dim.key, type: 'key' }));
            _this.dimensionSegments.push(_this.uiSegmentSrv.newSegment({ value: dim.operator, type: 'operator' }));
            _this.dimensionSegments.push(_this.uiSegmentSrv.newSegment({ value: dim.value, type: 'value' }));
          }
          _this.dimensionSegments.push(_this.uiSegmentSrv.newPlusButton());
          return _this;
        }

        // ****************************** Options **********************************

        _createClass(OCIDatasourceQueryCtrl, [{
          key: 'getRegions',
          value: function getRegions() {
            var _this2 = this;

            return this.datasource.getRegions().then(function (regions) {
              return _this2.appendVariables([].concat(_toConsumableArray(regions)), regionsQueryRegex);
            });
          }
        }, {
          key: 'getCompartments',
          value: function getCompartments() {
            var _this3 = this;

            return this.datasource.getCompartments().then(function (compartments) {
              return _this3.appendVariables([].concat(_toConsumableArray(compartments)), compartmentsQueryRegex);
            });
          }
        }, {
          key: 'getNamespaces',
          value: function getNamespaces() {
            var _this4 = this;

            return this.datasource.getNamespaces(this.target).then(function (namespaces) {
              return _this4.appendVariables([].concat(_toConsumableArray(namespaces)), namespacesQueryRegex);
            });
          }
        }, {
          key: 'getResourceGroups',
          value: function getResourceGroups() {
            var _this5 = this;

            return this.datasource.getResourceGroups(this.target).then(function (resourcegroups) {
              return _this5.appendVariables([].concat(_toConsumableArray(resourcegroups)), resourcegroupsQueryRegex);
            });
          }
        }, {
          key: 'getMetrics',
          value: function getMetrics() {
            var _this6 = this;

            return this.datasource.metricFindQuery(this.target).then(function (metrics) {
              return _this6.appendVariables([].concat(_toConsumableArray(metrics)), metricsQueryRegex);
            });
          }
        }, {
          key: 'getAggregations',
          value: function getAggregations() {
            return this.datasource.getAggregations().then(function (aggs) {
              return aggs.map(function (val) {
                return { text: val, value: val };
              });
            });
          }
        }, {
          key: 'getWindows',
          value: function getWindows() {
            return this.appendWindowsAndResolutionVariables([].concat(_toConsumableArray(windows)), windowsAndResolutionRegex);
          }
        }, {
          key: 'getResolutions',
          value: function getResolutions() {
            return this.appendWindowsAndResolutionVariables([].concat(_toConsumableArray(resolutions)), windowsAndResolutionRegex);
          }
        }, {
          key: 'getDimensionOptions',
          value: function getDimensionOptions(segment, index) {
            var _this7 = this;

            if (segment.type === 'key' || segment.type === 'plus-button') {
              return this.getDimensionsCache().then(function (cache) {
                var keys = Object.keys(cache);
                var vars = _this7.datasource.getVariables(dimensionKeysQueryRegex) || [];
                var keysWithVariables = vars.concat(keys);
                var segments = keysWithVariables.map(function (key) {
                  return _this7.uiSegmentSrv.newSegment({ value: key });
                });
                segments.unshift(_this7.removeDimensionSegment);
                return segments;
              });
            }

            if (segment.type === 'value') {
              return this.getDimensionsCache().then(function (cache) {
                var keySegment = _this7.dimensionSegments[index - 2];
                var key = _this7.datasource.getVariableValue(keySegment.value);
                var options = cache[key] || [];

                // return all the values for the key
                var vars = _this7.datasource.getVariables(dimensionValuesQueryRegex) || [];
                var optionsWithVariables = vars.concat(options);
                var segments = optionsWithVariables.map(function (v) {
                  return _this7.uiSegmentSrv.newSegment({ value: v });
                });
                return segments;
              });
            }

            return this.q.when([]);
          }
        }, {
          key: 'getDimensionsCache',
          value: function getDimensionsCache() {
            var _this8 = this;

            var targetSelector = JSON.stringify({
              region: this.datasource.getVariableValue(this.target.region),
              compartment: this.datasource.getVariableValue(this.target.compartment),
              namespace: this.datasource.getVariableValue(this.target.namespace),
              resourcegroup: this.datasource.getVariableValue(this.target.resourcegroup),
              metric: this.datasource.getVariableValue(this.target.metric)
            });

            if (this.dimensionsCache[targetSelector]) {
              return this.q.when(this.dimensionsCache[targetSelector]);
            }

            return this.datasource.getDimensions(this.target).then(function (dimensions) {
              var cache = dimensions.reduce(function (data, item) {
                var values = item.value.split('=') || [];
                var key = values[0] || item.value;
                var value = values[1];

                if (!data[key]) {
                  data[key] = [];
                }
                data[key].push(value);
                return data;
              }, {});
              _this8.dimensionsCache[targetSelector] = cache;
              return _this8.dimensionsCache[targetSelector];
            });
          }
        }, {
          key: 'appendVariables',
          value: function appendVariables(options, varQeueryRegex) {
            var vars = this.datasource.getVariables(varQeueryRegex) || [];
            vars.forEach(function (value) {
              options.unshift({ value: value, text: value });
            });
            return options;
          }
        }, {
          key: 'appendWindowsAndResolutionVariables',
          value: function appendWindowsAndResolutionVariables(options, varQeueryRegex) {
            var vars = this.datasource.getVariables(varQeueryRegex) || [];
            return [].concat(_toConsumableArray(options), _toConsumableArray(vars)).map(function (value) {
              return { value: value, text: value };
            });
          }
        }, {
          key: 'toggleEditorMode',
          value: function toggleEditorMode() {
            this.target.rawQuery = !this.target.rawQuery;
          }
        }, {
          key: 'onChangeInternal',
          value: function onChangeInternal() {
            this.panelCtrl.refresh(); // Asks the panel to refresh data.
          }
        }, {
          key: 'onDimensionsChange',
          value: function onDimensionsChange(segment, index) {
            var _this9 = this;

            if (segment.value === this.removeDimensionSegment.value) {
              // remove dimension: key - op - value
              this.dimensionSegments.splice(index, 3);
              // remove last comma
              if (this.dimensionSegments.length > 2) {
                this.dimensionSegments.splice(Math.max(index - 1, 0), 1);
              }
            } else if (segment.type === 'plus-button') {
              if (index > 2) {
                // add comma in front of plus button
                this.dimensionSegments.splice(index, 0, this.uiSegmentSrv.newCondition(','));
              }
              // replace plus button with key segment
              segment.type = 'key';
              segment.cssClass = 'query-segment-key';
              this.dimensionSegments.push(this.getDimensionOperatorSegment());
              this.dimensionSegments.push(this.getSelectDimensionValueSegment());
            } else if (segment.type === 'key') {
              this.getDimensionsCache().then(function (cache) {
                //update value to be part of the available options
                var value = _this9.dimensionSegments[index + 2].value;
                var options = cache[segment.value] || [];
                if (!_this9.datasource.isVariable(value) && options.indexOf(value) < 0) {
                  _this9.dimensionSegments[index + 2] = _this9.getSelectDimensionValueSegment();
                }

                _this9.updateQueryWithDimensions();
              });
            }

            // add plus button at the end
            if (this.dimensionSegments.length === 0 || this.dimensionSegments[this.dimensionSegments.length - 1].type !== 'plus-button') {
              this.dimensionSegments.push(this.uiSegmentSrv.newPlusButton());
            }

            this.updateQueryWithDimensions();
          }
        }, {
          key: 'updateQueryWithDimensions',
          value: function updateQueryWithDimensions() {
            var dimensions = [];
            var index = void 0;

            this.dimensionSegments.forEach(function (s) {
              if (s.type === 'key') {
                if (dimensions.length === 0) {
                  dimensions.push({});
                  index = 0;
                }
                dimensions[index].key = s.value;
              } else if (s.type === 'value') {
                dimensions[index].value = s.value;
              } else if (s.type === 'condition') {
                dimensions.push({});
                index++;
              } else if (s.type === 'operator') {
                dimensions[index].operator = s.value;
              }
            });

            this.target.dimensions = dimensions;
            this.panelCtrl.refresh();
          }
        }]);

        return OCIDatasourceQueryCtrl;
      }(QueryCtrl));

      _export('OCIDatasourceQueryCtrl', OCIDatasourceQueryCtrl);

      OCIDatasourceQueryCtrl.templateUrl = 'partials/query.editor.html';
    }
  };
});
//# sourceMappingURL=query_ctrl.js.map
