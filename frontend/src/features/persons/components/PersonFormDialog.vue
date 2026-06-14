<script setup lang="ts">
import AppFormCreateDialog from '../../../components/dialogs/AppFormCreateDialog.vue'
import AppFormEditDialog from '../../../components/dialogs/AppFormEditDialog.vue'
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
  delete: []
}>()
</script>

<template>
  <AppFormEditDialog v-if="!isCreateMode" :open="open" data-element="persons-form-dialog"
    width="min(36rem, calc(100vw - 2rem))" :title="title" :can-submit="values.name.trim().length > 0"
    :is-submitting="loading ?? false" @update:open="$emit('update:open', $event)" @submit="$emit('submit')"
    @cancel="$emit('cancel')" @delete="$emit('delete')">
    <PersonFormCard data-element="persons-edit-form" :title="title" submit-label="Save" :values="values"
      :weight-input-label="weightInputLabel" :weight-options="weightOptions" :loading="loading" bare
      @update:values="$emit('update:values', $event)" />
  </AppFormEditDialog>

  <AppFormCreateDialog v-else :open="open" data-element="persons-form-dialog" width="min(36rem, calc(100vw - 2rem))"
    :title="title" :can-submit="values.name.trim().length > 0" :is-submitting="loading ?? false"
    @update:open="$emit('update:open', $event)" @submit="$emit('submit')" @cancel="$emit('cancel')">
    <PersonFormCard data-element="persons-create-form" :title="title" submit-label="Save" :values="values"
      :weight-input-label="weightInputLabel" :weight-options="weightOptions" :loading="loading" bare
      @update:values="$emit('update:values', $event)" />
  </AppFormCreateDialog>
</template>
