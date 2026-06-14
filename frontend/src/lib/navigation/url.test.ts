import { safeHttpUrl } from './url'

describe('safeHttpUrl', () => {
  it('returns http(s) URLs unchanged', () => {
    expect(safeHttpUrl('https://komoot.com/tour/1')).toBe('https://komoot.com/tour/1')
    expect(safeHttpUrl('  http://example.com  ')).toBe('http://example.com')
  })

  it('rejects empty/blank/nullish values', () => {
    expect(safeHttpUrl(null)).toBeUndefined()
    expect(safeHttpUrl(undefined)).toBeUndefined()
    expect(safeHttpUrl('   ')).toBeUndefined()
  })

  it('rejects dangerous or non-http schemes (XSS guard)', () => {
    expect(safeHttpUrl('javascript:alert(1)')).toBeUndefined()
    expect(safeHttpUrl('data:text/html;base64,x')).toBeUndefined()
    expect(safeHttpUrl('not a url')).toBeUndefined()
  })
})
