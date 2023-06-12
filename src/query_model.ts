import { OCIQuery, QueryPlaceholder, AggregationOptions, IntervalOptions } from './types';
import { ScopedVars } from '@grafana/data';
import { TemplateSrv } from '@grafana/runtime';

export default class QueryModel {
  target: OCIQuery;
  templateSrv: any;
  scopedVars: any;
  refId?: string;

  constructor(incomingQuery: OCIQuery, templateSrv?: TemplateSrv, scopedVars?: ScopedVars) {
    this.target = incomingQuery;
    this.templateSrv = templateSrv;
    this.scopedVars = scopedVars;

    this.target.tenancyOCID = incomingQuery.tenancyOCID || QueryPlaceholder.Tenancy;
    // this.target.compartmentOCID = incomingQuery.compartmentOCID || QueryPlaceholder.Compartment;
    this.target.compartmentOCID = incomingQuery.compartmentOCID || '';
    this.target.region = incomingQuery.region || QueryPlaceholder.Region;
    this.target.namespace = incomingQuery.namespace || QueryPlaceholder.Namespace;
    this.target.metric = incomingQuery.metric || QueryPlaceholder.Metric;
    this.target.statistic = incomingQuery.statistic || QueryPlaceholder.Aggregation;
    this.target.interval = incomingQuery.interval || QueryPlaceholder.Interval;
    this.target.resourceGroup = incomingQuery.resourceGroup || QueryPlaceholder.ResourceGroup;
    this.target.dimensionValues = incomingQuery.dimensionValues || [];
    this.target.tagsValues = incomingQuery.tagsValues || [];
    this.target.groupBy = incomingQuery.groupBy || QueryPlaceholder.GroupBy;

    this.target.hide = incomingQuery.hide ?? false;

    if (this.target.resourceGroup === QueryPlaceholder.ResourceGroup) {
      this.target.resourceGroup = '';
    }

    if (this.target.tenancyOCID === QueryPlaceholder.Tenancy) {
      this.target.tenancyOCID = 'DEFAULT/';
    }    

    // handle pre query gui panels gracefully, so by default we will have raw editor
    this.target.rawQuery = incomingQuery.rawQuery ?? true;

    if (this.target.rawQuery) {
      this.target.queryText =
        incomingQuery.queryText || 'metric[interval]{dimensionname="dimensionvalue"}.groupingfunction.statistic';
    } else {
      this.target.queryText = incomingQuery.queryText || this.buildQuery();
    }
  }

  isQueryReady() {
    // check if the query is ready to be built
    if (
      this.target.tenancyOCID === QueryPlaceholder.Tenancy ||
      this.target.region === QueryPlaceholder.Region ||
      this.target.namespace === QueryPlaceholder.Namespace ||
      this.target.metric === QueryPlaceholder.Metric
    ) {
      return false;
    }

    return true;
  }

  buildQuery() {
    let queryText = this.target.metric;

    // for default interval
    if (this.target.interval === QueryPlaceholder.Interval) {
      this.target.interval = IntervalOptions[0].value;
    }
    queryText += this.target.interval;

    // for dimensions
    let dimensionParams = '{';
    let noOfDimensions = this.target.dimensionValues?.length ?? 0;
    if (noOfDimensions !== 0) {
      this.target.dimensionValues?.forEach((dv) => {
        dimensionParams += dv;
        noOfDimensions--;

        if (noOfDimensions !== 0) {
          dimensionParams += ',';
        }
      });
      dimensionParams += '}';

      queryText += dimensionParams;
    }

    // for groupBy option
    if (this.target.groupBy !== QueryPlaceholder.GroupBy) {
      queryText += '.groupBy(' + this.target.groupBy + ')';
    }

    // for default statistics
    if (this.target.statistic === QueryPlaceholder.Aggregation) {
      this.target.statistic = AggregationOptions[0].value;
    }

    queryText += '.' + this.target.statistic;

    return queryText;
  }
}
