import { Observable } from 'rxjs';
import _,{ isString} from 'lodash';
import { DataSourceInstanceSettings, DataQueryRequest, DataQueryResponse, ScopedVars, MetricFindValue } from '@grafana/data';
// import { DataSourceInstanceSettings, ScopedVars, MetricFindValue } from '@grafana/data';

import { DataSourceWithBackend, TemplateSrv, getTemplateSrv } from '@grafana/runtime';
import { OCIDataSourceOptions, OCIQuery, OCIResourceCall, QueryPlaceholder } from './types';
import {
  OCIResourceItem,
  OCINamespaceWithMetricNamesItem,
  OCIResourceGroupWithMetricNamesItem,
  ResponseParser,
  OCIResourceMetadataItem,
} from './resource.response.parser';
import {
  // aggregations,
  // dimensionKeysQueryRegex,
  namespacesQueryRegex,
  resourcegroupsQueryRegex,
  metricsQueryRegex,
  regionsQueryRegex,
  tenanciesQueryRegex,
  DEFAULT_TENANCY,
  compartmentsQueryRegex,
  // dimensionValuesQueryRegex,
  // removeQuotes,
  // AUTO,
} from "./constants";
// import _ from 'lodash';
import QueryModel from './query_model';


export class OCIDataSource extends DataSourceWithBackend<OCIQuery, OCIDataSourceOptions> {
  private jsonData: any;
  // private backendSrv: BackendSrv;
  // private templateSrv: TemplateSrv;

  constructor(instanceSettings: DataSourceInstanceSettings<OCIDataSourceOptions>, private readonly templateSrv: TemplateSrv = getTemplateSrv()) {
    super(instanceSettings);
    this.jsonData = instanceSettings.jsonData;

    // this.backendSrv = getBackendSrv();
    // this.templateSrv = getTemplateSrv();

  }

  /**
   * Override to apply template variables
   *
   * @param {string} query Query
   * @param {ScopedVars} scopedVars Scoped variables
   */

  query(options: DataQueryRequest<OCIQuery>): Observable<DataQueryResponse> {
    console.log("query(options: ")

    return super.query(options);
  }


  /**
   * Override to apply template variables
   *
   * @param {string} query Query
   * @param {ScopedVars} scopedVars Scoped variables
   */
  applyTemplateVariables(query: OCIQuery, scopedVars: ScopedVars) {
    const templateSrv = getTemplateSrv();
    console.log("applyTemplateVariables: before region: " + query.region)
    console.log("applyTemplateVariables: before compartmentOCID: " + query.compartmentOCID)
    console.log("applyTemplateVariables: before tenancyOCID: " + query.tenancyOCID)
    console.log("applyTemplateVariables: before namespace: " + query.namespace)
    console.log("applyTemplateVariables: before resourceGroup: " + query.resourceGroup)
    console.log("applyTemplateVariables: before metric: " + query.metric)

    query.region = templateSrv.replace(query.region, scopedVars);
    query.tenancyOCID = templateSrv.replace(query.tenancyOCID, scopedVars);
    query.compartmentOCID = templateSrv.replace(query.compartmentOCID, scopedVars);
    query.namespace = templateSrv.replace(query.namespace, scopedVars);
    query.resourceGroup = templateSrv.replace(query.resourceGroup, scopedVars);
    query.metric = templateSrv.replace(query.metric, scopedVars);
    
    const queryModel = new QueryModel(query, getTemplateSrv());
    if (queryModel.isQueryReady()) {
      query.queryText = queryModel.buildQuery(String(query.metric));
    }
    
    console.log("applyTemplateVariables: after region: " + query.region)
    console.log("applyTemplateVariables: after compartmentOCID: " + query.compartmentOCID)
    console.log("applyTemplateVariables: after tenancyOCID: " + query.tenancyOCID)
    console.log("applyTemplateVariables: after namespace: " + query.namespace)
    console.log("applyTemplateVariables: after resourceGroup: " + query.resourceGroup)
    console.log("applyTemplateVariables: after metric: " + query.metric)

    return query;
  }


  interpolateProps<T extends Record<string, any>>(object: T, scopedVars: ScopedVars = {}): T {
    const templateSrv = getTemplateSrv();
    return Object.entries(object).reduce((acc, [key, value]) => {
      return {
        ...acc,
        [key]: value && isString(value) ? templateSrv.replace(value, scopedVars) : value,
      };
    }, {} as T);
  }

  // // **************************** Template variable helpers ****************************

  // /**
  //  * Matches the regex from creating template variables and returns options for the corresponding variable.
  //  * Example:
  //  * template variable with the query "regions()" will be matched with the regionsQueryRegex and list of available regions will be returned.
  //  */
  // templateMetricQuery(query: OCIQuery,varString: string) {


  //   let dimensionsQuery = varString.match(dimensionKeysQueryRegex);
  //   if (dimensionsQuery) {
  //     if (this.tenancymode === "multitenancy") {
  //       let target = {
  //         tenancy: removeQuotes(this.getVariableValue(dimensionsQuery[1])),
  //         region: removeQuotes(this.getVariableValue(dimensionsQuery[2])),
  //         compartment: removeQuotes(this.getVariableValue(dimensionsQuery[3])),
  //         namespace: removeQuotes(this.getVariableValue(dimensionsQuery[4])),
  //         metric: removeQuotes(this.getVariableValue(dimensionsQuery[5])),
  //         resourcegroup: removeQuotes(this.getVariableValue(dimensionsQuery[6])),
  //       };
  //       return this.getDimensionKeys(target).catch((err) => {
  //         throw new Error("Unable to get dimensions: " + err);
  //       });
  //     } else {
  //       let target = {
  //         tenancy: DEFAULT_TENANCY,
  //         region: removeQuotes(this.getVariableValue(dimensionsQuery[1])),
  //         compartment: removeQuotes(this.getVariableValue(dimensionsQuery[2])),
  //         namespace: removeQuotes(this.getVariableValue(dimensionsQuery[3])),
  //         metric: removeQuotes(this.getVariableValue(dimensionsQuery[4])),
  //         resourcegroup: removeQuotes(this.getVariableValue(dimensionsQuery[5])),
  //       };
  //       return this.getDimensionKeys(target).catch((err) => {
  //         throw new Error("Unable to get dimensions: " + err);
  //       });        
  //     }      
  //   }

  //   let dimensionOptionsQuery = varString.match(dimensionValuesQueryRegex);
  //   if (dimensionOptionsQuery) {
  //     if (this.tenancymode === "multitenancy") {
  //       let target = {
  //         tenancy: removeQuotes(this.getVariableValue(dimensionOptionsQuery[1])),
  //         region: removeQuotes(this.getVariableValue(dimensionOptionsQuery[2])),
  //         compartment: removeQuotes(this.getVariableValue(dimensionOptionsQuery[3])),
  //         namespace: removeQuotes(this.getVariableValue(dimensionOptionsQuery[4])),
  //         metric: removeQuotes(this.getVariableValue(dimensionOptionsQuery[5])),
  //         resourcegroup: removeQuotes(this.getVariableValue(dimensionOptionsQuery[7])),
  //       };
  //       let dimensionKey = removeQuotes(this.getVariableValue(dimensionOptionsQuery[6]));
  //       return this.getDimensionValues(target, dimensionKey).catch((err) => {
  //         throw new Error("Unable to get dimension options: " + err);
  //       });
  //     } else {
  //       let target = {
  //         tenancy: DEFAULT_TENANCY,
  //         region: removeQuotes(this.getVariableValue(dimensionOptionsQuery[1])),
  //         compartment: removeQuotes(this.getVariableValue(dimensionOptionsQuery[2])),
  //         namespace: removeQuotes(this.getVariableValue(dimensionOptionsQuery[3])),
  //         metric: removeQuotes(this.getVariableValue(dimensionOptionsQuery[4])),
  //         resourcegroup: removeQuotes(this.getVariableValue(dimensionOptionsQuery[6])),
  //       };
  //       let dimensionKey = removeQuotes(this.getVariableValue(dimensionOptionsQuery[5]));        
  //       return this.getDimensionValues(target, dimensionKey).catch((err) => {
  //         throw new Error("Unable to get dimension options: " + err);
  //       });        
  //     }
  //   }

  //   throw new Error("Unable to parse templating string");
  // }

  async metricFindQuery?(query: any, options?: any): Promise<MetricFindValue[]> {
    const templateSrv = getTemplateSrv();
    const tmode = this.getJsonData().tenancymode;
    console.log("uga "+tmode)
    console.log("ciao "+this.jsonData.tenancymode)

    const tenancyQuery = query.match(tenanciesQueryRegex);
    if (tenancyQuery) {
      const tenancy = await this.getTenancies();
      return tenancy.map(n => {
        return { text: n.name, value: n.ocid };
      });   
    }    

    const regionQuery = query.match(regionsQueryRegex);
    if (regionQuery) {
      if (tmode === "multitenancy") {
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
          return { text: n.name, value: n.ocid };
        });
      } else {
        const compartments = await this.getCompartments(DEFAULT_TENANCY);
        return compartments.map(n => {
          return { text: n.name, value: n.ocid };
        }); 
      }   
    }    


    const namespaceQuery = query.match(namespacesQueryRegex);
    if (namespaceQuery) {
      if (this.jsonData.tenancymode === "multitenancy") {
        const tenancy = templateSrv.replace(namespaceQuery[1]);
        const region = templateSrv.replace(namespaceQuery[2]);
        const compartment = templateSrv.replace(namespaceQuery[3]);
        const namespaces = await this.getNamespacesWithMetricNames(tenancy, compartment, region);
        return namespaces.map(n => {
          return { text: n.namespace, value: n.namespace };
        });        
      } else {
        const tenancy = DEFAULT_TENANCY;
        const region = templateSrv.replace(namespaceQuery[1]);
        const compartment = templateSrv.replace(namespaceQuery[2]);
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
        const compartment = templateSrv.replace(resourcegroupQuery[3]);
        const namespace = templateSrv.replace(resourcegroupQuery[4]);
        const resource_group = await this.getResourceGroupsWithMetricNames(tenancy, compartment, region, namespace);
        return resource_group.map(n => {
          return { text: n.resource_group, value: n.resource_group };
        });
      } else {
        const tenancy = DEFAULT_TENANCY;
        const region = templateSrv.replace(resourcegroupQuery[1]);
        const compartment = templateSrv.replace(resourcegroupQuery[2]);
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
        const compartment = templateSrv.replace(metricQuery[3]);
        const namespace = templateSrv.replace(metricQuery[4]);
        // const resource_group = templateSrv.replace(metricQuery[4]);
        const metric_names = await this.getResourceGroupsWithMetricNames(tenancy, compartment, region, namespace);
        return metric_names.flatMap(n => {
          return n.metric_names.map(name => {
            console.log("metric_names "+name)
            return { text: name, value: name };
          });
        });        
      } else {
        const tenancy = DEFAULT_TENANCY;
        const region = templateSrv.replace(metricQuery[1]);
        const compartment = templateSrv.replace(metricQuery[2]);
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


  // targetContainsTemplate(target: any) {
  //   console.log(target)
  //   // console.log(target.tenancyOCID)

  //   if (this.templateSrv.containsTemplate(target)) {
  //     return true;
  //   }    
  //   // if (target.tenancyOCID && target.tenancyOCID.length > 0) {
  //   //   for (let i = 0; i < target.tenancyOCID.length; i++) {
  //   //     if (this.templateSrv.containsTemplate(target.tenancyOCID[i].filter)) {
  //   //       return true;
  //   //     }
  //   //   }
  //   // }
    
  //   return false;
  // }  


 // **************************** Template variables helpers ****************************

  /**
   * Get all template variable descriptors
   */
  getVariableDescriptors(regex: string, includeCustom = true) {
    const vars = this.templateSrv.getVariables() || [];

    if (regex) {
      let regexVars = vars.filter((item) => _.isString(item.name) && item.name.match(regex) !== null);
      if (includeCustom) {
        const custom = vars.filter(
          (item) => item.type === "custom" || item.type === "constant"
        );
        regexVars = regexVars.concat(custom);
      }
      const uniqueRegexVarsMap = new Map();
      regexVars.forEach((varObj) =>
        uniqueRegexVarsMap.set(varObj.name, varObj)
      );
      return Array.from(uniqueRegexVarsMap.values());
    }
    return vars;
  }

  /**
   * List all variable names optionally filtered by regex or/and type
   * Returns list of names with '$' at the beginning. Example: ['$dimensionKey', '$dimensionValue']
   *
   * Updates:
   * Notes on implementation :
   * If a custom or constant is in  variables and  includeCustom, default is false.
   * Hence,the varDescriptors list is filtered for a unique set of var names
   */
  // getVariables2(regex: string, includeCustom: string) {
  //   const varDescriptors =
  //     this.getVariableDescriptors(regex, includeCustom) || [];
  //   return varDescriptors.map((item) => `$${item.name}`);
  // }

  /**
   * @param varName valid varName contains '$'. Example: '$dimensionKey'
   * Returns an array with variable values or empty array
   */
  getVariableValue(varName: string, scopedVars = {}) {
    return this.templateSrv.replace(varName, scopedVars) || varName;
  }

  /**
   * @param varName valid varName contains '$'. Example: '$dimensionKey'
   * Returns true if variable with the given name is found
   */
  isVariable(varName: string) {
    const varNames = this.getVariables() || [];
    console.log('variabili '+ varNames)
    return !!varNames.find((item) => item === varName);
  }

//  appendVariables(options: string, varQueryRegex:string) {
//     const vars = this.getVariables(varQueryRegex) || [];
//     vars.forEach(value => {
//       options.unshift({ value, text: value });
//     });
//     return options;
//   }

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

  async getSubscribedRegions(tenancyOCID: string): Promise<string[]> {
    if (this.isVariable(tenancyOCID)) {
      let { tenancyOCID: var_tenancy} = this.interpolateProps({tenancyOCID});
      console.log("region vartenancy "+var_tenancy)
      if (var_tenancy !== "") { 
        tenancyOCID = var_tenancy
      }      
    } else {
      console.log("region tenancyOCID "+tenancyOCID)
    }
  
    if (tenancyOCID === '') {
      return [];
    }
    const reqBody: JSON = {
      tenancy: tenancyOCID,
    } as unknown as JSON;
    return this.postResource(OCIResourceCall.Regions, reqBody).then((response) => {
      return new ResponseParser().parseRegions(response);
    });
  }

  async getCompartments(tenancyOCID: string): Promise<OCIResourceItem[]> {
    if (this.isVariable(tenancyOCID)) {
      let { tenancyOCID: var_tenancy} = this.interpolateProps({tenancyOCID});
      console.log("COMP vartenancy "+var_tenancy)
      if (var_tenancy !== "") { 
        tenancyOCID = var_tenancy
      }      
    } else {
      console.log("COMP tenancyOCID "+tenancyOCID)
    }    
    if (tenancyOCID === '') {
      return [];
    }
    const reqBody: JSON = {
      tenancy: tenancyOCID,
    } as unknown as JSON;
    return this.postResource(OCIResourceCall.Compartments, reqBody).then((response) => {
      return new ResponseParser().parseCompartments(response);
    });
  }

  async getNamespacesWithMetricNames(
    tenancyOCID: string,
    compartmentOCID: any,
    region: any
  ): Promise<OCINamespaceWithMetricNamesItem[]> {
    let { tenancyOCID: var_tenancy, region: var_region, compartmentOCID: var_compartment } = this.interpolateProps({ tenancyOCID, region, compartmentOCID });

    console.log("NS")
    console.log("NS "+tenancyOCID)
    console.log("NS "+compartmentOCID)
    console.log("NS "+region)
    console.log("NS "+var_region)
    console.log("NS "+var_tenancy)
    console.log("NS "+var_compartment)
    if (this.isVariable(tenancyOCID)) {
      let { tenancyOCID: var_tenancy} = this.interpolateProps({tenancyOCID});
      console.log("NS vartenancy "+var_tenancy)
      if (var_tenancy !== "") { 
        tenancyOCID = var_tenancy
      }      
    } else {
      console.log("NS tenancyOCID "+tenancyOCID)
    }

    if (this.isVariable(compartmentOCID)) {
      let { compartmentOCID: var_compartment} = this.interpolateProps({compartmentOCID});
      console.log("NS vartenancy "+var_compartment)
      if (var_compartment !== "") { 
        compartmentOCID = var_compartment
      }      
    } else {
      console.log("NS compartmentOCID "+compartmentOCID)
    }

    if (this.isVariable(region)) {
      let { region: var_region} = this.interpolateProps({region});
      console.log("NS vartenancy "+var_region)
      if (var_region !== "") { 
        region = var_region
      }      
    } else {
      console.log("NS region "+region)
    }

    if (tenancyOCID === '') {
      console.log("NS notenancy")
      return [];
    }
    if (region === undefined || region === QueryPlaceholder.Region) {
      console.log("NS noregion")
      return [];
    }

    if (compartmentOCID === undefined || compartmentOCID === QueryPlaceholder.Compartment) {
      console.log("NS compartmentOCID")
      compartmentOCID = '';
    }

    console.log("NS2 "+tenancyOCID)
    console.log("NS2 "+compartmentOCID)
    console.log("NS2 "+region)

    const reqBody: JSON = {
      tenancy: tenancyOCID,
      compartment: compartmentOCID,
      region: region,
    } as unknown as JSON;
    return this.postResource(OCIResourceCall.Namespaces, reqBody).then((response) => {
      return new ResponseParser().parseNamespacesWithMetricNames(response);
    });
  }


  async getResourceGroupsWithMetricNames(
    tenancyOCID: any,
    compartmentOCID: any,
    region: any,
    namespace: any
  ): Promise<OCIResourceGroupWithMetricNamesItem[]> {
    let { tenancyOCID: var_tenancy, region: var_region, compartmentOCID: var_compartment, namespace: var_namespace } = this.interpolateProps({ tenancyOCID, region, compartmentOCID, namespace });
    console.log("RG")
    console.log("RG "+tenancyOCID)
    console.log("RG "+compartmentOCID)
    console.log("RG "+region)
    console.log("RG "+var_region)
    console.log("RG "+var_tenancy)
    console.log("RG "+var_compartment)
    console.log("RG "+var_namespace)

    if (this.isVariable(tenancyOCID)) {
      let { tenancyOCID: var_tenancy} = this.interpolateProps({tenancyOCID});
      console.log("RG vartenancy "+var_tenancy)
      if (var_tenancy !== "") { 
        tenancyOCID = var_tenancy
      }      
    } else {
      console.log("RG tenancyOCID "+tenancyOCID)
    }

    if (this.isVariable(compartmentOCID)) {
      let { compartmentOCID: var_compartment} = this.interpolateProps({compartmentOCID});
      console.log("RG vartenancy "+var_compartment)
      if (var_compartment !== "") { 
        compartmentOCID = var_compartment
      }      
    } else {
      console.log("RG compartmentOCID "+compartmentOCID)
    }

    if (this.isVariable(region)) {
      let { region: var_region} = this.interpolateProps({region});
      console.log("RG vartenancy "+var_region)
      if (var_region !== "") { 
        region = var_region
      }      
    } else {
      console.log("RG region "+region)
    }

    if (this.isVariable(namespace)) {
      let { namespace: var_namespace} = this.interpolateProps({namespace});
      console.log("RG vartenancy "+var_namespace)
      if (var_namespace !== "") { 
        namespace = var_namespace
      }      
    } else {
      console.log("RG namespace "+namespace)
    }    

    if (tenancyOCID === '') {
      console.log("RG notenancy")
      return [];
    }
    if (region === undefined || region === QueryPlaceholder.Region) {
      console.log("RG noregion")
      return [];
    }

    if (compartmentOCID === undefined || compartmentOCID === QueryPlaceholder.Compartment) {
      console.log("RG compartmentOCID")
      compartmentOCID = '';
    }

    console.log("RG2 "+tenancyOCID)
    console.log("RG2 "+compartmentOCID)
    console.log("RG2 "+region)    
    console.log("RG2 "+namespace)    

    if (tenancyOCID === '') {
      return [];
    }
    if (region === undefined || namespace === undefined) {
      return [];
    }
    if (region === QueryPlaceholder.Region || namespace === QueryPlaceholder.Namespace) {
      return [];
    }

    if (compartmentOCID === undefined || compartmentOCID === QueryPlaceholder.Compartment) {
      compartmentOCID = '';
    }

    const reqBody: JSON = {
      tenancy: tenancyOCID,
      compartment: compartmentOCID,
      region: region,
      namespace: namespace,
    } as unknown as JSON;
    return this.postResource(OCIResourceCall.ResourceGroups, reqBody).then((response) => {
      return new ResponseParser().parseResourceGroupWithMetricNames(response);
    });
  }

  async getDimensions(
    tenancyOCID: any,
    compartmentOCID: any,
    region: any,
    namespace: any,
    metricName: any
  ): Promise<OCIResourceMetadataItem[]> {
    let { tenancyOCID: var_tenancy, region: var_region, compartmentOCID: var_compartment, namespace: var_namespace, metricName: var_metric } = this.interpolateProps({ tenancyOCID, region, compartmentOCID, namespace, metricName });
    console.log("DIM")
    console.log("DIM "+tenancyOCID)
    console.log("DIM "+compartmentOCID)
    console.log("DIM "+region)
    console.log("DIM "+var_region)
    console.log("DIM "+var_tenancy)
    console.log("DIM "+var_compartment)
    console.log("DIM "+var_namespace)
    console.log("DIM "+var_metric)


    if (this.isVariable(tenancyOCID)) {
      let { tenancyOCID: var_tenancy} = this.interpolateProps({tenancyOCID});
      console.log("DIM vartenancy "+var_tenancy)
      if (var_tenancy !== "") { 
        tenancyOCID = var_tenancy
      }      
    } else {
      console.log("DIM tenancyOCID "+tenancyOCID)
    }

    if (this.isVariable(compartmentOCID)) {
      let { compartmentOCID: var_compartment} = this.interpolateProps({compartmentOCID});
      console.log("DIM vartenancy "+var_compartment)
      if (var_compartment !== "") { 
        compartmentOCID = var_compartment
      }      
    } else {
      console.log("DIM compartmentOCID "+compartmentOCID)
    }

    if (this.isVariable(region)) {
      let { region: var_region} = this.interpolateProps({region});
      console.log("DIM vartenancy "+var_region)
      if (var_region !== "") { 
        region = var_region
      }      
    } else {
      console.log("DIM region "+region)
    }

    if (this.isVariable(namespace)) {
      let { namespace: var_namespace} = this.interpolateProps({namespace});
      console.log("DIM vartenancy "+var_namespace)
      if (var_namespace !== "") { 
        namespace = var_namespace
      }      
    } else {
      console.log("DIM namespace "+namespace)
    }

    if (this.isVariable(metricName)) {
      let { metricName: var_metric} = this.interpolateProps({metricName});
      console.log("DIM vartenancy "+var_metric)
      if (var_metric !== "") { 
        metricName = var_metric
      }      
    } else {
      console.log("DIM metric "+metricName)
    }       

    if (tenancyOCID === '') {
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

    if (compartmentOCID === undefined || compartmentOCID === QueryPlaceholder.Compartment) {
      compartmentOCID = '';
    }

    console.log("DIM2 "+tenancyOCID)
    console.log("DIM2 "+compartmentOCID)
    console.log("DIM2 "+region)    
    console.log("DIM2 "+namespace)   
    console.log("DIM2 "+metricName)   

    const reqBody: JSON = {
      tenancy: tenancyOCID,
      compartment: compartmentOCID,
      region: region,
      namespace: namespace,
      metric_name: metricName,
    } as unknown as JSON;
    return this.postResource(OCIResourceCall.Dimensions, reqBody).then((response) => {
      console.log("DO OK")
      return new ResponseParser().parseDimensions(response);
    });
  }
  async getTags(
    tenancyOCID: any,
    compartmentOCID: any,
    compartmentName: any,
    region: any,
    namespace: any
  ): Promise<OCIResourceMetadataItem[]> {
    if (tenancyOCID === '') {
      return [];
    }
    if (region === undefined || namespace === undefined) {
      return [];
    }
    if (region === QueryPlaceholder.Region || namespace === QueryPlaceholder.Namespace) {
      return [];
    }

    if (compartmentOCID === undefined || compartmentOCID === QueryPlaceholder.Compartment) {
      compartmentOCID = '';
    }
    if (compartmentName === undefined) {
      compartmentName = '';
    }

    const reqBody: JSON = {
      tenancy: tenancyOCID,
      compartment: compartmentOCID,
      compartment_name: compartmentName,
      region: region,
      namespace: namespace,
    } as unknown as JSON;
    return this.postResource(OCIResourceCall.Tags, reqBody).then((response) => {
      return new ResponseParser().parseTags(response);
    });
  }
}
