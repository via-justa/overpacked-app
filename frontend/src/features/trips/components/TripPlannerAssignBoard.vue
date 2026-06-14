<script setup lang="ts">
import TripPlannerPoolColumn from './TripPlannerPoolColumn.vue'
import TripPlannerPersonColumn from './TripPlannerPersonColumn.vue'
import { useTripPlanner } from '../composables/useTripPlanner'

const planner = useTripPlanner()
</script>

<template>
    <div data-element="trip-planner-board" class="grid gap-4 lg:grid-cols-2">
        <TripPlannerPoolColumn />

        <div class="flex flex-col gap-4">
            <TripPlannerPersonColumn v-for="person in planner.persons.value" :key="person.personId" :person="person" />
        </div>
    </div>
</template>
