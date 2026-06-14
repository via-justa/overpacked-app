import type { components } from '../../lib/api/schema'

// Server types are sourced from the generated OpenAPI schema (single source of truth).
export type Settings = components['schemas']['Settings']
export type SettingsUpdate = components['schemas']['SettingsUpdate']

// Named enums derived from the schema for use in forms/selects/unit conversions.
export type WeightUnit = Settings['weight_unit']
export type VolumeUnit = Settings['volume_unit']
export type Currency = Settings['currency']

// Backup feature types (single source of truth: generated OpenAPI schema).
export type BackupConfig = components['schemas']['BackupConfig']
export type BackupConfigUpdate = components['schemas']['BackupConfigUpdate']
export type BackupImportResult = components['schemas']['BackupImportResult']
export type BackupRunResult = components['schemas']['BackupRunResult']
export type BackupImportMode = BackupImportResult['mode']
