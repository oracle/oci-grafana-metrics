import { DataSourcePlugin } from '@grafana/data';
import { OCIDataSource } from './datasource';
import { ConfigEditor } from './ConfigEditor';
import { QueryEditor } from './QueryEditor';
import { OCIQuery, OCIDataSourceOptions } from './types';

/**
 * @module OCIPlugin
 * @description
 * This module serves as the entry point for the OCI (Oracle Cloud Infrastructure) data source plugin in Grafana.
 * It defines the plugin's core components, including the data source class, configuration editor, and query editor.
 * It uses the Grafana `DataSourcePlugin` class to integrate these components into the Grafana ecosystem.
 */

/**
 * @constant plugin
 * @description
 * This constant represents the OCI data source plugin instance. It is created using the `DataSourcePlugin` class
 * from `@grafana/data` and configured with the necessary components for the OCI data source.
 *
 * @type {DataSourcePlugin<OCIDataSource, OCIQuery, OCIDataSourceOptions>}
 * @exports plugin
 *
 * @example
 * // Example of how this plugin is used within Grafana:
 * // Grafana loads this module, and the 'plugin' constant is used to register the OCI data source.
 * // When a user adds a new data source of type 'oci', Grafana uses this plugin to render the configuration and query editors.
 *
 * @see {@link OCIDataSource} - The data source class responsible for interacting with OCI.
 * @see {@link ConfigEditor} - The configuration editor component for setting up the OCI data source.
 * @see {@link QueryEditor} - The query editor component for building and editing OCI metric queries.
 * @see {@link OCIQuery} - The interface defining the structure of an OCI query.
 * @see {@link OCIDataSourceOptions} - The interface defining the options for the OCI data source.
 */
export const plugin = new DataSourcePlugin<OCIDataSource, OCIQuery, OCIDataSourceOptions>(OCIDataSource)
  /**
   * @method setConfigEditor
   * @description
   * Sets the configuration editor component for the OCI data source plugin.
   * This component is used to configure the data source settings, such as authentication details and tenancy information.
   *
   * @param {typeof ConfigEditor} ConfigEditor - The configuration editor component.
   * @returns {DataSourcePlugin<OCIDataSource, OCIQuery, OCIDataSourceOptions>} - The plugin instance for method chaining.
   * @see {@link ConfigEditor}
   */
  .setConfigEditor(ConfigEditor)
  /**
   * @method setQueryEditor
   * @description
   * Sets the query editor component for the OCI data source plugin.
   * This component is used to build and edit queries for retrieving metrics from OCI.
   *
   * @param {typeof QueryEditor} QueryEditor - The query editor component.
   * @returns {DataSourcePlugin<OCIDataSource, OCIQuery, OCIDataSourceOptions>} - The plugin instance for method chaining.
   * @see {@link QueryEditor}
   */
  .setQueryEditor(QueryEditor);
