<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMutation, useQuery } from '@tanstack/vue-query'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { useToast } from 'primevue/usetoast'
import AppTemplateDialog from '../../../components/AppTemplateDialog.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { queryClient } from '../../../lib/query/client'
import { getSettings } from '../../settings/api/settingsApi'
import {
  createPerson,
  listPersons,
  removePerson,
  updatePerson,
} from '../api/personsApi'
import PersonFormCard from '../components/PersonFormCard.vue'
import type { Person, PersonCreate, PersonFormValues, PersonUpdate } from '../types'
import type { WeightUnit } from '../../settings/types'

import { getPersonRecommendedMaxWeightGrams } from '../utils'

const toast = useToast()
const route = useRoute()
const router = useRouter()
const GRAMS_PER_KILOGRAM = 1000
const LB_PER_KG = 2.2046226218

type BodyWeightInputUnit = 'kg' | 'lb'

const settingsQuery = useQuery({
  queryKey: ['settings'],
  queryFn: getSettings,
})

const weightUnit = computed<WeightUnit>(() => settingsQuery.data.value?.weight_unit ?? 'g')
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

const createMutation = useMutation({
  mutationFn: createPerson,
  onSuccess: async () => {
    createValues.value = emptyFormValues()
    isFormDialogOpen.value = false
    await queryClient.invalidateQueries({ queryKey: ['persons'] })
    toast.add({
      severity: 'success',
      summary: 'Person created',
      detail: 'New person has been added.',
      life: 3000,
    })
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Create failed',
      detail: error instanceof Error ? error.message : 'Unable to create person.',
      life: 3500,
    })
  },
})

const updateMutation = useMutation({
  mutationFn: async (params: { personId: string; payload: PersonUpdate }) => {
    return updatePerson(params.personId, params.payload)
  },
  onSuccess: async () => {
    editingPersonId.value = null
    editValues.value = emptyFormValues()
    isFormDialogOpen.value = false
    await queryClient.invalidateQueries({ queryKey: ['persons'] })
    toast.add({
      severity: 'success',
      summary: 'Person updated',
      detail: 'Person details were saved.',
      life: 3000,
    })
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Update failed',
      detail: error instanceof Error ? error.message : 'Unable to update person.',
      life: 3500,
    })
  },
})

const deleteMutation = useMutation({
  mutationFn: removePerson,
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['persons'] })
    toast.add({
      severity: 'success',
      summary: 'Person deleted',
      detail: 'Person was removed successfully.',
      life: 3000,
    })
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Delete failed',
      detail: error instanceof Error ? error.message : 'Unable to delete person.',
      life: 3500,
    })
  },
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
const formSubmitLabel = computed(() => (isCreateMode.value ? 'Create Person' : 'Save Changes'))
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
    <AppTemplateDialog v-model="isFormDialogOpen" data-element="persons-form-dialog"
      width="min(36rem, calc(100vw - 2rem))" @hide="onCancelEdit">
      <PersonFormCard :data-element="isCreateMode ? 'persons-create-form' : 'persons-edit-form'" :title="formTitle"
        :submit-label="formSubmitLabel" :values="activeFormValues" :weight-input-label="inputWeightLabel"
        :weight-options="weightOptions" :loading="formLoading" show-cancel
        @update:values="(values) => { activeFormValues = values }" @submit="onSubmitForm" @cancel="onCancelEdit" />
    </AppTemplateDialog>

    <Message v-if="personsQuery.isError.value" data-element="persons-error" severity="error" :closable="false">
      {{ personsQuery.error.value instanceof Error ? personsQuery.error.value.message : 'Unable to load persons.' }}
    </Message>

    <div v-if="personsQuery.isPending.value" data-element="persons-loading"
      class="border-line-subtle bg-surface-muted text-copy-muted rounded-2xl border px-4 py-3 text-sm font-medium">
      Loading persons...
    </div>

    <div v-else-if="canShowEmptyState" data-element="persons-empty-state"
      class="border-line-subtle bg-surface-elevated text-copy-muted rounded-2xl border px-5 py-6 text-sm">
      Current crew count: 0. Morale remains surprisingly high. Add your first person to get started!
    </div>

    <div v-else data-element="persons-list" class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      <article v-for="person in personsQuery.data.value" :key="person.id" data-element="person-card"
        :data-person-id="person.id" class="border-line-subtle bg-surface-elevated rounded-2xl border p-4 shadow-panel">
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
            <span class="text-copy font-medium">Max recommended pack weight:</span>
            <span class="ml-1">{{ formatRecommendedMaxWeight(person) }}</span>
          </p>
        </div>

        <div class="mt-4 flex items-center justify-between gap-2">
          <Button data-element="person-edit" label="Edit" icon="pi pi-pencil" size="small" outlined
            @click="onStartEdit(person)" />
          <Button data-element="person-delete" label="Delete" icon="pi pi-trash" size="small" severity="danger" outlined
            :loading="deleteMutation.isPending.value" @click="onDelete(person.id)" />
        </div>
      </article>
    </div>
  </section>
</template>
