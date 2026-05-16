import { computed, ref } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { normalizeTitleWords } from '../../lib/text/normalize'
import { getSettings } from '../settings/api/settingsApi'
import { listItemTypeFields, listItemTypes, listManufacturers, updateItem } from './api/itemsApi'
import type { Item, ItemCreate, ItemFormValues, ItemTypeField, ItemUpdate, Manufacturer } from './types'

type WeightInputUnit = 'g' | 'oz'
type VolumeInputUnit = 'ml' | 'fl_oz'

const GRAMS_PER_OUNCE = 28.349523125
const ML_PER_FL_OZ = 29.5735295625

const toRoundedString = (value: number): string => {
  if (!Number.isFinite(value)) {
    return ''
  }

  return Number.parseFloat(value.toFixed(2)).toString()
}

const toIntegerString = (value?: number | null): string => {
  if (typeof value !== 'number') {
    return ''
  }

  return String(Math.trunc(value))
}

const toNumberString = (value?: number | null): string => {
  if (typeof value !== 'number') {
    return ''
  }

  return toRoundedString(value)
}

const gramsToInput = (value: number, unit: WeightInputUnit): number => {
  return unit === 'oz' ? value / GRAMS_PER_OUNCE : value
}

const inputToGrams = (value: number, unit: WeightInputUnit): number => {
  return unit === 'oz' ? value * GRAMS_PER_OUNCE : value
}

const mlToInput = (value: number, unit: VolumeInputUnit): number => {
  return unit === 'fl_oz' ? value / ML_PER_FL_OZ : value
}

const inputToMl = (value: number, unit: VolumeInputUnit): number => {
  return unit === 'fl_oz' ? value * ML_PER_FL_OZ : value
}

const parseNumber = (value: string): number | undefined => {
  if (!value.trim()) {
    return undefined
  }

  const parsed = Number(value)
  return Number.isNaN(parsed) ? undefined : parsed
}

const parseInteger = (value: string): number | undefined => {
  if (!value.trim()) {
    return undefined
  }

  const parsed = Number(value)
  if (Number.isNaN(parsed) || !Number.isInteger(parsed)) {
    return undefined
  }

  return parsed
}

const emptyFormValues = (): ItemFormValues => ({
  name: '',
  type: 'consumable',
  is_active: true,
  manufacturer_id: '',
  description: '',
  source_url: '',
  value: '',
  default_quantity: '',
  default_carry_status: 'packed',
  is_default: true,
  weight_value: '',
  volume_value: '',
  image_blob: '',
  image_mime_type: '',
  image_size_bytes: '',
  attributes: {},
  label_ids: [],
})

const toFormAttributes = (attributes?: Record<string, unknown> | null): Record<string, string | boolean> => {
  if (!attributes) {
    return {}
  }

  const formAttributes: Record<string, string | boolean> = {}
  for (const [key, value] of Object.entries(attributes)) {
    if (typeof value === 'boolean') {
      formAttributes[key] = value
      continue
    }

    if (typeof value === 'string') {
      formAttributes[key] = value
      continue
    }

    if (typeof value === 'number') {
      formAttributes[key] = String(value)
    }
  }

  return formAttributes
}

const toFormValues = (item: Item, weightInputUnit: WeightInputUnit, volumeInputUnit: VolumeInputUnit): ItemFormValues => ({
  name: item.name,
  type: item.type,
  is_active: item.is_active,
  manufacturer_id: item.manufacturer_id,
  description: item.description ?? '',
  source_url: item.source_url ?? '',
  value: toNumberString(item.value),
  default_quantity: toIntegerString(item.default_quantity),
  default_carry_status: item.default_carry_status ?? 'packed',
  is_default: Boolean(item.is_default),
  weight_value:
    typeof item.weight_grams === 'number'
      ? toRoundedString(gramsToInput(item.weight_grams, weightInputUnit))
      : '',
  volume_value:
    typeof item.volume_ml === 'number'
      ? toRoundedString(mlToInput(item.volume_ml, volumeInputUnit))
      : '',
  image_blob: item.image_blob ?? '',
  image_mime_type: item.image_mime_type ?? '',
  image_size_bytes: toIntegerString(item.image_size_bytes),
  attributes: toFormAttributes(item.attributes),
  label_ids: [],
})

const applyCommonPayloadFields = (
  payload: ItemCreate | ItemUpdate,
  values: ItemFormValues,
  weightInputUnit: WeightInputUnit,
  volumeInputUnit: VolumeInputUnit,
) => {
  if (values.description.trim()) {
    payload.description = values.description.trim()
  }

  if (values.source_url.trim()) {
    payload.source_url = values.source_url.trim()
  }

  const parsedValue = parseNumber(values.value)
  if (parsedValue !== undefined) {
    payload.value = parsedValue
  }

  const parsedDefaultQuantity = parseNumber(values.default_quantity)
  if (parsedDefaultQuantity !== undefined) {
    payload.default_quantity = parsedDefaultQuantity
  }

  payload.default_carry_status = values.default_carry_status
  payload.is_default = values.is_default

  if (values.weight_value.trim()) {
    const parsedWeight = parseInteger(values.weight_value)
    if (parsedWeight !== undefined) {
      payload.weight_grams = Math.round(inputToGrams(parsedWeight, weightInputUnit) * 100) / 100
    }
  }

  if (values.volume_value.trim()) {
    const parsedVolume = parseNumber(values.volume_value)
    if (parsedVolume !== undefined) {
      payload.volume_ml = Math.round(inputToMl(parsedVolume, volumeInputUnit) * 100) / 100
    }
  }
}

const applyImagePayloadFields = (payload: ItemCreate | ItemUpdate, values: ItemFormValues) => {
  if (!values.image_blob.trim()) {
    return
  }

  payload.image_blob = values.image_blob.trim()
  if (values.image_mime_type.trim()) {
    payload.image_mime_type = values.image_mime_type.trim()
  }

  const parsedImageSize = parseInteger(values.image_size_bytes)
  if (parsedImageSize !== undefined) {
    payload.image_size_bytes = parsedImageSize
  }
}

const toDynamicAttributeValue = (field: ItemTypeField, rawValue: string | boolean | undefined): unknown => {
  if (field.data_type === 'boolean') {
    return typeof rawValue === 'boolean' ? rawValue : undefined
  }

  if (typeof rawValue !== 'string' || !rawValue.trim()) {
    return undefined
  }

  if (field.data_type === 'integer') {
    return parseInteger(rawValue)
  }

  if (field.data_type === 'number') {
    return parseNumber(rawValue)
  }

  return rawValue.trim()
}

const toDynamicAttributes = (values: ItemFormValues, fields: ItemTypeField[]): Record<string, unknown> | undefined => {
  if (fields.length === 0) {
    return undefined
  }

  const attributes: Record<string, unknown> = {}

  for (const field of fields) {
    const rawValue = values.attributes[field.field_key]
    const parsedValue = toDynamicAttributeValue(field, rawValue)
    if (parsedValue !== undefined) {
      attributes[field.field_key] = parsedValue
    }
  }

  return attributes
}

const toPayload = (
  values: ItemFormValues,
  dynamicFields: ItemTypeField[],
  weightInputUnit: WeightInputUnit,
  volumeInputUnit: VolumeInputUnit,
): ItemCreate | ItemUpdate => {
  const payload: ItemCreate | ItemUpdate = {
    name: normalizeTitleWords(values.name),
    type: values.type,
    is_active: values.is_active,
    manufacturer_id: values.manufacturer_id,
  }

  applyCommonPayloadFields(payload, values, weightInputUnit, volumeInputUnit)
  applyImagePayloadFields(payload, values)

  const dynamicAttributes = toDynamicAttributes(values, dynamicFields)
  if (dynamicAttributes !== undefined) {
    payload.attributes = dynamicAttributes
  }

  return payload
}

const itemTypeFieldDisplayOrder: Record<ItemTypeField['data_type'], number> = {
  enum: 0,
  string: 1,
  number: 2,
  integer: 3,
  boolean: 4,
}

export const useItemEditDialog = () => {
  const settingsQuery = useQuery({
    queryKey: ['settings'],
    queryFn: getSettings,
  })

  const itemTypesQuery = useQuery({
    queryKey: ['item-types'],
    queryFn: listItemTypes,
  })

  const manufacturersQuery = useQuery({
    queryKey: ['manufacturers'],
    queryFn: listManufacturers,
  })

  const editingItemId = ref<string | null>(null)
  const editValues = ref<ItemFormValues>(emptyFormValues())
  const isEditDialogOpen = ref(false)
  const isSubmittingEdit = ref(false)

  const weightInputUnit = computed<WeightInputUnit>(() => settingsQuery.data.value?.weight_unit ?? 'g')
  const volumeInputUnit = computed<VolumeInputUnit>(() => settingsQuery.data.value?.volume_unit ?? 'ml')
  const weightInputLabel = computed<'g' | 'oz'>(() => (weightInputUnit.value === 'oz' ? 'oz' : 'g'))
  const volumeInputLabel = computed<'ml' | 'fl oz'>(() => (volumeInputUnit.value === 'fl_oz' ? 'fl oz' : 'ml'))
  const currentItemFormType = computed(() => editValues.value.type)

  const itemTypeFieldsQuery = useQuery({
    queryKey: computed(() => ['item-type-fields', currentItemFormType.value]),
    queryFn: () => listItemTypeFields(currentItemFormType.value),
    enabled: computed(() => isEditDialogOpen.value && currentItemFormType.value.length > 0),
  })

  const itemFormTypeOptions = computed<Array<{ label: string; value: string }>>(() => {
    const fetched = itemTypesQuery.data.value ?? []
    return fetched
      .map((itemType) => ({ label: normalizeTitleWords(itemType.name), value: itemType.id }))
      .sort((left, right) => left.label.localeCompare(right.label))
  })

  const dynamicFields = computed<ItemTypeField[]>(() => {
    const fields = [...(itemTypeFieldsQuery.data.value ?? [])]

    return fields.sort((left, right) => {
      const dataTypeOrder = itemTypeFieldDisplayOrder[left.data_type] - itemTypeFieldDisplayOrder[right.data_type]
      if (dataTypeOrder !== 0) {
        return dataTypeOrder
      }

      const sortOrder = left.sort_order - right.sort_order
      if (sortOrder !== 0) {
        return sortOrder
      }

      return left.field_label.localeCompare(right.field_label)
    })
  })

  const openItemEditDialog = (item: Item) => {
    editingItemId.value = item.id
    editValues.value = toFormValues(item, weightInputUnit.value, volumeInputUnit.value)
    isEditDialogOpen.value = true
  }

  const closeItemEditDialog = () => {
    isEditDialogOpen.value = false
    editingItemId.value = null
    editValues.value = emptyFormValues()
  }

  const submitItemEdit = async () => {
    if (!editingItemId.value) {
      return null
    }

    isSubmittingEdit.value = true
    try {
      const payload = toPayload(editValues.value, dynamicFields.value, weightInputUnit.value, volumeInputUnit.value)
      return await updateItem(editingItemId.value, payload)
    } finally {
      isSubmittingEdit.value = false
    }
  }

  return {
    isEditDialogOpen,
    editValues,
    itemFormTypeOptions,
    dynamicFields,
    dynamicFieldsLoading: computed(() => itemTypeFieldsQuery.isPending.value),
    manufacturers: computed<Manufacturer[]>(() => manufacturersQuery.data.value ?? []),
    weightInputLabel,
    volumeInputLabel,
    isSubmittingEdit,
    openItemEditDialog,
    closeItemEditDialog,
    submitItemEdit,
  }
}
