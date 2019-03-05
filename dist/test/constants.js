'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});
var regions = exports.regions = ['ca-toronto-1', 'eu-frankfurt-1', 'uk-london-1', 'us-ashburn-1', 'us-phoenix-1'];
var namespaces = exports.namespaces = ['oci_computeagent', 'oci_blockstore', 'oci_lbaas', 'oci_telemetry'];
var aggregations = exports.aggregations = ['count()', 'max()', 'mean()', 'min()', 'rate()', 'sum()', 'percentile(.90)', 'percentile(.95)', 'percentile(.99)'];
var windows = exports.windows = ['1m', '5m', '1h'];
var environments = exports.environments = ['local', 'OCI Instance'];
//# sourceMappingURL=constants.js.map
