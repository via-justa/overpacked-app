<script setup lang="ts">
import { computed } from 'vue'
import ItemLabel from '../../items/components/ItemLabel.vue'
import type { PackingList, Label } from '../types'

const props = defineProps<{
  packingList: PackingList
  labels: Label[]
}>()

const emit = defineEmits<{
  openEdit: [packingList: PackingList]
}>()

const displayedLabels = computed(() => props.labels.slice(0, 5))
const remainingCount = computed(() => Math.max(0, props.labels.length - 5))
</script>

<template>
  <article data-component="packing-list-card"
    class="border-line-subtle bg-surface-elevated hover:bg-surface-soft group cursor-pointer rounded-2xl border p-4 shadow-card transition"
    @click="emit('openEdit', packingList)">
    <div class="flex flex-col gap-3">
      <div>
        <h3 class="text-ink text-base font-semibold">{{ packingList.name }}</h3>
        <p v-if="packingList.description" class="text-copy-muted mt-1 text-sm line-clamp-2">
          {{ packingList.description }}
        </p>
      </div>

      <div v-if="labels.length > 0" class="flex flex-wrap gap-1.5">
        <ItemLabel v-for="label in displayedLabels" :key="label.id" :label="label" size="sm" />
        <span v-if="remainingCount > 0"
          class="bg-surface-muted text-copy-subtle inline-flex items-center rounded-full border border-gray-300 px-2 py-0.5 text-xs font-medium">
          +{{ remainingCount }}
        </span>
      </div>

      <div v-else class="text-copy-muted text-xs">
        No labels added yet
      </div>
    </div>
  </article>
</template>
