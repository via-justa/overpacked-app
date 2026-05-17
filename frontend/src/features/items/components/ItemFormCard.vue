<script setup lang="ts">
import { computed } from 'vue'
import Button from 'primevue/button'
import AppSelect from '../../../components/forms/AppSelect.vue'
import AppToggleGroup from '../../../components/forms/AppToggleGroup.vue'
import ItemLabelsSelector from './ItemLabelsSelector.vue'
import type {
  DefaultCarryStatus,
  ItemFormValues,
  ItemTypeField,
  Label,
  Manufacturer,
} from '../types'

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

defineExpose({
  onSubmit,
  onCancel,
})
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
            <template v-if="dynamicFieldsLoading">
              <p class="text-copy-subtle text-sm md:col-span-2">Loading category fields...</p>
            </template>

            <template v-else-if="dynamicFields.length === 0">
              <p class="text-copy-subtle text-sm md:col-span-2">No custom fields defined for
                this category.</p>
            </template>

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
                    {{ getDynamicFieldError(field.field_key) ?? ' ' }}
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
                    {{ getDynamicFieldError(field.field_key) ?? ' ' }}
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
                    {{ getDynamicFieldError(field.field_key) ?? ' ' }}
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
                    {{ getDynamicFieldError(field.field_key) ?? ' ' }}
                  </span>
                </div>
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