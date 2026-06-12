<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { iconRegistry } from '../../../lib/icons'
import { useMutationWithToast } from '../../../composables/useMutationWithToast'
import {
  downloadBackup,
  getBackupConfig,
  importBackup,
  runBackupNow,
  updateBackupConfig,
  type ImportBackupParams,
} from '../api/backupApi'
import SettingsSectionCard from './SettingsSectionCard.vue'
import SettingsBackupImportDialog from './SettingsBackupImportDialog.vue'
import type { BackupConfig, BackupConfigUpdate, BackupRunResult } from '../types'

const SCHEDULE_PRESETS = [
  { label: 'Nightly (02:00)', value: '0 2 * * *' },
  { label: 'Weekly (Sun 02:00)', value: '0 2 * * 0' },
  { label: 'Monthly (1st, 02:00)', value: '0 2 1 * *' },
] as const

const CUSTOM = 'custom'

const enabled = ref(false)
const schedulePreset = ref<string>(SCHEDULE_PRESETS[0].value)
const customSchedule = ref('')
const retentionCount = ref(7)

const isImportDialogOpen = ref(false)
const importSuccessToken = ref(0)

const configQuery = useQuery({
  queryKey: ['backup-config'],
  queryFn: getBackupConfig,
})

const applyConfig = (config: BackupConfig) => {
  enabled.value = config.enabled
  retentionCount.value = config.retention_count

  const preset = SCHEDULE_PRESETS.find((option) => option.value === config.schedule)
  if (preset) {
    schedulePreset.value = preset.value
    customSchedule.value = ''
  } else {
    schedulePreset.value = CUSTOM
    customSchedule.value = config.schedule
  }
}

watch(
  () => configQuery.data.value,
  (config) => {
    if (config) {
      applyConfig(config)
    }
  },
  { immediate: true },
)

const resolvedSchedule = computed(() =>
  schedulePreset.value === CUSTOM ? customSchedule.value.trim() : schedulePreset.value,
)

const errorMessage = computed(() => {
  const error = configQuery.error.value
  return error instanceof Error ? error.message : 'Unable to load backup settings.'
})

const lastRunLabel = computed(() => {
  const config = configQuery.data.value
  if (!config?.last_run_at) {
    return null
  }
  const when = new Date(config.last_run_at).toLocaleString()
  return config.last_status === 'error' ? `Last run failed at ${when}` : `Last run succeeded at ${when}`
})

const updateMutation = useMutationWithToast<BackupConfig, Error, BackupConfigUpdate>({
  mutationFn: (payload) => updateBackupConfig(payload),
  successMessage: { summary: 'Backup settings saved', detail: 'Scheduled backup configuration was updated.' },
  errorMessage: { summary: 'Save failed', detail: 'Unable to save backup settings.' },
  setQueryData: { queryKey: ['backup-config'], updater: (config) => config },
})

const downloadMutation = useMutationWithToast<void, Error, void>({
  mutationFn: () => downloadBackup(),
  successMessage: { summary: 'Backup ready', detail: 'Your backup download has started.' },
  errorMessage: { summary: 'Backup failed', detail: 'Unable to download backup.' },
})

const runNowMutation = useMutationWithToast<BackupRunResult, Error, void>({
  mutationFn: () => runBackupNow(),
  successMessage: { summary: 'Backup written', detail: 'A backup was written to the configured path.' },
  errorMessage: { summary: 'Backup failed', detail: 'Unable to run backup.' },
  invalidateQueries: [['backup-config']],
})

const importMutation = useMutationWithToast<unknown, Error, ImportBackupParams>({
  mutationFn: (params) => importBackup(params),
  successMessage: { summary: 'Restore complete', detail: 'Your data was restored from the backup.', life: 3500 },
  errorMessage: { summary: 'Restore failed', detail: 'Unable to import backup.' },
  invalidateAllQueries: true,
  onSuccess: () => {
    importSuccessToken.value += 1
  },
})

const isScheduleValid = computed(() => resolvedSchedule.value.length > 0)

const onSave = () => {
  if (!isScheduleValid.value) {
    return
  }
  void updateMutation.mutateAsync({
    enabled: enabled.value,
    schedule: resolvedSchedule.value,
    retention_count: retentionCount.value,
  })
}

const onImport = (params: ImportBackupParams) => {
  void importMutation.mutateAsync(params)
}
</script>

<template>
  <SettingsSectionCard title="Backup & Restore"
    description="Download a full backup of all your data and images, restore from one, or schedule automatic backups to the server.">
    <Message v-if="configQuery.isError.value" data-element="settings-backup-error" severity="error" :closable="false"
      class="mb-4">
      {{ errorMessage }}
    </Message>

    <div class="flex flex-col gap-4">
      <!-- Manual backup / restore -->
      <div
        class="border-line-subtle bg-surface-base flex flex-wrap items-center justify-between gap-3 rounded-xl border px-4 py-3">
        <div>
          <h3 class="text-ink text-sm font-semibold">Manual backup</h3>
          <p class="text-copy-muted text-xs">Download a full backup archive, or restore your data from one.</p>
        </div>
        <div class="flex flex-wrap gap-2">
          <Button data-element="settings-backup-download" label="Download Backup"
            :icon="`pi ${iconRegistry.action.download}`" :loading="downloadMutation.isPending.value"
            :disabled="downloadMutation.isPending.value" @click="downloadMutation.mutateAsync()" />
          <Button data-element="settings-backup-restore" label="Restore" :icon="`pi ${iconRegistry.action.upload}`"
            severity="secondary" outlined @click="isImportDialogOpen = true" />
        </div>
      </div>

      <!-- Scheduled backups -->
      <div class="border-line-subtle bg-surface-base flex flex-col gap-4 rounded-xl border px-4 py-4">
        <label class="text-copy flex items-center gap-2 text-sm font-semibold">
          <input data-element="settings-backup-enabled" v-model="enabled" type="checkbox" />
          Enable scheduled backups
        </label>

        <div class="grid gap-3 sm:grid-cols-2">
          <label class="grid gap-1">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Schedule</span>
            <select data-element="settings-backup-schedule-preset" v-model="schedulePreset" class="input-shell">
              <option v-for="option in SCHEDULE_PRESETS" :key="option.value" :value="option.value">
                {{ option.label }}
              </option>
              <option :value="CUSTOM">Custom (cron expression)</option>
            </select>
          </label>

          <label v-if="schedulePreset === CUSTOM" class="grid gap-1">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Cron expression</span>
            <input data-element="settings-backup-schedule-custom" v-model="customSchedule" class="input-shell"
              type="text" placeholder="0 2 * * *" />
          </label>

          <label class="grid gap-1">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Keep last N backups</span>
            <input data-element="settings-backup-retention" v-model.number="retentionCount" class="input-shell"
              type="number" min="1" />
          </label>
        </div>

        <p v-if="lastRunLabel" data-element="settings-backup-last-run" class="text-copy-muted text-xs">
          {{ lastRunLabel }}
          <span v-if="configQuery.data.value?.last_status === 'error' && configQuery.data.value?.last_error"
            class="text-copy-subtle">— {{ configQuery.data.value.last_error }}</span>
        </p>

        <footer class="flex flex-wrap items-center gap-3">
          <Button data-element="settings-backup-save" label="Save Schedule"
            :icon="`pi ${iconRegistry.action.confirm}`" :loading="updateMutation.isPending.value"
            :disabled="!isScheduleValid || updateMutation.isPending.value" @click="onSave" />
          <Button data-element="settings-backup-run-now" label="Run Now" :icon="`pi ${iconRegistry.action.refresh}`"
            severity="secondary" outlined :loading="runNowMutation.isPending.value"
            :disabled="runNowMutation.isPending.value" @click="runNowMutation.mutateAsync()" />
        </footer>
      </div>
    </div>
  </SettingsSectionCard>

  <SettingsBackupImportDialog v-model="isImportDialogOpen" :is-pending="importMutation.isPending.value"
    :close-dialog-token="importSuccessToken" @import="onImport" />
</template>
