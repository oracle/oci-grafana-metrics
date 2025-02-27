/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/

import _,{ isString} from 'lodash';
import { DataSourceInstanceSettings, ScopedVars, MetricFindValue } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import {
  OCIResourceItem,
  OCINamespaceWithMetricNamesItem,
  OCIResourceGroupWithMetricNamesItem,
  ResponseParser,
  OCIResourceMetadataItem,
} from './resource.response.parser';
import {
  OCIDataSourceOptions,
  OCIQuery,
  OCIResourceCall,
  QueryPlaceholder,
  dimensionQueryRegex,
  namespacesQueryRegex,
  resourcegroupsQueryRegex,
  metricsQueryRegex,
  regionsQueryRegex,
  tenanciesQueryRegex,
  DEFAULT_TENANCY,
  compartmentsQueryRegex,
  SetAutoInterval,
} from "./types";
import QueryModel from './query_model';

export class OCIDataSource extends DataSourceWithBackend<OCIQuery, OCIDataSourceOptions> {
  private jsonData: any;
  ocidCompartmentStore: Record<string, string> ={}

  constructor(instanceSettings: DataSourceInstanceSettings<OCIDataSourceOptions>) {
    super(instanceSettings);
    this.jsonData = instanceSettings.jsonData;
  }
 

  /**
   * Filters disabled/hidden queries
   *
   * @param {string} query Query
   */
  filterQuery(query: OCIQuery): boolean {
    if (query.hide) {
      return false;
    }
    return true;
  }

  compartmentFormatter = (value: string): string => {
    // if (typeof value === 'string') {
    //   return value;
    // }
    if (this.ocidCompartmentStore[value] || this.isVariable(value)) {
      return this.ocidCompartmentStore[value]
    } else {
      return value
    }
  };

  /**
   * Override to apply template variables
   *
   * @param {string} query Query
   * @param {ScopedVars} scopedVars Scoped variables
   */
  applyTemplateVariables(query: OCIQuery, scopedVars: ScopedVars) {
    const templateSrv = getTemplateSrv();
    const interpolatedQ = _.cloneDeep(query);

    const TimeStart = parseInt(getTemplateSrv().replace("${__from}"), 10)
    const TimeEnd  = parseInt(getTemplateSrv().replace("${__to}"), 10)
    if (this.isVariable(interpolatedQ.interval)) {
      interpolatedQ.interval = templateSrv.replace(interpolatedQ.interval, scopedVars);
    }
    if (interpolatedQ.interval === QueryPlaceholder.Interval || interpolatedQ.interval === "auto" || interpolatedQ.interval === undefined){
      interpolatedQ.interval = SetAutoInterval(TimeStart, TimeEnd);
    }
    interpolatedQ.region = templateSrv.replace(interpolatedQ.region, scopedVars);
    interpolatedQ.tenancy = templateSrv.replace(interpolatedQ.tenancy, scopedVars);
    interpolatedQ.compartment = templateSrv.replace(interpolatedQ.compartment, scopedVars, this.compartmentFormatter);
    interpolatedQ.namespace = templateSrv.replace(interpolatedQ.namespace, scopedVars);
    interpolatedQ.resourcegroup = templateSrv.replace(interpolatedQ.resourcegroup, scopedVars);
    interpolatedQ.metric = templateSrv.replace(interpolatedQ.metric, scopedVars);
    interpolatedQ.queryTextRaw = templateSrv.replace(interpolatedQ.queryTextRaw, scopedVars);

    if (interpolatedQ.dimensionValues) {
      for (let i = 0; i < interpolatedQ.dimensionValues.length; i++) {
        interpolatedQ.dimensionValues[i] = templateSrv.replace(interpolatedQ.dimensionValues[i], scopedVars);
      }
    }
    if (interpolatedQ.tenancy) {
      interpolatedQ.tenancy = templateSrv.replace(interpolatedQ.tenancy, scopedVars);
    }
    if (interpolatedQ.compartment) {
      interpolatedQ.compartment = templateSrv.replace(interpolatedQ.compartment, scopedVars, this.compartmentFormatter);
    }
    if (interpolatedQ.resourcegroup) {
      interpolatedQ.resourcegroup = templateSrv.replace(interpolatedQ.resourcegroup, scopedVars);
    }
    
    const queryModel = new QueryModel(interpolatedQ, getTemplateSrv());
    if (queryModel.isQueryReady()) {
      if (interpolatedQ.rawQuery === false && interpolatedQ.queryTextRaw !== '') {
        interpolatedQ.queryTextRaw = templateSrv.replace(interpolatedQ.queryTextRaw, scopedVars);
        interpolatedQ.queryText = queryModel.buildQuery(String(interpolatedQ.queryTextRaw));
      } else {
        interpolatedQ.queryText = queryModel.buildQuery(String(interpolatedQ.metric));
      }
      
    }    
    return interpolatedQ;
  }


  interpolateProps<T extends Record<string, any>>(object: T, scopedVars: ScopedVars = {}): T {
    const templateSrv = getTemplateSrv();
    return Object.entries(object).reduce((acc: any, [key, value]) => {
      if (value && isString(value)) {
        const formatter = key === "compartment" ? this.compartmentFormatter : undefined;
        acc[key] = templateSrv.replace(value, scopedVars, formatter);
      } else {
        acc[key] = value;
      }
      return acc as T;
    }, {});
  }

  // // **************************** Template variable helpers ****************************

  // /**
  //  * Matches the regex from creating template variables and returns options for the corresponding variable.
  //  * Example:
  //  * template variable with the query "regions()" will be matched with the regionsQueryRegex and list of available regions will be returned.
  //  */

  async metricFindQuery?(query: any, options?: any): Promise<MetricFindValue[]> {
    const templateSrv = getTemplateSrv();

    const tenancyQuery = query.match(tenanciesQueryRegex);
    if (tenancyQuery) {
      const tenancy = await this.getTenancies();
      return tenancy.map(n => {
        return { text: n.name, value: n.ocid };
      });   
    }    

    const regionQuery = query.match(regionsQueryRegex);
    if (regionQuery) {
      if (this.jsonData.tenancymode === "multitenancy") {
        const tenancy = templateSrv.replace(regionQuery[1]);
        const regions = await this.getSubscribedRegions(tenancy);
        return regions.map(n => {
          return { text: n, value: n };
        });
      } else {     
        const regions = await this.getSubscribedRegions(DEFAULT_TENANCY);
        return regions.map(n => {
          return { text: n, value: n };
        });       
      }
    }

    const compartmentQuery = query.match(compartmentsQueryRegex);
    if (compartmentQuery){
      if (this.jsonData.tenancymode === "multitenancy") {
        const tenancy = templateSrv.replace(compartmentQuery[1]);
        const compartments = await this.getCompartments(tenancy);
        return compartments.map(n => {
          this.ocidCompartmentStore[n.name]=n.ocid; 
          return { text: n.name, value: n.name };
        });
      } else {
        const compartments = await this.getCompartments(DEFAULT_TENANCY);
        return compartments.map(n => {
          this.ocidCompartmentStore[n.name]=n.ocid; 
          return { text: n.name, value: n.name };
        }); 
      }   
    }    


    const namespaceQuery = query.match(namespacesQueryRegex);
    if (namespaceQuery) {
      if (this.jsonData.tenancymode === "multitenancy") {
        const tenancy = templateSrv.replace(namespaceQuery[1]);
        const region = templateSrv.replace(namespaceQuery[2]);
        const compartment = templateSrv.replace(namespaceQuery[3], undefined, this.compartmentFormatter);
        const namespaces = await this.getNamespacesWithMetricNames(tenancy, compartment, region);
        return namespaces.map(n => {
          return { text: n.namespace, value: n.namespace };
        });        
      } else {
        const tenancy = DEFAULT_TENANCY;
        const region = templateSrv.replace(namespaceQuery[1]);
        const compartment = templateSrv.replace(namespaceQuery[2], undefined, this.compartmentFormatter);
        const namespaces = await this.getNamespacesWithMetricNames(tenancy, compartment, region);
        return namespaces.map(n => {
          return { text: n.namespace, value: n.namespace };
        });      
      }
    }

    let resourcegroupQuery = query.match(resourcegroupsQueryRegex);
    if (resourcegroupQuery) {
      if (this.jsonData.tenancymode === "multitenancy") {
        const tenancy = templateSrv.replace(resourcegroupQuery[1]);
        const region = templateSrv.replace(resourcegroupQuery[2]);
        const compartment = templateSrv.replace(resourcegroupQuery[3], undefined, this.compartmentFormatter);
        const namespace = templateSrv.replace(resourcegroupQuery[4]);
        const resource_group = await this.getResourceGroupsWithMetricNames(tenancy, compartment, region, namespace);
        return resource_group.map(n => {
          return { text: n.resource_group, value: n.resource_group };
        });
      } else {
        const tenancy = DEFAULT_TENANCY;
        const region = templateSrv.replace(resourcegroupQuery[1]);
        const compartment = templateSrv.replace(resourcegroupQuery[2], undefined, this.compartmentFormatter);
        const namespace = templateSrv.replace(resourcegroupQuery[3]);
        const resource_group = await this.getResourceGroupsWithMetricNames(tenancy, compartment, region, namespace);
        return resource_group.map(n => {
          return { text: n.resource_group, value: n.resource_group };
        });     
      }
    }

    const metricQuery = query.match(metricsQueryRegex);
    if (metricQuery) {
      if (this.jsonData.tenancymode === "multitenancy") {
        const tenancy = templateSrv.replace(metricQuery[1]);
        const region = templateSrv.replace(metricQuery[2]);
        const compartment = templateSrv.replace(metricQuery[3], undefined, this.compartmentFormatter);
        const namespace = templateSrv.replace(metricQuery[4]);
        // const resourcegroup = templateSrv.replace(metricQuery[4]);
        const metric_names = await this.getResourceGroupsWithMetricNames(tenancy, compartment, region, namespace);
        return metric_names.flatMap(n => {
          return n.metric_names.map(name => {
            return { text: name, value: name };
          });
        });        
      } else {
        const tenancy = DEFAULT_TENANCY;
        const region = templateSrv.replace(metricQuery[1]);
        const compartment = templateSrv.replace(metricQuery[2], undefined, this.compartmentFormatter);
        const namespace = templateSrv.replace(metricQuery[3]);
        // const resource_group = templateSrv.replace(metricQuery[4]);
        const metric_names = await this.getResourceGroupsWithMetricNames(tenancy, compartment, region, namespace); 
        return metric_names.flatMap(n => {
          return n.metric_names.map(name => {
            return { text: name, value: name };
          });
        });       
      }  
    }    

    const dimensionsQuery = query.match(dimensionQueryRegex);
    if (dimensionsQuery) {
      if (this.jsonData.tenancymode === "multitenancy") {
        const tenancy = templateSrv.replace(dimensionsQuery[1]);
        const region = templateSrv.replace(dimensionsQuery[2]);
        const compartment = templateSrv.replace(dimensionsQuery[3], undefined, this.compartmentFormatter);
        const namespace = templateSrv.replace(dimensionsQuery[4]);
        const metric = templateSrv.replace(dimensionsQuery[5]);
        const dimension_values = await this.getDimensions(tenancy, compartment, region, namespace, metric);
        return dimension_values.flatMap(res => {
          return res.values.map(val => {
              return { text: res.key + ' - ' + val, value: res.key + '="' + val + '"' };
          });
        }); 
      } else {
        const tenancy = DEFAULT_TENANCY;
        const region = templateSrv.replace(dimensionsQuery[1]);
        const compartment = templateSrv.replace(dimensionsQuery[2], undefined, this.compartmentFormatter);
        const namespace = templateSrv.replace(dimensionsQuery[3]);
        const metric = templateSrv.replace(dimensionsQuery[4]);
        const dimension_values = await this.getDimensions(tenancy, compartment, region, namespace, metric);
        return dimension_values.flatMap(res => {
          return res.values.map(val => {
              return { text: res.key + ' - ' + val, value: res.key + '="' + val + '"' };
          });
        }); 
      }      
    } 

    return [];
  }


  getJsonData() {
    return this.jsonData;
  }
  
  getVariables() {
    const templateSrv = getTemplateSrv();
    return templateSrv.getVariables().map((v) => `$${v.name}`);
  }

  getVariablesRaw() {
    const templateSrv = getTemplateSrv();
    return templateSrv.getVariables();
  }  


 // **************************** Template variables helpers ****************************

  /**
   * List all variable names optionally filtered by regex or/and type
   * Returns list of names with '$' at the beginning. Example: ['$dimensionKey', '$dimensionValue']
   *
   * Updates:
   * Notes on implementation :
   * If a custom or constant is in  variables and  includeCustom, default is false.
   * Hence,the varDescriptors list is filtered for a unique set of var names
   */

  /**
   * @param varName valid varName contains '$'. Example: '$dimensionKey'
   * Returns true if variable with the given name is found
   */
  isVariable(varName: string) {
    const varNames = this.getVariables() || [];
    return !!varNames.find((item) => item === varName);
  }


  // main caller to call resource handler for get call
  async getResource(path: string): Promise<any> {
    return super.getResource(path);
  }
  // main caller to call resource handler for post call
  async postResource(path: string, body: any): Promise<any> {
    return super.postResource(path, body);
  }


  async getTenancies(): Promise<OCIResourceItem[]> {
    return this.getResource(OCIResourceCall.Tenancies).then((response) => {
      return new ResponseParser().parseTenancies(response);
    });
  }

  async getSubscribedRegions(tenancy: string): Promise<string[]> {
    if (this.isVariable(tenancy)) {
      let { tenancy: var_tenancy} = this.interpolateProps({tenancy});
      if (var_tenancy !== "") { 
        tenancy = var_tenancy
      }      
    }
    if (tenancy === '') {
      return [];
    }
    const reqBody: JSON = {
      tenancy: tenancy,
    } as unknown as JSON;
    return this.postResource(OCIResourceCall.Regions, reqBody).then((response) => {
      return new ResponseParser().parseRegions(response);
    });
  }

  async getCompartments(tenancy: string): Promise<OCIResourceItem[]> {
    if (this.isVariable(tenancy)) {
      let { tenancy: var_tenancy} = this.interpolateProps({tenancy});
      if (var_tenancy !== "") { 
        tenancy = var_tenancy
      }      
    }   
    if (tenancy === '') {
      return [];
    }
    const reqBody: JSON = {
      tenancy: tenancy,
    } as unknown as JSON;
    return this.postResource(OCIResourceCall.Compartments, reqBody).then((response) => {
      return new ResponseParser().parseCompartments(response);
    });
  }

  async getNamespacesWithMetricNames(
    tenancy: string,
    compartment: any,
    region: any
  ): Promise<OCINamespaceWithMetricNamesItem[]> {
    if (this.isVariable(tenancy)) {
      let { tenancy: var_tenancy} = this.interpolateProps({tenancy});
      if (var_tenancy !== "") { 
        tenancy = var_tenancy
      }      
    }

    if (this.isVariable(compartment)) {
      let { compartment: var_compartment} = this.interpolateProps({compartment});
      if (var_compartment !== "") { 
        compartment = var_compartment
      }      
    }

    if (this.isVariable(region)) {
      let { region: var_region} = this.interpolateProps({region});
      if (var_region !== "") { 
        region = var_region
      }      
    }

    if (tenancy === '') {
      return [];
    }
    if (region === undefined || region === QueryPlaceholder.Region) {
      return [];
    }

    if (compartment === undefined || compartment === QueryPlaceholder.Compartment) {
      compartment = '';
    }

    const reqBody: JSON = {
      tenancy: tenancy,
      compartment: compartment,
      region: region,
    } as unknown as JSON;
    return this.postResource(OCIResourceCall.Namespaces, reqBody).then((response) => {
      return new ResponseParser().parseNamespacesWithMetricNames(response);
    });
  }


  async getResourceGroupsWithMetricNames(
    tenancy: any,
    compartment: any,
    region: any,
    namespace: any
  ): Promise<OCIResourceGroupWithMetricNamesItem[]> {

    if (this.isVariable(tenancy)) {
      let { tenancy: var_tenancy} = this.interpolateProps({tenancy});
      if (var_tenancy !== "") { 
        tenancy = var_tenancy
      }      
    }

    if (this.isVariable(compartment)) {
      let { compartment: var_compartment} = this.interpolateProps({compartment});
      if (var_compartment !== "") { 
        compartment = var_compartment
      }      
    }

    if (this.isVariable(region)) {
      let { region: var_region} = this.interpolateProps({region});
      if (var_region !== "") { 
        region = var_region
      }      
    }

    if (this.isVariable(namespace)) {
      let { namespace: var_namespace} = this.interpolateProps({namespace});
      if (var_namespace !== "") { 
        namespace = var_namespace
      }      
    }    


    if (tenancy === '') {
      return [];
    }
    if (region === undefined || region === QueryPlaceholder.Region) {
      return [];
    }

    if (compartment === undefined || compartment === QueryPlaceholder.Compartment) {
      compartment = '';
    } 

    if (region === QueryPlaceholder.Region || namespace === QueryPlaceholder.Namespace) {
      return [];
    }


    const reqBody: JSON = {
      tenancy: tenancy,
      compartment: compartment,
      region: region,
      namespace: namespace,
    } as unknown as JSON;
    return this.postResource(OCIResourceCall.ResourceGroups, reqBody).then((response) => {
      return new ResponseParser().parseResourceGroupWithMetricNames(response);
    });
  }

  async getDimensions(
    tenancy: any,
    compartment: any,
    region: any,
    namespace: any,
    metricName: any
  ): Promise<OCIResourceMetadataItem[]> {

    if (this.isVariable(tenancy)) {
      let { tenancy: var_tenancy} = this.interpolateProps({tenancy});
      if (var_tenancy !== "") { 
        tenancy = var_tenancy
      }      
    }

    if (this.isVariable(compartment)) {
      let { compartment: var_compartment} = this.interpolateProps({compartment});
      if (var_compartment !== "") { 
        compartment = var_compartment
      }      
    }

    if (this.isVariable(region)) {
      let { region: var_region} = this.interpolateProps({region});
      if (var_region !== "") { 
        region = var_region
      }      
    }

    if (this.isVariable(namespace)) {
      let { namespace: var_namespace} = this.interpolateProps({namespace});
      if (var_namespace !== "") { 
        namespace = var_namespace
      }      
    }

    if (this.isVariable(metricName)) {
      let { metricName: var_metric} = this.interpolateProps({metricName});
      if (var_metric !== "") { 
        metricName = var_metric
      }      
    }       

    if (tenancy === '') {
      return [];
    }
    if (region === undefined || namespace === undefined || metricName === undefined) {
      return [];
    }
    if (
      region === QueryPlaceholder.Region ||
      namespace === QueryPlaceholder.Namespace ||
      metricName === QueryPlaceholder.Metric
    ) {
      return [];
    }

    if (compartment === undefined || compartment === QueryPlaceholder.Compartment) {
      compartment = '';
    }

    const reqBody: JSON = {
      tenancy: tenancy,
      compartment: compartment,
      region: region,
      namespace: namespace,
      metric_name: metricName,
    } as unknown as JSON;
    return this.postResource(OCIResourceCall.Dimensions, reqBody).then((response) => {
      return new ResponseParser().parseDimensions(response);
    });
  }
  async getTags(
    tenancy: any,
    compartment: any,
    compartmentName: any,
    region: any,
    namespace: any
  ): Promise<OCIResourceMetadataItem[]> {
    if (tenancy === '') {
      return [];
    }
    if (region === undefined || namespace === undefined) {
      return [];
    }
    if (region === QueryPlaceholder.Region || namespace === QueryPlaceholder.Namespace) {
      return [];
    }

    if (compartment === undefined || compartment === QueryPlaceholder.Compartment) {
      compartment = '';
    }
    if (compartmentName === undefined) {
      compartmentName = '';
    }

    const reqBody: JSON = {
      tenancy: tenancy,
      compartment: compartment,
      compartment_name: compartmentName,
      region: region,
      namespace: namespace,
    } as unknown as JSON;
    return this.postResource(OCIResourceCall.Tags, reqBody).then((response) => {
      return new ResponseParser().parseTags(response);
    });
  }
}
