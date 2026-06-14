import {
  calculateRecommendedMaxWeightGrams,
  formatAge,
  formatGender,
  getPersonRecommendedMaxWeightGrams,
} from './utils'
import { personFixture } from '../../test/fixtures'

// Age math reads `new Date()`; fix the clock so results are deterministic.
beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(new Date('2026-06-14T00:00:00.000Z'))
})
afterEach(() => vi.useRealTimers())

describe('formatAge', () => {
  it('returns "Not set" for missing or invalid dates', () => {
    expect(formatAge(null)).toBe('Not set')
    expect(formatAge()).toBe('Not set')
    expect(formatAge('garbage')).toBe('Not set')
  })

  it('computes whole years, adjusting for an upcoming birthday', () => {
    expect(formatAge('1990-01-01')).toBe('36') // birthday passed
    expect(formatAge('2000-07-01')).toBe('25') // birthday not yet reached this year
  })

  it('returns "Not set" for a future birthdate', () => {
    expect(formatAge('2030-01-01')).toBe('Not set')
  })
})

describe('formatGender', () => {
  it('capitalizes or falls back', () => {
    expect(formatGender('male')).toBe('Male')
    expect(formatGender('female')).toBe('Female')
    expect(formatGender(null)).toBe('Not set')
  })
})

describe('calculateRecommendedMaxWeightGrams', () => {
  it('returns 0 when body weight is missing or non-positive', () => {
    expect(calculateRecommendedMaxWeightGrams(0, '1990-01-01', 'male', 'average')).toBe(0)
    expect(calculateRecommendedMaxWeightGrams(null, null, null, null)).toBe(0)
  })

  it('applies the body × 0.12 × age × gender × conditioning formula', () => {
    // 70 kg, no birthdate (ageFactor 1.1), female (0.95), average (1):
    // 70 * 0.12 * 1.1 * 0.95 * 1 = 8.778 kg
    expect(calculateRecommendedMaxWeightGrams(70000, null, 'female', 'average')).toBeCloseTo(8778, 0)
  })

  it.each([
    ['2023-01-01', 0.75], // age 3  (<5)
    ['2019-01-01', 0.75], // age 7  (5-8)
    ['2015-01-01', 0.85], // age 11 (9-12)
    ['2012-01-01', 0.95], // age 14 (13-15)
    ['2009-01-01', 1], // age 17 (16-18)
    ['1996-01-01', 1.1], // age 30 (19-50 peak)
    ['1966-01-01', 0.9], // age 60 (50+)
  ])('uses the right age factor for birthdate %s', (birthdate, ageFactor) => {
    // Neutral gender + average conditioning isolate the age factor.
    const result = calculateRecommendedMaxWeightGrams(10000, birthdate, 'other', 'average')
    expect(result).toBeCloseTo(10 * 0.12 * ageFactor * 1 * 1 * 1000, 0)
  })

  it.each([
    ['male', 1.05],
    ['female', 0.95],
    ['other', 1],
  ])('uses the right gender factor for %s', (gender, genderFactor) => {
    const result = calculateRecommendedMaxWeightGrams(10000, '1996-01-01', gender, 'average')
    expect(result).toBeCloseTo(10 * 0.12 * 1.1 * genderFactor * 1 * 1000, 0)
  })

  it.each([
    ['sedentary', 0.85],
    ['average', 1],
    ['athletic', 1.15],
    ['military', 1.2],
  ] as const)('uses the right conditioning factor for %s', (level, factor) => {
    const result = calculateRecommendedMaxWeightGrams(10000, '1996-01-01', 'other', level)
    expect(result).toBeCloseTo(10 * 0.12 * 1.1 * 1 * factor * 1000, 0)
  })
})

describe('getPersonRecommendedMaxWeightGrams', () => {
  it('derives all inputs from the person object', () => {
    const person = personFixture({
      body_weight_grams: 70000,
      birthdate: null,
      gender: 'female',
      conditioning_level: 'average',
    })
    expect(getPersonRecommendedMaxWeightGrams(person)).toBeCloseTo(8778, 0)
  })
})
