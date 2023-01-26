/*
** Copyright © 2022 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
import { regions, environments, tenancymodes } from './constants'

export class OCIConfigCtrl {
  /** @ngInject */
  constructor ($scope, backendSrv) {
    this.backendSrv = backendSrv
    this.environment = this.current.jsonData.environment      
    this.tenancymode = this.current.jsonData.tenancymode
    this.region0 = this.current.jsonData.region0
    this.region1 = this.current.jsonData.region1
    this.region2 = this.current.jsonData.region2
    this.region3 = this.current.jsonData.region3
    this.region4 = this.current.jsonData.region4
    this.region5 = this.current.jsonData.region5
    this.profile0 = this.current.jsonData.profile0
    this.profile1 = this.current.jsonData.profile1
    this.profile2 = this.current.jsonData.profile2
    this.profile3 = this.current.jsonData.profile3
    this.profile4 = this.current.jsonData.profile4
    this.profile5 = this.current.jsonData.profile5
    this.addon1 = this.current.jsonData.addon1
    this.addon2 = this.current.jsonData.addon2
    this.addon3 = this.current.jsonData.addon3
    this.addon4 = this.current.jsonData.addon4

  }

  getRegions () {
    return regions
  }

  getEnvironments () { 
    return environments
  }

  getTenancyModes () { 
    return tenancymodes
  }   

}

OCIConfigCtrl.templateUrl = 'partials/config.html'
