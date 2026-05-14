import { ref } from 'vue'

/**
 * Reusable composable for delete confirmation dialogs.
 * Manages confirmation state and provides helpers for single/bulk delete flows.
 */
export type DeleteConfirmationState =
  | { kind: 'single'; id: string; name: string }
  | { kind: 'bulk'; ids: string[]; count: number }
  | null

export function useDeleteConfirmation() {
  const confirmState = ref<DeleteConfirmationState>(null)

  const requestSingleDelete = (id: string, name: string) => {
    confirmState.value = { kind: 'single', id, name }
  }

  const requestBulkDelete = (ids: string[]) => {
    confirmState.value = { kind: 'bulk', ids, count: ids.length }
  }

  const closeConfirm = () => {
    confirmState.value = null
  }

  const getConfirmMessage = (): string => {
    if (!confirmState.value) return ''

    if (confirmState.value.kind === 'single') {
      return `Delete ${confirmState.value.name}?`
    }

    return `Delete ${confirmState.value.count} selected item(s)?`
  }

  return {
    confirmState,
    requestSingleDelete,
    requestBulkDelete,
    closeConfirm,
    getConfirmMessage,
  }
}
