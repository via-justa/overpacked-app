<script setup lang="ts">
import { computed } from 'vue'
import AppSelect from '../../../components/AppSelect.vue'
import AppFormCreateDialog from '../../../components/AppFormCreateDialog.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import type { ItemTypeEntity } from '../../items/types'

type Props = {
  modelValue: boolean
  editingSetId: string | null
  setNameInput: string
  setCategoryInput: string
  itemTypeOptions: ItemTypeEntity[]
  isSubmittingSet: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:setNameInput': [value: string]
  'update:setCategoryInput': [value: string]
  submit: []
}>()

const setNameModel = computed({
  get: () => props.setNameInput,
  set: (value: string) => emit('update:setNameInput', value),
})

const setCategoryModel = computed({
  get: () => props.setCategoryInput,
  set: (value: string) => emit('update:setCategoryInput', value),
})

const closeDialog = () => {
  emit('update:modelValue', false)
}

const onDialogUpdate = (value: boolean) => {
  emit('update:modelValue', value)
}

const canSubmit = computed(() => {
  return setNameModel.value.trim().length > 0 && setCategoryModel.value.trim().length > 0
})
</script>

<template>
  <AppFormCreateDialog :open="modelValue" data-element="sets-form-dialog" width="min(32rem, calc(100vw - 2rem))"
    :title="editingSetId ? 'Edit Set' : 'Create Set'" :can-submit="canSubmit" :is-submitting="isSubmittingSet"
    @update:open="onDialogUpdate" @submit="emit('submit')" @cancel="closeDialog">
    <label class="grid gap-1">
      <span class="label-field">Set name</span>
      <input v-model="setNameModel" class="input-shell" type="text" placeholder="Weekend Essentials" />
    </label>

    <div class="mt-3 grid gap-1">
        <span class="label-field">Set category</span>
      <AppSelect v-model="setCategoryModel">
        <option value="">Select category</option>
        <option v-for="itemType in itemTypeOptions" :key="itemType.id" :value="itemType.id">
          {{ normalizeTitleWords(itemType.name) }}
        </option>
      </AppSelect>
    </div>
  </AppFormCreateDialog>
</template>
