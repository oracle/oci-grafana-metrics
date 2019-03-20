import { QueryCtrl } from 'app/plugins/sdk'
import './css/query-editor.css!'
import { windows } from './constants'
import _ from 'lodash'

export class OCIDatasourceQueryCtrl extends QueryCtrl {
  constructor ($scope, $injector, $q, uiSegmentSrv) {
    super($scope, $injector)

    this.scope = $scope
    this.uiSegmentSrv = uiSegmentSrv
    this.target.region = this.target.region || 'select region'
    this.target.compartment = this.target.compartment || 'select compartment'
    this.target.resolution = this.target.resolution || '1m'
    this.target.namespace = this.target.namespace || 'select namespace'
    this.target.window = this.target.window || '1m'
    this.target.metric = this.target.metric || ''
    this.target.aggregation = this.target.aggregation || 'mean()'
    this.target.tags = this.target.tags || []
    this.q = $q

    this.target.dimension = this.target.dimension || ''

    this.tagSegments = []
    this.dimCache = {}
    this.removeTagFilterSegment = uiSegmentSrv.newSegment({
      fake: true,
      value: '-- remove tag filter --'
    })

    for (let i = 0; i < this.target.tags.length; i++) {
      if (i > 0) {
        this.tagSegments.push(this.uiSegmentSrv.newCondition(','))
      }
      const obj = this.target.tags[i]
      this.tagSegments.push(this.uiSegmentSrv.newSegment({
        fake: false,
        key: obj.key,
        value: obj.key,
        type: 'key'
      }))
      this.tagSegments.push(this.uiSegmentSrv.newSegment({
        fake: false,
        key: obj.operator,
        type: 'operator',
        value: obj.operator
      }))
      this.tagSegments.push(this.uiSegmentSrv.newSegment({
        fake: false,
        key: obj.value,
        type: 'value',
        value: obj.value
      }))
    }
    this.tagSegments.push(this.uiSegmentSrv.newPlusButton())
  }

  toggleEditorMode () {
    this.target.rawQuery = !this.target.rawQuery
  }

  getNamespaces () {
    return this.datasource.getNamespaces(this.target)
      .then((namespaces) => {
        namespaces.push({ text: '$namespace', value: '$namespace' })
        return namespaces
      })
  }

  getMetrics () {
    return this.datasource.metricFindQuery(this.target)
      .then((metrics) => {
        metrics.push({ text: '$metric', value: '$metric' })
        return metrics
      })
  }

  getAggregations () {
    return this.datasource.getAggregations().then((aggs) => {
      return aggs.map((val) => {
        return { text: val, value: val }
      })
    })
  }

  onChangeInternal () {
    this.panelCtrl.refresh() // Asks the panel to refresh data.
  }

  getRegions () {
    return this.datasource.getRegions()
      .then((regs) => {
        regs.push({ text: '$region', value: '$region' })
        return regs
      }).catch((err) => { console.error(err) })
  }

  getCompartments () {
    return this.datasource.getCompartments()
      .then((item) => {
        item.push({ text: '$compartment', value: '$compartment' })
        return item
      })
  }

  getWindows () {
    return windows
  }

  getDimensions () {
    return this.datasource.getDimensions(this.target)
  }

  handleQueryError (err) {
    this.error = err.message || 'Failed to issue metric query'
    return []
  }

  getTagsOrValues (segment, index) {
    if (segment.type === 'operator') {
      return this.q.when([])
    }

    if (segment.type === 'key' || segment.type === 'plus-button') {
      return this.getDimensions()
        .then(this.mapToSegment.bind(this))
        .catch(this.handleQueryError.bind(this))
    }
    const key = this.tagSegments[index - 2]
    const options = this.dimCache[key.value]
    const that = this
    const optSegments = options.map(v => that.uiSegmentSrv.newSegment({
      value: v
    }))
    return this.q.when(optSegments)
  }

  mapToSegment (dimensions) {
    const dimCache = {}
    const dims = dimensions.map((v) => {
      const values = v.text.split('=')
      const key = values[0]
      const value = values[1]
      if (!(key in dimCache)) {
        dimCache[key] = []
      }
      dimCache[key].push(value)
      return this.uiSegmentSrv.newSegment({
        value: values[0]
      })
    })
    dims.unshift(this.removeTagFilterSegment)
    this.dimCache = dimCache
    return dims
  }

  tagSegmentUpdated (segment, index) {
    this.tagSegments[index] = segment

    // handle remove tag condition
    if (segment.value === this.removeTagFilterSegment.value) {
      this.tagSegments.splice(index, 3)
      if (this.tagSegments.length === 0) {
        this.tagSegments.push(this.uiSegmentSrv.newPlusButton())
      } else if (this.tagSegments.length > 2) {
        this.tagSegments.splice(Math.max(index - 1, 0), 1)
        if (this.tagSegments[this.tagSegments.length - 1].type !== 'plus-button') {
          this.tagSegments.push(this.uiSegmentSrv.newPlusButton())
        }
      }
    } else {
      if (segment.type === 'plus-button') {
        if (index > 2) {
          this.tagSegments.splice(index, 0, this.uiSegmentSrv.newCondition(','))
        }
        this.tagSegments.push(this.uiSegmentSrv.newOperator('='))
        this.tagSegments.push(this.uiSegmentSrv.newFake('select tag value', 'value', 'query-segment-value'))
        segment.type = 'key'
        segment.cssClass = 'query-segment-key'
      }

      if (index + 1 === this.tagSegments.length) {
        this.tagSegments.push(this.uiSegmentSrv.newPlusButton())
      }
    }

    this.rebuildTargetTagConditions()
  }

  rebuildTargetTagConditions () {
    const tags = []
    let tagIndex = 0

    _.each(this.tagSegments, (segment2, index) => {
      if (segment2.type === 'key') {
        if (tags.length === 0) {
          tags.push({})
        }
        tags[tagIndex].key = segment2.value
      } else if (segment2.type === 'value') {
        tags[tagIndex].value = segment2.value
      } else if (segment2.type === 'condition') {
        tags.push({ condition: segment2.value })
        tagIndex += 1
      } else if (segment2.type === 'operator') {
        tags[tagIndex].operator = segment2.value
      }
    })

    this.target.tags = tags
    this.panelCtrl.refresh()
  }
}

OCIDatasourceQueryCtrl.templateUrl = 'partials/query.editor.html'
