<script setup lang="ts">
import type { Item, Label } from '../types'
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
  tableDetailMode: 'simple' | 'expanded'
  selectionMode: boolean
  selectedItemIds: string[]
  totalWeightLabel: string
  totalValueLabel: string
  itemLabelsMap: Map<string, Label[]>
}

defineProps<{
  viewMode: 'cards' | 'table'
  items: Item[]
  tableSections: ItemTableSection[]
  getImageSrc: (item: Item) => string
  itemLabelsMap: Map<string, Label[]>
}>()

defineEmits<{
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
  <div v-if="viewMode === 'table'" data-element="items-table-view" class="w-full space-y-6">
    <ItemsTypeTable v-for="section in tableSections" :key="section.type" :title="section.title" :items="section.items"
      :base-fields="section.baseFields" :table-detail-mode="section.tableDetailMode"
      :selection-mode="section.selectionMode" :selected-item-ids="section.selectedItemIds"
      :total-weight-label="section.totalWeightLabel" :total-value-label="section.totalValueLabel"
      :item-labels-map="section.itemLabelsMap" @edit="$emit('row:edit', $event)"
      @update:selection-mode="$emit('update:tableSelectionMode', section.type, $event)"
      @toggle:item-selection="(itemId, checked) => $emit('toggle:tableItemSelection', section.type, itemId, checked)"
      @toggle:select-all="$emit('toggle:tableSelectAll', section.type, $event)"
      @bulk:set-active="$emit('bulk:setActive', section.type, $event)"
      @bulk:set-default="$emit('bulk:setDefault', section.type, $event)"
      @bulk:delete="$emit('bulk:delete', section.type)" @row:edit="$emit('row:edit', $event)"
      @row:duplicate="$emit('row:duplicate', $event)" @row:toggle-active="$emit('row:toggleActive', $event)"
      @row:toggle-default="$emit('row:toggleDefault', $event)" @row:delete="$emit('row:delete', $event)" />
  </div>

  <div v-else data-element="items-card-view" class="grid gap-2 sm:gap-4 sm:grid-cols-3 xl:grid-cols-5">
    <ItemCard v-for="item in items" :key="item.id" :item="item" :image-src="getImageSrc(item)"
      :item-labels="itemLabelsMap.get(item.id) ?? []" @edit="$emit('row:edit', $event)">
      <template #additional-info>
        <slot name="card-additional-info" :item="item" />
      </template>
    </ItemCard>
  </div>
</template>
