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
