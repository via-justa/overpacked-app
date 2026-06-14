import { fireEvent, screen } from '@testing-library/vue'
import { renderWithProviders } from '../../test/renderWithProviders'
import AppConfirmDialog from './AppConfirmDialog.vue'

// The dialog teleports to <body>, so query the document via `screen`.
describe('AppConfirmDialog', () => {
  it('emits confirm when the confirm button is clicked', async () => {
    const { emitted } = renderWithProviders(AppConfirmDialog, {
      props: { open: true, message: 'Delete this trip?', confirmLabel: 'Confirm' },
    })

    await fireEvent.click(await screen.findByRole('button', { name: 'Confirm' }))
    expect(emitted().confirm).toHaveLength(1)
  })

  it('emits cancel and closes when the cancel button is clicked', async () => {
    const { emitted } = renderWithProviders(AppConfirmDialog, {
      props: { open: true, message: 'Delete this trip?', cancelLabel: 'Cancel' },
    })

    await fireEvent.click(await screen.findByRole('button', { name: 'Cancel' }))
    expect(emitted().cancel).toHaveLength(1)
    expect(emitted()['update:open']).toContainEqual([false])
  })

  it('renders the message and title', async () => {
    renderWithProviders(AppConfirmDialog, {
      props: { open: true, title: 'Delete trip', message: 'This cannot be undone.' },
    })
    expect(await screen.findByText('This cannot be undone.')).toBeInTheDocument()
    expect(screen.getByText('Delete trip')).toBeInTheDocument()
  })
})
