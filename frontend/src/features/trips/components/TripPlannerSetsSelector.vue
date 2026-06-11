<script setup lang="ts">
import { computed, ref } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import Button from 'primevue/button'
import AppMultiSelect from '../../../components/forms/AppMultiSelect.vue'
import AppIcon from '../../../components/icons/AppIcon.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { listSets } from '../../sets/api/setsApi'
import { useTripPlanner } from '../composables/useTripPlanner'

const planner = useTripPlanner()

const setsQuery = useQuery({ queryKey: ['sets'], queryFn: listSets })
const sets = computed(() => setsQuery.data.value ?? [])
const selectedSetIds = ref<string[]>([])

const onAddSets = async (): Promise<void> => {
    for (const setId of selectedSetIds.value) {
        await planner.addSet(setId)
    }
    selectedSetIds.value = []
}
</script>

<template>
    <div data-element="trip-planner-sets" class="surface-panel flex flex-col gap-2 p-4">
        <h2 class="text-copy flex items-center justify-between gap-1.5 text-lg font-semibold">
            Sets
            <AppIcon category="feedback" name="info" size="xs" class="text-copy-subtle cursor-help"
                v-tooltip.top="'Sets are reusable bundles of gear. Add one to drop all of its items into your trip at once.'" />
        </h2>
        <div class="grid gap-2 sm:grid-cols-[1fr_auto]">
            <AppMultiSelect v-model="selectedSetIds" placeholder="Add sets…" :max-selected-labels="0"
                selected-items-label="{0} sets selected">
                <option v-for="set in sets" :key="set.id" :value="set.id">{{ normalizeTitleWords(set.name) }}</option>
            </AppMultiSelect>
            <Button label="Add" size="small" outlined :disabled="selectedSetIds.length === 0" @click="onAddSets" />
        </div>
    </div>
</template>
