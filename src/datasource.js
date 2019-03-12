import _ from 'lodash'
import { namespaces } from './constants'
import retryOrThrow from './util/retry'

export default class OCIDatasource {
  constructor (instanceSettings, $q, backendSrv, templateSrv, timeSrv) {
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

  query (options) {
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

  buildQueryParameters (options) {
    // remove placeholder targets
    options.targets = _.filter(options.targets, target => {
      return target.metric !== 'select metric'
    })

    var targets = _.map(options.targets, target => {
      let region = target.region
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
        query: `${this.templateSrv.replace(target.metric, options.scopedVars || {})}[${target.window}]${dimension}.${target.aggregation}`
      }
    })

    options.targets = targets

    return options
  }

  testDatasource () {
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

  templateMeticSearch (varString) {
    let compartmentQuery = varString.match(/^compartments\(\)/)
    if (compartmentQuery) {
      return this.getCompartments()
    }

    let regionQuery = varString.match(/^regions\(\)/)
    if (regionQuery) {
      return this.getRegions()
    }

    let metricQuery = varString.match(/metrics\((\$?\w+)(,\s*\$\w+)*\)/)
    if (metricQuery) {
      let target = {
        namespace: this.templateSrv.replace(metricQuery[1]),
        compartment: this.templateSrv.replace(metricQuery[2]).replace(',', '').trim()
      }
      return this.metricFindQuery(target)
    }

    let namespaceQuery = varString.match(/namespaces\(\)/)
    if (namespaceQuery) {
      let names = namespaces.map((reg) => {
        return { row: reg, value: reg }
      })
      return this.q.when(names)
    }
    throw new Error('Unable to parse templating string')
  }

  metricFindQuery (target) {
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

  mapToTextValue (result, searchField) {
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

  getCompartments () {
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

  getDimensions (target) {
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

  getNamespaces (target) {
    let region = target.region
    if (region === 'select region') {
      region = this.defaultRegion
    }
    return this.doRequest({ targets: [{
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

  getRegions () {
    return this.doRequest({ targets: [{
      environment: this.environment,
      queryType: 'regions',
      datasourceId: this.id
    }],
    range: this.timeSrv.timeRange()
    }).then((regions) => { return this.mapToTextValue(regions, 'regions') })
  }

  doRequest (options) {
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
