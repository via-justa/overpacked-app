<script setup lang="ts">
import Button from 'primevue/button'
import AppTemplateDialog from './AppTemplateDialog.vue'

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
    <article class="border-line-subtle bg-surface-elevated rounded-2xl border p-4 shadow-panel flex flex-col">
      <h2 class="text-ink text-lg font-semibold shrink-0">{{ title }}</h2>
      <div class="flex-1 mt-4">
        <slot />
      </div>

      <footer class="mt-4 flex flex-wrap items-center gap-2 shrink-0">
        <Button label="Save" icon="pi pi-check" :disabled="canSubmit === false || isSubmitting"
          :loading="isSubmitting ?? false" @click="$emit('submit')" />
        <Button label="Cancel" icon="pi pi-times" severity="secondary" outlined
          :disabled="isSubmitting ?? false" @click="$emit('cancel')" />
      </footer>
    </article>
  </AppTemplateDialog>
</template>
