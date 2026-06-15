<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import Message from 'primevue/message'
import AppDialogActions from '../../../components/dialogs/AppDialogActions.vue'
import AppTemplateDialog from '../../../components/dialogs/AppTemplateDialog.vue'
import ItemLabel from '../../items/components/ItemLabel.vue'
import type { PackingList, Label } from '../types'

const props = defineProps<{
  open: boolean
  isCreateMode: boolean
  packingList: PackingList | null
  selectedLabels: Label[]
  availableLabels: Label[]
  isLoading: boolean
  labelsLoading: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [name: string, description: string]
  delete: []
  'label:add': [labelId: string]
  'label:remove': [labelId: string]
}>()

const formValues = ref({ name: '', description: '' })
const formError = ref('')
const labelSearch = ref('')

watch(() => props.open, (isOpen) => {
  if (!isOpen) {
    formValues.value = { name: '', description: '' }
    formError.value = ''
    labelSearch.value = ''
    return
  }

  if (props.packingList) {
    formValues.value = {
      name: props.packingList.name,
      description: props.packingList.description || '',
    }
  }
  else {
    formValues.value = { name: '', description: '' }
  }

  formError.value = ''
})

const handleSubmit = () => {
  formError.value = ''

  const name = formValues.value.name.trim()
  if (!name) {
    formError.value = 'Name is required'
    return
  }

  emit('submit', name, formValues.value.description.trim())
}

const handleDelete = () => {
  emit('delete')
}

const handleLabelClick = (label: Label) => {
  const isSelected = props.selectedLabels.some((l) => l.id === label.id)
  if (isSelected) {
    emit('label:remove', label.id)
  }
  else {
    emit('label:add', label.id)
  }
}

const filteredAvailableLabels = computed(() => {
  const search = labelSearch.value.trim().toLowerCase()
  if (!search) return props.availableLabels

  return props.availableLabels.filter((label) =>
    label.name.toLowerCase().includes(search)
  )
})

const dialogTitle = computed(() => props.isCreateMode ? 'Create Packing List' : 'Edit Packing List')
const submitLabel = computed(() => props.isCreateMode ? 'Create' : 'Save')
const dialogMode = computed<'create' | 'edit'>(() => props.isCreateMode ? 'create' : 'edit')
const canSubmit = computed(() => formValues.value.name.trim().length > 0)
</script>

<template>
  <AppTemplateDialog :model-value="open" data-element="packing-list-form-dialog" width="min(28rem, calc(100vw - 2rem))"
    @update:model-value="emit('update:open', $event)">
    <article class="border-line-subtle bg-surface-elevated relative rounded-2xl border p-5 shadow-panel">
      <h2 class="text-ink text-lg font-semibold pr-24">{{ dialogTitle }}</h2>

      <form class="mt-4 space-y-4" @submit.prevent="handleSubmit">
        <label class="grid gap-1">
          <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Name</span>
          <input v-model="formValues.name" type="text" class="input-shell" placeholder="Weekend trip" required />
        </label>

        <label class="grid gap-1">
          <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Description (Optional)</span>
          <textarea v-model="formValues.description" class="input-shell min-h-20"
            placeholder="Essential items for a weekend backpacking trip" />
        </label>

        <div class="space-y-2">
          <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Added Labels</span>
          <div v-if="selectedLabels.length > 0" class="flex max-h-35 flex-wrap gap-1.5 overflow-y-auto">
            <button v-for="label in selectedLabels" :key="label.id" type="button"
              class="inline-flex transition-opacity hover:opacity-75" @click="handleLabelClick(label)">
              <ItemLabel :label="label" size="sm" />
            </button>
          </div>
          <div v-else class="text-copy-muted text-xs">
            No labels added yet
          </div>
        </div>

        <div class="space-y-2 pt-3">
          <div v-if="labelsLoading" class="text-copy-muted text-sm">
            Loading labels...
          </div>
          <div v-else-if="availableLabels.length === 0" class="text-copy-muted text-sm">
            No labels available. Create labels in the gear section first.
          </div>
          <template v-else>
            <input v-model="labelSearch" aria-label="Search labels" type="text" class="input-shell w-full" placeholder="Search labels..." />

            <div v-if="filteredAvailableLabels.length === 0" class="text-copy-muted text-sm">
              No labels match your search.
            </div>
            <div v-else class="flex max-h-35 flex-wrap gap-1.5 overflow-y-auto">
              <button v-for="label in filteredAvailableLabels" :key="label.id" type="button"
                class="inline-flex transition-opacity hover:opacity-75" @click="handleLabelClick(label)">
                <ItemLabel :label="label" size="sm" />
              </button>
            </div>
          </template>
        </div>

        <Message v-if="formError" severity="error" :closable="false" class="mt-3">
          {{ formError }}
        </Message>
      </form>

      <AppDialogActions :mode="dialogMode" :submit-label="submitLabel" :can-submit="canSubmit" :is-pending="isLoading"
        :is-creating="isCreateMode && isLoading" :is-updating="!isCreateMode && isLoading"
        submit-data-element="packing-list-submit" cancel-data-element="packing-list-cancel"
        delete-data-element="packing-list-delete" @submit="handleSubmit" @cancel="emit('update:open', false)"
        @delete="handleDelete" />
    </article>
  </AppTemplateDialog>
</template>
