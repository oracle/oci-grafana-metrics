'use strict';

System.register(['../constants'], function (_export, _context) {
  "use strict";

  var AUTO, autoTimeIntervals, getWindowAndResolution, resolveAutoWinRes;
  return {
    setters: [function (_constants) {
      AUTO = _constants.AUTO;
      autoTimeIntervals = _constants.autoTimeIntervals;
    }],
    execute: function () {
      _export('getWindowAndResolution', getWindowAndResolution = function getWindowAndResolution(autoWinResConfig, timeRange) {
        var i = -1;
        do {
          i++;
        } while (i < autoWinResConfig.length - 1 && timeRange > autoWinResConfig[i][0]);
        var _autoWinResConfig$i$ = autoWinResConfig[i][1],
            window = _autoWinResConfig$i$.window,
            resolution = _autoWinResConfig$i$.resolution;

        return { window: window, resolution: resolution };
      });

      _export('getWindowAndResolution', getWindowAndResolution);

      _export('resolveAutoWinRes', resolveAutoWinRes = function resolveAutoWinRes(windowSelected, resolutionSelected, timeRangeSelected) {
        var result = { window: windowSelected, resolution: resolutionSelected };
        if (windowSelected !== AUTO && resolutionSelected !== AUTO) return result;

        var _getWindowAndResoluti = getWindowAndResolution(autoTimeIntervals, timeRangeSelected),
            window = _getWindowAndResoluti.window,
            resolution = _getWindowAndResoluti.resolution;

        if (windowSelected === AUTO) result.window = window;
        if (resolutionSelected === AUTO) result.resolution = resolution;
        return result;
      });

      _export('resolveAutoWinRes', resolveAutoWinRes);
    }
  };
});
//# sourceMappingURL=utilFunctions.js.map
