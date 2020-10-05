'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.resolveAutoWinRes = exports.getWindowAndResolution = undefined;

var _constants = require('../constants');

/** getWindowAndResolution
 *
 * @param autoWinResConfig is an array of Object with length always greater than 1,
 * i.e config array should contain at least 1 object
 * @param timeRange
 * @returns {{window, resolution}}
 */
var getWindowAndResolution = exports.getWindowAndResolution = function getWindowAndResolution(autoWinResConfig, timeRange) {
  var i = -1;
  do {
    i++;
  } while (i < autoWinResConfig.length - 1 && timeRange > autoWinResConfig[i][0]);
  var _autoWinResConfig$i$ = autoWinResConfig[i][1],
      window = _autoWinResConfig$i$.window,
      resolution = _autoWinResConfig$i$.resolution;

  return { window: window, resolution: resolution };
};

/** resolveAutoWinRes
 *
 * @param windowSelected
 * @param resolutionSelected
 * @param timeRangeSelected
 * @returns {{window: *, resolution: *}}
 */
var resolveAutoWinRes = exports.resolveAutoWinRes = function resolveAutoWinRes(windowSelected, resolutionSelected, timeRangeSelected) {
  var result = { window: windowSelected, resolution: resolutionSelected };
  if (windowSelected !== _constants.AUTO && resolutionSelected !== _constants.AUTO) return result;

  var _getWindowAndResoluti = getWindowAndResolution(_constants.autoTimeIntervals, timeRangeSelected),
      window = _getWindowAndResoluti.window,
      resolution = _getWindowAndResoluti.resolution;

  if (windowSelected === _constants.AUTO) result.window = window;
  if (resolutionSelected === _constants.AUTO) result.resolution = resolution;
  return result;
};
//# sourceMappingURL=utilFunctions.js.map
