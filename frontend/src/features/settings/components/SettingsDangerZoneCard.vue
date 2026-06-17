<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import Button from 'primevue/button'
import { iconRegistry } from '../../../lib/icons'
import AppTemplateDialog from '../../../components/dialogs/AppTemplateDialog.vue'
import SettingsSectionCard from './SettingsSectionCard.vue'

const props = defineProps<{
  isPending: boolean
  closeDialogToken: number
}>()

const emit = defineEmits<{
  startFresh: [payload: { password: string; reseed: boolean }]
}>()

const isDialogOpen = ref(false)
const password = ref('')
const reseed = ref(true)

const canSubmit = computed(() => password.value.trim().length > 0)

const dangerZoneCardStyle = {
  borderColor: 'color-mix(in srgb, var(--color-danger-500) 55%, var(--color-line-subtle))',
  backgroundColor: 'color-mix(in srgb, var(--color-danger-500) 22%, var(--color-surface-elevated))',
}

const openDialog = () => {
  password.value = ''
  reseed.value = true
  isDialogOpen.value = true
}

const closeDialog = () => {
  password.value = ''
  reseed.value = true
  isDialogOpen.value = false
}

watch(
  () => props.closeDialogToken,
  () => {
    closeDialog()
  },
)

const submit = () => {
  if (!canSubmit.value || props.isPending) {
    return
  }

  emit('startFresh', { password: password.value, reseed: reseed.value })
}
</script>

<template>
  <SettingsSectionCard title="Danger Zone" description="Irreversible actions that permanently remove your app data."
    :style="dangerZoneCardStyle">
    <div
      class="border-line-subtle bg-surface-base flex flex-wrap items-center justify-between gap-3 rounded-xl border px-4 py-3">
      <div>
        <h3 class="text-ink text-sm font-semibold">Start fresh</h3>
        <p class="text-copy-muted text-xs">
          Permanently erase all persons, items, sets, manufacturers, and custom categories.
        </p>
      </div>

      <Button data-element="settings-start-fresh" label="Start Fresh" :icon="`pi ${iconRegistry.action.delete}`"
        severity="danger" outlined :disabled="isPending" @click="openDialog" />
    </div>
  </SettingsSectionCard>

  <AppTemplateDialog v-model="isDialogOpen" data-element="settings-start-fresh-dialog"
    width="min(32rem, calc(100vw - 2rem))" @hide="closeDialog">
    <article class="border-line-subtle bg-surface-elevated rounded-2xl border p-5 shadow-panel">
      <header class="mb-3">
        <h3 class="text-copy text-sm font-semibold uppercase tracking-[0.08em]">Confirm Start Fresh</h3>
      </header>

      <p class="text-copy-subtle text-sm leading-6">
        This will permanently erase all app data and reset defaults. This action cannot be undone.
      </p>

      <label class="mt-4 grid gap-1">
        <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Enter password to continue</span>
        <input data-element="settings-start-fresh-password" v-model="password" class="input-shell" type="password"
          autocomplete="current-password" :disabled="isPending" />
      </label>

      <label class="text-copy mt-4 flex items-start gap-2 text-sm">
        <input data-element="settings-start-fresh-reseed" v-model="reseed" type="checkbox" class="mt-0.5"
          :disabled="isPending" />
        <span>
          Restore default catalog data (labels &amp; manufacturers)
          <span class="text-copy-subtle block text-xs">
            Leave unchecked to start with a completely empty database.
          </span>
        </span>
      </label>

      <footer class="mt-5 flex justify-end gap-2">
        <Button data-element="settings-start-fresh-cancel" label="Cancel" :icon="`pi ${iconRegistry.action.cancel}`"
          severity="secondary" outlined :disabled="isPending" @click="closeDialog" />
        <Button data-element="settings-start-fresh-confirm" label="Delete All Data"
          :icon="`pi ${iconRegistry.action.delete}`" severity="danger" :disabled="!canSubmit || isPending"
          :loading="isPending" @click="submit" />
      </footer>
    </article>
  </AppTemplateDialog>
</template>
