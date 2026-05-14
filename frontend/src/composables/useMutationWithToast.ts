import { useMutation, type UseMutationReturnType } from '@tanstack/vue-query'
import { useToast } from 'primevue/usetoast'
import { queryClient } from '../lib/query/client'

export interface ToastMessage {
  summary: string
  detail: string
  life?: number
}

export interface MutationWithToastOptions<TData, TError, TVariables, TContext> {
  /**
   * The mutation function to execute
   */
  mutationFn: (variables: TVariables) => Promise<TData>
  
  /**
   * Success toast message configuration
   */
  successMessage: ToastMessage
  
  /**
   * Error toast message configuration (detail can be overridden by error.message)
   */
  errorMessage: ToastMessage
  
  /**
   * Custom success handler called before toast and query invalidation
   */
  onSuccess?: (data: TData, variables: TVariables, context: TContext | undefined) => void | Promise<void>
  
  /**
   * Custom error handler called before error toast
   */
  onError?: (error: TError, variables: TVariables, context: TContext | undefined) => void | Promise<void>
  
  /**
   * Query keys to invalidate on success. If not provided, no queries are invalidated.
   */
  invalidateQueries?: string[][]
  
  /**
   * If true, invalidates all queries on success
   */
  invalidateAllQueries?: boolean
  
  /**
   * If provided, sets query data directly instead of invalidating
   */
  setQueryData?: {
    queryKey: string[]
    updater: (data: TData) => unknown
  }
}

/**
 * A wrapper around useMutation that automatically handles success/error toasts
 * and query invalidation patterns.
 * 
 * @example
 * ```ts
 * const createMutation = useMutationWithToast({
 *   mutationFn: createItem,
 *   successMessage: { summary: 'Item created', detail: 'New item added.' },
 *   errorMessage: { summary: 'Create failed', detail: 'Unable to create item.' },
 *   invalidateQueries: [['items']],
 *   onSuccess: () => {
 *     closeDialog()
 *     resetForm()
 *   }
 * })
 * ```
 */
export function useMutationWithToast<TData = unknown, TError = Error, TVariables = void, TContext = unknown>(
  options: MutationWithToastOptions<TData, TError, TVariables, TContext>
): UseMutationReturnType<TData, TError, TVariables, TContext> {
  const toast = useToast()
  
  const {
    successMessage,
    errorMessage,
    onSuccess: customOnSuccess,
    onError: customOnError,
    invalidateQueries,
    invalidateAllQueries,
    setQueryData,
    ...mutationOptions
  } = options
  
  return useMutation({
    ...mutationOptions,
    onSuccess: async (data, variables, context) => {
      // Call custom success handler first
      if (customOnSuccess) {
        await customOnSuccess(data, variables, context)
      }
      
      // Handle query cache updates
      if (setQueryData) {
        queryClient.setQueryData(setQueryData.queryKey, setQueryData.updater(data))
      } else if (invalidateAllQueries) {
        await queryClient.invalidateQueries()
      } else if (invalidateQueries && invalidateQueries.length > 0) {
        await Promise.all(
          invalidateQueries.map((queryKey) => 
            queryClient.invalidateQueries({ queryKey })
          )
        )
      }
      
      // Show success toast
      toast.add({
        severity: 'success',
        summary: successMessage.summary,
        detail: successMessage.detail,
        life: successMessage.life ?? 3000,
      })
    },
    onError: async (error, variables, context) => {
      // Call custom error handler first
      if (customOnError) {
        await customOnError(error, variables, context)
      }
      
      // Show error toast with error message if available
      const detail = error instanceof Error && error.message 
        ? error.message 
        : errorMessage.detail
      
      toast.add({
        severity: 'error',
        summary: errorMessage.summary,
        detail,
        life: errorMessage.life ?? 3500,
      })
    },
  })
}
