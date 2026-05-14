<script setup lang="ts">
import { computed } from 'vue'
import Button from 'primevue/button'
import AppSelect from '../../../components/AppSelect.vue'
import AppTemplateDialog from '../../../components/AppTemplateDialog.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import type { Item } from '../../items/types'
import type { ItemSet, SetItemWithDetails } from '../types'

type SetStats = {
  itemCount: number
  totalWeightGrams: number
}

type DraftField = 'quantity' | 'notes'

type Props = {
  modelValue: boolean
  activeSet: ItemSet | null
  activeSetItems: SetItemWithDetails[]
  activeSetStats: SetStats
  setItemsLoading: boolean
  availableItemsForAdd: Item[]
  addItemId: string
  addItemQuantity: string
  addItemNotes: string
  isAddingItem: boolean
  itemDraftsById: Record<string, { quantity: string; notes: string }>
  savingItemIds: Record<string, boolean>
  getItemTypeLabel: (categoryId: string) => string
  formatDisplayWeight: (grams: number) => string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:addItemId': [value: string]
  'update:addItemQuantity': [value: string]
  'update:addItemNotes': [value: string]
  updateItemDraft: [payload: { itemId: string; field: DraftField; value: string }]
  addItem: []
  saveSetItem: [itemId: string]
  requestRemoveSetItem: [payload: { itemId: string; itemName: string }]
}>()

const addItemIdModel = computed({
  get: () => props.addItemId,
  set: (value: string) => emit('update:addItemId', value),
})

const addItemQuantityModel = computed({
  get: () => props.addItemQuantity,
  set: (value: string) => emit('update:addItemQuantity', value),
})

const addItemNotesModel = computed({
  get: () => props.addItemNotes,
  set: (value: string) => emit('update:addItemNotes', value),
})

const closeDialog = () => {
  emit('update:modelValue', false)
}

const getDraftValue = (itemId: string, field: DraftField): string => {
  const draft = props.itemDraftsById[itemId]
  if (!draft) {
    return ''
  }
  return field === 'quantity' ? draft.quantity : draft.notes
}

const onDraftInput = (itemId: string, field: DraftField, event: Event) => {
  const target = event.target as HTMLInputElement
  emit('updateItemDraft', { itemId, field, value: target.value })
}

const onDraftBlur = (itemId: string) => {
  emit('saveSetItem', itemId)
}
</script>

<template>
  <AppTemplateDialog :model-value="modelValue" data-element="set-details-dialog" width="min(64rem, calc(100vw - 2rem))"
    @update:model-value="(value) => { if (!value) closeDialog() }">
    <article v-if="activeSet" class="border-line-subtle bg-surface-elevated rounded-2xl border p-4 shadow-panel">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 class="text-ink text-xl font-semibold">{{ normalizeTitleWords(activeSet.name) }}</h2>
          <p class="text-copy-muted mt-1 text-sm">
            Category: {{ getItemTypeLabel(activeSet.set_category) }}
            <span class="text-line mx-2">/</span>
            {{ activeSetStats.itemCount }} items
            <span class="text-line mx-2">/</span>
            {{ formatDisplayWeight(activeSetStats.totalWeightGrams) }}
          </p>
        </div>

        <Button label="Close" severity="secondary" outlined @click="closeDialog" />
      </div>

      <section class="border-line-subtle bg-surface-muted mt-4 rounded-xl border p-3">
        <h3 class="heading-section">Add Item</h3>
        <div class="mt-3 grid gap-2 sm:grid-cols-[1fr_8rem_1fr_auto]">
          <AppSelect v-model="addItemIdModel">
            <option value="">Select gear item</option>
            <option v-for="item in availableItemsForAdd" :key="item.id" :value="item.id">
              {{ normalizeTitleWords(item.name) }}
            </option>
          </AppSelect>
          <input v-model="addItemQuantityModel" class="input-shell" type="number" min="0.1" step="0.1"
            placeholder="Qty" />
          <input v-model="addItemNotesModel" class="input-shell" type="text" placeholder="Notes (optional)" />
          <Button label="Add" icon="pi pi-plus" :loading="isAddingItem" :disabled="!addItemIdModel"
            @click="emit('addItem')" />
        </div>
      </section>

      <div class="border-line-subtle mt-4 overflow-x-auto rounded-xl border">
        <table class="divide-line min-w-full divide-y text-sm">
          <thead class="bg-surface-muted text-copy-subtle text-left text-xs font-semibold uppercase tracking-[0.06em]">
            <tr>
              <th class="px-3 py-2">Item</th>
              <th class="px-3 py-2">Qty</th>
              <th class="px-3 py-2">Notes</th>
              <th class="px-3 py-2 text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-line divide-y">
            <tr v-if="setItemsLoading">
              <td colspan="4" class="text-copy-muted px-3 py-3">Loading set items...</td>
            </tr>
            <tr v-else-if="activeSetItems.length === 0">
              <td colspan="4" class="text-copy-muted px-3 py-3">No items in this set yet.</td>
            </tr>
            <tr v-for="entry in activeSetItems" :key="entry.item_id">
              <td class="px-3 py-2">
                <span class="text-copy font-medium">{{ normalizeTitleWords(entry.item.name) }}</span>
              </td>
              <td class="px-3 py-2">
                <input :value="getDraftValue(entry.item_id, 'quantity')" class="input-shell w-24" type="number"
                  min="0.1" step="0.1" :disabled="savingItemIds[entry.item_id]"
                  @input="onDraftInput(entry.item_id, 'quantity', $event)" @blur="onDraftBlur(entry.item_id)" />
              </td>
              <td class="px-3 py-2">
                <input :value="getDraftValue(entry.item_id, 'notes')" class="input-shell" type="text"
                  placeholder="Optional notes" :disabled="savingItemIds[entry.item_id]"
                  @input="onDraftInput(entry.item_id, 'notes', $event)" @blur="onDraftBlur(entry.item_id)" />
              </td>
              <td class="px-3 py-2 text-right">
                <Button size="small" label="Remove" icon="pi pi-trash" severity="danger" outlined
                  :disabled="savingItemIds[entry.item_id]"
                  @click="emit('requestRemoveSetItem', { itemId: entry.item_id, itemName: entry.item.name })" />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>
  </AppTemplateDialog>
</template>
