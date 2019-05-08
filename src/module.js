/*
** Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
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
