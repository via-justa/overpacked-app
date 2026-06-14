import { safeRedirectPath } from './redirect'

describe('safeRedirectPath', () => {
  it('accepts root-relative internal paths', () => {
    expect(safeRedirectPath('/trips')).toBe('/trips')
    expect(safeRedirectPath('/trips?tab=packed#x')).toBe('/trips?tab=packed#x')
  })

  it('rejects non-strings', () => {
    expect(safeRedirectPath(undefined)).toBeNull()
    expect(safeRedirectPath(['/trips'])).toBeNull()
  })

  it('rejects off-origin and protocol-relative paths (open-redirect guard)', () => {
    expect(safeRedirectPath('https://evil.com')).toBeNull()
    expect(safeRedirectPath('//evil.com')).toBeNull()
    expect(safeRedirectPath(String.raw`/\evil.com`)).toBeNull()
    expect(safeRedirectPath('relative')).toBeNull()
  })
})
