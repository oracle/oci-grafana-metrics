import { getWindowAndResolution, resolveAutoWinRes } from './utilFunctions'
import {
  autoTimeIntervals,
  SEVEN_DAYS,
  THIRTY_DAYS,
  d8To30Config,
  d0To7Config,
  d31toInfConfig,
  AUTO
} from '../constants'

describe('getWindowAndResolution Tests : Test for config generation on days given', () => {
  test('getWindowAndResolution : 0 days', () => {
    expect(getWindowAndResolution(autoTimeIntervals, SEVEN_DAYS))
      .toMatchObject(d0To7Config)
  })
  test('getWindowAndResolution : 1 day', () => {
    expect(getWindowAndResolution(autoTimeIntervals, SEVEN_DAYS))
      .toMatchObject(d0To7Config)
  })
  test('getWindowAndResolution : 7 days', () => {
    expect(getWindowAndResolution(autoTimeIntervals, SEVEN_DAYS))
      .toMatchObject(d0To7Config)
  })
  test('getWindowAndResolution : 14 days', () => {
    const D14 = '14'
    expect(getWindowAndResolution(autoTimeIntervals, D14))
      .toMatchObject(d8To30Config)
  })
  test('getWindowAndResolution : 30 days', () => {
    expect(getWindowAndResolution(autoTimeIntervals, THIRTY_DAYS))
      .toMatchObject(d8To30Config)
  })

  test('getWindowAndResolution : 41 days', () => {
    const D999 = '999'
    expect(getWindowAndResolution(autoTimeIntervals, D999))
      .toMatchObject(d31toInfConfig)
  })
})

describe('resolveAutoWinRes Tests : Test to check replacement of auto with time duration with time duration ' +
  'if auto is selected', () => {
  const TIME_IN_DURATION = '5m'

  test('resolveAutoWinRes : with auto mode on ', () => {
    expect(resolveAutoWinRes(AUTO, AUTO, SEVEN_DAYS))
      .toMatchObject(d0To7Config)
  })

  test('resolveAutoWinRes : with auto mode only on window, resolution given a time duration', () => {
    expect(resolveAutoWinRes(AUTO, TIME_IN_DURATION, SEVEN_DAYS))
      .toMatchObject({ window: d0To7Config.window, resolution: TIME_IN_DURATION })
  })

  test('resolveAutoWinRes : with auto mode only on resolution, window given a time duration', () => {
    expect(resolveAutoWinRes(TIME_IN_DURATION, AUTO, SEVEN_DAYS))
      .toMatchObject({ window: TIME_IN_DURATION, resolution: d0To7Config.resolution })
  })

  test('resolveAutoWinRes : with no auto', () => {
    expect(resolveAutoWinRes(TIME_IN_DURATION, TIME_IN_DURATION, SEVEN_DAYS))
      .toMatchObject({ window: TIME_IN_DURATION, resolution: TIME_IN_DURATION })
  })
}
)
