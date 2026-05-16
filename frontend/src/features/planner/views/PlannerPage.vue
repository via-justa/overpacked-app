<script setup lang="ts">
import { RouterLink } from 'vue-router'
import SetsSection from '../components/SetsSection.vue'
import PackingListsSection from '../components/PackingListsSection.vue'
import PersonsSection from '../components/PersonsSection.vue'
</script>

<template>
  <section data-component="planner-page" class="flex w-full flex-col gap-4">
    <!-- Mobile: Launcher Cards -->
    <div class="flex flex-col gap-3 md:hidden">
      <RouterLink to="/sets" data-element="planner-sets-launcher"
        class="border-line-subtle bg-surface-elevated hover:border-brand-300 flex items-center gap-4 rounded-xl border p-4 transition">
        <div class="bg-brand-100 text-brand-600 flex h-12 w-12 shrink-0 items-center justify-center rounded-lg">
          <i class="pi pi-sitemap text-xl" aria-hidden="true"></i>
        </div>
        <div class="flex-1">
          <h3 class="text-copy text-base font-semibold">Sets</h3>
          <p class="text-copy-muted text-sm">Manage gear collections</p>
        </div>
        <i class="pi pi-chevron-right text-copy-subtle" aria-hidden="true"></i>
      </RouterLink>

      <RouterLink to="/lists" data-element="planner-lists-launcher"
        class="border-line-subtle bg-surface-elevated hover:border-brand-300 flex items-center gap-4 rounded-xl border p-4 transition">
        <div class="bg-brand-100 text-brand-600 flex h-12 w-12 shrink-0 items-center justify-center rounded-lg">
          <i class="pi pi-check-square text-xl" aria-hidden="true"></i>
        </div>
        <div class="flex-1">
          <h3 class="text-copy text-base font-semibold">Packing Lists</h3>
          <p class="text-copy-muted text-sm">Trip checklist templates</p>
        </div>
        <i class="pi pi-chevron-right text-copy-subtle" aria-hidden="true"></i>
      </RouterLink>

      <RouterLink to="/persons" data-element="planner-persons-launcher"
        class="border-line-subtle bg-surface-elevated hover:border-brand-300 flex items-center gap-4 rounded-xl border p-4 transition">
        <div class="bg-brand-100 text-brand-600 flex h-12 w-12 shrink-0 items-center justify-center rounded-lg">
          <i class="pi pi-users text-xl" aria-hidden="true"></i>
        </div>
        <div class="flex-1">
          <h3 class="text-copy text-base font-semibold">Persons</h3>
          <p class="text-copy-muted text-sm">Manage crew members</p>
        </div>
        <i class="pi pi-chevron-right text-copy-subtle" aria-hidden="true"></i>
      </RouterLink>
    </div>

    <!-- Desktop: All Sections -->
    <div class="hidden md:flex md:flex-col md:gap-8">
      <SetsSection />
      <PackingListsSection />
      <PersonsSection />
    </div>
  </section>
</template>
