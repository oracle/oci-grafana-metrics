/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/

import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';


export enum DefaultOCIOptions {
  ConfigProfile = 'DEFAULT',
}

export const DEFAULT_TENANCY = "DEFAULT/";
export const compartmentsQueryRegex = /^compartments\(\s*(\".+\"|\'.+\'|\$\w+)\s*\)|^compartments\(\)\s*/;
export const regionsQueryRegex = /^regions\(\s*(\".+\"|\'.+\'|\$\w+)\s*\)|^regions\(\)\s*/;
export const tenanciesQueryRegex = /^tenancies\(\)\s*/;
export const namespacesQueryRegex = /^namespaces\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const resourcegroupsQueryRegex = /^resourcegroups\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const metricsQueryRegex = /^metrics\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const dimensionQueryRegex = /^dimensions\(\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*,\s*(\".+\"|\'.+\'|\$\w+)\s*(?:,\s*(\".+\"|\'.+\'|\$\w+)\s*)?\)/;
export const windowsAndResolutionRegex = /^[0-9]+[mhs]$/;

/**
 * Enum representing the different OCI resource API calls.
 */
export enum OCIResourceCall {
  /**
   * Represents the API call to list tenancies.
   */
  Tenancies = 'tenancies',
  /**
   * Represents the API call to get tenancy mode information.
   */
  TenancyMode = 'tenancymode',
  /**
   * Represents the API call to list compartments.
   */
  Compartments = 'compartments',
  /**
   * Represents the API call to list regions.
   */
  Regions = 'regions',
  /**
   * Represents the API call to list namespaces.
   */
  Namespaces = 'namespaces',
  /**
   * Represents the API call to list resource groups.
   */
  ResourceGroups = 'resourcegroups',
  /**
   * Represents the API call to list dimensions.
   */
  Dimensions = 'dimensions',
  /**
   * Represents the API call to list tags.
   */
  Tags = 'tags',
}

/**
 * Enum representing the different query placeholders used in the UI.
 */
export enum QueryPlaceholder {
  /**
   * Placeholder for the tenancy selection.
   */
  Tenancy = 'select tenancy',
  /**
   * Placeholder for the compartment selection.
   */
  Compartment = 'select compartment',
  /**
   * Placeholder for the region selection.
   */
  Region = 'select region',
  /**
   * Placeholder for the namespace selection.
   */
  Namespace = 'select namespace',
  /**
   * Placeholder for the metric selection.
   */
  Metric = 'select metric',
  /**
   * Placeholder for the aggregation selection.
   */
  Aggregation = 'select aggregation',
  /**
   * Placeholder for the interval selection.
   */
  Interval = 'select interval',
  /**
   * Placeholder for the dimensions selection.
   */
  Dimensions = 'select dimensions (optional)',
  /**
   * Placeholder for the resource group selection.
   */
  ResourceGroup = 'select resource group',
  /**
   * Placeholder for the tags selection.
   */
  Tags = 'select resource tags (optional)',
  /**
   * Placeholder for the group by selection.
   */
  GroupBy = 'select option (optional)',
}

export interface DimensionPart {
  type: string;
  params: Array<string | number>;
  name?: string;
}

export type UnitOptions = 'minute' | 'hour';

/**
 * Represents the available interval options for metric queries.
 */
export const IntervalOptions = [
  { label: '1 minute', value: '[1m]', description: 'Maximum time range supported: 7 days' },
  { label: '5 minutes', value: '[5m]', description: 'Maximum time range supported: 30 days' },
  { label: '1 hour', value: '[1h]', description: 'Maximum time range supported: 90 days' },
  { label: '1 day', value: '[1d]', description: 'Maximum time range supported: 90 days' },
  { label: 'Auto', value: 'auto', description: 'Automatic selection of interval accordingly to OCI default' },
];

/**
 * Represents the available aggregation options for metric queries.
 */
export const AggregationOptions = [
  { label: 'average', value: 'avg()' },
  { label: 'count', value: 'count()' },
  { label: 'per-interval change', value: 'increment()' },
  { label: 'maximum', value: 'max()' },
  { label: 'minimum', value: 'min()' },
  { label: 'per-interval average rate of change', value: 'rate()' },
  { label: 'sum', value: 'sum()' },
  { label: 'P90', value: 'percentile(.90)' },
  { label: 'P95', value: 'percentile(.95)' },
  { label: 'P99', value: 'percentile(.99)' },
  { label: 'P99.9', value: 'percentile(.999)' },
];

/**
 * Represents the available group by options for metric queries.
 */
export const GroupOptions = [
  { label: 'group by', value: 'groupBy()' },
  { label: 'grouping', value: 'grouping()' },
];

/**
 * Represents the structure of an OCI query.
 */
export interface OCIQuery extends DataQuery {
  /**
   * The MQL query text.
   */
  queryText?: string;
  /**
   * The raw MQL query text entered by the user, before any processing.
   */
  queryTextRaw?: string;
  /**
   * Indicates whether the query is a raw query or a builder-based query.
   */
  rawQuery: boolean;
  /**
   * The name of the tenancy.
   */
  tenancyName: string;
  /**
   * The OCID of the tenancy.
   */
  tenancy: string;
  /**
   * The tenancy mode (e.g., single-tenancy, multi-tenancy).
   */
  tenancymode: string;
  /**
   * The name of the compartment.
   */
  compartmentName?: string;
  /**
   * The OCID of the compartment.
   */
  compartment?: string;
  /**
   * The region to query.
   */
  region?: string;
  /**
   * The namespace to query.
   */
  namespace?: string;
  /**
   * A list of metric names.
   */
  metricNames?: string[];
  /**
   * The metric to query.
   */
  metric?: string;
  /**
   * The query interval.
   */
  interval: string;
  /**
   * The label of the interval.
   */
  intervalLabel?: string;
  /**
   * The legend format for the query.
   */
  legendFormat?: string;
  /**
   * The aggregation statistic to use.
   */
  statistic: string;
  /**
   * The label of the statistic.
   */
  statisticLabel?: string;
  /**
   * The resource group to query.
   */
  resourcegroup?: string;
  /**
   * A list of dimension values.
   */
  dimensionValues?: string[];
  /**
   * A list of tag values.
   */
  tagsValues?: string[];
  /**
   * The group by option.
   */
  groupBy?: string;
}

export const defaultQuery: Partial<OCIQuery> = {};

/**
 * These are options configured for each DataSource instance
 */
export interface OCIDataSourceOptions extends DataSourceJsonData {
  tenancyName: string; // name of the base tenancy
  environment?: string; // oci-cli, oci-instance
  tenancymode?: string; // multi-profile, cross-tenancy-policy
  xtenancy0: string;


  addon1: boolean;
  addon2: boolean;
  addon3: boolean;
  addon4: boolean;

  customregionbool0: boolean;
  customregionbool1: boolean;
  customregionbool2: boolean;
  customregionbool3: boolean;
  customregionbool4: boolean;
  customregionbool5: boolean;

  customregion0: string
  customregion1: string  
  customregion2: string  
  customregion3: string  
  customregion4: string  
  customregion5: string  


	profile0: string;
	region0: string;

	profile1: string;
	region1: string;

	profile2: string;
	region2: string;

	profile3: string;
	region3: string;

	profile4: string;
	region4: string;

	profile5: string;
	region5: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface OCISecureJsonData {
	tenancy0: string;
	user0: string;
	privkey0: string;
	fingerprint0: string;
  customdomain0: string

	tenancy1: string;
	user1: string;
	fingerprint1: string;
	privkey1: string;
  customdomain1: string

	tenancy2: string;
	user2: string;
	fingerprint2: string;
	privkey2: string;
  customdomain2: string

	tenancy3: string;
	user3: string;
	fingerprint3: string;
	privkey3: string;
  customdomain3: string

	tenancy4: string;
	user4: string;
	fingerprint4: string;
	privkey4: string;
  customdomain4: string

	tenancy5: string;
	user5: string;
	fingerprint5: string;
	privkey5: string;
  customdomain5: string
}

/**
 * Function to automatically determine the appropriate interval based on the time range.
 *
 * @param timestamp1 - The start timestamp in milliseconds.
 * @param timestamp2 - The end timestamp in milliseconds.
 * @returns The suggested interval string (e.g., "[1m]", "[5m]", "[1h]").
 */
export const SetAutoInterval = (timestamp1: number, timestamp2: number): string => {
  const differenceInMs = timestamp2 - timestamp1;
  const differenceInHours = differenceInMs / (1000 * 60 * 60);

  // use limits and defaults specified here: https://docs.oracle.com/en-us/iaas/Content/Monitoring/Reference/mql.htm#Interval
  if (differenceInHours <= 6) {
    return "[1m]"; // Equal or Less than 6 hours, set to 1 minute interval
  } else if (differenceInHours < 36) {
    return "[5m]"; // Between 6 and 36 hours, set to 5 minute interval
  } else {
    return "[1h]"; // More than 36 hours, set to 1 hour interval
  }
};
