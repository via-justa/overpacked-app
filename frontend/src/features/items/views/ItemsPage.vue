<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMutation, useQuery } from '@tanstack/vue-query'
import Message from 'primevue/message'
import { useToast } from 'primevue/usetoast'
import AppToggleGroup from '../../../components/AppToggleGroup.vue'
import AppConfirmDialog from '../../../components/AppConfirmDialog.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { queryClient } from '../../../lib/query/client'
import { getSettings } from '../../settings/api/settingsApi'
import { listItems, listItemTypeFields, listItemTypes, listManufacturers, removeItem, updateItem, createItem } from '../api/itemsApi'
import ItemDetailsDialog from '../components/ItemDetailsDialog.vue'
import ItemFormDialog from '../components/ItemFormDialog.vue'
import ItemsCreateOptionsMenu from '../components/ItemsCreateOptionsMenu.vue'
import ItemsListView from '../components/ItemsListView.vue'
import ItemsManufacturerDialog from '../components/ItemsManufacturerDialog.vue'
import ItemsCategoryDialog from '../components/ItemsCategoryDialog.vue'
import ItemsImportDialog from '../components/ItemsImportDialog.vue'
import type { Item, ItemCreate, ItemFormValues, ItemTypeField, ItemUpdate, KnownItemType } from '../types'
import { KNOWN_ITEM_TYPES, isKnownItemType } from '../types'
import type { Currency } from '../../settings/types'

const toast = useToast()
const route = useRoute()
const router = useRouter()

const GRAMS_PER_OUNCE = 28.349523125
const GRAMS_PER_KILOGRAM = 1000
const OUNCES_PER_POUND = 16
const ML_PER_FL_OZ = 29.5735295625

type WeightInputUnit = 'g' | 'oz'
type VolumeInputUnit = 'ml' | 'fl_oz'
type ItemsViewMode = 'cards' | 'table'
type ItemsTableDetailMode = 'simple' | 'expanded'
type ItemTypeFilter = 'all' | string
type CreateTarget = 'item' | 'manufacturer' | 'category' | 'import'
type TableFieldKey = 'type' | 'manufacturer' | 'active' | 'default_carry' | 'default_quantity' | 'is_default' | 'weight' | 'volume' | 'value' | 'description' | 'source_url' | 'dose_count' | 'calories' | 'calories_per_serving' | 'requires_water' | 'season' | 'layer' | 'waterproof' | 'size' | 'color' | 'capacity_people' | 'season_rating' | 'freestanding' | 'has_footprint' | 'comfort_temp_c' | 'fill_type' | 'r_value' | 'capacity_mah' | 'charge_port' | 'rechargeable'

type TableFieldOption = {
  key: TableFieldKey
  label: string
}

type TableFieldDefinition = TableFieldOption & {
  render: (item: Item) => string
  renderHref?: (item: Item) => string | undefined
  renderBoolean?: (item: Item) => boolean | null | undefined
}

type DetailEntry = {
  label: string
  value: string
  href?: string
  booleanValue?: boolean | null
}

type ItemTableSection = {
  type: string
  title: string
  items: Item[]
  baseFields: TableFieldDefinition[]
  extraFields: TableFieldDefinition[]
  tableDetailMode: ItemsTableDetailMode
  selectionMode: boolean
  selectedItemIds: string[]
  totalWeightLabel: string
  totalValueLabel: string
}

const ITEMS_VIEW_MODE_STORAGE_KEY = 'items:view-mode'
const ITEMS_TYPE_FILTER_STORAGE_KEY = 'items:type-filter'
const ITEMS_TABLE_DETAIL_MODE_BY_TYPE_STORAGE_KEY = 'items:table-detail-mode-by-type'
const commonTableFieldOptions: TableFieldOption[] = [
  { key: 'type', label: 'Type' },
  { key: 'manufacturer', label: 'Manufacturer' },
  { key: 'active', label: 'Active' },
  { key: 'default_carry', label: 'Default Carry' },
  { key: 'default_quantity', label: 'Qty' },
  { key: 'is_default', label: 'Default' },
  { key: 'weight', label: 'Weight' },
  { key: 'volume', label: 'Volume' },
  { key: 'value', label: 'Value' },
  { key: 'description', label: 'Notes' },
  { key: 'source_url', label: 'URL' },
]

const extraTableFieldOptionsByType: Record<KnownItemType, TableFieldOption[]> = {
  consumable: [
    { key: 'dose_count', label: 'Dose Count' },
    { key: 'calories', label: 'Calories' },
    { key: 'calories_per_serving', label: 'Cal/Serving' },
    { key: 'requires_water', label: 'Requires Water' },
  ],
  wearable: [
    { key: 'season', label: 'Season' },
    { key: 'layer', label: 'Layer' },
    { key: 'waterproof', label: 'Waterproof' },
    { key: 'size', label: 'Size' },
    { key: 'color', label: 'Color' },
  ],
  shelter: [
    { key: 'capacity_people', label: 'Capacity' },
    { key: 'season_rating', label: 'Season Rating' },
    { key: 'freestanding', label: 'Freestanding' },
    { key: 'has_footprint', label: 'Has Footprint' },
  ],
  sleep: [
    { key: 'comfort_temp_c', label: 'Comfort Temp (°C)' },
    { key: 'fill_type', label: 'Fill Type' },
    { key: 'r_value', label: 'R-Value' },
  ],
  electronics: [
    { key: 'capacity_mah', label: 'Capacity (mAh)' },
    { key: 'charge_port', label: 'Charge Port' },
    { key: 'rechargeable', label: 'Rechargeable' },
  ],
}

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

const readStoredItemsViewMode = (): ItemsViewMode => {
  if (globalThis.window === undefined) {
    return 'table'
  }

  const stored = globalThis.localStorage.getItem(ITEMS_VIEW_MODE_STORAGE_KEY)
  return stored === 'cards' || stored === 'table' ? stored : 'table'
}

const readStoredItemsTypeFilter = (): ItemTypeFilter => {
  if (globalThis.window === undefined) {
    return 'all'
  }

  const stored = globalThis.localStorage.getItem(ITEMS_TYPE_FILTER_STORAGE_KEY)
  return stored && stored.trim().length > 0 ? stored : 'all'
}

const readStoredItemsTableDetailModeByType = (): Record<string, ItemsTableDetailMode> => {
  if (globalThis.window === undefined) {
    return {}
  }

  const stored = globalThis.localStorage.getItem(ITEMS_TABLE_DETAIL_MODE_BY_TYPE_STORAGE_KEY)
  if (!stored) {
    return {}
  }

  try {
    const parsed = JSON.parse(stored) as Record<string, unknown>
    const next: Record<string, ItemsTableDetailMode> = {}

    for (const [type, rawValue] of Object.entries(parsed)) {
      if (rawValue === 'simple' || rawValue === 'expanded') {
        next[type] = rawValue
      }
    }

    return next
  } catch {
    return {}
  }
}

const toNumberString = (value?: number | null): string => {
  if (typeof value !== 'number') {
    return ''
  }

  return toRoundedString(value)
}

const toBooleanValue = Boolean

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
  dose_count: '',
  calories: '',
  calories_per_serving: '',
  requires_water: false,
  season: '',
  layer: '',
  waterproof: false,
  size: '',
  color: '',
  capacity_people: '',
  season_rating: '',
  freestanding: false,
  has_footprint: false,
  comfort_temp_c: '',
  fill_type: '',
  r_value: '',
  capacity_mah: '',
  charge_port: '',
  rechargeable: false,
  image_blob: '',
  image_mime_type: '',
  image_size_bytes: '',
  attributes: {},
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
  dose_count: toIntegerString(item.dose_count),
  calories: toNumberString(item.calories),
  calories_per_serving: toNumberString(item.calories_per_serving),
  requires_water: toBooleanValue(item.requires_water),
  season: item.season ? item.season as ItemFormValues['season'] : '',
  layer: item.layer ? item.layer as ItemFormValues['layer'] : '',
  waterproof: toBooleanValue(item.waterproof),
  size: item.size ?? '',
  color: item.color ?? '',
  capacity_people: toIntegerString(item.capacity_people),
  season_rating: item.season_rating ? item.season_rating as ItemFormValues['season_rating'] : '',
  freestanding: toBooleanValue(item.freestanding),
  has_footprint: toBooleanValue(item.has_footprint),
  comfort_temp_c: toNumberString(item.comfort_temp_c),
  fill_type: item.fill_type ? item.fill_type as ItemFormValues['fill_type'] : '',
  r_value: toNumberString(item.r_value),
  capacity_mah: toIntegerString(item.capacity_mah),
  charge_port: item.charge_port ? item.charge_port as ItemFormValues['charge_port'] : '',
  rechargeable: toBooleanValue(item.rechargeable),
  image_blob: item.image_blob ?? '',
  image_mime_type: item.image_mime_type ?? '',
  image_size_bytes: toIntegerString(item.image_size_bytes),
  attributes: toFormAttributes(item.attributes),
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

const applyConsumablePayloadFields = (payload: ItemCreate | ItemUpdate, values: ItemFormValues) => {
  const parsedDoseCount = parseInteger(values.dose_count)
  if (parsedDoseCount !== undefined) {
    payload.dose_count = parsedDoseCount
  }

  const parsedCalories = parseInteger(values.calories)
  if (parsedCalories !== undefined) {
    payload.calories = parsedCalories
  }

  if (parsedCalories !== undefined && parsedDoseCount !== undefined && parsedDoseCount > 0) {
    payload.calories_per_serving = Math.round((parsedCalories / parsedDoseCount) * 100) / 100
  }

  payload.requires_water = values.requires_water
}

const applyWearablePayloadFields = (payload: ItemCreate | ItemUpdate, values: ItemFormValues) => {
  if (values.season) {
    payload.season = values.season
  }
  if (values.layer) {
    payload.layer = values.layer
  }
  payload.waterproof = values.waterproof
  if (values.size.trim()) {
    payload.size = values.size.trim()
  }
  if (values.color.trim()) {
    payload.color = values.color.trim()
  }
}

const applyShelterPayloadFields = (payload: ItemCreate | ItemUpdate, values: ItemFormValues) => {
  const parsedCapacityPeople = parseNumber(values.capacity_people)
  if (parsedCapacityPeople !== undefined) {
    payload.capacity_people = parsedCapacityPeople
  }
  if (values.season_rating) {
    payload.season_rating = values.season_rating
  }
  payload.freestanding = values.freestanding
  payload.has_footprint = values.has_footprint
}

const applySleepPayloadFields = (payload: ItemCreate | ItemUpdate, values: ItemFormValues) => {
  const parsedComfortTempC = parseNumber(values.comfort_temp_c)
  if (parsedComfortTempC !== undefined) {
    payload.comfort_temp_c = parsedComfortTempC
  }
  if (values.fill_type) {
    payload.fill_type = values.fill_type
  }
  const parsedRValue = parseNumber(values.r_value)
  if (parsedRValue !== undefined) {
    payload.r_value = parsedRValue
  }
}

const applyElectronicsPayloadFields = (payload: ItemCreate | ItemUpdate, values: ItemFormValues) => {
  const parsedCapacityMah = parseInteger(values.capacity_mah)
  if (parsedCapacityMah !== undefined) {
    payload.capacity_mah = parsedCapacityMah
  }
  if (values.charge_port) {
    payload.charge_port = values.charge_port
  }
  payload.rechargeable = values.rechargeable
}

const applyTypeSpecificPayloadFields = (payload: ItemCreate | ItemUpdate, values: ItemFormValues) => {
  switch (values.type) {
    case 'consumable':
      applyConsumablePayloadFields(payload, values)
      break
    case 'wearable':
      applyWearablePayloadFields(payload, values)
      break
    case 'shelter':
      applyShelterPayloadFields(payload, values)
      break
    case 'sleep':
      applySleepPayloadFields(payload, values)
      break
    case 'electronics':
      applyElectronicsPayloadFields(payload, values)
      break
    default:
      break
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
  if (isKnownItemType(values.type) || fields.length === 0) {
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
  applyTypeSpecificPayloadFields(payload, values)

  const dynamicAttributes = toDynamicAttributes(values, dynamicFields)
  if (dynamicAttributes !== undefined) {
    payload.attributes = dynamicAttributes
  }

  return payload
}

const formatDisplayWeight = (valueGrams: number): string => {
  if (weightInputUnit.value === 'oz') {
    const ounces = gramsToInput(valueGrams, 'oz')
    if (Math.abs(ounces) >= OUNCES_PER_POUND) {
      return `${toRoundedString(ounces / OUNCES_PER_POUND)} lb`
    }

    return `${toRoundedString(ounces)} oz`
  }

  if (Math.abs(valueGrams) >= GRAMS_PER_KILOGRAM) {
    return `${toRoundedString(valueGrams / GRAMS_PER_KILOGRAM)} kg`
  }

  return `${toRoundedString(valueGrams)} g`
}

const formatWeight = (value?: number | null) => {
  if (typeof value !== 'number') {
    return 'Not set'
  }

  return formatDisplayWeight(value)
}

const formatTotalWeight = (valueGrams: number): string => {
  return formatDisplayWeight(valueGrams)
}

const formatVolume = (value?: number | null) => {
  if (typeof value !== 'number') {
    return 'Not set'
  }

  return `${toRoundedString(mlToInput(value, volumeInputUnit.value))} ${volumeInputLabel.value}`
}

const formatNumber = (value?: number | null) => {
  if (typeof value !== 'number') {
    return 'Not set'
  }

  return toRoundedString(value)
}

const formatValue = (value?: number | null): string => {
  if (!value) {
    return 'Not set'
  }
  const formatted = formatNumber(value)
  const currencySymbol = currency.value === 'usd' ? '$' : '€'
  return `${formatted} ${currencySymbol}`
}

const formatCarryStatus = (value?: string | null) => {
  if (!value) {
    return 'Not set'
  }

  if (value === 'packed') {
    return 'Packed'
  }
  if (value === 'worn') {
    return 'Worn'
  }
  return value
}

const formatType = (value: string) => {
  return value
    .split('_')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}

const formatText = (value?: string | null) => {
  if (!value?.trim()) {
    return 'Not set'
  }

  return value
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
    { label: 'Value', value: typeof item.value === 'number' ? formatNumber(item.value) : '' },
  ]

  return entries.filter((entry) => entry.value)
}

const getCardStatEntries = (item: Item) => {
  return getCardSummaryEntries(item).filter((entry) => entry.label === 'Weight' || entry.label === 'Volume' || entry.label === 'Value')
}

const getBaseDetails = (item: Item) => {
  return [
    { label: 'Type', value: formatType(item.type) },
    { label: 'Manufacturer', value: manufacturersById.value.get(item.manufacturer_id) ?? item.manufacturer_id },
    { label: 'Active', value: '', booleanValue: item.is_active },
    { label: 'Default carry', value: formatCarryStatus(item.default_carry_status) },
    { label: 'Default quantity', value: formatNumber(item.default_quantity) },
    { label: 'Is default', value: '', booleanValue: item.is_default },
    { label: 'Weight', value: formatWeight(item.weight_grams) },
    { label: 'Volume', value: formatVolume(item.volume_ml) },
    { label: 'Value', value: formatValue(item.value) },
  ]
}

const getDetailedEntries = (item: Item) => {
  const entries: DetailEntry[] = [
    ...getBaseDetails(item),
    { label: 'Description', value: formatText(item.description) },
    { label: 'URL', value: item.source_url?.trim() ? 'URL' : 'Not set', href: item.source_url ?? undefined },
  ]

  if (item.type === 'consumable') {
    entries.push(
      { label: 'Dose count', value: formatNumber(item.dose_count) },
      { label: 'Calories', value: formatNumber(item.calories) },
      { label: 'Calories per serving', value: formatNumber(item.calories_per_serving) },
      { label: 'Requires water', value: '', booleanValue: item.requires_water },
    )
  }

  if (item.type === 'wearable') {
    entries.push(
      { label: 'Season', value: formatText(item.season) },
      { label: 'Layer', value: formatText(item.layer) },
      { label: 'Size', value: formatText(item.size) },
      { label: 'Color', value: formatText(item.color) },
      { label: 'Waterproof', value: '', booleanValue: item.waterproof },
    )
  }

  if (item.type === 'shelter') {
    entries.push(
      { label: 'Capacity people', value: formatNumber(item.capacity_people) },
      { label: 'Season rating', value: formatText(item.season_rating) },
      { label: 'Freestanding', value: '', booleanValue: item.freestanding },
      { label: 'Has footprint', value: '', booleanValue: item.has_footprint },
    )
  }

  if (item.type === 'sleep') {
    entries.push(
      { label: 'Comfort temp C', value: formatNumber(item.comfort_temp_c) },
      { label: 'Fill type', value: formatText(item.fill_type) },
      { label: 'R value', value: formatNumber(item.r_value) },
    )
  }

  if (item.type === 'electronics') {
    entries.push(
      { label: 'Capacity mAh', value: formatNumber(item.capacity_mah) },
      { label: 'Charge port', value: formatText(item.charge_port) },
      { label: 'Rechargeable', value: '', booleanValue: item.rechargeable },
    )
  }

  if (item.image_mime_type || item.image_size_bytes || item.image_width_px || item.image_height_px) {
    entries.push(
      { label: 'Image type', value: formatText(item.image_mime_type) },
      { label: 'Image size bytes', value: formatNumber(item.image_size_bytes) },
      { label: 'Image width px', value: formatNumber(item.image_width_px) },
      { label: 'Image height px', value: formatNumber(item.image_height_px) },
    )
  }

  return entries
}

const settingsQuery = useQuery({
  queryKey: ['settings'],
  queryFn: getSettings,
})

const weightInputUnit = computed<WeightInputUnit>(() => settingsQuery.data.value?.weight_unit ?? 'g')
const volumeInputUnit = computed<VolumeInputUnit>(() => settingsQuery.data.value?.volume_unit ?? 'ml')
const currency = computed<Currency>(() => settingsQuery.data.value?.currency ?? 'usd')
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

const manufacturersById = computed(() => {
  const map = new Map<string, string>()
  for (const manufacturer of manufacturersQuery.data.value ?? []) {
    map.set(manufacturer.id, normalizeTitleWords(manufacturer.name))
  }

  return map
})

const createValues = ref<ItemFormValues>(emptyFormValues())
const editingItemId = ref<string | null>(null)
const editValues = ref<ItemFormValues>(emptyFormValues())
const isFormDialogOpen = ref(false)
const itemsViewMode = ref<ItemsViewMode>(readStoredItemsViewMode())
const itemsTableDetailModeByType = ref<Record<string, ItemsTableDetailMode>>(readStoredItemsTableDetailModeByType())
const itemsTableSelectionModeByType = ref<Record<string, boolean>>({})
const selectedTableItemIdsByType = ref<Record<string, string[]>>({})
const selectedDetailItemId = ref<string | null>(null)
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
  { value: 'item', label: 'Create', description: 'Add a new gear item.', icon: 'pi pi-box' },
  { value: 'manufacturer', label: 'Manage Manufacturers', description: 'Create and edit manufacturers.', icon: 'pi pi-building' },
  { value: 'category', label: 'Manage Categories', description: 'Create and edit custom categories.', icon: 'pi pi-tag' },
  { value: 'import', label: 'Import CSV', description: 'Preview and import gear from CSV.', icon: 'pi pi-upload' },
]

const itemFormTypeOptions = computed<Array<{ label: string; value: string }>>(() => {
  const fetched = itemTypesQuery.data.value ?? []
  if (fetched.length > 0) {
    return fetched
      .map((itemType) => ({ label: normalizeTitleWords(itemType.name), value: itemType.id }))
      .sort((left, right) => left.label.localeCompare(right.label))
  }

  return KNOWN_ITEM_TYPES.map((value) => ({
    value,
    label: formatType(value),
  }))
})

const filteredItems = computed(() => {
  const items = itemsQuery.data.value ?? []
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
      case 'requires_water': return item.requires_water
      case 'waterproof': return item.waterproof
      case 'freestanding': return item.freestanding
      case 'has_footprint': return item.has_footprint
      case 'rechargeable': return item.rechargeable
      default: return undefined
    }
  }

  const renderHref = (item: Item): string | undefined => {
    if (key !== 'source_url') {
      return undefined
    }

    return item.source_url?.trim() ? item.source_url : undefined
  }

  const render = (item: Item): string => {
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
      case 'dose_count': return formatNumber(item.dose_count)
      case 'calories': return formatNumber(item.calories)
      case 'calories_per_serving': return formatNumber(item.calories_per_serving)
      case 'requires_water': return ''
      case 'season': return formatText(item.season)
      case 'layer': return formatText(item.layer)
      case 'waterproof': return ''
      case 'size': return formatText(item.size)
      case 'color': return formatText(item.color)
      case 'capacity_people': return formatNumber(item.capacity_people)
      case 'season_rating': return formatText(item.season_rating)
      case 'freestanding': return ''
      case 'has_footprint': return ''
      case 'comfort_temp_c': return formatNumber(item.comfort_temp_c)
      case 'fill_type': return formatText(item.fill_type)
      case 'r_value': return formatNumber(item.r_value)
      case 'capacity_mah': return formatNumber(item.capacity_mah)
      case 'charge_port': return formatText(item.charge_port)
      case 'rechargeable': return ''
      default: return 'Not set'
    }
  }

  return { key, label, render, renderHref, renderBoolean }
}

const commonTableFieldDefinitions = commonTableFieldOptions.map(createTableFieldDefinition)
const extraTableFieldDefinitionsByType: Record<KnownItemType, TableFieldDefinition[]> = {
  consumable: extraTableFieldOptionsByType.consumable.map(createTableFieldDefinition),
  wearable: extraTableFieldOptionsByType.wearable.map(createTableFieldDefinition),
  shelter: extraTableFieldOptionsByType.shelter.map(createTableFieldDefinition),
  sleep: extraTableFieldOptionsByType.sleep.map(createTableFieldDefinition),
  electronics: extraTableFieldOptionsByType.electronics.map(createTableFieldDefinition),
}

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
      const knownType = isKnownItemType(type) ? type : null
      return {
        type,
        title: formatType(type),
        items: groupedItems,
        baseFields: commonTableFieldDefinitions,
        extraFields: knownType ? extraTableFieldDefinitionsByType[knownType] : [],
        tableDetailMode: itemsTableDetailModeByType.value[type] ?? 'simple',
        selectionMode: itemsTableSelectionModeByType.value[type] ?? false,
        selectedItemIds: selectedTableItemIdsByType.value[type] ?? [],
        totalWeightLabel: formatTotalWeight(getItemsTotalWeightGrams(groupedItems)),
        totalValueLabel: formatTotalValue(getItemsTotalValue(groupedItems)),
      }
    })
    .filter((section) => section.items.length > 0)
})

const updateTableDetailMode = (type: string, mode: ItemsTableDetailMode) => {
  itemsTableDetailModeByType.value = {
    ...itemsTableDetailModeByType.value,
    [type]: mode,
  }
}

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
  await queryClient.invalidateQueries({ queryKey: ['items'] })
}

const onRowToggleActive = async (item: Item) => {
  try {
    await withItemsRefresh(async () => {
      await updateItem(item.id, { is_active: !item.is_active })
    })
    toast.add({
      severity: 'success',
      summary: 'Item updated',
      detail: `Item is now ${item.is_active ? 'inactive' : 'active'}.`,
      life: 2500,
    })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: 'Update failed',
      detail: error instanceof Error ? error.message : 'Unable to update item.',
      life: 3500,
    })
  }
}

const onRowToggleDefault = async (item: Item) => {
  try {
    await withItemsRefresh(async () => {
      await updateItem(item.id, { is_default: !item.is_default })
    })
    toast.add({
      severity: 'success',
      summary: 'Item updated',
      detail: item.is_default ? 'Item is no longer default.' : 'Item is now default.',
      life: 2500,
    })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: 'Update failed',
      detail: error instanceof Error ? error.message : 'Unable to update item.',
      life: 3500,
    })
  }
}

const onRowDuplicate = async (item: Item) => {
  try {
    const duplicateValues = toFormValues(item)
    duplicateValues.name = `${item.name} Copy`
    const payload = toPayload(duplicateValues, []) as ItemCreate
    if (!isKnownItemType(item.type) && item.attributes) {
      payload.attributes = item.attributes
    }

    await withItemsRefresh(async () => {
      await createItem(payload)
    })

    toast.add({
      severity: 'success',
      summary: 'Item duplicated',
      detail: 'A copy was created successfully.',
      life: 3000,
    })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: 'Duplicate failed',
      detail: error instanceof Error ? error.message : 'Unable to duplicate item.',
      life: 3500,
    })
  }
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

const createMutation = useMutation({
  mutationFn: createItem,
  onSuccess: async () => {
    createValues.value = emptyFormValues()
    isFormDialogOpen.value = false
    await queryClient.invalidateQueries({ queryKey: ['items'] })
    toast.add({
      severity: 'success',
      summary: 'Item created',
      detail: 'New item has been added.',
      life: 3000,
    })
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Create failed',
      detail: error instanceof Error ? error.message : 'Unable to create item.',
      life: 3500,
    })
  },
})

const updateMutation = useMutation({
  mutationFn: async (params: { itemId: string; payload: ItemUpdate }) => {
    return updateItem(params.itemId, params.payload)
  },
  onSuccess: async () => {
    editingItemId.value = null
    editValues.value = emptyFormValues()
    isFormDialogOpen.value = false
    await queryClient.invalidateQueries({ queryKey: ['items'] })
    toast.add({
      severity: 'success',
      summary: 'Item updated',
      detail: 'Item details were saved.',
      life: 3000,
    })
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Update failed',
      detail: error instanceof Error ? error.message : 'Unable to update item.',
      life: 3500,
    })
  },
})

const deleteMutation = useMutation({
  mutationFn: removeItem,
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['items'] })
    toast.add({
      severity: 'success',
      summary: 'Item deleted',
      detail: 'Item was removed successfully.',
      life: 3000,
    })
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Delete failed',
      detail: error instanceof Error ? error.message : 'Unable to delete item.',
      life: 3500,
    })
  },
})

const canShowEmptyState = computed(() => {
  return !itemsQuery.isPending.value && !itemsQuery.isError.value && (itemsQuery.data.value?.length ?? 0) === 0
})

const hasFilteredItems = computed(() => filteredItems.value.length > 0)

const selectedDetailItem = computed(() => {
  if (!selectedDetailItemId.value) {
    return null
  }

  return itemsQuery.data.value?.find((item) => item.id === selectedDetailItemId.value) ?? null
})

const isDetailDialogOpen = computed({
  get: () => selectedDetailItem.value !== null,
  set: (value: boolean) => {
    if (!value) {
      selectedDetailItemId.value = null
    }
  },
})

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
const formSubmitLabel = computed(() => (isCreateMode.value ? 'Create' : 'Save Changes'))
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

watch(
  () => route.query.create,
  () => {
    void consumeCreateQuery()
  },
  { immediate: true },
)

watch(itemsViewMode, (value) => {
  if (globalThis.window === undefined) {
    return
  }

  globalThis.localStorage.setItem(ITEMS_VIEW_MODE_STORAGE_KEY, value)
})

watch(itemTypeFilter, (value) => {
  if (globalThis.window === undefined) {
    return
  }

  globalThis.localStorage.setItem(ITEMS_TYPE_FILTER_STORAGE_KEY, value)
})

watch(itemsTableDetailModeByType, (value) => {
  if (globalThis.window === undefined) {
    return
  }

  globalThis.localStorage.setItem(ITEMS_TABLE_DETAIL_MODE_BY_TYPE_STORAGE_KEY, JSON.stringify(value))
}, { deep: true })

watch(isFormDialogOpen, (value) => {
  if (!value) {
    onCancelEdit()
  }
})

const onCreate = async () => {
  const payload = toPayload(createValues.value, itemFormDynamicFields.value) as ItemCreate
  await createMutation.mutateAsync(payload)
}

const onStartEdit = (item: Item) => {
  editingItemId.value = item.id
  editValues.value = toFormValues(item)
  selectedDetailItemId.value = null
  isManufacturerDialogOpen.value = false
  isCategoryDialogOpen.value = false
  closeCreateOptions()
  isFormDialogOpen.value = true
}

const onCancelEdit = () => {
  isFormDialogOpen.value = false

  if (isCreateMode.value) {
    createValues.value = emptyFormValues()
    return
  }

  editingItemId.value = null
  editValues.value = emptyFormValues()
}

const onSaveEdit = async () => {
  if (!editingItemId.value) {
    return
  }

  const payload = toPayload(editValues.value, itemFormDynamicFields.value)
  await updateMutation.mutateAsync({ itemId: editingItemId.value, payload })
}

const onDelete = async (itemId: string) => {
  await deleteMutation.mutateAsync(itemId)
}

const openDetails = (item: Item) => {
  selectedDetailItemId.value = item.id
}

const onEditFromDetails = () => {
  if (!selectedDetailItem.value) {
    return
  }

  const item = selectedDetailItem.value
  selectedDetailItemId.value = null
  onStartEdit(item)
}

const onDeleteFromDetails = async (itemId: string) => {
  selectedDetailItemId.value = null
  await onDelete(itemId)
}

const onSubmitForm = async () => {
  if (isCreateMode.value) {
    await onCreate()
    return
  }

  await onSaveEdit()
}
</script>

<template>
  <section data-component="items-page" class="flex w-full flex-col gap-4">
    <AppConfirmDialog :open="confirmDialogState !== null" title="Confirm delete" :message="confirmDialogMessage"
      confirm-label="Delete" confirm-tone="danger" @update:open="(value) => { if (!value) closeConfirmDialog() }"
      @cancel="closeConfirmDialog" @confirm="onConfirmDelete" />

    <ItemDetailsDialog :open="isDetailDialogOpen" :selected-item="selectedDetailItem" :get-image-src="getItemImageSrc"
      :get-detailed-entries="getDetailedEntries" :format-type="formatType" :manufacturers-by-id="manufacturersById"
      :is-delete-loading="deleteMutation.isPending.value" @update:open="isDetailDialogOpen = $event"
      @edit="onEditFromDetails" @delete="onDeleteFromDetails" />

    <ItemFormDialog :open="isFormDialogOpen" :is-create-mode="isCreateMode" :title="formTitle"
      :submit-label="formSubmitLabel" :values="activeFormValues" :item-type-options="itemFormTypeOptions"
      :dynamic-fields="itemFormDynamicFields" :dynamic-fields-loading="itemTypeFieldsQuery.isPending.value"
      :manufacturers="manufacturersQuery.data.value ?? []" :weight-input-label="weightInputLabel"
      :volume-input-label="volumeInputLabel" :is-loading="formLoading" @update:open="isFormDialogOpen = $event"
      @update:values="(values) => { activeFormValues = values }" @request:manufacturer-create="openManufacturerDialog"
      @submit="onSubmitForm" @cancel="onCancelEdit" />

    <ItemsManufacturerDialog v-model:open="isManufacturerDialogOpen"
      :manufacturers="manufacturersQuery.data.value ?? []" :items="itemsQuery.data.value ?? []"
      @manufacturer-created="(id) => { activeFormValues = { ...activeFormValues, manufacturer_id: id } }" />

    <ItemsCategoryDialog v-model:open="isCategoryDialogOpen" :items="itemsQuery.data.value ?? []" />

    <ItemsImportDialog v-model:open="isImportDialogOpen" />

    <Message v-if="itemsQuery.isError.value" data-element="items-error" severity="error" :closable="false">
      {{ itemsQuery.error.value instanceof Error ? itemsQuery.error.value.message : 'Unable to load gear.' }}
    </Message>

    <Message v-if="manufacturersQuery.isError.value" data-element="items-manufacturers-error" severity="error"
      :closable="false">
      {{ manufacturersQuery.error.value instanceof Error ? manufacturersQuery.error.value.message :
        'Unable to load manufacturers.' }}
    </Message>

    <ItemsCreateOptionsMenu :open="isCreateOptionsOpen" :position="createOptionsPosition" :options="createTargetOptions"
      @update:open="isCreateOptionsOpen = $event" @select="onSelectCreateTarget" />

    <div v-if="itemsQuery.isPending.value" data-element="items-loading"
      class="border-line-subtle bg-surface-muted text-copy-muted rounded-2xl border px-4 py-3 text-sm font-medium">
      Loading gear...
    </div>

    <div v-else-if="canShowEmptyState" data-element="items-empty-state"
      class="border-line-subtle bg-surface-elevated text-copy-muted rounded-2xl border px-5 py-6 text-sm">
      Your closet is currently operating at true ultralight standards. Add some gear to get started!
    </div>

    <div v-else data-element="items-list" class="space-y-3">
      <div data-element="items-summary" class="grid gap-2 sm:grid-cols-3">
        <div class="border-line-subtle bg-surface-elevated text-copy rounded-xl border px-3 py-2 text-sm">
          <span class="text-copy-subtle text-xs font-semibold uppercase tracking-[0.08em]">Total gear</span>
          <p class="text-ink mt-1 text-base font-semibold">{{ filteredSummary.totalItems }}</p>
        </div>
        <div class="border-line-subtle bg-surface-elevated text-copy rounded-xl border px-3 py-2 text-sm">
          <span class="text-copy-subtle text-xs font-semibold uppercase tracking-[0.08em]">Total weight</span>
          <p class="text-ink mt-1 text-base font-semibold">{{ filteredSummary.totalWeightLabel }}</p>
        </div>
        <div class="border-line-subtle bg-surface-elevated text-copy rounded-xl border px-3 py-2 text-sm">
          <span class="text-copy-subtle text-xs font-semibold uppercase tracking-[0.08em]">Total value</span>
          <p class="text-ink mt-1 text-base font-semibold">{{ filteredSummary.totalValueLabel }}</p>
        </div>
      </div>

      <!-- Click-outside backdrop -->
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div class="flex flex-wrap items-center gap-3">
          <!-- View Mode Toggle -->
          <AppToggleGroup name="items-view-mode" data-element="items-view-mode" :model-value="itemsViewMode"
            :options="itemViewOptions" fit-content
            @update:model-value="(value) => { itemsViewMode = value as ItemsViewMode }" />

          <nav data-element="items-type-filter" aria-label="Item type filter"
            class="text-copy-subtle flex flex-wrap items-center text-xs font-semibold uppercase tracking-[0.08em]">
            <template v-for="(option, index) in itemTypeFilterOptions" :key="option.value">
              <span v-if="index > 0" class="text-line mx-1">/</span>
              <button type="button" class="rounded px-2 py-1 transition"
                :class="itemTypeFilter === option.value ? 'bg-brand-50 text-brand-800' : 'text-copy-subtle hover:bg-surface-soft hover:text-copy'"
                @click="itemTypeFilter = option.value">
                {{ option.label }}
              </button>
            </template>
          </nav>
        </div>
      </div>

      <div v-if="!hasFilteredItems" data-element="items-filter-empty-state"
        class="border-line-subtle bg-surface-elevated text-copy-muted rounded-2xl border px-5 py-6 text-sm">
        No gear matches the selected type filter.
      </div>

      <ItemsListView v-else :view-mode="itemsViewMode" :items="filteredItems" :table-sections="itemTableSections"
        :get-image-src="getItemImageSrc" @open-details="openDetails"
        @update:table-detail-mode="(type, mode) => updateTableDetailMode(type, mode)"
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
