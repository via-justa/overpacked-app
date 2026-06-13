<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import AppConfirmDialog from '../../../components/dialogs/AppConfirmDialog.vue'
import AppQueryError from '../../../components/feedback/AppQueryError.vue'
import AppLoadingState from '../../../components/feedback/AppLoadingState.vue'
import AppEmptyState from '../../../components/feedback/AppEmptyState.vue'
import TripsCollectionView from '../components/TripsCollectionView.vue'
import { useDeleteConfirmation } from '../../../composables/useDeleteConfirmation'
import { useMutationWithToast } from '../../../composables/useMutationWithToast'
import { getTrip, listTrips, removeTrip } from '../api/tripsApi'
import { computeTripStats } from '../utils'
import type { Trip, TripStats } from '../types'

const router = useRouter()

const tripsQuery = useQuery({
    queryKey: ['trips'],
    queryFn: listTrips,
})

const allTrips = computed<Trip[]>(() => tripsQuery.data.value ?? [])

// Trips are sorted by most recently updated first (future: pluggable sort options).
const sortedTrips = computed<Trip[]>(() =>
    [...allTrips.value].sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()),
)

const tripIdsKey = computed(() => allTrips.value.map((trip) => trip.id).sort().join(','))

// Fetch nested details for every trip so cards can show aggregated stats.
const tripDetailsQuery = useQuery({
    queryKey: computed(() => ['trips-details', tripIdsKey.value]),
    queryFn: async () => {
        const trips = allTrips.value
        if (trips.length === 0) {
            return []
        }
        return Promise.all(trips.map((trip) => getTrip(trip.id)))
    },
    enabled: computed(() => allTrips.value.length > 0),
})

const statsByTripId = computed<Record<string, TripStats>>(() => {
    const map: Record<string, TripStats> = {}
    for (const detail of tripDetailsQuery.data.value ?? []) {
        map[detail.id] = computeTripStats(detail)
    }
    return map
})

const canShowEmptyState = computed(
    () => !tripsQuery.isPending.value && !tripsQuery.isError.value && sortedTrips.value.length === 0,
)

// ─── Create / edit navigation ────────────────────────────────────────────────

const openCreateDialog = () => {
    router.push('/trips/new')
}

const onStartEdit = (trip: Trip) => {
    router.push(`/trips/${trip.id}/edit`)
}

// ─── Delete ──────────────────────────────────────────────────────────────────

const { confirmState, requestSingleDelete, closeConfirm, getConfirmMessage } = useDeleteConfirmation()

const deleteMutation = useMutationWithToast({
    mutationFn: (tripId: string) => removeTrip(tripId),
    successMessage: { summary: 'Trip deleted', detail: 'The trip was removed.' },
    errorMessage: { summary: 'Delete failed', detail: 'Unable to delete trip.' },
    invalidateQueries: [['trips'], ['trips-details']],
    onSuccess: () => closeConfirm(),
})

const onRequestDelete = (trip: Trip) => {
    requestSingleDelete(trip.id, trip.name)
}

const onConfirmDelete = () => {
    if (confirmState.value?.kind === 'single') {
        deleteMutation.mutate(confirmState.value.id)
    }
}
</script>

<template>
    <section data-component="trips-page" class="flex w-full flex-col gap-4">
        <div class="hidden items-center justify-between md:flex">
            <h1 class="text-copy text-2xl font-bold">Trips</h1>
        </div>

        <AppConfirmDialog :open="confirmState !== null" title="Delete trip" :message="getConfirmMessage()"
            confirm-label="Delete" confirm-tone="danger" @update:open="(value) => { if (!value) closeConfirm() }"
            @cancel="closeConfirm" @confirm="onConfirmDelete" />

        <AppQueryError :query="tripsQuery" fallback-message="Unable to load trips." data-element="trips-error" />

        <AppLoadingState v-if="tripsQuery.isPending.value" message="Loading trips..." data-element="trips-loading" />

        <div v-else-if="canShowEmptyState" class="flex flex-col items-start gap-3">
            <AppEmptyState message="No trips yet. Plan your first adventure and start packing!"
                data-element="trips-empty-state" />
            <button type="button"
                class="bg-brand-600 text-ink-inverse hover:bg-brand-700 rounded-lg px-4 py-2 text-sm font-medium transition"
                @click="openCreateDialog">
                Create your first trip
            </button>
        </div>

        <TripsCollectionView v-else :trips="sortedTrips" :stats-by-trip-id="statsByTripId" @edit="onStartEdit"
            @delete="onRequestDelete" />
    </section>
</template>
