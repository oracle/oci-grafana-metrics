"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = retryOrThrow;
/*
** Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
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
