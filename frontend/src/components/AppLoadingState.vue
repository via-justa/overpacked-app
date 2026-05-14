<script setup lang="ts">
defineProps<{
  message: string
  dataElement?: string
}>()
</script>

<template>
  <div :data-element="dataElement"
    class="border-line-subtle bg-surface-muted text-copy-muted rounded-2xl border px-4 py-3 text-sm font-medium">
    {{ message }}
  </div>
</template>
