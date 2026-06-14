import { useMutationWithToast } from './useMutationWithToast'
import { withSetup } from '../test/withSetup'
import { queryClient } from '../lib/query/client'

const add = vi.hoisted(() => vi.fn())
vi.mock('primevue/usetoast', () => ({ useToast: () => ({ add }) }))

const messages = {
  successMessage: { summary: 'Saved', detail: 'It worked.' },
  errorMessage: { summary: 'Failed', detail: 'It did not work.' },
}

beforeEach(() => add.mockClear())

describe('useMutationWithToast', () => {
  it('shows a success toast and invalidates the configured queries', async () => {
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries').mockResolvedValue()

    const { result, unmount } = withSetup(() =>
      useMutationWithToast<string, Error, string>({
        mutationFn: async (value: string) => value.toUpperCase(),
        ...messages,
        invalidateQueries: [['items']],
      }),
    )

    await result.mutateAsync('ok')

    await vi.waitFor(() =>
      expect(add).toHaveBeenCalledWith(expect.objectContaining({ severity: 'success', summary: 'Saved' })),
    )
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['items'] })

    invalidate.mockRestore()
    unmount()
  })

  it('writes the cache directly via setQueryData and runs the custom onSuccess', async () => {
    const setData = vi.spyOn(queryClient, 'setQueryData').mockReturnValue(undefined)
    const onSuccess = vi.fn()

    const { result, unmount } = withSetup(() =>
      useMutationWithToast<string, Error, string>({
        mutationFn: async (value: string) => value,
        ...messages,
        onSuccess,
        setQueryData: { queryKey: ['item', '1'], updater: (data) => ({ name: data }) },
      }),
    )

    await result.mutateAsync('alpha')
    await vi.waitFor(() => expect(onSuccess).toHaveBeenCalled())
    expect(setData).toHaveBeenCalledWith(['item', '1'], { name: 'alpha' })

    setData.mockRestore()
    unmount()
  })

  it('invalidates all queries when requested', async () => {
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries').mockResolvedValue()

    const { result, unmount } = withSetup(() =>
      useMutationWithToast<string, Error, string>({
        mutationFn: async (value: string) => value,
        ...messages,
        invalidateAllQueries: true,
      }),
    )

    await result.mutateAsync('x')
    await vi.waitFor(() => expect(invalidate).toHaveBeenCalledWith())

    invalidate.mockRestore()
    unmount()
  })

  it('shows an error toast carrying the error message', async () => {
    const { result, unmount } = withSetup(() =>
      useMutationWithToast<string, Error, string>({
        mutationFn: async () => {
          throw new Error('boom')
        },
        ...messages,
      }),
    )

    await expect(result.mutateAsync('x')).rejects.toThrow('boom')

    await vi.waitFor(() =>
      expect(add).toHaveBeenCalledWith(expect.objectContaining({ severity: 'error', detail: 'boom' })),
    )
    unmount()
  })

  it('runs the custom onError and falls back to the configured detail for non-Error throws', async () => {
    const onError = vi.fn()

    const { result, unmount } = withSetup(() =>
      useMutationWithToast<string, string, string>({
        mutationFn: async () => {
          throw 'plain string' // eslint-disable-line no-throw-literal
        },
        ...messages,
        onError,
      }),
    )

    await expect(result.mutateAsync('x')).rejects.toBe('plain string')
    await vi.waitFor(() => expect(onError).toHaveBeenCalled())
    expect(add).toHaveBeenCalledWith(
      expect.objectContaining({ severity: 'error', detail: 'It did not work.' }),
    )
    unmount()
  })
})
