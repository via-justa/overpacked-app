<script setup lang="ts">
import AppSelect from '../../../components/forms/AppSelect.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { useTripPlanner } from '../composables/useTripPlanner'

const planner = useTripPlanner()

// Toggle the pool filter for a label (clicking the active one clears it).
const onLabelClick = (labelId: string): void => {
    planner.filterLabelId.value = planner.filterLabelId.value === labelId ? null : labelId
}
</script>

<template>
    <aside data-element="trip-planner-labels" class="surface-panel flex h-fit flex-col gap-3 p-5">
        <h2 class="text-copy text-lg font-semibold">Packing list</h2>

        <AppSelect v-model="planner.selectedPackingListId.value">
            <option value="">No packing list</option>
            <option v-for="list in planner.packingLists.value" :key="list.id" :value="list.id">
                {{ normalizeTitleWords(list.name) }}
            </option>
        </AppSelect>

        <p v-if="!planner.selectedPackingListId.value" class="text-copy-subtle text-sm">
            Select a packing list to see its labels. Added gear grays out matching labels.
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
