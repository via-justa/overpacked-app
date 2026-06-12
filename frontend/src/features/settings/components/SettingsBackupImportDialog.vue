<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { iconRegistry } from '../../../lib/icons'
import AppTemplateDialog from '../../../components/dialogs/AppTemplateDialog.vue'
import type { BackupImportMode } from '../types'

const props = defineProps<{
  modelValue: boolean
  isPending: boolean
  closeDialogToken: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  import: [payload: { file: File; mode: BackupImportMode; password: string }]
}>()

const file = ref<File | null>(null)
const mode = ref<BackupImportMode>('merge')
const password = ref('')

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const canSubmit = computed(() => file.value !== null && password.value.trim().length > 0)

const reset = () => {
  file.value = null
  mode.value = 'merge'
  password.value = ''
}

watch(
  () => props.closeDialogToken,
  () => {
    isOpen.value = false
    reset()
  },
)

const onFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  file.value = target.files?.[0] ?? null
}

const submit = () => {
  if (!canSubmit.value || props.isPending || !file.value) {
    return
  }
  emit('import', { file: file.value, mode: mode.value, password: password.value })
}

const close = () => {
  isOpen.value = false
  reset()
}
</script>

<template>
  <AppTemplateDialog v-model="isOpen" data-element="settings-backup-import-dialog"
    width="min(34rem, calc(100vw - 2rem))" @hide="reset">
    <article class="border-line-subtle bg-surface-elevated rounded-2xl border p-5 shadow-panel">
      <header class="mb-3">
        <h3 class="text-copy text-sm font-semibold uppercase tracking-[0.08em]">Restore from backup</h3>
      </header>

      <p class="text-copy-subtle text-sm leading-6">
        Upload a backup archive (.zip) previously exported from this app.
      </p>

      <label class="mt-4 grid gap-1">
        <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Backup archive</span>
        <input data-element="settings-backup-import-file" class="input-shell" type="file" accept=".zip,application/zip"
          :disabled="isPending" @change="onFileChange" />
      </label>

      <fieldset class="mt-4 grid gap-2" :disabled="isPending">
        <legend class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Restore mode</legend>
        <label class="border-line-subtle bg-surface-base flex items-start gap-2 rounded-xl border px-3 py-2">
          <input v-model="mode" type="radio" value="merge" name="restore-mode" class="mt-1" />
          <span>
            <span class="text-ink block text-sm font-semibold">Merge</span>
            <span class="text-copy-muted block text-xs">Keep existing data and add or update entries from the
              backup.</span>
          </span>
        </label>
        <label class="border-line-subtle bg-surface-base flex items-start gap-2 rounded-xl border px-3 py-2">
          <input v-model="mode" type="radio" value="replace" name="restore-mode" class="mt-1" />
          <span>
            <span class="text-ink block text-sm font-semibold">Replace</span>
            <span class="text-copy-muted block text-xs">Erase current data first, then import the backup.</span>
          </span>
        </label>
      </fieldset>

      <Message v-if="mode === 'replace'" data-element="settings-backup-import-replace-warning" severity="warn"
        :closable="false" class="mt-3">
        Replace will permanently delete all current items, sets, trips, and related data before importing.
      </Message>

      <label class="mt-4 grid gap-1">
        <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Enter password to continue</span>
        <input data-element="settings-backup-import-password" v-model="password" class="input-shell" type="password"
          autocomplete="current-password" :disabled="isPending" />
      </label>

      <footer class="mt-5 flex justify-end gap-2">
        <Button data-element="settings-backup-import-cancel" label="Cancel"
          :icon="`pi ${iconRegistry.action.cancel}`" severity="secondary" outlined :disabled="isPending"
          @click="close" />
        <Button data-element="settings-backup-import-confirm" label="Restore"
          :icon="`pi ${iconRegistry.action.upload}`" :severity="mode === 'replace' ? 'danger' : 'primary'"
          :disabled="!canSubmit || isPending" :loading="isPending" @click="submit" />
      </footer>
    </article>
  </AppTemplateDialog>
</template>
