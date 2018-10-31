'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.OCIDatasource = undefined;

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _lodash = require('lodash');

var _lodash2 = _interopRequireDefault(_lodash);

var _constants = require('./constants');

var _retry = require('./util/retry');

var _retry2 = _interopRequireDefault(_retry);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var OCIDatasource = exports.OCIDatasource = function () {
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
  }

  _createClass(OCIDatasource, [{
    key: 'query',
    value: function query(options) {
      var query = this.buildQueryParameters(options);
      query.targets = query.targets.filter(function (t) {
        return !t.hide;
      });
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
  }, {
    key: 'buildQueryParameters',
    value: function buildQueryParameters(options) {
      var _this2 = this;

      // remove placeholder targets
      options.targets = _lodash2.default.filter(options.targets, function (target) {
        return target.metric !== 'select metric';
      });

      var targets = _lodash2.default.map(options.targets, function (target) {
        var region = target.region;
        var t = [];
        if (target.hasOwnProperty('tags')) {
          for (var i = 0; i < target.tags.length; i++) {
            if (target.tags[i].value !== 'select tag value') {
              t.push(target.tags[i].key + ' ' + target.tags[i].operator + ' "' + target.tags[i].value + '"');
            }
          }
          t.join();
        }

        if (target.region === 'select region') {
          region = _this2.defaultRegion;
        }
        var dimension = t.length === 0 ? '' : '{' + t + '}';

        return {
          compartment: _this2.templateSrv.replace(target.compartment, options.scopedVars || {}),
          environment: _this2.environment,
          queryType: 'query',
          region: _this2.templateSrv.replace(region, options.scopedVars || {}),
          tenancyOCID: _this2.tenancyOCID,
          namespace: _this2.templateSrv.replace(target.namespace, options.scopedVars || {}),
          resolution: target.resolution,
          refId: target.refId,
          hide: target.hide,
          type: target.type || 'timeserie',
          datasourceId: _this2.id,
          query: _this2.templateSrv.replace(target.metric, options.scopedVars || {}) + '[' + target.window + ']' + dimension + '.' + target.aggregation
        };
      });

      options.targets = targets;

      return options;
    }
  }, {
    key: 'testDatasource',
    value: function testDatasource() {
      return this.doRequest({
        targets: [{
          queryType: 'query',
          refId: 'test',
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
      });
    }
  }, {
    key: 'annotationQuery',
    value: function annotationQuery(options) {
      var query = this.templateSrv.replace(options.annotation.query, {}, 'glob');
      var annotationQuery = {
        range: options.range,
        annotation: {
          name: options.annotation.name,
          datasource: options.annotation.datasource,
          enable: options.annotation.enable,
          iconColor: options.annotation.iconColor,
          query: query
        },
        rangeRaw: options.rangeRaw
      };

      return this.doRequest({
        url: this.url + '/annotations',
        method: 'POST',
        data: annotationQuery
      }).then(function (result) {
        return result.data;
      });
    }
  }, {
    key: 'templateMeticSearch',
    value: function templateMeticSearch(varString) {
      var compartmentQuery = varString.match(/^compartments\(\)/);
      if (compartmentQuery) {
        return this.getCompartments();
      }

      var regionQuery = varString.match(/^regions\(\)/);
      if (regionQuery) {
        var regs = _constants.regions.map(function (reg) {
          return { row: reg, value: reg };
        });
        return this.q.when(regs);
      }

      var metricQuery = varString.match(/metrics\((\$?\w+)(,\s*\$\w+)*\)/);
      if (metricQuery) {
        var target = {
          namespace: this.templateSrv.replace(metricQuery[1]),
          compartment: this.templateSrv.replace(metricQuery[2]).replace(',', '').trim()
        };
        return this.metricFindQuery(target);
      }

      var namespaceQuery = varString.match(/namespaces\(\)/);
      if (namespaceQuery) {
        var names = _constants.namespaces.map(function (reg) {
          return { row: reg, value: reg };
        });
        return this.q.when(names);
      }
      throw new Error('Unable to parse templating string');
    }
  }, {
    key: 'metricFindQuery',
    value: function metricFindQuery(target) {
      var _this3 = this;

      if (typeof target === 'string') {
        return this.templateMeticSearch(target);
      }

      var range = this.timeSrv.timeRange();
      var region = this.defaultRegion;
      if (target.namespace === 'select namespace') {
        target.namespace = '';
      }
      if (target.compartment === 'select compartment') {
        target.compartment = '';
      }
      if (Object.hasOwnProperty(target, 'region') && target.region !== 'select region') {
        region = target.region;
      }

      var targets = [{
        compartment: this.templateSrv.replace(target.compartment),
        environment: this.environment,
        queryType: 'search',
        tenancyOCID: this.tenancyOCID,
        region: this.templateSrv.replace(region),
        datasourceId: this.id,
        refId: 'search',
        namespace: this.templateSrv.replace(target.namespace)
      }];
      var options = {
        range: range,
        targets: targets
      };
      return this.doRequest(options).then(function (res) {
        return _this3.mapToTextValue(res, 'search');
      });
    }
  }, {
    key: 'mapToTextValue',
    value: function mapToTextValue(result, searchField) {
      var table = result.data.results[searchField].tables[0];
      if (!table) {
        return [];
      }

      var m = _lodash2.default.map(table.rows, function (row, i) {
        if (row.length > 1) {
          return { text: row[0], value: row[1] };
        } else if (_lodash2.default.isObject(row[0])) {
          return { text: row[0], value: i };
        }
        return { text: row[0], value: row[0] };
      });
      return m;
    }
  }, {
    key: 'getCompartments',
    value: function getCompartments() {
      var _this4 = this;

      var range = this.timeSrv.timeRange();
      var targets = [{
        environment: this.environment,
        region: this.defaultRegion,
        tenancyOCID: this.tenancyOCID,
        queryType: 'compartments',
        datasourceId: this.id,
        refId: 'compartments'
      }];
      var options = {
        range: range,
        targets: targets
      };
      return this.doRequest(options).then(function (res) {
        return _this4.mapToTextValue(res, 'compartment');
      });
    }
  }, {
    key: 'getDimensions',
    value: function getDimensions(target) {
      var _this5 = this;

      var range = this.timeSrv.timeRange();
      var region = target.region;
      if (target.namespace === 'select namespace') {
        target.namespace = '';
      }
      if (target.compartment === 'select compartment') {
        target.compartment = '';
      }
      if (target.metric === 'select metric') {
        return [];
      }
      if (region === 'select region') {
        region = this.defaultRegion;
      }

      var targets = [{
        compartment: this.templateSrv.replace(target.compartment),
        environment: this.environment,
        queryType: 'dimensions',
        region: this.templateSrv.replace(region),
        tenancyOCID: this.tenancyOCID,

        datasourceId: this.id,
        refId: 'dimensions',
        metric: this.templateSrv.replace(target.metric),
        namespace: this.templateSrv.replace(target.namespace)
      }];

      var options = {
        range: range,
        targets: targets
      };
      return this.doRequest(options).then(function (res) {
        return _this5.mapToTextValue(res, 'dimensions');
      });
    }
  }, {
    key: 'getNamespaces',
    value: function getNamespaces(target) {
      var _this6 = this;

      var region = target.region;
      if (region === 'select region') {
        region = this.defaultRegion;
      }
      return this.doRequest({ targets: [{
          // commonRequestParameters
          compartment: this.templateSrv.replace(target.compartment),
          environment: this.environment,
          queryType: 'namespaces',
          region: this.templateSrv.replace(region),
          tenancyOCID: this.tenancyOCID,

          datasourceId: this.id,
          refId: 'namespaces'
        }],
        range: this.timeSrv.timeRange()
      }).then(function (namespaces) {
        return _this6.mapToTextValue(namespaces, 'namespaces');
      });
    }
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
  }]);

  return OCIDatasource;
}();
//# sourceMappingURL=datasource.js.map
