<script setup lang="ts">
import TripCard from './TripCard.vue'
import type { Trip, TripStats } from '../types'

defineProps<{
    trips: Trip[]
    statsByTripId: Record<string, TripStats>
}>()

const emit = defineEmits<{
    edit: [trip: Trip]
    delete: [trip: Trip]
}>()
</script>

<template>
    <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <TripCard v-for="trip in trips" :key="trip.id" :trip="trip" :stats="statsByTripId[trip.id]"
            @edit="emit('edit', $event)" @delete="emit('delete', $event)" />
    </div>
</template>
