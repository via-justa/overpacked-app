import { ref } from 'vue'
import type { UseQueryReturnType } from '@tanstack/vue-query'
import { renderWithProviders } from '../../test/renderWithProviders'
import AppQueryState from './AppQueryState.vue'

// Minimal fake of the bits of a vue-query result that AppQueryState reads.
const fakeQuery = (o: { pending?: boolean; error?: Error | null; data?: unknown }) =>
  ({
    isPending: ref(o.pending ?? false),
    isError: ref(Boolean(o.error)),
    error: ref(o.error ?? null),
    data: ref(o.data ?? null),
  }) as unknown as UseQueryReturnType<unknown, Error>

describe('AppQueryState', () => {
  it('shows the loading state while pending', () => {
    const { getByText } = renderWithProviders(AppQueryState, {
      props: { query: fakeQuery({ pending: true }), loadingMessage: 'Loading gear…' },
    })
    expect(getByText('Loading gear…')).toBeInTheDocument()
  })

  it('shows the error message from the query error', () => {
    const { getByText } = renderWithProviders(AppQueryState, {
      props: { query: fakeQuery({ error: new Error('boom') }) },
    })
    expect(getByText('boom')).toBeInTheDocument()
  })

  it('shows the empty state for an empty array result', () => {
    const { getByText } = renderWithProviders(AppQueryState, {
      props: { query: fakeQuery({ data: [] }), emptyMessage: 'Nothing yet' },
    })
    expect(getByText('Nothing yet')).toBeInTheDocument()
  })

  it('renders the default slot when data is present', () => {
    const { getByText } = renderWithProviders(AppQueryState, {
      props: { query: fakeQuery({ data: [{ id: 1 }] }) },
      slots: { default: '<div>list content</div>' },
    })
    expect(getByText('list content')).toBeInTheDocument()
  })
})
