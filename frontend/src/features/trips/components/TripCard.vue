<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { AppIcon } from '../../../components/icons'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { useSettings } from '../../../composables/useSettings'
import { formatDisplayWeight, toRoundedString } from '../../../lib/units/conversions'
import { formatValue } from '../../../lib/format/display'
import { getRoutePreview } from '../api/tripsApi'
import {
    ROUTE_SERVICE_ICONS,
    averageDistancePerDay,
    formatDurationDays,
    formatTripType,
    getTripRouteService,
    getTripRouteUrl,
} from '../utils'
import type { Trip, TripStats } from '../types'

const props = defineProps<{
    trip: Trip
    stats?: TripStats
}>()

const emit = defineEmits<{
    edit: [trip: Trip]
    delete: [trip: Trip]
}>()

const { weightUnit, currency } = useSettings()

const routeUrl = computed(() => getTripRouteUrl(props.trip))
const routeService = computed(() => getTripRouteService(props.trip))
const routeIconName = computed(() => ROUTE_SERVICE_ICONS[routeService.value])

// Fetch Open Graph preview metadata only when the trip has a recognized route service.
const routePreviewQuery = useQuery({
    queryKey: computed(() => ['trip-route-preview', routeService.value, routeUrl.value]),
    queryFn: () => getRoutePreview(routeService.value as Exclude<typeof routeService.value, 'unknown'>, routeUrl.value),
    enabled: computed(() => routeUrl.value.length > 0 && routeService.value !== 'unknown'),
    staleTime: 1000 * 60 * 60,
})

const previewImageUrl = computed(() => routePreviewQuery.data.value?.image_url ?? null)

const weight = (grams: number) => formatDisplayWeight(grams, weightUnit.value)
const value = (amount: number) => formatValue(amount, currency.value, toRoundedString)

const durationLabel = computed(() => formatDurationDays(props.trip.duration))

const distanceLabel = computed(() => {
    const distance = props.trip.total_distance_km
    return typeof distance === 'number' ? `${toRoundedString(distance)} km` : 'Not set'
})

const avgPerDayLabel = computed(() => {
    const avg = averageDistancePerDay(props.trip.total_distance_km, props.trip.duration)
    return avg === null ? 'Not set' : `${toRoundedString(avg)} km/day`
})

// Show the creation date until the trip is updated, then show the update date.
const dateLabel = computed(() => {
    const isUpdated = props.trip.updated_at !== props.trip.created_at
    return isUpdated ? `Updated ${formatDate(props.trip.updated_at)}` : `Created ${formatDate(props.trip.created_at)}`
})

function formatDate(value: string): string {
    const parsed = new Date(value)
    if (Number.isNaN(parsed.getTime())) {
        return value
    }
    const day = String(parsed.getDate()).padStart(2, '0')
    const month = String(parsed.getMonth() + 1).padStart(2, '0')
    const year = parsed.getFullYear()
    return `${day}-${month}-${year}`
}
</script>

<template>
    <article class="surface-panel relative flex flex-col overflow-hidden p-0 text-left">
        <button type="button"
            class="text-copy-muted hover:bg-surface-muted hover:text-danger absolute right-2 top-2 z-10 inline-flex h-8 w-8 items-center justify-center rounded-full bg-surface/80 transition"
            aria-label="Delete trip" @click.stop="emit('delete', trip)">
            <AppIcon category="action" name="delete" size="sm" />
        </button>

        <button type="button" class="flex flex-1 flex-col text-left" @click="emit('edit', trip)">
            <!-- Image preview from the route link -->
            <div class="bg-surface-muted relative aspect-video w-full overflow-hidden">
                <img v-if="previewImageUrl" :src="previewImageUrl" :alt="`${trip.name} route preview`"
                    class="h-full w-full object-cover" loading="lazy" />
                <div v-else class="text-copy-subtle flex h-full w-full items-center justify-center">
                    <AppIcon category="content" name="imagePlaceholder" size="2xl" />
                </div>
            </div>

            <div class="flex flex-1 flex-col gap-2 p-4">
                <h3 class="text-ink text-lg font-semibold">{{ normalizeTitleWords(trip.name) }}</h3>

                <p class="text-copy-muted text-sm">
                    {{ formatTripType(trip.trip_type) }}
                    <span class="text-line mx-1.5">/</span>{{ durationLabel }}
                    <span class="text-line mx-1.5">/</span>{{ distanceLabel }}
                    <span class="text-line mx-1.5">/</span>{{ avgPerDayLabel }}
                </p>

                <p class="text-copy-muted text-sm">
                    Total: {{ weight(stats?.totalWeightGrams ?? 0) }}
                    <span class="text-line mx-1.5">/</span>Packed: {{ weight(stats?.packedWeightGrams ?? 0) }}
                    <span class="text-line mx-1.5">/</span>Worn: {{ weight(stats?.wornWeightGrams ?? 0) }}
                    <span class="text-line mx-1.5">/</span>Value: {{ value(stats?.totalValue ?? 0) }}
                </p>

                <p class="text-copy-muted text-sm">
                    Packs: {{ stats?.packsCount ?? 0 }}
                    <span class="text-line mx-1.5">/</span>Travelers: {{ stats?.travelersCount ?? 0 }}
                </p>

                <p v-if="trip.notes" class="text-copy-subtle line-clamp-2 text-sm">{{ trip.notes }}</p>

                <div class="mt-auto flex items-center justify-between gap-2 pt-2">
                    <a v-if="routeUrl" :href="routeUrl" target="_blank" rel="noopener noreferrer"
                        class="text-brand-600 hover:text-brand-700 inline-flex items-center gap-1 text-sm" @click.stop>
                        <AppIcon category="content" :name="routeIconName" size="sm" />
                        <span>Route</span>
                    </a>
                    <span v-else />
                    <span class="text-copy-subtle text-xs">{{ dateLabel }}</span>
                </div>
            </div>
        </button>
    </article>
</template>
