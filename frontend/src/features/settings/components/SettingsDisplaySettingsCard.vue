<script setup lang="ts">
import Button from 'primevue/button'
import Message from 'primevue/message'
import { iconRegistry } from '../../../lib/icons'
import UnitSettingField from './UnitSettingField.vue'
import SettingsSectionCard from './SettingsSectionCard.vue'
import type { Settings } from '../types'

type UnitSettingKey = keyof Pick<
  Settings,
  'weight_unit' | 'distance_unit' | 'temperature_unit' | 'volume_unit' | 'currency'
>

type FieldConfig = {
  key: UnitSettingKey
  label: string
  helper: string
  options: Array<{ label: string; value: string }>
}

defineProps<{
  hasSettingsError: boolean
  settingsErrorMessage: string
  isLoadingSettings: boolean
  editableSettings: Settings | null
  fieldConfigs: Array<FieldConfig>
  isSavingSettings: boolean
  isDirty: boolean
  isFieldDirty: (key: UnitSettingKey) => boolean
}>()

const emit = defineEmits<{
  save: []
  reset: []
  updateField: [payload: { key: UnitSettingKey; value: string }]
}>()
</script>

<template>
  <SettingsSectionCard title="Display Settings"
    description="Customize how units are displayed across the app. Changes will apply to all weight, distance, temperature, and volume values shown in the UI and exported reports.">
    <Message v-if="hasSettingsError" data-element="settings-error" severity="error" :closable="false" class="mb-4">
      {{ settingsErrorMessage }}
    </Message>

    <div v-if="isLoadingSettings" data-element="settings-loading"
      class="border-line-subtle bg-surface-muted text-copy-muted rounded-2xl border px-4 py-3 text-sm font-medium">
      Loading settings...
    </div>

    <div v-else-if="editableSettings" data-element="settings-fields" class="flex flex-col gap-2">
      <UnitSettingField v-for="field in fieldConfigs" :id="field.key" :key="field.key"
        :data-element="`settings-field-${field.key}`" :label="field.label" :helper="field.helper"
        :model-value="editableSettings[field.key]" :options="field.options" :disabled="isSavingSettings"
        :dirty="isFieldDirty(field.key)"
        @update:model-value="(value) => emit('updateField', { key: field.key, value })" />
    </div>

    <footer data-element="settings-actions" class="mt-6 flex flex-wrap items-center gap-3">
      <Button data-element="settings-save" label="Save Changes" :icon="`pi ${iconRegistry.action.confirm}`"
        :disabled="!isDirty || isSavingSettings || isLoadingSettings" :loading="isSavingSettings"
        @click="emit('save')" />
      <Button data-element="settings-reset" label="Reset" :icon="`pi ${iconRegistry.action.reset}`" severity="secondary"
        outlined :disabled="!isDirty || isSavingSettings || isLoadingSettings" @click="emit('reset')" />
    </footer>
  </SettingsSectionCard>
</template>
