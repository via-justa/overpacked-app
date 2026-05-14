<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <section data-component="settings-section-card"
    class="surface-panel p-6 sm:p-7 rounded-3xl backdrop-blur">
    <header data-element="settings-section-header" class="mb-5">
      <h2 class="text-ink text-xl font-bold sm:text-2xl">{{ title }}</h2>
      <p v-if="description" class="text-copy-muted mt-2 text-sm sm:text-base">{{ description }}</p>
    </header>
    <slot />
  </section>
</template>
