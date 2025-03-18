/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/

import { SelectableValue } from '@grafana/data';

/**
 * @module ConfigOptions
 * @description
 * This module defines the configuration options and enums used within the OCI (Oracle Cloud Infrastructure)
 * data source plugin for Grafana. It includes authentication providers, tenancy choices, and selectable
 * value options for configuring the plugin's behavior.
 */

/**
 * @enum AuthProviders
 * @description
 * Enumerates the available authentication providers for the OCI data source.
 *
 * @property {string} OCI_USER - Represents the 'local' authentication method, where OCI user credentials are used.
 * @property {string} OCI_INSTANCE - Represents the 'OCI Instance' authentication method, where the Grafana instance is running within an OCI environment and uses instance principals.
 */
export enum AuthProviders {
  OCI_USER = 'local',
  OCI_INSTANCE = 'OCI Instance',
}

/**
 * @constant namespaces
 * @description
 * An array of commonly used OCI namespaces.
 *
 * @type {string[]}
 * @example
 * // Example usage:
 * // const myNamespace = namespaces[0]; // 'oci_computeagent'
 */
export const namespaces = ['oci_computeagent', 'oci_blockstore', 'oci_lbaas', 'oci_telemetry'];

/**
 * @constant environments
 * @description
 * An array of valid environment types for the OCI data source.
 *
 * @type {string[]}
 * @example
 * // Example usage:
 * // const myEnvironment = environments[1]; // 'OCI Instance'
 */
export const environments = ['local', 'OCI Instance'];

/**
 * @enum TenancyChoices
 * @description
 * Enumerates the available tenancy modes for the OCI data source.
 *
 * @property {string} multitenancy - Represents the 'multitenancy' mode, where the plugin can fetch metrics across multiple OCI tenancies.
 * @property {string} single - Represents the 'single' tenancy mode, where the plugin is configured for a single OCI tenancy.
 */
export enum TenancyChoices {
  multitenancy = 'multitenancy',
  single = 'single',
}

/**
 * @constant TenancyChoiceOptions
 * @description
 * An array of selectable value options for choosing the tenancy mode.
 *
 * @type {SelectableValue<string>[]}
 * @example
 * // Example usage:
 * // <Select options={TenancyChoiceOptions} />
 */
export const TenancyChoiceOptions = [
  {
    label: 'Multi Tenancy',
    value: TenancyChoices.multitenancy,
  },
  {
    label: 'Single Tenancy',
    value: TenancyChoices.single,
  },
] as Array<SelectableValue<string>>;

/**
 * @constant AuthProviderOptions
 * @description
 * An array of selectable value options for choosing the authentication provider.
 *
 * @type {SelectableValue<string>[]}
 * @example
 * // Example usage:
 * // <Select options={AuthProviderOptions} />
 */
export const AuthProviderOptions = [
  {
    label: 'OCI User',
    value: AuthProviders.OCI_USER,
    description: 'The grafana instance is configured with oci user principals',
  },
  {
    label: 'OCI Instance',
    value: AuthProviders.OCI_INSTANCE,
    description: 'The grafana instance is configured in OCI environment',
  },
] as Array<SelectableValue<string>>;
