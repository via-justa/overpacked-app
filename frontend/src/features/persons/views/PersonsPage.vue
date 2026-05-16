<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { useMutationWithToast } from '../../../composables/useMutationWithToast'
import AppQueryError from '../../../components/AppQueryError.vue'
import AppLoadingState from '../../../components/AppLoadingState.vue'
import AppEmptyState from '../../../components/AppEmptyState.vue'
import {
  createPerson,
  listPersons,
  removePerson,
  updatePerson,
} from '../api/personsApi'
import PersonFormDialog from '../components/PersonFormDialog.vue'
import type { Person, PersonCreate, PersonFormValues, PersonUpdate } from '../types'

import { getPersonRecommendedMaxWeightGrams } from '../utils'
import { useSettings } from '../../../composables/useSettings'
import { GRAMS_PER_KILOGRAM, LB_PER_KG } from '../../../lib/units/conversions'

const route = useRoute()
const router = useRouter()

type BodyWeightInputUnit = 'kg' | 'lb'

const { weightUnit } = useSettings()
const defaultInputUnit = computed<BodyWeightInputUnit>(() => (weightUnit.value === 'oz' ? 'lb' : 'kg'))
const inputWeightLabel = computed<'kg' | 'lb'>(() => (defaultInputUnit.value === 'lb' ? 'lb' : 'kg'))

// Backend always stores and returns grams. These functions convert between
// the user-facing input unit (kg or lb) and grams.
const gramsToInput = (grams: number, inputUnit: BodyWeightInputUnit): number => {
  const kg = grams / GRAMS_PER_KILOGRAM
  return inputUnit === 'kg' ? kg : kg * LB_PER_KG
}

const inputToGrams = (value: number, inputUnit: BodyWeightInputUnit): number => {
  const kg = inputUnit === 'kg' ? value : value / LB_PER_KG
  return Math.round(kg * GRAMS_PER_KILOGRAM)
}

const buildRange = (start: number, end: number, step: number): number[] => {
  const options: number[] = []
  for (let value = start; value <= end; value += step) {
    options.push(value)
  }
  return options
}

const weightOptions = computed<number[]>(() => {
  if (defaultInputUnit.value === 'lb') {
    return buildRange(66, 440, 1)
  }

  return buildRange(30, 200, 1)
})

const emptyFormValues = (): PersonFormValues => ({
  name: '',
  gender: 'male',
  birthdate: '',
  body_weight_value: '',
  conditioning_level: 'average',
})

const toFormValues = (person: Person): PersonFormValues => ({
  name: person.name,
  gender: person.gender ?? 'male',
  birthdate: person.birthdate ?? '',
  conditioning_level: person.conditioning_level ?? 'average',
  body_weight_value:
    typeof person.body_weight_grams === 'number'
      ? String(Math.round(gramsToInput(person.body_weight_grams, defaultInputUnit.value)))
      : ''
})

const toPayload = (
  values: PersonFormValues,
): PersonCreate | PersonUpdate => {
  const payload: PersonCreate | PersonUpdate = {
    name: normalizeTitleWords(values.name),
    gender: values.gender,
  }

  if (values.birthdate) {
    payload.birthdate = values.birthdate
  }

  payload.conditioning_level = values.conditioning_level

  if (values.body_weight_value.trim()) {
    const parsedInput = Number(values.body_weight_value)
    if (!Number.isNaN(parsedInput)) {
      payload.body_weight_grams = inputToGrams(parsedInput, defaultInputUnit.value)
    }
  }

  return payload
}

const formatAge = (birthdate?: string | null): string => {
  if (!birthdate) {
    return 'Not set'
  }

  const parsed = new Date(birthdate)
  if (Number.isNaN(parsed.getTime())) {
    return 'Not set'
  }

  const today = new Date()
  let age = today.getFullYear() - parsed.getFullYear()
  const monthDiff = today.getMonth() - parsed.getMonth()
  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < parsed.getDate())) {
    age -= 1
  }

  return age >= 0 ? String(age) : 'Not set'
}

const formatGender = (value?: string | null) => {
  if (!value) {
    return 'Not set'
  }

  return value.charAt(0).toUpperCase() + value.slice(1)
}

const formatWeight = (value?: number | null) => {
  if (typeof value !== 'number') {
    return 'Not set'
  }

  return `${gramsToInput(value, defaultInputUnit.value).toFixed(1)} ${inputWeightLabel.value}`
}

const formatRecommendedMaxWeight = (person: Person): string => {
  const recommendedGrams = getPersonRecommendedMaxWeightGrams(person)
  if (recommendedGrams <= 0) {
    return 'Not set'
  }

  return `${gramsToInput(recommendedGrams, defaultInputUnit.value).toFixed(1)} ${inputWeightLabel.value}`
}

const personsQuery = useQuery({
  queryKey: ['persons'],
  queryFn: listPersons,
})

const createValues = ref<PersonFormValues>(emptyFormValues())
const editingPersonId = ref<string | null>(null)
const editValues = ref<PersonFormValues>(emptyFormValues())
const isFormDialogOpen = ref(false)

const createMutation = useMutationWithToast<Person, Error, PersonCreate>({
  mutationFn: createPerson,
  successMessage: {
    summary: 'Person created',
    detail: 'New person has been added.',
  },
  errorMessage: {
    summary: 'Create failed',
    detail: 'Unable to create person.',
  },
  invalidateQueries: [['persons']],
  onSuccess: () => {
    createValues.value = emptyFormValues()
    isFormDialogOpen.value = false
  },
})

const updateMutation = useMutationWithToast<Person, Error, { personId: string; payload: PersonUpdate }>({
  mutationFn: async (params: { personId: string; payload: PersonUpdate }) => {
    return updatePerson(params.personId, params.payload)
  },
  successMessage: {
    summary: 'Person updated',
    detail: 'Person details were saved.',
  },
  errorMessage: {
    summary: 'Update failed',
    detail: 'Unable to update person.',
  },
  invalidateQueries: [['persons']],
  onSuccess: () => {
    editingPersonId.value = null
    editValues.value = emptyFormValues()
    isFormDialogOpen.value = false
  },
})

const deleteMutation = useMutationWithToast<void, Error, string>({
  mutationFn: removePerson,
  successMessage: {
    summary: 'Person deleted',
    detail: 'Person was removed successfully.',
  },
  errorMessage: {
    summary: 'Delete failed',
    detail: 'Unable to delete person.',
  },
  invalidateQueries: [['persons']],
})

const canShowEmptyState = computed(() => {
  return !personsQuery.isPending.value && !personsQuery.isError.value && (personsQuery.data.value?.length ?? 0) === 0
})

const isCreateMode = computed(() => editingPersonId.value === null)

const activeFormValues = computed<PersonFormValues>({
  get() {
    return isCreateMode.value ? createValues.value : editValues.value
  },
  set(values) {
    if (isCreateMode.value) {
      createValues.value = values
      return
    }

    editValues.value = values
  },
})

const formTitle = computed(() => (isCreateMode.value ? 'Add Person' : 'Edit Person'))
const formLoading = computed(() => (isCreateMode.value ? createMutation.isPending.value : updateMutation.isPending.value))

const openCreateDialog = () => {
  editingPersonId.value = null
  createValues.value = emptyFormValues()
  isFormDialogOpen.value = true
}

const consumeCreateQuery = async () => {
  if (route.query.create !== '1') {
    return
  }

  openCreateDialog()
  const nextQuery = { ...route.query }
  delete nextQuery.create
  await router.replace({
    path: route.path,
    query: nextQuery,
  })
}

watch(
  () => route.query.create,
  () => {
    void consumeCreateQuery()
  },
  { immediate: true },
)

const onCreate = async () => {
  const payload = toPayload(createValues.value) as PersonCreate
  await createMutation.mutateAsync(payload)
}

const onStartEdit = (person: Person) => {
  editingPersonId.value = person.id
  editValues.value = toFormValues(person)
  isFormDialogOpen.value = true
}

const onCancelEdit = () => {
  isFormDialogOpen.value = false

  if (isCreateMode.value) {
    createValues.value = emptyFormValues()
    return
  }

  editingPersonId.value = null
  editValues.value = emptyFormValues()
}

const onSaveEdit = async () => {
  if (!editingPersonId.value) {
    return
  }

  const payload = toPayload(editValues.value)
  await updateMutation.mutateAsync({ personId: editingPersonId.value, payload })
}

const onDelete = async (personId: string) => {
  await deleteMutation.mutateAsync(personId)
}

const onDeleteFromDialog = async () => {
  if (!editingPersonId.value) return
  await onDelete(editingPersonId.value)
  onCancelEdit()
}

const onSubmitForm = async () => {
  if (isCreateMode.value) {
    await onCreate()
    return
  }

  await onSaveEdit()
}
</script>

<template>
  <section data-component="persons-page" class="flex w-full flex-col gap-4">
    <!-- Header -->
    <div class="hidden items-center justify-between md:flex">
      <h1 class="text-copy text-2xl font-bold">Persons</h1>
      <RouterLink to="/planner" class="text-brand-500 hover:text-brand-600 text-sm font-medium">
        ← Back to Planner
      </RouterLink>
    </div>

    <PersonFormDialog :open="isFormDialogOpen" :is-create-mode="isCreateMode" :title="formTitle"
      :values="activeFormValues" :weight-input-label="inputWeightLabel" :weight-options="weightOptions"
      :loading="formLoading" @update:open="(value) => { if (!value) onCancelEdit(); isFormDialogOpen = value }"
      @update:values="(values) => { activeFormValues = values }" @submit="onSubmitForm" @cancel="onCancelEdit"
      @delete="onDeleteFromDialog" />

    <AppQueryError :query="personsQuery" fallback-message="Unable to load persons." data-element="persons-error" />

    <AppLoadingState v-if="personsQuery.isPending.value" message="Loading persons..." data-element="persons-loading" />

    <AppEmptyState v-else-if="canShowEmptyState"
      message="Current crew count: 0. Morale remains surprisingly high. Add your first person to get started!"
      data-element="persons-empty-state" />

    <div v-else data-element="persons-list" class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      <button v-for="person in personsQuery.data.value" :key="person.id" data-element="person-card"
        :data-person-id="person.id" type="button"
        class="surface-panel hover:border-brand-300 cursor-pointer p-4 text-left transition"
        @click="onStartEdit(person)">
        <h3 class="text-ink text-lg font-semibold">{{ normalizeTitleWords(person.name) }}</h3>

        <div class="text-copy-muted mt-2 space-y-1 text-sm">
          <p>
            <span class="text-copy font-medium">Gender:</span>
            <span class="ml-1">{{ formatGender(person.gender) }}</span>
            <span class="text-line mx-2">|</span>
            <span class="text-copy font-medium">Age:</span>
            <span class="ml-1">{{ formatAge(person.birthdate) }}</span>
            <span class="text-line mx-2">|</span>
            <span class="text-copy font-medium">Weight:</span>
            <span class="ml-1">{{ formatWeight(person.body_weight_grams) }}</span>
          </p>
          <p>
            <span class="text-copy font-medium">Max recommended carry weight:</span>
            <span class="ml-1">{{ formatRecommendedMaxWeight(person) }}</span>
          </p>
        </div>
      </button>
    </div>
  </section>
</template>
