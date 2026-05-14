<script setup lang="ts">
import { computed } from 'vue'
import Button from 'primevue/button'
import DatePicker from 'primevue/datepicker'
import AppSelect from '../../../components/AppSelect.vue'
import AppToggleGroup from '../../../components/AppToggleGroup.vue'
import type { PersonFormValues } from '../types'

const props = defineProps<{
  title: string
  submitLabel: string
  values: PersonFormValues
  weightInputLabel: 'kg' | 'lb'
  weightOptions: number[]
  loading?: boolean
  showCancel?: boolean
}>()

const emit = defineEmits<{
  submit: []
  cancel: []
  'update:values': [values: PersonFormValues]
}>()

const canSubmit = computed(() => props.values.name.trim().length > 0)
const genderOptions = [
  { label: 'Male', value: 'male' },
  { label: 'Female', value: 'female' },
  { label: 'Other', value: 'other' },
]

const conditioningLevelOptions = [
  { label: 'Sedentary', value: 'sedentary' },
  { label: 'Average', value: 'average' },
  { label: 'Athletic', value: 'athletic' },
  { label: 'Military', value: 'military' },
]

const selectableWeightOptions = computed<number[]>(() => {
  const parsedCurrent = Number(props.values.body_weight_value)
  if (Number.isNaN(parsedCurrent)) {
    return props.weightOptions
  }

  if (props.weightOptions.includes(parsedCurrent)) {
    return props.weightOptions
  }

  return [...props.weightOptions, parsedCurrent].sort((a, b) => a - b)
})

const toDate = (value: string): Date | null => {
  if (!value) {
    return null
  }

  const [year, month, day] = value.split('-').map(Number)
  if (!year || !month || !day) {
    return null
  }

  return new Date(year, month - 1, day)
}

const toDateString = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const birthdateModel = computed<Date | null>({
  get() {
    return toDate(props.values.birthdate)
  },
  set(nextValue) {
    updateField('birthdate', nextValue ? toDateString(nextValue) : '')
  },
})

const updateField = <K extends keyof PersonFormValues>(key: K, value: PersonFormValues[K]) => {
  emit('update:values', {
    ...props.values,
    [key]: value,
  })
}

const onSubmit = () => {
  if (!canSubmit.value || props.loading) {
    return
  }

  emit('submit')
}

const onCancel = () => {
  emit('cancel')
}
</script>

<template>
  <section data-component="person-form-card"
    class="border-line-subtle bg-surface-elevated rounded-2xl border p-4 shadow-panel backdrop-blur sm:p-5">
    <h2 class="text-ink text-lg font-semibold">{{ title }}</h2>

    <div class="mt-4 grid gap-3">
      <label class="grid gap-1">
        <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Name</span>
        <input data-element="person-name" class="input-shell" :value="values.name" type="text"
          @input="updateField('name', ($event.target as HTMLInputElement).value)" />
      </label>

      <div class="grid gap-1">
        <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Gender</span>
        <AppToggleGroup name="gender" data-element="person-gender" :model-value="values.gender" :options="genderOptions"
          fit-content @update:model-value="(value) => updateField('gender', value as PersonFormValues['gender'])" />
      </div>

      <div class="grid gap-1">
        <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Birthdate</span>
        <DatePicker data-element="person-birthdate" v-model="birthdateModel" class="app-date-picker w-full" fluid
          show-icon icon-display="input" date-format="dd-mm-yy" :pt="{ panel: { class: 'app-date-panel' } }"
          :manual-input="false" />
      </div>

      <div class="grid gap-1">
        <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Body Weight (<span
            class="lowercase">{{ weightInputLabel }}</span>)</span>
        <AppSelect data-element="person-body-weight" :model-value="values.body_weight_value"
          @update:model-value="(value) => updateField('body_weight_value', value)">
          <option value="">Select body weight</option>
          <option v-for="option in selectableWeightOptions" :key="option" :value="String(option)">
            {{ option }}
          </option>
        </AppSelect>
      </div>

      <div class="grid gap-1">
        <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Conditioning Level</span>
        <AppToggleGroup name="conditioning-level" data-element="person-conditioning-level"
          :model-value="values.conditioning_level" :options="conditioningLevelOptions" fit-content
          @update:model-value="(value) => updateField('conditioning_level', value as PersonFormValues['conditioning_level'])" />
      </div>
    </div>

    <footer data-element="person-form-actions" class="mt-4 flex flex-wrap items-center gap-2">
      <Button data-element="person-form-submit" :label="submitLabel" icon="pi pi-check"
        :disabled="!canSubmit || loading" :loading="loading" @click="onSubmit" />
      <Button v-if="showCancel" data-element="person-form-cancel" label="Cancel" icon="pi pi-times" severity="secondary"
        outlined :disabled="loading" @click="onCancel" />
    </footer>
  </section>
</template>
