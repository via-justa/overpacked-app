import {
  formatCarryStatus,
  formatNumber,
  formatText,
  formatType,
  formatValue,
} from './display'

describe('formatNumber', () => {
  it('falls back to "Not set" for non-numbers', () => {
    expect(formatNumber(null)).toBe('Not set')
    expect(formatNumber(undefined)).toBe('Not set')
  })

  it('stringifies, optionally via a custom formatter', () => {
    expect(formatNumber(42)).toBe('42')
    expect(formatNumber(1.5, (n) => n.toFixed(1))).toBe('1.5')
  })
})

describe('formatValue', () => {
  it('falls back to "Not set" for non-numbers', () => {
    expect(formatValue(null, 'usd')).toBe('Not set')
  })

  it('appends the currency symbol ($ for usd, € otherwise)', () => {
    expect(formatValue(10, 'usd')).toBe('10 $')
    expect(formatValue(10, 'eur')).toBe('10 €')
    expect(formatValue(10)).toBe('10 €')
  })

  it('applies the optional formatter', () => {
    expect(formatValue(10, 'usd', (n) => n.toFixed(2))).toBe('10.00 $')
  })
})

describe('formatCarryStatus', () => {
  it('maps known statuses and falls back', () => {
    expect(formatCarryStatus('packed')).toBe('Packed')
    expect(formatCarryStatus('worn')).toBe('Worn')
    expect(formatCarryStatus(null)).toBe('Not set')
    expect(formatCarryStatus('custom')).toBe('custom')
  })
})

describe('formatType', () => {
  it('title-cases snake_case', () => {
    expect(formatType('day_hike')).toBe('Day Hike')
    expect(formatType('shelter')).toBe('Shelter')
  })
})

describe('formatText', () => {
  it('falls back to "Not set" for empty/blank', () => {
    expect(formatText('')).toBe('Not set')
    expect(formatText('   ')).toBe('Not set')
    expect(formatText(null)).toBe('Not set')
  })

  it('returns the text otherwise', () => {
    expect(formatText('hello')).toBe('hello')
  })
})
