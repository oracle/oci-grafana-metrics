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
  // resourcegroupsQueryRegex,
  // metricsQueryRegex,
  regionsQueryRegex,
  tenanciesQueryRegex,
  DEFAULT_TENANCY,
  compartmentsQueryRegex,
  // dimensionValuesQueryRegex,
  // removeQuotes,
  // AUTO,
} from "./constants";
// import _ from 'lodash';

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
    return super.query(options);
  }

  // applyTemplateVariables(query: OCIQuery, scopedVars: ScopedVars) {
  //   // TODO: pass scopedVars to templateSrv.replace()
  //   const templateSrv = getTemplateSrv();

  //   console.log("queryregion1: "+query.region)
  //   console.log("compo1: "+query.compartmentOCID)
  //   console.log("name1: "+query.namespace)


  //   query.tenancyOCID = templateSrv.replace(query.tenancyOCID, scopedVars);
  //   query.region = templateSrv.replace(query.region, scopedVars);
  //   query.compartmentOCID = templateSrv.replace(query.compartmentOCID, scopedVars);

  //   // query.namespace = templateSrv.replace('$namespace', scopedVars);
  //   console.log("queryregion2: "+query.region)
  //   console.log("compo2: "+query.compartmentOCID)
  //   console.log("name2: "+query.namespace)



  //   return {
  //     ...query,
  //     datasource: this.getRef(),
  //     region: query.region,
  //     compartmentOCID: query.compartmentOCID,
  //     // timeSeriesList: timeSeriesList && {
  //     //   ...this.interpolateProps(timeSeriesList, scopedVars),
  //     //   projectName: this.templateSrv.replace(
  //     //     timeSeriesList.projectName ? timeSeriesList.projectName : this.getDefaultProject(),
  //     //     scopedVars
  //     //   ),
  //     //   filters: this.interpolateFilters(timeSeriesList.filters || [], scopedVars),
  //     //   groupBys: this.interpolateGroupBys(timeSeriesList.groupBys || [], scopedVars),
  //     //   view: timeSeriesList.view || 'FULL',
  //     // },
  //     // timeSeriesQuery: timeSeriesQuery && {
  //     //   ...this.interpolateProps(timeSeriesQuery, scopedVars),
  //     //   projectName: this.templateSrv.replace(
  //     //     timeSeriesQuery.projectName ? timeSeriesQuery.projectName : this.getDefaultProject(),
  //     //     scopedVars
  //     //   ),
  //     // },
  //     tenancyOCID: query.tenancyOCID,
  //   };
  //   // return query;
  // }  

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
    query.region = templateSrv.replace(query.region, scopedVars);
    query.tenancyOCID = templateSrv.replace(query.tenancyOCID, scopedVars);
    query.compartmentOCID = templateSrv.replace(query.compartmentOCID, scopedVars);
    query.namespace = templateSrv.replace(query.namespace, scopedVars);
    console.log("applyTemplateVariables: after region: " + query.region)
    console.log("applyTemplateVariables: after compartmentOCID: " + query.compartmentOCID)
    console.log("applyTemplateVariables: after tenancyOCID: " + query.tenancyOCID)
    console.log("applyTemplateVariables: after namespace: " + query.namespace)
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





  //   let resourcegroupQuery = varString.match(resourcegroupsQueryRegex);
  //   if (resourcegroupQuery) {
  //     if (this.tenancymode === "multitenancy") {
  //       let target = {
  //         tenancy: removeQuotes(this.getVariableValue(resourcegroupQuery[1])),
  //         region: removeQuotes(this.getVariableValue(resourcegroupQuery[2])),
  //         compartment: removeQuotes(this.getVariableValue(resourcegroupQuery[3])),
  //         namespace: removeQuotes(this.getVariableValue(resourcegroupQuery[4])),
  //       };
  //       return this.getResourceGroups(target).catch((err) => {
  //         throw new Error("Unable to get resourcegroups: " + err);
  //       });
  //     } else {
  //       let target = {
  //         tenancy: DEFAULT_TENANCY,
  //         region: removeQuotes(this.getVariableValue(resourcegroupQuery[1])),
  //         compartment: removeQuotes(this.getVariableValue(resourcegroupQuery[2])),
  //         namespace: removeQuotes(this.getVariableValue(resourcegroupQuery[3])),
  //       };
  //       return this.getResourceGroups(target).catch((err) => {
  //         throw new Error("Unable to get resourcegroups: " + err);
  //       });        
  //     }
  //   }

  //   let metricQuery = varString.match(metricsQueryRegex);
  //   if (metricQuery) {
  //     if (this.tenancymode === "multitenancy") {
  //       let target = {
  //         tenancy: removeQuotes(this.getVariableValue(metricQuery[1])),
  //         region: removeQuotes(this.getVariableValue(metricQuery[2])),
  //         compartment: removeQuotes(this.getVariableValue(metricQuery[3])),
  //         namespace: removeQuotes(this.getVariableValue(metricQuery[4])),
  //         resourcegroup: removeQuotes(this.getVariableValue(metricQuery[5])),
  //       };
  //       return this.metricFindQuery(target).catch((err) => {
  //         throw new Error("Unable to get metrics: " + err);
  //       });
  //     } else {
  //       let target = {
  //         tenancy: DEFAULT_TENANCY,
  //         region: removeQuotes(this.getVariableValue(metricQuery[1])),
  //         compartment: removeQuotes(this.getVariableValue(metricQuery[2])),
  //         namespace: removeQuotes(this.getVariableValue(metricQuery[3])),
  //         resourcegroup: removeQuotes(this.getVariableValue(metricQuery[4])),
  //       };
  //       return this.metricFindQuery(target).catch((err) => {
  //         throw new Error("Unable to get metrics: " + err);
  //       });        
  //     }  
  //   }

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
        const region = templateSrv.replace(namespaceQuery[2]);
        const compartment = templateSrv.replace(namespaceQuery[3]);
        const namespaces = await this.getNamespacesWithMetricNames(tenancy, compartment, region);
        return namespaces.map(n => {
          return { text: n.namespace, value: n.namespace };
        });      
      }
    }


    // const namedQueryQueryQuery = query.match(/^named_query_queries\(([^,]+?),\s?([^,]+)(,\s?.+)?\)/);
    // if (namedQueryQueryQuery) {
    //   const region = templateSrv.replace(namedQueryQueryQuery[1]);
    //   const pattern = templateSrv.replace(namedQueryQueryQuery[2], {}, 'regex');
    //   let workGroup = namedQueryQueryQuery[3];
    //   if (workGroup) {
    //     workGroup = workGroup.substr(1); //remove the comma
    //     workGroup = workGroup.trim();
    //   } else {
    //     workGroup = '';
    //   }
    //   workGroup = templateSrv.replace(workGroup);
    //   const namedQueryQueries = await this.getNamedQueryQueries(region, pattern, workGroup);
    //   return namedQueryQueries.map(n => {
    //     return { text: n, value: n };
    //   });
    // }

    // const queryExecutionIdsQuery = query.match(/^query_execution_ids\(([^,]+?),\s?([^,]+?),\s?([^,]+)(,\s?.+)?\)/);
    // if (queryExecutionIdsQuery) {
    //   const region = templateSrv.replace(queryExecutionIdsQuery[1]);
    //   const limit = parseInt(templateSrv.replace(queryExecutionIdsQuery[2]), 10);
    //   const pattern = templateSrv.replace(queryExecutionIdsQuery[3], {}, 'regex');
    //   let workGroup = queryExecutionIdsQuery[4];
    //   if (workGroup) {
    //     workGroup = workGroup.substr(1); //remove the comma
    //     workGroup = workGroup.trim();
    //   } else {
    //     workGroup = '';
    //   }
    //   workGroup = templateSrv.replace(workGroup);
    //   const to = new Date(parseInt(templateSrv.replace('$__to'), 10)).toISOString();

    //   const queryExecutions = await this.getQueryExecutions(region, limit, pattern, workGroup, to);
    //   return queryExecutions.map(n => {
    //     const id = n.QueryExecutionId;
    //     return { text: id, value: id };
    //   });
    // }

    // const queryExecutionIdsByNameQuery = query.match(
    //   /^query_execution_ids_by_name\(([^,]+?),\s?([^,]+?),\s?([^,]+)(,\s?.+)?\)/
    // );
    // if (queryExecutionIdsByNameQuery) {
    //   const region = templateSrv.replace(queryExecutionIdsByNameQuery[1]);
    //   const limit = parseInt(templateSrv.replace(queryExecutionIdsByNameQuery[2]), 10);
    //   const pattern = templateSrv.replace(queryExecutionIdsByNameQuery[3], {}, 'regex');
    //   let workGroup = queryExecutionIdsByNameQuery[4];
    //   if (workGroup) {
    //     workGroup = workGroup.substr(1); //remove the comma
    //     workGroup = workGroup.trim();
    //   } else {
    //     workGroup = '';
    //   }
    //   workGroup = templateSrv.replace(workGroup);
    //   const to = new Date(parseInt(templateSrv.replace('$__to'), 10)).toISOString();

    //   const queryExecutionsByName = await this.getQueryExecutionsByName(region, limit, pattern, workGroup, to);
    //   return queryExecutionsByName.map(n => {
    //     const id = n.QueryExecutionId;
    //     return { text: id, value: id };
    //   });
    // }

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

    if (var_tenancy !== "") { 
      tenancyOCID = var_tenancy
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
