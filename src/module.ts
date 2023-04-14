/*
** Copyright © 2022 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
import OCIDatasource from './datasource'
import { OCIDatasourceQueryCtrl } from './query_ctrl'
// import { OCIConfigCtrl } from './config_ctrl'
import { DataSourcePlugin } from '@grafana/data';
// import { DataSource } from './datasource';
import { ConfigEditor } from './ConfigEditor';
// import { QueryEditor } from './QueryEditor';
import { MyQuery, MyDataSourceOptions } from './types';

// export const plugin = new DataSourcePlugin<DataSource, MyQuery, MyDataSourceOptions>(DataSource)
//   .setConfigEditor(ConfigEditor)
//   .setQueryEditor(QueryEditor);

// class OCIQueryOptionsCtrl {}
// OCIQueryOptionsCtrl.templateUrl = 'partials/query.options.html'

// class OCIAnnotationsQueryCtrl {}
// OCIAnnotationsQueryCtrl.templateUrl = 'partials/annotations.editor.html'

// export {
//   OCIDatasource as Datasource,
//   OCIDatasourceQueryCtrl as QueryCtrl,
//   ConfigEditor as ConfigCtrl
//   // OCIQueryOptionsCtrl as QueryOptionsCtrl,
//   // OCIAnnotationsQueryCtrl as AnnotationsQueryCtrl
// }

export const plugin = new DataSourcePlugin<Datasource, MyQuery, MyDataSourceOptions>(Datasource)
  .setConfigEditor(ConfigEditor);
  // .setQueryEditor(QueryEditor);
