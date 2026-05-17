<script setup lang="ts">
import AppTemplateDialog from './AppTemplateDialog.vue'

withDefaults(defineProps<{
  open: boolean
  title?: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  confirmTone?: 'default' | 'danger'
}>(), {
  title: 'Confirm action',
  confirmLabel: 'Confirm',
  cancelLabel: 'Cancel',
  confirmTone: 'default',
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  confirm: []
  cancel: []
}>()

const onCancel = () => {
  emit('cancel')
  emit('update:open', false)
}

const onConfirm = () => {
  emit('confirm')
}
</script>

<template>
  <AppTemplateDialog :model-value="open" data-element="app-confirm-dialog" width="min(28rem, calc(100vw - 2rem))"
    @update:model-value="$emit('update:open', $event)">
    <article class="surface-panel p-4">
      <header class="mb-2">
        <h3 class="text-copy text-sm font-semibold uppercase tracking-[0.08em]">{{ title }}</h3>
      </header>

      <p class="text-copy-subtle text-sm leading-6">{{ message }}</p>

      <div class="mt-4 flex justify-end gap-2">
        <button type="button"
          class="border-line-subtle text-copy hover:bg-surface-soft rounded-md border px-3 py-1.5 text-sm font-medium"
          @click="onCancel">
          {{ cancelLabel }}
        </button>
        <button type="button" class="rounded-md px-3 py-1.5 text-sm font-semibold" :class="confirmTone === 'danger'
          ? 'border hover:bg-[color-mix(in_srgb,var(--color-danger-500)_22%,var(--color-surface-elevated))]'
          : 'bg-brand-600 text-ink-inverse hover:bg-brand-700'"
          :style="confirmTone === 'danger' ? { borderColor: 'color-mix(in srgb, var(--color-danger-500) 55%, var(--color-line-subtle))', color: 'var(--color-danger-500)' } : undefined"
          @click="onConfirm">
          {{ confirmLabel }}
        </button>
      </div>
    </article>
  </AppTemplateDialog>
</template>
