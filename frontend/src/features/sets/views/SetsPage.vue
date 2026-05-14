<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import { useToast } from 'primevue/usetoast'
import AppConfirmDialog from '../../../components/AppConfirmDialog.vue'
import AppToggleGroup from '../../../components/AppToggleGroup.vue'
import AppQueryError from '../../../components/AppQueryError.vue'
import AppLoadingState from '../../../components/AppLoadingState.vue'
import AppEmptyState from '../../../components/AppEmptyState.vue'
import AppSummaryCard from '../../../components/AppSummaryCard.vue'
import type { AppItemTableField } from '../../../components/AppItemTableRowContent.vue'
import ItemDetailsDialog from '../../items/components/ItemDetailsDialog.vue'
import SetDetailsDialog from '../components/SetDetailsDialog.vue'
import SetFormDialog from '../components/SetFormDialog.vue'
import SetsCollectionView from '../components/SetsCollectionView.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { queryClient } from '../../../lib/query/client'
import { listItemTypes, listItems, listManufacturers } from '../../items/api/itemsApi'
import { useSettings } from '../../../composables/useSettings'
import { toRoundedString, formatDisplayWeight, mlToInput } from '../../../lib/units/conversions'
import {
  formatNumber as formatNumberDisplay,
  formatValue as formatValueDisplay,
  formatCarryStatus as formatCarryStatusDisplay,
  formatType as formatTypeDisplay,
  formatText as formatTextDisplay,
} from '../../../lib/format/display'
import {
  addSetItem,
  createSet,
  listSetItems,
  listSets,
  removeSet,
  removeSetItem,
  updateSet,
  updateSetItem,
} from '../api/setsApi'
import type { ItemSet, SetItemWithDetails } from '../types'
import type { Item, ItemTypeEntity } from '../../items/types'

const toast = useToast()
const route = useRoute()
const router = useRouter()

type SetsViewMode = 'cards' | 'table'
type TableDetailMode = 'simple' | 'expanded'
type WeightInputUnit = 'g' | 'oz'
type VolumeInputUnit = 'ml' | 'fl_oz'

const SETS_VIEW_MODE_STORAGE_KEY = 'sets:view-mode'
const SETS_TABLE_DETAIL_MODE_STORAGE_KEY = 'sets:table-detail-mode'

const readStoredSetsViewMode = (): SetsViewMode => {
  if (globalThis.window === undefined) {
    return 'cards'
  }

  const stored = globalThis.localStorage.getItem(SETS_VIEW_MODE_STORAGE_KEY)
  return stored === 'cards' || stored === 'table' ? stored : 'cards'
}

const readStoredSetsTableDetailMode = (): TableDetailMode => {
  if (globalThis.window === undefined) {
    return 'simple'
  }

  const stored = globalThis.localStorage.getItem(SETS_TABLE_DETAIL_MODE_STORAGE_KEY)
  return stored === 'simple' || stored === 'expanded' ? stored : 'simple'
}


const formatDate = (value: string): string => {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) {
    return value
  }

  const day = String(parsed.getDate()).padStart(2, '0')
  const month = String(parsed.getMonth() + 1).padStart(2, '0')
  const year = parsed.getFullYear()
  return `${day}-${month}-${year}`
}

const { weightUnit, volumeUnit, currency } = useSettings()
const weightInputUnit = computed<WeightInputUnit>(() => weightUnit.value)

// Wrapper for formatDisplayWeight that uses current weightInputUnit
const formatWeight = (valueGrams: number): string => {
  return formatDisplayWeight(valueGrams, weightInputUnit.value)
}

const setsQuery = useQuery({
  queryKey: ['sets'],
  queryFn: listSets,
})

const itemsQuery = useQuery({
  queryKey: ['items'],
  queryFn: listItems,
})

const manufacturersQuery = useQuery({
  queryKey: ['manufacturers'],
  queryFn: listManufacturers,
})

const itemTypesQuery = useQuery({
  queryKey: ['item-types'],
  queryFn: listItemTypes,
})

const setsViewMode = ref<SetsViewMode>(readStoredSetsViewMode())
const tableDetailMode = ref<TableDetailMode>(readStoredSetsTableDetailMode())
const tableSelectionMode = ref(false)
const selectedSetIds = ref<string[]>([])
const isFormDialogOpen = ref(false)
const isDetailsDialogOpen = ref(false)
const isItemDetailsDialogOpen = ref(false)
const editingSetId = ref<string | null>(null)
const activeSetId = ref<string | null>(null)
const selectedItemDetailId = ref<string | null>(null)
const setNameInput = ref('')
const setCategoryInput = ref('')
const addItemId = ref('')
const addItemQuantity = ref('1')
const addItemNotes = ref('')

const setItemsBySetId = ref<Record<string, SetItemWithDetails[]>>({})
const setItemsLoadingBySetId = ref<Record<string, boolean>>({})
const setStatsById = ref<Record<string, { itemCount: number; totalWeightGrams: number }>>({})
const itemDraftsById = ref<Record<string, { quantity: string; notes: string }>>({})

const confirmDialogState = ref<
  | { kind: 'set-delete'; setId: string; setName: string }
  | { kind: 'set-bulk-delete'; setIds: string[] }
  | { kind: 'set-item-remove'; setId: string; itemId: string; itemName: string }
  | null
>(null)

const viewOptions: Array<{ label: string; value: SetsViewMode }> = [
  { label: 'Cards', value: 'cards' },
  { label: 'Table', value: 'table' },
]

const allSets = computed<ItemSet[]>(() => setsQuery.data.value ?? [])

const itemTypeOptions = computed<ItemTypeEntity[]>(() => itemTypesQuery.data.value ?? [])

const volumeInputUnit = computed<VolumeInputUnit>(() => volumeUnit.value)

const manufacturersById = computed(() => {
  const map = new Map<string, string>()
  for (const manufacturer of manufacturersQuery.data.value ?? []) {
    map.set(manufacturer.id, normalizeTitleWords(manufacturer.name))
  }

  return map
})

const itemTypeNameById = computed(() => {
  const map = new Map<string, string>()
  for (const itemType of itemTypeOptions.value) {
    map.set(itemType.id, normalizeTitleWords(itemType.name))
  }

  return map
})

const getItemTypeLabel = (categoryId: string): string => {
  return itemTypeNameById.value.get(categoryId) ?? categoryId
}

const formatType = (value: string) => {
  return formatTypeDisplay(value)
}

const formatText = (value?: string | null) => {
  return formatTextDisplay(value)
}

const formatNumber = (value?: number | null) => {
  return formatNumberDisplay(value, toRoundedString)
}

const formatCarryStatus = (value?: string | null) => {
  return formatCarryStatusDisplay(value)
}

const formatVolume = (value?: number | null) => {
  if (typeof value !== 'number') {
    return 'Not set'
  }

  if (volumeInputUnit.value === 'fl_oz') {
    return `${toRoundedString(mlToInput(value, 'fl_oz'))} fl oz`
  }

  return `${toRoundedString(value)} ml`
}

const formatValue = (value?: number | null): string => {
  return formatValueDisplay(value, currency.value, toRoundedString)
}

const getItemImageSrc = (item: Item) => {
  if (!item.image_blob) {
    return ''
  }

  return `data:${item.image_mime_type ?? 'image/*'};base64,${item.image_blob}`
}

type DetailEntry = {
  label: string
  value: string
  href?: string
  booleanValue?: boolean | null
}

const getItemDetailedEntries = (item: Item): DetailEntry[] => {
  return [
    { label: 'Type', value: formatType(item.type) },
    { label: 'Manufacturer', value: manufacturersById.value.get(item.manufacturer_id) ?? item.manufacturer_id },
    { label: 'Active', value: '', booleanValue: item.is_active },
    { label: 'Default carry', value: formatCarryStatus(item.default_carry_status) },
    { label: 'Default quantity', value: formatNumber(item.default_quantity) },
    { label: 'Is default', value: '', booleanValue: item.is_default },
    {
      label: 'Weight',
      value: typeof item.weight_grams === 'number' ? formatWeight(item.weight_grams) : 'Not set',
    },
    { label: 'Volume', value: formatVolume(item.volume_ml) },
    { label: 'Value', value: formatValue(item.value) },
    { label: 'Description', value: formatText(item.description) },
    { label: 'URL', value: item.source_url?.trim() ? 'URL' : 'Not set', href: item.source_url ?? undefined },
  ]
}

const itemTableFields = computed<AppItemTableField[]>(() => {
  return [
    {
      key: 'type',
      label: 'Type',
      render: (item: Item) => formatType(item.type),
    },
    {
      key: 'manufacturer',
      label: 'Manufacturer',
      render: (item: Item) => manufacturersById.value.get(item.manufacturer_id) ?? item.manufacturer_id,
    },
    {
      key: 'active',
      label: 'Active',
      render: () => '',
      renderBoolean: (item: Item) => item.is_active,
    },
    {
      key: 'default_carry',
      label: 'Default Carry',
      render: (item: Item) => formatCarryStatus(item.default_carry_status),
    },
    {
      key: 'default_quantity',
      label: 'Qty',
      render: (item: Item) => formatNumber(item.default_quantity),
    },
    {
      key: 'is_default',
      label: 'Default',
      render: () => '',
      renderBoolean: (item: Item) => item.is_default,
    },
    {
      key: 'weight',
      label: 'Weight',
      render: (item: Item) => {
        if (typeof item.weight_grams !== 'number') {
          return 'Not set'
        }
        return formatWeight(item.weight_grams)
      },
    },
    {
      key: 'volume',
      label: 'Volume',
      render: (item: Item) => formatVolume(item.volume_ml),
    },
    {
      key: 'value',
      label: 'Value',
      render: (item: Item) => formatValue(item.value),
    },
    {
      key: 'description',
      label: 'Notes',
      render: (item: Item) => formatText(item.description),
    },
    {
      key: 'source_url',
      label: 'URL',
      render: (item: Item) => (item.source_url?.trim() ? 'URL' : 'Not set'),
      renderHref: (item: Item) => (item.source_url?.trim() ? item.source_url : undefined),
    },
  ]
})

const activeSet = computed<ItemSet | null>(() => {
  if (!activeSetId.value) {
    return null
  }

  return allSets.value.find((set) => set.id === activeSetId.value) ?? null
})

const selectedItemDetail = computed<Item | null>(() => {
  if (!selectedItemDetailId.value) {
    return null
  }

  return (itemsQuery.data.value ?? []).find((item) => item.id === selectedItemDetailId.value) ?? null
})

const activeSetItems = computed<SetItemWithDetails[]>(() => {
  if (!activeSetId.value) {
    return []
  }

  return setItemsBySetId.value[activeSetId.value] ?? []
})

const activeSetStats = computed(() => {
  if (!activeSetId.value) {
    return { itemCount: 0, totalWeightGrams: 0 }
  }

  return setStatsById.value[activeSetId.value] ?? { itemCount: 0, totalWeightGrams: 0 }
})

const availableItemsForAdd = computed(() => {
  const activeCategory = activeSet.value?.set_category?.trim()
  const existing = new Set(activeSetItems.value.map((entry) => entry.item_id))
  return (itemsQuery.data.value ?? []).filter((item) => {
    if (existing.has(item.id)) {
      return false
    }

    if (!activeCategory) {
      return true
    }

    return item.type === activeCategory
  })
})

const setSummary = computed(() => {
  const sets = allSets.value
  const totalItems = sets.reduce((sum, entry) => sum + (setStatsById.value[entry.id]?.itemCount ?? 0), 0)
  const totalWeight = sets.reduce((sum, entry) => sum + (setStatsById.value[entry.id]?.totalWeightGrams ?? 0), 0)

  return {
    totalSets: sets.length,
    totalItems,
    totalWeightLabel: formatWeight(totalWeight),
  }
})

const confirmDialogMessage = computed(() => {
  if (confirmDialogState.value?.kind === 'set-delete') {
    return `Delete ${confirmDialogState.value.setName}?`
  }

  if (confirmDialogState.value?.kind === 'set-bulk-delete') {
    return `Delete ${confirmDialogState.value.setIds.length} selected set(s)?`
  }

  if (confirmDialogState.value?.kind === 'set-item-remove') {
    return `Remove ${confirmDialogState.value.itemName} from this set?`
  }

  return ''
})

const canShowEmptyState = computed(() => {
  return !setsQuery.isPending.value && !setsQuery.isError.value && allSets.value.length === 0
})

const isSubmittingSet = ref(false)
const isAddingItem = ref(false)
const savingItemIds = ref<Record<string, boolean>>({})

const computeSetStats = (setItems: SetItemWithDetails[]) => {
  const itemCount = setItems.length
  const totalWeightGrams = setItems.reduce((sum, entry) => {
    const itemWeight = typeof entry.item.weight_grams === 'number' ? entry.item.weight_grams : 0
    return sum + (itemWeight * entry.quantity)
  }, 0)

  return { itemCount, totalWeightGrams }
}

const loadSetItems = async (setId: string) => {
  setItemsLoadingBySetId.value = {
    ...setItemsLoadingBySetId.value,
    [setId]: true,
  }

  try {
    const data = await listSetItems(setId)
    setItemsBySetId.value = {
      ...setItemsBySetId.value,
      [setId]: data,
    }
    setStatsById.value = {
      ...setStatsById.value,
      [setId]: computeSetStats(data),
    }
  } finally {
    setItemsLoadingBySetId.value = {
      ...setItemsLoadingBySetId.value,
      [setId]: false,
    }
  }
}

const refreshAllSetStats = async () => {
  const sets = allSets.value
  if (sets.length === 0) {
    setStatsById.value = {}
    return
  }

  const next: Record<string, { itemCount: number; totalWeightGrams: number }> = {}
  await Promise.all(
    sets.map(async (set) => {
      try {
        const data = await listSetItems(set.id)
        next[set.id] = computeSetStats(data)
        setItemsBySetId.value = {
          ...setItemsBySetId.value,
          [set.id]: data,
        }
      } catch {
        next[set.id] = { itemCount: 0, totalWeightGrams: 0 }
      }
    }),
  )
  setStatsById.value = next
}

watch(
  () => setsQuery.data.value,
  () => {
    void refreshAllSetStats()
  },
  { immediate: true },
)

watch(setsViewMode, (value) => {
  if (globalThis.window === undefined) {
    return
  }

  globalThis.localStorage.setItem(SETS_VIEW_MODE_STORAGE_KEY, value)
})

watch(tableDetailMode, (value) => {
  if (globalThis.window === undefined) {
    return
  }

  globalThis.localStorage.setItem(SETS_TABLE_DETAIL_MODE_STORAGE_KEY, value)
})

watch(allSets, (sets) => {
  const validIds = new Set(sets.map((entry) => entry.id))
  selectedSetIds.value = selectedSetIds.value.filter((id) => validIds.has(id))
  if (selectedSetIds.value.length === 0) {
    tableSelectionMode.value = false
  }
})

watch(activeSetItems, (items) => {
  const drafts: Record<string, { quantity: string; notes: string }> = {}
  for (const entry of items) {
    drafts[entry.item_id] = {
      quantity: String(entry.quantity),
      notes: entry.notes ?? '',
    }
  }
  itemDraftsById.value = drafts
}, { immediate: true })

const openCreateDialog = () => {
  editingSetId.value = null
  setNameInput.value = ''
  setCategoryInput.value = ''
  isFormDialogOpen.value = true
}

const onStartEdit = (set: ItemSet) => {
  editingSetId.value = set.id
  setNameInput.value = set.name
  setCategoryInput.value = set.set_category
  isFormDialogOpen.value = true
}

const onOpenDetails = async (set: ItemSet) => {
  activeSetId.value = set.id
  isDetailsDialogOpen.value = true
  await loadSetItems(set.id)
}

const onOpenItemDetails = (item: Item) => {
  selectedItemDetailId.value = item.id
  isItemDetailsDialogOpen.value = true
}

const closeDetailsDialog = () => {
  isDetailsDialogOpen.value = false
  activeSetId.value = null
  addItemId.value = ''
  addItemQuantity.value = '1'
  addItemNotes.value = ''
}

const closeItemDetailsDialog = () => {
  isItemDetailsDialogOpen.value = false
  selectedItemDetailId.value = null
}

const consumeCreateQuery = async () => {
  if (route.query.create !== '1') {
    return
  }

  openCreateDialog()
  const nextQuery = { ...route.query }
  delete nextQuery.create
  await router.replace({
    path: route.path,
    query: nextQuery,
  })
}

watch(
  () => route.query.create,
  () => {
    void consumeCreateQuery()
  },
  { immediate: true },
)

const onSubmitSet = async () => {
  const name = normalizeTitleWords(setNameInput.value.trim())
  if (!name) {
    return
  }
  const setCategory = setCategoryInput.value.trim()
  if (!setCategory) {
    return
  }

  isSubmittingSet.value = true
  try {
    if (editingSetId.value) {
      await updateSet(editingSetId.value, { name, set_category: setCategory })
      toast.add({
        severity: 'success',
        summary: 'Set updated',
        detail: 'Set details were saved.',
        life: 2500,
      })
    } else {
      await createSet({ name, set_category: setCategory })
      toast.add({
        severity: 'success',
        summary: 'Set created',
        detail: 'New set has been created.',
        life: 2500,
      })
    }

    await queryClient.invalidateQueries({ queryKey: ['sets'] })
    isFormDialogOpen.value = false
    setNameInput.value = ''
    setCategoryInput.value = ''
    editingSetId.value = null
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: 'Save failed',
      detail: error instanceof Error ? error.message : 'Unable to save set.',
      life: 3500,
    })
  } finally {
    isSubmittingSet.value = false
  }
}

const requestDeleteSet = (set: ItemSet) => {
  confirmDialogState.value = {
    kind: 'set-delete',
    setId: set.id,
    setName: normalizeTitleWords(set.name),
  }
}

const onUpdateTableSelectionMode = (value: boolean) => {
  tableSelectionMode.value = value
  if (!value) {
    selectedSetIds.value = []
  }
}

const onToggleSetSelection = (setId: string, checked: boolean) => {
  const next = new Set(selectedSetIds.value)
  if (checked) {
    next.add(setId)
  } else {
    next.delete(setId)
  }
  selectedSetIds.value = Array.from(next)
  if (selectedSetIds.value.length === 0) {
    tableSelectionMode.value = false
  }
}

const onToggleSelectAllSets = (checked: boolean) => {
  selectedSetIds.value = checked ? allSets.value.map((set) => set.id) : []
  tableSelectionMode.value = checked
}

const onRequestBulkDelete = () => {
  if (selectedSetIds.value.length === 0) {
    return
  }

  confirmDialogState.value = {
    kind: 'set-bulk-delete',
    setIds: [...selectedSetIds.value],
  }
}

const requestRemoveSetItem = (itemId: string, itemName: string) => {
  if (!activeSetId.value) {
    return
  }

  confirmDialogState.value = {
    kind: 'set-item-remove',
    setId: activeSetId.value,
    itemId,
    itemName: normalizeTitleWords(itemName),
  }
}

const closeConfirmDialog = () => {
  confirmDialogState.value = null
}

const onConfirmDelete = async () => {
  const current = confirmDialogState.value
  if (!current) {
    return
  }

  closeConfirmDialog()
  try {
    if (current.kind === 'set-delete') {
      await removeSet(current.setId)
      await queryClient.invalidateQueries({ queryKey: ['sets'] })
      if (activeSetId.value === current.setId) {
        closeDetailsDialog()
      }
      toast.add({
        severity: 'success',
        summary: 'Set deleted',
        detail: 'Set removed successfully.',
        life: 2500,
      })
      return
    }

    if (current.kind === 'set-bulk-delete') {
      await Promise.all(current.setIds.map(async (setId) => removeSet(setId)))
      await queryClient.invalidateQueries({ queryKey: ['sets'] })
      if (activeSetId.value && current.setIds.includes(activeSetId.value)) {
        closeDetailsDialog()
      }
      selectedSetIds.value = []
      tableSelectionMode.value = false
      toast.add({
        severity: 'success',
        summary: 'Sets deleted',
        detail: `${current.setIds.length} set(s) removed successfully.`,
        life: 2500,
      })
      return
    }

    await removeSetItem(current.setId, current.itemId)
    await loadSetItems(current.setId)
    toast.add({
      severity: 'success',
      summary: 'Item removed',
      detail: 'Item removed from set.',
      life: 2500,
    })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: 'Action failed',
      detail: error instanceof Error ? error.message : 'Unable to complete action.',
      life: 3500,
    })
  }
}

const onAddItem = async () => {
  if (!activeSetId.value || !addItemId.value) {
    return
  }

  const quantity = Number(addItemQuantity.value)
  if (!Number.isFinite(quantity) || quantity <= 0) {
    return
  }

  isAddingItem.value = true
  try {
    await addSetItem(activeSetId.value, {
      item_id: addItemId.value,
      quantity,
      ...(addItemNotes.value.trim() ? { notes: addItemNotes.value.trim() } : {}),
    })

    await loadSetItems(activeSetId.value)
    addItemId.value = ''
    addItemQuantity.value = '1'
    addItemNotes.value = ''
    toast.add({
      severity: 'success',
      summary: 'Item added',
      detail: 'Item added to set.',
      life: 2200,
    })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: 'Add failed',
      detail: error instanceof Error ? error.message : 'Unable to add item.',
      life: 3500,
    })
  } finally {
    isAddingItem.value = false
  }
}

const onSaveSetItem = async (itemId: string) => {
  if (!activeSetId.value) {
    return
  }

  const draft = itemDraftsById.value[itemId]
  if (!draft) {
    return
  }

  const quantity = Number(draft.quantity)
  if (!Number.isFinite(quantity) || quantity <= 0) {
    return
  }

  savingItemIds.value = {
    ...savingItemIds.value,
    [itemId]: true,
  }

  try {
    await updateSetItem(activeSetId.value, itemId, {
      quantity,
      notes: draft.notes.trim() ? draft.notes.trim() : '',
    })
    await loadSetItems(activeSetId.value)
    toast.add({
      severity: 'success',
      summary: 'Item updated',
      detail: 'Set item updated.',
      life: 2000,
    })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: 'Update failed',
      detail: error instanceof Error ? error.message : 'Unable to update set item.',
      life: 3500,
    })
  } finally {
    savingItemIds.value = {
      ...savingItemIds.value,
      [itemId]: false,
    }
  }
}

const onUpdateItemDraft = (payload: { itemId: string; field: 'quantity' | 'notes'; value: string }) => {
  const current = itemDraftsById.value[payload.itemId] ?? { quantity: '1', notes: '' }
  itemDraftsById.value = {
    ...itemDraftsById.value,
    [payload.itemId]: {
      ...current,
      [payload.field]: payload.value,
    },
  }
}
</script>

<template>
  <section data-component="sets-page" class="flex w-full flex-col gap-4">
    <AppConfirmDialog :open="confirmDialogState !== null" title="Confirm delete" :message="confirmDialogMessage"
      confirm-label="Delete" confirm-tone="danger" @update:open="(value) => { if (!value) closeConfirmDialog() }"
      @cancel="closeConfirmDialog" @confirm="onConfirmDelete" />

    <SetFormDialog :model-value="isFormDialogOpen" :editing-set-id="editingSetId" :set-name-input="setNameInput"
      :set-category-input="setCategoryInput" :item-type-options="itemTypeOptions" :is-submitting-set="isSubmittingSet"
      @update:model-value="(value) => { isFormDialogOpen = value; if (!value) { editingSetId = null; setNameInput = ''; setCategoryInput = '' } }"
      @update:set-name-input="(value) => { setNameInput = value }"
      @update:set-category-input="(value) => { setCategoryInput = value }" @submit="onSubmitSet" />

    <ItemDetailsDialog :open="isItemDetailsDialogOpen" :selected-item="selectedItemDetail"
      :get-image-src="getItemImageSrc" :get-detailed-entries="getItemDetailedEntries" :format-type="formatType"
      :manufacturers-by-id="manufacturersById" :is-delete-loading="false" :show-edit-action="false"
      :show-delete-action="false" @update:open="(value) => { if (!value) closeItemDetailsDialog() }"
      @edit="closeItemDetailsDialog" @delete="closeItemDetailsDialog" />

    <SetDetailsDialog :model-value="isDetailsDialogOpen" :active-set="activeSet" :active-set-items="activeSetItems"
      :active-set-stats="activeSetStats"
      :set-items-loading="activeSet ? (setItemsLoadingBySetId[activeSet.id] ?? false) : false"
      :available-items-for-add="availableItemsForAdd" :add-item-id="addItemId" :add-item-quantity="addItemQuantity"
      :add-item-notes="addItemNotes" :is-adding-item="isAddingItem" :item-drafts-by-id="itemDraftsById"
      :saving-item-ids="savingItemIds" :get-item-type-label="getItemTypeLabel"
      :format-display-weight="formatWeight" @update:model-value="(value) => { if (!value) closeDetailsDialog() }"
      @update:add-item-id="(value) => { addItemId = value }"
      @update:add-item-quantity="(value) => { addItemQuantity = value }"
      @update:add-item-notes="(value) => { addItemNotes = value }" @update-item-draft="onUpdateItemDraft"
      @add-item="onAddItem" @save-set-item="onSaveSetItem"
      @request-remove-set-item="(payload) => { requestRemoveSetItem(payload.itemId, payload.itemName) }" />

    <AppQueryError :query="setsQuery" fallback-message="Unable to load sets." data-element="sets-error" />

    <AppQueryError :query="itemsQuery" fallback-message="Unable to load gear items." data-element="sets-items-error" />

    <AppQueryError :query="manufacturersQuery" fallback-message="Unable to load manufacturers."
      data-element="sets-manufacturers-error" />

    <AppLoadingState v-if="setsQuery.isPending.value" message="Loading sets..." data-element="sets-loading" />

    <AppEmptyState v-else-if="canShowEmptyState"
      message="Your organizational system is currently powered by memory and hope. Add your first set to get started!"
      data-element="sets-empty-state" />

    <div v-else class="space-y-3">
      <div class="grid gap-2 sm:grid-cols-3">
        <AppSummaryCard label="Total sets" :value="setSummary.totalSets" />
        <AppSummaryCard label="Assigned items" :value="setSummary.totalItems" />
        <AppSummaryCard label="Total weight" :value="setSummary.totalWeightLabel" />
      </div>

      <div class="flex flex-wrap items-center justify-between gap-3">
        <AppToggleGroup name="sets-view-mode" data-element="sets-view-mode" :model-value="setsViewMode"
          :options="viewOptions" fit-content
          @update:model-value="(value) => { setsViewMode = value as SetsViewMode }" />
      </div>

      <SetsCollectionView :sets="allSets" :sets-view-mode="setsViewMode" :table-detail-mode="tableDetailMode"
        :selection-mode="tableSelectionMode" :selected-set-ids="selectedSetIds" :set-stats-by-id="setStatsById"
        :set-items-by-set-id="setItemsBySetId" :item-table-fields="itemTableFields"
        :get-item-type-label="getItemTypeLabel" :format-display-weight="formatWeight" :format-date="formatDate"
        @open-details="onOpenDetails" @open-item-details="onOpenItemDetails" @start-edit="onStartEdit"
        @request-delete="requestDeleteSet" @update:table-detail-mode="(mode) => { tableDetailMode = mode }"
        @update:selection-mode="onUpdateTableSelectionMode" @toggle:set-selection="onToggleSetSelection"
        @toggle:select-all="onToggleSelectAllSets" @bulk:delete="onRequestBulkDelete" />
    </div>
  </section>
</template>
