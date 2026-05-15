<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import Button from 'primevue/button'
import AppSelect from '../../../components/AppSelect.vue'
import AppTemplateDialog from '../../../components/AppTemplateDialog.vue'
import AppNotSetValue from '../../../components/AppNotSetValue.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import type { Item, ItemTypeEntity } from '../../items/types'
import type { ItemSet, SetItemWithDetails } from '../types'

type SetStats = {
  itemCount: number
  totalWeightGrams: number
  totalValue: number
}

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
  getItemTypeLabel: (categoryId: string) => string
  formatDisplayWeight: (grams: number) => string
  setNameInput: string
  setCategoryInput: string
  itemTypeOptions: ItemTypeEntity[]
  isSubmittingSet: boolean
  manufacturersById: Map<string, string>
  volumeInputUnit: 'ml' | 'fl_oz'
  currency: 'usd' | 'eur'
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:addItemId': [value: string]
  'update:addItemQuantity': [value: string]
  'update:addItemNotes': [value: string]
  'update:setNameInput': [value: string]
  'update:setCategoryInput': [value: string]
  addItem: []
  saveSet: []
  requestCategoryChange: [newCategory: string]
  requestRemoveSetItem: [payload: { itemId: string; itemName: string }]
  editSetItem: [payload: { itemId: string; quantity: number; notes: string }]
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

const setNameModel = computed({
  get: () => props.setNameInput,
  set: (value: string) => emit('update:setNameInput', value),
})

const setCategoryModel = computed({
  get: () => props.setCategoryInput,
  set: (value: string) => emit('update:setCategoryInput', value),
})

const closeDialog = () => {
  emit('update:modelValue', false)
}

const sortedItemTypeOptions = computed(() => {
  return [...props.itemTypeOptions].sort((a, b) => {
    const nameA = normalizeTitleWords(a.name)
    const nameB = normalizeTitleWords(b.name)
    return nameA.localeCompare(nameB)
  })
})

// Auto-save logic
const isInitialized = ref(false)
const initialCategory = ref('')

const shouldSave = (): boolean => {
  if (!isInitialized.value || !props.activeSet) {
    return false
  }

  const name = normalizeTitleWords(props.setNameInput.trim())
  const category = props.setCategoryInput.trim()

  // Only save if values are valid and different from current
  if (!name || !category) {
    return false
  }

  if (name === normalizeTitleWords(props.activeSet.name) && category === props.activeSet.set_category) {
    return false
  }

  return true
}

const onNameBlur = () => {
  if (shouldSave()) {
    emit('saveSet')
  }
}

// Watch for category changes and emit request for validation
watch(() => props.setCategoryInput, (newCategory) => {
  // Only process if initialized and the value changed from the initial value
  if (isInitialized.value && newCategory !== initialCategory.value) {
    emit('requestCategoryChange', newCategory)
  }
})

// Track dialog open state and initialize
watch(() => props.modelValue, (isOpen) => {
  if (isOpen) {
    // Reset initialization flag when dialog opens
    isInitialized.value = false
    // Capture initial category value
    initialCategory.value = props.setCategoryInput
    // Set initialized after a short delay to allow props to settle
    setTimeout(() => {
      isInitialized.value = true
    }, 100)
  } else {
    isInitialized.value = false
    initialCategory.value = ''
  }
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
  const currencySymbol = props.currency === 'usd' ? '$' : '€'
  return `${item.value.toFixed(2)} ${currencySymbol}`
}

const isEditingExistingItem = computed(() => {
  if (!props.addItemId) {
    return false
  }
  return props.activeSetItems.some((entry) => entry.item_id === props.addItemId)
})

// Auto-populate quantity with item's default_quantity when item is selected
watch(() => props.addItemId, (newItemId) => {
  if (!newItemId) {
    return
  }

  const selectedItem = props.availableItemsForAdd.find(item => item.id === newItemId)
  if (selectedItem && typeof selectedItem.default_quantity === 'number') {
    emit('update:addItemQuantity', String(selectedItem.default_quantity))
  } else {
    emit('update:addItemQuantity', '1')
  }
})
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
            <span class="text-line mx-2">/</span>
            {{ currency === 'usd' ? '$' : '€' }}{{ activeSetStats.totalValue.toFixed(2) }}
          </p>
        </div>

        <Button label="Close" severity="secondary" outlined @click="closeDialog" />
      </div>

      <section class="border-line-subtle bg-surface-muted mt-4 rounded-xl border p-3">
        <div class="mt-3 grid gap-3 sm:grid-cols-2">
          <label class="grid gap-1">
            <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Set name</span>
            <input v-model="setNameModel" class="input-shell" type="text" @blur="onNameBlur" />
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
      </section>

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
          <Button :label="isEditingExistingItem ? 'Update' : 'Add'" :icon="isEditingExistingItem ? 'pi pi-check' : 'pi pi-plus'" 
            :loading="isAddingItem" :disabled="!addItemIdModel"
            @click="emit('addItem')" />
        </div>
      </section>

      <div class="border-line-subtle mt-4 overflow-x-auto rounded-xl border">
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
            <tr v-if="setItemsLoading">
              <td colspan="8" class="text-copy-muted px-3 py-3">Loading set items...</td>
            </tr>
            <tr v-else-if="activeSetItems.length === 0">
              <td colspan="8" class="text-copy-muted px-3 py-3">No items in this set yet.</td>
            </tr>
            <tr v-for="entry in activeSetItems" :key="entry.item_id">
              <td class="px-3 py-2">
                <span class="text-copy font-medium">{{ normalizeTitleWords(entry.item.name) }}</span>
              </td>
              <td class="px-3 py-2">
                <span class="text-copy-subtle text-xs">{{ getManufacturerName(entry.item) }}</span>
              </td>
              <td class="px-3 py-2">
                <span class="text-copy-subtle text-xs">{{ getItemWeight(entry.item) }}</span>
              </td>
              <td class="px-3 py-2">
                <span class="text-copy-subtle text-xs">{{ getItemVolume(entry.item) }}</span>
              </td>
              <td class="px-3 py-2">
                <span class="text-copy-subtle text-xs">{{ getItemValue(entry.item) }}</span>
              </td>
              <td class="px-3 py-2">
                <span class="text-copy-subtle text-xs">{{ entry.quantity }}</span>
              </td>
              <td class="px-3 py-2">
                <div class="group/note relative inline-flex" v-if="entry.notes">
                  <i class="pi pi-file-edit text-copy-subtle hover:text-copy cursor-default text-sm" aria-hidden="true" />
                  <span
                    class="pointer-events-none absolute bottom-full left-1/2 z-20 mb-1.5 w-max max-w-xs -translate-x-1/2 rounded-lg border border-line-subtle bg-surface-elevated px-3 py-2 text-xs text-copy opacity-0 shadow-panel transition-opacity group-hover/note:opacity-100">
                    {{ entry.notes }}
                  </span>
                </div>
                <AppNotSetValue v-else label="Notes" />
              </td>
              <td class="px-3 py-2 text-right">
                <div class="flex items-center justify-end gap-1">
                  <button type="button"
                    class="text-copy-muted hover:text-copy inline-flex h-8 w-8 items-center justify-center rounded-full transition"
                    :aria-label="`Edit ${entry.item.name}`"
                    @click="emit('editSetItem', { itemId: entry.item_id, quantity: entry.quantity, notes: entry.notes ?? '' })">
                    <i class="pi pi-pencil text-sm" aria-hidden="true" />
                  </button>
                  <button type="button"
                    class="inline-flex h-8 w-8 items-center justify-center rounded-full text-red-700 transition hover:text-red-900"
                    :aria-label="`Remove ${entry.item.name} from set`"
                    @click="emit('requestRemoveSetItem', { itemId: entry.item_id, itemName: entry.item.name })">
                    <i class="pi pi-trash text-sm" aria-hidden="true" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>
  </AppTemplateDialog>
</template>
