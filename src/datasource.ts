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


/**
 * The OCIDataSource class extends the DataSourceWithBackend class to provide
 * functionality for interacting with Oracle Cloud Infrastructure (OCI) metrics.
 * It includes methods for filtering queries, formatting compartment values,
 * applying template variables, and executing queries for template variable values.
 * 
 * @extends DataSourceWithBackend<OCIQuery, OCIDataSourceOptions>
 */
export class OCIDataSource extends DataSourceWithBackend<OCIQuery, OCIDataSourceOptions> {
  private jsonData: any;
  ocidCompartmentStore: Record<string, string> ={}

    /**
   * Constructor for the OCIDataSource class.
   *
   * @param {DataSourceInstanceSettings<OCIDataSourceOptions>} instanceSettings - The settings for the data source instance.
   */
  constructor(instanceSettings: DataSourceInstanceSettings<OCIDataSourceOptions>) {
    super(instanceSettings);
    this.jsonData = instanceSettings.jsonData;
  }
 

  /**
   * Filters disabled/hidden queries.
   *
   * @param {OCIQuery} query - The query to filter.
   * @returns {boolean} True if the query is not hidden, false otherwise.
   */
  filterQuery(query: OCIQuery): boolean {
    if (query.hide) {
      return false;
    }
    return true;
  }

  /**
   * Formats compartment values, resolving names to OCIDs if available in the store.
   *
   * @param {string} value - The compartment name or OCID.
   * @returns {string} The resolved compartment OCID or the original value if not found.
   */
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
   * Applies template variables to the query, interpolating values and building the MQL query.
   *
   * @param {OCIQuery} query - The query object to apply variables to.
   * @param {ScopedVars} scopedVars - The scoped variables to use for interpolation.
   * @returns {OCIQuery} The query object with template variables applied.
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


  /**
   * Interpolates properties of an object using template variables.
   *
   * @param {T} object - The object whose properties to interpolate.
   * @param {ScopedVars} [scopedVars={}] - The scoped variables to use for interpolation.
   * @returns {T} The object with interpolated properties.
   */
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
  /**
   * Executes a query for template variable values and returns the results.
   *
   * @param {any} query - The query string or object.
   * @param {any} [options] - Optional query options.
   * @returns {Promise<MetricFindValue[]>} A promise that resolves to an array of MetricFindValue objects.
   */
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

  /**
   * Gets the JSON data associated with this data source.
   *
   * @returns {any} The JSON data.
   */
  getJsonData() {
    return this.jsonData;
  }
  
  /**
   * Gets the list of variable names.
   *
   * @returns {string[]} An array of variable names with '$' at the beginning.
   */
  getVariables() {
    const templateSrv = getTemplateSrv();
    return templateSrv.getVariables().map((v) => `$${v.name}`);
  }

  /**
   * Gets the raw list of variables.
   *
   * @returns {any[]} An array of raw variable objects.
   */
  getVariablesRaw() {
    const templateSrv = getTemplateSrv();
    return templateSrv.getVariables();
  }  


   // **************************** Template variables helpers ****************************
  /**
   * Checks if a given name is a variable.
   *
   * @param {string} varName - The name to check, expected to contain '$'.
   * @returns {boolean} True if the name is a variable, false otherwise.
   */
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
   * Checks if a given name is a variable.
   *
   * @param {string} varName - The name to check, expected to contain '$'.
   * @returns {boolean} True if the name is a variable, false otherwise.
   */
  isVariable(varName: string) {
    const varNames = this.getVariables() || [];
    return !!varNames.find((item) => item === varName);
  }


  /**
   * Calls the backend to fetch a resource.
   *
   * @param {string} path - The path of the resource to fetch.
   * @returns {Promise<any>} A promise that resolves to the resource data.
   */
  async getResource(path: string): Promise<any> {
    return super.getResource(path);
  }

  /**
   * Calls the backend to post data to a resource.
   *
   * @param {string} path - The path of the resource.
   * @param {any} body - The request body.
   * @returns {Promise<any>} A promise that resolves to the response data.
   */
  async postResource(path: string, body: any): Promise<any> {
    return super.postResource(path, body);
  }


  /**
   * Retrieves a list of tenancies from the OCI (Oracle Cloud Infrastructure).
   *
   * @returns {Promise<OCIResourceItem[]>} A promise that resolves to an array of OCIResourceItem objects representing the tenancies.
   */
  async getTenancies(): Promise<OCIResourceItem[]> {
    return this.getResource(OCIResourceCall.Tenancies).then((response) => {
      return new ResponseParser().parseTenancies(response);
    });
  }

  /**
   * Retrieves the list of subscribed regions for a given tenancy.
   *
   * @param tenancy - The tenancy identifier. If the tenancy is a variable, it will be interpolated.
   * @returns A promise that resolves to an array of subscribed region names.
   *
   * @throws Will return an empty array if the tenancy is an empty string.
   */
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

  /**
   * Retrieves the compartments for a given tenancy.
   *
   * @param {string} tenancy - The tenancy OCID or a variable representing the tenancy.
   * @returns {Promise<OCIResourceItem[]>} A promise that resolves to an array of OCIResourceItem objects representing the compartments.
   *
   * This method first checks if the provided tenancy is a variable and interpolates its value if necessary.
   * If the tenancy is an empty string, it returns an empty array.
   * Otherwise, it sends a request to retrieve the compartments for the specified tenancy and parses the response.
   */
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

  /**
   * Retrieves namespaces with their associated metric names for a given tenancy, compartment, and region.
   * 
   * @param tenancy - The tenancy OCID or a variable representing the tenancy.
   * @param compartment - The compartment OCID or a variable representing the compartment.
   * @param region - The region name or a variable representing the region.
   * @returns A promise that resolves to an array of OCINamespaceWithMetricNamesItem objects.
   * 
   * The function interpolates the tenancy, compartment, and region if they are variables.
   * If the tenancy is empty, or the region is undefined or a placeholder, it returns an empty array.
   * If the compartment is undefined or a placeholder, it sets the compartment to an empty string.
   * 
   * The function sends a POST request to the OCI Namespaces resource with the tenancy, compartment, and region
   * in the request body, and parses the response to extract the namespaces with their metric names.
   */
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


  /**
   * Retrieves resource groups along with their metric names for a given tenancy, compartment, region, and namespace.
   * 
   * @param tenancy - The tenancy identifier, which can be a variable.
   * @param compartment - The compartment identifier, which can be a variable.
   * @param region - The region identifier, which can be a variable.
   * @param namespace - The namespace identifier, which can be a variable.
   * @returns A promise that resolves to an array of OCIResourceGroupWithMetricNamesItem objects.
   * 
   * The function interpolates the provided parameters if they are variables. If any of the required parameters
   * (tenancy, region, namespace) are missing or placeholders, it returns an empty array. Otherwise, it constructs
   * a request body and makes a POST request to retrieve the resource groups and their metric names.
   */
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

  /**
   * Retrieves the dimensions for a specified metric in Oracle Cloud Infrastructure (OCI).
   *
   * @param tenancy - The tenancy identifier, which can be a variable.
   * @param compartment - The compartment identifier, which can be a variable.
   * @param region - The region identifier, which can be a variable.
   * @param namespace - The namespace of the metric, which can be a variable.
   * @param metricName - The name of the metric, which can be a variable.
   * @returns A promise that resolves to an array of OCIResourceMetadataItem objects representing the dimensions of the specified metric.
   *
   * The function interpolates the provided parameters if they are variables, and then constructs a request body to fetch the dimensions
   * from the OCI resource. If any required parameter is missing or invalid, it returns an empty array.
   */
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


  /**
   * Retrieves tags for a specified OCI resource.
   * WARNING: This function is not yet implemented.
   *
   * @param tenancy - The tenancy identifier.
   * @param compartment - The compartment identifier.
   * @param compartmentName - The name of the compartment.
   * @param region - The region identifier.
   * @param namespace - The namespace identifier.
   * @returns A promise that resolves to an array of OCIResourceMetadataItem objects.
   */
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
