import { useToast } from 'primevue/usetoast'

export interface InlineMutationMessages {
  successSummary: string
  successDetail: string | ((result?: unknown) => string)
  errorSummary: string
  errorDetail?: string
  successLife?: number
  errorLife?: number
}

/**
 * A helper for inline mutations (e.g., row actions) that handles try-catch + toast pattern.
 * Unlike useMutationWithToast, this is for imperative mutations called directly from event handlers.
 * 
 * @example
 * ```ts
 * const { executeInlineMutation } = useInlineMutation()
 * 
 * const onRowToggle = async (item: Item) => {
 *   await executeInlineMutation(
 *     async () => {
 *       await updateItem(item.id, { is_active: !item.is_active })
 *       await queryClient.invalidateQueries({ queryKey: ['items'] })
 *     },
 *     {
 *       successSummary: 'Item updated',
 *       successDetail: `Item is now ${item.is_active ? 'inactive' : 'active'}.`,
 *       errorSummary: 'Update failed',
 *       errorDetail: 'Unable to update item.',
 *     }
 *   )
 * }
 * ```
 */
export function useInlineMutation() {
  const toast = useToast()

  const executeInlineMutation = async <T = void>(
    mutationFn: () => Promise<T>,
    messages: InlineMutationMessages
  ): Promise<T | undefined> => {
    try {
      const result = await mutationFn()
      
      const successDetail = typeof messages.successDetail === 'function'
        ? messages.successDetail(result)
        : messages.successDetail

      toast.add({
        severity: 'success',
        summary: messages.successSummary,
        detail: successDetail,
        life: messages.successLife ?? 2500,
      })

      return result
    } catch (error) {
      const errorDetail = error instanceof Error && error.message
        ? error.message
        : (messages.errorDetail ?? 'An error occurred.')

      toast.add({
        severity: 'error',
        summary: messages.errorSummary,
        detail: errorDetail,
        life: messages.errorLife ?? 3500,
      })

      return undefined
    }
  }

  return {
    executeInlineMutation,
  }
}
