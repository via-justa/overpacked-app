export type KnownItemType = 'consumable' | 'wearable' | 'shelter' | 'sleep' | 'electronics'
export type DefaultCarryStatus = 'packed' | 'worn'

export type Item = {
  id: string
  name: string
  type: string
  is_active: boolean
  manufacturer_id: string
  description?: string | null
  source_url?: string | null
  value?: number | null
  weight_grams?: number | null
  volume_ml?: number | null
  default_quantity?: number | null
  default_carry_status?: DefaultCarryStatus | null
  is_default?: boolean | null
  image_blob?: string | null
  image_mime_type?: string | null
  image_size_bytes?: number | null
  image_width_px?: number | null
  image_height_px?: number | null
  attributes?: Record<string, unknown> | null
  created_at: string
  updated_at: string
}

export type ItemCreate = {
  name: string
  type: string
  is_active: boolean
  manufacturer_id: string
  description?: string
  source_url?: string
  value?: number
  weight_grams?: number
  volume_ml?: number
  default_quantity?: number
  default_carry_status?: DefaultCarryStatus
  is_default?: boolean
  image_blob?: string
  image_mime_type?: string
  image_size_bytes?: number
  attributes?: Record<string, unknown>
}

export type ItemUpdate = {
  name?: string
  type?: string
  is_active?: boolean
  manufacturer_id?: string
  description?: string
  source_url?: string
  value?: number
  weight_grams?: number
  volume_ml?: number
  default_quantity?: number
  default_carry_status?: DefaultCarryStatus
  is_default?: boolean
  image_blob?: string
  image_mime_type?: string
  image_size_bytes?: number
  attributes?: Record<string, unknown>
}

export type Manufacturer = {
  id: string
  name: string
  website?: string | null
  created_at: string
  updated_at: string
}

export type ManufacturerCreate = {
  name: string
  website?: string
}

export type ManufacturerUpdate = {
  name?: string
  website?: string
}

export type ItemTypeCreateBaseProfile = 'consumable' | 'wearable' | 'shelter' | 'sleep' | 'electronics'

export type ItemTypeCreate = {
  id: string
  name: string
  description?: string
  base_profile?: ItemTypeCreateBaseProfile
}

export type ItemTypeUpdate = {
  name?: string
  description?: string
  base_profile?: ItemTypeCreateBaseProfile
}

export type ItemTypeEntity = {
  id: string
  name: string
  description?: string | null
  base_profile?: ItemTypeCreateBaseProfile | null
  is_system: boolean
}

export type ItemTypeFieldDataType = 'string' | 'integer' | 'number' | 'boolean' | 'enum'

export type ItemTypeFieldInput = {
  field_key: string
  field_label: string
  data_type: ItemTypeFieldDataType
  is_required?: boolean
  enum_options?: string[]
  min_value?: number
  max_value?: number
  unit?: string
  sort_order?: number
}

export type ItemTypeField = {
  id: string
  item_type_id: string
  field_key: string
  field_label: string
  data_type: ItemTypeFieldDataType
  is_required: boolean
  enum_options?: string[] | null
  min_value?: number | null
  max_value?: number | null
  unit?: string | null
  sort_order: number
}

export type Label = {
  id: string
  name: string
  color?: string | null
  created_at: string
  updated_at: string
}

export type LabelCreate = {
  name: string
  color?: string | null
}

export type LabelUpdate = {
  name?: string
  color?: string | null
}

export type ItemLabelAdd = {
  label_id: string
}

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