<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import AppActionsMenu from '../actions/AppActionsMenu.vue'
import type { ActionTarget } from '../actions/actionOptions'
import { useActionRouter } from '../actions/useActionRouter'
import { useActionRail } from '../../composables/useActionRail'
import AppGlobalSearch from './AppGlobalSearch.vue'
import { AppIcon } from '../icons'
import type { IconCategory } from '../../lib/icons'

type NavItem = {
  to: string
  label: string
  iconCategory: IconCategory
  iconName: string
}

const props = defineProps<{
  navItems: NavItem[]
  currentPath: string
}>()

const emit = defineEmits<{
  logout: []
}>()

const onLogout = () => {
  emit('logout')
}

const { pinned, togglePinned } = useActionRail()
const { selectAction } = useActionRouter(onLogout)

const isActionsMenuOpen = ref(false)
const actionsMenuPosition = ref<{ top: number; left: number }>({ top: 84, left: 16 })

// Desktop pins/unpins the action rail; mobile opens the overlay actions menu.
const onMenuButtonClick = (event: Event) => {
  const isMobile = (globalThis.window?.innerWidth ?? 1024) < 768

  if (isMobile) {
    openActionsMenu(event)
    return
  }

  togglePinned()
}

// Position actions menu relative to trigger button, responsive to viewport
const openActionsMenu = (event: Event) => {
  const trigger = event.currentTarget
  if (!(trigger instanceof HTMLElement)) {
    isActionsMenuOpen.value = true
    return
  }

  const rect = trigger.getBoundingClientRect()
  const viewportWidth = globalThis.window?.innerWidth ?? 1024

  // On mobile, center the menu; on desktop, align to button
  const isMobile = viewportWidth < 768

  actionsMenuPosition.value = {
    top: rect.bottom + 8,
    left: isMobile ? 16 : Math.max(16, rect.left),
  }

  isActionsMenuOpen.value = true
}

const closeActionsMenu = () => {
  isActionsMenuOpen.value = false
}

const onSelectAction = async (target: ActionTarget) => {
  closeActionsMenu()
  await selectAction(target)
}

const primaryNavItems = props.navItems.filter((item) => item.to !== '/settings')
</script>

<template>
  <header data-component="app-top-nav"
    class="border-line-subtle bg-surface-elevated fixed inset-x-0 top-0 z-40 border-b backdrop-blur">
    <div class="flex w-full flex-wrap items-center justify-between gap-3 px-4 py-3 sm:px-6 lg:px-10">
      <div class="flex items-center gap-3">
        <!-- On md+ the burger sits in a flush-left 64px slot so it caps the action rail column;
             negative margins cancel the nav's own left padding. -->
        <div class="flex shrink-0 justify-center md:-ml-6 md:w-16 lg:-ml-10">
          <button ref="actionsButtonRef" type="button" data-element="nav-menu-button"
            class="inline-flex h-8 w-8 items-center justify-center rounded-lg transition"
            :class="pinned
              ? 'bg-surface-inverse text-ink-inverse'
              : 'text-copy hover:bg-surface-soft hover:text-ink'"
            aria-label="Menu" :aria-pressed="pinned" @click="onMenuButtonClick">
            <AppIcon category="navigation" name="menu" size="sm" />
          </button>
        </div>
        <RouterLink to="/trips" class="flex h-10 items-center overflow-hidden" aria-label="Overpacked home">
          <img src="/logo.png" alt="Overpacked" class="h-full w-44 object-cover object-center" />
        </RouterLink>
      </div>

      <nav data-element="top-nav-primary" class="hidden items-center gap-1 md:flex">
        <RouterLink v-for="item in primaryNavItems" :key="item.to" :to="item.to"
          :data-element="`nav-link-${item.to.replace('/', '') || 'home'}`"
          :aria-current="currentPath.startsWith(item.to) ? 'page' : undefined"
          class="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition" :class="currentPath.startsWith(item.to)
            ? 'bg-surface-inverse text-ink-inverse'
            : 'text-copy hover:bg-surface-soft hover:text-ink'">
          <AppIcon :category="item.iconCategory" :name="item.iconName" size="sm" />
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>

      <!-- Global search sits on the right on desktop and wraps to full width on mobile -->
      <div data-element="top-nav-search" class="flex w-full justify-end md:w-auto md:max-w-md md:flex-1">
        <AppGlobalSearch />
      </div>
    </div>

    <AppActionsMenu :open="isActionsMenuOpen" :position="actionsMenuPosition" :current-path="props.currentPath"
      @update:open="isActionsMenuOpen = $event" @select="onSelectAction" />
  </header>
</template>
