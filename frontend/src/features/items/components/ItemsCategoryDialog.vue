<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useMutation, useQuery } from '@tanstack/vue-query'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { useToast } from 'primevue/usetoast'
import AppSelect from '../../../components/AppSelect.vue'
import AppTemplateDialog from '../../../components/AppTemplateDialog.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { queryClient } from '../../../lib/query/client'
import { createItemType, listItemTypeFields, listItemTypes, removeItemType, replaceItemTypeFields, updateItemType } from '../api/itemsApi'
import type { Item, ItemTypeCreate, ItemTypeField, ItemTypeFieldInput, ItemTypeUpdate } from '../types'
import { slugifyCategoryId } from '../utils/itemUtils'

type CategoryDialogMode = 'create' | 'edit'
type CategoryFieldType = 'bool' | 'string' | 'number' | 'float' | 'select'
type CategoryFieldForm = {
  name: string
  type: CategoryFieldType
  selectOptions: string
}

const props = defineProps<{
  open: boolean
  items: Item[]
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
}>()

const toast = useToast()

const mode = ref<CategoryDialogMode>('create')
const editingId = ref('')
const formValues = ref({ name: '', description: '' })
const fieldValues = ref<CategoryFieldForm[]>([{ name: '', type: 'string', selectOptions: '' }])
const formError = ref('')
const isFieldsLoading = ref(false)

const itemTypesQuery = useQuery({
  queryKey: ['item-types'],
  queryFn: listItemTypes,
})

const itemTypeOptions = computed(() =>
  [...(itemTypesQuery.data.value ?? [])].sort((a, b) => a.name.localeCompare(b.name)),
)

const dialogTitle = computed(() => (mode.value === 'create' ? 'Add Category' : 'Edit Category'))
const submitLabel = computed(() => (mode.value === 'create' ? 'Create' : 'Save'))

const canCreate = computed(() => {
  if (!formValues.value.name.trim()) return false
  if (fieldValues.value.length === 0) return false
  return fieldValues.value.every((field) => {
    if (!field.name.trim()) return false
    if (field.type !== 'select') return true
    return field.selectOptions
      .split(',')
      .map((v) => v.trim())
      .some((v) => v.length > 0)
  })
})
const canEdit = computed(() => editingId.value.length > 0 && canCreate.value)
const canSubmit = computed(() => (mode.value === 'create' ? canCreate.value : canEdit.value))

const usageCount = computed(() => {
  if (!editingId.value) return 0
  return props.items.filter((item) => item.type === editingId.value).length
})

const isPending = computed(
  () => createMutation.isPending.value || updateMutation.isPending.value || deleteMutation.isPending.value,
)

const categoryFieldTypeOptions: Array<{ value: CategoryFieldType; label: string }> = [
  { value: 'bool', label: 'bool' },
  { value: 'string', label: 'string' },
  { value: 'number', label: 'number' },
  { value: 'float', label: 'float' },
  { value: 'select', label: 'select' },
]

const reset = () => {
  mode.value = 'create'
  editingId.value = ''
  formValues.value = { name: '', description: '' }
  fieldValues.value = [{ name: '', type: 'string', selectOptions: '' }]
  formError.value = ''
  isFieldsLoading.value = false
}

const close = () => {
  reset()
  emit('update:open', false)
}

const toCategoryFieldForm = (field: ItemTypeField): CategoryFieldForm => {
  if (field.data_type === 'boolean') return { name: field.field_label, type: 'bool', selectOptions: '' }
  if (field.data_type === 'integer') return { name: field.field_label, type: 'number', selectOptions: '' }
  if (field.data_type === 'number') return { name: field.field_label, type: 'float', selectOptions: '' }
  if (field.data_type === 'enum') {
    return { name: field.field_label, type: 'select', selectOptions: (field.enum_options ?? []).join(', ') }
  }
  return { name: field.field_label, type: 'string', selectOptions: '' }
}

const loadForEdit = async (typeId: string) => {
  const category = itemTypeOptions.value.find((t) => t.id === typeId)
  if (!category) return
  formValues.value = { name: category.name, description: category.description ?? '' }
  isFieldsLoading.value = true
  try {
    const fields = await listItemTypeFields(typeId)
    const mapped = fields.sort((a, b) => a.sort_order - b.sort_order).map(toCategoryFieldForm)
    fieldValues.value = mapped.length > 0 ? mapped : [{ name: '', type: 'string', selectOptions: '' }]
  } catch (err) {
    formError.value = err instanceof Error ? err.message : 'Unable to load category fields.'
    fieldValues.value = [{ name: '', type: 'string', selectOptions: '' }]
  } finally {
    isFieldsLoading.value = false
  }
}

const setMode = async (next: CategoryDialogMode) => {
  mode.value = next
  formError.value = ''

  if (next === 'create') {
    editingId.value = ''
    formValues.value = { name: '', description: '' }
    fieldValues.value = [{ name: '', type: 'string', selectOptions: '' }]
    return
  }

  const first = itemTypeOptions.value[0]
  if (!first) {
    editingId.value = ''
    formValues.value = { name: '', description: '' }
    fieldValues.value = [{ name: '', type: 'string', selectOptions: '' }]
    return
  }

  editingId.value = first.id
  await loadForEdit(first.id)
}

const onEditTargetChange = async (typeId: string) => {
  editingId.value = typeId
  formError.value = ''
  await loadForEdit(typeId)
}

const addField = () => {
  fieldValues.value = [...fieldValues.value, { name: '', type: 'string', selectOptions: '' }]
}

const removeField = (index: number) => {
  if (fieldValues.value.length <= 1) return
  fieldValues.value = fieldValues.value.filter((_, i) => i !== index)
}

const updateField = (index: number, next: CategoryFieldForm) => {
  fieldValues.value = fieldValues.value.map((f, i) => (i === index ? next : f))
}

const toFieldInput = (field: CategoryFieldForm, index: number): ItemTypeFieldInput => {
  const fieldKey = slugifyCategoryId(field.name) || `field_${index + 1}`
  if (field.type === 'bool') {
    return { field_key: fieldKey, field_label: normalizeTitleWords(field.name), data_type: 'boolean', sort_order: index + 1 }
  }
  if (field.type === 'number') {
    return { field_key: fieldKey, field_label: normalizeTitleWords(field.name), data_type: 'integer', sort_order: index + 1 }
  }
  if (field.type === 'float') {
    return { field_key: fieldKey, field_label: normalizeTitleWords(field.name), data_type: 'number', sort_order: index + 1 }
  }
  if (field.type === 'select') {
    return {
      field_key: fieldKey,
      field_label: normalizeTitleWords(field.name),
      data_type: 'enum',
      enum_options: field.selectOptions
        .split(',')
        .map((v) => v.trim())
        .filter((v) => v.length > 0),
      sort_order: index + 1,
    }
  }
  return { field_key: fieldKey, field_label: normalizeTitleWords(field.name), data_type: 'string', sort_order: index + 1 }
}

const checkDuplicateFieldKey = (): string => {
  const keySet = new Set<string>()
  for (let i = 0; i < fieldValues.value.length; i += 1) {
    const key = slugifyCategoryId(fieldValues.value[i].name) || `field_${i + 1}`
    if (keySet.has(key)) return key
    keySet.add(key)
  }
  return ''
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) reset()
  },
)

const createMutation = useMutation({
  mutationFn: async (payload: { categoryPayload: ItemTypeCreate; fields: ItemTypeFieldInput[] }) => {
    const created = await createItemType(payload.categoryPayload)
    await replaceItemTypeFields(created.id, payload.fields)
    return created
  },
  onSuccess: async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['item-types'] }),
      queryClient.invalidateQueries({ queryKey: ['items'] }),
    ])
    close()
    toast.add({ severity: 'success', summary: 'Category created', detail: 'New category has been added.', life: 3000 })
  },
  onError: (err) => {
    toast.add({
      severity: 'error',
      summary: 'Create failed',
      detail: err instanceof Error ? err.message : 'Unable to create category.',
      life: 3500,
    })
  },
})

const updateMutation = useMutation({
  mutationFn: async (params: { typeId: string; payload: ItemTypeUpdate; fields: ItemTypeFieldInput[] }) => {
    await updateItemType(params.typeId, params.payload)
    await replaceItemTypeFields(params.typeId, params.fields)
  },
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['item-types'] })
    toast.add({ severity: 'success', summary: 'Category updated', detail: 'Category details were saved.', life: 3000 })
  },
  onError: (err) => {
    toast.add({
      severity: 'error',
      summary: 'Update failed',
      detail: err instanceof Error ? err.message : 'Unable to update category.',
      life: 3500,
    })
  },
})

const deleteMutation = useMutation({
  mutationFn: removeItemType,
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['item-types'] })
    close()
    toast.add({ severity: 'success', summary: 'Category deleted', detail: 'Category was deleted.', life: 3000 })
  },
  onError: (err) => {
    formError.value = err instanceof Error ? err.message : 'Unable to delete category.'
    toast.add({
      severity: 'error',
      summary: 'Delete failed',
      detail: err instanceof Error ? err.message : 'Unable to delete category.',
      life: 3500,
    })
  },
})

const onCreate = async () => {
  if (!canCreate.value || isPending.value) return
  formError.value = ''
  const duplicateKey = checkDuplicateFieldKey()
  if (duplicateKey) {
    formError.value = 'Property names must be unique. Rename duplicate fields and try again.'
    return
  }
  const categoryPayload: ItemTypeCreate = {
    id: slugifyCategoryId(formValues.value.name),
    name: normalizeTitleWords(formValues.value.name),
  }
  if (formValues.value.description.trim()) categoryPayload.description = formValues.value.description.trim()
  try {
    await createMutation.mutateAsync({ categoryPayload, fields: fieldValues.value.map(toFieldInput) })
  } catch (err) {
    formError.value = err instanceof Error ? err.message : 'Unable to create category.'
  }
}

const onUpdate = async () => {
  if (!canEdit.value || isPending.value) return
  formError.value = ''
  const duplicateKey = checkDuplicateFieldKey()
  if (duplicateKey) {
    formError.value = 'Property names must be unique. Rename duplicate fields and try again.'
    return
  }
  const payload: ItemTypeUpdate = { name: normalizeTitleWords(formValues.value.name) }
  if (formValues.value.description.trim()) payload.description = formValues.value.description.trim()
  try {
    await updateMutation.mutateAsync({ typeId: editingId.value, payload, fields: fieldValues.value.map(toFieldInput) })
  } catch (err) {
    formError.value = err instanceof Error ? err.message : 'Unable to update category.'
  }
}

const onDelete = async () => {
  if (!editingId.value || isPending.value) return
  formError.value = ''
  if (usageCount.value > 0) {
    formError.value = `This category is used by ${usageCount.value} gear items. Delete those items first.`
    return
  }
  const confirmed = globalThis.window?.confirm('Delete selected category? This action cannot be undone.') ?? false
  if (!confirmed) return
  await deleteMutation.mutateAsync(editingId.value)
}
</script>

<template>
  <AppTemplateDialog :model-value="open" data-element="category-create-dialog" width="min(30rem, calc(100vw - 2rem))"
    @update:model-value="(v) => emit('update:open', v as boolean)" @hide="close">
    <section
      class="border-line-subtle bg-surface-elevated flex max-h-[calc(100vh-8rem)] w-full flex-col rounded-2xl border p-4 shadow-panel backdrop-blur sm:p-5">
      <h3 class="text-ink shrink-0 text-lg font-semibold">{{ dialogTitle }}</h3>
      <div class="mt-3 grid grid-cols-2 gap-2">
        <Button data-element="category-mode-create" label="Create" icon="pi pi-plus"
          :severity="mode === 'create' ? undefined : 'secondary'" :outlined="mode !== 'create'"
          class="w-full justify-center" @click="void setMode('create')" />
        <Button data-element="category-mode-edit" label="Edit" icon="pi pi-pencil"
          :severity="mode === 'edit' ? undefined : 'secondary'" :outlined="mode !== 'edit'"
          class="w-full justify-center" @click="void setMode('edit')" />
      </div>

      <div class="mt-4 flex-1 overflow-y-auto pr-1">
        <div class="grid gap-3">
          <div v-if="mode === 'edit'" class="grid gap-1">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Category</span>
            <AppSelect data-element="category-edit-target" :model-value="editingId"
              @update:model-value="(value) => { void onEditTargetChange(value) }">
              <option v-for="category in itemTypeOptions" :key="category.id" :value="category.id">
                {{ category.name }}
              </option>
            </AppSelect>
          </div>

          <label class="grid gap-1">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Name</span>
            <input data-element="category-name" class="input-shell" :value="formValues.name" type="text"
              @input="formValues.name = ($event.target as HTMLInputElement).value" />
          </label>

          <div class="grid gap-2">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Properties</span>
            <p v-if="isFieldsLoading" class="text-copy-subtle text-sm">Loading category fields...</p>
            <div v-for="(field, index) in fieldValues" :key="`field-${index}`"
              class="border-line-subtle bg-surface-muted grid gap-2 rounded-lg border p-2">
              <div class="grid gap-2 sm:grid-cols-[1fr,12rem,auto]">
                <input :data-element="`category-field-name-${index}`" class="input-shell" :value="field.name"
                  type="text" placeholder="Property name"
                  @input="updateField(index, { ...field, name: ($event.target as HTMLInputElement).value })" />
                <AppSelect :data-element="`category-field-type-${index}`" :model-value="field.type"
                  @update:model-value="(value) => updateField(index, { ...field, type: value as CategoryFieldType, selectOptions: value === 'select' ? field.selectOptions : '' })">
                  <option v-for="option in categoryFieldTypeOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </AppSelect>
                <Button :data-element="`category-field-remove-${index}`" icon="pi pi-trash" severity="secondary"
                  outlined :disabled="fieldValues.length <= 1" @click="removeField(index)" />
              </div>
              <input v-if="field.type === 'select'" :data-element="`category-field-options-${index}`"
                class="input-shell" :value="field.selectOptions" type="text"
                placeholder="Select options (comma separated)"
                @input="updateField(index, { ...field, selectOptions: ($event.target as HTMLInputElement).value })" />
            </div>
            <Button data-element="category-field-add" label="Add field" icon="pi pi-plus" severity="secondary" outlined
              @click="addField" />
          </div>

          <label class="grid gap-1">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Description</span>
            <textarea data-element="category-description" class="input-shell min-h-4.5rem"
              :value="formValues.description"
              @input="formValues.description = ($event.target as HTMLTextAreaElement).value" />
          </label>
        </div>

        <Message v-if="formError" severity="error" :closable="false" class="mt-3">
          {{ formError }}
        </Message>
      </div>

      <div class="mt-4 flex shrink-0 items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-2">
          <Button data-element="category-create-submit" :label="submitLabel" icon="pi pi-check"
            :disabled="!canSubmit || isPending || isFieldsLoading"
            :loading="createMutation.isPending.value || updateMutation.isPending.value"
            @click="mode === 'create' ? onCreate() : onUpdate()" />
          <Button data-element="category-create-cancel" label="Cancel" icon="pi pi-times" severity="secondary" outlined
            :disabled="isPending" @click="close" />
        </div>
        <Button v-if="mode === 'edit'" data-element="category-delete" label="Delete" icon="pi pi-trash"
          severity="danger" outlined class="ml-auto" :disabled="isPending" :loading="deleteMutation.isPending.value"
          @click="onDelete" />
      </div>
    </section>
  </AppTemplateDialog>
</template>
