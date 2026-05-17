<script setup lang="ts">
import { RouterLink } from 'vue-router'
import AppSectionPlaceholder from '../../../components/actions/AppSectionPlaceholder.vue'
</script>

<template>
  <section data-component="packing-lists-page" class="flex w-full flex-col gap-4">
    <!-- Header -->
    <div class="hidden items-center justify-between md:flex">
      <h1 class="text-copy text-2xl font-bold">Packing Lists</h1>
      <RouterLink to="/planner" class="text-brand-500 hover:text-brand-600 text-sm font-medium">
        ← Back to Planner
      </RouterLink>
    </div>

    <AppSectionPlaceholder title="Packing Lists"
      description="Create and manage trip packing checklist templates for different adventures." />
  </section>
</template>
