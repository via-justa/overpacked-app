import { computed } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { getSettings } from '../features/settings/api/settingsApi'
import type { Currency, VolumeUnit, WeightUnit } from '../features/settings/types'

/**
 * Composable for accessing application settings
 * Provides reactive access to user preferences for units and currency
 */
export function useSettings() {
  const settingsQuery = useQuery({
    queryKey: ['settings'],
    queryFn: getSettings,
  })

  const weightUnit = computed<WeightUnit>(() => settingsQuery.data.value?.weight_unit ?? 'g')
  const volumeUnit = computed<VolumeUnit>(() => settingsQuery.data.value?.volume_unit ?? 'ml')
  const currency = computed<Currency>(() => settingsQuery.data.value?.currency ?? 'usd')

  return {
    settingsQuery,
    weightUnit,
    volumeUnit,
    currency,
  }
}
