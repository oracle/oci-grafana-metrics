import { DataSourcePlugin } from '@grafana/data';
import { OCIDataSource } from './datasource';
import { ConfigEditor } from './ConfigEditor';
import { QueryEditor } from './QueryEditor';
import { OCIQuery, OCIDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<OCIDataSource, OCIQuery, OCIDataSourceOptions>(OCIDataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor)
