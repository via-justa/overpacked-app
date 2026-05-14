export type KnownItemType = 'consumable' | 'wearable' | 'shelter' | 'sleep' | 'electronics'
export type DefaultCarryStatus = 'packed' | 'worn'
export type ItemSeason = 'summer' | 'winter' | 'year_round'
export type ItemLayer = 'base' | 'mid' | 'shell' | 'accessory'
export type SeasonRating = '3-season' | '4-season'
export type SleepFillType = 'down' | 'synthetic' | 'foam' | 'air' | 'other'
export type ChargePort = 'usb-c' | 'micro-usb' | 'lightning' | 'dc'

type ItemBase = {
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
  dose_count?: number | null
  calories?: number | null
  calories_per_serving?: number | null
  requires_water?: boolean | null
  season?: ItemSeason | null
  layer?: ItemLayer | null
  waterproof?: boolean | null
  size?: string | null
  color?: string | null
  capacity_people?: number | null
  season_rating?: SeasonRating | null
  freestanding?: boolean | null
  has_footprint?: boolean | null
  comfort_temp_c?: number | null
  fill_type?: SleepFillType | null
  r_value?: number | null
  capacity_mah?: number | null
  charge_port?: ChargePort | null
  rechargeable?: boolean | null
  image_blob?: string | null
  image_mime_type?: string | null
  image_size_bytes?: number | null
  image_width_px?: number | null
  image_height_px?: number | null
  attributes?: Record<string, unknown> | null
  created_at: string
  updated_at: string
}

export type CustomItem = ItemBase & {
  type: string
}

export type ConsumableItem = ItemBase & {
  type: 'consumable'
  dose_count?: number | null
  calories?: number | null
  calories_per_serving?: number | null
  requires_water?: boolean | null
}

export type WearableItem = ItemBase & {
  type: 'wearable'
  season?: 'summer' | 'winter' | 'year_round' | null
  layer?: 'base' | 'mid' | 'shell' | 'accessory' | null
  waterproof?: boolean | null
  size?: string | null
  color?: string | null
}

export type ShelterItem = ItemBase & {
  type: 'shelter'
  capacity_people?: number | null
  season_rating?: '3-season' | '4-season' | null
  freestanding?: boolean | null
  has_footprint?: boolean | null
}

export type SleepItem = ItemBase & {
  type: 'sleep'
  comfort_temp_c?: number | null
  fill_type?: 'down' | 'synthetic' | 'foam' | 'air' | 'other' | null
  r_value?: number | null
}

export type ElectronicsItem = ItemBase & {
  type: 'electronics'
  capacity_mah?: number | null
  charge_port?: 'usb-c' | 'micro-usb' | 'lightning' | 'dc' | null
  rechargeable?: boolean | null
}

export type Item = ConsumableItem | WearableItem | ShelterItem | SleepItem | ElectronicsItem | CustomItem

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
  dose_count?: number
  calories?: number
  calories_per_serving?: number
  requires_water?: boolean
  season?: ItemSeason
  layer?: ItemLayer
  waterproof?: boolean
  size?: string
  color?: string
  capacity_people?: number
  season_rating?: SeasonRating
  freestanding?: boolean
  has_footprint?: boolean
  comfort_temp_c?: number
  fill_type?: SleepFillType
  r_value?: number
  capacity_mah?: number
  charge_port?: ChargePort
  rechargeable?: boolean
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
  dose_count?: number
  calories?: number
  calories_per_serving?: number
  requires_water?: boolean
  season?: ItemSeason
  layer?: ItemLayer
  waterproof?: boolean
  size?: string
  color?: string
  capacity_people?: number
  season_rating?: SeasonRating
  freestanding?: boolean
  has_footprint?: boolean
  comfort_temp_c?: number
  fill_type?: SleepFillType
  r_value?: number
  capacity_mah?: number
  charge_port?: ChargePort
  rechargeable?: boolean
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

export const KNOWN_ITEM_TYPES: KnownItemType[] = ['consumable', 'wearable', 'shelter', 'sleep', 'electronics']

export const isKnownItemType = (value: string): value is KnownItemType => {
  return KNOWN_ITEM_TYPES.includes(value as KnownItemType)
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
  dose_count: string
  calories: string
  calories_per_serving: string
  requires_water: boolean
  season: ItemSeason | ''
  layer: ItemLayer | ''
  waterproof: boolean
  size: string
  color: string
  capacity_people: string
  season_rating: SeasonRating | ''
  freestanding: boolean
  has_footprint: boolean
  comfort_temp_c: string
  fill_type: SleepFillType | ''
  r_value: string
  capacity_mah: string
  charge_port: ChargePort | ''
  rechargeable: boolean
  image_blob: string
  image_mime_type: string
  image_size_bytes: string
  attributes: Record<string, string | boolean>
}