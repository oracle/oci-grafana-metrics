/*
** Copyright © 2022 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
import React, { PureComponent } from 'react';
import { CoreApp, HistoryItem, PanelData, QueryEditorProps, TimeRange } from '@grafana/data';
import { DataQuery } from '@grafana/schema';
import {
  windows,
  namespacesQueryRegex,
  resourcegroupsQueryRegex,
  metricsQueryRegex,
  regionsQueryRegex,
  tenanciesQueryRegex,
  compartmentsQueryRegex,
  dimensionKeysQueryRegex,
  dimensionValuesQueryRegex,
  windowsAndResolutionRegex, resolutions, AUTO
} from './constants'
import { OCIConfig, OCIConfigSec } from './types';


export const SELECT_PLACEHOLDERS = {
  DIMENSION_KEY: 'select dimension',
  DIMENSION_VALUE: 'select value',
  COMPARTMENT: 'select compartment',
  REGION: 'select region',
  TENANCY: 'select tenancy',
  NAMESPACE: 'select namespace',
  RESOURCEGROUP: 'select resource group',
  METRIC: 'select metric',
  WINDOW: 'select window'
}


interface ITarget {
  region: string;
  tenancy: string;
  MultiTenancy?: boolean;
}

export class OCIDatasourceQueryCtrl implements QueryEditorProps<any, any> {
  private q: any;
  private uiSegmentSrv: IUiSegmentSrv;
  private dimensionsCache: Record<string, any>;
  private dimensionSegments: Record<string, any>;
  target: any;
  removeDimensionSegment: any;
  getSelectDimensionKeySegment: () => any;
  getDimensionOperatorSegment: () => any;
  getSelectDimensionValueSegment: () => any;
  panelCtrl: any;
  static templateUrl: string;


  constructor($scope: any, $injector: any, private $q: ng.IQService, uiSegmentSrv: IUiSegmentSrv) {
    // super($scope, $injector)

    this.q = $q;
    this.uiSegmentSrv = uiSegmentSrv;

    this.target.region = this.target.region || SELECT_PLACEHOLDERS.REGION;
    this.target.tenancy = this.target.tenancy || SELECT_PLACEHOLDERS.TENANCY;
    this.target.compartment = this.target.compartment || SELECT_PLACEHOLDERS.COMPARTMENT;
    this.target.namespace = this.target.namespace || SELECT_PLACEHOLDERS.NAMESPACE;
    this.target.resourcegroup = this.target.resourcegroup || SELECT_PLACEHOLDERS.RESOURCEGROUP;
    this.target.metric = this.target.metric || SELECT_PLACEHOLDERS.METRIC;
    this.target.resolution = this.target.resolution || AUTO;
    this.target.window = this.target.window || AUTO;
    this.target.aggregation = this.target.aggregation || 'mean()'
    this.target.dimensions = this.target.dimensions || [];
    this.target.legendFormat = this.target.legendFormat || ''
    this.target.tenancymode = this.datasource.tenancymode || ''

    this.dimensionSegments = [];
    this.removeDimensionSegment = uiSegmentSrv.newSegment({ fake: true, value: '-- remove dimension --' });
    this.getSelectDimensionKeySegment = () => uiSegmentSrv.newSegment({ value: SELECT_PLACEHOLDERS.DIMENSION_KEY, type: 'key' });
    this.getDimensionOperatorSegment = () => this.uiSegmentSrv.newOperator('=');
    this.getSelectDimensionValueSegment = () => uiSegmentSrv.newSegment({ value: SELECT_PLACEHOLDERS.DIMENSION_VALUE, type: 'value' });

    this.dimensionsCache = {};
    if (this.datasource.tenancymode === "multitenancy") {
      this.target.MultiTenancy = true;
    }

    // rebuild dimensionSegments on query editor load
    for (let i = 0; i < this.target.dimensions.length; i++) {
      const dim = this.target.dimensions[i];
      if (i > 0) {
        this.dimensionSegments.push(this.uiSegmentSrv.newCondition(','))
      }
      this.dimensionSegments.push(this.uiSegmentSrv.newSegment({ value: dim.key, type: 'key' }));
      this.dimensionSegments.push(this.uiSegmentSrv.newSegment({ value: dim.operator, type: 'operator' }));
      this.dimensionSegments.push(this.uiSegmentSrv.newSegment({ value: dim.value, type: 'value' }));
    }
    this.dimensionSegments.push(this.uiSegmentSrv.newPlusButton())
  }
  datasource: any;
  query: any;
  onRunQuery!: () => void;
  onChange!: (value: any) => void;
  onBlur?: (() => void) | undefined;
  onAddQuery?: ((query: any) => void) | undefined;
  data?: PanelData | undefined;
  range?: TimeRange | undefined;
  exploreId?: any;
  history?: HistoryItem<any>[] | undefined;
  queries?: DataQuery[] | undefined;
  app?: CoreApp | undefined;

  // ****************************** Options **********************************

  getRegions() {
    return this.datasource.getRegions(this.target).then((regions: any) => {
      return this.appendVariables([ ...regions], regionsQueryRegex);
    });
  }

  getTenancies() {
    return this.datasource.getTenancies().then((tenancies: any) => {
      return this.appendVariables([ ...tenancies], tenanciesQueryRegex);
    });
  }

  getCompartments() {
    return this.datasource.getCompartments(this.target).then((compartments: any) => {
      return this.appendVariables([...compartments], compartmentsQueryRegex);
    });
  }

  getNamespaces() {
    return this.datasource.getNamespaces(this.target).then((namespaces: any) => {
      return this.appendVariables([...namespaces], namespacesQueryRegex);
    });
  }

  getResourceGroups() {
    return this.datasource.getResourceGroups(this.target).then((resourcegroups: any) => {
      return this.appendVariables([...resourcegroups], resourcegroupsQueryRegex);
    });
  }

  getMetrics() {
    return this.datasource.metricFindQuery(this.target).then((metrics: any) => {
      return this.appendVariables([...metrics], metricsQueryRegex);
    });
  }

  getAggregations() {
    return this.datasource.getAggregations().then((aggs: any[]) => {
      return aggs.map((val: any) => ({ text: val, value: val }));
    });
  }

  getWindows () {
    return this.appendWindowsAndResolutionVariables([...windows], windowsAndResolutionRegex)
  }

  getResolutions () {
    return this.appendWindowsAndResolutionVariables([...resolutions], windowsAndResolutionRegex)
  }

  /**
   * Get options for the dimension segment: of type 'key' or type 'value'
   * @param segment 
   * @param index 
   */
  getDimensionOptions(segment: { type: string; }, index: number) {
    if (segment.type === 'key' || segment.type === 'plus-button') {
      return this.getDimensionsCache().then((cache: {}) => {
        const keys = Object.keys(cache);
        const vars = this.datasource.getVariables(dimensionKeysQueryRegex) || [];
        const keysWithVariables = vars.concat(keys);
        const segments = keysWithVariables.map((key: any) => this.uiSegmentSrv.newSegment({ value: key }));
        segments.unshift(this.removeDimensionSegment);
        return segments;
      });
    }

    if (segment.type === 'value') {
      return this.getDimensionsCache().then((cache: { [x: string]: never[]; }) => {
        const keySegment = this.dimensionSegments[index - 2];
        const key = this.datasource.getVariableValue(keySegment.value);
        const options = cache[key] || [];

        // return all the values for the key
        const vars = this.datasource.getVariables(dimensionValuesQueryRegex) || [];
        const optionsWithVariables = vars.concat(options);
        const segments = optionsWithVariables.map((v: any) => this.uiSegmentSrv.newSegment({ value: v }));
        return segments;
      });
    }

    return this.q.when([]);
  }

  getDimensionsCache() {
    const targetSelector = JSON.stringify({
      region: this.datasource.getVariableValue(this.target.region),
      compartment: this.datasource.getVariableValue(this.target.compartment),
      namespace: this.datasource.getVariableValue(this.target.namespace),
      resourcegroup: this.datasource.getVariableValue(this.target.resourcegroup),
      metric: this.datasource.getVariableValue(this.target.metric)
    });

    if (this.dimensionsCache[targetSelector]) {
      return this.q.when(this.dimensionsCache[targetSelector]);
    }

    return this.datasource.getDimensions(this.target).then((dimensions: any[]) => {
      const cache = dimensions.reduce((data: { [x: string]: any[]; }, item: { value: { split: (arg0: string) => never[]; }; }) => {
        const values = item.value.split('=') || [];
        const key = values[0] || item.value;
        const value = values[1];

        if (!data[key]) {
          data[key] = []
        }
        data[key].push(value);
        return data;
      }, {});
      this.dimensionsCache[targetSelector] = cache;
      return this.dimensionsCache[targetSelector];
    })
  }

  appendVariables(options: any[], varQueryRegex: RegExp) {
    const vars = this.datasource.getVariables(varQueryRegex) || [];
    vars.forEach((value: any) => {
      options.unshift({ value, text: value });
    });
    return options;
  }

  appendWindowsAndResolutionVariables (options: string[], varQueryRegex: RegExp) {
    const vars = this.datasource.getVariables(varQueryRegex) || []
    return [...options, ...vars].map(value => ({ value, text: value }))
  }
  // ****************************** Callbacks **********************************

  toggleEditorMode() {
    this.target.rawQuery = !this.target.rawQuery;
  }

  onChangeInternal() {
    this.panelCtrl.refresh(); // Asks the panel to refresh data.
    const namespc=(this.datasource.getNamespaces(this.target))
    namespc.then((value: string | any[]) => {
      if (value.length === 0){
        this.target.namespace = SELECT_PLACEHOLDERS.NAMESPACE;
        this.panelCtrl.refresh(); // Asks the panel to refresh data.
      }       
    });
  }


  /**
   * On dimension segment change callback
   * @param segment 
   * @param index 
   */
  onDimensionsChange(segment: { value: string | number; type: string; cssClass: string; }, index: number) {
    if (segment.value === this.removeDimensionSegment.value) {
      // remove dimension: key - op - value
      this.dimensionSegments.splice(index, 3);
      // remove last comma
      if (this.dimensionSegments.length > 2) {
        this.dimensionSegments.splice(Math.max(index - 1, 0), 1);
      }
    } else if (segment.type === 'plus-button') {
      if (index > 2) {
        // add comma in front of plus button
        this.dimensionSegments.splice(index, 0, this.uiSegmentSrv.newCondition(','))
      }
      // replace plus button with key segment
      segment.type = 'key';
      segment.cssClass = 'query-segment-key';
      this.dimensionSegments.push(this.getDimensionOperatorSegment());
      this.dimensionSegments.push(this.getSelectDimensionValueSegment());
    } else if (segment.type === 'key') {
      this.getDimensionsCache().then((cache: { [x: string]: never[]; }) => {
        //update value to be part of the available options
        const value = this.dimensionSegments[index + 2].value;
        const options = cache[segment.value] || [];
        if (!this.datasource.isVariable(value) && options.indexOf(value) < 0) {
          this.dimensionSegments[index + 2] = this.getSelectDimensionValueSegment();
        }

        this.updateQueryWithDimensions();
      });
    }

    // add plus button at the end
    if (this.dimensionSegments.length === 0 || this.dimensionSegments[this.dimensionSegments.length - 1].type !== 'plus-button') {
      this.dimensionSegments.push(this.uiSegmentSrv.newPlusButton());
    }

    this.updateQueryWithDimensions();
  }

  /**
   * Collect data from  dimension segments to pass to query
   */
  updateQueryWithDimensions() {
    const dimensions: any = [];
    let index: number;

    this.dimensionSegments.forEach((s: { type: string; value: any; }) => {
      if (s.type === 'key') {
        if (dimensions.length === 0) {
          dimensions.push({});
          index = 0;
        }
        dimensions[index].key = s.value;
      } else if (s.type === 'value') {
        dimensions[index].value = s.value;
      } else if (s.type === 'condition') {
        dimensions.push({});
        index++;
      } else if (s.type === 'operator') {
        dimensions[index].operator = s.value;
      }
    });

    this.target.dimensions = dimensions;
    this.panelCtrl.refresh();
  }
}

OCIDatasourceQueryCtrl.templateUrl = 'partials/query.editor.html'
