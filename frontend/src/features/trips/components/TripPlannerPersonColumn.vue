<script setup lang="ts">
import { computed, ref } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import AppIcon from '../../../components/icons/AppIcon.vue'
import TripPlannerItemCard from './TripPlannerItemCard.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { useSettings } from '../../../composables/useSettings'
import { formatDisplayWeight } from '../../../lib/units/conversions'
import { formatValue } from '../../../lib/format/display'
import { TRIP_PLANNER_DND_GROUP } from '../plannerTypes'
import type { PlannerPerson, PlannerPlacement } from '../plannerTypes'
import { useTripPlanner } from '../composables/useTripPlanner'

const props = defineProps<{
    person: PlannerPerson
}>()

const planner = useTripPlanner()
const { weightUnit, currency } = useSettings()

const stats = computed(() => planner.personStats(props.person))

// Main pack is shown first, remaining packs follow in their existing order.
const orderedPacks = computed(() => {
    const main = props.person.packs.find((pack) => pack.localId === props.person.mainPackLocalId)
    const rest = props.person.packs.filter((pack) => pack.localId !== props.person.mainPackLocalId)
    return main ? [main, ...rest] : rest
})

const WEIGHT_STATUS_CLASS: Record<string, string> = {
    ok: 'bg-brand-50 text-brand-700',
    warn: 'bg-warning-200 text-warning-900',
    over: 'bg-danger-500/10 text-danger-500',
    unknown: 'bg-surface-muted text-copy-subtle',
}

const weightChipClass = computed(() => WEIGHT_STATUS_CLASS[stats.value.weightStatus])

const statChips = computed(() => [
    { label: 'Packed', value: formatDisplayWeight(stats.value.packedWeightGrams, weightUnit.value), class: weightChipClass.value },
    { label: 'Worn', value: formatDisplayWeight(stats.value.wornWeightGrams, weightUnit.value) },
    { label: 'Recommended max', value: formatDisplayWeight(stats.value.recommendedMaxGrams, weightUnit.value) },
    { label: 'Gear cost', value: formatValue(stats.value.totalValue, currency.value) },
    { label: 'Items', value: String(stats.value.itemCount) },
])

// Empty inbox that powers the person-level auto-assign drop zone.
const autoInbox = ref<PlannerPlacement[]>([])

const onAutoDrop = (): void => {
    const placement = autoInbox.value[0]
    autoInbox.value = []
    if (placement) {
        planner.autoAssignToPerson(props.person.personId, placement)
    }
}
</script>

<template>
    <section data-element="trip-planner-person" class="surface-panel flex flex-col gap-3 p-4">
        <header class="flex items-center justify-between gap-2">
            <h3 class="text-copy text-base font-semibold">{{ normalizeTitleWords(person.person.name) }}</h3>
        </header>

        <div class="flex flex-wrap gap-1.5">
            <span v-for="chip in statChips" :key="chip.label"
                class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs"
                :class="chip.class ?? 'bg-surface-muted'">
                <span class="text-copy-subtle font-semibold uppercase tracking-wider">{{ chip.label }}</span>
                <span class="text-copy font-semibold">{{ chip.value }}</span>
            </span>
        </div>

        <!-- Person-level auto-assign target -->
        <VueDraggable v-model="autoInbox" :group="TRIP_PLANNER_DND_GROUP" :animation="150" :on-add="onAutoDrop"
            class="border-line-subtle text-copy-subtle flex items-center justify-center gap-2 rounded-lg border border-dashed px-3 py-3 text-xs">
            <AppIcon category="directional" name="arrowDown" size="xs" />
            <span>Drop here to auto-assign (worn / main pack)</span>
        </VueDraggable>

        <!-- Worn -->
        <div class="flex flex-col gap-1.5">
            <span class="text-copy-subtle text-xs font-semibold uppercase tracking-[0.06em]">Worn</span>
            <VueDraggable v-model="person.worn" :group="TRIP_PLANNER_DND_GROUP" :animation="150"
                class="bg-surface-muted flex min-h-12 flex-col gap-1.5 rounded-lg p-1.5">
                <TripPlannerItemCard v-for="placement in person.worn" :key="placement.localId" :placement="placement" />
            </VueDraggable>
        </div>

        <!-- Packs (main first) -->
        <div v-for="pack in orderedPacks" :key="pack.localId" class="flex flex-col gap-1.5">
            <span class="text-copy-subtle flex items-center gap-1.5 text-xs font-semibold uppercase tracking-[0.06em]">
                <AppIcon v-if="pack.localId === person.mainPackLocalId" category="navigation" name="packs" size="xs"
                    color="text-brand-500" />
                {{ normalizeTitleWords(pack.name) }}
            </span>
            <VueDraggable v-model="pack.items" :group="TRIP_PLANNER_DND_GROUP" :animation="150"
                class="bg-surface-muted flex min-h-12 flex-col gap-1.5 rounded-lg p-1.5">
                <TripPlannerItemCard v-for="placement in pack.items" :key="placement.localId" :placement="placement" />
            </VueDraggable>
        </div>
    </section>
</template>
