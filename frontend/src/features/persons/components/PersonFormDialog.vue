<script setup lang="ts">
import AppFormCreateDialog from '../../../components/AppFormCreateDialog.vue'
import PersonFormCard from './PersonFormCard.vue'
import type { PersonFormValues } from '../types'

defineProps<{
  open: boolean
  isCreateMode: boolean
  title: string
  values: PersonFormValues
  weightInputLabel: 'kg' | 'lb'
  weightOptions: number[]
  loading?: boolean
}>()

defineEmits<{
  'update:open': [value: boolean]
  'update:values': [values: PersonFormValues]
  submit: []
  cancel: []
}>()
</script>

<template>
  <AppFormCreateDialog :open="open" data-element="persons-form-dialog" width="min(36rem, calc(100vw - 2rem))"
    :title="title" :can-submit="values.name.trim().length > 0" :is-submitting="loading ?? false"
    @update:open="$emit('update:open', $event)" @submit="$emit('submit')" @cancel="$emit('cancel')">
    <PersonFormCard :data-element="isCreateMode ? 'persons-create-form' : 'persons-edit-form'" :title="title"
      submit-label="Save" :values="values" :weight-input-label="weightInputLabel"
      :weight-options="weightOptions" :loading="loading" bare
      @update:values="$emit('update:values', $event)" />
  </AppFormCreateDialog>
</template>
