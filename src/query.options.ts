/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/

export const AUTO = 'auto'
export const aggregations = ['count()', 'max()', 'mean()', 'min()', 'rate()', 'sum()', 'percentile(.90)', 'percentile(.95)', 'percentile(.99)', 'last()']
export const windows = [AUTO, '1m', '5m', '1h']
export const resolutions = [AUTO, '1m', '5m', '1h']
export const DEFAULT_TENANCY = "DEFAULT/";


export const compartmentsQueryRegex = /^compartments\(\s*(\".+\"|\'.+\'|\$\w+)\s*\)|^compartments\(\)\s*/;
export const regionsQueryRegex = /^regions\(\s*(\".+\"|\'.+\'|\$\w+)\s*\)|^regions\(\)\s*/;
export const tenanciesQueryRegex = /^tenancies\(\)\s*/;
export const namespacesQueryRegex = /^namespaces\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const resourcegroupsQueryRegex = /^resourcegroups\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const metricsQueryRegex = /^metrics\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const dimensionQueryRegex = /^dimensions\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const dimensionValuesQueryRegex = /^dimensionOptions\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const windowsAndResolutionRegex = /^[0-9]+[mhs]$/;

// export const removeQuotes = (str: string): string => {
//     if (!str) return str;

//     let res = str;
//     if (str.startsWith("'") || str.startsWith('"')) {
//         res = res.slice(1);
//     }
//     if (str.endsWith("'") || str.endsWith('"')) {
//         res = res.slice(0, res.length - 1);
//     }
//     return res;
// };


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
