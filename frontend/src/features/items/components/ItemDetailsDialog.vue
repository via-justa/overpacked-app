<script setup lang="ts">
import Button from 'primevue/button'
import AppBooleanValue from '../../../components/display/AppBooleanValue.vue'
import AppFormViewDialog from '../../../components/dialogs/AppFormViewDialog.vue'
import AppNotSetValue from '../../../components/display/AppNotSetValue.vue'
import { AppIcon } from '../../../components/icons'
import { iconRegistry } from '../../../lib/icons'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import type { Item } from '../types'

interface DetailEntry {
  label: string
  value: string
  href?: string
  booleanValue?: boolean | null
}

defineProps<{
  open: boolean
  selectedItem: Item | null
  getImageSrc: (item: Item) => string
  getDetailedEntries: (item: Item) => DetailEntry[]
  formatType: (value: string) => string
  manufacturersById: Map<string, string>
  isDeleteLoading: boolean
  showEditAction?: boolean
  showDeleteAction?: boolean
}>()

defineEmits<{
  'update:open': [value: boolean]
  edit: [item: Item]
  delete: [itemId: string]
}>()
</script>

<template>
  <AppFormViewDialog :open="open" title="Item Details" data-element="item-details-dialog"
    width="min(44rem, calc(100vw - 2rem))" @update:open="$emit('update:open', $event)">
    <article v-if="selectedItem" data-element="item-details-card" class="px-1">
      <div v-if="getImageSrc(selectedItem)"
        class="border-line-subtle bg-surface-muted mb-4 overflow-hidden rounded-xl border">
        <img :src="getImageSrc(selectedItem)" :alt="normalizeTitleWords(selectedItem.name)"
          class="h-56 w-full object-cover" />
      </div>

      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 class="text-ink text-2xl font-semibold">{{ normalizeTitleWords(selectedItem.name) }}</h2>
          <p class="text-copy-muted mt-1 text-sm">
            {{ formatType(selectedItem.type) }}
            <span class="text-line mx-2">/</span>
            {{ manufacturersById.get(selectedItem.manufacturer_id)
              ? normalizeTitleWords(manufacturersById.get(selectedItem.manufacturer_id) ?? '')
              : selectedItem.manufacturer_id }}
          </p>
        </div>

        <Button v-if="showEditAction !== false" data-element="item-details-edit" label="Edit"
          :icon="`pi ${iconRegistry.action.edit}`" outlined @click="$emit('edit', selectedItem)" />
      </div>

      <div class="mt-5 grid gap-3 sm:grid-cols-2">
        <div v-for="entry in getDetailedEntries(selectedItem)" :key="`${selectedItem.id}-${entry.label}`"
          class="border-line-subtle bg-surface-muted rounded-xl border px-3 py-2">
          <p class="text-copy-subtle text-xs font-semibold uppercase tracking-[0.06em]">{{ entry.label }}</p>
          <a v-if="entry.href" :href="entry.href" target="_blank" rel="noreferrer"
            class="text-brand-500 mt-1 inline-flex items-center gap-1 text-sm" :aria-label="`Open ${entry.label}`">
            <AppIcon category="content" name="externalLink" size="sm" />
            <span class="sr-only">Open {{ entry.label }}</span>
          </a>
          <p v-else-if="typeof entry.booleanValue === 'boolean'" class="mt-1 text-sm">
            <AppBooleanValue :value="entry.booleanValue" :label="entry.label" />
          </p>
          <p v-else-if="entry.value === 'Not set'" class="mt-1 text-sm">
            <AppNotSetValue :label="entry.label" />
          </p>
          <p v-else class="text-copy mt-1 wrap-break-word text-sm">{{ entry.value }}</p>
        </div>
      </div>

      <div v-if="showDeleteAction !== false" class="mt-5 flex flex-wrap items-center gap-2">
        <Button data-element="item-details-delete" label="Delete" :icon="`pi ${iconRegistry.action.delete}`"
          severity="danger" outlined :loading="isDeleteLoading" @click="$emit('delete', selectedItem.id)" />
      </div>
    </article>
  </AppFormViewDialog>
</template>
