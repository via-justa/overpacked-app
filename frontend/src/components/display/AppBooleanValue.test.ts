import { render } from '@testing-library/vue'
import AppBooleanValue from './AppBooleanValue.vue'

// Presentational, provider-free component: a plain render is enough.
describe('AppBooleanValue', () => {
  it('renders "Yes" with an accessible label when true', () => {
    const { getByText, getByLabelText } = render(AppBooleanValue, {
      props: { value: true, label: 'Active' },
    })
    expect(getByLabelText('Active: Yes')).toBeInTheDocument()
    expect(getByText('Yes')).toBeInTheDocument()
  })

  it('renders "No" when false', () => {
    const { getByText, getByLabelText } = render(AppBooleanValue, {
      props: { value: false, label: 'Active' },
    })
    expect(getByLabelText('Active: No')).toBeInTheDocument()
    expect(getByText('No')).toBeInTheDocument()
  })

  it('falls back to a "Not set" placeholder when the value is missing', () => {
    const { getByText } = render(AppBooleanValue, { props: { value: null } })
    expect(getByText('Not set')).toBeInTheDocument()
  })
})
