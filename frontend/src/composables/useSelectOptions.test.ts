import { Comment, Fragment, h, type Slots } from 'vue'
import { useFieldMessageClass, useSelectOptions } from './useSelectOptions'

describe('useSelectOptions', () => {
  it('parses <option> nodes into structured options', () => {
    const slots = {
      default: () => [
        h('option', { value: 'a' }, 'Apple'),
        h('option', { value: 'b', disabled: true }, 'Banana'),
      ],
    }
    const { parsedOptions } = useSelectOptions(slots)
    expect(parsedOptions.value).toEqual([
      { value: 'a', label: 'Apple', disabled: false },
      { value: 'b', label: 'Banana', disabled: true },
    ])
  })

  it('flattens fragments, skips comments/text, and recurses nested labels', () => {
    const slots = {
      default: () => [
        'loose text', // not a vnode → skipped
        h(Comment, null, ''), // comment → skipped
        h(Fragment, null, [h('option', { value: 1 }, [h('span', null, 'One')])]),
      ],
    } as unknown as Slots
    const { parsedOptions } = useSelectOptions(slots)
    expect(parsedOptions.value).toEqual([{ value: 1, label: 'One', disabled: false }])
  })

  it('falls back to the value when an option has no text', () => {
    const slots = { default: () => [h('option', { value: 'x' })] }
    const { parsedOptions } = useSelectOptions(slots)
    expect(parsedOptions.value[0].label).toBe('x')
  })

  it('handles a missing default slot', () => {
    const { parsedOptions } = useSelectOptions({})
    expect(parsedOptions.value).toEqual([])
  })
})

describe('useFieldMessageClass', () => {
  it('uses danger styling for an invalid message', () => {
    expect(useFieldMessageClass({ message: 'Bad', invalid: true }).value).toContain('text-danger-500')
  })

  it('uses muted styling for a non-error message', () => {
    expect(useFieldMessageClass({ message: 'Hint', invalid: false }).value).toContain('text-copy-muted')
  })

  it('reserves space (invisible) or hides when there is no message', () => {
    expect(useFieldMessageClass({ reserveMessageSpace: true }).value).toContain('invisible')
    expect(useFieldMessageClass({}).value).toContain('hidden')
  })

  it('appends a custom message class', () => {
    expect(useFieldMessageClass({ message: 'x', invalid: true, messageClass: 'mt-2' }).value).toContain('mt-2')
  })
})
