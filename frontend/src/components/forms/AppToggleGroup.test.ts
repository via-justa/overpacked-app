import { fireEvent, render } from '@testing-library/vue'
import AppToggleGroup from './AppToggleGroup.vue'

const options = [
  { label: 'Cards', value: 'cards' },
  { label: 'Table', value: 'table' },
]

describe('AppToggleGroup', () => {
  it('reflects the current modelValue as the checked radio', () => {
    const { getByLabelText } = render(AppToggleGroup, {
      props: { name: 'view', modelValue: 'cards', options },
    })
    expect((getByLabelText('Cards') as HTMLInputElement).checked).toBe(true)
    expect((getByLabelText('Table') as HTMLInputElement).checked).toBe(false)
  })

  it('emits update:modelValue when another option is chosen', async () => {
    const { getByLabelText, emitted } = render(AppToggleGroup, {
      props: { name: 'view', modelValue: 'cards', options },
    })
    await fireEvent.click(getByLabelText('Table'))
    expect(emitted()['update:modelValue'][0]).toEqual(['table'])
  })
})
