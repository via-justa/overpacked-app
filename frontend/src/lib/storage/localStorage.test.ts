import {
  getStoredValue,
  isDetailMode,
  isViewMode,
  setStoredValue,
} from './localStorage'

beforeEach(() => globalThis.window.localStorage.clear())

describe('getStoredValue / setStoredValue', () => {
  it('returns the default when nothing is stored', () => {
    expect(getStoredValue('view', isViewMode, 'cards')).toBe('cards')
  })

  it('round-trips a valid stored value', () => {
    setStoredValue('view', 'table')
    expect(getStoredValue('view', isViewMode, 'cards')).toBe('table')
  })

  it('falls back to the default when the stored value fails validation', () => {
    setStoredValue('view', 'garbage')
    expect(getStoredValue('view', isViewMode, 'cards')).toBe('cards')
  })
})

describe('type guards', () => {
  it('isViewMode', () => {
    expect(isViewMode('cards')).toBe(true)
    expect(isViewMode('table')).toBe(true)
    expect(isViewMode('grid')).toBe(false)
  })

  it('isDetailMode', () => {
    expect(isDetailMode('simple')).toBe(true)
    expect(isDetailMode('expanded')).toBe(true)
    expect(isDetailMode('full')).toBe(false)
  })
})
