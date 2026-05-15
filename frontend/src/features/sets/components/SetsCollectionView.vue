<script setup lang="ts">
import { computed } from 'vue'
import Button from 'primevue/button'
import AppItemTableRowContent from '../../../components/AppItemTableRowContent.vue'
import AppToggleGroup from '../../../components/AppToggleGroup.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { useRowActionsMenu } from '../../../composables/useRowActionsMenu'
import type { AppItemTableField } from '../../../components/AppItemTableRowContent.vue'
import type { Item } from '../../items/types'
import type { ItemSet, SetItemWithDetails } from '../types'

type SetStats = {
  itemCount: number
  totalWeightGrams: number
}

type Props = {
  sets: ItemSet[]
  setsViewMode: 'cards' | 'table'
  tableDetailMode: 'simple' | 'expanded'
  selectionMode: boolean
  selectedSetIds: string[]
  setStatsById: Record<string, SetStats>
  setItemsBySetId: Record<string, SetItemWithDetails[]>
  itemTableFields: AppItemTableField[]
  getItemTypeLabel: (categoryId: string) => string
  formatDisplayWeight: (grams: number) => string
  formatDate: (value: string) => string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  openDetails: [set: ItemSet]
  openItemDetails: [item: Item]
  startEdit: [set: ItemSet]
  requestDelete: [set: ItemSet]
  'update:tableDetailMode': [mode: 'simple' | 'expanded']
  'update:selectionMode': [value: boolean]
  'toggle:setSelection': [setId: string, checked: boolean]
  'toggle:selectAll': [checked: boolean]
  'bulk:delete': []
}>()

type TableField = {
  key: 'category' | 'items' | 'weight'
  label: string
  render: (set: ItemSet) => string
}

type ExpandedFieldDisplay = {
  key: string
  label: string
  value: string
}

const detailModeOptions: Array<{ label: string; value: 'simple' | 'expanded' }> = [
  { label: 'Simple', value: 'simple' },
  { label: 'Expanded', value: 'expanded' },
]

const detailRowArrowImageSrc = 'https://assets.streamlinehq.com/image/private/w_300,h_300,ar_1/f_auto/v1/icons/3/long-arrow-down-right-mbokvtnvyn8z8i5oh7j9v.png/long-arrow-down-right-xo63kk83xpiwai9nokmz4a.png?_a=DATAiZAAZAA0'

const visibleFields = computed<TableField[]>(() => [
  {
    key: 'category',
    label: 'Category',
    render: (set) => props.getItemTypeLabel(set.set_category),
  },
  {
    key: 'items',
    label: 'Items',
    render: (set) => String(props.setStatsById[set.id]?.itemCount ?? 0),
  },
  {
    key: 'weight',
    label: 'Weight',
    render: (set) => props.formatDisplayWeight(props.setStatsById[set.id]?.totalWeightGrams ?? 0),
  },
])

const allRowsSelected = computed(
  () => props.sets.length > 0 && props.selectedSetIds.length === props.sets.length,
)

const { openActionsForId: openActionsForSetId, menuPosition: rowActionsMenuPosition, closeActions: closeRowActions, toggleActions: toggleRowActions } = useRowActionsMenu({
  menuWidth: 176,
  menuHeight: 110,
  dataElement: 'sets-row-actions'
})

const activeMenuSet = computed(() => {
  if (!openActionsForSetId.value) return null
  return props.sets.find(set => set.id === openActionsForSetId.value) ?? null
})

const getExpandedFieldDisplays = (set: ItemSet): ExpandedFieldDisplay[] => {
  return [
    {
      key: 'updated',
      label: 'Updated',
      value: props.formatDate(set.updated_at),
    },
  ]
}
</script>

<template>
  <div v-if="setsViewMode === 'cards'" class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
    <article v-for="set in sets" :key="set.id"
      class="surface-panel p-4">
      <h3 class="text-ink text-lg font-semibold">{{ normalizeTitleWords(set.name) }}</h3>
      <p class="text-copy-muted mt-2 text-sm">
        {{ getItemTypeLabel(set.set_category) }}
        <span class="text-line mx-2">/</span>
        {{ setStatsById[set.id]?.itemCount ?? 0 }} items
        <span class="text-line mx-2">/</span>
        {{ formatDisplayWeight(setStatsById[set.id]?.totalWeightGrams ?? 0) }}
      </p>
      <p class="text-copy-subtle mt-1 text-xs">Updated {{ formatDate(set.updated_at) }}</p>

      <div class="mt-4 flex flex-wrap gap-2">
        <Button size="small" label="Manage set gear" icon="pi pi-folder-open" outlined
          @click="emit('openDetails', set)" />
        <Button size="small" label="Edit" icon="pi pi-pencil" outlined @click="emit('startEdit', set)" />
        <Button size="small" label="Delete" icon="pi pi-trash" severity="danger" outlined
          @click="emit('requestDelete', set)" />
      </div>
    </article>
  </div>

  <section v-else class="surface-panel overflow-visible">
    <div class="border-line-subtle relative border-b py-3 pl-12 pr-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-3">
          <h3 class="text-copy text-sm font-semibold uppercase tracking-[0.08em]">Sets</h3>
          <AppToggleGroup name="sets-table-detail-mode" data-element="sets-table-detail-mode"
            :model-value="tableDetailMode" :options="detailModeOptions" fit-content
            @update:model-value="(value) => emit('update:tableDetailMode', value as 'simple' | 'expanded')" />
          <template v-if="selectionMode">
            <span class="text-copy-subtle text-xs font-medium">{{ selectedSetIds.length }} selected</span>
            <button type="button"
              class="rounded-md border border-red-200 px-2 py-1 text-xs font-medium text-red-700 hover:bg-red-50"
              :disabled="selectedSetIds.length === 0" @click="emit('bulk:delete')">
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
          <span>{{ sets.length }} sets</span>
        </div>
      </div>
    </div>

    <div v-if="sets.length === 0" class="text-copy-subtle px-4 py-6 text-sm">
      No sets in this section.
    </div>

    <div v-else class="overflow-x-auto overflow-y-visible">
      <table class="divide-line text-copy min-w-full table-fixed divide-y text-sm">
        <thead class="bg-surface-muted text-copy-subtle text-left text-xs font-semibold uppercase tracking-[0.06em]">
          <tr>
            <th class="w-8 px-3 py-3">
              <input type="checkbox" :checked="allRowsSelected" aria-label="Select all sets" @change="(event) => {
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
          <template v-for="set in sets" :key="set.id">
            <tr :data-set-id="set.id" class="group hover:bg-surface-muted">
              <td class="w-8 px-3 py-3 align-top">
                <input type="checkbox" :checked="selectionMode && selectedSetIds.includes(set.id)"
                  :aria-label="`Select ${set.name}`" class="transition-opacity" @change="(event) => {
                    const checked = (event.target as HTMLInputElement).checked
                    if (!selectionMode) emit('update:selectionMode', true)
                    emit('toggle:setSelection', set.id, checked)
                  }" />
              </td>
              <td class="w-80 px-4 py-3 align-top">
                <button type="button"
                  class="text-brand-500 decoration-brand-200 block max-w-full truncate text-left font-semibold underline underline-offset-2"
                  :title="normalizeTitleWords(set.name)" @click="emit('openDetails', set)">
                  {{ normalizeTitleWords(set.name) }}
                </button>
              </td>
              <td v-for="field in visibleFields" :key="`${set.id}-${field.key}`"
                class="whitespace-nowrap px-4 py-3 align-top">
                {{ field.render(set) }}
              </td>
              <td class="w-px px-4 py-3 align-top">
                <div data-element="sets-row-actions" class="relative flex justify-end">
                  <button type="button"
                    class="text-copy-muted hover:text-copy hover:bg-surface-soft inline-flex h-7 w-7 items-center justify-center rounded-full transition"
                    :aria-label="`Open actions for ${set.name}`" @click="(event) => toggleRowActions(set.id, event)">
                    <i class="pi pi-ellipsis-h text-xs" aria-hidden="true" />
                  </button>
                </div>
              </td>
            </tr>

            <tr v-if="tableDetailMode === 'expanded'" class="bg-surface-muted/40">
              <td :colspan="visibleFields.length + 3" class="border-line-subtle border-t px-4 py-2 align-top">
                <div class="relative ml-8 pl-3">
                  <img :src="detailRowArrowImageSrc" alt="" aria-hidden="true"
                    class="pointer-events-none absolute -left-7 top-0.5 h-3 w-3 select-none opacity-70" />
                  <div class="space-y-2">
                    <div class="flex flex-wrap gap-x-4 gap-y-1 text-xs">
                      <span v-for="field in getExpandedFieldDisplays(set)" :key="`${set.id}-expanded-${field.key}`"
                        class="inline-flex items-center gap-1">
                        <span class="text-copy font-medium">{{ field.label }}:</span>
                        <span>{{ field.value }}</span>
                      </span>
                    </div>

                    <div v-if="(setItemsBySetId[set.id] ?? []).length > 0"
                      class="border-line-subtle overflow-x-auto rounded-lg border bg-surface-elevated">
                      <table class="divide-line text-copy min-w-full table-fixed divide-y text-sm">
                        <thead
                          class="bg-surface-muted text-copy-subtle text-left text-xs font-semibold uppercase tracking-[0.06em]">
                          <tr>
                            <th class="w-80 px-4 py-3">Name</th>
                            <th v-for="field in itemTableFields" :key="`set-${set.id}-field-${field.key}`"
                              class="whitespace-nowrap px-4 py-3">{{ field.label }}</th>
                          </tr>
                        </thead>
                        <tbody class="divide-line divide-y">
                          <tr v-for="entry in (setItemsBySetId[set.id] ?? [])" :key="`${set.id}-${entry.item_id}`"
                            class="group hover:bg-surface-muted">
                            <AppItemTableRowContent :item="entry.item" :visible-fields="itemTableFields"
                              @open-details="emit('openItemDetails', $event)" />
                          </tr>
                        </tbody>
                      </table>
                    </div>
                    <div v-else class="text-copy-muted text-xs">No items in this set yet.</div>
                  </div>
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
    <div v-if="activeMenuSet" data-element="sets-row-actions-menu"
      class="border-line-subtle bg-surface-elevated fixed z-30 w-44 rounded-lg border py-1 shadow-sm"
      :style="{
        top: `${rowActionsMenuPosition.top}px`,
        left: `${rowActionsMenuPosition.left}px`,
      }">
      <button type="button"
        class="text-copy-subtle hover:text-copy hover:bg-surface-soft block w-full px-3 py-2 text-left text-xs font-medium"
        @click="emit('openDetails', activeMenuSet); closeRowActions()">
        Manage set gear
      </button>
      <button type="button"
        class="text-copy-subtle hover:text-copy hover:bg-surface-soft block w-full px-3 py-2 text-left text-xs font-medium"
        @click="emit('startEdit', activeMenuSet); closeRowActions()">
        Edit
      </button>
      <button type="button"
        class="block w-full px-3 py-2 text-left text-xs font-medium text-red-700 hover:bg-red-50"
        @click="emit('requestDelete', activeMenuSet); closeRowActions()">
        Delete
      </button>
    </div>
  </Teleport>
</template>
