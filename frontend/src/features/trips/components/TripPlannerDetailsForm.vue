<script setup lang="ts">
import { ref, watch } from 'vue'
import AppSelect from '../../../components/forms/AppSelect.vue'
import { getRoutePreview } from '../api/tripsApi'
import { cleanRouteTitle, detectRouteService, ROUTE_SERVICE_OPTIONS, TRIP_TYPE_OPTIONS } from '../utils'
import { useTripPlanner } from '../composables/useTripPlanner'

const planner = useTripPlanner()
const details = planner.details

// Day hikes / overnights have an implied duration, so set it when the type changes.
watch(
    () => details.tripType,
    (tripType) => {
        if (tripType === 'day_hike') {
            details.durationDays = '1'
        }
        if (tripType === 'overnight') {
            details.durationDays = '2'
        }
    },
)

const isFetchingRouteName = ref(false)

const isHttpUrl = (value: string): boolean => /^https?:\/\//i.test(value.trim())

// Pre-select the service from a recognizable host (komoot/strava).
const preselectServiceFromUrl = (): void => {
    const detected = detectRouteService(details.routeUrl)
    if (detected !== 'unknown') {
        details.routeService = detected
    }
}

// Best-effort autofill of the trip name from the route page title when name is empty.
const autofillNameFromRoute = async (): Promise<void> => {
    preselectServiceFromUrl()

    const url = details.routeUrl.trim()
    const service = details.routeService
    if (!url || !isHttpUrl(url) || service === 'unknown' || details.name.trim().length > 0) {
        return
    }

    isFetchingRouteName.value = true
    try {
        const preview = await getRoutePreview(service, url)
        const name = cleanRouteTitle(preview.title)
        if (name && details.name.trim().length === 0) {
            details.name = name
        }
    } catch {
        // Route preview is best-effort; ignore failures.
    } finally {
        isFetchingRouteName.value = false
    }
}
</script>

<template>
    <div class="flex flex-col gap-4">
        <div class="grid gap-3 sm:grid-cols-[160px_1fr]">
            <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Route service</span>
                <AppSelect v-model="details.routeService">
                    <option v-for="option in ROUTE_SERVICE_OPTIONS" :key="option.value" :value="option.value">
                        {{ option.label }}
                    </option>
                </AppSelect>
            </div>
            <label class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Route URL</span>
                <input v-model="details.routeUrl" class="input-shell" type="url"
                    placeholder="https://www.komoot.com/tour/..." @blur="autofillNameFromRoute" />
            </label>
        </div>

        <label class="grid gap-1">
            <span class="text-copy flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.06em]">
                Trip name
                <span v-if="isFetchingRouteName" class="text-copy-subtle normal-case tracking-normal">
                    fetching from route…
                </span>
            </span>
            <input v-model="details.name" class="input-shell" type="text" placeholder="Weekend in the Alps" />
        </label>

        <label class="grid gap-1">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Description</span>
            <textarea v-model="details.notes" class="input-shell min-h-24 resize-y"
                placeholder="Notes about this trip" />
        </label>

        <div class="grid gap-3 sm:grid-cols-3">
            <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Type</span>
                <AppSelect v-model="details.tripType">
                    <option v-for="option in TRIP_TYPE_OPTIONS" :key="option.value" :value="option.value">
                        {{ option.label }}
                    </option>
                </AppSelect>
            </div>

            <label v-if="details.tripType !== 'day_hike' && details.tripType !== 'overnight'" class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Duration (days)</span>
                <input v-model="details.durationDays" class="input-shell" type="number" min="1" step="1"
                    placeholder="3" />
            </label>
            <label class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Distance (km)</span>
                <input v-model="details.distanceKm" class="input-shell" type="number" min="0" step="0.1"
                    placeholder="42.5" />
            </label>
        </div>
    </div>
</template>
