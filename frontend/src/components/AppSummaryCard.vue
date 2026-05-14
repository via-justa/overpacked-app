<script setup lang="ts">
/**
 * Reusable summary stat card.
 * Used for displaying aggregate metrics in ItemsPage, SetsPage, etc.
 */
defineProps<{
  label: string
  value: string | number
}>()
</script>

<template>
  <div class="border-line-subtle bg-surface-elevated text-copy rounded-xl border px-3 py-2 text-sm">
    <span class="text-copy-subtle text-xs font-semibold uppercase tracking-[0.08em]">{{ label }}</span>
    <p class="text-ink mt-1 text-base font-semibold">{{ value }}</p>
  </div>
</template>
