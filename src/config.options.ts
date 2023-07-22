/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/

import { SelectableValue } from '@grafana/data';

export enum AuthProviders {
  OCI_USER = 'local',
  OCI_INSTANCE = 'OCI Instance',
}

export const regions = ['af-johannesburg-1', 'ap-chiyoda-1', 'ap-chuncheon-1', 'ap-dcc-canberra-1', 'ap-hyderabad-1', 'ap-ibaraki-1', 'ap-melbourne-1',
    'ap-mumbai-1', 'ap-osaka-1', 'ap-seoul-1', 'ap-singapore-1', 'ap-sydney-1', 'ap-tokyo-1', 'ca-montreal-1', 'ca-toronto-1',
    'eu-amsterdam-1', 'eu-frankfurt-1', 'eu-madrid-1', 'eu-marseille-1', 'eu-milan-1', 'eu-paris-1', 'eu-stockholm-1', 'eu-zurich-1',
    'il-jerusalem-1', 'me-abudhabi-1', 'me-dubai-1', 'me-jeddah-1', 'me-dcc-muscat-1', 'mx-queretaro-1', 'sa-santiago-1', 'sa-saopaulo-1', 'sa-vinhedo-1',
    'uk-cardiff-1', 'uk-gov-cardiff-1', 'uk-gov-london-1', 'uk-london-1', 'us-ashburn-1', 'us-chicago-1', 'us-gov-ashburn-1',
    'us-gov-chicago-1', 'us-gov-phoenix-1', 'us-langley-1', 'us-luke-1', 'us-phoenix-1', 'us-sanjose-1'];

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
