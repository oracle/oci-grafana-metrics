import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

// export interface OCIQuery extends DataQuery {
//   multiplier: number;
//   queryType: string;
//   region: string;
//   compartment: string;
//   tenancy: string;
//   namespace: string
//   resourcegroup: string;  
// }

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


export enum Protocol {
  NATIVE = 'native',
  HTTP = 'http',
}

export enum Format {
  TIMESERIES = 0,
  TABLE = 1,
  LOGS = 2,
}

//#region Query
export enum QueryType {
  SQL = 'sql',
  Builder = 'builder',
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
  ConfigPath = '~/.oci/config',
  MultiTenanciesFile = '~/.oci/tenancies',
  ConfigProfile = 'DEFAULT',
}

export enum OCIResourceCall {
  Tenancies = 'tenancies',
  Compartments = 'compartments',
  Regions = 'regions',
  Namespaces = 'namespaces',
  ResourceGroups = 'resourcegroups',
  Dimensions = 'dimensions',
  Tags = 'tags',
}

export enum QueryPlaceholder {
  Tenancy = 'select tenancy',
  Compartment = 'select compartment (optional)',
  Region = 'select region',
  Namespace = 'select namespace',
  Metric = 'select metric',
  Aggregation = 'select aggregation',
  Interval = 'select interval',
  Dimensions = 'select dimensions (optional)',
  ResourceGroup = 'select resourcegroup (optional)',
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
  tenancyOCID: string;
  compartments?: any;
  compartmentName?: string;
  compartmentOCID?: string;
  regions?: any;
  region?: string;
  namespace?: string;
  merticNames?: string[];
  merticNamesFromNS?: string[];
  metric?: string;
  interval: string;
  intervalLabel?: string;
  statistic: string;
  statisticLabel?: string;
  resourceGroup?: string;
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
  Environment: string; // oci-cli, oci-instance
  multiTenancyMode?: string; // multi-profile, cross-tenancy-policy
  multiTenancyChoice?: string; // yes, no
  TenancyMode?: string; // multi-profile, cross-tenancy-policy
  TenancyChoice?: string; // yes, no 
  multiTenancyFile?: string; // Default is ~/.oci/tenancies, if enabled
  configPath?: string; // Config file path. Default is ~/.oci/config
  configProfile?: string; // Config profile name, as specified in ~/.oci/config. Default is DEFAULT
  enableCMDB?: boolean; // Choice to enable oracel cmdb datasource mapping
  enableCMDBUploadFile?: boolean; // Choice to enable user to upload customer mapping file
  cmdbFileContent: string; // CMDB file content as json string

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
