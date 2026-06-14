import { slugifyCategoryId } from './itemUtils'

describe('slugifyCategoryId', () => {
  it('lowercases and joins words with underscores', () => {
    expect(slugifyCategoryId('Hello World')).toBe('hello_world')
    expect(slugifyCategoryId('My Category!')).toBe('my_category')
  })

  it('collapses runs of separators and trims edges', () => {
    expect(slugifyCategoryId('  leading')).toBe('leading')
    expect(slugifyCategoryId('a -- b')).toBe('a_b')
    expect(slugifyCategoryId('trailing-')).toBe('trailing')
  })

  it('keeps alphanumerics and drops non-alphanumerics', () => {
    expect(slugifyCategoryId('café')).toBe('caf')
    expect(slugifyCategoryId('Item 2 Go')).toBe('item_2_go')
  })

  it('falls back to a timestamped id when nothing usable remains', () => {
    expect(slugifyCategoryId('!!!')).toMatch(/^category_\d+$/)
    expect(slugifyCategoryId('   ')).toMatch(/^category_\d+$/)
  })
})
