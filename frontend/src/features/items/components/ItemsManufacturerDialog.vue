<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useMutation } from '@tanstack/vue-query'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { useToast } from 'primevue/usetoast'
import AppSelect from '../../../components/AppSelect.vue'
import AppTemplateDialog from '../../../components/AppTemplateDialog.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { queryClient } from '../../../lib/query/client'
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

const toast = useToast()

const mode = ref<ManufacturerDialogMode>('create')
const editingId = ref('')
const formValues = ref({ name: '', website: '' })
const formError = ref('')

const dialogTitle = computed(() => (mode.value === 'create' ? 'Add Manufacturer' : 'Edit Manufacturer'))
const submitLabel = computed(() => (mode.value === 'create' ? 'Create' : 'Save'))

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

const createMutation = useMutation({
  mutationFn: createManufacturer,
  onSuccess: async (manufacturer) => {
    await queryClient.invalidateQueries({ queryKey: ['manufacturers'] })
    emit('manufacturer-created', manufacturer.id)
    close()
    toast.add({
      severity: 'success',
      summary: 'Manufacturer created',
      detail: 'Manufacturer was added and selected.',
      life: 3000,
    })
  },
  onError: (err) => {
    toast.add({
      severity: 'error',
      summary: 'Create failed',
      detail: err instanceof Error ? err.message : 'Unable to create manufacturer.',
      life: 3500,
    })
  },
})

const updateMutation = useMutation({
  mutationFn: async (params: { manufacturerId: string; payload: ManufacturerUpdate }) =>
    updateManufacturer(params.manufacturerId, params.payload),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['manufacturers'] })
    toast.add({
      severity: 'success',
      summary: 'Manufacturer updated',
      detail: 'Manufacturer details were saved.',
      life: 3000,
    })
  },
  onError: (err) => {
    toast.add({
      severity: 'error',
      summary: 'Update failed',
      detail: err instanceof Error ? err.message : 'Unable to update manufacturer.',
      life: 3500,
    })
  },
})

const deleteMutation = useMutation({
  mutationFn: removeManufacturer,
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['manufacturers'] })
    reset()
    toast.add({
      severity: 'success',
      summary: 'Manufacturer deleted',
      detail: 'Manufacturer was deleted.',
      life: 3000,
    })
  },
  onError: (err) => {
    formError.value = err instanceof Error ? err.message : 'Unable to delete manufacturer.'
    toast.add({
      severity: 'error',
      summary: 'Delete failed',
      detail: err instanceof Error ? err.message : 'Unable to delete manufacturer.',
      life: 3500,
    })
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

const onDelete = async () => {
  if (!editingId.value || isPending.value) return
  formError.value = ''
  if (usageCount.value > 0) {
    formError.value = `This manufacturer is used by ${usageCount.value} gear items. Delete those items first.`
    return
  }
  const confirmed = globalThis.window?.confirm('Delete selected manufacturer? This action cannot be undone.') ?? false
  if (!confirmed) return
  await deleteMutation.mutateAsync(editingId.value)
}
</script>

<template>
  <AppTemplateDialog :model-value="open" data-element="manufacturer-create-dialog"
    width="min(30rem, calc(100vw - 2rem))" @update:model-value="(v) => emit('update:open', v as boolean)" @hide="close">
    <section
      class="border-line-subtle bg-surface-elevated flex max-h-[calc(100vh-8rem)] w-full flex-col rounded-2xl border p-4 shadow-panel backdrop-blur sm:p-5">
      <h3 class="text-ink shrink-0 text-lg font-semibold">{{ dialogTitle }}</h3>
      <div class="mt-3 grid grid-cols-2 gap-2">
        <Button data-element="manufacturer-mode-create" label="Create" icon="pi pi-plus"
          :severity="mode === 'create' ? undefined : 'secondary'" :outlined="mode !== 'create'"
          class="w-full justify-center" @click="setMode('create')" />
        <Button data-element="manufacturer-mode-edit" label="Edit" icon="pi pi-pencil"
          :severity="mode === 'edit' ? undefined : 'secondary'" :outlined="mode !== 'edit'"
          class="w-full justify-center" @click="setMode('edit')" />
      </div>

      <div class="mt-4 flex-1 overflow-y-auto pr-1">
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

      <div class="mt-4 flex shrink-0 items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-2">
          <Button data-element="manufacturer-create-submit" :label="submitLabel" icon="pi pi-check"
            :disabled="!canSubmit || isPending"
            :loading="createMutation.isPending.value || updateMutation.isPending.value"
            @click="mode === 'create' ? onCreate() : onUpdate()" />
          <Button data-element="manufacturer-create-cancel" label="Cancel" icon="pi pi-times" severity="secondary"
            outlined :disabled="isPending" @click="close" />
        </div>
        <Button v-if="mode === 'edit'" data-element="manufacturer-delete" label="Delete" icon="pi pi-trash"
          severity="danger" outlined class="ml-auto" :disabled="isPending" :loading="deleteMutation.isPending.value"
          @click="onDelete" />
      </div>
    </section>
  </AppTemplateDialog>
</template>
