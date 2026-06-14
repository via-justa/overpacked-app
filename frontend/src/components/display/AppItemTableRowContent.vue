<script setup lang="ts">
import { AppIcon } from '../icons'
import AppBooleanValue from './AppBooleanValue.vue'
import AppNotSetValue from './AppNotSetValue.vue'
import { normalizeTitleWords } from '../../lib/text/normalize'
import type { Item, Label } from '../../features/items/types'

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
  itemLabels?: Label[]
}>(), {
  showNameLink: true,
  itemLabels: () => [],
})

const emit = defineEmits<{
  edit: [item: Item]
}>()

// Determine if text should be light or dark based on background color luminance
const getContrastColor = (color?: string | null): 'light' | 'dark' => {
  if (!color) {
    return 'light'
  }

  // Handle HSL colors
  if (color.startsWith('hsl')) {
    const match = color.match(/hsl\((\d+),\s*(\d+)%,\s*(\d+)%\)/)
    if (match) {
      const lightness = Number.parseInt(match[3], 10)
      return lightness > 55 ? 'dark' : 'light'
    }
  }

  // Handle hex colors
  const hex = color.replace('#', '')
  const r = Number.parseInt(hex.substring(0, 2), 16)
  const g = Number.parseInt(hex.substring(2, 4), 16)
  const b = Number.parseInt(hex.substring(4, 6), 16)

  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255

  return luminance > 0.5 ? 'dark' : 'light'
}

const getLabelTextColor = (color?: string | null): string => {
  const contrast = getContrastColor(color)
  return contrast === 'light' ? 'text-ink-inverse' : 'text-ink'
}

const getLabelBorderColor = (color?: string | null): string => {
  const contrast = getContrastColor(color)
  return contrast === 'light'
    ? 'rgba(255, 255, 255, 0.2)'
    : 'rgba(0, 0, 0, 0.1)'
}
</script>

<template>
  <td class="w-80 px-4 py-2 align-middle">
    <button v-if="showNameLink" type="button"
      class="text-brand-500 decoration-brand-200 block max-w-full truncate text-left font-semibold underline underline-offset-2"
      :title="normalizeTitleWords(item.name)" @click="emit('edit', item)">
      {{ normalizeTitleWords(item.name) }}
    </button>
    <span v-else class="text-copy block max-w-full truncate font-semibold" :title="normalizeTitleWords(item.name)">
      {{ normalizeTitleWords(item.name) }}
    </span>
  </td>
  <td v-for="field in visibleFields" :key="`${item.id}-${field.key}`" class="whitespace-nowrap px-4 py-2 align-middle"
    :class="field.key === 'manufacturer' ? '' : 'text-center'">
    <template v-if="field.key === 'labels'">
      <span class="group/labels relative inline-flex items-center gap-1.5"
        :aria-label="`${itemLabels.length} label${itemLabels.length === 1 ? '' : 's'}`">
        <AppIcon category="content" name="tag" size="sm" class="text-copy-subtle hover:text-copy cursor-default" />
        <span class="text-copy-subtle hover:text-copy cursor-default text-xs font-medium">{{ itemLabels.length }}</span>
        <div v-if="itemLabels.length > 0"
          class="pointer-events-none absolute bottom-full left-1/2 z-20 mb-1.5 w-max max-w-xs -translate-x-1/2 rounded-lg border border-line-subtle bg-surface-elevated px-3 py-2 opacity-0 shadow-panel transition-opacity group-hover/labels:opacity-100">
          <div class="flex flex-wrap gap-1.5">
            <span v-for="label in itemLabels" :key="label.id"
              class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium"
              :class="getLabelTextColor(label.color)" :style="{
                backgroundColor: label.color ?? 'var(--color-label-fallback)',
                border: `1px solid ${getLabelBorderColor(label.color)}`
              }">
              {{ label.name }}
            </span>
          </div>
        </div>
      </span>
    </template>
    <template v-else-if="field.key === 'description'">
      <span v-if="field.render(item) !== 'Not set'" class="group/note relative inline-flex"
        :aria-label="field.render(item)">
        <AppIcon category="action" name="editField" size="sm" class="text-copy-subtle hover:text-copy cursor-default" />
        <span
          class="pointer-events-none absolute bottom-full left-1/2 z-20 mb-1.5 w-max max-w-xs -translate-x-1/2 rounded-lg border border-line-subtle bg-surface-elevated px-3 py-2 text-xs text-copy opacity-0 shadow-panel transition-opacity group-hover/note:opacity-100">
          {{ field.render(item) }}
        </span>
      </span>
      <AppNotSetValue v-else :label="field.label" />
    </template>
    <template v-else>
      <a v-if="field.key === 'manufacturer' && field.renderHref?.(item)" :href="field.renderHref(item)" target="_blank"
        rel="noreferrer" class="text-brand-500 decoration-brand-200 underline underline-offset-2"
        :title="field.render(item)" :aria-label="`Open ${field.render(item)} website`">
        {{ field.render(item) }}
      </a>
      <a v-else-if="field.renderHref?.(item)" :href="field.renderHref(item)" target="_blank" rel="noreferrer"
        class="text-brand-500 inline-flex items-center gap-1" :aria-label="`Open ${field.label} for ${item.name}`">
        <AppIcon category="content" name="externalLink" size="sm" />
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
