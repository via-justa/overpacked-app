import { validatePlannerDetails } from './schema'

describe('validatePlannerDetails', () => {
  it('accepts empty optional fields', () => {
    expect(validatePlannerDetails({ durationDays: '', distanceKm: '' })).toEqual({})
  })

  it('accepts valid values', () => {
    expect(validatePlannerDetails({ durationDays: '3', distanceKm: '12.5' })).toEqual({})
  })

  it('rejects non-positive / non-integer durations with the user-facing message', () => {
    expect(validatePlannerDetails({ durationDays: '0', distanceKm: '' }).durationDays).toMatch(
      /greater than 0/,
    )
    expect(validatePlannerDetails({ durationDays: '1.5', distanceKm: '' }).durationDays).toMatch(
      /greater than 0/,
    )
    expect(validatePlannerDetails({ durationDays: 'abc', distanceKm: '' }).durationDays).toBeTruthy()
  })

  it('rejects negative / non-numeric distances', () => {
    expect(validatePlannerDetails({ durationDays: '', distanceKm: '-1' }).distanceKm).toMatch(
      /0 or more/,
    )
    expect(validatePlannerDetails({ durationDays: '', distanceKm: 'abc' }).distanceKm).toBeTruthy()
  })

  it('reports both fields at once', () => {
    const errors = validatePlannerDetails({ durationDays: '0', distanceKm: '-1' })
    expect(errors.durationDays).toBeTruthy()
    expect(errors.distanceKm).toBeTruthy()
  })
})
