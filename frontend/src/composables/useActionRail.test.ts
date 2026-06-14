import { useActionRail } from './useActionRail'

const STORAGE_KEY = 'overpacked-app.ui.actionRailPinned'

describe('useActionRail', () => {
  it('toggles the shared pinned state and persists it', () => {
    const { pinned, togglePinned } = useActionRail()
    const start = pinned.value

    togglePinned()
    expect(pinned.value).toBe(!start)
    expect(globalThis.window.localStorage.getItem(STORAGE_KEY)).toBe(String(!start))

    togglePinned()
    expect(pinned.value).toBe(start)
    expect(globalThis.window.localStorage.getItem(STORAGE_KEY)).toBe(String(start))
  })

  it('shares one source of truth across call sites', () => {
    const a = useActionRail()
    const b = useActionRail()
    const before = a.pinned.value
    a.togglePinned()
    expect(b.pinned.value).toBe(!before)
    a.togglePinned() // restore
  })
})
