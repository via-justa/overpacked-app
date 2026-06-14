import { normalizeTitleWords } from './normalize'

describe('normalizeTitleWords', () => {
  it('returns empty string for blank input', () => {
    expect(normalizeTitleWords('')).toBe('')
    expect(normalizeTitleWords('   ')).toBe('')
  })

  it('title-cases each word and collapses whitespace', () => {
    expect(normalizeTitleWords('hello   WORLD')).toBe('Hello World')
    expect(normalizeTitleWords('  multi  space ')).toBe('Multi Space')
    expect(normalizeTitleWords('aLpInE tReK')).toBe('Alpine Trek')
  })
})
