import { getErrorMessage, readString } from './errors'

describe('readString', () => {
  it('returns non-empty strings, else null', () => {
    expect(readString('hello')).toBe('hello')
    expect(readString('   ')).toBeNull()
    expect(readString(42)).toBeNull()
    expect(readString(null)).toBeNull()
  })
})

describe('getErrorMessage', () => {
  it('falls back when the error is not an object', () => {
    expect(getErrorMessage(null, 'fallback')).toBe('fallback')
    expect(getErrorMessage('boom', 'fallback')).toBe('fallback')
  })

  it('reads the first populated direct field (error/message/detail/details)', () => {
    expect(getErrorMessage({ error: 'e' }, 'fb')).toBe('e')
    expect(getErrorMessage({ message: 'm' }, 'fb')).toBe('m')
    expect(getErrorMessage({ detail: 'd' }, 'fb')).toBe('d')
    expect(getErrorMessage({ details: 'ds' }, 'fb')).toBe('ds')
  })

  it('reads a nested error object', () => {
    expect(getErrorMessage({ error: { message: 'nested' } }, 'fb')).toBe('nested')
    expect(getErrorMessage({ error: { detail: 'nd' } }, 'fb')).toBe('nd')
  })

  it('falls back when nothing usable is present', () => {
    expect(getErrorMessage({}, 'fb')).toBe('fb')
    expect(getErrorMessage({ error: {} }, 'fb')).toBe('fb')
  })
})
