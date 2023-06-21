import { Observable } from 'rxjs';

import { DataSourceInstanceSettings, DataQueryRequest, DataQueryResponse, ScopedVars, MetricFindValue } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
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
  // namespacesQueryRegex,
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

export class OCIDataSource extends DataSourceWithBackend<OCIQuery, OCIDataSourceOptions> {
  private jsonData: any;
  // private backendSrv: BackendSrv;
  // private templateSrv: TemplateSrv;

  constructor(instanceSettings: DataSourceInstanceSettings<OCIDataSourceOptions>) {
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
  // applyTemplateVariables(query: OCIQuery, scopedVars: ScopedVars) {
  //   const templateSrv = getTemplateSrv();
  //   /**
  //    * Replace variables
  //    */
  //   return {
  //     ...query,
  //     tenancyOCID: query.tenancyOCID ? templateSrv.replace(query.tenancyOCID, scopedVars) : '',
  //     region: query.region ? templateSrv.replace(query.region, scopedVars) : '',
  //     compartment: query.compartmentOCID ? templateSrv.replace(query.compartmentOCID, scopedVars) : '',
  //     namespace: query.namespace ? templateSrv.replace(query.namespace, scopedVars) : '',
  //     resourceGroup: query.resourceGroup ? templateSrv.replace(query.resourceGroup, scopedVars) : '',
  //     metric: query.metric ? templateSrv.replace(query.metric, scopedVars) : '',
  //     // dimension: query.dimensionValues ? templateSrv.replace(query.dimensionValues, scopedVars) : '',
  //   };
  // }

  applyTemplateVariables(query: OCIQuery, scopedVars: ScopedVars) {
    // TODO: pass scopedVars to templateSrv.replace()
    const templateSrv = getTemplateSrv();
    query.tenancyOCID = templateSrv.replace(query.tenancyOCID, scopedVars);
    query.region = templateSrv.replace(query.region, scopedVars);
    query.compartmentOCID = templateSrv.replace(query.compartmentOCID, scopedVars);
    
    // query.maxRows = query.maxRows || '';
    // query.cacheDuration = query.cacheDuration || '';
    // if (typeof query.queryString === 'undefined' || query.queryString === '') {
    //   query.queryExecutionId = templateSrv.replace(query.queryExecutionId, scopedVars);
    //   query.inputs = query.queryExecutionId.split(/,/).map(id => {
    //     return {
    //       queryExecutionId: id,
    //     };
    //   });
    // } else {
    //   query.queryExecutionId = '';
    //   query.inputs = [];
    // }
    // query.queryString = templateSrv.replace(query.queryString, scopedVars) || '';
    // query.outputLocation = this.outputLocation;
    return query;
  }  


  // // **************************** Template variable helpers ****************************

  // /**
  //  * Matches the regex from creating template variables and returns options for the corresponding variable.
  //  * Example:
  //  * template variable with the query "regions()" will be matched with the regionsQueryRegex and list of available regions will be returned.
  //  */
  // templateMetricQuery(query: OCIQuery,varString: string) {




  //   let namespaceQuery = varString.match(namespacesQueryRegex);
  //   if (namespaceQuery) {
  //     if (this.tenancymode === "multitenancy") {
  //       let target = {
  //         tenancy: removeQuotes(this.getVariableValue(namespaceQuery[1])),
  //         region: removeQuotes(this.getVariableValue(namespaceQuery[2])),
  //         compartment: removeQuotes(this.getVariableValue(namespaceQuery[3])),
  //       };
  //       return this.getNamespaces(target).catch((err) => {
  //         throw new Error("Unable to get namespaces: " + err);
  //       });
  //     } else {
  //       let target = {
  //         tenancy: DEFAULT_TENANCY,
  //         region: removeQuotes(this.getVariableValue(namespaceQuery[1])),
  //         compartment: removeQuotes(this.getVariableValue(namespaceQuery[2])),
  //       };
  //       return this.getNamespaces(target).catch((err) => {
  //         throw new Error("Unable to get namespaces: " + err);
  //       });        
  //     }
  //   }

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
    // const templateSrv = getTemplateSrv();

    const tenancyQuery = query.match(tenanciesQueryRegex);
    if (tenancyQuery) {
      const tenancy = await this.getTenancies();
      return tenancy.map(n => {
        return { text: n.name, value: n.ocid };
      });   
    }    

    const regionQuery = query.match(regionsQueryRegex);
    if (regionQuery) {
      if (query.tenancymode === "multitenancy") {
        const regions = await this.getSubscribedRegions(query.tenancyOCID);
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
      if (query.tenancymode === "multitenancy") {
        const compartments = await this.getCompartments(query.tenancyOCID);
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


    // const workgroupNamesQuery = query.match(/^workgroup_names\(([^\)]+?)\)/);
    // if (workgroupNamesQuery) {
    //   const region = templateSrv.replace(workgroupNamesQuery[1]);
    //   const workgroupNames = await this.getWorkgroupNames(region);
    //   return workgroupNames.map(n => {
    //     return { text: n, value: n };
    //   });
    // }

    // const namedQueryNamesQuery = query.match(/^named_query_names\(([^\)]+?)(,\s?.+)?\)/);
    // if (namedQueryNamesQuery) {
    //   const region = templateSrv.replace(namedQueryNamesQuery[1]);
    //   let workGroup = namedQueryNamesQuery[2];
    //   if (workGroup) {
    //     workGroup = workGroup.substr(1); //remove the comma
    //     workGroup = workGroup.trim();
    //   } else {
    //     workGroup = '';
    //   }
    //   workGroup = templateSrv.replace(workGroup);
    //   const namedQueryNames = await this.getNamedQueryNames(region, workGroup);
    //   return namedQueryNames.map(n => {
    //     return { text: n, value: n };
    //   });
    // }

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
  
  query(options: DataQueryRequest<OCIQuery>): Observable<DataQueryResponse> {
    return super.query(options);
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
      console.log ("Ritorno di Tenanci");
      return new ResponseParser().parseTenancies(response);
    });
  }

  async getSubscribedRegions(tenancyOCID: string): Promise<string[]> {
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
    console.log("NS")
    console.log("NS "+tenancyOCID)
    console.log("NS "+compartmentOCID)
    console.log("NS "+region)


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
