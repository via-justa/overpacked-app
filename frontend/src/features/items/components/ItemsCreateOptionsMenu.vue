<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

type CreateTarget = 'item' | 'manufacturer' | 'category' | 'import'

interface CreateOption {
  value: CreateTarget
  label: string
  description: string
  icon: string
}

const props = defineProps<{
  open: boolean
  position: { top: number; left: number }
  options: CreateOption[]
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  select: [target: CreateTarget]
}>()

const menuRef = ref<HTMLElement | null>(null)
// The element that opened the menu; focus is restored to it on close.
let triggerEl: HTMLElement | null = null

const handleBackdropClick = () => {
  emit('update:open', false)
}

const handleOptionClick = (target: CreateTarget) => {
  emit('select', target)
}

const optionButtons = (): HTMLButtonElement[] => {
  if (!menuRef.value) {
    return []
  }
  return [...menuRef.value.querySelectorAll<HTMLButtonElement>('[data-create-option]')]
}

const focusOption = (index: number) => {
  const buttons = optionButtons()
  if (buttons.length === 0) {
    return
  }
  buttons[(index + buttons.length) % buttons.length]?.focus()
}

// Keyboard navigation: arrows/Home/End move focus, Escape closes, Tab is trapped.
const handleKeyDown = (event: KeyboardEvent) => {
  if (!props.open) {
    return
  }
  const buttons = optionButtons()
  if (buttons.length === 0) {
    return
  }
  const currentIndex = buttons.indexOf(globalThis.document?.activeElement as HTMLButtonElement)

  switch (event.key) {
    case 'Escape':
      event.preventDefault()
      emit('update:open', false)
      break
    case 'ArrowDown':
      event.preventDefault()
      focusOption(currentIndex < 0 ? 0 : currentIndex + 1)
      break
    case 'ArrowUp':
      event.preventDefault()
      focusOption(currentIndex < 0 ? buttons.length - 1 : currentIndex - 1)
      break
    case 'Home':
      event.preventDefault()
      focusOption(0)
      break
    case 'End':
      event.preventDefault()
      focusOption(buttons.length - 1)
      break
    case 'Tab':
      event.preventDefault()
      focusOption(event.shiftKey ? currentIndex - 1 : currentIndex + 1)
      break
  }
}

// Move focus into the menu on open; restore it to the trigger on close.
watch(() => props.open, async (isOpen) => {
  if (!isOpen) {
    triggerEl?.focus()
    triggerEl = null
    return
  }

  const active = globalThis.document?.activeElement
  triggerEl = active instanceof HTMLElement ? active : null

  await nextTick()
  focusOption(0)
})

onMounted(() => {
  globalThis.document?.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  globalThis.document?.removeEventListener('keydown', handleKeyDown)
})
</script>

<template>
  <div v-if="open" class="fixed inset-0 z-30" @click="handleBackdropClick" />

  <div v-if="open" ref="menuRef" data-element="items-create-options" role="menu" aria-label="Create"
    class="fixed z-40 w-[min(24rem,calc(100vw-2rem))]"
    :style="{ top: `${position.top}px`, left: `${position.left}px` }">
    <section class="border-line-subtle bg-surface-elevated w-full rounded-2xl border p-3 shadow-panel backdrop-blur">
      <p class="text-copy-subtle px-2 pb-2 text-xs font-semibold uppercase tracking-[0.08em]">Create</p>
      <div class="grid gap-1">
        <button v-for="option in options" :key="option.value" type="button" role="menuitem" data-create-option
          class="hover:bg-surface-soft flex items-start gap-3 rounded-lg px-3 py-2 text-left transition"
          @click="handleOptionClick(option.value)">
          <span class="bg-brand-50 text-brand-500 mt-0.5 inline-flex h-7 w-7 items-center justify-center rounded-full">
            <i :class="option.icon" class="text-sm" />
          </span>
          <span>
            <span class="text-ink block text-sm font-semibold">{{ option.label }}</span>
            <span class="text-copy-muted block text-xs">{{ option.description }}</span>
          </span>
        </button>
      </div>
    </section>
  </div>
</template>
