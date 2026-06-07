<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { RouterLink } from 'vue-router'
import AppQueryState from '../../../components/feedback/AppQueryState.vue'
import { listPackingLists } from '../../lists/api/listsApi'

const packingListsQuery = useQuery({
  queryKey: ['packing-lists'],
  queryFn: listPackingLists,
})

const displayLists = computed(() => {
  const lists = packingListsQuery.data.value ?? []
  return lists.slice(0, 3) // Show first 3
})
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

    <AppQueryState :query="packingListsQuery" loading-message="Loading packing lists..."
      empty-message="Trip checklist templates coming soon. Create reusable packing lists for different adventures!"
      error-fallback="Unable to load packing lists.">
      <template #empty>
        <div class="border-line-subtle bg-surface-elevated text-copy-muted rounded-2xl border px-5 py-6 text-sm">
          Trip checklist templates coming soon. Create reusable packing lists for different adventures!
          <RouterLink to="/lists?create=1" class="text-brand-500 hover:text-brand-600 font-medium">
            Create your first
          </RouterLink>
        </div>
      </template>
      <div class="grid gap-3 sm:grid-cols-3">
        <RouterLink v-for="list in displayLists" :key="list.id" :to="{ path: '/lists', query: { open: list.id } }"
          class="border-line-subtle bg-surface-elevated hover:bg-surface-soft block rounded-xl border px-4 py-3 transition">
          <h3 class="text-ink text-sm font-semibold">{{ list.name }}</h3>
          <p v-if="list.description" class="text-copy-muted mt-1 text-xs line-clamp-2">
            {{ list.description }}
          </p>
        </RouterLink>
      </div>
    </AppQueryState>
  </section>
</template>
