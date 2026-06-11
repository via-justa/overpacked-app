<script setup lang="ts">
import AppSelect from '../../../components/forms/AppSelect.vue'
import AppIcon from '../../../components/icons/AppIcon.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { useTripPlanner } from '../composables/useTripPlanner'

const planner = useTripPlanner()

const packingListTooltip =
    'Pick a packing list to use as a checklist to ensure you have all necessary items for the trip. Click a label to filter the gear pool to just those items.'

// Toggle the pool filter for a label (clicking the active one clears it).
const onLabelClick = (labelId: string): void => {
    planner.filterLabelId.value = planner.filterLabelId.value === labelId ? null : labelId
}
</script>

<template>
    <aside data-element="trip-planner-labels" class="surface-panel flex h-fit flex-col gap-3 p-5">
        <h2 class="text-copy flex items-center justify-between gap-1.5 text-lg font-semibold">
            Packing list
            <AppIcon category="feedback" name="info" size="xs" class="text-copy-subtle cursor-help"
                v-tooltip.top="packingListTooltip" />
        </h2>

        <AppSelect v-model="planner.selectedPackingListId.value">
            <option value="">No packing list</option>
            <option v-for="list in planner.packingLists.value" :key="list.id" :value="list.id">
                {{ normalizeTitleWords(list.name) }}
            </option>
        </AppSelect>

        <p v-if="!planner.selectedPackingListId.value" class="text-copy-subtle text-sm">
            {{ packingListTooltip }}
        </p>

        <p v-else-if="planner.packingListLabels.value.length === 0" class="text-copy-subtle text-sm">
            This packing list has no labels.
        </p>

        <ul v-else class="flex flex-col gap-1">
            <li v-for="label in planner.packingListLabels.value" :key="label.id">
                <button type="button"
                    class="flex w-full items-center justify-between gap-2 rounded-lg px-2 py-1.5 text-left text-sm transition"
                    :class="[
                        planner.coveredLabelIds.value.has(label.id) ? 'text-copy-subtle line-through' : 'text-copy',
                        planner.filterLabelId.value === label.id ? 'bg-surface-inverse text-ink-inverse' : 'hover:bg-surface-muted',
                    ]" @click="onLabelClick(label.id)">
                    <span class="truncate">{{ label.name }}</span>
                    <span v-if="planner.coveredLabelIds.value.has(label.id)" class="text-xs">added</span>
                </button>
            </li>
        </ul>
    </aside>
</template>
