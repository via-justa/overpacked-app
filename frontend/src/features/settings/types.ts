export type WeightUnit = 'g' | 'oz'
export type DistanceUnit = 'km' | 'mi'
export type TemperatureUnit = 'c' | 'f'
export type VolumeUnit = 'ml' | 'fl_oz'
export type Currency = 'usd' | 'eur'

export type Settings = {
  id: number
  weight_unit: WeightUnit
  distance_unit: DistanceUnit
  temperature_unit: TemperatureUnit
  volume_unit: VolumeUnit
  currency: Currency
}

export type SettingsUpdate = Partial<Pick<Settings, 'weight_unit' | 'distance_unit' | 'temperature_unit' | 'volume_unit' | 'currency'>>
