/*
** Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
import { QueryCtrl } from 'app/plugins/sdk'
import './css/query-editor.css!'
import { windows, namespacesQueryRegex, metricsQueryRegex, regionsQueryRegex, compartmentsQueryRegex, dimensionKeysQueryRegex, dimensionValuesQueryRegex } from './constants'
import _ from 'lodash'

export const SELECT_PLACEHOLDERS = {
  DIMENSION_KEY: 'select dimension',
  DIMENSION_VALUE: 'select value',
  COMPARTMENT: 'select compartment',
  REGION: 'select region',
  NAMESPACE: 'select namespace',
  METRIC: 'select metric'
};

export class OCIDatasourceQueryCtrl extends QueryCtrl {
  constructor($scope, $injector, $q, uiSegmentSrv) {
    super($scope, $injector)

    this.q = $q;
    this.uiSegmentSrv = uiSegmentSrv;

    this.target.region = this.target.region || SELECT_PLACEHOLDERS.REGION;
    this.target.compartment = this.target.compartment || SELECT_PLACEHOLDERS.COMPARTMENT;
    this.target.namespace = this.target.namespace || SELECT_PLACEHOLDERS.NAMESPACE;
    this.target.metric = this.target.metric || SELECT_PLACEHOLDERS.METRIC;
    this.target.resolution = this.target.resolution || '1m';
    this.target.window = this.target.window || '1m';
    this.target.aggregation = this.target.aggregation || 'mean()'
    this.target.dimensions = this.target.dimensions || [];

    this.dimensionSegments = [];
    this.removeDimensionSegment = uiSegmentSrv.newSegment({ fake: true, value: '-- remove dimension --' });
    this.getSelectDimensionKeySegment = () => uiSegmentSrv.newSegment({ value: SELECT_PLACEHOLDERS.DIMENSION_KEY, type: 'key' });
    this.getDimensionOperatorSegment = () => this.uiSegmentSrv.newOperator('=');
    this.getSelectDimensionValueSegment = () => uiSegmentSrv.newSegment({ value: SELECT_PLACEHOLDERS.DIMENSION_VALUE, type: 'value' });

    this.regionsCache = [];
    this.compartmentsCache = [];
    this.dimensionsCache = {};

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

  // ****************************** Options **********************************

  getRegions() {
    if (!_.isEmpty(this.regionsCache)) {
      return this.q.when(this.regionsCache);
    }
    return this.datasource.getRegions()
      .then((regions) => {
        this.regionsCache = this.appendVariables(regions, regionsQueryRegex);
        return this.regionsCache;
      });
  }

  getCompartments() {
    if (!_.isEmpty(this.compartmentsCache)) {
      return this.q.when(this.compartmentsCache);
    }
    return this.datasource.getCompartments()
      .then((compartments) => {
        this.compartmentsCache = this.appendVariables(compartments, compartmentsQueryRegex);
        return this.compartmentsCache;
      });
  }

  getNamespaces() {
    return this.datasource.getNamespaces(this.target)
      .then((namespaces) => {
        return this.appendVariables(namespaces, namespacesQueryRegex);
      });
  }

  getMetrics() {
    return this.datasource.metricFindQuery(this.target)
      .then((metrics) => {
        return this.appendVariables(metrics, metricsQueryRegex);
      });
  }

  getAggregations() {
    return this.datasource.getAggregations().then((aggs) => {
      return aggs.map((val) => {
        return { text: val, value: val };
      })
    });
  }

  getWindows() {
    return windows;
  }

  /**
   * Get options for the dimension segment: of type 'key' or type 'value'
   * @param segment 
   * @param index 
   */
  getDimensionOptions(segment, index) {
    if (segment.type === 'key' || segment.type === 'plus-button') {
      return this.getDimensionsCache().then(cache => {
        const keys = Object.keys(cache);
        const vars = this.datasource.getVariables(dimensionKeysQueryRegex) || [];
        const keysWithVariables = vars.concat(keys);
        const segments = keysWithVariables.map(key => this.uiSegmentSrv.newSegment({ value: key }));
        segments.unshift(this.removeDimensionSegment);
        return segments;
      });
    }

    if (segment.type === 'value') {
      return this.getDimensionsCache().then(cache => {
        const keySegment = this.dimensionSegments[index - 2];
        const key = this.datasource.getVariableValue(keySegment.value);
        const options = cache[key] || [];

        // return all the values for the key
        const vars = this.datasource.getVariables(dimensionValuesQueryRegex) || [];
        const custom = this.datasource.getVariables(null, 'custom') || [];
        const optionsWithVariables = vars.concat(custom).concat(options);
        const segments = optionsWithVariables.map(v => this.uiSegmentSrv.newSegment({ value: v }));
        return segments;
      });
    }

    return this.q.when([]);
  }

  getDimensionsCache() {
    const targetSelector = JSON.stringify({
      region: this.target.region,
      compartment: this.target.compartment,
      namespace: this.target.namespace,
      metric: this.target.metric
    });

    if (this.dimensionsCache[targetSelector]) {
      return this.q.when(this.dimensionsCache[targetSelector]);
    }

    return this.datasource.getDimensions(this.target).then(dimensions => {
      const cache = dimensions.reduce((data, item) => {
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

  appendVariables(options, varQeueryRegex) {
    const vars = this.datasource.getVariables(varQeueryRegex) || [];
    vars.forEach(value => {
      options.unshift({ value, text: value });
    });
    return options;
  }

  // ****************************** Callbacks **********************************

  toggleEditorMode() {
    this.target.rawQuery = !this.target.rawQuery;
  }

  onChangeInternal() {
    this.panelCtrl.refresh(); // Asks the panel to refresh data.
  }

  /**
   * On dimension segment change callback
   * @param segment 
   * @param index 
   */
  onDimensionsChange(segment, index) {
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
      this.getDimensionsCache().then(cache => {
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
    const dimensions = [];
    let index;

    this.dimensionSegments.forEach(s => {
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
