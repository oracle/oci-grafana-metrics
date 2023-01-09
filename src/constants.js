/*
** Copyright © 2022 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
export const AUTO = 'auto'
export const regions = ['af-johannesburg-1', 'ap-chiyoda-1', 'ap-chuncheon-1', 'ap-dcc-canberra-1', 'ap-hyderabad-1', 'ap-ibaraki-1', 'ap-melbourne-1',
    'ap-mumbai-1', 'ap-osaka-1', 'ap-seoul-1', 'ap-singapore-1', 'ap-sydney-1', 'ap-tokyo-1', 'ca-montreal-1', 'ca-toronto-1',
    'eu-amsterdam-1', 'eu-frankfurt-1', 'eu-madrid-1', 'eu-marseille-1', 'eu-milan-1', 'eu-paris-1', 'eu-stockholm-1', 'eu-zurich-1',
    'il-jerusalem-1', 'me-abudhabi-1', 'me-dubai-1', 'me-jeddah-1', 'me-dcc-muscat-1', 'mx-queretaro-1', 'sa-santiago-1', 'sa-saopaulo-1', 'sa-vinhedo-1',
    'uk-cardiff-1', 'uk-gov-cardiff-1', 'uk-gov-london-1', 'uk-london-1', 'us-ashburn-1', 'us-chicago-1', 'us-gov-ashburn-1',
    'us-gov-chicago-1', 'us-gov-phoenix-1', 'us-langley-1', 'us-luke-1', 'us-phoenix-1', 'us-sanjose-1']

export const namespaces = ['oci_computeagent', 'oci_blockstore', 'oci_lbaas', 'oci_telemetry']
export const aggregations = ['count()', 'max()', 'mean()', 'min()', 'rate()', 'sum()', 'percentile(.90)', 'percentile(.95)', 'percentile(.99)', 'last()']
export const windows = [AUTO, '1m', '5m', '1h']
export const resolutions = [AUTO, '1m', '5m', '1h']
export const environments = ['local', 'OCI Instance']
export const tenancymodes = ['single', 'multitenancy']


export const compartmentsQueryRegex = /^compartments\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)|^compartments\(\)\s*/;
export const regionsQueryRegex = /^regions\(\s*(\".+\"|\'.+\'|\$\w+)\s*\)|^regions\(\)\s*/;
export const tenanciesQueryRegex = /^tenancies\(\)\s*/;
export const namespacesQueryRegex = /^namespaces\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const resourcegroupsQueryRegex = /^resourcegroups\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const metricsQueryRegex = /^metrics\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const dimensionKeysQueryRegex = /^dimensions\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const dimensionValuesQueryRegex = /^dimensionOptions\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const windowsAndResolutionRegex = /^[0-9]+[mhs]$/;

export const removeQuotes = str => {
    if (!str) return str;

    let res = str;
    if (str.startsWith("'") || str.startsWith('"')) {
        res = res.slice(1);
    }
    if (str.endsWith("'") || str.endsWith('"')) {
        res = res.slice(0, res.length - 1);
    }
    return res;
}

// if the user selects a time range less than 7 days ->  window will be 1m and resolution will be 1 min
//
// if the user selects a time range less than 30 days and more than 7 days ->   window will be 5m and resolution will be 5 min.
//
//   if the user select time range less than 90 days and more than 30 days -> a window will be 1h and resolution will be 1 h

export const SEVEN_DAYS = 7
export const THIRTY_DAYS = 30
export const NINETY_DAYS = 90

export const d0To7Config = { window: '1m', resolution: '1m' }
export const d8To30Config = { window: '5m', resolution: '5m' }
export const d31toInfConfig = { window: '1h', resolution: '1h' }

export const autoTimeIntervals = [
  [SEVEN_DAYS, d0To7Config],
  [THIRTY_DAYS, d8To30Config],
  [NINETY_DAYS, d31toInfConfig]
]
