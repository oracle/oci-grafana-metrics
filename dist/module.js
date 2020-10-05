'use strict';

System.register(['./datasource', './query_ctrl', './config_ctrl'], function (_export, _context) {
  "use strict";

  var OCIDatasource, OCIDatasourceQueryCtrl, OCIConfigCtrl, OCIQueryOptionsCtrl, OCIAnnotationsQueryCtrl;

  function _classCallCheck(instance, Constructor) {
    if (!(instance instanceof Constructor)) {
      throw new TypeError("Cannot call a class as a function");
    }
  }

  return {
    setters: [function (_datasource) {
      OCIDatasource = _datasource.default;
    }, function (_query_ctrl) {
      OCIDatasourceQueryCtrl = _query_ctrl.OCIDatasourceQueryCtrl;
    }, function (_config_ctrl) {
      OCIConfigCtrl = _config_ctrl.OCIConfigCtrl;
    }],
    execute: function () {
      _export('QueryOptionsCtrl', OCIQueryOptionsCtrl = function OCIQueryOptionsCtrl() {
        _classCallCheck(this, OCIQueryOptionsCtrl);
      });

      OCIQueryOptionsCtrl.templateUrl = 'partials/query.options.html';

      _export('AnnotationsQueryCtrl', OCIAnnotationsQueryCtrl = function OCIAnnotationsQueryCtrl() {
        _classCallCheck(this, OCIAnnotationsQueryCtrl);
      });

      OCIAnnotationsQueryCtrl.templateUrl = 'partials/annotations.editor.html';

      _export('Datasource', OCIDatasource);

      _export('QueryCtrl', OCIDatasourceQueryCtrl);

      _export('ConfigCtrl', OCIConfigCtrl);

      _export('QueryOptionsCtrl', OCIQueryOptionsCtrl);

      _export('AnnotationsQueryCtrl', OCIAnnotationsQueryCtrl);
    }
  };
});
//# sourceMappingURL=module.js.map
