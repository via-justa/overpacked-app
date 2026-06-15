<script setup lang="ts">
import { reactive } from 'vue'
import Button from 'primevue/button'
import AppIcon from '../../../components/icons/AppIcon.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { useTripPlanner } from '../composables/useTripPlanner'
import type { Person } from '../../persons/types'

const planner = useTripPlanner()

// Tracks the new-pack name input per person id.
const newPackName = reactive<Record<string, string>>({})

const onTogglePerson = (person: Person): void => {
    planner.togglePerson(person)
}

const onAddPack = (personId: string): void => {
    planner.addPack(personId, newPackName[personId] ?? '')
    newPackName[personId] = ''
}
</script>

<template>
    <div data-element="trip-planner-people" class="surface-panel flex flex-col gap-3 p-5">
        <h2 class="text-copy text-lg font-semibold">People &amp; packs</h2>

        <p v-if="planner.availablePersons.value.length === 0" class="text-copy-subtle text-sm">
            No persons available. Create persons first to add them to a trip.
        </p>

        <div v-else class="flex flex-col gap-2">
            <!-- Selectable people -->
            <div class="border-line-subtle bg-surface-muted flex flex-wrap gap-2 rounded-xl border p-2">
                <button v-for="person in planner.availablePersons.value" :key="person.id" type="button"
                    class="rounded-full px-3 py-1 text-sm font-medium transition" :class="planner.isPersonSelected(person.id)
                        ? 'bg-surface-inverse text-ink-inverse'
                        : 'bg-surface-soft text-copy hover:bg-surface'" @click="onTogglePerson(person)">
                    {{ normalizeTitleWords(person.name) }}
                </button>
            </div>

            <!-- Pack management per selected person -->
            <div v-for="person in planner.persons.value" :key="person.personId"
                class="border-line-subtle flex flex-col gap-2 rounded-xl border p-3">
                <div class="flex items-center justify-between gap-2">
                    <span class="text-copy text-sm font-semibold">{{ normalizeTitleWords(person.person.name) }}</span>
                    <span class="text-copy-subtle text-xs">{{ person.packs.length }}
                        {{ person.packs.length === 1 ? 'pack' : 'packs' }}</span>
                </div>

                <ul class="flex flex-col gap-1">
                    <li v-for="pack in person.packs" :key="pack.localId"
                        class="bg-surface-muted flex items-center gap-2 rounded-lg px-2 py-1.5">
                        <button type="button" class="shrink-0"
                            :title="person.mainPackLocalId === pack.localId ? 'Main pack' : 'Set as main pack'"
                            :aria-label="person.mainPackLocalId === pack.localId ? 'Main pack' : 'Set as main pack'"
                            @click="planner.setMainPack(person.personId, pack.localId)">
                            <AppIcon category="status"
                                :name="person.mainPackLocalId === pack.localId ? 'active' : 'incomplete'" size="sm"
                                :color="person.mainPackLocalId === pack.localId ? 'text-brand-500' : 'text-copy-subtle'" />
                        </button>
                        <input :value="pack.name" aria-label="Pack name" class="input-shell flex-1 py-1 text-sm"
                            @input="planner.renamePack(person.personId, pack.localId, ($event.target as HTMLInputElement).value)" />
                        <button type="button" class="text-copy-subtle hover:text-danger shrink-0" title="Remove pack"
                            aria-label="Remove pack" @click="planner.removePack(person.personId, pack.localId)">
                            <AppIcon category="action" name="delete" size="sm" />
                        </button>
                    </li>
                </ul>

                <div class="flex items-center gap-2">
                    <input v-model="newPackName[person.personId]" aria-label="New pack name" class="input-shell flex-1 py-1 text-sm"
                        placeholder="New pack name" @keyup.enter="onAddPack(person.personId)" />
                    <Button label="Add pack" size="small" outlined @click="onAddPack(person.personId)" />
                </div>
            </div>
        </div>
    </div>
</template>
