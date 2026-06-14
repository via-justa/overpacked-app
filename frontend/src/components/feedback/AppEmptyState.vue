<script setup lang="ts">
defineProps<{
  message: string
  dataElement?: string
}>()
</script>

<template>
  <div :data-element="dataElement"
    class="border-line-subtle bg-surface-elevated text-copy-muted rounded-2xl border px-5 py-6 text-sm">
    {{ message }}
  </div>
</template>
