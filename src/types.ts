/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/

import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';


export interface MyDataSourceOptions extends DataSourceJsonData {}

export interface OCIConfig extends DataSourceJsonData {
  refId: string;
  name: string;
  region?: string;
  environments: string;
  tenancyMode: string;
}

export interface OCIConfigSec {
  userOcid?: string;
  tenancyOcid?: string;
  fingerprint?: string;
  apiKey?: string;
}


export enum BuilderMetricFieldAggregation {
  Sum = 'sum',
  Average = 'avg',
  Min = 'min',
  Max = 'max',
  Count = 'count',
  Any = 'any',
  // Count_Distinct = 'count_distinct',
}
export type BuilderMetricField = {
  field: string;
  aggregation: BuilderMetricFieldAggregation;
  alias?: string;
};

export interface FullEntity {
  name: string;
  label: string;
  custom: boolean;
  queryable: boolean;
}
interface FullFieldPickListItem {
  value: string;
  label: string;
}
export interface FullField {
  name: string;
  label: string;
  type: string;
  picklistValues: FullFieldPickListItem[];
  filterable?: boolean;
  sortable?: boolean;
  groupable?: boolean;
  aggregatable?: boolean;
}
export enum OrderByDirection {
  ASC = 'ASC',
  DESC = 'DESC',
}

export interface OrderBy {
  name: string;
  dir: OrderByDirection;
}

export enum FilterOperator {
  IsNull = 'IS NULL',
  IsNotNull = 'IS NOT NULL',
  Equals = '=',
  NotEquals = '!=',
  LessThan = '<',
  LessThanOrEqual = '<=',
  GreaterThan = '>',
  GreaterThanOrEqual = '>=',
  Like = 'LIKE',
  NotLike = 'NOT LIKE',
  In = 'IN',
  NotIn = 'NOT IN',
  WithInGrafanaTimeRange = 'WITH IN DASHBOARD TIME RANGE',
  OutsideGrafanaTimeRange = 'OUTSIDE DASHBOARD TIME RANGE',
}
export interface CommonFilterProps {
  filterType: 'custom';
  key: string;
  type: string;
  condition: 'AND' | 'OR';
}
export interface NullFilter extends CommonFilterProps {
  operator: FilterOperator.IsNull | FilterOperator.IsNotNull;
}
export interface BooleanFilter extends CommonFilterProps {
  type: 'boolean';
  operator: FilterOperator.Equals | FilterOperator.NotEquals;
  value: boolean;
}
export interface StringFilter extends CommonFilterProps {
  operator: FilterOperator.Equals | FilterOperator.NotEquals | FilterOperator.Like | FilterOperator.NotLike;
  value: string;
}

export interface NumberFilter extends CommonFilterProps {
  operator:
    | FilterOperator.Equals
    | FilterOperator.NotEquals
    | FilterOperator.LessThan
    | FilterOperator.LessThanOrEqual
    | FilterOperator.GreaterThan
    | FilterOperator.GreaterThanOrEqual;
  value: number;
}

export interface DateFilterWithValue extends CommonFilterProps {
  type: 'datetime' | 'date';
  operator:
    | FilterOperator.Equals
    | FilterOperator.NotEquals
    | FilterOperator.LessThan
    | FilterOperator.LessThanOrEqual
    | FilterOperator.GreaterThan
    | FilterOperator.GreaterThanOrEqual;
  value: string;
}
export interface DateFilterWithoutValue extends CommonFilterProps {
  type: 'datetime' | 'date';
  operator: FilterOperator.WithInGrafanaTimeRange | FilterOperator.OutsideGrafanaTimeRange;
}
export type DateFilter = DateFilterWithValue | DateFilterWithoutValue;

export interface MultiFilter extends CommonFilterProps {
  operator: FilterOperator.In | FilterOperator.NotIn;
  value: string[];
}

export type Filter = NullFilter | BooleanFilter | NumberFilter | DateFilter | StringFilter | MultiFilter;

//#endregion

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

export enum OCIResourceCall {
  Tenancies = 'tenancies',
  TenancyMode = 'tenancymode',
  Compartments = 'compartments',
  Regions = 'regions',
  Namespaces = 'namespaces',
  ResourceGroups = 'resourcegroups',
  Dimensions = 'dimensions',
  Tags = 'tags',
}

export enum QueryPlaceholder {
  Tenancy = 'select tenancy',
  Compartment = 'select compartment',
  Region = 'select region',
  Namespace = 'select namespace',
  Metric = 'select metric',
  Aggregation = 'select aggregation',
  Interval = 'select interval',
  Dimensions = 'select dimensions (optional)',
  ResourceGroup = 'select resource group',
  Tags = 'select resource tags (optional)',
  GroupBy = 'select option (optional)',
}

export interface DimensionPart {
  type: string;
  params: Array<string | number>;
  name?: string;
}

export type UnitOptions = 'minute' | 'hour';

export const IntervalOptions = [
  { label: '1 minute', value: '[1m]', description: 'Maximum time range supported: 7 days' },
  { label: '5 minutes', value: '[5m]', description: 'Maximum time range supported: 30 days' },
  { label: '1 hour', value: '[1h]', description: 'Maximum time range supported: 90 days' },
  { label: '1 day', value: '[1d]', description: 'Maximum time range supported: 90 days' },
];

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

export const GroupOptions = [
  { label: 'group by', value: 'groupBy()' },
  { label: 'grouping', value: 'grouping()' },
];

export interface OCIQuery extends DataQuery {
  queryText?: string;
  rawQuery: boolean;
  //hide: boolean;
  tenancyName: string;
  // tenancyOCID: string;
  tenancy: string;
  tenancymode: string;
  compartments?: any;
  compartmentName?: string;
  // compartmentOCID?: string;
  compartment?: string; // for legacy compatibility
  regions?: any;
  region?: string;
  namespace?: string;
  metricNames?: string[];
  metricNamesFromNS?: string[];
  metric?: string;
  interval: string;
  intervalLabel?: string;
  legendFormat?: string;
  statistic: string;
  statisticLabel?: string;
  resourcegroup?: string;
  dimensionValues?: string[];
  tagsValues?: string[];
  groupBy?: string;
}

export const defaultQuery: Partial<OCIQuery> = {};

/**
 * These are options configured for each DataSource instance
 */
export interface OCIDataSourceOptions extends DataSourceJsonData {
  tenancyName: string; // name of the base tenancy
  defaultRegion: string; // name of the base region
  environment?: string; // oci-cli, oci-instance
  multiTenancyMode?: string; // multi-profile, cross-tenancy-policy
  multiTenancyChoice?: string; // yes, no
  tenancymode?: string; // multi-profile, cross-tenancy-policy
  TenancyChoice?: string; // yes, no 
  multiTenancyFile?: string; // Default is ~/.oci/tenancies, if enabled
  configPath?: string; // Config file path. Default is ~/.oci/config
  configProfile?: string; // Config profile name, as specified in ~/.oci/config. Default is DEFAULT

  addon1: boolean;
  addon2: boolean;
  addon3: boolean;
  addon4: boolean;

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

	tenancy1: string;
	user1: string;
	fingerprint1: string;
	privkey1: string;

	tenancy2: string;
	user2: string;
	fingerprint2: string;
	privkey2: string;

	tenancy3: string;
	user3: string;
	fingerprint3: string;
	privkey3: string;

	tenancy4: string;
	user4: string;
	fingerprint4: string;
	privkey4: string;

	tenancy5: string;
	user5: string;
	fingerprint5: string;
	privkey5: string;
}
