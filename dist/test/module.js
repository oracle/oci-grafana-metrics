'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.AnnotationsQueryCtrl = exports.QueryOptionsCtrl = exports.ConfigCtrl = exports.QueryCtrl = exports.Datasource = undefined;

var _datasource = require('./datasource');

var _query_ctrl = require('./query_ctrl');

var _config_ctrl = require('./config_ctrl');

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var OCIQueryOptionsCtrl = function OCIQueryOptionsCtrl() {
  _classCallCheck(this, OCIQueryOptionsCtrl);
};

OCIQueryOptionsCtrl.templateUrl = 'partials/query.options.html';

var OCIAnnotationsQueryCtrl = function OCIAnnotationsQueryCtrl() {
  _classCallCheck(this, OCIAnnotationsQueryCtrl);
};

OCIAnnotationsQueryCtrl.templateUrl = 'partials/annotations.editor.html';

exports.Datasource = _datasource.OCIDatasource;
exports.QueryCtrl = _query_ctrl.OCIDatasourceQueryCtrl;
exports.ConfigCtrl = _config_ctrl.OCIConfigCtrl;
exports.QueryOptionsCtrl = OCIQueryOptionsCtrl;
exports.AnnotationsQueryCtrl = OCIAnnotationsQueryCtrl;
//# sourceMappingURL=module.js.map
