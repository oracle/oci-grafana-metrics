/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/

import { OCIQuery, QueryPlaceholder, AggregationOptions } from './types';
// import {SetAutoInterval} from './datasource'
import { ScopedVars } from '@grafana/data';
import { TemplateSrv, getTemplateSrv } from '@grafana/runtime';

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
    this.target.hide = incomingQuery.hide ?? false;

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
      console.log("buildQuery nel Query_model")
      console.log("incomingQuery.queryText nel Query_model" +incomingQuery.queryText)
      console.log("this.target.metric nel Query_model" + this.target.metric)
      this.target.queryText = incomingQuery.queryText || this.buildQuery(String(this.target.metric));
      console.log("this.target.queryText nel Query_model" +this.target.queryText)

    }
  }

  SetAutoInterval(timestamp1: number, timestamp2: number): string {
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
  }

  isQueryReady() {
    // check if the query is ready to be built
    console.log("this.target.metric "+this.target.metric)
    if (
      this.target.tenancy === QueryPlaceholder.Tenancy ||
      this.target.region === QueryPlaceholder.Region ||
      this.target.namespace === QueryPlaceholder.Namespace ||
      ((this.target.metric === QueryPlaceholder.Metric || this.target.metric === undefined) && this.target.queryTextRaw === '')
    ) {
      return false;
    }

    return true;
  }


  buildQuery(queryText: string) { 
    //check if a raw query is being used
    if (this.target.queryTextRaw !== '' && this.target.rawQuery === false) {
      queryText = String(this.target.queryTextRaw);
    }  else {
      // if builder mode is used then:
      // add interval
      console.log ("this.target.interval "+this.target.interval)
      if (this.target.interval === QueryPlaceholder.Interval || this.target.interval === "auto" || this.target.interval === undefined){
        const TimeStart = parseInt(getTemplateSrv().replace("${__from}"), 10)
        const TimeEnd  = parseInt(getTemplateSrv().replace("${__to}"), 10)
        console.log ("TimeStart "+TimeStart)
        console.log ("TimeEnd "+TimeEnd)
        if (isNaN(TimeStart) || isNaN(TimeEnd)){
          this.target.interval = "[1m]"
        } else {
          this.target.interval = this.SetAutoInterval(TimeStart, TimeEnd);

        }
      }
      queryText += this.target.interval;

      console.log ("queryText "+queryText)

      // add dimensions
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

      // add groupBy option
      if (this.target.groupBy !== QueryPlaceholder.GroupBy) {
        queryText += '.groupBy(' + this.target.groupBy + ')';
      }

      // add default statistics
      if (this.target.statistic === QueryPlaceholder.Aggregation) {
        this.target.statistic = AggregationOptions[0].value;
      }
      queryText += '.' + this.target.statistic;
      } 

    return queryText;
  }
}
