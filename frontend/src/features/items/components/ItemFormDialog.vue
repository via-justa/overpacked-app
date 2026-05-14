<script setup lang="ts">
import AppTemplateDialog from '../../../components/AppTemplateDialog.vue'
import ItemFormCard from './ItemFormCard.vue'
import type { ItemFormValues, ItemTypeField, Manufacturer } from '../types'

defineProps<{
  open: boolean
  isCreateMode: boolean
  title: string
  submitLabel: string
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
  <AppTemplateDialog :model-value="open" data-element="items-form-dialog" width="min(36rem, calc(100vw - 2rem))"
    @update:model-value="$emit('update:open', $event)">
    <ItemFormCard :data-element="isCreateMode ? 'items-create-form' : 'items-edit-form'" :title="title"
      :submit-label="submitLabel" :values="values" :item-type-options="itemTypeOptions" :dynamic-fields="dynamicFields"
      :dynamic-fields-loading="dynamicFieldsLoading" :manufacturers="manufacturers"
      :weight-input-label="weightInputLabel" :volume-input-label="volumeInputLabel" :loading="isLoading" show-cancel
      @update:values="$emit('update:values', $event)"
      @request:manufacturer-create="$emit('request:manufacturer-create')" @submit="$emit('submit')"
      @cancel="$emit('cancel')" />
  </AppTemplateDialog>
</template>
