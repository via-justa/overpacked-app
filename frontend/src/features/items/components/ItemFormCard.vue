<script setup lang="ts">
import { computed } from 'vue'
import Button from 'primevue/button'
import AppSelect from '../../../components/AppSelect.vue'
import AppToggleGroup from '../../../components/AppToggleGroup.vue'
import ItemLabelsSelector from './ItemLabelsSelector.vue'
import type {
  ChargePort,
  DefaultCarryStatus,
  ItemFormValues,
  ItemTypeField,
  ItemLayer,
  ItemSeason,
  Label,
  Manufacturer,
  SeasonRating,
  SleepFillType,
} from '../types'
import { isKnownItemType } from '../types'

const props = defineProps<{
  title: string
  submitLabel: string
  values: ItemFormValues
  itemTypeOptions: Array<{ label: string; value: string }>
  dynamicFields: ItemTypeField[]
  dynamicFieldsLoading?: boolean
  manufacturers: Manufacturer[]
  weightInputLabel: 'g' | 'oz'
  volumeInputLabel: 'ml' | 'fl oz'
  loading?: boolean
  showCancel?: boolean
  showButtons?: boolean
  bare?: boolean
  allLabels: Label[]
  selectedLabels: Label[]
  labelsLoading?: boolean
}>()

const emit = defineEmits<{
  submit: []
  cancel: []
  'request:manufacturer-create': []
  'update:values': [values: ItemFormValues]
  'label:add': [label: Label]
  'label:remove': [labelId: string]
  'label:create': [name: string]
}>()

const carryStatusOptions: Array<{ label: string; value: DefaultCarryStatus }> = [
  { label: 'Packed', value: 'packed' },
  { label: 'Worn', value: 'worn' },
]

const seasonOptions: Array<{ label: string; value: ItemSeason }> = [
  { label: 'Summer', value: 'summer' },
  { label: 'Winter', value: 'winter' },
  { label: 'Year round', value: 'year_round' },
]

const layerOptions: Array<{ label: string; value: ItemLayer }> = [
  { label: 'Base', value: 'base' },
  { label: 'Mid', value: 'mid' },
  { label: 'Shell', value: 'shell' },
  { label: 'Accessory', value: 'accessory' },
]

const seasonRatingOptions: Array<{ label: string; value: SeasonRating }> = [
  { label: '3-season', value: '3-season' },
  { label: '4-season', value: '4-season' },
]

const fillTypeOptions: Array<{ label: string; value: SleepFillType }> = [
  { label: 'Down', value: 'down' },
  { label: 'Synthetic', value: 'synthetic' },
  { label: 'Foam', value: 'foam' },
  { label: 'Air', value: 'air' },
  { label: 'Other', value: 'other' },
]

const chargePortOptions: Array<{ label: string; value: ChargePort }> = [
  { label: 'USB-C', value: 'usb-c' },
  { label: 'Micro-USB', value: 'micro-usb' },
  { label: 'Lightning', value: 'lightning' },
  { label: 'DC', value: 'dc' },
]

const activeOptions = [
  { label: 'Active', value: 'active' },
  { label: 'Inactive', value: 'inactive' },
]

const yesNoOptions = [
  { label: 'Yes', value: 'yes' },
  { label: 'No', value: 'no' },
]

const canSubmit = computed(() => {
  return props.values.name.trim().length > 0 && props.values.manufacturer_id.length > 0
})

const activeValue = computed(() => (props.values.is_active ? 'active' : 'inactive'))
const defaultCarryValue = computed(() => props.values.default_carry_status)
const defaultQuantityOptions = computed<number[]>(() => {
  const base = Array.from({ length: 10 }, (_, index) => index + 1)
  const parsedCurrent = Number(props.values.default_quantity)

  if (Number.isNaN(parsedCurrent) || base.includes(parsedCurrent)) {
    return base
  }

  return [...base, parsedCurrent].sort((a, b) => a - b)
})

const doseCountOptions = new Set(Array.from({ length: 10 }, (_, index) => String(index + 1)))
const capacityPeopleOptions = new Set(['1', '2', '3'])
const sizeOptions = new Set(['XS', 'S', 'M', 'L', 'XL', 'XXL'])

function validURL(str: string): boolean {
  try {
    const parsed = new URL(str)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

const isValidUrl = (value: string): boolean => {
  if (!value.trim()) {
    return true
  }

  return validURL(value)
}

const isFloat = (value: string): boolean => {
  if (!value.trim()) {
    return true
  }

  const parsed = Number(value)
  return !Number.isNaN(parsed)
}

const isInteger = (value: string): boolean => {
  if (!value.trim()) {
    return true
  }

  return /^-?\d+$/.test(value.trim())
}

const caloriesPerServing = computed<string>(() => {
  const calories = Number(props.values.calories)
  const doseCount = Number(props.values.dose_count)

  if (Number.isNaN(calories) || Number.isNaN(doseCount) || doseCount <= 0) {
    return ''
  }

  return String(Math.round((calories / doseCount) * 100) / 100)
})

const appendBaseValidationErrors = (errors: Record<string, string>) => {
  if (!isValidUrl(props.values.source_url)) {
    errors.source_url = 'Enter a valid URL (http/https).'
  }

  if (!isFloat(props.values.value)) {
    errors.value = 'Value must be a valid float number.'
  }

  if (!isInteger(props.values.weight_value)) {
    errors.weight_value = 'Weight must be a whole number.'
  }

  if (!isFloat(props.values.volume_value)) {
    errors.volume_value = 'Volume must be a valid float number.'
  }
}

const appendTypeValidationErrors = (errors: Record<string, string>) => {
  if (props.values.type === 'consumable') {
    if (props.values.dose_count && !doseCountOptions.has(props.values.dose_count)) {
      errors.dose_count = 'Dose count must be between 1 and 10.'
    }

    if (!isInteger(props.values.calories)) {
      errors.calories = 'Calories must be a whole number.'
    }
  }

  if (props.values.type === 'wearable' && props.values.size && !sizeOptions.has(props.values.size)) {
    errors.size = 'Size must be one of XS, S, M, L, XL, XXL.'
  }

  if (props.values.type === 'shelter' && props.values.capacity_people && !capacityPeopleOptions.has(props.values.capacity_people)) {
    errors.capacity_people = 'Capacity people must be between 1 and 3.'
  }

  if (props.values.type === 'sleep' && !isFloat(props.values.r_value)) {
    errors.r_value = 'R value must be a valid float number.'
  }

  if (props.values.type === 'electronics' && !isInteger(props.values.capacity_mah)) {
    errors.capacity_mah = 'Capacity mAh must be a whole number.'
  }
}

const validateDynamicField = (field: ItemTypeField, rawValue: string | boolean | undefined): string | null => {
  if (field.data_type === 'boolean') {
    if (field.is_required && typeof rawValue !== 'boolean') {
      return `${field.field_label} is required.`
    }

    return null
  }

  const value = typeof rawValue === 'string' ? rawValue.trim() : ''

  if (field.is_required && !value) {
    return `${field.field_label} is required.`
  }

  if (!value) {
    return null
  }

  if (field.data_type === 'integer' && !isInteger(value)) {
    return `${field.field_label} must be a whole number.`
  }

  if (field.data_type === 'number' && !isFloat(value)) {
    return `${field.field_label} must be a valid number.`
  }

  if (field.data_type === 'enum' && field.enum_options && field.enum_options.length > 0 && !field.enum_options.includes(value)) {
    return `Choose a valid option for ${field.field_label}.`
  }

  return null
}

const appendDynamicValidationErrors = (errors: Record<string, string>) => {
  if (isKnownItemType(props.values.type)) {
    return
  }

  for (const field of props.dynamicFields) {
    const fieldErrorKey = `attributes.${field.field_key}`
    const rawValue = props.values.attributes[field.field_key]

    const message = validateDynamicField(field, rawValue)
    if (message) {
      errors[fieldErrorKey] = message
    }
  }
}

const validationErrors = computed<Record<string, string>>(() => {
  const errors: Record<string, string> = {}

  appendBaseValidationErrors(errors)
  appendTypeValidationErrors(errors)
  appendDynamicValidationErrors(errors)

  return errors
})

const imagePreviewSrc = computed(() => {
  if (!props.values.image_blob) {
    return ''
  }

  const mimeType = props.values.image_mime_type || 'image/*'
  return `data:${mimeType};base64,${props.values.image_blob}`
})

const updateField = <K extends keyof ItemFormValues>(key: K, value: ItemFormValues[K]) => {
  emit('update:values', {
    ...props.values,
    [key]: value,
  })
}

const updateFields = (partialValues: Partial<ItemFormValues>) => {
  emit('update:values', {
    ...props.values,
    ...partialValues,
  })
}

const onImageChange = (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]

  if (!file) {
    return
  }

  if (!file.type.startsWith('image/')) {
    input.value = ''
    return
  }

  const reader = new FileReader()
  reader.onload = () => {
    const result = typeof reader.result === 'string' ? reader.result : ''
    const base64 = result.includes(',') ? result.split(',')[1] ?? '' : ''

    updateFields({
      image_blob: base64,
      image_mime_type: file.type,
      image_size_bytes: String(file.size),
    })
  }

  reader.readAsDataURL(file)
  input.value = ''
}

const clearImage = () => {
  updateFields({
    image_blob: '',
    image_mime_type: '',
    image_size_bytes: '',
  })
}

const onTypeChange = (nextType: string) => {
  updateFields({
    type: nextType,
    attributes: {},
  })
}

const onManufacturerChange = (nextManufacturerId: string) => {
  if (nextManufacturerId === '__create_new__') {
    emit('request:manufacturer-create')
    return
  }

  updateField('manufacturer_id', nextManufacturerId)
}

const getAttributeStringValue = (fieldKey: string): string => {
  const value = props.values.attributes[fieldKey]
  return typeof value === 'string' ? value : ''
}

const getAttributeBooleanValue = (fieldKey: string): boolean => {
  return props.values.attributes[fieldKey] === true
}

const updateAttributeValue = (fieldKey: string, value: string | boolean) => {
  updateField('attributes', {
    ...props.values.attributes,
    [fieldKey]: value,
  })
}

const getDynamicFieldError = (fieldKey: string): string => {
  return validationErrors.value[`attributes.${fieldKey}`] ?? ''
}

const onSubmit = () => {
  if (!canSubmit.value || props.loading || Object.keys(validationErrors.value).length > 0) {
    return
  }

  emit('submit')
}

const onCancel = () => {
  emit('cancel')
}
</script>

<template>
  <section data-component="item-form-card"
    :class="bare
      ? 'flex flex-col'
      : 'border-line-subtle bg-surface-elevated flex max-h-[calc(100vh-8rem)] flex-col rounded-2xl border p-4 shadow-panel backdrop-blur sm:p-5'">
    <h2 v-if="!bare" class="text-ink shrink-0 text-lg font-semibold">{{ title }}</h2>

    <div class="mt-4 flex-1 overflow-y-auto pr-1">
      <div class="space-y-4">
        <div class="grid gap-3 md:grid-cols-2">
          <label class="grid gap-1">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Name</span>
            <input data-element="item-name" class="input-shell" :value="values.name" type="text"
              @input="updateField('name', ($event.target as HTMLInputElement).value)" />
          </label>

          <div class="grid gap-1">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Status</span>
            <AppToggleGroup name="item-active" data-element="item-active" :model-value="activeValue"
              :options="activeOptions" fit-content
              @update:model-value="(value) => updateField('is_active', value === 'active')" />
          </div>

          <div class="grid gap-1">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Manufacturer</span>
            <AppSelect data-element="item-manufacturer" :model-value="values.manufacturer_id"
              @update:model-value="onManufacturerChange">
              <option value="">Select manufacturer</option>
              <option v-for="manufacturer in manufacturers" :key="manufacturer.id" :value="manufacturer.id">
                {{ manufacturer.name }}
              </option>
              <option value="__create_new__">+ Create new manufacturer</option>
            </AppSelect>
          </div>

          <label class="grid gap-1 md:col-start-2 md:row-span-2">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Description</span>
            <textarea data-element="item-description" class="h-full min-h-7.5rem input-shell"
              :value="values.description"
              @input="updateField('description', ($event.target as HTMLTextAreaElement).value)" />
          </label>

          <div class="grid gap-1">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Type</span>
            <AppSelect data-element="item-type" :model-value="values.type" @update:model-value="onTypeChange">
              <option v-for="option in itemTypeOptions" :key="option.value" :value="option.value">
                {{ option.label }}
              </option>
            </AppSelect>
          </div>

          <div class="md:col-span-2">
            <ItemLabelsSelector :selected-labels="selectedLabels" :available-labels="allLabels" :loading="labelsLoading"
              @add="emit('label:add', $event)" @remove="emit('label:remove', $event)"
              @create="emit('label:create', $event)" />
          </div>

          <label class="grid gap-1 md:col-span-2">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">URL</span>
            <input data-element="item-source-url" class="input-shell" :value="values.source_url" type="url"
              @input="updateField('source_url', ($event.target as HTMLInputElement).value)" />
            <span class="block min-h-4 truncate text-xs font-medium"
              :class="validationErrors.source_url ? 'text-danger-500' : 'invisible'">
              {{ validationErrors.source_url ?? ' ' }}
            </span>
          </label>

          <div class="grid gap-2 md:col-span-2">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Image</span>

            <div class="flex flex-wrap items-center gap-2">
              <input data-element="item-image-upload"
                class="text-copy text-sm file:mr-3 file:rounded-lg file:border-0 file:bg-surface-soft file:px-3 file:py-2 file:text-sm file:font-medium file:text-copy hover:file:bg-surface-muted"
                type="file" accept="image/*" @change="onImageChange" />
              <Button v-if="imagePreviewSrc" data-element="item-image-clear" label="Remove" icon="pi pi-trash"
                severity="secondary" outlined size="small" @click="clearImage" />
            </div>

            <div v-if="imagePreviewSrc"
              class="border-line-subtle bg-surface-muted h-24 w-24 overflow-hidden rounded-lg border">
              <img :src="imagePreviewSrc" alt="Item preview" class="h-full w-full object-cover" />
            </div>
          </div>

          <div class="grid gap-3 md:col-span-2 md:grid-cols-3">
            <label class="grid min-w-0 gap-1">
              <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Value</span>
              <input data-element="item-value" class="w-full min-w-0 input-shell" :value="values.value"
                inputmode="decimal" type="text"
                @input="updateField('value', ($event.target as HTMLInputElement).value)" />
              <span class="block min-h-4 truncate text-xs font-medium"
                :class="validationErrors.value ? 'text-danger-500' : 'invisible'">
                {{ validationErrors.value ?? ' ' }}
              </span>
            </label>

            <label class="grid gap-1">
              <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Weight (<span
                  class="lowercase">{{ weightInputLabel }}</span>)</span>
              <input data-element="item-weight" class="w-full min-w-0 input-shell" :value="values.weight_value"
                inputmode="decimal" type="text" :aria-invalid="Boolean(validationErrors.weight_value)"
                :aria-describedby="validationErrors.weight_value ? 'item-weight-error' : undefined"
                @input="updateField('weight_value', ($event.target as HTMLInputElement).value)" />
              <span id="item-weight-error" class="block min-h-4 truncate text-xs font-medium"
                :class="validationErrors.weight_value ? 'text-danger-500' : 'invisible'">
                {{ validationErrors.weight_value ?? ' ' }}
              </span>
            </label>

            <label class="grid gap-1">
              <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Volume (<span
                  class="lowercase">{{ volumeInputLabel }}</span>)</span>
              <input data-element="item-volume" class="w-full min-w-0 input-shell" :value="values.volume_value"
                inputmode="decimal" type="text" :aria-invalid="Boolean(validationErrors.volume_value)"
                :aria-describedby="validationErrors.volume_value ? 'item-volume-error' : undefined"
                @input="updateField('volume_value', ($event.target as HTMLInputElement).value)" />
              <span id="item-volume-error" class="block min-h-4 truncate text-xs font-medium"
                :class="validationErrors.volume_value ? 'text-danger-500' : 'invisible'">
                {{ validationErrors.volume_value ?? ' ' }}
              </span>
            </label>
          </div>
        </div>

        <div class="border-line-subtle border-t pt-4">
          <h3 class="heading-section">Defaults</h3>
          <div class="mt-3 grid gap-3 md:grid-cols-3">
            <div class="grid gap-1">
              <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Default quantity</span>
              <AppSelect data-element="item-default-quantity" :model-value="values.default_quantity"
                @update:model-value="(value) => updateField('default_quantity', value)">
                <option value="">Select quantity</option>
                <option v-for="option in defaultQuantityOptions" :key="option" :value="String(option)">
                  {{ option }}
                </option>
              </AppSelect>
            </div>

            <div class="grid gap-1">
              <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Default carry status</span>
              <AppToggleGroup name="item-default-carry-status" data-element="item-default-carry-status"
                :model-value="defaultCarryValue" :options="carryStatusOptions" fit-content
                @update:model-value="(value) => updateField('default_carry_status', value as DefaultCarryStatus)" />
            </div>

            <div class="grid gap-1">
              <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Is default</span>
              <AppToggleGroup name="item-is-default" data-element="item-is-default"
                :model-value="values.is_default ? 'active' : 'inactive'" :options="activeOptions" fit-content
                @update:model-value="(value) => updateField('is_default', value === 'active')" />
            </div>
          </div>
        </div>

        <div class="border-line-subtle border-t pt-4">
          <h3 class="heading-section">Additional properties</h3>

          <div class="grid gap-3 md:grid-cols-2">
            <template v-if="values.type === 'consumable'">
              <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Dose count</span>
                <AppSelect data-element="item-dose-count" :model-value="values.dose_count"
                  :invalid="Boolean(validationErrors.dose_count)" :message="validationErrors.dose_count"
                  reserve-message-space @update:model-value="(value) => updateField('dose_count', value)">
                  <option value="">Select dose count</option>
                  <option v-for="option in doseCountOptions" :key="option" :value="option">
                    {{ option }}
                  </option>
                </AppSelect>
              </div>

              <label class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Calories</span>
                <input data-element="item-calories" class="input-shell" :value="values.calories" inputmode="decimal"
                  type="text" @input="updateField('calories', ($event.target as HTMLInputElement).value)" />
                <span class="block min-h-4 truncate text-xs font-medium"
                  :class="validationErrors.calories ? 'text-danger-500' : 'invisible'">
                  {{ validationErrors.calories ?? ' ' }}
                </span>
              </label>

              <label class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Calories per serving</span>
                <input data-element="item-calories-per-serving"
                  class="border-line-subtle bg-surface-muted text-copy rounded-lg border px-3 py-2 text-sm outline-none"
                  :value="caloriesPerServing" readonly type="text" />
              </label>

              <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Requires water</span>
                <AppToggleGroup name="item-requires-water" data-element="item-requires-water"
                  :model-value="values.requires_water ? 'yes' : 'no'" :options="yesNoOptions" fit-content
                  @update:model-value="(value) => updateField('requires_water', value === 'yes')" />
              </div>
            </template>

            <template v-else-if="values.type === 'wearable'">
              <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Season</span>
                <AppSelect data-element="item-season" :model-value="values.season"
                  @update:model-value="(value) => updateField('season', value as ItemFormValues['season'])">
                  <option value="">Select season</option>
                  <option v-for="option in seasonOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </AppSelect>
              </div>

              <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Layer</span>
                <AppSelect data-element="item-layer" :model-value="values.layer"
                  @update:model-value="(value) => updateField('layer', value as ItemFormValues['layer'])">
                  <option value="">Select layer</option>
                  <option v-for="option in layerOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </AppSelect>
              </div>

              <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Size</span>
                <AppSelect data-element="item-size" :model-value="values.size" :invalid="Boolean(validationErrors.size)"
                  :message="validationErrors.size" reserve-message-space
                  @update:model-value="(value) => updateField('size', value)">
                  <option value="">Select size</option>
                  <option v-for="option in sizeOptions" :key="option" :value="option">
                    {{ option }}
                  </option>
                </AppSelect>
              </div>

              <label class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Color</span>
                <input data-element="item-color" class="input-shell" :value="values.color" type="text"
                  @input="updateField('color', ($event.target as HTMLInputElement).value)" />
              </label>

              <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Waterproof</span>
                <AppToggleGroup name="item-waterproof" data-element="item-waterproof"
                  :model-value="values.waterproof ? 'active' : 'inactive'" :options="activeOptions" fit-content
                  @update:model-value="(value) => updateField('waterproof', value === 'active')" />
              </div>
            </template>

            <template v-else-if="values.type === 'shelter'">
              <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Capacity people</span>
                <AppSelect data-element="item-capacity-people" :model-value="values.capacity_people"
                  :invalid="Boolean(validationErrors.capacity_people)" :message="validationErrors.capacity_people"
                  reserve-message-space @update:model-value="(value) => updateField('capacity_people', value)">
                  <option value="">Select capacity</option>
                  <option v-for="option in capacityPeopleOptions" :key="option" :value="option">
                    {{ option }}
                  </option>
                </AppSelect>
              </div>

              <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Season rating</span>
                <AppSelect data-element="item-season-rating" :model-value="values.season_rating"
                  @update:model-value="(value) => updateField('season_rating', value as ItemFormValues['season_rating'])">
                  <option value="">Select season rating</option>
                  <option v-for="option in seasonRatingOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </AppSelect>
              </div>

              <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Freestanding</span>
                <AppToggleGroup name="item-freestanding" data-element="item-freestanding"
                  :model-value="values.freestanding ? 'active' : 'inactive'" :options="activeOptions" fit-content
                  @update:model-value="(value) => updateField('freestanding', value === 'active')" />
              </div>

              <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Has footprint</span>
                <AppToggleGroup name="item-has-footprint" data-element="item-has-footprint"
                  :model-value="values.has_footprint ? 'active' : 'inactive'" :options="activeOptions" fit-content
                  @update:model-value="(value) => updateField('has_footprint', value === 'active')" />
              </div>
            </template>

            <template v-else-if="values.type === 'sleep'">
              <label class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Comfort temp C</span>
                <input data-element="item-comfort-temp-c" class="input-shell" :value="values.comfort_temp_c"
                  inputmode="decimal" type="text"
                  @input="updateField('comfort_temp_c', ($event.target as HTMLInputElement).value)" />
              </label>

              <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Fill type</span>
                <AppSelect data-element="item-fill-type" :model-value="values.fill_type"
                  @update:model-value="(value) => updateField('fill_type', value as ItemFormValues['fill_type'])">
                  <option value="">Select fill type</option>
                  <option v-for="option in fillTypeOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </AppSelect>
              </div>

              <label class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">R value</span>
                <input data-element="item-r-value" class="input-shell" :value="values.r_value" inputmode="decimal"
                  type="text" @input="updateField('r_value', ($event.target as HTMLInputElement).value)" />
                <span class="block min-h-4 truncate text-xs font-medium"
                  :class="validationErrors.r_value ? 'text-danger-500' : 'invisible'">
                  {{ validationErrors.r_value ?? ' ' }}
                </span>
              </label>
            </template>

            <template v-else-if="values.type === 'electronics'">
              <label class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Capacity mAh</span>
                <input data-element="item-capacity-mah" class="input-shell" :value="values.capacity_mah"
                  inputmode="numeric" type="text"
                  @input="updateField('capacity_mah', ($event.target as HTMLInputElement).value)" />
                <span class="block min-h-4 truncate text-xs font-medium"
                  :class="validationErrors.capacity_mah ? 'text-danger-500' : 'invisible'">
                  {{ validationErrors.capacity_mah ?? ' ' }}
                </span>
              </label>

              <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Charge port</span>
                <AppSelect data-element="item-charge-port" :model-value="values.charge_port"
                  @update:model-value="(value) => updateField('charge_port', value as ItemFormValues['charge_port'])">
                  <option value="">Select charge port</option>
                  <option v-for="option in chargePortOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </AppSelect>
              </div>

              <div class="grid gap-1">
                <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Rechargeable</span>
                <AppToggleGroup name="item-rechargeable" data-element="item-rechargeable"
                  :model-value="values.rechargeable ? 'active' : 'inactive'" :options="activeOptions" fit-content
                  @update:model-value="(value) => updateField('rechargeable', value === 'active')" />
              </div>
            </template>

            <template v-else>
              <p v-if="dynamicFieldsLoading" class="text-copy-subtle text-sm">Loading category fields...</p>

              <p v-else-if="dynamicFields.length === 0" class="text-copy-subtle text-sm">No custom fields defined for
                this category.</p>

              <template v-else>
                <template v-for="field in dynamicFields" :key="field.field_key">
                  <div v-if="field.data_type === 'enum'" class="grid gap-1">
                    <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">{{ field.field_label
                    }}</span>
                    <AppSelect :data-element="`item-attribute-${field.field_key}`"
                      :model-value="getAttributeStringValue(field.field_key)"
                      :invalid="Boolean(getDynamicFieldError(field.field_key))"
                      :message="getDynamicFieldError(field.field_key)" reserve-message-space
                      @update:model-value="(value) => updateAttributeValue(field.field_key, value)">
                      <option value="">Select option</option>
                      <option v-for="option in field.enum_options ?? []" :key="option" :value="option">
                        {{ option }}
                      </option>
                    </AppSelect>
                  </div>

                  <label v-else-if="field.data_type === 'string'" class="grid gap-1">
                    <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">{{ field.field_label
                    }}</span>
                    <input :data-element="`item-attribute-${field.field_key}`" class="input-shell"
                      :value="getAttributeStringValue(field.field_key)" type="text"
                      @input="updateAttributeValue(field.field_key, ($event.target as HTMLInputElement).value)" />
                    <span class="block min-h-4 truncate text-xs font-medium"
                      :class="getDynamicFieldError(field.field_key) ? 'text-danger-500' : 'invisible'">

                    </span>
                  </label>

                  <label v-else-if="field.data_type === 'number'" class="grid gap-1">
                    <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">{{ field.field_label
                    }}</span>
                    <input :data-element="`item-attribute-${field.field_key}`" class="input-shell"
                      :value="getAttributeStringValue(field.field_key)" inputmode="decimal" type="text"
                      @input="updateAttributeValue(field.field_key, ($event.target as HTMLInputElement).value)" />
                    <span class="block min-h-4 truncate text-xs font-medium"
                      :class="getDynamicFieldError(field.field_key) ? 'text-danger-500' : 'invisible'">

                    </span>
                  </label>

                  <label v-else-if="field.data_type === 'integer'" class="grid gap-1">
                    <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">{{ field.field_label
                    }}</span>
                    <input :data-element="`item-attribute-${field.field_key}`" class="input-shell"
                      :value="getAttributeStringValue(field.field_key)" inputmode="numeric" type="text"
                      @input="updateAttributeValue(field.field_key, ($event.target as HTMLInputElement).value)" />
                    <span class="block min-h-4 truncate text-xs font-medium"
                      :class="getDynamicFieldError(field.field_key) ? 'text-danger-500' : 'invisible'">

                    </span>
                  </label>

                  <div v-else class="grid gap-1">
                    <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">{{ field.field_label
                    }}</span>
                    <AppToggleGroup :name="`item-attribute-${field.field_key}`"
                      :data-element="`item-attribute-${field.field_key}`"
                      :model-value="getAttributeBooleanValue(field.field_key) ? 'yes' : 'no'" :options="yesNoOptions"
                      fit-content
                      @update:model-value="(value) => updateAttributeValue(field.field_key, value === 'yes')" />
                    <span class="block min-h-4 truncate text-xs font-medium"
                      :class="getDynamicFieldError(field.field_key) ? 'text-danger-500' : 'invisible'">

                    </span>
                  </div>
                </template>
              </template>
            </template>

          </div>
        </div>
      </div>
    </div>

    <footer v-if="!bare && showButtons !== false" data-element="item-form-actions"
      class="mt-4 flex shrink-0 flex-wrap items-center gap-2">
      <Button data-element="item-form-submit" :label="submitLabel" icon="pi pi-check"
        :disabled="!canSubmit || loading || Object.keys(validationErrors).length > 0" :loading="loading"
        @click="onSubmit" />
      <Button v-if="showCancel" data-element="item-form-cancel" label="Cancel" icon="pi pi-times" severity="secondary"
        outlined :disabled="loading" @click="onCancel" />
    </footer>
  </section>
</template>