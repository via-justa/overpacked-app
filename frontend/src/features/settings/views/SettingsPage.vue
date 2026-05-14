<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useMutation, useQuery } from '@tanstack/vue-query'
import { useToast } from 'primevue/usetoast'
import { queryClient } from '../../../lib/query/client'
import { getSettings, patchSettings, startFresh } from '../api/settingsApi'
import SettingsDangerZoneCard from '../components/SettingsDangerZoneCard.vue'
import SettingsDisplaySettingsCard from '../components/SettingsDisplaySettingsCard.vue'
import type { Settings, SettingsUpdate } from '../types'

const toast = useToast()
const editableSettings = ref<Settings | null>(null)
const startFreshSuccessToken = ref(0)

const settingsQuery = useQuery({
  queryKey: ['settings'],
  queryFn: getSettings,
})

watch(
  () => settingsQuery.data.value,
  (settings) => {
    if (settings && editableSettings.value === null) {
      editableSettings.value = { ...settings }
    }
  },
  { immediate: true },
)

const updateMutation = useMutation({
  mutationFn: patchSettings,
  onSuccess: (updatedSettings) => {
    queryClient.setQueryData(['settings'], updatedSettings)
    editableSettings.value = { ...updatedSettings }
    toast.add({
      severity: 'success',
      summary: 'Settings saved',
      detail: 'Display units were updated successfully.',
      life: 3000,
    })
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Save failed',
      detail: error instanceof Error ? error.message : 'Unable to save settings.',
      life: 3500,
    })
  },
})

const startFreshMutation = useMutation({
  mutationFn: async (password: string) => startFresh(password),
  onSuccess: async () => {
    startFreshSuccessToken.value += 1
    await queryClient.invalidateQueries()
    toast.add({
      severity: 'success',
      summary: 'Fresh start complete',
      detail: 'All app data was cleared and settings were reset to defaults.',
      life: 3500,
    })
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Start fresh failed',
      detail: error instanceof Error ? error.message : 'Unable to reset app data.',
      life: 3500,
    })
  },
})

const isLoadingSettings = computed(() => settingsQuery.isPending.value)
const hasSettingsError = computed(() => settingsQuery.isError.value)
const settingsErrorMessage = computed(() => {
  const error = settingsQuery.error.value
  return error instanceof Error ? error.message : 'Unable to load settings.'
})
const isSavingSettings = computed(() => updateMutation.isPending.value)

type UnitSettingKey = keyof Pick<
  Settings,
  'weight_unit' | 'distance_unit' | 'temperature_unit' | 'volume_unit' | 'currency'
>

const isDirty = computed(() => {
  const current = settingsQuery.data.value
  const edited = editableSettings.value

  if (!current || !edited) {
    return false
  }

  return (
    current.weight_unit !== edited.weight_unit ||
    current.distance_unit !== edited.distance_unit ||
    current.temperature_unit !== edited.temperature_unit ||
    current.volume_unit !== edited.volume_unit ||
    current.currency !== edited.currency
  )
})

const fieldConfigs: Array<{
  key: UnitSettingKey
  label: string
  helper: string
  options: Array<{ label: string; value: string }>
}> = [
    {
      key: 'weight_unit',
      label: 'Weight unit',
      helper: 'Choose how weight is displayed in the UI and exports.',
      options: [
        { label: 'g', value: 'g' },
        { label: 'oz', value: 'oz' },
      ],
    },
    {
      key: 'distance_unit',
      label: 'Distance unit',
      helper: 'Select preferred distance display for trips and reports.',
      options: [
        { label: 'km', value: 'km' },
        { label: 'mi', value: 'mi' },
      ],
    },
    {
      key: 'temperature_unit',
      label: 'Temperature unit',
      helper: 'Set forecast and weather temperature display units.',
      options: [
        { label: 'C', value: 'c' },
        { label: 'F', value: 'f' },
      ],
    },
    {
      key: 'volume_unit',
      label: 'Volume unit',
      helper: 'Choose liquid volume display while backend keeps canonical ml.',
      options: [
        { label: 'ml', value: 'ml' },
        { label: 'fl oz', value: 'fl_oz' },
      ],
    },
    {
      key: 'currency',
      label: 'Currency',
      helper: 'Select your preferred currency for item values and reports.',
      options: [
        { label: '€', value: 'eur' },
        { label: '$', value: 'usd' },
      ],
    },
  ]

const onReset = () => {
  if (!settingsQuery.data.value) {
    return
  }

  editableSettings.value = { ...settingsQuery.data.value }
}

const onSave = async () => {
  if (!editableSettings.value) {
    return
  }

  const payload: SettingsUpdate = {
    weight_unit: editableSettings.value.weight_unit,
    distance_unit: editableSettings.value.distance_unit,
    temperature_unit: editableSettings.value.temperature_unit,
    volume_unit: editableSettings.value.volume_unit,
    currency: editableSettings.value.currency,
  }

  await updateMutation.mutateAsync(payload)
}

const onConfirmStartFresh = async (password: string) => {
  await startFreshMutation.mutateAsync(password)
}

const setFieldValue = (
  key: UnitSettingKey,
  value: string,
) => {
  if (!editableSettings.value) {
    return
  }

  editableSettings.value = {
    ...editableSettings.value,
    [key]: value,
  }
}

const isFieldDirty = (
  key: UnitSettingKey,
) => {
  const current = settingsQuery.data.value
  const edited = editableSettings.value

  if (!current || !edited) {
    return false
  }

  return current[key] !== edited[key]
}
</script>

<template>
  <section data-component="settings-page" class="flex w-full flex-col gap-5">
    <SettingsDisplaySettingsCard :has-settings-error="hasSettingsError" :settings-error-message="settingsErrorMessage"
      :is-loading-settings="isLoadingSettings" :editable-settings="editableSettings" :field-configs="fieldConfigs"
      :is-saving-settings="isSavingSettings" :is-dirty="isDirty" :is-field-dirty="isFieldDirty" @save="onSave"
      @reset="onReset" @update-field="({ key, value }) => setFieldValue(key, value)" />

    <SettingsDangerZoneCard :is-pending="startFreshMutation.isPending.value"
      :close-dialog-token="startFreshSuccessToken" @start-fresh="onConfirmStartFresh" />
  </section>
</template>
