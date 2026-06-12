<script setup lang="ts">
import { ref } from 'vue'
import Button from 'primevue/button'
import { iconRegistry } from '../../../lib/icons'
import { useMutationWithToast } from '../../../composables/useMutationWithToast'
import { exportItems } from '../api/backupApi'
import SettingsSectionCard from './SettingsSectionCard.vue'

const includeImages = ref(false)

const exportMutation = useMutationWithToast<void, Error, boolean>({
  mutationFn: (withImages: boolean) => exportItems(withImages),
  successMessage: {
    summary: 'Export ready',
    detail: 'Your items export download has started.',
  },
  errorMessage: {
    summary: 'Export failed',
    detail: 'Unable to export items.',
  },
})

const onExport = () => {
  void exportMutation.mutateAsync(includeImages.value)
}
</script>

<template>
  <SettingsSectionCard title="Export Items"
    description="Download all your items as a CSV file for use in other applications. Manufacturer, type, and label names are included; identifiers are not.">
    <div
      class="border-line-subtle bg-surface-base flex flex-wrap items-center justify-between gap-3 rounded-xl border px-4 py-3">
      <label class="text-copy flex items-center gap-2 text-sm">
        <input data-element="settings-export-include-images" v-model="includeImages" type="checkbox"
          :disabled="exportMutation.isPending.value" />
        Include item images (downloads a .zip instead of a .csv)
      </label>

      <Button data-element="settings-export-items" label="Export Items"
        :icon="`pi ${iconRegistry.action.download}`" :loading="exportMutation.isPending.value"
        :disabled="exportMutation.isPending.value" @click="onExport" />
    </div>
  </SettingsSectionCard>
</template>
