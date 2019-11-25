/*
** Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
*/
import OCIDatasource from './datasource'

describe('OCIDatasource', () => {
  const instanceSettings = {
    jsonData: {
      tenancyOCID: 'ocid1.tenancy.oc1..aaaa5sdafasdfadsf2e3d2qmq4a',
      defaultRegion: 'us-ashburn-1',
      environment: 'local'
    }
  }

  describe('building a query', () => {
    const backendSrvMock = { datasourceRequest: jest.fn() }
    const templateSrvMock = {
      getAdhocFilters: () => [],
      replace: a => a
    }
    test('should populate a region if none is there', () => {
      const options = {
        targets: [
          {
            region: 'select region',
            compartment: 'ocid1.compartment.oc1..aaaa5sdafasdfadsf2e3d2qmq4a',
            resolution: '1m',
            namespace: 'oci_computeagent',
            window: '1m',
            metric: 'CpuUtilization',
            aggregation: 'mean()',
            dimensions: []
          }
        ] }
      const ds = new OCIDatasource(instanceSettings, null, backendSrvMock, templateSrvMock)
      let opts = ds.buildQueryParameters(options)
      expect(opts.targets.length).toBe(1)
      expect(opts.targets[0].region).toMatch(/us-ashburn-1/)
    })
    test('should not override a region if none is there', () => {
      const options = {
        targets: [
          {
            region: 'uk-london-1',
            compartment: 'ocid1.compartment.oc1..aaaa5sdafasdfadsf2e3d2qmq4a',
            resolution: '1m',
            namespace: 'oci_computeagent',
            window: '1m',
            metric: 'CpuUtilization',
            aggregation: 'mean()',
            dimensions: []
          }
        ] }
      const ds = new OCIDatasource(instanceSettings, null, backendSrvMock, templateSrvMock)
      let opts = ds.buildQueryParameters(options)
      expect(opts.targets.length).toBe(1)
      expect(opts.targets[0].region).toMatch(/uk-london-1/)
    })
    test('should not have dimensions in query if there are no dimensions', () => {
      const options = {
        targets: [
          {
            region: 'uk-london-1',
            compartment: 'ocid1.compartment.oc1..aaaa5sdafasdfadsf2e3d2qmq4a',
            resolution: '1m',
            namespace: 'oci_computeagent',
            window: '1m',
            metric: 'CpuUtilization',
            aggregation: 'mean()',
            dimensions: []
          }
        ] }
      const ds = new OCIDatasource(instanceSettings, null, backendSrvMock, templateSrvMock)
      let opts = ds.buildQueryParameters(options)
      expect(opts.targets.length).toBe(1)
      expect(opts.targets[0].query).toMatch(/CpuUtilization\[1m\]\.mean\(\)/)
    })
    test('should have dimensions in query if there are dimensions', () => {
      const options = {
        targets: [
          {
            region: 'uk-london-1',
            compartment: 'ocid1.compartment.oc1..aaaa5sdafasdfadsf2e3d2qmq4a',
            resolution: '1m',
            namespace: 'oci_computeagent',
            window: '1m',
            metric: 'CpuUtilization',
            aggregation: 'mean()',
            dimensions: [{ key: 'key', operator: '=', value: 'value' }]
          }
        ] }
      const ds = new OCIDatasource(instanceSettings, null, backendSrvMock, templateSrvMock)
      let opts = ds.buildQueryParameters(options)
      expect(opts.targets.length).toBe(1)
      expect(opts.targets[0].query).toMatch(/CpuUtilization\[1m\]\{key = "value"}.mean\(\)/)
    })
  })
  describe('testing the datasource', () => {
    const templateSrvMock = {
      getAdhocFilters: () => [],
      replace: a => a
    }
    const timeSrvMock = {
      timeRange: () => { return { from: { valueOf: () => '' }, to: { valueOf: () => '' } } }
    }
    test('should handle errors gracefully', (done) => {
      const backendSrvMock = { datasourceRequest: jest.fn(() => { throw new Error('Network Request') }) }
      const myDS = new OCIDatasource(instanceSettings, null, backendSrvMock, templateSrvMock, timeSrvMock)
      return myDS.testDatasource().then((err) => {
        expect(err).toEqual({ status: 'error', message: 'Data source is not working', title: 'Failure' })
        done()
      })
    })

    test('should return success on a 200', (done) => {
      const backendSrvMock = { datasourceRequest: jest.fn(() => Promise.resolve({ status: 200 })) }
      const myDS = new OCIDatasource(instanceSettings, null, backendSrvMock, templateSrvMock, timeSrvMock)
      return myDS.testDatasource().then((data) => {
        expect(data).toEqual({ status: 'success', message: 'Data source is working', title: 'Success' })
        done()
      })
    })
  })
})
