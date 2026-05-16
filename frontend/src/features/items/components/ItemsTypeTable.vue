<script setup lang="ts">
import { computed } from 'vue'
import AppToggleGroup from '../../../components/AppToggleGroup.vue'
import AppBooleanValue from '../../../components/AppBooleanValue.vue'
import AppItemTableRowContent from '../../../components/AppItemTableRowContent.vue'
import { useRowActionsMenu } from '../../../composables/useRowActionsMenu'
import type { Item, Label } from '../types'

type TableField = {
  key: string
  label: string
  render: (item: Item) => string
  renderHref?: (item: Item) => string | undefined
  renderBoolean?: (item: Item) => boolean | null | undefined
}

const props = defineProps<{
  title: string
  items: Item[]
  baseFields: TableField[]
  extraFields: TableField[]
  tableDetailMode: 'simple' | 'expanded'
  selectionMode: boolean
  selectedItemIds: string[]
  totalWeightLabel: string
  totalValueLabel: string
  itemLabelsMap?: Map<string, Label[]>
}>()

const emit = defineEmits<{
  openDetails: [item: Item]
  'update:tableDetailMode': [mode: 'simple' | 'expanded']
  'update:selectionMode': [value: boolean]
  'toggle:itemSelection': [itemId: string, checked: boolean]
  'toggle:selectAll': [checked: boolean]
  'bulk:setActive': [value: boolean]
  'bulk:setDefault': [value: boolean]
  'bulk:delete': []
  'row:edit': [item: Item]
  'row:duplicate': [item: Item]
  'row:toggleActive': [item: Item]
  'row:toggleDefault': [item: Item]
  'row:delete': [item: Item]
}>()

const detailModeOptions: Array<{ label: string; value: 'simple' | 'expanded' }> = [
  { label: 'Simple', value: 'simple' },
  { label: 'Expanded', value: 'expanded' },
]

const detailRowArrowImageSrc = 'https://assets.streamlinehq.com/image/private/w_300,h_300,ar_1/f_auto/v1/icons/3/long-arrow-down-right-mbokvtnvyn8z8i5oh7j9v.png/long-arrow-down-right-xo63kk83xpiwai9nokmz4a.png?_a=DATAiZAAZAA0'

const visibleFields = computed(() => props.baseFields)
const allRowsSelected = computed(() => props.items.length > 0 && props.selectedItemIds.length === props.items.length)

const { openActionsForId: openActionsForItemId, menuPosition: rowActionsMenuPosition, closeActions: closeRowActions, toggleActions: toggleRowActions } = useRowActionsMenu({
  menuWidth: 176,
  menuHeight: 170,
  dataElement: 'items-row-actions'
})

const activeMenuItem = computed(() => {
  if (!openActionsForItemId.value) return null
  return props.items.find(item => item.id === openActionsForItemId.value) ?? null
})

type ExpandedFieldDisplay = {
  key: string
  label: string
  value: string
  href?: string
  booleanValue?: boolean | null
}

const getExpandedFieldDisplays = (item: Item): ExpandedFieldDisplay[] => {
  const displays: ExpandedFieldDisplay[] = []

  for (const field of props.extraFields) {
    const href = field.renderHref?.(item)
    if (href) {
      displays.push({
        key: field.key,
        label: field.label,
        value: 'URL',
        href,
      })
      continue
    }

    const booleanValue = field.renderBoolean?.(item)
    if (typeof booleanValue === 'boolean') {
      displays.push({
        key: field.key,
        label: field.label,
        value: '',
        booleanValue,
      })
      continue
    }

    const value = field.render(item)
    if (!value || value === 'Not set') {
      continue
    }

    displays.push({
      key: field.key,
      label: field.label,
      value,
    })
  }

  return displays
}
</script>

<template>
  <section class="surface-panel w-full overflow-visible">
    <div class="border-line-subtle relative border-b py-3 pl-12 pr-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-3">
          <h3 class="heading-section">{{ title }}</h3>
          <div class="hidden md:block">
            <AppToggleGroup :name="`items-table-detail-mode-${title.toLowerCase().replace(/\s+/g, '-')}`"
              data-element="items-table-detail-mode" :model-value="tableDetailMode" :options="detailModeOptions"
              fit-content
              @update:model-value="(value) => emit('update:tableDetailMode', value as 'simple' | 'expanded')" />
          </div>
          <template v-if="selectionMode">
            <span class="text-copy-subtle text-xs font-medium">{{ selectedItemIds.length }} selected</span>
            <button type="button"
              class="border-line-subtle text-copy hover:bg-surface-soft rounded-md border px-2 py-1 text-xs font-medium"
              :disabled="selectedItemIds.length === 0" @click="emit('bulk:setActive', true)">
              Set active
            </button>
            <button type="button"
              class="border-line-subtle text-copy hover:bg-surface-soft rounded-md border px-2 py-1 text-xs font-medium"
              :disabled="selectedItemIds.length === 0" @click="emit('bulk:setActive', false)">
              Set inactive
            </button>
            <button type="button"
              class="border-line-subtle text-copy hover:bg-surface-soft rounded-md border px-2 py-1 text-xs font-medium"
              :disabled="selectedItemIds.length === 0" @click="emit('bulk:setDefault', true)">
              Set default
            </button>
            <button type="button"
              class="border-line-subtle text-copy hover:bg-surface-soft rounded-md border px-2 py-1 text-xs font-medium"
              :disabled="selectedItemIds.length === 0" @click="emit('bulk:setDefault', false)">
              Clear default
            </button>
            <button type="button"
              class="rounded-md border border-red-200 px-2 py-1 text-xs font-medium text-red-700 hover:bg-red-50"
              :disabled="selectedItemIds.length === 0" @click="emit('bulk:delete')">
              Delete selected
            </button>
            <button type="button"
              class="text-copy-muted hover:text-copy text-xs font-medium underline underline-offset-2"
              @click="emit('update:selectionMode', false)">
              Clear
            </button>
          </template>
        </div>

        <div class="text-copy-subtle flex flex-wrap items-center gap-3 text-xs font-medium">
          <span>{{ items.length }} gear items</span>
          <span>Weight: {{ totalWeightLabel }}</span>
          <span>Value: {{ totalValueLabel }}</span>
        </div>
      </div>
    </div>

    <div v-if="items.length === 0" class="text-copy-subtle px-4 py-6 text-sm">
      No gear in this section.
    </div>

    <div v-else class="overflow-x-auto overflow-y-visible">
      <table class="divide-line text-copy min-w-full table-auto divide-y text-sm">
        <caption class="sr-only">{{ title }} gear items with {{ visibleFields.length }} columns</caption>
        <thead class="bg-surface-muted text-copy-subtle text-left text-xs font-semibold uppercase tracking-[0.06em]">
          <tr>
            <th class="w-8 px-3 py-3">
              <input type="checkbox" :checked="allRowsSelected" :aria-label="`Select all rows in ${title}`" @change="(event) => {
                const checked = (event.target as HTMLInputElement).checked
                if (checked && !selectionMode) emit('update:selectionMode', true)
                emit('toggle:selectAll', checked)
              }" />
            </th>
            <th class="w-80 px-4 py-3">Name</th>
            <th v-for="field in visibleFields" :key="field.key" class="whitespace-nowrap px-4 py-3">{{ field.label }}
            </th>
            <th class="w-px px-4 py-3 text-right">
              <span class="sr-only">Row actions</span>
            </th>
          </tr>
        </thead>
        <tbody class="divide-line divide-y">
          <template v-for="item in items" :key="item.id">
            <tr :data-item-id="item.id" class="group hover:bg-surface-muted" :class="{ 'opacity-50': !item.is_active }">
              <td class="w-8 px-3 py-3 align-top">
                <input type="checkbox" :checked="selectionMode && selectedItemIds.includes(item.id)"
                  :aria-label="`Select ${item.name}`" class="transition-opacity" @change="(event) => {
                    const checked = (event.target as HTMLInputElement).checked
                    if (!selectionMode) emit('update:selectionMode', true)
                    emit('toggle:itemSelection', item.id, checked)
                  }" />
              </td>
              <AppItemTableRowContent :item="item" :visible-fields="visibleFields"
                :item-labels="itemLabelsMap?.get(item.id) ?? []" @open-details="emit('openDetails', $event)" />
              <td class="w-px px-4 py-3 align-top">
                <div data-element="items-row-actions" class="relative flex justify-end">
                  <button type="button"
                    class="text-copy-muted hover:text-copy hover:bg-surface-soft inline-flex h-7 w-7 items-center justify-center rounded-full transition"
                    :aria-label="`Open actions for ${item.name}`" @click="(event) => toggleRowActions(item.id, event)">
                    <i class="pi pi-ellipsis-h text-xs" aria-hidden="true" />
                  </button>
                </div>
              </td>
            </tr>

            <tr v-if="tableDetailMode === 'expanded'" class="bg-surface-muted/40 hidden md:table-row">
              <td :colspan="visibleFields.length + 3" class="border-line-subtle border-t px-4 py-2 align-top">
                <div class="relative ml-8 pl-3">
                  <img :src="detailRowArrowImageSrc" alt="" aria-hidden="true"
                    class="pointer-events-none absolute -left-7 top-0.5 h-3 w-3 select-none opacity-70" />
                  <div v-if="getExpandedFieldDisplays(item).length > 0" class="flex flex-wrap gap-x-4 gap-y-1 text-xs">
                    <span v-for="field in getExpandedFieldDisplays(item)" :key="`${item.id}-expanded-${field.key}`"
                      class="inline-flex items-center gap-1">
                      <span class="text-copy font-medium">{{ field.label }}:</span>
                      <a v-if="field.href" :href="field.href" target="_blank" rel="noreferrer"
                        class="text-brand-500 inline-flex items-center"
                        :aria-label="`Open ${field.label} for ${item.name}`">
                        <i class="pi pi-external-link" aria-hidden="true"></i>
                        <span class="sr-only">Open {{ field.label }}</span>
                      </a>
                      <AppBooleanValue v-else-if="typeof field.booleanValue === 'boolean'" :value="field.booleanValue"
                        :label="field.label" />
                      <span v-else>{{ field.value }}</span>
                    </span>
                  </div>
                  <span v-else class="text-copy-muted text-xs">No additional details.</span>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>
  </section>

  <!-- Teleport menu to body to escape overflow container -->
  <Teleport to="body">
    <div v-if="activeMenuItem" data-element="items-row-actions-menu"
      class="border-line-subtle bg-surface-elevated fixed z-30 w-44 rounded-lg border py-1 shadow-sm" :style="{
        top: `${rowActionsMenuPosition.top}px`,
        left: `${rowActionsMenuPosition.left}px`,
      }">
      <button type="button"
        class="text-copy-subtle hover:text-copy hover:bg-surface-soft block w-full px-3 py-2 text-left text-xs font-medium"
        @click="emit('row:edit', activeMenuItem); closeRowActions()">
        Edit
      </button>
      <button type="button"
        class="text-copy-subtle hover:text-copy hover:bg-surface-soft block w-full px-3 py-2 text-left text-xs font-medium"
        @click="emit('row:duplicate', activeMenuItem); closeRowActions()">
        Duplicate
      </button>
      <button type="button"
        class="text-copy-subtle hover:text-copy hover:bg-surface-soft block w-full px-3 py-2 text-left text-xs font-medium"
        @click="emit('row:toggleActive', activeMenuItem); closeRowActions()">
        {{ activeMenuItem.is_active ? 'Deactivate' : 'Activate' }}
      </button>
      <button type="button"
        class="text-copy-subtle hover:text-copy hover:bg-surface-soft block w-full px-3 py-2 text-left text-xs font-medium"
        @click="emit('row:toggleDefault', activeMenuItem); closeRowActions()">
        {{ activeMenuItem.is_default ? 'Unset default' : 'Set default' }}
      </button>
      <button type="button" class="block w-full px-3 py-2 text-left text-xs font-medium text-red-700 hover:bg-red-50"
        @click="emit('row:delete', activeMenuItem); closeRowActions()">
        Delete
      </button>
    </div>
  </Teleport>
</template>
