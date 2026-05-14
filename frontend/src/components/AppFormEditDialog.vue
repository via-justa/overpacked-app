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
  isDeleting?: boolean
}>()

defineEmits<{
  'update:open': [value: boolean]
  submit: []
  cancel: []
  delete: []
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

      <footer class="mt-4 flex flex-wrap items-center justify-between gap-2 shrink-0">
        <div class="flex flex-wrap items-center gap-2">
          <Button label="Save" icon="pi pi-check" :disabled="canSubmit === false || isSubmitting || isDeleting"
            :loading="isSubmitting ?? false" @click="$emit('submit')" />
          <Button label="Cancel" icon="pi pi-times" severity="secondary" outlined
            :disabled="(isSubmitting ?? false) || (isDeleting ?? false)" @click="$emit('cancel')" />
        </div>
        <Button label="Delete" icon="pi pi-trash" severity="danger" outlined
          :disabled="(isSubmitting ?? false) || (isDeleting ?? false)" :loading="isDeleting ?? false"
          @click="$emit('delete')" />
      </footer>
    </article>
  </AppTemplateDialog>
</template>
