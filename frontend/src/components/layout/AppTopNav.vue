<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import AppActionsMenu, { type ActionTarget } from '../actions/AppActionsMenu.vue'
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

const router = useRouter()

const isActionsMenuOpen = ref(false)
const actionsMenuPosition = ref<{ top: number; left: number }>({ top: 84, left: 16 })

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

  if (target === 'logout') {
    onLogout()
    return
  }

  const actionRoutes: Record<Exclude<ActionTarget, 'logout'>, { path: string; query?: Record<string, string> }> = {
    'add-item': { path: '/gear', query: { action: 'create-item' } },
    'add-set': { path: '/sets', query: { create: '1' } },
    'add-person': { path: '/persons', query: { create: '1' } },
    'add-packing-list': { path: '/lists', query: { create: '1' } },
    'manage-manufacturers': { path: '/gear', query: { action: 'manufacturers' } },
    'manage-categories': { path: '/gear', query: { action: 'categories' } },
    'import-csv': { path: '/gear', query: { action: 'import' } },
    'settings': { path: '/settings' },
    'dashboard': { path: '/dashboard' },
    'planner': { path: '/planner' },
    'sets': { path: '/sets' },
    'lists': { path: '/lists' },
    'gear': { path: '/gear' },
    'persons': { path: '/persons' },
  }

  const route = actionRoutes[target]
  await router.push(route)
}

const primaryNavItems = props.navItems.filter((item) => item.to !== '/settings')
</script>

<template>
  <header data-component="app-top-nav"
    class="border-line-subtle bg-surface-elevated fixed inset-x-0 top-0 z-40 border-b backdrop-blur">
    <div class="flex w-full flex-wrap items-center justify-between gap-3 px-4 py-3 sm:px-6 lg:px-10">
      <div class="flex items-center gap-3">
        <button ref="actionsButtonRef" type="button" data-element="nav-menu-button"
          class="text-copy hover:bg-surface-soft hover:text-ink inline-flex h-8 w-8 items-center justify-center rounded-lg transition"
          aria-label="Menu" @click="openActionsMenu">
          <AppIcon category="navigation" name="menu" size="sm" />
        </button>
        <p class="text-brand-500 text-xs font-semibold uppercase tracking-[0.18em]">Overpacked</p>
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
