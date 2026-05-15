import type { ComputedRef } from 'vue'
import type { Item } from '../../items/types'
import { formatDisplayWeight, mlToFluidOunces, toRoundedString } from '../../../lib/units/conversions'
import type { VolumeUnit } from '../../../lib/units/conversions'
import type { Currency } from '../../settings/types'

export interface SetItemFormatters {
  getManufacturerName: (item: Item) => string
  getItemWeight: (item: Item) => string
  getItemVolume: (item: Item) => string
  getItemValue: (item: Item) => string
}

export function useSetItemFormatters(
  manufacturersById: ComputedRef<Map<string, string>>,
  weightUnit: ComputedRef<'g' | 'oz'>,
  volumeUnit: ComputedRef<VolumeUnit>,
  currency: ComputedRef<Currency>
): SetItemFormatters {
  const getManufacturerName = (item: Item): string => {
    return manufacturersById.value.get(item.manufacturer_id) ?? 'Unknown'
  }

  const getItemWeight = (item: Item): string => {
    if (typeof item.weight_grams !== 'number') {
      return 'Not set'
    }
    return formatDisplayWeight(item.weight_grams, weightUnit.value)
  }

  const getItemVolume = (item: Item): string => {
    if (typeof item.volume_ml !== 'number') {
      return 'Not set'
    }

    if (volumeUnit.value === 'fl_oz') {
      const flOz = mlToFluidOunces(item.volume_ml)
      return `${toRoundedString(flOz)} fl oz`
    }

    return `${toRoundedString(item.volume_ml)} ml`
  }

  const getItemValue = (item: Item): string => {
    if (typeof item.value !== 'number') {
      return 'Not set'
    }

    const currencySymbol = currency.value === 'usd' ? '$' : '€'
    return `${toRoundedString(item.value)} ${currencySymbol}`
  }

  return {
    getManufacturerName,
    getItemWeight,
    getItemVolume,
    getItemValue,
  }
}
