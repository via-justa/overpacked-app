<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { iconRegistry } from '../../../lib/icons'
import AppConfirmDialog from '../../../components/dialogs/AppConfirmDialog.vue'
import AppSelect from '../../../components/forms/AppSelect.vue'
import AppFormCreateDialog from '../../../components/dialogs/AppFormCreateDialog.vue'
import AppFormEditDialog from '../../../components/dialogs/AppFormEditDialog.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { useMutationWithToast } from '../../../composables/useMutationWithToast'
import { createManufacturer, removeManufacturer, updateManufacturer } from '../api/itemsApi'
import type { Item, Manufacturer, ManufacturerCreate, ManufacturerUpdate } from '../types'

type ManufacturerDialogMode = 'create' | 'edit'

const props = defineProps<{
  open: boolean
  manufacturers: Manufacturer[]
  items: Item[]
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'manufacturer-created', id: string): void
}>()

const mode = ref<ManufacturerDialogMode>('create')
const editingId = ref('')
const formValues = ref({ name: '', website: '' })
const formError = ref('')
const isDeleteConfirmOpen = ref(false)

const dialogTitle = computed(() => (mode.value === 'create' ? 'Add Manufacturer' : 'Edit Manufacturer'))

const canCreate = computed(() => formValues.value.name.trim().length > 0)
const canEdit = computed(() => editingId.value.length > 0 && formValues.value.name.trim().length > 0)
const canSubmit = computed(() => (mode.value === 'create' ? canCreate.value : canEdit.value))

const usageCount = computed(() => {
  if (!editingId.value) return 0
  return props.items.filter((item) => item.manufacturer_id === editingId.value).length
})

const isPending = computed(
  () => createMutation.isPending.value || updateMutation.isPending.value || deleteMutation.isPending.value,
)

const reset = () => {
  mode.value = 'create'
  editingId.value = ''
  formValues.value = { name: '', website: '' }
  formError.value = ''
  isDeleteConfirmOpen.value = false
}

const close = () => {
  reset()
  emit('update:open', false)
}

const setMode = (next: ManufacturerDialogMode) => {
  mode.value = next
  formError.value = ''

  if (next === 'create') {
    editingId.value = ''
    formValues.value = { name: '', website: '' }
    return
  }

  const first = props.manufacturers[0]
  if (!first) {
    editingId.value = ''
    formValues.value = { name: '', website: '' }
    return
  }

  editingId.value = first.id
  formValues.value = { name: first.name, website: first.website ?? '' }
}

const onEditTargetChange = (manufacturerId: string) => {
  editingId.value = manufacturerId
  formError.value = ''
  const selected = props.manufacturers.find((m) => m.id === manufacturerId)
  formValues.value = selected
    ? { name: selected.name, website: selected.website ?? '' }
    : { name: '', website: '' }
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) reset()
  },
)

const createMutation = useMutationWithToast<Manufacturer, Error, ManufacturerCreate>({
  mutationFn: createManufacturer,
  successMessage: {
    summary: 'Manufacturer created',
    detail: 'Manufacturer was added and selected.',
  },
  errorMessage: {
    summary: 'Create failed',
    detail: 'Unable to create manufacturer.',
  },
  invalidateQueries: [['manufacturers']],
  onSuccess: (manufacturer) => {
    emit('manufacturer-created', manufacturer.id)
    close()
  },
})

const updateMutation = useMutationWithToast<Manufacturer, Error, { manufacturerId: string; payload: ManufacturerUpdate }>({
  mutationFn: async (params: { manufacturerId: string; payload: ManufacturerUpdate }) =>
    updateManufacturer(params.manufacturerId, params.payload),
  successMessage: {
    summary: 'Manufacturer updated',
    detail: 'Manufacturer details were saved.',
  },
  errorMessage: {
    summary: 'Update failed',
    detail: 'Unable to update manufacturer.',
  },
  invalidateQueries: [['manufacturers']],
})

const deleteMutation = useMutationWithToast<void, Error, string>({
  mutationFn: removeManufacturer,
  successMessage: {
    summary: 'Manufacturer deleted',
    detail: 'Manufacturer was deleted.',
  },
  errorMessage: {
    summary: 'Delete failed',
    detail: 'Unable to delete manufacturer.',
  },
  invalidateQueries: [['manufacturers']],
  onSuccess: () => {
    reset()
  },
  onError: (err) => {
    formError.value = err instanceof Error ? err.message : 'Unable to delete manufacturer.'
  },
})

const onCreate = async () => {
  if (!canCreate.value || isPending.value) return
  const payload: ManufacturerCreate = { name: normalizeTitleWords(formValues.value.name) }
  if (formValues.value.website.trim()) payload.website = formValues.value.website.trim()
  await createMutation.mutateAsync(payload)
}

const onUpdate = async () => {
  if (!canEdit.value || isPending.value) return
  const payload: ManufacturerUpdate = { name: normalizeTitleWords(formValues.value.name) }
  if (formValues.value.website.trim()) payload.website = formValues.value.website.trim()
  await updateMutation.mutateAsync({ manufacturerId: editingId.value, payload })
}

const onDeleteRequest = () => {
  if (!editingId.value || isPending.value) return
  formError.value = ''
  if (usageCount.value > 0) {
    formError.value = `This manufacturer is used by ${usageCount.value} gear items. Delete those items first.`
    return
  }
  isDeleteConfirmOpen.value = true
}

const onDelete = async () => {
  if (!editingId.value || isPending.value) return
  isDeleteConfirmOpen.value = false
  await deleteMutation.mutateAsync(editingId.value)
}
</script>

<template>
  <component :is="mode === 'edit' ? AppFormEditDialog : AppFormCreateDialog" :open="open"
    data-element="manufacturer-create-dialog" width="min(30rem, calc(100vw - 2rem))" :title="dialogTitle"
    :can-submit="canSubmit" :is-submitting="createMutation.isPending.value || updateMutation.isPending.value"
    :is-deleting="deleteMutation.isPending.value"
    @update:open="(v: boolean) => { emit('update:open', v); if (!v) close() }"
    @submit="mode === 'create' ? onCreate() : onUpdate()" @cancel="close" @delete="onDeleteRequest">
    <div class="mt-3 grid grid-cols-2 gap-2">
      <Button data-element="manufacturer-mode-create" label="Create" :icon="`pi ${iconRegistry.action.create}`"
        :severity="mode === 'create' ? undefined : 'secondary'" :outlined="mode !== 'create'"
        class="w-full justify-center" @click="setMode('create')" />
      <Button data-element="manufacturer-mode-edit" label="Edit" :icon="`pi ${iconRegistry.action.edit}`"
        :severity="mode === 'edit' ? undefined : 'secondary'" :outlined="mode !== 'edit'" class="w-full justify-center"
        @click="setMode('edit')" />
    </div>

    <div class="mt-4 overflow-y-auto pr-1">
      <div class="grid gap-3">
        <div v-if="mode === 'edit'" class="grid gap-1">
          <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Manufacturer</span>
          <AppSelect data-element="manufacturer-edit-target" :model-value="editingId"
            @update:model-value="onEditTargetChange">
            <option v-for="manufacturer in manufacturers" :key="manufacturer.id" :value="manufacturer.id">
              {{ manufacturer.name }}
            </option>
          </AppSelect>
        </div>

        <label class="grid gap-1">
          <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Name</span>
          <input data-element="manufacturer-name" class="input-shell" :value="formValues.name" type="text"
            @input="formValues.name = ($event.target as HTMLInputElement).value" />
        </label>

        <label class="grid gap-1">
          <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Website</span>
          <input data-element="manufacturer-website" class="input-shell" :value="formValues.website" type="url"
            @input="formValues.website = ($event.target as HTMLInputElement).value" />
        </label>
      </div>

      <Message v-if="formError" severity="error" :closable="false" class="mt-3">
        {{ formError }}
      </Message>
    </div>
  </component>

  <AppConfirmDialog :open="isDeleteConfirmOpen" title="Delete Manufacturer"
    message="Delete selected manufacturer? This action cannot be undone." confirm-label="Delete" confirm-tone="danger"
    @update:open="isDeleteConfirmOpen = $event" @confirm="onDelete" />
</template>
