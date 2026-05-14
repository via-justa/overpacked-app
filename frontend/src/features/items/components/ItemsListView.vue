<script setup lang="ts">
import type { Item } from '../types'
import ItemCard from './ItemCard.vue'
import ItemsTypeTable from './ItemsTypeTable.vue'

type TableField = {
  key: string
  label: string
  render: (item: Item) => string
  renderHref?: (item: Item) => string | undefined
  renderBoolean?: (item: Item) => boolean | null | undefined
}

interface ItemTableSection {
  type: string
  title: string
  items: Item[]
  baseFields: TableField[]
  extraFields: TableField[]
  tableDetailMode: 'simple' | 'expanded'
  selectionMode: boolean
  selectedItemIds: string[]
  totalWeightLabel: string
  totalValueLabel: string
}

defineProps<{
  viewMode: 'cards' | 'table'
  items: Item[]
  tableSections: ItemTableSection[]
  getImageSrc: (item: Item) => string
}>()

defineEmits<{
  openDetails: [item: Item]
  'update:tableDetailMode': [type: string, mode: 'simple' | 'expanded']
  'update:tableSelectionMode': [type: string, value: boolean]
  'toggle:tableItemSelection': [type: string, itemId: string, checked: boolean]
  'toggle:tableSelectAll': [type: string, checked: boolean]
  'bulk:setActive': [type: string, value: boolean]
  'bulk:setDefault': [type: string, value: boolean]
  'bulk:delete': [type: string]
  'row:edit': [item: Item]
  'row:duplicate': [item: Item]
  'row:toggleActive': [item: Item]
  'row:toggleDefault': [item: Item]
  'row:delete': [item: Item]
}>()
</script>

<template>
  <div v-if="viewMode === 'table'" data-element="items-table-view" class="space-y-6">
    <ItemsTypeTable v-for="section in tableSections" :key="section.type" :title="section.title" :items="section.items"
      :base-fields="section.baseFields" :extra-fields="section.extraFields" :table-detail-mode="section.tableDetailMode"
      :selection-mode="section.selectionMode" :selected-item-ids="section.selectedItemIds"
      :total-weight-label="section.totalWeightLabel" :total-value-label="section.totalValueLabel"
      @open-details="$emit('openDetails', $event)"
      @update:table-detail-mode="$emit('update:tableDetailMode', section.type, $event)"
      @update:selection-mode="$emit('update:tableSelectionMode', section.type, $event)"
      @toggle:item-selection="(itemId, checked) => $emit('toggle:tableItemSelection', section.type, itemId, checked)"
      @toggle:select-all="$emit('toggle:tableSelectAll', section.type, $event)"
      @bulk:set-active="$emit('bulk:setActive', section.type, $event)"
      @bulk:set-default="$emit('bulk:setDefault', section.type, $event)"
      @bulk:delete="$emit('bulk:delete', section.type)" @row:edit="$emit('row:edit', $event)"
      @row:duplicate="$emit('row:duplicate', $event)" @row:toggle-active="$emit('row:toggleActive', $event)"
      @row:toggle-default="$emit('row:toggleDefault', $event)" @row:delete="$emit('row:delete', $event)" />
  </div>

  <div v-else data-element="items-card-view" class="grid gap-4 sm:grid-cols-3 xl:grid-cols-4">
    <ItemCard v-for="item in items" :key="item.id" :item="item" :image-src="getImageSrc(item)"
      @open-details="$emit('openDetails', $event)">
      <template #additional-info>
        <slot name="card-additional-info" :item="item" />
      </template>
    </ItemCard>
  </div>
</template>
