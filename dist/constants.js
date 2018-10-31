'use strict';

System.register([], function (_export, _context) {
  "use strict";

  var regions, namespaces, aggregations, windows, environments;
  return {
    setters: [],
    execute: function () {
      _export('regions', regions = ['SEA', 'PHX', 'IAD', 'FRA', 'LHR', 'LFI', 'LUF']);

      _export('regions', regions);

      _export('namespaces', namespaces = ['oci_computeagent', 'oci_blockstore', 'oci_lbaas', 'oci_telemetry']);

      _export('namespaces', namespaces);

      _export('aggregations', aggregations = ['count()', 'max()', 'mean()', 'min()', 'rate()', 'sum()', 'percentile(.90)', 'percentile(.95)', 'percentile(.99)']);

      _export('aggregations', aggregations);

      _export('windows', windows = ['1m', '5m', '1h']);

      _export('windows', windows);

      _export('environments', environments = ['local', 'OCI Instance']);

      _export('environments', environments);
    }
  };
});
//# sourceMappingURL=constants.js.map
