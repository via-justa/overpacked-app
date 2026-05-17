<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue'
import { RouterLink } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import AppQueryState from '../../../components/feedback/AppQueryState.vue'
import { listSets, listSetItems } from '../../sets/api/setsApi'
import { listItemTypes, listItemLabels } from '../../items/api/itemsApi'
import { useSettings } from '../../../composables/useSettings'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { formatDisplayWeight } from '../../../lib/units/conversions'
import { formatValue } from '../../../lib/format/display'
import type { SetItemWithDetails } from '../../sets/types'
import type { Label } from '../../items/types'

const { weightUnit, currency } = useSettings()

const setsQuery = useQuery({
  queryKey: ['sets'],
  queryFn: listSets,
})

const itemTypesQuery = useQuery({
  queryKey: ['item-types'],
  queryFn: listItemTypes,
})

const displaySets = computed(() => {
  if (!setsQuery.data.value) return []
  return setsQuery.data.value.slice(0, 6)
})

const totalSets = computed(() => setsQuery.data.value?.length ?? 0)

const canShowContent = computed(() => {
  return !setsQuery.isPending.value && !setsQuery.isError.value && totalSets.value > 0
})

const setStatsById = ref<Record<string, { itemCount: number; totalWeightGrams: number; totalValue: number }>>({})
const setItemsBySetId = ref<Record<string, SetItemWithDetails[]>>({})

const getItemTypeLabel = (categoryId: string): string => {
  if (!categoryId) return 'Not set'
  const type = itemTypesQuery.data.value?.find(t => t.id === categoryId)
  return type ? normalizeTitleWords(type.name) : categoryId
}

const computeSetStats = (setItems: SetItemWithDetails[]) => {
  const itemCount = setItems.length
  const totalWeightGrams = setItems.reduce((sum, entry) => {
    const itemWeight = typeof entry.item.weight_grams === 'number' ? entry.item.weight_grams : 0
    return sum + (itemWeight * entry.quantity)
  }, 0)
  const totalValue = setItems.reduce((sum, entry) => {
    const itemValue = typeof entry.item.value === 'number' ? entry.item.value : 0
    return sum + (itemValue * entry.quantity)
  }, 0)

  return { itemCount, totalWeightGrams, totalValue }
}

const loadSetItems = async (setId: string) => {
  try {
    const data = await listSetItems(setId)
    setItemsBySetId.value = {
      ...setItemsBySetId.value,
      [setId]: data,
    }
    setStatsById.value = {
      ...setStatsById.value,
      [setId]: computeSetStats(data),
    }
  } catch {
    // Silently fail for section view
  }
}

watchEffect(() => {
  if (displaySets.value.length > 0) {
    displaySets.value.forEach(set => {
      if (!setItemsBySetId.value[set.id]) {
        void loadSetItems(set.id)
      }
    })
  }
})

// Labels
const itemLabelsData = ref<Record<string, Label[]>>({})

watchEffect(() => {
  // Load labels for all items in displayed sets
  displaySets.value.forEach(set => {
    const setItems = setItemsBySetId.value[set.id]
    if (setItems) {
      setItems.forEach(async si => {
        if (!itemLabelsData.value[si.item.id]) {
          try {
            const labels = await listItemLabels(si.item.id)
            itemLabelsData.value = {
              ...itemLabelsData.value,
              [si.item.id]: labels,
            }
          } catch {
            // Silently fail
          }
        }
      })
    }
  })
})

const getSetLabels = (setId: string): Label[] => {
  const setItems = setItemsBySetId.value[setId]
  if (!setItems || setItems.length === 0) return []

  const labelsMap = new Map<string, Label>()
  for (const si of setItems) {
    const labels = itemLabelsData.value[si.item.id]
    if (labels) {
      for (const label of labels) {
        if (!labelsMap.has(label.id)) {
          labelsMap.set(label.id, label)
        }
      }
    }
  }

  return Array.from(labelsMap.values())
}

const getContrastColor = (color?: string | null): 'light' | 'dark' => {
  if (!color) return 'light'

  if (color.startsWith('hsl')) {
    const match = color.match(/hsl\((\d+),\s*(\d+)%,\s*(\d+)%\)/)
    if (match) {
      const lightness = Number.parseInt(match[3], 10)
      return lightness > 55 ? 'dark' : 'light'
    }
  }

  const hex = color.replace('#', '')
  const r = Number.parseInt(hex.substring(0, 2), 16)
  const g = Number.parseInt(hex.substring(2, 4), 16)
  const b = Number.parseInt(hex.substring(4, 6), 16)

  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255

  return luminance > 0.5 ? 'dark' : 'light'
}

const getLabelTextColor = (color?: string | null): string => {
  const contrast = getContrastColor(color)
  return contrast === 'light' ? '#ffffff' : '#111827'
}

const getLabelBorderColor = (color?: string | null): string => {
  const contrast = getContrastColor(color)
  return contrast === 'light'
    ? 'rgba(255, 255, 255, 0.2)'
    : 'rgba(0, 0, 0, 0.1)'
}

const formatDate = (date: string): string => {
  return new Date(date).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}
</script>

<template>
  <section data-component="sets-section" class="space-y-3">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <h2 class="text-copy text-xl font-bold">Sets</h2>
        <span v-if="canShowContent"
          class="bg-brand-100 text-brand-700 inline-flex h-6 min-w-6 items-center justify-center rounded-full px-2 text-xs font-semibold">
          {{ totalSets }}
        </span>
      </div>
      <RouterLink v-if="canShowContent" to="/sets" class="text-brand-500 hover:text-brand-600 text-sm font-medium">
        View All →
      </RouterLink>
    </div>

    <AppQueryState :query="setsQuery" loading-message="Loading sets..."
      empty-message="Nothing sorted yet. Peak entropy achieved. Time to start creating some sets to organize your items!"
      error-fallback="Unable to load sets.">
      <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <RouterLink v-for="set in displaySets" :key="set.id" :to="`/sets`"
          class="border-line-subtle bg-surface-elevated hover:border-brand-300 block rounded-xl border p-4 transition">
          <h3 class="text-copy text-base font-semibold">{{ normalizeTitleWords(set.name) }}</h3>
          <p class="text-copy-muted mt-2 text-sm">
            {{ getItemTypeLabel(set.set_category) }}
            <span class="text-line mx-2">/</span>
            {{ setStatsById[set.id]?.itemCount ?? 0 }} items
            <span class="text-line mx-2">/</span>
            {{ formatDisplayWeight(setStatsById[set.id]?.totalWeightGrams ?? 0, weightUnit) }}
            <span class="text-line mx-2">/</span>
            {{ formatValue(setStatsById[set.id]?.totalValue ?? 0, currency) }}
          </p>

          <div v-if="getSetLabels(set.id).length > 0" class="mt-2 flex flex-wrap gap-1">
            <span v-for="label in getSetLabels(set.id)" :key="label.id"
              class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium" :style="{
                backgroundColor: label.color ?? '#6b7280',
                color: getLabelTextColor(label.color),
                border: `1px solid ${getLabelBorderColor(label.color)}`
              }">
              {{ label.name }}
            </span>
          </div>

          <p class="text-copy-subtle mt-1 text-xs">Updated {{ formatDate(set.updated_at) }}</p>
        </RouterLink>
      </div>
    </AppQueryState>
  </section>
</template>
