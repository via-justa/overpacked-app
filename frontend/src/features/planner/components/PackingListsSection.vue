<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <section data-component="packing-lists-section" class="space-y-3">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <h2 class="text-copy text-xl font-bold">Packing Lists</h2>
      </div>
      <RouterLink to="/lists" class="text-brand-500 hover:text-brand-600 text-sm font-medium">
        View All →
      </RouterLink>
    </div>

    <!-- Placeholder Content -->
    <div class="border-line-subtle bg-surface-elevated text-copy-muted rounded-xl border px-4 py-3 text-sm">
      Trip checklist templates coming soon. Create reusable packing lists for different adventures!
    </div>
  </section>
</template>
