import { ref } from 'vue'
import { defineStore } from 'pinia'

type SidebarMode = 'expanded' | 'collapsed'

const SIDEBAR_MODE_KEY = 'overpacked-app.ui.sidebarMode'

export const useUiStore = defineStore('ui', () => {
  const sidebarMode = ref<SidebarMode>((localStorage.getItem(SIDEBAR_MODE_KEY) as SidebarMode) || 'expanded')

  const setSidebarMode = (mode: SidebarMode) => {
    sidebarMode.value = mode
    localStorage.setItem(SIDEBAR_MODE_KEY, mode)
  }

  const toggleSidebar = () => {
    setSidebarMode(sidebarMode.value === 'expanded' ? 'collapsed' : 'expanded')
  }

  return {
    sidebarMode,
    setSidebarMode,
    toggleSidebar,
  }
})
