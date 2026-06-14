<script setup lang="ts">
import AppIcon from '../icons/AppIcon.vue'

defineProps<{
  label?: string
}>()
</script>

<template>
  <span class="text-copy-muted inline-flex items-center opacity-70" :aria-label="label ?? 'Not set'"
    :title="label ?? 'Not set'">
    <AppIcon category="status" name="notSet" />
    <span class="sr-only">{{ label ?? 'Not set' }}</span>
  </span>
</template>
