'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.OCIConfigCtrl = undefined;

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _constants = require('./constants');

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var OCIConfigCtrl = exports.OCIConfigCtrl = function () {
  /** @ngInject */
  function OCIConfigCtrl($scope, backendSrv) {
    _classCallCheck(this, OCIConfigCtrl);

    this.backendSrv = backendSrv;
    this.tenancyOCID = this.current.jsonData.tenancyOCID;
    this.defaultRegion = this.current.jsonData.defaultRegion;
    this.environment = this.current.jsonData.environment;
  }

  _createClass(OCIConfigCtrl, [{
    key: 'getRegions',
    value: function getRegions() {
      return _constants.regions;
    }
  }, {
    key: 'getEnvironments',
    value: function getEnvironments() {
      return _constants.environments;
    }
  }]);

  return OCIConfigCtrl;
}();

OCIConfigCtrl.templateUrl = 'partials/config.html';
//# sourceMappingURL=config_ctrl.js.map
