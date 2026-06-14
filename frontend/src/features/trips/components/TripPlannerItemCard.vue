<script setup lang="ts">
import { computed } from 'vue'
import AppIcon from '../../../components/icons/AppIcon.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { useSettings } from '../../../composables/useSettings'
import { formatDisplayWeight } from '../../../lib/units/conversions'
import { useTripPlanner } from '../composables/useTripPlanner'
import type { PlannerPlacement } from '../plannerTypes'

const props = withDefaults(defineProps<{
    placement: PlannerPlacement
    selectable?: boolean
    selected?: boolean
}>(), {
    selectable: false,
    selected: false,
})

const emit = defineEmits<{
    'update:selected': [value: boolean]
}>()

const planner = useTripPlanner()
const { weightUnit } = useSettings()

const lineWeight = computed(() =>
    formatDisplayWeight((props.placement.item.weight_grams ?? 0) * props.placement.quantity, weightUnit.value),
)

const decrement = (): void => {
    planner.setPlacementQuantity(props.placement, props.placement.quantity - 1)
}

const increment = (): void => {
    planner.setPlacementQuantity(props.placement, props.placement.quantity + 1)
}
</script>

<template>
    <div data-element="trip-planner-card"
        class="border-line-subtle bg-surface-elevated flex cursor-grab items-center gap-2 rounded-lg border px-2 py-1.5 active:cursor-grabbing">
        <input v-if="selectable" type="checkbox" class="shrink-0 cursor-pointer" :checked="selected"
            aria-label="Select item" @click.stop @change="emit('update:selected', ($event.target as HTMLInputElement).checked)" />
        <AppIcon v-if="!selectable" category="action" name="menuBars" size="xs" class="text-copy-subtle shrink-0" />
        <div class="min-w-0 flex-1">
            <p class="text-copy truncate text-sm font-medium">{{ normalizeTitleWords(placement.item.name) }}</p>
            <p class="text-copy-subtle text-xs">{{ lineWeight }}</p>
        </div>

        <div class="flex items-center gap-1">
            <button type="button" class="text-copy-subtle hover:text-copy" title="Decrease" aria-label="Decrease quantity"
                @click="decrement">
                <AppIcon category="directional" name="chevronDown" size="xs" />
            </button>
            <span class="text-copy w-5 text-center text-sm tabular-nums">{{ placement.quantity }}</span>
            <button type="button" class="text-copy-subtle hover:text-copy" title="Increase" aria-label="Increase quantity"
                @click="increment">
                <AppIcon category="directional" name="chevronUp" size="xs" />
            </button>
        </div>

        <button v-if="placement.quantity > 1" type="button" class="text-copy-subtle hover:text-copy shrink-0"
            title="Split one off" aria-label="Split one off" @click="planner.splitPlacement(placement)">
            <AppIcon category="action" name="duplicate" size="xs" />
        </button>
        <button type="button" class="text-copy-subtle hover:text-danger shrink-0" title="Return to pool"
            aria-label="Return to pool" @click="planner.removePlacement(placement)">
            <AppIcon category="action" name="close" size="xs" />
        </button>
    </div>
</template>
