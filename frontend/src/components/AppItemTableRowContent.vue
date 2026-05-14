<script setup lang="ts">
import AppBooleanValue from './AppBooleanValue.vue'
import AppNotSetValue from './AppNotSetValue.vue'
import { normalizeTitleWords } from '../lib/text/normalize'
import type { Item } from '../features/items/types'

export type AppItemTableField = {
  key: string
  label: string
  render: (item: Item) => string
  renderHref?: (item: Item) => string | undefined
  renderBoolean?: (item: Item) => boolean | null | undefined
}

const props = withDefaults(defineProps<{
  item: Item
  visibleFields: AppItemTableField[]
  showNameLink?: boolean
}>(), {
  showNameLink: true,
})

const emit = defineEmits<{
  openDetails: [item: Item]
}>()
</script>

<template>
  <td class="w-80 px-4 py-3 align-top">
    <button v-if="showNameLink" type="button"
      class="text-brand-500 decoration-brand-200 block max-w-full truncate text-left font-semibold underline underline-offset-2"
      :title="normalizeTitleWords(item.name)" @click="emit('openDetails', item)">
      {{ normalizeTitleWords(item.name) }}
    </button>
    <span v-else class="text-copy block max-w-full truncate font-semibold" :title="normalizeTitleWords(item.name)">
      {{ normalizeTitleWords(item.name) }}
    </span>
  </td>
  <td v-for="field in visibleFields" :key="`${item.id}-${field.key}`" class="whitespace-nowrap px-4 py-3 align-top">
    <template v-if="field.key === 'description'">
      <span v-if="field.render(item) !== 'Not set'" class="group/note relative inline-flex"
        :aria-label="field.render(item)">
        <i class="pi pi-file-edit text-copy-subtle hover:text-copy cursor-default text-sm" aria-hidden="true" />
        <span
          class="pointer-events-none absolute bottom-full left-1/2 z-20 mb-1.5 w-max max-w-xs -translate-x-1/2 rounded-lg border border-line-subtle bg-surface-elevated px-3 py-2 text-xs text-copy opacity-0 shadow-panel transition-opacity group-hover/note:opacity-100">
          {{ field.render(item) }}
        </span>
      </span>
      <AppNotSetValue v-else :label="field.label" />
    </template>
    <template v-else>
      <a v-if="field.renderHref?.(item)" :href="field.renderHref(item)" target="_blank" rel="noreferrer"
        class="text-brand-500 inline-flex items-center gap-1" :aria-label="`Open ${field.label} for ${item.name}`">
        <i class="pi pi-external-link" aria-hidden="true"></i>
        <span class="sr-only">Open {{ field.label }}</span>
      </a>
      <AppBooleanValue v-else-if="typeof field.renderBoolean?.(item) === 'boolean'" :value="field.renderBoolean?.(item)"
        :label="field.label" />
      <span v-else-if="field.render(item) === 'Not set'">
        <AppNotSetValue :label="field.label" />
      </span>
      <span v-else>{{ field.render(item) }}</span>
    </template>
  </td>
</template>
