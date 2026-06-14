<script setup lang="ts">
import { normalizeTitleWords } from '../../../lib/text/normalize'
import type { Label } from '../../items/types'
import type { ItemSet, SetItemWithDetails } from '../types'

type SetStats = {
  itemCount: number
  totalWeightGrams: number
  totalValue: number
}

type Props = {
  sets: ItemSet[]
  setStatsById: Record<string, SetStats>
  setItemsBySetId: Record<string, SetItemWithDetails[]>
  getItemTypeLabel: (categoryId: string) => string
  formatDisplayWeight: (grams: number) => string
  formatDate: (value: string) => string
  formatValue: (value: number) => string
  getSetLabels: (setId: string) => Label[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  openDetails: [set: ItemSet]
  startEdit: [set: ItemSet]
}>()

const getContrastColor = (color?: string | null): 'light' | 'dark' => {
  if (!color) {
    return 'light'
  }

  // Handle HSL colors
  if (color.startsWith('hsl')) {
    const match = color.match(/hsl\((\d+),\s*(\d+)%,\s*(\d+)%\)/)
    if (match) {
      const lightness = Number.parseInt(match[3], 10)
      return lightness > 55 ? 'dark' : 'light'
    }
  }

  // Handle hex colors
  const hex = color.replace('#', '')
  const r = Number.parseInt(hex.substring(0, 2), 16)
  const g = Number.parseInt(hex.substring(2, 4), 16)
  const b = Number.parseInt(hex.substring(4, 6), 16)

  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255

  return luminance > 0.5 ? 'dark' : 'light'
}

const getLabelTextColor = (color?: string | null): string => {
  const contrast = getContrastColor(color)
  return contrast === 'light' ? 'text-ink-inverse' : 'text-ink'
}

const getLabelBorderColor = (color?: string | null): string => {
  const contrast = getContrastColor(color)
  return contrast === 'light'
    ? 'rgba(255, 255, 255, 0.2)'
    : 'rgba(0, 0, 0, 0.1)'
}
</script>

<template>
  <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
    <button v-for="set in sets" :key="set.id" type="button"
      class="surface-panel hover:border-brand-300 cursor-pointer p-4 text-left transition"
      @click="emit('openDetails', set)">
      <h3 class="text-ink text-lg font-semibold">{{ normalizeTitleWords(set.name) }}</h3>
      <p class="text-copy-muted mt-2 text-sm">
        {{ getItemTypeLabel(set.set_category) }}
        <span class="text-line mx-2">/</span>
        {{ setStatsById[set.id]?.itemCount ?? 0 }} items
        <span class="text-line mx-2">/</span>
        {{ formatDisplayWeight(setStatsById[set.id]?.totalWeightGrams ?? 0) }}
        <span class="text-line mx-2">/</span>
        {{ formatValue(setStatsById[set.id]?.totalValue ?? 0) }}
      </p>

      <div v-if="getSetLabels(set.id).length > 0" class="mt-2 flex flex-wrap gap-1">
        <span v-for="label in getSetLabels(set.id)" :key="label.id"
          class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
          :class="getLabelTextColor(label.color)" :style="{
            backgroundColor: label.color ?? 'var(--color-label-fallback)',
            border: `1px solid ${getLabelBorderColor(label.color)}`
          }">
          {{ label.name }}
        </span>
      </div>

      <p class="text-copy-subtle mt-1 text-xs">Updated {{ formatDate(set.updated_at) }}</p>
    </button>
  </div>
</template>
