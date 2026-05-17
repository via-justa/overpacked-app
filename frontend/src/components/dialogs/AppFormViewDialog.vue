<script setup lang="ts">
import Button from 'primevue/button'
import AppTemplateDialog from './AppTemplateDialog.vue'

// Dialog wrapper for read-only content display with close button
defineProps<{
  open: boolean
  title: string
  width?: string
  dataElement?: string
}>()

defineEmits<{
  'update:open': [value: boolean]
}>()
</script>

<template>
  <AppTemplateDialog :model-value="open" :width="width ?? 'min(44rem, calc(100vw - 2rem))'"
    :data-element="dataElement" @update:model-value="$emit('update:open', $event)">
    <article class="surface-panel p-4 flex flex-col">
      <h2 class="text-ink text-lg font-semibold shrink-0">{{ title }}</h2>
      <div class="flex-1 mt-4">
        <slot />
      </div>

      <footer class="mt-4 flex shrink-0">
        <Button label="Close" icon="pi pi-times" severity="secondary" outlined
          @click="$emit('update:open', false)" />
      </footer>
    </article>
  </AppTemplateDialog>
</template>
