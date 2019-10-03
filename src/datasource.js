/*
** Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
import _ from 'lodash'
import { aggregations, namespaces } from './constants'
import retryOrThrow from './util/retry'

export default class OCIDatasource {
  constructor(instanceSettings, $q, backendSrv, templateSrv, timeSrv) {
    this.type = instanceSettings.type
    this.url = instanceSettings.url
    this.name = instanceSettings.name
    this.id = instanceSettings.id
    this.tenancyOCID = instanceSettings.jsonData.tenancyOCID
    this.defaultRegion = instanceSettings.jsonData.defaultRegion
    this.environment = instanceSettings.jsonData.environment
    this.q = $q
    this.backendSrv = backendSrv
    this.templateSrv = templateSrv
    this.timeSrv = timeSrv
  }

  query(options) {
    var query = this.buildQueryParameters(options)
    query.targets = query.targets.filter(t => !t.hide)
    if (query.targets.length <= 0) {
      return this.q.when({ data: [] })
    }

    return this.doRequest(query)
      .then(result => {
        var res = []
        _.forEach(result.data.results, r => {
          _.forEach(r.series, s => {
            res.push({ target: s.name, datapoints: s.points })
          })
          _.forEach(r.tables, t => {
            t.type = 'table'
            t.refId = r.refId
            res.push(t)
          })
        })

        result.data = res
        return result
      })
  }

  buildQueryParameters(options) {
    // remove placeholder targets
    options.targets = _.filter(options.targets, target => {
      return target.metric !== 'select metric'
    })

    var targets = _.map(options.targets, target => {
      let region = target.region
      // check to see if we have tags (dimensions) in the query and turn those into a string for MQL
      let t = []
      if (target.hasOwnProperty('tags')) {
        for (let i = 0; i < target.tags.length; i++) {
          if (target.tags[i].value !== 'select tag value') {
            t.push(`${target.tags[i].key} ${target.tags[i].operator} "${target.tags[i].value}"`)
          }
        }
        t.join()
      }

      if (target.region === 'select region') {
        region = this.defaultRegion
      }
      // If there's nothing there return a blank string otherwise return the dimensions encapsuled by these {}
      let dimension = (t.length === 0) ? '' : `{${t}}`

      return {
        compartment: this.templateSrv.replace(target.compartment, options.scopedVars || {}),
        environment: this.environment,
        queryType: 'query',
        region: this.templateSrv.replace(region, options.scopedVars || {}),
        tenancyOCID: this.tenancyOCID,
        namespace: this.templateSrv.replace(target.namespace, options.scopedVars || {}),
        resolution: target.resolution,
        refId: target.refId,
        hide: target.hide,
        type: target.type || 'timeserie',
        datasourceId: this.id,
        // pass the MQL string we built here
        query: `${this.templateSrv.replace(target.metric, options.scopedVars || {})}[${target.window}]${dimension}.${target.aggregation}`
      }
    })

    options.targets = targets

    return options
  }

  testDatasource() {
    return this.doRequest({
      targets: [{
        queryType: 'test',
        region: this.defaultRegion,
        tenancyOCID: this.tenancyOCID,
        compartment: '',
        environment: this.environment,
        datasourceId: this.id
      }],
      range: this.timeSrv.timeRange()
    }).then((response) => {
      if (response.status === 200) {
        return { status: 'success', message: 'Data source is working', title: 'Success' }
      }
    }).catch(() => {
      return { status: 'error', message: 'Data source is not working', title: 'Failure' }
    })
  }

  // helps match the regex's from creating template variables in grafana
  templateMeticSearch(varString) {
    let compartmentQuery = varString.match(/^compartments\(\)/)
    if (compartmentQuery) {
      return this.getCompartments().catch(err => { throw new Error('Unable to make request ' + err) })
    }

    let regionQuery = varString.match(/^regions\(\)/)
    if (regionQuery) {
      return this.getRegions().catch(err => { throw new Error('Unable to make request ' + err) })
    }

    let metricQuery = varString.match(/metrics\((\s*\$?\w+)(\s*,\s*\$\w+)(\s*,\s*\$\w+\s*)*\)/)
    if (metricQuery) {
      let target = {
        region: this.templateSrv.replace(metricQuery[1].trim()),
        compartment: this.templateSrv.replace(metricQuery[2].replace(',', '').trim()),
        namespace: this.templateSrv.replace(metricQuery[3].replace(',', '').trim())
      }
      return this.metricFindQuery(target).catch(err => { throw new Error('Unable to make request ' + err) })
    }

    let namespaceQuery = varString.match(/namespaces\((\$?\w+)(,\s*\$\w+)*\)/)
    if (namespaceQuery) {
   
      let target = {
        region: this.templateSrv.replace(namespaceQuery[1]),
        compartment: this.templateSrv.replace(namespaceQuery[2]).replace(',', '').trim()
      }
      return this.getNamespaces(target).catch(err => { throw new Error('Unable to make request(get namespaces) ' + err) })
    }
    throw new Error('Unable to parse templating string')
  }

  // this function does 2 things - its the entrypoint for finding the metrics from the query editor
  // and is the entrypoint for templating according to grafana -- since this wasn't docuemnted
  // in grafana I was lead to believe that it did the former
  // TODO: break the metric finding for the query editor out into a different function
  metricFindQuery(target) {
    if (typeof (target) === 'string') {
      return this.templateMeticSearch(target)
    }

    var range = this.timeSrv.timeRange()
    let region = this.defaultRegion
    if (target.namespace === 'select namespace') {
      target.namespace = ''
    }
    if (target.compartment === 'select compartment') {
      target.compartment = ''
    }
    if (Object.hasOwnProperty(target, 'region') && target.region !== 'select region') {
      region = target.region
    }

    var targets = [{
      compartment: this.templateSrv.replace(target.compartment),
      environment: this.environment,
      queryType: 'search',
      tenancyOCID: this.tenancyOCID,
      region: this.templateSrv.replace(region),
      datasourceId: this.id,
      namespace: this.templateSrv.replace(target.namespace)
    }]
    var options = {
      range: range,
      targets: targets
    }
    return this.doRequest(options).then((res) => {
      return this.mapToTextValue(res, 'search')
    })
  }

  mapToTextValue(result, searchField) {
    var table = result.data.results[searchField].tables[0]
    if (!table) {
      return []
    }

    var m = _.map(table.rows, (row, i) => {
      if (row.length > 1) {
        return { text: row[0], value: row[1] }
      } else if (_.isObject(row[0])) {
        return { text: row[0], value: i }
      }
      return { text: row[0], value: row[0] }
    })
    return m
  }

  getCompartments() {
    var range = this.timeSrv.timeRange()
    var targets = [{
      environment: this.environment,
      region: this.defaultRegion,
      tenancyOCID: this.tenancyOCID,
      queryType: 'compartments',
      datasourceId: this.id
    }]
    var options = {
      range: range,
      targets: targets
    }
    return this.doRequest(options).then((res) => { return this.mapToTextValue(res, 'compartment') })
  }

  getDimensions(target) {
    var range = this.timeSrv.timeRange()
    let region = target.region
    if (target.namespace === 'select namespace') {
      target.namespace = ''
    }
    if (target.compartment === 'select compartment') {
      target.compartment = ''
    }
    if (target.metric === 'select metric') {
      return []
    }
    if (region === 'select region') {
      region = this.defaultRegion
    }

    var targets = [{
      compartment: this.templateSrv.replace(target.compartment),
      environment: this.environment,
      queryType: 'dimensions',
      region: this.templateSrv.replace(region),
      tenancyOCID: this.tenancyOCID,

      datasourceId: this.id,
      metric: this.templateSrv.replace(target.metric),
      namespace: this.templateSrv.replace(target.namespace)
    }]

    var options = {
      range: range,
      targets: targets
    }
    return this.doRequest(options).then((res) => { return this.mapToTextValue(res, 'dimensions') })
  }

  getNamespaces(target) {
    let region = target.region
    if (region === 'select region') {
      region = this.defaultRegion
    }
    return this.doRequest({
      targets: [{
        // commonRequestParameters
        compartment: this.templateSrv.replace(target.compartment),
        environment: this.environment,
        queryType: 'namespaces',
        region: this.templateSrv.replace(region),
        tenancyOCID: this.tenancyOCID,

        datasourceId: this.id
      }],
      range: this.timeSrv.timeRange()
    }).then((namespaces) => { return this.mapToTextValue(namespaces, 'namespaces') })
  }

  getRegions() {
    return this.doRequest({
      targets: [{
        environment: this.environment,
        queryType: 'regions',
        datasourceId: this.id
      }],
      range: this.timeSrv.timeRange()
    }).then((regions) => { return this.mapToTextValue(regions, 'regions') })
  }

  getAggregations() {
    return this.q.when(aggregations)
  }

  // retries all request to the backend grafana 10 times before failure
  doRequest(options) {
    let _this = this
    return retryOrThrow(() => {
      return _this.backendSrv.datasourceRequest({
        url: '/api/tsdb/query',
        method: 'POST',
        data: {
          from: options.range.from.valueOf().toString(),
          to: options.range.to.valueOf().toString(),
          queries: options.targets
        }
      })
    }, 10)
  }
}
