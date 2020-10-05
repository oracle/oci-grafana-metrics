'use strict';

System.register([], function (_export, _context) {
    "use strict";

    var AUTO, regions, namespaces, aggregations, windows, resolutions, environments, compartmentsQueryRegex, regionsQueryRegex, namespacesQueryRegex, resourcegroupsQueryRegex, metricsQueryRegex, dimensionKeysQueryRegex, dimensionValuesQueryRegex, windowsAndResolutionRegex, removeQuotes, SEVEN_DAYS, THIRTY_DAYS, NINETY_DAYS, d0To7Config, d8To30Config, d31toInfConfig, autoTimeIntervals;
    return {
        setters: [],
        execute: function () {
            _export('AUTO', AUTO = 'auto');

            _export('AUTO', AUTO);

            _export('regions', regions = ['ap-chuncheon-1', 'ap-hyderabad-1', 'ap-melbourne-1', 'ap-mumbai-1', 'ap-osaka-1', 'ap-seoul-1', 'ap-sydney-1', 'ap-tokyo-1', 'ca-montreal-1', 'ca-toronto-1', 'eu-amsterdam-1', 'eu-frankfurt-1', 'eu-zurich-1', 'me-jeddah-1', 'sa-saopaulo-1', 'uk-london-1', 'us-ashburn-1', 'us-phoenix-1']);

            _export('regions', regions);

            _export('namespaces', namespaces = ['oci_computeagent', 'oci_blockstore', 'oci_lbaas', 'oci_telemetry']);

            _export('namespaces', namespaces);

            _export('aggregations', aggregations = ['count()', 'max()', 'mean()', 'min()', 'rate()', 'sum()', 'percentile(.90)', 'percentile(.95)', 'percentile(.99)']);

            _export('aggregations', aggregations);

            _export('windows', windows = [AUTO, '1m', '5m', '1h']);

            _export('windows', windows);

            _export('resolutions', resolutions = [AUTO, '1m', '5m', '1h']);

            _export('resolutions', resolutions);

            _export('environments', environments = ['local', 'OCI Instance']);

            _export('environments', environments);

            _export('compartmentsQueryRegex', compartmentsQueryRegex = /^compartments\(\)\s*/);

            _export('compartmentsQueryRegex', compartmentsQueryRegex);

            _export('regionsQueryRegex', regionsQueryRegex = /^regions\(\)\s*/);

            _export('regionsQueryRegex', regionsQueryRegex);

            _export('namespacesQueryRegex', namespacesQueryRegex = /^namespaces\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/);

            _export('namespacesQueryRegex', namespacesQueryRegex);

            _export('resourcegroupsQueryRegex', resourcegroupsQueryRegex = /^resourcegroups\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/);

            _export('resourcegroupsQueryRegex', resourcegroupsQueryRegex);

            _export('metricsQueryRegex', metricsQueryRegex = /^metrics\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/);

            _export('metricsQueryRegex', metricsQueryRegex);

            _export('dimensionKeysQueryRegex', dimensionKeysQueryRegex = /^dimensions\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/);

            _export('dimensionKeysQueryRegex', dimensionKeysQueryRegex);

            _export('dimensionValuesQueryRegex', dimensionValuesQueryRegex = /^dimensionOptions\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*\)/);

            _export('dimensionValuesQueryRegex', dimensionValuesQueryRegex);

            _export('windowsAndResolutionRegex', windowsAndResolutionRegex = /^[0-9]+[mhs]$/);

            _export('windowsAndResolutionRegex', windowsAndResolutionRegex);

            _export('removeQuotes', removeQuotes = function removeQuotes(str) {
                if (!str) return str;

                var res = str;
                if (str.startsWith("'") || str.startsWith('"')) {
                    res = res.slice(1);
                }
                if (str.endsWith("'") || str.endsWith('"')) {
                    res = res.slice(0, res.length - 1);
                }
                return res;
            });

            _export('removeQuotes', removeQuotes);

            _export('SEVEN_DAYS', SEVEN_DAYS = 7);

            _export('SEVEN_DAYS', SEVEN_DAYS);

            _export('THIRTY_DAYS', THIRTY_DAYS = 30);

            _export('THIRTY_DAYS', THIRTY_DAYS);

            _export('NINETY_DAYS', NINETY_DAYS = 90);

            _export('NINETY_DAYS', NINETY_DAYS);

            _export('d0To7Config', d0To7Config = { window: '1m', resolution: '1m' });

            _export('d0To7Config', d0To7Config);

            _export('d8To30Config', d8To30Config = { window: '5m', resolution: '5m' });

            _export('d8To30Config', d8To30Config);

            _export('d31toInfConfig', d31toInfConfig = { window: '1h', resolution: '1h' });

            _export('d31toInfConfig', d31toInfConfig);

            _export('autoTimeIntervals', autoTimeIntervals = [[SEVEN_DAYS, d0To7Config], [THIRTY_DAYS, d8To30Config], [NINETY_DAYS, d31toInfConfig]]);

            _export('autoTimeIntervals', autoTimeIntervals);
        }
    };
});
//# sourceMappingURL=constants.js.map
