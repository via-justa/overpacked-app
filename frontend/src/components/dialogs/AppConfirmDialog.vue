<script setup lang="ts">
import { computed } from 'vue'
import AppTemplateDialog from './AppTemplateDialog.vue'
import AppActionButton from '../actions/AppActionButton.vue'
import AppActionCluster from '../actions/AppActionCluster.vue'

const props = withDefaults(defineProps<{
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

const confirmToneOverride = computed<'danger' | 'primary'>(() => (props.confirmTone === 'danger' ? 'danger' : 'primary'))
</script>

<template>
  <AppTemplateDialog :model-value="open" data-element="app-confirm-dialog" width="min(28rem, calc(100vw - 2rem))"
    @update:model-value="$emit('update:open', $event)">
    <article class="surface-panel relative p-4">
      <header class="mb-2">
        <h3 class="text-copy text-sm font-semibold uppercase tracking-[0.08em] pr-20">{{ title }}</h3>
      </header>

      <p class="text-copy-subtle text-sm leading-6">{{ message }}</p>

      <AppActionCluster data-element="confirm-dialog-actions">
        <AppActionButton action="cancel" :label="cancelLabel" data-element="confirm-dialog-cancel" @click="onCancel" />
        <AppActionButton action="confirm" :label="confirmLabel" :tone="confirmToneOverride"
          data-element="confirm-dialog-confirm" @click="onConfirm" />
      </AppActionCluster>
    </article>
  </AppTemplateDialog>
</template>
