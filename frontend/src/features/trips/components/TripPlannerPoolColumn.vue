<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import TripPlannerItemCard from './TripPlannerItemCard.vue'
import { TRIP_PLANNER_DND_GROUP } from '../plannerTypes'
import { useTripPlanner } from '../composables/useTripPlanner'

const planner = useTripPlanner()
</script>

<template>
    <div data-element="trip-planner-pool-column" class="surface-panel flex h-full flex-col gap-3 p-4">
        <div class="flex items-center justify-between">
            <h2 class="text-copy text-base font-semibold">Unassigned gear</h2>
            <span class="text-copy-subtle text-xs">{{ planner.unassigned.value.length }} cards</span>
        </div>

        <VueDraggable v-model="planner.unassigned.value" :group="TRIP_PLANNER_DND_GROUP" :animation="150"
            class="flex min-h-24 flex-1 flex-col gap-1.5">
            <TripPlannerItemCard v-for="placement in planner.unassigned.value" :key="placement.localId"
                :placement="placement" />
        </VueDraggable>

        <p v-if="planner.unassigned.value.length === 0" class="text-copy-subtle py-6 text-center text-sm">
            All gear assigned. Drag cards here to unassign them.
        </p>
    </div>
</template>
