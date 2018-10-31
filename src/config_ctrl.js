import { regions, environments } from './constants'

export class OCIConfigCtrl {
  /** @ngInject */
  constructor ($scope, backendSrv) {
    this.backendSrv = backendSrv
    this.tenancyOCID = this.current.jsonData.tenancyOCID
    this.defaultRegion = this.current.jsonData.defaultRegion
    this.environment = this.current.jsonData.environment
  }

  getRegions () {
    return regions
  }

  getEnvironments () {
    return environments
  }
}

OCIConfigCtrl.templateUrl = 'partials/config.html'
