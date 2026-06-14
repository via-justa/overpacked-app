import { useDeleteConfirmation } from './useDeleteConfirmation'

describe('useDeleteConfirmation', () => {
  it('tracks a single-delete request and message', () => {
    const { confirmState, requestSingleDelete, getConfirmMessage } = useDeleteConfirmation()
    requestSingleDelete('id-1', 'Tent')
    expect(confirmState.value).toEqual({ kind: 'single', id: 'id-1', name: 'Tent' })
    expect(getConfirmMessage()).toBe('Delete Tent?')
  })

  it('tracks a bulk-delete request and message', () => {
    const { confirmState, requestBulkDelete, getConfirmMessage } = useDeleteConfirmation()
    requestBulkDelete(['a', 'b', 'c'])
    expect(confirmState.value).toEqual({ kind: 'bulk', ids: ['a', 'b', 'c'], count: 3 })
    expect(getConfirmMessage()).toBe('Delete 3 selected item(s)?')
  })

  it('closes the confirmation and yields an empty message', () => {
    const { confirmState, requestSingleDelete, closeConfirm, getConfirmMessage } = useDeleteConfirmation()
    requestSingleDelete('id-1', 'Tent')
    closeConfirm()
    expect(confirmState.value).toBeNull()
    expect(getConfirmMessage()).toBe('')
  })
})
