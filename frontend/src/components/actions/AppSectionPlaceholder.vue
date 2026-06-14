<script setup lang="ts">
import AppIcon from '../icons/AppIcon.vue'

defineProps<{
  title: string
  description: string
}>()
</script>

<template>
  <section data-component="app-section-placeholder" :data-section="title.toLowerCase()"
    class="surface-panel p-8 sm:p-10">
    <h1 class="text-ink text-3xl font-bold sm:text-4xl">{{ title }}</h1>
    <p class="text-copy-muted mt-3 max-w-3xl text-base sm:text-lg">{{ description }}</p>
    <div data-element="placeholder-ready-banner"
      class="border-brand-200 bg-brand-50 text-brand-800 mt-8 flex items-center gap-3 rounded-2xl border px-4 py-3 text-sm font-medium">
      <AppIcon category="status" name="success" />
      <span>Page scaffolded and ready for feature implementation.</span>
    </div>
  </section>
</template>
