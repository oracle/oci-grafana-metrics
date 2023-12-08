/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/

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

    this.target.tenancy = incomingQuery.tenancy || QueryPlaceholder.Tenancy;
    this.target.compartment = incomingQuery.compartment || '';
    this.target.region = incomingQuery.region || QueryPlaceholder.Region;
    this.target.namespace = incomingQuery.namespace || QueryPlaceholder.Namespace;
    this.target.metric = incomingQuery.metric || QueryPlaceholder.Metric;
    this.target.statistic = incomingQuery.statistic || QueryPlaceholder.Aggregation;
    this.target.interval = incomingQuery.interval || QueryPlaceholder.Interval;
    this.target.resourcegroup = incomingQuery.resourcegroup || QueryPlaceholder.ResourceGroup;
    this.target.dimensionValues = incomingQuery.dimensionValues || [];
    this.target.tagsValues = incomingQuery.tagsValues || [];
    this.target.groupBy = incomingQuery.groupBy || QueryPlaceholder.GroupBy;
    this.target.queryTextRaw = incomingQuery.queryTextRaw || '';
    this.target.hide = incomingQuery.hide ?? true;

    if (this.target.resourcegroup === QueryPlaceholder.ResourceGroup) {
      this.target.resourcegroup = '';
    }

    if (this.target.tenancy === QueryPlaceholder.Tenancy) {
        this.target.tenancy = 'DEFAULT/';
    }   

    // handle pre query gui panels gracefully, so by default we will NOT have raw editor. Here we are using this logic: Query builder: true, Raw Editor: false
    this.target.rawQuery = incomingQuery.rawQuery ?? true;

    if (this.target.rawQuery === false) {
      this.target.queryText =
        incomingQuery.queryTextRaw || 'metric[interval]{dimensionname="dimensionvalue"}.groupingfunction.statistic';
    } else {
      this.target.queryText = incomingQuery.queryText || this.buildQuery(String(this.target.metric));
    }
  }

  isQueryReady() {
    // check if the query is ready to be built
    if (
      this.target.tenancy === QueryPlaceholder.Tenancy ||
      this.target.region === QueryPlaceholder.Region ||
      this.target.namespace === QueryPlaceholder.Namespace ||
      (this.target.metric === QueryPlaceholder.Metric && this.target.queryTextRaw === '')
    ) {
      return false;
    }

    return true;
  }
  

  buildQuery(queryText: string) { 

    if (this.target.queryTextRaw !== '' && this.target.rawQuery === false) {
      queryText = String(this.target.queryTextRaw);
    }  else {
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
      } 


    return queryText;
  }
}
