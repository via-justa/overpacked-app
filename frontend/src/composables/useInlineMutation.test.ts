import { useInlineMutation } from './useInlineMutation'

const add = vi.hoisted(() => vi.fn())
vi.mock('primevue/usetoast', () => ({ useToast: () => ({ add }) }))

beforeEach(() => add.mockClear())

describe('useInlineMutation', () => {
  it('runs the mutation, returns its result, and shows a success toast', async () => {
    const { executeInlineMutation } = useInlineMutation()
    const result = await executeInlineMutation(async () => 'done', {
      successSummary: 'Updated',
      successDetail: (r) => `Result: ${r}`,
      errorSummary: 'Failed',
    })

    expect(result).toBe('done')
    expect(add).toHaveBeenCalledWith(
      expect.objectContaining({ severity: 'success', summary: 'Updated', detail: 'Result: done' }),
    )
  })

  it('accepts a static success detail string', async () => {
    const { executeInlineMutation } = useInlineMutation()
    await executeInlineMutation(async () => undefined, {
      successSummary: 'Saved',
      successDetail: 'All good.',
      errorSummary: 'Failed',
    })
    expect(add).toHaveBeenCalledWith(expect.objectContaining({ detail: 'All good.' }))
  })

  it('uses a generic message when neither error nor errorDetail is informative', async () => {
    const { executeInlineMutation } = useInlineMutation()
    await executeInlineMutation(
      async () => {
        throw 'x' // eslint-disable-line no-throw-literal
      },
      { successSummary: 'ok', successDetail: 'ok', errorSummary: 'Failed' },
    )
    expect(add).toHaveBeenCalledWith(expect.objectContaining({ detail: 'An error occurred.' }))
  })

  it('catches errors, returns undefined, and shows the error message', async () => {
    const { executeInlineMutation } = useInlineMutation()
    const result = await executeInlineMutation(
      async () => {
        throw new Error('nope')
      },
      { successSummary: 'ok', successDetail: 'ok', errorSummary: 'Failed' },
    )

    expect(result).toBeUndefined()
    expect(add).toHaveBeenCalledWith(expect.objectContaining({ severity: 'error', detail: 'nope' }))
  })

  it('falls back to the configured error detail for non-Error throws', async () => {
    const { executeInlineMutation } = useInlineMutation()
    await executeInlineMutation(
      async () => {
        throw 'oops' // eslint-disable-line no-throw-literal
      },
      { successSummary: 'ok', successDetail: 'ok', errorSummary: 'Failed', errorDetail: 'Custom fallback' },
    )

    expect(add).toHaveBeenCalledWith(expect.objectContaining({ detail: 'Custom fallback' }))
  })
})
