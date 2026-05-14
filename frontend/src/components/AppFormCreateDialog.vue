<script setup lang="ts">
import AppTemplateDialog from './AppTemplateDialog.vue'
import AppDialogActions from './AppDialogActions.vue'

defineProps<{
  open: boolean
  title: string
  width?: string
  dataElement?: string
  canSubmit?: boolean
  isSubmitting?: boolean
}>()

defineEmits<{
  'update:open': [value: boolean]
  submit: []
  cancel: []
}>()
</script>

<template>
  <AppTemplateDialog :model-value="open" :width="width ?? 'min(36rem, calc(100vw - 2rem))'"
    :data-element="dataElement" @update:model-value="$emit('update:open', $event)">
    <article class="surface-panel p-4 flex flex-col">
      <h2 class="text-ink text-lg font-semibold shrink-0">{{ title }}</h2>
      <div class="flex-1 mt-4">
        <slot />
      </div>

      <AppDialogActions mode="create" :can-submit="canSubmit" :is-creating="isSubmitting"
        @submit="$emit('submit')" @cancel="$emit('cancel')" />
    </article>
  </AppTemplateDialog>
</template>
