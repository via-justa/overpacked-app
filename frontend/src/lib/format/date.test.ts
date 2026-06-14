import { formatDate, formatDateMedium } from './date'

// Tests run with TZ=UTC (see vitest.config.ts) so these are deterministic.
describe('formatDate', () => {
  it('formats an ISO date as DD-MM-YYYY with zero padding', () => {
    expect(formatDate('2026-06-11T12:00:00.000Z')).toBe('11-06-2026')
    expect(formatDate('2026-01-05T12:00:00.000Z')).toBe('05-01-2026')
  })

  it('returns the input unchanged when unparseable', () => {
    expect(formatDate('not-a-date')).toBe('not-a-date')
  })
})

describe('formatDateMedium', () => {
  it('formats as a localized medium date', () => {
    expect(formatDateMedium('2026-06-11T12:00:00.000Z')).toBe('Jun 11, 2026')
  })
})
