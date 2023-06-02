import _ from 'lodash';

export interface OCIResourceItem {
  name: string;
  ocid: string;
}

export interface OCINamespaceWithMetricNamesItem {
  namespace: string;
  metric_names: string[];
}

export interface OCIResourceGroupWithMetricNamesItem {
  resource_group: string;
  metric_names: string[];
}

export interface OCIResourceMetadataItem {
  key: string;
  values: string[];
}

export class ResponseParser {
  parseTenancies(results: any): OCIResourceItem[] {
    const tenancies: OCIResourceItem[] = [];
    if (!results) {
      return tenancies;
    }

    let tList: OCIResourceItem[] = JSON.parse(JSON.stringify(results));
    return tList;
  }

  parseRegions(results: any): string[] {
    const regions: string[] = [];
    if (!results) {
      return regions;
    }

    let rList: string[] = JSON.parse(JSON.stringify(results));
    return rList;
  }

  parseTenancyMode(results: any): string {
    // const tenancymode: string = "";
    // if (!results) {
    //   return tenancymode;
    // }

    let rList: string = JSON.parse(JSON.stringify(results));
    return rList;
  }

  parseCompartments(results: any): OCIResourceItem[] {
    const compartments: OCIResourceItem[] = [];
    if (!results) {
      return compartments;
    }

    let cList: OCIResourceItem[] = JSON.parse(JSON.stringify(results));
    return cList;
  }

  parseNamespacesWithMetricNames(results: any): OCINamespaceWithMetricNamesItem[] {
    const namespaceWithMetricNames: OCINamespaceWithMetricNamesItem[] = [];
    if (!results) {
      return namespaceWithMetricNames;
    }

    let nmList: OCINamespaceWithMetricNamesItem[] = JSON.parse(JSON.stringify(results));
    return nmList;
  }

  parseResourceGroupWithMetricNames(results: any): OCIResourceGroupWithMetricNamesItem[] {
    const rgWithMetricNames: OCIResourceGroupWithMetricNamesItem[] = [];
    if (!results) {
      return rgWithMetricNames;
    }

    let rgList: OCIResourceGroupWithMetricNamesItem[] = JSON.parse(JSON.stringify(results));
    return rgList;
  }

  parseDimensions(results: any): OCIResourceMetadataItem[] {
    const dimensions: OCIResourceMetadataItem[] = [];
    if (!results) {
      return dimensions;
    }

    let dList: OCIResourceMetadataItem[] = JSON.parse(JSON.stringify(results));
    return dList;
  }

  parseTags(results: any): OCIResourceMetadataItem[] {
    const tags: OCIResourceMetadataItem[] = [];
    if (!results) {
      return tags;
    }

    let tList: OCIResourceMetadataItem[] = JSON.parse(JSON.stringify(results));
    return tList;
  }
}
