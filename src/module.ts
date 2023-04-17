/*
** Copyright © 2022 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
import { OCIDatasourceQueryCtrl } from './query_ctrl'
// import { OCIConfigCtrl } from './config_ctrl'
// import { DataSourcePlugin } from '@grafana/data';
// import { QueryEditor } from './QueryEditor';


import { DataSourcePlugin } from '@grafana/data';
import { OCIDatasource } from './datasource';
import { ConfigEditor } from './ConfigEditor';
import { QueryEditor } from './QueryEditor';
import { OCIConfig } from './types';


export const plugin = new DataSourcePlugin<OCIDatasource, OCIConfig>(OCIDatasource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);