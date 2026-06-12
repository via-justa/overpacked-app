<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppActionButton from '../../../components/actions/AppActionButton.vue'
import AppActionCluster from '../../../components/actions/AppActionCluster.vue'
import AppLoadingState from '../../../components/feedback/AppLoadingState.vue'
import TripPlannerDetailsForm from '../components/TripPlannerDetailsForm.vue'
import TripPlannerStats from '../components/TripPlannerStats.vue'
import TripPlannerPeoplePicker from '../components/TripPlannerPeoplePicker.vue'
import TripPlannerItemPool from '../components/TripPlannerItemPool.vue'
import TripPlannerSetsSelector from '../components/TripPlannerSetsSelector.vue'
import TripPlannerListLabelsPanel from '../components/TripPlannerListLabelsPanel.vue'
import TripPlannerAssignBoard from '../components/TripPlannerAssignBoard.vue'
import { useMutationWithToast } from '../../../composables/useMutationWithToast'
import { provideTripPlanner } from '../composables/useTripPlanner'
import { persistNewTrip, persistTripUpdate } from '../api/tripPersistence'

const props = defineProps<{
    tripId?: string
}>()

const router = useRouter()
const planner = provideTripPlanner()

onMounted(() => {
    if (props.tripId) {
        void planner.loadExisting(props.tripId)
    }
})

const pageTitle = computed(() => (planner.isEditMode.value ? 'Edit Trip' : 'Plan Trip'))

const saveMutation = useMutationWithToast({
    mutationFn: () => {
        const staged = planner.buildStagedTrip()
        const original = planner.originalDetails.value
        return original ? persistTripUpdate(original, staged) : persistNewTrip(staged)
    },
    successMessage: { summary: 'Trip saved', detail: 'Your trip was saved.' },
    errorMessage: { summary: 'Save failed', detail: 'Unable to save trip.' },
    invalidateQueries: [['trips'], ['trips-details']],
    onSuccess: () => {
        void router.push('/trips')
    },
})

const goToStep2 = () => {
    if (planner.canProceedToStep2.value) {
        planner.step.value = 2
    }
}

const goToStep1 = () => {
    planner.step.value = 1
}

const onCancel = () => {
    void router.push('/trips')
}

const onSave = () => {
    if (planner.canSave.value) {
        saveMutation.mutate()
    }
}
</script>

<template>
    <section data-component="trip-planner-page" class="relative flex w-full flex-col gap-4">
        <AppActionCluster data-element="trip-planner-actions">
            <AppActionButton action="cancel" label="Cancel" data-element="trip-planner-cancel" @click="onCancel" />
            <AppActionButton v-if="planner.step.value === 2" action="save" label="Save trip"
                data-element="trip-planner-save" :loading="saveMutation.isPending.value"
                :disabled="!planner.canSave.value" @click="onSave" />
        </AppActionCluster>

        <header class="flex flex-col gap-1">
            <div class="flex items-center justify-between gap-2 pr-24">
                <h1 class="text-copy text-2xl font-bold">{{ pageTitle }}</h1>
                <div class="flex items-center gap-2">
                    <button v-if="planner.step.value === 2" type="button"
                        class="text-brand-500 hover:text-brand-600 text-sm font-medium" @click="goToStep1">
                        ←
                    </button>
                    <span class="text-copy-subtle text-sm">Step {{ planner.step.value }} of 2</span>
                    <button v-if="planner.step.value === 1" type="button"
                        class="text-brand-500 hover:text-brand-600 text-sm font-medium disabled:opacity-40"
                        :disabled="!planner.canProceedToStep2.value" @click="goToStep2">
                        →
                    </button>
                </div>
            </div>
            <p class="text-copy-subtle text-sm">
                {{ planner.step.value === 1 ? 'Trip details, people, and gear' :
                    'Distribute gear across people and packs' }}
            </p>
        </header>

        <AppLoadingState v-if="planner.isLoading.value" message="Loading trip..." />

        <!-- ─── Step 1: details + people + gear ─────────────────────────────── -->
        <div v-else-if="planner.step.value === 1" class="flex flex-col gap-4">
            <div class="surface-panel flex flex-col gap-4 p-5">
                <h2 class="text-copy text-lg font-semibold">Trip details</h2>
                <TripPlannerDetailsForm />
            </div>

            <div class="grid gap-4 lg:grid-cols-[1fr_18rem]">
                <div class="flex flex-col gap-4">
                    <TripPlannerPeoplePicker />
                    <TripPlannerItemPool />
                </div>
                <div class="flex flex-col gap-4">
                    <TripPlannerStats />
                    <TripPlannerSetsSelector />
                    <TripPlannerListLabelsPanel />
                </div>
            </div>
        </div>

        <!-- ─── Step 2: drag-and-drop assignment ────────────────────────────── -->
        <div v-else class="flex flex-col gap-4">
            <div class="surface-panel flex flex-col gap-1 p-5">
                <h2 class="text-copy text-lg font-semibold">{{ planner.details.name || 'Untitled trip' }}</h2>
                <p class="text-copy-subtle text-sm">
                    {{ planner.persons.value.length }} {{ planner.persons.value.length === 1 ? 'person' : 'people' }}
                </p>
            </div>

            <TripPlannerAssignBoard />
        </div>
    </section>
</template>
