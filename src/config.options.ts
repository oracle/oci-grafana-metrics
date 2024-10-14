/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/

import { SelectableValue } from '@grafana/data';

export enum AuthProviders {
  OCI_USER = 'local',
  OCI_INSTANCE = 'OCI Instance',
}


export const namespaces = ['oci_computeagent', 'oci_blockstore', 'oci_lbaas', 'oci_telemetry'];
export const environments = ['local', 'OCI Instance'];


export enum TenancyChoices {
  multitenancy = 'multitenancy',
  single = 'single',
}

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
