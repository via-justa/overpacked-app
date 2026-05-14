<script setup lang="ts">
import AppFormCreateDialog from '../../../components/AppFormCreateDialog.vue'
import ItemFormCard from './ItemFormCard.vue'
import type { ItemFormValues, ItemTypeField, Manufacturer } from '../types'

defineProps<{
  open: boolean
  isCreateMode: boolean
  title: string
  values: ItemFormValues
  itemTypeOptions: Array<{ label: string; value: string }>
  dynamicFields: ItemTypeField[]
  dynamicFieldsLoading: boolean
  manufacturers: Manufacturer[]
  weightInputLabel: 'g' | 'oz'
  volumeInputLabel: 'ml' | 'fl oz'
  isLoading: boolean
}>()

defineEmits<{
  'update:open': [value: boolean]
  'update:values': [values: ItemFormValues]
  'request:manufacturer-create': []
  submit: []
  cancel: []
}>()
</script>

<template>
  <AppFormCreateDialog :open="open" data-element="items-form-dialog" width="min(36rem, calc(100vw - 2rem))"
    :title="title" :can-submit="!isLoading" :is-submitting="isLoading"
    @update:open="$emit('update:open', $event)" @submit="$emit('submit')" @cancel="$emit('cancel')">
    <ItemFormCard :data-element="isCreateMode ? 'items-create-form' : 'items-edit-form'" :title="title"
      submit-label="Save" :values="values" :item-type-options="itemTypeOptions" :dynamic-fields="dynamicFields"
      :dynamic-fields-loading="dynamicFieldsLoading" :manufacturers="manufacturers"
      :weight-input-label="weightInputLabel" :volume-input-label="volumeInputLabel" :loading="isLoading"
      bare @update:values="$emit('update:values', $event)"
      @request:manufacturer-create="$emit('request:manufacturer-create')" />
  </AppFormCreateDialog>
</template>
