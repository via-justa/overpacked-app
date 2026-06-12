<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import { useToast } from 'primevue/usetoast'
import { iconRegistry } from '../../../lib/icons'
import { AppIcon } from '../../../components/icons'
import AppToggleGroup from '../../../components/forms/AppToggleGroup.vue'
import AppConfirmDialog from '../../../components/dialogs/AppConfirmDialog.vue'
import AppQueryError from '../../../components/feedback/AppQueryError.vue'
import AppLoadingState from '../../../components/feedback/AppLoadingState.vue'
import AppEmptyState from '../../../components/feedback/AppEmptyState.vue'
import AppCategoryFilter from '../../../components/actions/AppCategoryFilter.vue'
import AppSummaryCard from '../../../components/layout/AppSummaryCard.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { queryClient } from '../../../lib/query/client'
import { getStoredValue, setStoredValue } from '../../../lib/storage/localStorage'
import { listItems, listItemTypeFields, listItemTypes, listManufacturers, removeItem, updateItem, createItem, listItemLabels, listLabels, createLabel, addItemLabel, removeItemLabel } from '../api/itemsApi'
import { useMutationWithToast } from '../../../composables/useMutationWithToast'
import { useInlineMutation } from '../../../composables/useInlineMutation'
import { useSettings } from '../../../composables/useSettings'
import {
  gramsToInput,
  inputToGrams,
  mlToInput,
  inputToMl,
  toRoundedString,
  formatDisplayWeight,
} from '../../../lib/units/conversions'
import {
  formatNumber as formatNumberDisplay,
  formatValue as formatValueDisplay,
  formatCarryStatus as formatCarryStatusDisplay,
  formatType as formatTypeDisplay,
  formatText as formatTextDisplay,
} from '../../../lib/format/display'
import ItemFormDialog from '../components/ItemFormDialog.vue'
import ItemsCreateOptionsMenu from '../components/ItemsCreateOptionsMenu.vue'
import ItemsListView from '../components/ItemsListView.vue'
import ItemsManufacturerDialog from '../components/ItemsManufacturerDialog.vue'
import ItemsCategoryDialog from '../components/ItemsCategoryDialog.vue'
import ItemsImportDialog from '../components/ItemsImportDialog.vue'
import type { Item, ItemCreate, ItemFormValues, ItemTypeField, ItemUpdate, Label, LabelCreate } from '../types'

const toast = useToast()
const { executeInlineMutation } = useInlineMutation()
const route = useRoute()
const router = useRouter()

type WeightInputUnit = 'g' | 'oz'
type VolumeInputUnit = 'ml' | 'fl_oz'
type ItemsViewMode = 'cards' | 'table'
type ItemsTableDetailMode = 'simple' | 'expanded'
type ItemTypeFilter = 'all' | string
type CreateTarget = 'item' | 'manufacturer' | 'category' | 'import'
type TableFieldKey = 'type' | 'manufacturer' | 'active' | 'default_carry' | 'default_quantity' | 'is_default' | 'weight' | 'volume' | 'weight_volume' | 'value' | 'labels' | 'description' | 'source_url'

type TableFieldOption = {
  key: TableFieldKey
  label: string
}

type TableFieldDefinition = TableFieldOption & {
  render: (item: Item) => string
  renderHref?: (item: Item) => string | undefined
  renderBoolean?: (item: Item) => boolean | null | undefined
}

type ItemTableSection = {
  type: string
  title: string
  items: Item[]
  baseFields: TableFieldDefinition[]
  tableDetailMode: ItemsTableDetailMode
  selectionMode: boolean
  selectedItemIds: string[]
  totalWeightLabel: string
  totalValueLabel: string
  itemLabelsMap: Map<string, Label[]>
}

const ITEMS_VIEW_MODE_STORAGE_KEY = 'items:view-mode'
const ITEMS_TYPE_FILTER_STORAGE_KEY = 'items:type-filter'
const ITEMS_TABLE_EXPANDED_VIEW_STORAGE_KEY = 'items:table-expanded-view'
const ITEMS_SHOW_INACTIVE_STORAGE_KEY = 'items:show-inactive'

const isViewMode = (value: string): value is 'cards' | 'table' => value === 'cards' || value === 'table'

const commonTableFieldOptions: TableFieldOption[] = [
  { key: 'manufacturer', label: 'Manufacturer' },
  { key: 'default_carry', label: 'Default Carry' },
  { key: 'default_quantity', label: 'Qty' },
  { key: 'is_default', label: 'Default' },
  { key: 'weight_volume', label: 'W/V' },
  { key: 'value', label: 'Value' },
  { key: 'labels', label: 'Labels' },
  { key: 'description', label: 'Notes' },
  { key: 'source_url', label: 'URL' },
]

const toIntegerString = (value?: number | null): string => {
  if (typeof value !== 'number') {
    return ''
  }
  return String(Math.trunc(value))
}

const readStoredItemsViewMode = (): ItemsViewMode => {
  return getStoredValue(ITEMS_VIEW_MODE_STORAGE_KEY, isViewMode, 'table')
}

const readStoredItemsTypeFilter = (): ItemTypeFilter => {
  return getStoredValue(ITEMS_TYPE_FILTER_STORAGE_KEY, (value): value is string => value.trim().length > 0, 'all')
}

const readStoredShowInactive = (): boolean => {
  if (globalThis.window === undefined) {
    return false
  }

  const stored = globalThis.localStorage.getItem(ITEMS_SHOW_INACTIVE_STORAGE_KEY)
  return stored === 'true'
}

const readStoredExpandedView = (): boolean => {
  if (globalThis.window === undefined) {
    return false
  }

  const stored = globalThis.localStorage.getItem(ITEMS_TABLE_EXPANDED_VIEW_STORAGE_KEY)
  return stored === 'true'
}

const toNumberString = (value?: number | null): string => {
  if (typeof value !== 'number') {
    return ''
  }

  return toRoundedString(value)
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

const toFormValues = (item: Item): ItemFormValues => ({
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
      ? toRoundedString(gramsToInput(item.weight_grams, weightInputUnit.value))
      : '',
  volume_value:
    typeof item.volume_ml === 'number'
      ? toRoundedString(mlToInput(item.volume_ml, volumeInputUnit.value))
      : '',
  image_blob: item.image_blob ?? '',
  image_mime_type: item.image_mime_type ?? '',
  image_size_bytes: toIntegerString(item.image_size_bytes),
  attributes: toFormAttributes(item.attributes),
  label_ids: [],
})

const applyCommonPayloadFields = (payload: ItemCreate | ItemUpdate, values: ItemFormValues) => {
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
      payload.weight_grams = Math.round(inputToGrams(parsedWeight, weightInputUnit.value) * 100) / 100
    }
  }

  if (values.volume_value.trim()) {
    const parsedVolume = parseNumber(values.volume_value)
    if (parsedVolume !== undefined) {
      payload.volume_ml = Math.round(inputToMl(parsedVolume, volumeInputUnit.value) * 100) / 100
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

const toPayload = (values: ItemFormValues, dynamicFields: ItemTypeField[]): ItemCreate | ItemUpdate => {
  const payload: ItemCreate | ItemUpdate = {
    name: normalizeTitleWords(values.name),
    type: values.type,
    is_active: values.is_active,
    manufacturer_id: values.manufacturer_id,
  }

  applyCommonPayloadFields(payload, values)
  applyImagePayloadFields(payload, values)

  const dynamicAttributes = toDynamicAttributes(values, dynamicFields)
  if (dynamicAttributes !== undefined) {
    payload.attributes = dynamicAttributes
  }

  return payload
}

const formatWeight = (value?: number | null) => {
  if (typeof value !== 'number') {
    return 'Not set'
  }
  return formatDisplayWeight(value, weightInputUnit.value)
}

const formatTotalWeight = (valueGrams: number): string => {
  return formatDisplayWeight(valueGrams, weightInputUnit.value)
}

const formatVolume = (value?: number | null) => {
  if (typeof value !== 'number') {
    return 'Not set'
  }

  return `${toRoundedString(mlToInput(value, volumeInputUnit.value))} ${volumeInputLabel.value}`
}

const formatNumber = (value?: number | null) => {
  return formatNumberDisplay(value, toRoundedString)
}

const formatValue = (value?: number | null): string => {
  if (value === 0) {
    return 'Not set'
  }
  return formatValueDisplay(value, currency.value, toRoundedString)
}

const formatCarryStatus = (value?: string | null) => {
  return formatCarryStatusDisplay(value)
}

const formatType = (value: string) => {
  return formatTypeDisplay(value)
}

const formatText = (value?: string | null) => {
  return formatTextDisplay(value)
}

const getItemImageSrc = (item: Item) => {
  if (!item.image_blob) {
    return ''
  }

  return `data:${item.image_mime_type ?? 'image/*'};base64,${item.image_blob}`
}

const getCardSummaryEntries = (item: Item) => {
  const entries = [
    { label: 'Type', value: item.type ? formatType(item.type) : '' },
    { label: 'Manufacturer', value: manufacturersById.value.get(item.manufacturer_id) ?? '' },
    { label: 'Weight', value: typeof item.weight_grams === 'number' ? formatWeight(item.weight_grams) : '' },
    { label: 'Volume', value: typeof item.volume_ml === 'number' ? formatVolume(item.volume_ml) : '' },
    { label: 'Value', value: typeof item.value === 'number' ? formatValue(item.value) : '' },
  ]

  return entries.filter((entry) => entry.value)
}

const getCardStatEntries = (item: Item) => {
  return getCardSummaryEntries(item).filter((entry) => entry.label === 'Weight' || entry.label === 'Volume' || entry.label === 'Value')
}

const { weightUnit, volumeUnit, currency } = useSettings()
const weightInputUnit = computed<WeightInputUnit>(() => weightUnit.value)
const volumeInputUnit = computed<VolumeInputUnit>(() => volumeUnit.value)
const weightInputLabel = computed<'g' | 'oz'>(() => (weightInputUnit.value === 'oz' ? 'oz' : 'g'))
const volumeInputLabel = computed<'ml' | 'fl oz'>(() => (volumeInputUnit.value === 'fl_oz' ? 'fl oz' : 'ml'))

const itemsQuery = useQuery({
  queryKey: ['items'],
  queryFn: listItems,
})

const itemTypesQuery = useQuery({
  queryKey: ['item-types'],
  queryFn: listItemTypes,
})

const manufacturersQuery = useQuery({
  queryKey: ['manufacturers'],
  queryFn: listManufacturers,
})

const itemLabelsQueries = useQuery({
  queryKey: computed(() => ['items-labels', itemsQuery.data.value?.map(i => i.id).sort().join(',') ?? '']),
  queryFn: async () => {
    const items = itemsQuery.data.value ?? []
    if (items.length === 0) return []

    const labelsPromises = items.map(item =>
      listItemLabels(item.id).catch(() => [] as Label[])
    )
    const labelsArrays = await Promise.all(labelsPromises)

    return items.map((item, index) => ({
      itemId: item.id,
      labels: labelsArrays[index],
    }))
  },
  enabled: computed(() => (itemsQuery.data.value?.length ?? 0) > 0),
})

const itemLabelsMap = computed(() => {
  const map = new Map<string, Label[]>()
  for (const entry of itemLabelsQueries.data.value ?? []) {
    map.set(entry.itemId, entry.labels)
  }
  return map
})

const manufacturersById = computed(() => {
  const map = new Map<string, string>()
  for (const manufacturer of manufacturersQuery.data.value ?? []) {
    map.set(manufacturer.id, normalizeTitleWords(manufacturer.name))
  }

  return map
})

const manufacturerWebsitesById = computed(() => {
  const map = new Map<string, string>()
  for (const manufacturer of manufacturersQuery.data.value ?? []) {
    const website = manufacturer.website?.trim()
    if (website) {
      map.set(manufacturer.id, website)
    }
  }

  return map
})

const createValues = ref<ItemFormValues>(emptyFormValues())
const editingItemId = ref<string | null>(null)
const editValues = ref<ItemFormValues>(emptyFormValues())
const isFormDialogOpen = ref(false)
const itemsViewMode = ref<ItemsViewMode>(readStoredItemsViewMode())
const isMobileViewport = ref(false)

// Detect mobile viewport: table view and the cards/table selector are desktop/tablet only
const updateIsMobileViewport = () => {
  isMobileViewport.value = (globalThis.window?.innerWidth ?? 1024) < 768
}

// On mobile, only the cards view is available regardless of stored preference
const effectiveViewMode = computed<ItemsViewMode>(() => (isMobileViewport.value ? 'cards' : itemsViewMode.value))
const showExpandedView = ref<boolean>(readStoredExpandedView())
const itemsTableSelectionModeByType = ref<Record<string, boolean>>({})
const selectedTableItemIdsByType = ref<Record<string, string[]>>({})
const isCreateOptionsOpen = ref(false)
const isManufacturerDialogOpen = ref(false)
const isCategoryDialogOpen = ref(false)
const isImportDialogOpen = ref(false)
const createOptionsPosition = ref<{ top: number; left: number }>({ top: 84, left: 16 })
const confirmDialogState = ref<
  | { kind: 'single-delete'; itemId: string; itemName: string }
  | { kind: 'bulk-delete'; type: string; count: number }
  | null
>(null)

const allLabelsQuery = useQuery({
  queryKey: ['labels'],
  queryFn: listLabels,
  enabled: computed(() => isFormDialogOpen.value),
})

const createSelectedLabels = ref<Label[]>([])
const editSelectedLabels = ref<Label[]>([])

const selectedLabels = computed(() => (isCreateMode.value ? createSelectedLabels.value : editSelectedLabels.value))

const currentItemFormType = computed(() => (editingItemId.value === null ? createValues.value.type : editValues.value.type))

const itemTypeFieldsQuery = useQuery({
  queryKey: computed(() => ['item-type-fields', currentItemFormType.value]),
  queryFn: () => listItemTypeFields(currentItemFormType.value),
  enabled: computed(() => isFormDialogOpen.value && currentItemFormType.value.length > 0),
})

const itemTypeFieldDisplayOrder: Record<ItemTypeField['data_type'], number> = {
  enum: 0,
  string: 1,
  number: 2,
  integer: 3,
  boolean: 4,
}

const itemFormDynamicFields = computed<ItemTypeField[]>(() => {
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

const itemViewOptions: Array<{ label: string; value: ItemsViewMode }> = [
  { label: 'Cards', value: 'cards' },
  { label: 'Table', value: 'table' },
]

const showInactive = ref<boolean>(readStoredShowInactive())
const isSettingsMenuOpen = ref(false)
const settingsMenuPosition = ref({ top: 0, left: 0 })
const itemTypeFilter = ref<ItemTypeFilter>(readStoredItemsTypeFilter())
const itemTypeFilterOptions = computed<Array<{ value: ItemTypeFilter; label: string }>>(() => {
  const options: Array<{ value: ItemTypeFilter; label: string }> = [{ value: 'all', label: 'All' }]
  const types = [...new Set((itemsQuery.data.value ?? []).map((item) => item.type))].sort((left, right) => left.localeCompare(right))

  for (const type of types) {
    options.push({ value: type, label: formatType(type) })
  }

  return options
})

const createTargetOptions: Array<{ value: CreateTarget; label: string; description: string; icon: string }> = [
  { value: 'item', label: 'Create', description: 'Add a new gear item.', icon: `pi ${iconRegistry.navigation.gear}` },
  { value: 'manufacturer', label: 'Manage Manufacturers', description: 'Create and edit manufacturers.', icon: `pi ${iconRegistry.content.building}` },
  { value: 'category', label: 'Manage Categories', description: 'Create and edit custom categories.', icon: `pi ${iconRegistry.content.tag}` },
  { value: 'import', label: 'Import CSV', description: 'Preview and import gear from CSV.', icon: `pi ${iconRegistry.action.upload}` },
]

const itemFormTypeOptions = computed<Array<{ label: string; value: string }>>(() => {
  const fetched = itemTypesQuery.data.value ?? []
  return fetched
    .map((itemType) => ({ label: normalizeTitleWords(itemType.name), value: itemType.id }))
    .sort((left, right) => left.label.localeCompare(right.label))
})

const filteredItems = computed(() => {
  let items = itemsQuery.data.value ?? []

  // Filter by active status
  if (!showInactive.value) {
    items = items.filter((item) => item.is_active)
  }

  // Filter by type
  if (itemTypeFilter.value === 'all') {
    return items
  }

  return items.filter((item) => item.type === itemTypeFilter.value)
})

const getItemsTotalWeightGrams = (items: Item[]) => {
  return items.reduce((sum, item) => sum + (typeof item.weight_grams === 'number' ? item.weight_grams : 0), 0)
}

const getItemsTotalValue = (items: Item[]) => {
  return items.reduce((sum, item) => sum + (typeof item.value === 'number' ? item.value : 0), 0)
}

const formatTotalValue = (value: number): string => {
  const currencySymbol = currency.value === 'usd' ? '$' : '€'
  return `${formatNumber(value)} ${currencySymbol}`
}

const filteredSummary = computed(() => {
  const items = filteredItems.value
  return {
    totalItems: items.length,
    totalWeightLabel: formatTotalWeight(getItemsTotalWeightGrams(items)),
    totalValueLabel: formatTotalValue(getItemsTotalValue(items)),
  }
})

const createTableFieldDefinition = (field: TableFieldOption): TableFieldDefinition => {
  const { key, label } = field
  const renderBoolean = (item: Item): boolean | null | undefined => {
    switch (key) {
      case 'active': return item.is_active
      case 'is_default': return item.is_default
      default: return undefined
    }
  }

  const renderHref = (item: Item): string | undefined => {
    if (key === 'source_url') {
      return item.source_url?.trim() ? item.source_url : undefined
    }

    if (key === 'manufacturer') {
      return manufacturerWebsitesById.value.get(item.manufacturer_id)
    }

    return undefined
  }

  const render = (item: Item): string => {
    if (key === 'labels') {
      const count = itemLabelsMap.value.get(item.id)?.length ?? 0
      return count > 0 ? String(count) : 'Not set'
    }

    if (key === 'weight_volume') {
      const hasWeight = typeof item.weight_grams === 'number'
      const hasVolume = typeof item.volume_ml === 'number'

      if (hasWeight && hasVolume) {
        return `${formatWeight(item.weight_grams)} / ${formatVolume(item.volume_ml)}`
      }
      if (hasWeight) {
        return formatWeight(item.weight_grams)
      }
      if (hasVolume) {
        return formatVolume(item.volume_ml)
      }
      return 'Not set'
    }

    switch (key) {
      case 'type': return formatType(item.type)
      case 'manufacturer': return manufacturersById.value.get(item.manufacturer_id) ?? item.manufacturer_id
      case 'active': return ''
      case 'default_carry': return formatCarryStatus(item.default_carry_status)
      case 'default_quantity': return formatNumber(item.default_quantity)
      case 'is_default': return ''
      case 'weight': return formatWeight(item.weight_grams)
      case 'volume': return formatVolume(item.volume_ml)
      case 'value': return formatValue(item.value)
      case 'description': return formatText(item.description)
      case 'source_url': return item.source_url?.trim() ? 'URL' : 'Not set'
      default: return 'Not set'
    }
  }

  return { key, label, render, renderHref, renderBoolean }
}

const commonTableFieldDefinitions = commonTableFieldOptions.map(createTableFieldDefinition)

const itemTableSections = computed<ItemTableSection[]>(() => {
  const items = filteredItems.value

  const grouped = new Map<string, Item[]>()
  for (const item of items) {
    const existing = grouped.get(item.type)
    if (existing) {
      existing.push(item)
      continue
    }
    grouped.set(item.type, [item])
  }

  return [...grouped.entries()]
    .sort((left, right) => left[0].localeCompare(right[0]))
    .map(([type, groupedItems]) => {
      return {
        type,
        title: formatType(type),
        items: groupedItems,
        baseFields: commonTableFieldDefinitions,
        tableDetailMode: showExpandedView.value ? 'expanded' : 'simple',
        selectionMode: itemsTableSelectionModeByType.value[type] ?? false,
        selectedItemIds: selectedTableItemIdsByType.value[type] ?? [],
        totalWeightLabel: formatTotalWeight(getItemsTotalWeightGrams(groupedItems)),
        totalValueLabel: formatTotalValue(getItemsTotalValue(groupedItems)),
        itemLabelsMap: itemLabelsMap.value,
      }
    })
    .filter((section) => section.items.length > 0)
})

const clearSelectionForType = (type: string) => {
  selectedTableItemIdsByType.value = {
    ...selectedTableItemIdsByType.value,
    [type]: [],
  }
}

const updateTableSelectionMode = (type: string, enabled: boolean) => {
  itemsTableSelectionModeByType.value = {
    ...itemsTableSelectionModeByType.value,
    [type]: enabled,
  }

  if (!enabled) {
    clearSelectionForType(type)
  }
}

const toggleTableItemSelection = (type: string, itemId: string, checked: boolean) => {
  const current = selectedTableItemIdsByType.value[type] ?? []
  const next = checked
    ? Array.from(new Set([...current, itemId]))
    : current.filter((id) => id !== itemId)

  selectedTableItemIdsByType.value = {
    ...selectedTableItemIdsByType.value,
    [type]: next,
  }

  // Auto-exit selection mode when all rows are deselected
  if (next.length === 0 && itemsTableSelectionModeByType.value[type]) {
    itemsTableSelectionModeByType.value = {
      ...itemsTableSelectionModeByType.value,
      [type]: false,
    }
  }
}

const toggleTableSelectAll = (type: string, checked: boolean) => {
  const section = itemTableSections.value.find((entry) => entry.type === type)
  if (!section) {
    return
  }

  selectedTableItemIdsByType.value = {
    ...selectedTableItemIdsByType.value,
    [type]: checked ? section.items.map((item) => item.id) : [],
  }

  // Activate selection mode when checking all; deactivate when unchecking all
  itemsTableSelectionModeByType.value = {
    ...itemsTableSelectionModeByType.value,
    [type]: checked,
  }
}

const getSelectedIdsForType = (type: string): string[] => {
  return selectedTableItemIdsByType.value[type] ?? []
}

const withItemsRefresh = async (task: () => Promise<void>) => {
  await task()
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: ['items'] }),
    queryClient.invalidateQueries({ queryKey: ['items-labels'] }),
  ])
}

const onRowToggleActive = async (item: Item) => {
  await executeInlineMutation(
    async () => {
      await withItemsRefresh(async () => {
        await updateItem(item.id, { is_active: !item.is_active })
      })
    },
    {
      successSummary: 'Item updated',
      successDetail: `Item is now ${item.is_active ? 'inactive' : 'active'}.`,
      errorSummary: 'Update failed',
      errorDetail: 'Unable to update item.',
    }
  )
}

const onRowToggleDefault = async (item: Item) => {
  await executeInlineMutation(
    async () => {
      await withItemsRefresh(async () => {
        await updateItem(item.id, { is_default: !item.is_default })
      })
    },
    {
      successSummary: 'Item updated',
      successDetail: item.is_default ? 'Item is no longer default.' : 'Item is now default.',
      errorSummary: 'Update failed',
      errorDetail: 'Unable to update item.',
    }
  )
}

const onRowDuplicate = async (item: Item) => {
  await executeInlineMutation(
    async () => {
      const duplicateValues = toFormValues(item)
      duplicateValues.name = `${item.name} Copy`
      const payload = toPayload(duplicateValues, []) as ItemCreate
      if (item.attributes) {
        payload.attributes = item.attributes
      }

      await withItemsRefresh(async () => {
        await createItem(payload)
      })
    },
    {
      successSummary: 'Item duplicated',
      successDetail: 'A copy was created successfully.',
      errorSummary: 'Duplicate failed',
      errorDetail: 'Unable to duplicate item.',
      successLife: 3000,
    }
  )
}

const onRowDelete = async (item: Item) => {
  confirmDialogState.value = {
    kind: 'single-delete',
    itemId: item.id,
    itemName: item.name,
  }
}

const onBulkSetActive = async (type: string, value: boolean) => {
  const ids = getSelectedIdsForType(type)
  if (ids.length === 0) {
    return
  }

  try {
    await withItemsRefresh(async () => {
      await Promise.all(ids.map((id) => updateItem(id, { is_active: value })))
    })

    updateTableSelectionMode(type, false)
    toast.add({
      severity: 'success',
      summary: 'Bulk update complete',
      detail: `Updated ${ids.length} item${ids.length === 1 ? '' : 's'}.`,
      life: 3000,
    })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: 'Bulk update failed',
      detail: error instanceof Error ? error.message : 'Unable to update selected items.',
      life: 3500,
    })
  }
}

const onBulkSetDefault = async (type: string, value: boolean) => {
  const ids = getSelectedIdsForType(type)
  if (ids.length === 0) {
    return
  }

  try {
    await withItemsRefresh(async () => {
      await Promise.all(ids.map((id) => updateItem(id, { is_default: value })))
    })

    updateTableSelectionMode(type, false)
    toast.add({
      severity: 'success',
      summary: 'Bulk update complete',
      detail: `Updated ${ids.length} item${ids.length === 1 ? '' : 's'}.`,
      life: 3000,
    })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: 'Bulk update failed',
      detail: error instanceof Error ? error.message : 'Unable to update selected items.',
      life: 3500,
    })
  }
}

const onBulkDelete = async (type: string) => {
  const ids = getSelectedIdsForType(type)
  if (ids.length === 0) {
    return
  }

  confirmDialogState.value = {
    kind: 'bulk-delete',
    type,
    count: ids.length,
  }
}

const closeConfirmDialog = () => {
  confirmDialogState.value = null
}

const confirmDialogMessage = computed(() => {
  if (confirmDialogState.value?.kind === 'single-delete') {
    return `Delete ${confirmDialogState.value.itemName}?`
  }

  if (confirmDialogState.value?.kind === 'bulk-delete') {
    const count = confirmDialogState.value.count
    return `Delete ${count} selected item${count === 1 ? '' : 's'}?`
  }

  return ''
})

const onConfirmDelete = async () => {
  const current = confirmDialogState.value
  if (!current) {
    return
  }

  if (current.kind === 'single-delete') {
    closeConfirmDialog()
    await onDelete(current.itemId)
    return
  }

  const ids = getSelectedIdsForType(current.type)
  closeConfirmDialog()
  if (ids.length === 0) {
    return
  }

  try {
    await withItemsRefresh(async () => {
      await Promise.all(ids.map((id) => removeItem(id)))
    })

    updateTableSelectionMode(current.type, false)
    toast.add({
      severity: 'success',
      summary: 'Bulk delete complete',
      detail: `Deleted ${ids.length} item${ids.length === 1 ? '' : 's'}.`,
      life: 3000,
    })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: 'Bulk delete failed',
      detail: error instanceof Error ? error.message : 'Unable to delete selected items.',
      life: 3500,
    })
  }
}

const createMutation = useMutationWithToast<Item, Error, ItemCreate>({
  mutationFn: createItem,
  successMessage: {
    summary: 'Item created',
    detail: 'New item has been added.',
  },
  errorMessage: {
    summary: 'Create failed',
    detail: 'Unable to create item.',
  },
  invalidateQueries: [],
  onSuccess: () => {
    // Don't reset form or close dialog here - let onCreate handle it
  },
})

const updateMutation = useMutationWithToast<Item, Error, { itemId: string; payload: ItemUpdate }>({
  mutationFn: async (params: { itemId: string; payload: ItemUpdate }) => {
    return updateItem(params.itemId, params.payload)
  },
  successMessage: {
    summary: 'Item updated',
    detail: 'Item details were saved.',
  },
  errorMessage: {
    summary: 'Update failed',
    detail: 'Unable to update item.',
  },
  invalidateQueries: [],
  onSuccess: () => {
    // Don't reset form or close dialog here - let onSaveEdit handle it
  },
})

const deleteMutation = useMutationWithToast<void, Error, string>({
  mutationFn: removeItem,
  successMessage: {
    summary: 'Item deleted',
    detail: 'Item was removed successfully.',
  },
  errorMessage: {
    summary: 'Delete failed',
    detail: 'Unable to delete item.',
  },
  invalidateQueries: [['items'], ['items-labels']],
})

const canShowEmptyState = computed(() => {
  return !itemsQuery.isPending.value && !itemsQuery.isError.value && (itemsQuery.data.value?.length ?? 0) === 0
})

const hasFilteredItems = computed(() => filteredItems.value.length > 0)

const isCreateMode = computed(() => editingItemId.value === null)

const activeFormValues = computed<ItemFormValues>({
  get() {
    return isCreateMode.value ? createValues.value : editValues.value
  },
  set(values) {
    if (isCreateMode.value) {
      createValues.value = values
      return
    }

    editValues.value = values
  },
})

const formTitle = computed(() => (isCreateMode.value ? 'Add Item' : 'Edit Item'))
const formLoading = computed(() => (isCreateMode.value ? createMutation.isPending.value : updateMutation.isPending.value))

const closeCreateOptions = () => {
  isCreateOptionsOpen.value = false
}

const openCreateOptions = () => {
  if (globalThis.window !== undefined) {
    const trigger = globalThis.document.querySelector<HTMLElement>('[data-element="nav-create-person"]')
    if (trigger) {
      const rect = trigger.getBoundingClientRect()
      createOptionsPosition.value = {
        top: rect.bottom + 8,
        left: Math.max(16, rect.left + rect.width - 320),
      }
    }
  }

  isCreateOptionsOpen.value = true
}

const onSelectCreateTarget = (target: CreateTarget) => {
  closeCreateOptions()
  if (target === 'item') { openCreateDialog(); return }
  if (target === 'manufacturer') { isManufacturerDialogOpen.value = true; return }
  if (target === 'import') { isImportDialogOpen.value = true; return }
  isCategoryDialogOpen.value = true
}

const openManufacturerDialog = () => {
  isManufacturerDialogOpen.value = true
}

const openCreateDialog = () => {
  editingItemId.value = null
  createValues.value = emptyFormValues()
  createSelectedLabels.value = []
  isManufacturerDialogOpen.value = false
  isCategoryDialogOpen.value = false
  closeCreateOptions()
  isFormDialogOpen.value = true
}

const consumeCreateQuery = async () => {
  if (route.query.create !== '1') {
    return
  }

  openCreateOptions()
  const nextQuery = { ...route.query }
  delete nextQuery.create
  await router.replace({
    path: route.path,
    query: nextQuery,
  })
}

const consumeActionQuery = async () => {
  const action = route.query.action
  if (!action || typeof action !== 'string') {
    return
  }

  const nextQuery = { ...route.query }
  delete nextQuery.action
  await router.replace({
    path: route.path,
    query: nextQuery,
  })

  if (action === 'create-item') {
    openCreateDialog()
  } else if (action === 'manufacturers') {
    isManufacturerDialogOpen.value = true
  } else if (action === 'categories') {
    isCategoryDialogOpen.value = true
  } else if (action === 'import') {
    isImportDialogOpen.value = true
  }
}

watch(
  () => route.query.create,
  () => {
    void consumeCreateQuery()
  },
  { immediate: true },
)

watch(
  () => route.query.action,
  () => {
    void consumeActionQuery()
  },
  { immediate: true },
)

watch(itemsViewMode, (value) => {
  setStoredValue(ITEMS_VIEW_MODE_STORAGE_KEY, value)
})

watch(itemTypeFilter, (value) => {
  setStoredValue(ITEMS_TYPE_FILTER_STORAGE_KEY, value)
})

watch(showInactive, (value) => {
  setStoredValue(ITEMS_SHOW_INACTIVE_STORAGE_KEY, String(value))
})

watch(showExpandedView, (value) => {
  setStoredValue(ITEMS_TABLE_EXPANDED_VIEW_STORAGE_KEY, String(value))
})

watch(isFormDialogOpen, (value) => {
  if (!value) {
    onCancelEdit()
  }
})

onMounted(() => {
  if (globalThis.window) {
    updateIsMobileViewport()
    globalThis.window.addEventListener('resize', updateIsMobileViewport)
  }

  if (globalThis.document) {
    globalThis.document.addEventListener('click', onDocumentClickSettings)
  }
})

onBeforeUnmount(() => {
  if (globalThis.window) {
    globalThis.window.removeEventListener('resize', updateIsMobileViewport)
  }

  if (globalThis.document) {
    globalThis.document.removeEventListener('click', onDocumentClickSettings)
  }
})

const onCreate = async () => {
  const payload = toPayload(createValues.value, itemFormDynamicFields.value) as ItemCreate
  const createdItem = await createMutation.mutateAsync(payload)

  // Save label associations
  if (createSelectedLabels.value.length > 0) {
    try {
      await Promise.all(
        createSelectedLabels.value.map(label => addItemLabel(createdItem.id, { label_id: label.id }))
      )
    } catch (err) {
      // Log error but allow item creation to succeed
      // eslint-disable-next-line no-console
      console.error('Failed to add labels to item:', err)
      toast.add({
        severity: 'warn',
        summary: 'Labels partially saved',
        detail: 'Item created but some labels could not be added.',
        life: 3500,
      })
    }
  }

  // Invalidate queries and reset form after everything is done
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: ['items'] }),
    queryClient.invalidateQueries({ queryKey: ['items-labels'] }),
  ])
  createValues.value = emptyFormValues()
  createSelectedLabels.value = []
  isFormDialogOpen.value = false
}

const onStartEdit = async (item: Item) => {
  editingItemId.value = item.id
  editValues.value = toFormValues(item)

  // Load item labels
  try {
    const labels = await listItemLabels(item.id)
    editSelectedLabels.value = labels
    editValues.value = { ...editValues.value, label_ids: labels.map(l => l.id) }
  } catch (err) {
    // If labels fail to load, continue with empty labels
    // eslint-disable-next-line no-console
    console.error('Failed to load item labels:', err)
    editSelectedLabels.value = []
  }

  isManufacturerDialogOpen.value = false
  isCategoryDialogOpen.value = false
  closeCreateOptions()
  isFormDialogOpen.value = true
}

const onCancelEdit = () => {
  isFormDialogOpen.value = false

  if (isCreateMode.value) {
    createValues.value = emptyFormValues()
    createSelectedLabels.value = []
    return
  }

  editingItemId.value = null
  editValues.value = emptyFormValues()
  editSelectedLabels.value = []
}

// Open an item edit dialog when navigated to with ?open=<itemId> (e.g. from global search)
const consumeOpenQuery = async () => {
  const openId = route.query.open
  if (typeof openId !== 'string' || !openId) {
    return
  }

  const item = itemsQuery.data.value?.find((entry) => entry.id === openId)
  if (!item) {
    return
  }

  const nextQuery = { ...route.query }
  delete nextQuery.open
  await router.replace({
    path: route.path,
    query: nextQuery,
  })

  await onStartEdit(item)
}

watch(
  [() => route.query.open, () => itemsQuery.data.value],
  () => {
    void consumeOpenQuery()
  },
  { immediate: true },
)

const onSaveEdit = async () => {
  if (!editingItemId.value) {
    return
  }

  const payload = toPayload(editValues.value, itemFormDynamicFields.value)
  await updateMutation.mutateAsync({ itemId: editingItemId.value, payload })

  // Sync label associations
  try {
    const currentLabels = await listItemLabels(editingItemId.value)
    const currentLabelIds = new Set(currentLabels.map(l => l.id))
    const selectedLabelIds = new Set(editSelectedLabels.value.map(l => l.id))

    // Remove labels that are no longer selected
    const toRemove = currentLabels.filter(l => !selectedLabelIds.has(l.id))
    await Promise.all(toRemove.map(l => removeItemLabel(editingItemId.value!, l.id)))

    // Add labels that are newly selected
    const toAdd = editSelectedLabels.value.filter(l => !currentLabelIds.has(l.id))
    await Promise.all(toAdd.map(l => addItemLabel(editingItemId.value!, { label_id: l.id })))
  } catch (err) {
    // Log error but allow item update to succeed
    // eslint-disable-next-line no-console
    console.error('Failed to sync labels:', err)
    toast.add({
      severity: 'warn',
      summary: 'Labels partially saved',
      detail: 'Item updated but some labels could not be synced.',
      life: 3500,
    })
  }

  // Invalidate queries and reset form after everything is done
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: ['items'] }),
    queryClient.invalidateQueries({ queryKey: ['items-labels'] }),
  ])
  editingItemId.value = null
  editValues.value = emptyFormValues()
  editSelectedLabels.value = []
  isFormDialogOpen.value = false
}

const onDelete = async (itemId: string) => {
  await deleteMutation.mutateAsync(itemId)
}

const onDeleteFromForm = async () => {
  if (!editingItemId.value) {
    return
  }

  isFormDialogOpen.value = false
  await onDelete(editingItemId.value)
  editingItemId.value = null
  editValues.value = emptyFormValues()
  editSelectedLabels.value = []
}

const onAddLabel = (label: Label) => {
  if (isCreateMode.value) {
    if (!createSelectedLabels.value.some(l => l.id === label.id)) {
      createSelectedLabels.value = [...createSelectedLabels.value, label]
      createValues.value = { ...createValues.value, label_ids: createSelectedLabels.value.map(l => l.id) }
    }
    return
  }

  if (!editSelectedLabels.value.some(l => l.id === label.id)) {
    editSelectedLabels.value = [...editSelectedLabels.value, label]
    editValues.value = { ...editValues.value, label_ids: editSelectedLabels.value.map(l => l.id) }
  }
}

const onRemoveLabel = (labelId: string) => {
  if (isCreateMode.value) {
    createSelectedLabels.value = createSelectedLabels.value.filter(l => l.id !== labelId)
    createValues.value = { ...createValues.value, label_ids: createSelectedLabels.value.map(l => l.id) }
    return
  }

  editSelectedLabels.value = editSelectedLabels.value.filter(l => l.id !== labelId)
  editValues.value = { ...editValues.value, label_ids: editSelectedLabels.value.map(l => l.id) }
}

const generateRandomLabelColor = (): string => {
  const hue = Math.floor(Math.random() * 360)
  const saturation = 60 + Math.floor(Math.random() * 20) // 60-80%
  const lightness = 45 + Math.floor(Math.random() * 20) // 45-65%
  return `hsl(${hue}, ${saturation}%, ${lightness}%)`
}

const onCreateLabel = async (name: string) => {
  try {
    const payload: LabelCreate = { name, color: generateRandomLabelColor() }
    const created = await createLabel(payload)
    await queryClient.invalidateQueries({ queryKey: ['labels'] })
    onAddLabel(created)
  } catch (err) {
    toast.add({
      severity: 'error',
      summary: 'Label creation failed',
      detail: err instanceof Error ? err.message : 'Unable to create label.',
      life: 3500,
    })
  }
}

const onSubmitForm = async () => {
  if (isCreateMode.value) {
    await onCreate()
    return
  }

  await onSaveEdit()
}

const toggleSettingsMenu = (event: MouseEvent) => {
  if (isSettingsMenuOpen.value) {
    isSettingsMenuOpen.value = false
    return
  }

  const trigger = event.currentTarget
  if (!(trigger instanceof HTMLElement)) {
    isSettingsMenuOpen.value = true
    return
  }

  const rect = trigger.getBoundingClientRect()
  const menuWidth = 200
  const menuHeight = 60
  const gap = 8

  const left = Math.max(8, rect.right - menuWidth)
  const openUpward = rect.bottom + gap + menuHeight > globalThis.window.innerHeight - 8
  const top = openUpward ? Math.max(8, rect.top - gap - menuHeight) : rect.bottom + gap

  settingsMenuPosition.value = { top, left }
  isSettingsMenuOpen.value = true
}

const closeSettingsMenu = () => {
  isSettingsMenuOpen.value = false
}

const onDocumentClickSettings = (event: MouseEvent) => {
  const target = event.target
  if (!(target instanceof HTMLElement)) {
    closeSettingsMenu()
    return
  }

  if (target.closest('[data-element="items-settings-menu"]')) {
    return
  }

  closeSettingsMenu()
}

onMounted(() => {
  if (globalThis.document) {
    globalThis.document.addEventListener('click', onDocumentClickSettings)
  }
})

onBeforeUnmount(() => {
  if (globalThis.document) {
    globalThis.document.removeEventListener('click', onDocumentClickSettings)
  }
})
</script>

<template>
  <section data-component="items-page" class="flex w-full flex-col gap-4">
    <AppConfirmDialog :open="confirmDialogState !== null" title="Confirm delete" :message="confirmDialogMessage"
      confirm-label="Delete" confirm-tone="danger" @update:open="(value) => { if (!value) closeConfirmDialog() }"
      @cancel="closeConfirmDialog" @confirm="onConfirmDelete" />

    <ItemFormDialog :open="isFormDialogOpen" :is-create-mode="isCreateMode" :title="formTitle"
      :values="activeFormValues" :item-type-options="itemFormTypeOptions" :dynamic-fields="itemFormDynamicFields"
      :dynamic-fields-loading="itemTypeFieldsQuery.isPending.value" :manufacturers="manufacturersQuery.data.value ?? []"
      :weight-input-label="weightInputLabel" :volume-input-label="volumeInputLabel" :is-loading="formLoading"
      :all-labels="allLabelsQuery.data.value ?? []" :selected-labels="selectedLabels"
      :labels-loading="allLabelsQuery.isPending.value" @update:open="isFormDialogOpen = $event"
      @update:values="(values) => { activeFormValues = values }" @request:manufacturer-create="openManufacturerDialog"
      @label:add="onAddLabel" @label:remove="onRemoveLabel" @label:create="onCreateLabel" @submit="onSubmitForm"
      @cancel="onCancelEdit" @delete="onDeleteFromForm" />

    <ItemsManufacturerDialog v-model:open="isManufacturerDialogOpen"
      :manufacturers="manufacturersQuery.data.value ?? []" :items="itemsQuery.data.value ?? []"
      @manufacturer-created="(id) => { activeFormValues = { ...activeFormValues, manufacturer_id: id } }" />

    <ItemsCategoryDialog v-model:open="isCategoryDialogOpen" :items="itemsQuery.data.value ?? []" />

    <ItemsImportDialog v-model:open="isImportDialogOpen" />

    <AppQueryError :query="itemsQuery" fallback-message="Unable to load gear." data-element="items-error" />

    <AppQueryError :query="manufacturersQuery" fallback-message="Unable to load manufacturers."
      data-element="items-manufacturers-error" />

    <ItemsCreateOptionsMenu :open="isCreateOptionsOpen" :position="createOptionsPosition" :options="createTargetOptions"
      @update:open="isCreateOptionsOpen = $event" @select="onSelectCreateTarget" />

    <AppLoadingState v-if="itemsQuery.isPending.value" message="Loading gear..." data-element="items-loading" />

    <AppEmptyState v-else-if="canShowEmptyState"
      message="Your closet is currently operating at true ultralight standards. Add some gear to get started!"
      data-element="items-empty-state" />

    <div v-else data-element="items-list" class="space-y-3">
      <!-- Click-outside backdrop -->
      <div class="relative flex flex-wrap items-center gap-4 md:justify-center">
        <!-- View Mode Toggle: hidden on mobile (cards view only) -->
        <AppToggleGroup v-if="!isMobileViewport" name="items-view-mode" data-element="items-view-mode"
          class="md:absolute md:left-0 md:top-1/2 md:-translate-y-1/2" :model-value="itemsViewMode"
          :options="itemViewOptions" fit-content
          @update:model-value="(value) => { itemsViewMode = value as ItemsViewMode }" />

        <div data-element="items-summary" class="flex flex-wrap gap-2">
          <AppSummaryCard class="w-44" label="Gear" :value="filteredSummary.totalItems" />
          <AppSummaryCard class="w-44" label="Weight" :value="filteredSummary.totalWeightLabel" />
          <AppSummaryCard class="w-44" label="Value" :value="filteredSummary.totalValueLabel" />
        </div>

        <!-- Settings Button -->
        <div class="ml-auto md:absolute md:right-0 md:top-1/2 md:-translate-y-1/2">
          <div data-element="items-settings-menu" class="relative">
            <button type="button"
              class="text-copy-muted hover:text-copy hover:bg-surface-soft inline-flex h-9 w-9 items-center justify-center rounded-full transition"
              aria-label="View settings" @click="toggleSettingsMenu">
              <AppIcon category="navigation" name="settings" size="sm" />
            </button>

            <div v-if="isSettingsMenuOpen"
              class="border-line-subtle bg-surface-elevated fixed z-30 w-48 rounded-lg border shadow-sm" :style="{
                top: `${settingsMenuPosition.top}px`,
                left: `${settingsMenuPosition.left}px`,
              }">
              <div class="px-3 py-2.5">
                <label class="flex cursor-pointer items-center gap-2.5">
                  <input type="checkbox" v-model="showInactive" />
                  <span class="text-copy text-sm font-medium">Show inactive</span>
                </label>
                <label class="mt-2.5 flex cursor-pointer items-center gap-2.5">
                  <input type="checkbox" v-model="showExpandedView" />
                  <span class="text-copy text-sm font-medium">Expanded view</span>
                </label>
              </div>
            </div>
          </div>
        </div>
      </div>
      <!-- item type filter -->
      <AppCategoryFilter v-model="itemTypeFilter" :options="itemTypeFilterOptions" label="Item type filter"
        data-element="items-type-filter" />


      <div v-if="!hasFilteredItems" data-element="items-filter-empty-state"
        class="border-line-subtle bg-surface-elevated text-copy-muted rounded-2xl border px-5 py-6 text-sm">
        No gear matches the selected type filter.
      </div>

      <ItemsListView v-else :view-mode="effectiveViewMode" :items="filteredItems" :table-sections="itemTableSections"
        :get-image-src="getItemImageSrc" :item-labels-map="itemLabelsMap"
        @update:table-selection-mode="(type, value) => updateTableSelectionMode(type, value)"
        @toggle:table-item-selection="(type, itemId, checked) => toggleTableItemSelection(type, itemId, checked)"
        @toggle:table-select-all="(type, checked) => toggleTableSelectAll(type, checked)"
        @bulk:set-active="(type, value) => onBulkSetActive(type, value)"
        @bulk:set-default="(type, value) => onBulkSetDefault(type, value)" @bulk:delete="(type) => onBulkDelete(type)"
        @row:edit="onStartEdit" @row:duplicate="onRowDuplicate" @row:toggle-active="onRowToggleActive"
        @row:toggle-default="onRowToggleDefault" @row:delete="onRowDelete">
        <template #card-additional-info="{ item }">
          <p v-if="manufacturersById.get(item.manufacturer_id)" class="leading-6">
            <span class="text-copy font-medium">Manufacturer:</span>
            <span class="ml-1">{{ manufacturersById.get(item.manufacturer_id) }}</span>
          </p>

          <div v-if="getCardStatEntries(item).length > 0" class="flex flex-wrap items-center gap-x-3 gap-y-1 leading-6">
            <span v-for="entry in getCardStatEntries(item)" :key="`${item.id}-${entry.label}`"
              class="inline-flex items-center gap-1 whitespace-nowrap">
              <span class="text-copy font-medium">{{ entry.label }}:</span>
              <span>{{ entry.value }}</span>
            </span>
          </div>
        </template>
      </ItemsListView>
    </div>
  </section>
</template>
