<script setup lang="ts">
import AppFormCreateDialog from '../../../components/dialogs/AppFormCreateDialog.vue'
import AppFormEditDialog from '../../../components/dialogs/AppFormEditDialog.vue'
import ItemFormCard from './ItemFormCard.vue'
import type { ItemFormValues, ItemTypeField, Label, Manufacturer } from '../types'

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
  allLabels: Label[]
  selectedLabels: Label[]
  labelsLoading: boolean
}>()

defineEmits<{
  'update:open': [value: boolean]
  'update:values': [values: ItemFormValues]
  'request:manufacturer-create': []
  'label:add': [label: Label]
  'label:remove': [labelId: string]
  'label:create': [name: string]
  submit: []
  cancel: []
  delete: []
}>()
</script>

<template>
  <AppFormEditDialog v-if="!isCreateMode" :open="open" data-element="items-form-dialog"
    width="min(36rem, calc(100vw - 2rem))" :title="title" :can-submit="!isLoading" :is-submitting="isLoading"
    @update:open="$emit('update:open', $event)" @submit="$emit('submit')" @cancel="$emit('cancel')"
    @delete="$emit('delete')">
    <ItemFormCard data-element="items-edit-form" :title="title" submit-label="Save" :values="values"
      :item-type-options="itemTypeOptions" :dynamic-fields="dynamicFields"
      :dynamic-fields-loading="dynamicFieldsLoading" :manufacturers="manufacturers"
      :weight-input-label="weightInputLabel" :volume-input-label="volumeInputLabel" :loading="isLoading"
      :all-labels="allLabels" :selected-labels="selectedLabels" :labels-loading="labelsLoading" bare
      @update:values="$emit('update:values', $event)"
      @request:manufacturer-create="$emit('request:manufacturer-create')" @label:add="$emit('label:add', $event)"
      @label:remove="$emit('label:remove', $event)" @label:create="$emit('label:create', $event)" />
  </AppFormEditDialog>

  <AppFormCreateDialog v-else :open="open" data-element="items-form-dialog" width="min(36rem, calc(100vw - 2rem))"
    :title="title" :can-submit="!isLoading" :is-submitting="isLoading" @update:open="$emit('update:open', $event)"
    @submit="$emit('submit')" @cancel="$emit('cancel')">
    <ItemFormCard data-element="items-create-form" :title="title" submit-label="Save" :values="values"
      :item-type-options="itemTypeOptions" :dynamic-fields="dynamicFields"
      :dynamic-fields-loading="dynamicFieldsLoading" :manufacturers="manufacturers"
      :weight-input-label="weightInputLabel" :volume-input-label="volumeInputLabel" :loading="isLoading"
      :all-labels="allLabels" :selected-labels="selectedLabels" :labels-loading="labelsLoading" bare
      @update:values="$emit('update:values', $event)"
      @request:manufacturer-create="$emit('request:manufacturer-create')" @label:add="$emit('label:add', $event)"
      @label:remove="$emit('label:remove', $event)" @label:create="$emit('label:create', $event)" />
  </AppFormCreateDialog>
</template>
