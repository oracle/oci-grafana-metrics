/*
** Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
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

  onChangeInternal() {
    if (this.environment === "multitenancy") {
      this.target.MultiTenancy = true;
    }    
    this.panelCtrl.refresh(); // Asks the panel to refresh data.
  }

}

OCIConfigCtrl.templateUrl = 'partials/config.html'
