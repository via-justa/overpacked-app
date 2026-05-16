<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import AppQueryState from '../../../components/AppQueryState.vue'
import { listPersons } from '../../persons/api/personsApi'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { useSettings } from '../../../composables/useSettings'
import { getPersonRecommendedMaxWeightGrams } from '../../persons/utils'
import { GRAMS_PER_KILOGRAM, LB_PER_KG } from '../../../lib/units/conversions'
import type { Person } from '../../persons/types'

type BodyWeightInputUnit = 'kg' | 'lb'

const { weightUnit } = useSettings()
const defaultInputUnit = computed<BodyWeightInputUnit>(() => (weightUnit.value === 'oz' ? 'lb' : 'kg'))
const inputWeightLabel = computed<'kg' | 'lb'>(() => (defaultInputUnit.value === 'lb' ? 'lb' : 'kg'))

const gramsToInput = (grams: number, inputUnit: BodyWeightInputUnit): number => {
  const kg = grams / GRAMS_PER_KILOGRAM
  return inputUnit === 'kg' ? kg : kg * LB_PER_KG
}

const personsQuery = useQuery({
  queryKey: ['persons'],
  queryFn: listPersons,
})

const displayPersons = computed(() => {
  if (!personsQuery.data.value) return []
  return personsQuery.data.value.slice(0, 6)
})

const totalPersons = computed(() => personsQuery.data.value?.length ?? 0)

const canShowContent = computed(() => {
  return !personsQuery.isPending.value && !personsQuery.isError.value && totalPersons.value > 0
})

const formatGender = (value?: string | null) => {
  if (!value) return 'Not set'
  return value.charAt(0).toUpperCase() + value.slice(1)
}

const formatAge = (birthdate?: string | null): string => {
  if (!birthdate) return 'Not set'

  const parsed = new Date(birthdate)
  if (Number.isNaN(parsed.getTime())) return 'Not set'

  const today = new Date()
  let age = today.getFullYear() - parsed.getFullYear()
  const monthDiff = today.getMonth() - parsed.getMonth()
  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < parsed.getDate())) {
    age -= 1
  }

  return age >= 0 ? String(age) : 'Not set'
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
</script>

<template>
  <section data-component="persons-section" class="space-y-3">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <h2 class="text-copy text-xl font-bold">Persons</h2>
        <span v-if="canShowContent"
          class="bg-brand-100 text-brand-700 inline-flex h-6 min-w-6 items-center justify-center rounded-full px-2 text-xs font-semibold">
          {{ totalPersons }}
        </span>
      </div>
      <RouterLink v-if="canShowContent" to="/persons" class="text-brand-500 hover:text-brand-600 text-sm font-medium">
        View All →
      </RouterLink>
    </div>

    <AppQueryState :query="personsQuery" loading-message="Loading persons..."
      empty-message="Current crew count: 0. Morale remains surprisingly high. Time to recruit some people for your adventure!"
      error-fallback="Unable to load persons.">
      <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <RouterLink v-for="person in displayPersons" :key="person.id" :to="`/persons`"
          class="border-line-subtle bg-surface-elevated hover:border-brand-300 block rounded-xl border p-4 transition">
          <h3 class="text-copy text-base font-semibold">{{ normalizeTitleWords(person.name) }}</h3>
          <div class="text-copy-muted mt-2 space-y-1 text-xs">
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
        </RouterLink>
      </div>
    </AppQueryState>
  </section>
</template>
