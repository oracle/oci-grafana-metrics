"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = retryOrThrow;
function retryOrThrow(actionPromise, maxRetries) {
  var numberOfRetries = 1;
  return new Promise(function (resolve, reject) {
    action();
    function action() {
      actionPromise().then(function (response) {
        resolve(response);
      }).catch(function (error) {
        if (numberOfRetries >= maxRetries) {
          return reject(new Error("reject: too many failed attempts: " + JSON.stringify(error)));
        }
        var delay = Math.pow(2, numberOfRetries) + Math.floor(Math.random() * 1000);
        numberOfRetries++;
        setTimeout(function () {
          return action();
        }, delay);
      });
    }
  });
}
//# sourceMappingURL=retry.js.map
