'use strict';

Object.defineProperty(exports, "__esModule", {
    value: true
});
/*
** Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
var AUTO = exports.AUTO = 'auto';
var regions = exports.regions = ['ap-chuncheon-1', 'ap-hyderabad-1', 'ap-melbourne-1', 'ap-mumbai-1', 'ap-osaka-1', 'ap-seoul-1', 'ap-sydney-1', 'ap-tokyo-1', 'ca-montreal-1', 'ca-toronto-1', 'eu-amsterdam-1', 'eu-frankfurt-1', 'eu-zurich-1', 'me-jeddah-1', 'sa-saopaulo-1', 'uk-london-1', 'us-ashburn-1', 'us-phoenix-1'];
var namespaces = exports.namespaces = ['oci_computeagent', 'oci_blockstore', 'oci_lbaas', 'oci_telemetry'];
var aggregations = exports.aggregations = ['count()', 'max()', 'mean()', 'min()', 'rate()', 'sum()', 'percentile(.90)', 'percentile(.95)', 'percentile(.99)'];
var windows = exports.windows = [AUTO, '1m', '5m', '1h'];
var resolutions = exports.resolutions = [AUTO, '1m', '5m', '1h'];
var environments = exports.environments = ['local', 'OCI Instance'];

var compartmentsQueryRegex = exports.compartmentsQueryRegex = /^compartments\(\)\s*/;
var regionsQueryRegex = exports.regionsQueryRegex = /^regions\(\)\s*/;
var namespacesQueryRegex = exports.namespacesQueryRegex = /^namespaces\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/;
var resourcegroupsQueryRegex = exports.resourcegroupsQueryRegex = /^resourcegroups\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/;
var metricsQueryRegex = exports.metricsQueryRegex = /^metrics\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/;
var dimensionKeysQueryRegex = exports.dimensionKeysQueryRegex = /^dimensions\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/;
var dimensionValuesQueryRegex = exports.dimensionValuesQueryRegex = /^dimensionOptions\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/;
var windowsAndResolutionRegex = exports.windowsAndResolutionRegex = /^[0-9]+[mhs]$/;

var removeQuotes = exports.removeQuotes = function removeQuotes(str) {
    if (!str) return str;

    var res = str;
    if (str.startsWith("'") || str.startsWith('"')) {
        res = res.slice(1);
    }
    if (str.endsWith("'") || str.endsWith('"')) {
        res = res.slice(0, res.length - 1);
    }
    return res;
};

// if the user selects a time range less than 7 days ->  window will be 1m and resolution will be 1 min
//
// if the user selects a time range less than 30 days and more than 7 days ->   window will be 5m and resolution will be 5 min.
//
//   if the user select time range less than 90 days and more than 30 days -> a window will be 1h and resolution will be 1 h

var SEVEN_DAYS = exports.SEVEN_DAYS = 7;
var THIRTY_DAYS = exports.THIRTY_DAYS = 30;
var NINETY_DAYS = exports.NINETY_DAYS = 90;

var d0To7Config = exports.d0To7Config = { window: '1m', resolution: '1m' };
var d8To30Config = exports.d8To30Config = { window: '5m', resolution: '5m' };
var d31toInfConfig = exports.d31toInfConfig = { window: '1h', resolution: '1h' };

var autoTimeIntervals = exports.autoTimeIntervals = [[SEVEN_DAYS, d0To7Config], [THIRTY_DAYS, d8To30Config], [NINETY_DAYS, d31toInfConfig]];
//# sourceMappingURL=constants.js.map
