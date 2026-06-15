<script setup lang="ts">
import { computed } from 'vue'
import AppIcon from '../../../components/icons/AppIcon.vue'
import TripItemPicker from './TripItemPicker.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { useTripPlanner } from '../composables/useTripPlanner'
import type { CarryStatus } from '../types'
import type { PlannerPlacement } from '../plannerTypes'

const planner = useTripPlanner()

// Items filtered by the label clicked in the packing-list panel.
const filteredItems = computed(() => {
    const labelId = planner.filterLabelId.value
    if (!labelId) {
        return planner.availableItems.value
    }
    return planner.availableItems.value.filter((item) =>
        (planner.itemLabelsByItemId.value.get(item.id) ?? []).some((label) => label.id === labelId),
    )
})

const activeFilterLabel = computed(() => {
    const labelId = planner.filterLabelId.value
    if (!labelId) {
        return null
    }
    return planner.packingListLabels.value.find((label) => label.id === labelId) ?? null
})

const onAddItem = (payload: { itemId: string; quantity: number; carryStatus: CarryStatus }): void => {
    planner.addPoolItem(payload.itemId, payload.quantity, payload.carryStatus)
}

const onQuantityInput = (placement: PlannerPlacement, event: Event): void => {
    const value = Number.parseInt((event.target as HTMLInputElement).value, 10)
    planner.setPlacementQuantity(placement, Number.isFinite(value) ? value : 1)
}

const toggleCarry = (placement: PlannerPlacement): void => {
    placement.carryStatus = placement.carryStatus === 'worn' ? 'packed' : 'worn'
}
</script>

<template>
    <div data-element="trip-planner-pool" class="surface-panel flex flex-col gap-3 p-5">
        <div class="flex items-center justify-between gap-2">
            <h2 class="text-copy text-lg font-semibold">Gear</h2>
            <button v-if="activeFilterLabel" type="button"
                class="bg-surface-soft text-copy hover:bg-surface-muted flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium"
                @click="planner.filterLabelId.value = null">
                <span>Filter: {{ activeFilterLabel.name }}</span>
                <AppIcon category="action" name="close" size="xs" />
            </button>
        </div>

        <TripItemPicker :items="filteredItems" @add="onAddItem" />

        <p v-if="planner.unassigned.value.length === 0" class="text-copy-subtle text-sm">
            No gear added yet. Use the pickers above to build your gear list.
        </p>

        <ul v-else class="flex flex-col gap-1">
            <li v-for="placement in planner.unassigned.value" :key="placement.localId"
                class="bg-surface-muted flex items-center gap-2 rounded-lg px-2 py-1.5">
                <span class="text-copy flex-1 truncate text-sm">{{ normalizeTitleWords(placement.item.name) }}</span>
                <button type="button" class="rounded-full px-2 py-0.5 text-xs font-medium" :class="placement.carryStatus === 'worn'
                    ? 'bg-surface-inverse text-ink-inverse'
                    : 'bg-surface-soft text-copy'" @click="toggleCarry(placement)">
                    {{ placement.carryStatus === 'worn' ? 'Worn' : 'Packed' }}
                </button>
                <input :value="placement.quantity" aria-label="Quantity" type="number" min="1" step="1" class="input-shell w-16 py-1 text-sm"
                    @input="onQuantityInput(placement, $event)" />
                <button type="button" class="text-copy-subtle hover:text-danger shrink-0" title="Remove"
                    aria-label="Remove from pool" @click="planner.removePlacement(placement)">
                    <AppIcon category="action" name="delete" size="sm" />
                </button>
            </li>
        </ul>
    </div>
</template>
