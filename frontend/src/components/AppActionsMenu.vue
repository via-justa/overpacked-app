<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

export type ActionTarget = 'add-item' | 'add-set' | 'add-person' | 'add-packing-list' | 'manage-manufacturers' | 'manage-categories' | 'import-csv' | 'settings' | 'logout' | 'dashboard' | 'planner' | 'sets' | 'lists' | 'gear' | 'persons'

interface ActionOption {
  value: ActionTarget
  label: string
  description: string
  icon: string
}

const props = defineProps<{
  open: boolean
  position: { top: number; left: number }
  currentPath?: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  select: [target: ActionTarget]
}>()

const menuRef = ref<HTMLElement | null>(null)
const focusedIndex = ref(0)
const actionsExpanded = ref(false)

const actionOptions: ActionOption[] = [
  { value: 'add-item', label: 'Add Item', description: 'Create a new gear item.', icon: 'pi pi-box' },
  { value: 'add-set', label: 'Add Set', description: 'Create a new gear set.', icon: 'pi pi-sitemap' },
  { value: 'add-person', label: 'Add Person', description: 'Create a new person.', icon: 'pi pi-user' },
  { value: 'add-packing-list', label: 'Add Packing List', description: 'Create a new packing list.', icon: 'pi pi-check-square' },
  { value: 'manage-manufacturers', label: 'Manage Manufacturers', description: 'Create and edit manufacturers.', icon: 'pi pi-building' },
  { value: 'manage-categories', label: 'Manage Categories', description: 'Create and edit custom categories.', icon: 'pi pi-tag' },
  { value: 'import-csv', label: 'Import from CSV', description: 'Preview and import gear from CSV.', icon: 'pi pi-upload' },
  { value: 'settings', label: 'Settings', description: 'Configure app preferences.', icon: 'pi pi-cog' },
  { value: 'logout', label: 'Logout', description: 'Sign out of your account.', icon: 'pi pi-sign-out' },
]

const desktopNavigationOptions: ActionOption[] = [
  { value: 'dashboard' as ActionTarget, label: 'Dashboard', description: 'View your dashboard.', icon: 'pi pi-home' },
  { value: 'planner' as ActionTarget, label: 'Planner', description: 'Sets, lists, and persons.', icon: 'pi pi-list-check' },
  { value: 'gear' as ActionTarget, label: 'Gear', description: 'Manage your gear items.', icon: 'pi pi-box' },
]

const mobileNavigationOptions: ActionOption[] = [
  { value: 'dashboard' as ActionTarget, label: 'Dashboard', description: 'View your dashboard.', icon: 'pi pi-home' },
  { value: 'sets' as ActionTarget, label: 'Sets', description: 'Manage your gear sets.', icon: 'pi pi-sitemap' },
  { value: 'lists' as ActionTarget, label: 'Packing Lists', description: 'Trip checklist templates.', icon: 'pi pi-check-square' },
  { value: 'persons' as ActionTarget, label: 'Persons', description: 'Manage persons.', icon: 'pi pi-users' },
  { value: 'gear' as ActionTarget, label: 'Gear', description: 'Manage your gear items.', icon: 'pi pi-box' },
]

const isMobile = ref(false)

const updateMobile = () => {
  isMobile.value = globalThis.window?.innerWidth < 768
}

onMounted(() => {
  if (globalThis.window) {
    updateMobile()
    globalThis.window.addEventListener('resize', updateMobile)
  }
})

onUnmounted(() => {
  if (globalThis.window) {
    globalThis.window.removeEventListener('resize', updateMobile)
  }
})

const allOptions = computed(() => {
  const navigationItems = isMobile.value ? mobileNavigationOptions : desktopNavigationOptions
  if (isMobile.value) {
    // On mobile, if actions are collapsed, only show navigation + settings + logout
    if (!actionsExpanded.value) {
      return [...navigationItems, actionOptions[actionOptions.length - 2], actionOptions[actionOptions.length - 1]]
    }
  }
  return isMobile.value ? [...navigationItems, ...actionOptions] : actionOptions
})

const handleBackdropClick = () => {
  emit('update:open', false)
}

const handleOptionClick = (target: ActionTarget) => {
  emit('select', target)
}

const handleMouseEnter = (index: number) => {
  focusedIndex.value = index
  focusButton(index)
}

const handleKeyDown = (event: KeyboardEvent) => {
  if (!props.open) return

  const optionsLength = allOptions.value.length

  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      focusedIndex.value = (focusedIndex.value + 1) % optionsLength
      focusButton(focusedIndex.value)
      break
    case 'ArrowUp':
      event.preventDefault()
      focusedIndex.value = (focusedIndex.value - 1 + optionsLength) % optionsLength
      focusButton(focusedIndex.value)
      break
    case 'Home':
      event.preventDefault()
      focusedIndex.value = 0
      focusButton(0)
      break
    case 'End':
      event.preventDefault()
      focusedIndex.value = optionsLength - 1
      focusButton(optionsLength - 1)
      break
    case 'Enter':
    case ' ':
      event.preventDefault()
      handleOptionClick(allOptions.value[focusedIndex.value].value)
      break
    case 'Escape':
      event.preventDefault()
      emit('update:open', false)
      break
  }
}

const focusButton = (index: number) => {
  if (!menuRef.value) return
  const buttons = menuRef.value.querySelectorAll<HTMLButtonElement>('[data-action-option]')
  buttons[index]?.focus()
}

watch(() => props.open, async (isOpen) => {
  // Prevent body scroll when menu is open
  if (globalThis.document?.body) {
    if (isOpen) {
      globalThis.document.body.style.overflow = 'hidden'
    } else {
      globalThis.document.body.style.overflow = ''
    }
  }

  if (isOpen) {
    // Reset actions expanded state on mobile
    if (isMobile.value) {
      actionsExpanded.value = false
    }

    // Find current page in menu and focus it, otherwise focus first item
    const currentPath = props.currentPath ?? ''
    let targetIndex = 0

    if (currentPath) {
      const pathSegment = currentPath.split('/')[1] || 'dashboard'

      // Map path segments to option values
      let optionValue = pathSegment

      // Special mappings
      if (pathSegment === 'gear') {
        optionValue = 'gear'
      } else if (pathSegment === '' || pathSegment === 'dashboard') {
        optionValue = 'dashboard'
      } else if (!isMobile.value && (pathSegment === 'sets' || pathSegment === 'lists' || pathSegment === 'persons')) {
        // On desktop, sets/lists/persons should highlight planner
        optionValue = 'planner'
      }

      const index = allOptions.value.findIndex(opt => opt.value === optionValue)
      if (index !== -1) {
        targetIndex = index
      }
    }

    focusedIndex.value = targetIndex
    await nextTick()
    focusButton(targetIndex)
  }
})

onMounted(() => {
  if (globalThis.document) {
    globalThis.document.addEventListener('keydown', handleKeyDown)
  }
})

onUnmounted(() => {
  if (globalThis.document) {
    globalThis.document.removeEventListener('keydown', handleKeyDown)
    // Restore body scroll on unmount
    if (globalThis.document.body) {
      globalThis.document.body.style.overflow = ''
    }
  }
})
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-50" @click="handleBackdropClick" @touchmove.prevent />

    <div v-if="open" ref="menuRef" data-element="app-actions-menu" role="menu" aria-label="Actions menu"
      class="fixed z-60 w-[min(24rem,calc(100vw-2rem))] flex flex-col" :style="{
        top: `${position.top}px`,
        left: `${position.left}px`,
        maxHeight: `calc(100vh - ${position.top}px - 1rem)`,
        backgroundColor: 'white',
      }" @click.stop>
      <section
        class="border-line-subtle bg-surface-elevated w-full rounded-2xl border shadow-panel backdrop-blur overflow-hidden flex flex-col h-full">
        <p class="text-copy-subtle shrink-0 px-2 pb-2 pt-3 text-xs font-semibold uppercase tracking-[0.08em]">Menu</p>
        <div class="flex-1 overflow-y-auto px-3 pb-3" style="overscroll-behavior: contain;">
          <div class="grid gap-2 pb-32">
            <template v-for="(option, index) in allOptions" :key="option.value">
              <!-- Separator after navigation items on mobile -->
              <div v-if="isMobile && index === mobileNavigationOptions.length" class="border-line-subtle my-1 border-t">
              </div>
              <!-- Actions section header on mobile -->
              <button v-if="isMobile && index === mobileNavigationOptions.length" type="button"
                class="text-copy-subtle hover:text-copy flex items-center justify-between px-1 pb-1 pt-2 text-xs font-semibold uppercase tracking-[0.08em] transition"
                @click="actionsExpanded = !actionsExpanded">
                <span>Actions</span>
                <i :class="actionsExpanded ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" class="text-xs"
                  aria-hidden="true"></i>
              </button>
              <!-- Separator after collapsed Actions header -->
              <div v-if="isMobile && !actionsExpanded && index === mobileNavigationOptions.length"
                class="border-line-subtle my-1 border-t"></div>
              <!-- Separator before settings (after import-csv or last action option) -->
              <div v-if="index === allOptions.length - 2 && (!isMobile || actionsExpanded)"
                class="border-line-subtle my-1 border-t"></div>
              <button type="button" role="menuitem" :data-action-option="option.value"
                class="bg-surface-elevated hover:bg-surface-soft focus:bg-surface-soft outline-none ring-2 ring-transparent hover:ring-brand-500 focus:ring-brand-500 ring-offset-2 flex items-start gap-3 rounded-lg px-3 py-2 text-left transition"
                @mouseenter="handleMouseEnter(index)" @click="handleOptionClick(option.value)">
                <span
                  class="bg-brand-50 text-brand-500 mt-0.5 inline-flex h-7 w-7 items-center justify-center rounded-full"
                  aria-hidden="true">
                  <i :class="option.icon" class="text-sm" />
                </span>
                <span>
                  <span class="text-ink block text-sm font-semibold">{{ option.label }}</span>
                  <span class="text-copy-muted block text-xs">{{ option.description }}</span>
                </span>
              </button>
            </template>
          </div>
        </div>
      </section>
    </div>
  </Teleport>
</template>
