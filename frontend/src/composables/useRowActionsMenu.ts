import { onBeforeUnmount, onMounted, ref, type Ref } from 'vue'

export type MenuPosition = { top: number; left: number }

export interface UseRowActionsMenuOptions {
  menuWidth?: number
  menuHeight?: number
  gap?: number
  dataElement: string
}

export interface UseRowActionsMenuReturn {
  openActionsForId: Ref<string | null>
  menuPosition: Ref<MenuPosition>
  closeActions: () => void
  toggleActions: (id: string, event: MouseEvent) => void
}

/**
 * Composable for table/list row action menus with smart positioning.
 * Handles menu open/close state, viewport-aware positioning, and document click handling.
 */
export function useRowActionsMenu(options: UseRowActionsMenuOptions): UseRowActionsMenuReturn {
  const { menuWidth = 176, menuHeight = 188, gap = 6, dataElement } = options

  const openActionsForId = ref<string | null>(null)
  const menuPosition = ref<MenuPosition>({ top: 0, left: 0 })

  const closeActions = () => {
    openActionsForId.value = null
  }

  const toggleActions = (id: string, event: MouseEvent) => {
    if (openActionsForId.value === id) {
      openActionsForId.value = null
      return
    }

    const trigger = event.currentTarget
    if (!(trigger instanceof HTMLElement)) {
      openActionsForId.value = id
      return
    }

    // Calculate menu position: align right, flip upward if too close to bottom
    const rect = trigger.getBoundingClientRect()
    const left = Math.max(8, rect.right - menuWidth)
    const openUpward = rect.bottom + gap + menuHeight > window.innerHeight - 8
    const top = openUpward ? Math.max(8, rect.top - gap - menuHeight) : rect.bottom + gap

    menuPosition.value = { top, left }
    openActionsForId.value = id
  }

  const onDocumentClick = (event: MouseEvent) => {
    const target = event.target
    if (!(target instanceof HTMLElement)) {
      closeActions()
      return
    }

    if (target.closest(`[data-element="${dataElement}"]`)) {
      return
    }

    closeActions()
  }

  onMounted(() => {
    if (typeof document !== 'undefined') {
      document.addEventListener('click', onDocumentClick)
    }
  })

  onBeforeUnmount(() => {
    if (typeof document !== 'undefined') {
      document.removeEventListener('click', onDocumentClick)
    }
  })

  return {
    openActionsForId,
    menuPosition,
    closeActions,
    toggleActions,
  }
}
