import type { components } from '../../lib/api/schema'

// Server types are sourced from the generated OpenAPI schema (single source of truth).
export type Item = components['schemas']['Item']
export type ItemCreate = components['schemas']['ItemCreate']
export type ItemUpdate = components['schemas']['ItemUpdate']
export type Manufacturer = components['schemas']['Manufacturer']
export type ManufacturerCreate = components['schemas']['ManufacturerCreate']
export type ManufacturerUpdate = components['schemas']['ManufacturerUpdate']
export type ItemTypeCreate = components['schemas']['ItemTypeCreate']
export type ItemTypeUpdate = components['schemas']['ItemTypeUpdate']
export type ItemTypeEntity = components['schemas']['ItemType']
export type ItemTypeFieldInput = components['schemas']['ItemTypeFieldInput']
export type ItemTypeField = components['schemas']['ItemTypeField']
export type Label = components['schemas']['Label']
export type LabelCreate = components['schemas']['LabelCreate']
export type ItemLabelAdd = components['schemas']['ItemLabelAdd']

// Named enums derived from the schema for use in forms/selects.
export type DefaultCarryStatus = NonNullable<Item['default_carry_status']>

// UI-only form state — no server/spec equivalent.
export type ItemFormValues = {
  name: string
  type: string
  is_active: boolean
  manufacturer_id: string
  description: string
  source_url: string
  value: string
  default_quantity: string
  default_carry_status: DefaultCarryStatus
  is_default: boolean
  weight_value: string
  volume_value: string
  // Image bytes are uploaded separately (multipart) after the item is saved.
  // `imageFile` holds a newly picked file; `image_url` is the existing image
  // (if any); `imageRemoved` flags that the user cleared an existing image.
  imageFile: File | null
  image_url: string | null
  imageRemoved: boolean
  attributes: Record<string, string | boolean>
  label_ids: string[]
}
