import type { components } from '../../lib/api/schema'

// Server types are sourced from the generated OpenAPI schema (single source of truth).
export type Settings = components['schemas']['Settings']
export type SettingsUpdate = components['schemas']['SettingsUpdate']

// Named enums derived from the schema for use in forms/selects/unit conversions.
export type WeightUnit = Settings['weight_unit']
export type DistanceUnit = Settings['distance_unit']
export type TemperatureUnit = Settings['temperature_unit']
export type VolumeUnit = Settings['volume_unit']
export type Currency = Settings['currency']
