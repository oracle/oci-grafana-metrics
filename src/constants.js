/*
** Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
export const AUTO = 'auto'
export const regions = ['ap-chuncheon-1', 'ap-hyderabad-1', 'ap-melbourne-1', 'ap-mumbai-1', 'ap-osaka-1', 'ap-seoul-1', 'ap-sydney-1', 'ap-tokyo-1', 'ca-montreal-1', 'ca-toronto-1', 'eu-amsterdam-1', 'eu-frankfurt-1', 'eu-zurich-1', 'me-jeddah-1', 'sa-saopaulo-1', 'uk-london-1', 'us-ashburn-1', 'us-phoenix-1']
export const namespaces = ['oci_computeagent', 'oci_blockstore', 'oci_lbaas', 'oci_telemetry']
export const aggregations = ['count()', 'max()', 'mean()', 'min()', 'rate()', 'sum()', 'percentile(.90)', 'percentile(.95)', 'percentile(.99)']
export const windows = [AUTO, '1m', '5m', '1h']
export const resolutions = [AUTO, '1m', '5m', '1h']
export const environments = ['local', 'OCI Instance']


export const compartmentsQueryRegex = /^compartments\(\)\s*/;
export const regionsQueryRegex = /^regions\(\)\s*/;
export const namespacesQueryRegex = /^namespaces\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/;
export const resourcegroupsQueryRegex = /^resourcegroups\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/;
export const metricsQueryRegex = /^metrics\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/;
export const dimensionKeysQueryRegex = /^dimensions\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/;
export const dimensionValuesQueryRegex = /^dimensionOptions\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/;
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
