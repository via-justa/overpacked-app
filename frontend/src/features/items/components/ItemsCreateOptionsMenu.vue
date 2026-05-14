<script setup lang="ts">
type CreateTarget = 'item' | 'manufacturer' | 'category' | 'import'

interface CreateOption {
  value: CreateTarget
  label: string
  description: string
  icon: string
}

defineProps<{
  open: boolean
  position: { top: number; left: number }
  options: CreateOption[]
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  select: [target: CreateTarget]
}>()

const handleBackdropClick = () => {
  emit('update:open', false)
}

const handleOptionClick = (target: CreateTarget) => {
  emit('select', target)
}
</script>

<template>
  <div v-if="open" class="fixed inset-0 z-30" @click="handleBackdropClick" />

  <div v-if="open" data-element="items-create-options" class="fixed z-40 w-[min(24rem,calc(100vw-2rem))]"
    :style="{ top: `${position.top}px`, left: `${position.left}px` }">
    <section class="border-line-subtle bg-surface-elevated w-full rounded-2xl border p-3 shadow-panel backdrop-blur">
      <p class="text-copy-subtle px-2 pb-2 text-xs font-semibold uppercase tracking-[0.08em]">Create</p>
      <div class="grid gap-1">
        <button v-for="option in options" :key="option.value" type="button"
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
