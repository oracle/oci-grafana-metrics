import OCIDatasource from './datasource'
import { OCIDatasourceQueryCtrl } from './query_ctrl'
import { OCIConfigCtrl } from './config_ctrl'

class OCIQueryOptionsCtrl {}
OCIQueryOptionsCtrl.templateUrl = 'partials/query.options.html'

class OCIAnnotationsQueryCtrl {}
OCIAnnotationsQueryCtrl.templateUrl = 'partials/annotations.editor.html'

export {
  OCIDatasource as Datasource,
  OCIDatasourceQueryCtrl as QueryCtrl,
  OCIConfigCtrl as ConfigCtrl,
  OCIQueryOptionsCtrl as QueryOptionsCtrl,
  OCIAnnotationsQueryCtrl as AnnotationsQueryCtrl
}
