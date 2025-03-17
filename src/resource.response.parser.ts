/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/

import _ from 'lodash';

/**
 * @interface OCIResourceItem
 * @description Represents a generic OCI resource with a name and OCID.
 * @property {string} name - The display name of the OCI resource.
 * @property {string} ocid - The Oracle Cloud Identifier of the OCI resource.
 */
export interface OCIResourceItem {
  name: string;
  ocid: string;
}

/**
 * @interface OCINamespaceWithMetricNamesItem
 * @description Represents a namespace and its associated metric names.
 * @property {string} namespace - The OCI namespace.
 * @property {string[]} metric_names - An array of metric names within the namespace.
 */
export interface OCINamespaceWithMetricNamesItem {
  namespace: string;
  metric_names: string[];
}

/**
 * @interface OCIResourceGroupWithMetricNamesItem
 * @description Represents a resource group and its associated metric names.
 * @property {string} resource_group - The name of the OCI resource group.
 * @property {string[]} metric_names - An array of metric names within the resource group.
 */
export interface OCIResourceGroupWithMetricNamesItem {
  resource_group: string;
  metric_names: string[];
}

/**
 * @interface OCIResourceMetadataItem
 * @description Represents a metadata item with a key and an array of values.
 * @property {string} key - The metadata key.
 * @property {string[]} values - An array of values associated with the key.
 */
export interface OCIResourceMetadataItem {
  key: string;
  values: string[];
}

/**
 * @class ResponseParser
 * @description Provides methods for parsing responses from OCI API calls.
 */
export class ResponseParser {
  /**
   * @function parseTenancies
   * @description Parses the response from the OCI API call to list tenancies.
   * @param {any} results - The raw response from the OCI API.
   * @returns {OCIResourceItem[]} An array of OCIResourceItem representing the tenancies.
   */
  parseTenancies(results: any): OCIResourceItem[] {
    const tenancies: OCIResourceItem[] = [];
    if (!results) {
      return tenancies;
    }

    let tList: OCIResourceItem[] = JSON.parse(JSON.stringify(results));
    return tList;
  }

  /**
   * @function parseRegions
   * @description Parses the response from the OCI API call to list subscribed regions.
   * @param {any} results - The raw response from the OCI API.
   * @returns {string[]} An array of strings representing the subscribed regions.
   */
  parseRegions(results: any): string[] {
    const regions: string[] = [];
    if (!results) {
      return regions;
    }

    let rList: string[] = JSON.parse(JSON.stringify(results));
    return rList;
  }

  /**
   * @function parseTenancyMode
   * @description Parses the response from the OCI API call to get tenancy mode.
   * @param {any} results - The raw response from the OCI API.
   * @returns {string[]} An array of strings representing the tenancy modes.
   */
  parseTenancyMode(results: any): string[] {
    const tenancymodes: string[] = [];
    if (!results) {
      return tenancymodes;
    }

    let rList: string[] = JSON.parse(JSON.stringify(results));
    return rList;
  }

  /**
   * @function parseCompartments
   * @description Parses the response from the OCI API call to list compartments.
   * @param {any} results - The raw response from the OCI API.
   * @returns {OCIResourceItem[]} An array of OCIResourceItem representing the compartments.
   */
  parseCompartments(results: any): OCIResourceItem[] {
    const compartments: OCIResourceItem[] = [];
    if (!results) {
      return compartments;
    }

    let cList: OCIResourceItem[] = JSON.parse(JSON.stringify(results));
    return cList;
  }

  /**
   * @function parseNamespacesWithMetricNames
   * @description Parses the response from the OCI API call to list namespaces with their associated metric names.
   * @param {any} results - The raw response from the OCI API.
   * @returns {OCINamespaceWithMetricNamesItem[]} An array of OCINamespaceWithMetricNamesItem representing the namespaces and their metric names.
   */
  parseNamespacesWithMetricNames(results: any): OCINamespaceWithMetricNamesItem[] {
    const namespaceWithMetricNames: OCINamespaceWithMetricNamesItem[] = [];
    if (!results) {
      return namespaceWithMetricNames;
    }

    let nmList: OCINamespaceWithMetricNamesItem[] = JSON.parse(JSON.stringify(results));
    return nmList;
  }

  /**
   * @function parseResourceGroupWithMetricNames
   * @description Parses the response from the OCI API call to list resource groups with their associated metric names.
   * @param {any} results - The raw response from the OCI API.
   * @returns {OCIResourceGroupWithMetricNamesItem[]} An array of OCIResourceGroupWithMetricNamesItem representing the resource groups and their metric names.
   */
  parseResourceGroupWithMetricNames(results: any): OCIResourceGroupWithMetricNamesItem[] {
    const rgWithMetricNames: OCIResourceGroupWithMetricNamesItem[] = [];
    if (!results) {
      return rgWithMetricNames;
    }

    let rgList: OCIResourceGroupWithMetricNamesItem[] = JSON.parse(JSON.stringify(results));
    return rgList;
  }

  /**
   * @function parseDimensions
   * @description Parses the response from the OCI API call to list dimensions.
   * @param {any} results - The raw response from the OCI API.
   * @returns {OCIResourceMetadataItem[]} An array of OCIResourceMetadataItem representing the dimensions.
   */
  parseDimensions(results: any): OCIResourceMetadataItem[] {
    const dimensions: OCIResourceMetadataItem[] = [];
    if (!results) {
      return dimensions;
    }

    let dList: OCIResourceMetadataItem[] = JSON.parse(JSON.stringify(results));
    return dList;
  }

  /**
   * @function parseTags
   * @description Parses the response from the OCI API call to list tags.
   * @param {any} results - The raw response from the OCI API.
   * @returns {OCIResourceMetadataItem[]} An array of OCIResourceMetadataItem representing the tags.
   */
  parseTags(results: any): OCIResourceMetadataItem[] {
    const tags: OCIResourceMetadataItem[] = [];
    if (!results) {
      return tags;
    }

    let tList: OCIResourceMetadataItem[] = JSON.parse(JSON.stringify(results));
    return tList;
  }
}
