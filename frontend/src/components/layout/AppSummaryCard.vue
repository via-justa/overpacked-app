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
  <div
    class="border-line-subtle bg-surface-elevated text-copy flex items-center justify-between gap-3 rounded-xl border px-3 py-2 text-sm">
    <span class="text-copy-subtle text-xs font-semibold uppercase tracking-[0.08em]">{{ label }}</span>
    <span class="text-ink text-xs font-semibold">{{ value }}</span>
  </div>
</template>
