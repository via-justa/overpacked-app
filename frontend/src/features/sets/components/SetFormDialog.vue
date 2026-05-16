<script setup lang="ts">
import { computed, watch } from 'vue'
import Button from 'primevue/button'
import AppSelect from '../../../components/AppSelect.vue'
import AppFormCreateDialog from '../../../components/AppFormCreateDialog.vue'
import AppNotSetValue from '../../../components/AppNotSetValue.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import type { Item, ItemTypeEntity } from '../../items/types'

type TempItem = {
  itemId: string
  quantity: number
  notes: string
}

type TempItemWithDetails = {
  tempItem: TempItem
  item: Item | undefined
}

type Props = {
  modelValue: boolean
  editingSetId: string | null
  setNameInput: string
  setCategoryInput: string
  itemTypeOptions: ItemTypeEntity[]
  isSubmittingSet: boolean
  availableItemsForAdd: Item[]
  addItemId: string
  addItemQuantity: string
  addItemNotes: string
  tempItems: TempItemWithDetails[]
  manufacturersById: Map<string, string>
  formatDisplayWeight: (grams: number) => string
  formatValue: (value: number) => string
  volumeInputUnit: 'ml' | 'fl_oz'
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:setNameInput': [value: string]
  'update:setCategoryInput': [value: string]
  'update:addItemId': [value: string]
  'update:addItemQuantity': [value: string]
  'update:addItemNotes': [value: string]
  addItem: []
  removeItem: [itemId: string]
  editItem: [payload: { itemId: string; quantity: number; notes: string }]
  submit: []
}>()

const setNameModel = computed({
  get: () => props.setNameInput,
  set: (value: string) => emit('update:setNameInput', value),
})

const setCategoryModel = computed({
  get: () => props.setCategoryInput,
  set: (value: string) => emit('update:setCategoryInput', value),
})

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

const onDialogUpdate = (value: boolean) => {
  emit('update:modelValue', value)
}

const canSubmit = computed(() => {
  return setNameModel.value.trim().length > 0 && setCategoryModel.value.trim().length > 0
})

const sortedItemTypeOptions = computed(() => {
  return [...props.itemTypeOptions].sort((a, b) => {
    const nameA = normalizeTitleWords(a.name)
    const nameB = normalizeTitleWords(b.name)
    return nameA.localeCompare(nameB)
  })
})

const isEditingExistingItem = computed(() => {
  if (!props.addItemId) {
    return false
  }
  return props.tempItems.some((entry) => entry.tempItem.itemId === props.addItemId)
})

const getManufacturerName = (item: Item): string => {
  return props.manufacturersById.get(item.manufacturer_id) ?? 'Unknown'
}

const getItemWeight = (item: Item): string => {
  return typeof item.weight_grams === 'number' ? props.formatDisplayWeight(item.weight_grams) : 'Not set'
}

const getItemVolume = (item: Item): string => {
  if (typeof item.volume_ml !== 'number') {
    return 'Not set'
  }
  if (props.volumeInputUnit === 'fl_oz') {
    const flOz = item.volume_ml / 29.5735295625
    return `${flOz.toFixed(2)} fl oz`
  }
  return `${item.volume_ml.toFixed(2)} ml`
}

const getItemValue = (item: Item): string => {
  if (typeof item.value !== 'number') {
    return 'Not set'
  }
  return props.formatValue(item.value)
}

// Auto-populate quantity with item's default_quantity when item is selected
watch(() => props.addItemId, (newItemId) => {
  if (!newItemId) {
    return
  }

  const selectedItem = props.availableItemsForAdd.find((item) => item.id === newItemId)
  if (selectedItem && typeof selectedItem.default_quantity === 'number') {
    emit('update:addItemQuantity', String(selectedItem.default_quantity))
  } else {
    emit('update:addItemQuantity', '1')
  }
})
</script>

<template>
  <AppFormCreateDialog :open="modelValue" data-element="sets-form-dialog" width="min(56rem, calc(100vw - 2rem))"
    :title="editingSetId ? 'Edit Set' : 'Create Set'" :can-submit="canSubmit" :is-submitting="isSubmittingSet"
    @update:open="onDialogUpdate" @submit="emit('submit')" @cancel="closeDialog">
    <div class="grid gap-3 sm:grid-cols-2">
      <label class="grid gap-1">
        <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Set name</span>
        <input v-model="setNameModel" class="input-shell" type="text" placeholder="Weekend Essentials" />
      </label>
      <div class="grid gap-1">
        <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Category</span>
        <AppSelect v-model="setCategoryModel">
          <option value="">Select category</option>
          <option v-for="itemType in sortedItemTypeOptions" :key="itemType.id" :value="itemType.id">
            {{ normalizeTitleWords(itemType.name) }}
          </option>
        </AppSelect>
      </div>
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
        <Button :label="isEditingExistingItem ? 'Update' : 'Add'"
          :icon="isEditingExistingItem ? 'pi pi-check' : 'pi pi-plus'" :disabled="!addItemIdModel"
          @click="emit('addItem')" />
      </div>
    </section>

    <div v-if="tempItems.length > 0" class="border-line-subtle mt-4 overflow-x-auto rounded-xl border">
      <table class="divide-line min-w-full divide-y text-sm">
        <thead class="bg-surface-muted text-copy-subtle text-left text-xs font-semibold uppercase tracking-[0.06em]">
          <tr>
            <th class="px-3 py-2">Item</th>
            <th class="px-3 py-2">Manufacturer</th>
            <th class="px-3 py-2">Weight</th>
            <th class="px-3 py-2">Volume</th>
            <th class="px-3 py-2">Value</th>
            <th class="px-3 py-2">Qty</th>
            <th class="px-3 py-2">Notes</th>
            <th class="px-3 py-2 text-right">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-line divide-y">
          <tr v-for="entry in tempItems" :key="entry.tempItem.itemId">
            <td class="px-3 py-2">
              <span class="text-copy font-medium">{{ entry.item ? normalizeTitleWords(entry.item.name) : 'Unknown'
              }}</span>
            </td>
            <td class="px-3 py-2">
              <span class="text-copy-subtle text-xs">{{ entry.item ? getManufacturerName(entry.item) : 'Unknown'
              }}</span>
            </td>
            <td class="px-3 py-2">
              <span class="text-copy-subtle text-xs">{{ entry.item ? getItemWeight(entry.item) : 'Not set' }}</span>
            </td>
            <td class="px-3 py-2">
              <span class="text-copy-subtle text-xs">{{ entry.item ? getItemVolume(entry.item) : 'Not set' }}</span>
            </td>
            <td class="px-3 py-2">
              <span class="text-copy-subtle text-xs">{{ entry.item ? getItemValue(entry.item) : 'Not set' }}</span>
            </td>
            <td class="px-3 py-2">
              <span class="text-copy text-sm">{{ entry.tempItem.quantity }}</span>
            </td>
            <td class="px-3 py-2">
              <span v-if="entry.tempItem.notes" class="group/note relative inline-flex"
                :aria-label="entry.tempItem.notes">
                <i class="pi pi-file-edit text-copy-subtle hover:text-copy cursor-default text-sm" aria-hidden="true" />
                <span
                  class="border-line-subtle bg-surface-elevated text-copy pointer-events-none absolute bottom-full left-1/2 mb-2 w-max max-w-xs -translate-x-1/2 rounded-lg border px-2 py-1 text-xs opacity-0 shadow-sm transition group-hover/note:opacity-100">
                  {{ entry.tempItem.notes }}
                </span>
              </span>
              <AppNotSetValue v-else label="Notes" />
            </td>
            <td class="px-3 py-2 text-right">
              <div class="flex items-center justify-end gap-1">
                <button type="button"
                  class="text-copy-muted hover:text-copy inline-flex h-8 w-8 items-center justify-center rounded-full transition"
                  :aria-label="`Edit ${entry.item?.name ?? 'item'}`"
                  @click="emit('editItem', { itemId: entry.tempItem.itemId, quantity: entry.tempItem.quantity, notes: entry.tempItem.notes })">
                  <i class="pi pi-pencil text-sm" aria-hidden="true" />
                </button>
                <button type="button"
                  class="inline-flex h-8 w-8 items-center justify-center rounded-full text-red-700 transition hover:text-red-900"
                  :aria-label="`Remove ${entry.item?.name ?? 'item'} from set`"
                  @click="emit('removeItem', entry.tempItem.itemId)">
                  <i class="pi pi-trash text-sm" aria-hidden="true" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </AppFormCreateDialog>
</template>
