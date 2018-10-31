'use strict';

System.register(['app/plugins/sdk', './css/query-editor.css!', './constants', 'lodash'], function (_export, _context) {
  "use strict";

  var QueryCtrl, regions, aggregations, windows, _, _createClass, OCIDatasourceQueryCtrl;

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
      regions = _constants.regions;
      aggregations = _constants.aggregations;
      windows = _constants.windows;
    }, function (_lodash) {
      _ = _lodash.default;
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

      _export('OCIDatasourceQueryCtrl', OCIDatasourceQueryCtrl = function (_QueryCtrl) {
        _inherits(OCIDatasourceQueryCtrl, _QueryCtrl);

        function OCIDatasourceQueryCtrl($scope, $injector, $q, uiSegmentSrv) {
          _classCallCheck(this, OCIDatasourceQueryCtrl);

          var _this = _possibleConstructorReturn(this, (OCIDatasourceQueryCtrl.__proto__ || Object.getPrototypeOf(OCIDatasourceQueryCtrl)).call(this, $scope, $injector));

          _this.scope = $scope;
          _this.uiSegmentSrv = uiSegmentSrv;
          _this.target.region = _this.target.region || 'select region';
          _this.target.compartment = _this.target.compartment || 'select compartment';
          _this.target.resolution = _this.target.resolution || '1m';
          _this.target.namespace = _this.target.namespace || 'select namespace';
          _this.target.window = _this.target.window || '1m';
          _this.target.metric = _this.target.metric || '';
          _this.target.aggregation = _this.target.aggregation || 'mean()';
          _this.target.tags = _this.target.tags || [];
          _this.q = $q;

          _this.target.dimension = _this.target.dimension || '';

          _this.tagSegments = [];
          _this.dimCache = {};
          _this.removeTagFilterSegment = uiSegmentSrv.newSegment({
            fake: true,
            value: '-- remove tag filter --'
          });

          for (var i = 0; i < _this.target.tags.length; i++) {
            if (i > 0) {
              _this.tagSegments.push(_this.uiSegmentSrv.newCondition(','));
            }
            var obj = _this.target.tags[i];
            _this.tagSegments.push(_this.uiSegmentSrv.newSegment({
              fake: false,
              key: obj.key,
              value: obj.key,
              type: 'key'
            }));
            _this.tagSegments.push(_this.uiSegmentSrv.newSegment({
              fake: false,
              key: obj.operator,
              type: 'operator',
              value: obj.operator
            }));
            _this.tagSegments.push(_this.uiSegmentSrv.newSegment({
              fake: false,
              key: obj.value,
              type: 'value',
              value: obj.value
            }));
          }
          _this.tagSegments.push(_this.uiSegmentSrv.newPlusButton());
          return _this;
        }

        _createClass(OCIDatasourceQueryCtrl, [{
          key: 'toggleEditorMode',
          value: function toggleEditorMode() {
            this.target.rawQuery = !this.target.rawQuery;
          }
        }, {
          key: 'getNamespaces',
          value: function getNamespaces() {
            return this.datasource.getNamespaces(this.target).then(function (namespaces) {
              namespaces.push({ text: '$namespace', value: '$namespace' });
              return namespaces;
            });
          }
        }, {
          key: 'getMetrics',
          value: function getMetrics() {
            return this.datasource.metricFindQuery(this.target).then(function (metrics) {
              metrics.push({ text: '$metric', value: '$metric' });
              return metrics;
            });
          }
        }, {
          key: 'getAggregations',
          value: function getAggregations() {
            return aggregations;
          }
        }, {
          key: 'onChangeInternal',
          value: function onChangeInternal() {
            this.panelCtrl.refresh(); // Asks the panel to refresh data.
          }
        }, {
          key: 'getRegions',
          value: function getRegions() {
            var regs = _.clone(regions);
            regs.push('$region');
            return regs;
          }
        }, {
          key: 'getCompartments',
          value: function getCompartments() {
            return this.datasource.getCompartments().then(function (item) {
              item.push({ text: '$compartment', value: '$compartment' });
              return item;
            });
          }
        }, {
          key: 'getWindows',
          value: function getWindows() {
            return windows;
          }
        }, {
          key: 'getDimensions',
          value: function getDimensions() {
            return this.datasource.getDimensions(this.target);
          }
        }, {
          key: 'handleQueryError',
          value: function handleQueryError(err) {
            this.error = err.message || 'Failed to issue metric query';
            return [];
          }
        }, {
          key: 'getTagsOrValues',
          value: function getTagsOrValues(segment, index) {
            if (segment.type === 'operator') {
              return this.q.when([]);
            }

            if (segment.type === 'key' || segment.type === 'plus-button') {
              return this.getDimensions().then(this.mapToSegment.bind(this)).catch(this.handleQueryError.bind(this));
            }
            var key = this.tagSegments[index - 2];
            var options = this.dimCache[key.value];
            var that = this;
            var optSegments = options.map(function (v) {
              return that.uiSegmentSrv.newSegment({
                value: v
              });
            });
            return this.q.when(optSegments);
          }
        }, {
          key: 'mapToSegment',
          value: function mapToSegment(dimensions) {
            var _this2 = this;

            var dimCache = {};
            var dims = dimensions.map(function (v) {
              var values = v.text.split('=');
              var key = values[0];
              var value = values[1];
              if (!(key in dimCache)) {
                dimCache[key] = [];
              }
              dimCache[key].push(value);
              return _this2.uiSegmentSrv.newSegment({
                value: values[0]
              });
            });
            dims.unshift(this.removeTagFilterSegment);
            this.dimCache = dimCache;
            return dims;
          }
        }, {
          key: 'tagSegmentUpdated',
          value: function tagSegmentUpdated(segment, index) {
            this.tagSegments[index] = segment;

            // handle remove tag condition
            if (segment.value === this.removeTagFilterSegment.value) {
              this.tagSegments.splice(index, 3);
              if (this.tagSegments.length === 0) {
                this.tagSegments.push(this.uiSegmentSrv.newPlusButton());
              } else if (this.tagSegments.length > 2) {
                this.tagSegments.splice(Math.max(index - 1, 0), 1);
                if (this.tagSegments[this.tagSegments.length - 1].type !== 'plus-button') {
                  this.tagSegments.push(this.uiSegmentSrv.newPlusButton());
                }
              }
            } else {
              if (segment.type === 'plus-button') {
                if (index > 2) {
                  this.tagSegments.splice(index, 0, this.uiSegmentSrv.newCondition(','));
                }
                this.tagSegments.push(this.uiSegmentSrv.newOperator('='));
                this.tagSegments.push(this.uiSegmentSrv.newFake('select tag value', 'value', 'query-segment-value'));
                segment.type = 'key';
                segment.cssClass = 'query-segment-key';
              }

              if (index + 1 === this.tagSegments.length) {
                this.tagSegments.push(this.uiSegmentSrv.newPlusButton());
              }
            }

            this.rebuildTargetTagConditions();
          }
        }, {
          key: 'rebuildTargetTagConditions',
          value: function rebuildTargetTagConditions() {
            var tags = [];
            var tagIndex = 0;

            _.each(this.tagSegments, function (segment2, index) {
              if (segment2.type === 'key') {
                if (tags.length === 0) {
                  tags.push({});
                }
                tags[tagIndex].key = segment2.value;
              } else if (segment2.type === 'value') {
                tags[tagIndex].value = segment2.value;
              } else if (segment2.type === 'condition') {
                tags.push({ condition: segment2.value });
                tagIndex += 1;
              } else if (segment2.type === 'operator') {
                tags[tagIndex].operator = segment2.value;
              }
            });

            this.target.tags = tags;
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
