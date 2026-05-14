<script setup lang="ts">
import { computed } from 'vue'
import Button from 'primevue/button'
import AppSelect from '../../../components/AppSelect.vue'
import AppTemplateDialog from '../../../components/AppTemplateDialog.vue'
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
  <AppTemplateDialog :model-value="modelValue" data-element="sets-form-dialog" width="min(32rem, calc(100vw - 2rem))"
    @update:model-value="onDialogUpdate">
    <article class="border-line-subtle bg-surface-elevated rounded-2xl border p-4 shadow-panel">
      <h2 class="text-ink text-lg font-semibold">{{ editingSetId ? 'Edit Set' : 'Create Set' }}</h2>
      <label class="mt-4 grid gap-1">
        <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Set name</span>
        <input v-model="setNameModel" class="input-shell" type="text" placeholder="Weekend Essentials" />
      </label>

      <div class="mt-3 grid gap-1">
        <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Set category</span>
        <AppSelect v-model="setCategoryModel">
          <option value="">Select category</option>
          <option v-for="itemType in itemTypeOptions" :key="itemType.id" :value="itemType.id">
            {{ normalizeTitleWords(itemType.name) }}
          </option>
        </AppSelect>
      </div>

      <div class="mt-4 flex items-center justify-end gap-2">
        <Button label="Cancel" severity="secondary" outlined @click="closeDialog" />
        <Button :label="editingSetId ? 'Save Changes' : 'Create Set'" icon="pi pi-check" :loading="isSubmittingSet"
          :disabled="!canSubmit" @click="emit('submit')" />
      </div>
    </article>
  </AppTemplateDialog>
</template>
