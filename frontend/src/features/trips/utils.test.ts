import {
  averageDistancePerDay,
  cleanRouteTitle,
  computeTripStats,
  daysToInterval,
  detectRouteService,
  formatDurationDays,
  formatTripType,
  getTripRouteService,
  getTripRouteUrl,
  intervalToDays,
  serviceToUrlField,
} from './utils'
import type { Trip, TripWithDetails } from './types'

describe('formatTripType', () => {
  it('maps known types and falls back to the raw value', () => {
    expect(formatTripType('day_hike')).toBe('Day Hike')
    expect(formatTripType('thru_hike')).toBe('Thru Hike')
    expect(formatTripType('mystery' as never)).toBe('mystery')
  })
})

describe('detectRouteService', () => {
  it('detects services from the host', () => {
    expect(detectRouteService('https://www.komoot.com/tour/123')).toBe('komoot')
    expect(detectRouteService('https://komoot.de/tour/1')).toBe('komoot')
    expect(detectRouteService('https://strava.app.link/abc')).toBe('strava')
    expect(detectRouteService('https://www.strava.com/activities/1')).toBe('strava')
    expect(detectRouteService('https://wanderer.to/t/1')).toBe('wanderer')
  })

  it('returns "unknown" for empty, unparseable, or unrecognized URLs', () => {
    expect(detectRouteService('')).toBe('unknown')
    expect(detectRouteService('   ')).toBe('unknown')
    expect(detectRouteService('not a url')).toBe('unknown')
    expect(detectRouteService('https://example.com/x')).toBe('unknown')
  })
})

describe('serviceToUrlField', () => {
  it('maps the URL onto the selected service field', () => {
    expect(serviceToUrlField('komoot', 'u')).toEqual({ trip_komoot_url: 'u' })
    expect(serviceToUrlField('strava', 'u')).toEqual({ trip_strava_url: 'u' })
    expect(serviceToUrlField('wanderer', 'u')).toEqual({ trip_wanderer_url: 'u' })
  })

  it('returns {} for an empty URL or unknown service', () => {
    expect(serviceToUrlField('komoot', '   ')).toEqual({})
    expect(serviceToUrlField('unknown', 'u')).toEqual({})
  })
})

describe('trip route URL helpers', () => {
  const trip = (over: Partial<Trip>): Trip => over as Trip

  it('getTripRouteUrl returns the first populated field', () => {
    expect(getTripRouteUrl(trip({ trip_strava_url: 's' }))).toBe('s')
    expect(getTripRouteUrl(trip({}))).toBe('')
  })

  it('getTripRouteService reflects which field is populated', () => {
    expect(getTripRouteService(trip({ trip_komoot_url: 'k' }))).toBe('komoot')
    expect(getTripRouteService(trip({ trip_strava_url: 's' }))).toBe('strava')
    expect(getTripRouteService(trip({ trip_wanderer_url: 'w' }))).toBe('wanderer')
    expect(getTripRouteService(trip({}))).toBe('unknown')
  })
})

describe('cleanRouteTitle', () => {
  it('keeps the first segment before a pipe', () => {
    expect(cleanRouteTitle('My Hike | Hiking | Komoot')).toBe('My Hike')
    expect(cleanRouteTitle('  Trail  ')).toBe('Trail')
  })

  it('returns "" for nullish titles', () => {
    expect(cleanRouteTitle(null)).toBe('')
    expect(cleanRouteTitle(undefined)).toBe('')
  })
})

describe('duration interval helpers', () => {
  it('daysToInterval builds a PG interval or undefined', () => {
    expect(daysToInterval(3)).toBe('3 days')
    expect(daysToInterval('5')).toBe('5 days')
    expect(daysToInterval(0)).toBeUndefined()
    expect(daysToInterval(-1)).toBeUndefined()
    expect(daysToInterval('abc')).toBeUndefined()
  })

  it('intervalToDays parses days from an interval string', () => {
    expect(intervalToDays('3 days')).toBe(3)
    expect(intervalToDays('1 day')).toBe(1)
    expect(intervalToDays('3:00:00')).toBe(3)
    expect(intervalToDays('nope')).toBeNull()
    expect(intervalToDays(null)).toBeNull()
  })

  it('formatDurationDays renders a human label', () => {
    expect(formatDurationDays('1 day')).toBe('1 day')
    expect(formatDurationDays('3 days')).toBe('3 days')
    expect(formatDurationDays(null)).toBe('Not set')
  })
})

describe('averageDistancePerDay', () => {
  it('divides distance by parsed days, else null', () => {
    expect(averageDistancePerDay(30, '3 days')).toBe(10)
    expect(averageDistancePerDay(null, '3 days')).toBeNull()
    expect(averageDistancePerDay(30, null)).toBeNull()
    expect(averageDistancePerDay(30, '0 days')).toBeNull()
  })
})

describe('computeTripStats', () => {
  it('aggregates pack + worn weight, value, and counts over a nested trip', () => {
    const trip = {
      persons: [
        {
          packs: [
            {
              items: [
                { quantity: 2, carry_status: 'packed', item: { weight_grams: 100, value: 5 } },
                { quantity: 1, carry_status: 'worn', item: { weight_grams: 50, value: 0 } },
              ],
            },
          ],
          items: [
            { quantity: 1, carry_status: 'packed', item: { weight_grams: 300, value: 10 } },
          ],
        },
      ],
    } as unknown as TripWithDetails

    const stats = computeTripStats(trip)
    expect(stats.packedWeightGrams).toBe(500) // 2*100 + 300
    expect(stats.wornWeightGrams).toBe(50)
    expect(stats.totalWeightGrams).toBe(550)
    expect(stats.totalValue).toBe(20) // 2*5 + 0 + 10
    expect(stats.packsCount).toBe(1)
    expect(stats.travelersCount).toBe(1)
  })

  it('handles missing weights/values as zero', () => {
    const trip = {
      persons: [
        { packs: [], items: [{ quantity: 3, carry_status: 'packed', item: {} }] },
      ],
    } as unknown as TripWithDetails

    const stats = computeTripStats(trip)
    expect(stats.totalWeightGrams).toBe(0)
    expect(stats.totalValue).toBe(0)
    expect(stats.packsCount).toBe(0)
  })
})
