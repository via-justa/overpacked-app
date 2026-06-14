import { nextTick } from 'vue'
import { useRowActionsMenu } from './useRowActionsMenu'
import { withSetup } from '../test/withSetup'

const rect = (over: Partial<DOMRect>): DOMRect =>
  ({ top: 80, bottom: 100, left: 480, right: 500, width: 20, height: 20, x: 480, y: 80, toJSON() {}, ...over })

const makeTrigger = (r: DOMRect): HTMLButtonElement => {
  const btn = document.createElement('button')
  btn.getBoundingClientRect = () => r
  document.body.appendChild(btn)
  return btn
}

const clickEvent = (target: HTMLElement) => ({ currentTarget: target }) as unknown as MouseEvent

afterEach(() => {
  document.body.innerHTML = ''
})

describe('useRowActionsMenu', () => {
  it('opens below the trigger, right-aligned, when there is room', () => {
    Object.defineProperty(globalThis, 'innerHeight', { configurable: true, value: 1000 })
    const { result, unmount } = withSetup(() => useRowActionsMenu({ dataElement: 'row' }))

    result.toggleActions('1', clickEvent(makeTrigger(rect({}))))

    expect(result.openActionsForId.value).toBe('1')
    // left = max(8, right(500) - menuWidth(176)) = 324; top = bottom(100) + gap(6) = 106
    expect(result.menuPosition.value).toEqual({ top: 106, left: 324 })
    unmount()
  })

  it('flips upward when too close to the viewport bottom', () => {
    Object.defineProperty(globalThis, 'innerHeight', { configurable: true, value: 120 })
    const { result, unmount } = withSetup(() => useRowActionsMenu({ dataElement: 'row' }))

    result.toggleActions('1', clickEvent(makeTrigger(rect({ top: 80, bottom: 100 }))))
    // openUpward → top = max(8, top(80) - gap(6) - menuHeight(188)) = 8
    expect(result.menuPosition.value.top).toBe(8)
    unmount()
  })

  it('toggles closed when the same row is clicked again', () => {
    const { result, unmount } = withSetup(() => useRowActionsMenu({ dataElement: 'row' }))
    const trigger = makeTrigger(rect({}))

    result.toggleActions('1', clickEvent(trigger))
    expect(result.openActionsForId.value).toBe('1')
    result.toggleActions('1', clickEvent(trigger))
    expect(result.openActionsForId.value).toBeNull()
    unmount()
  })

  it('closes on an outside click', () => {
    const { result, unmount } = withSetup(() => useRowActionsMenu({ dataElement: 'row' }))
    result.toggleActions('1', clickEvent(makeTrigger(rect({}))))

    document.body.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    expect(result.openActionsForId.value).toBeNull()
    unmount()
  })

  it('navigates menu items with the keyboard and closes on Escape', async () => {
    const menu = document.createElement('div')
    menu.dataset.element = 'row-menu'
    const b1 = document.createElement('button')
    const b2 = document.createElement('button')
    menu.append(b1, b2)
    document.body.appendChild(menu)

    const { result, unmount } = withSetup(() => useRowActionsMenu({ dataElement: 'row' }))
    result.toggleActions('1', clickEvent(makeTrigger(rect({}))))
    await nextTick()
    expect(document.activeElement).toBe(b1) // focus moved into the menu

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown' }))
    expect(document.activeElement).toBe(b2)

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    expect(result.openActionsForId.value).toBeNull()
    unmount()
  })

  it('supports ArrowUp/Home/End/Tab navigation and wraps around', async () => {
    const menu = document.createElement('div')
    menu.dataset.element = 'row-menu'
    const buttons = [0, 1, 2].map(() => document.createElement('button'))
    menu.append(...buttons)
    document.body.appendChild(menu)

    const { result, unmount } = withSetup(() => useRowActionsMenu({ dataElement: 'row' }))
    result.toggleActions('1', clickEvent(makeTrigger(rect({}))))
    await nextTick()
    expect(document.activeElement).toBe(buttons[0])

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowUp' })) // wraps to last
    expect(document.activeElement).toBe(buttons[2])

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Home' }))
    expect(document.activeElement).toBe(buttons[0])

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'End' }))
    expect(document.activeElement).toBe(buttons[2])

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab' })) // wraps forward to first
    expect(document.activeElement).toBe(buttons[0])

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true })) // back to last
    expect(document.activeElement).toBe(buttons[2])
    unmount()
  })

  it('ignores key presses while the menu is closed', () => {
    const { result, unmount } = withSetup(() => useRowActionsMenu({ dataElement: 'row' }))
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown' }))
    expect(result.openActionsForId.value).toBeNull()
    unmount()
  })

  it('opens without positioning when the event has no element target', () => {
    const { result, unmount } = withSetup(() => useRowActionsMenu({ dataElement: 'row' }))
    result.toggleActions('1', { currentTarget: null } as unknown as MouseEvent)
    expect(result.openActionsForId.value).toBe('1')
    unmount()
  })
})
