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
export type LabelUpdate = components['schemas']['LabelUpdate']
export type ItemLabelAdd = components['schemas']['ItemLabelAdd']

// Named enums derived from the schema for use in forms/selects.
export type DefaultCarryStatus = NonNullable<Item['default_carry_status']>
export type ItemTypeCreateBaseProfile = NonNullable<ItemTypeCreate['base_profile']>
export type ItemTypeFieldDataType = ItemTypeField['data_type']

// UI-only: a curated subset of item types; the server `type` field is a free-form string.
export type KnownItemType = 'consumable' | 'wearable' | 'shelter' | 'sleep' | 'electronics'

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
  image_blob: string
  image_mime_type: string
  image_size_bytes: string
  attributes: Record<string, string | boolean>
  label_ids: string[]
}
