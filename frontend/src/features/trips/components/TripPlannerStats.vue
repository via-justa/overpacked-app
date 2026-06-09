<script setup lang="ts">
import { computed } from 'vue'
import { useSettings } from '../../../composables/useSettings'
import { formatDisplayWeight } from '../../../lib/units/conversions'
import { formatValue } from '../../../lib/format/display'
import { useTripPlanner } from '../composables/useTripPlanner'

const planner = useTripPlanner()
const { weightUnit, currency } = useSettings()

const cards = computed(() => [
    { label: 'Items', value: String(planner.stats.value.itemCount) },
    { label: 'Total weight', value: formatDisplayWeight(planner.stats.value.totalWeightGrams, weightUnit.value) },
    { label: 'Total value', value: formatValue(planner.stats.value.totalValue, currency.value) },
])
</script>

<template>
    <div data-element="trip-planner-stats" class="grid gap-3 sm:grid-cols-3">
        <div v-for="card in cards" :key="card.label" class="surface-panel flex flex-col gap-1 p-4">
            <span class="text-copy-subtle text-xs font-semibold uppercase tracking-[0.06em]">{{ card.label }}</span>
            <span class="text-copy text-xl font-bold">{{ card.value }}</span>
        </div>
    </div>
</template>
