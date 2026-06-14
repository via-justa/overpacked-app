import { nextTick, onBeforeUnmount, onMounted, ref, type Ref } from 'vue'

export type MenuPosition = { top: number; left: number }

export interface UseRowActionsMenuOptions {
  menuWidth?: number
  menuHeight?: number
  gap?: number
  /** `data-element` of the trigger wrapper (used for outside-click dismissal). */
  dataElement: string
  /** `data-element` of the rendered menu; defaults to `${dataElement}-menu`. */
  menuDataElement?: string
}

export interface UseRowActionsMenuReturn {
  openActionsForId: Ref<string | null>
  menuPosition: Ref<MenuPosition>
  closeActions: () => void
  toggleActions: (id: string, event: MouseEvent) => void
}

/**
 * Composable for table/list row action menus with smart positioning.
 * Handles menu open/close state, viewport-aware positioning, outside-click
 * dismissal, and keyboard operation (arrow/Home/End/Escape navigation, focus
 * moves into the menu on open and is restored to the trigger on close).
 *
 * The menu is rendered by the consumer; this composable locates it in the DOM
 * via `menuDataElement` and drives focus across its `<button>` children.
 */
export function useRowActionsMenu(options: UseRowActionsMenuOptions): UseRowActionsMenuReturn {
  const { menuWidth = 176, menuHeight = 188, gap = 6, dataElement } = options
  const menuDataElement = options.menuDataElement ?? `${dataElement}-menu`

  const openActionsForId = ref<string | null>(null)
  const menuPosition = ref<MenuPosition>({ top: 0, left: 0 })
  // Remembered so focus can be returned to the trigger when the menu closes.
  let triggerEl: HTMLElement | null = null

  const menuButtons = (): HTMLButtonElement[] => {
    const menu = globalThis.document?.querySelector(`[data-element="${menuDataElement}"]`)
    if (!menu) {
      return []
    }
    return [...menu.querySelectorAll<HTMLButtonElement>('button')]
  }

  const focusMenuItem = (index: number) => {
    const buttons = menuButtons()
    if (buttons.length === 0) {
      return
    }
    buttons[(index + buttons.length) % buttons.length]?.focus()
  }

  const closeActions = () => {
    // Only pull focus back to the trigger if a keyboard user was inside the menu;
    // an outside mouse click should leave focus wherever the user clicked.
    const menu = globalThis.document?.querySelector(`[data-element="${menuDataElement}"]`)
    const active = globalThis.document?.activeElement ?? null
    const focusWasInMenu = menu instanceof HTMLElement && menu.contains(active)

    openActionsForId.value = null
    if (focusWasInMenu) {
      triggerEl?.focus()
    }
    triggerEl = null
  }

  const toggleActions = (id: string, event: MouseEvent) => {
    if (openActionsForId.value === id) {
      closeActions()
      return
    }

    const trigger = event.currentTarget
    if (!(trigger instanceof HTMLElement)) {
      openActionsForId.value = id
      triggerEl = null
      return
    }

    triggerEl = trigger

    // Calculate menu position: align right, flip upward if too close to bottom
    const rect = trigger.getBoundingClientRect()
    const left = Math.max(8, rect.right - menuWidth)
    const openUpward = rect.bottom + gap + menuHeight > window.innerHeight - 8
    const top = openUpward ? Math.max(8, rect.top - gap - menuHeight) : rect.bottom + gap

    menuPosition.value = { top, left }
    openActionsForId.value = id
    // Move focus into the menu so it is operable by keyboard.
    void nextTick(() => focusMenuItem(0))
  }

  const onDocumentClick = (event: MouseEvent) => {
    const target = event.target
    if (!(target instanceof HTMLElement)) {
      closeActions()
      return
    }

    if (target.closest(`[data-element="${dataElement}"]`) || target.closest(`[data-element="${menuDataElement}"]`)) {
      return
    }

    closeActions()
  }

  const onKeyDown = (event: KeyboardEvent) => {
    if (openActionsForId.value === null) {
      return
    }

    const buttons = menuButtons()
    if (buttons.length === 0) {
      return
    }
    const currentIndex = buttons.indexOf(globalThis.document?.activeElement as HTMLButtonElement)

    switch (event.key) {
      case 'Escape':
        event.preventDefault()
        closeActions()
        break
      case 'ArrowDown':
        event.preventDefault()
        focusMenuItem(currentIndex < 0 ? 0 : currentIndex + 1)
        break
      case 'ArrowUp':
        event.preventDefault()
        focusMenuItem(currentIndex < 0 ? buttons.length - 1 : currentIndex - 1)
        break
      case 'Home':
        event.preventDefault()
        focusMenuItem(0)
        break
      case 'End':
        event.preventDefault()
        focusMenuItem(buttons.length - 1)
        break
      case 'Tab':
        // Keep focus within the open menu.
        event.preventDefault()
        focusMenuItem(event.shiftKey ? currentIndex - 1 : currentIndex + 1)
        break
    }
  }

  onMounted(() => {
    if (typeof document !== 'undefined') {
      document.addEventListener('click', onDocumentClick)
      document.addEventListener('keydown', onKeyDown)
    }
  })

  onBeforeUnmount(() => {
    if (typeof document !== 'undefined') {
      document.removeEventListener('click', onDocumentClick)
      document.removeEventListener('keydown', onKeyDown)
    }
  })

  return {
    openActionsForId,
    menuPosition,
    closeActions,
    toggleActions,
  }
}
