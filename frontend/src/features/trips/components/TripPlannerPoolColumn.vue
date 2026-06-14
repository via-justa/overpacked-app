<script setup lang="ts">
import { computed, ref } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import AppIcon from '../../../components/icons/AppIcon.vue'
import TripPlannerItemCard from './TripPlannerItemCard.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { TRIP_PLANNER_DND_GROUP } from '../plannerTypes'
import { useTripPlanner } from '../composables/useTripPlanner'

const planner = useTripPlanner()

const isPersonMenuOpen = ref(false)
const selectedIds = ref<Set<string>>(new Set())

// Only count selections that still exist in the pool (items may be dragged out).
const selectedCount = computed(
    () => planner.unassigned.value.filter((placement) => selectedIds.value.has(placement.localId)).length,
)

const canAssign = computed(
    () => planner.persons.value.length > 0 && planner.unassigned.value.length > 0,
)

const assignTooltip = computed(() =>
    selectedCount.value > 0 ? 'Assign selected gear to a person' : 'Assign all gear to a person',
)

const toggleSelect = (localId: string, value: boolean): void => {
    const next = new Set(selectedIds.value)
    if (value) {
        next.add(localId)
    } else {
        next.delete(localId)
    }
    selectedIds.value = next
}

const assignTo = (personId: string): void => {
    // Selected items if any were ticked, otherwise the whole pool.
    const targets =
        selectedCount.value > 0
            ? planner.unassigned.value.filter((placement) => selectedIds.value.has(placement.localId))
            : undefined
    planner.assignToPerson(personId, targets)
    selectedIds.value = new Set()
    isPersonMenuOpen.value = false
}

// One person: assign straight away. Multiple: open the picker to choose who.
const onAssignClick = (): void => {
    const persons = planner.persons.value
    if (persons.length === 1) {
        assignTo(persons[0].personId)
        return
    }
    isPersonMenuOpen.value = !isPersonMenuOpen.value
}
</script>

<template>
    <div data-element="trip-planner-pool-column" class="surface-panel flex h-full flex-col gap-3 p-4">
        <div class="flex items-center justify-between">
            <h2 class="text-copy text-base font-semibold">Unassigned gear</h2>
            <div class="flex items-center gap-2">
                <span class="text-copy-subtle text-xs">
                    {{ selectedCount > 0 ? `${selectedCount} selected` : `${planner.unassigned.value.length} items` }}
                </span>
                <div v-if="canAssign" class="relative">
                    <button type="button"
                        class="text-copy-subtle hover:text-copy hover:bg-surface-soft flex h-6 w-6 items-center justify-center rounded-full transition"
                        v-tooltip.top="assignTooltip" @click="onAssignClick">
                        <AppIcon category="directional" name="chevronRight" size="xs" />
                    </button>

                    <template v-if="isPersonMenuOpen">
                        <div class="fixed inset-0 z-40" @click="isPersonMenuOpen = false" />
                        <div
                            class="border-line-subtle bg-surface-elevated shadow-panel absolute right-0 z-50 mt-1 flex min-w-40 flex-col rounded-lg border py-1">
                            <button v-for="entry in planner.persons.value" :key="entry.personId" type="button"
                                class="text-copy hover:bg-surface-soft px-3 py-1.5 text-left text-sm transition"
                                @click="assignTo(entry.personId)">
                                {{ normalizeTitleWords(entry.person.name) }}
                            </button>
                        </div>
                    </template>
                </div>
            </div>
        </div>

        <VueDraggable v-model="planner.unassigned.value" :group="TRIP_PLANNER_DND_GROUP" :animation="150"
            class="flex min-h-24 flex-1 flex-col gap-1.5">
            <TripPlannerItemCard v-for="placement in planner.unassigned.value" :key="placement.localId"
                :placement="placement" selectable :selected="selectedIds.has(placement.localId)"
                @update:selected="(value) => toggleSelect(placement.localId, value)" />
        </VueDraggable>

        <p v-if="planner.unassigned.value.length === 0" class="text-copy-subtle py-6 text-center text-sm">
            All gear assigned. Drag items here to unassign them.
        </p>
    </div>
</template>
