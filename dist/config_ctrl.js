'use strict';

System.register(['./constants'], function (_export, _context) {
  "use strict";

  var regions, environments, _createClass, OCIConfigCtrl;

  function _classCallCheck(instance, Constructor) {
    if (!(instance instanceof Constructor)) {
      throw new TypeError("Cannot call a class as a function");
    }
  }

  return {
    setters: [function (_constants) {
      regions = _constants.regions;
      environments = _constants.environments;
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

      _export('OCIConfigCtrl', OCIConfigCtrl = function () {
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
            return regions;
          }
        }, {
          key: 'getEnvironments',
          value: function getEnvironments() {
            return environments;
          }
        }]);

        return OCIConfigCtrl;
      }());

      _export('OCIConfigCtrl', OCIConfigCtrl);

      OCIConfigCtrl.templateUrl = 'partials/config.html';
    }
  };
});
//# sourceMappingURL=config_ctrl.js.map
