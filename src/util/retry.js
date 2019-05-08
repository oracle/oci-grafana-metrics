/*
** Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
export default function retryOrThrow (actionPromise, maxRetries) {
  let numberOfRetries = 1
  return new Promise((resolve, reject) => {
    action()
    function action () {
      actionPromise()
        .then((response) => {
          resolve(response)
        })
        .catch((error) => {
          if (numberOfRetries >= maxRetries) {
            return reject(new Error(`reject: too many failed attempts: ${JSON.stringify(error)}`))
          }
          let delay = Math.pow(2, numberOfRetries) + Math.floor(Math.random() * 1000)
          numberOfRetries++
          setTimeout(() => action(), delay)
        })
    }
  })
}
