<script setup lang="ts">
defineProps<{
  label?: string
}>()
</script>

<template>
  <span class="text-copy-muted inline-flex items-center opacity-70" :aria-label="label ?? 'Not set'"
    :title="label ?? 'Not set'">
    <i class="pi pi-minus-circle" aria-hidden="true"></i>
    <span class="sr-only">{{ label ?? 'Not set' }}</span>
  </span>
</template>
