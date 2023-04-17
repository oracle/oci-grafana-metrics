import { DataSourceJsonData } from '@grafana/data';

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
