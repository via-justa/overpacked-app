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
import type { AppItemTableField } from '../../../components/AppItemTableRowContent.vue'
import ItemDetailsDialog from '../../items/components/ItemDetailsDialog.vue'
import SetDetailsDialog from '../components/SetDetailsDialog.vue'
import SetFormDialog from '../components/SetFormDialog.vue'
import SetsCollectionView from '../components/SetsCollectionView.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { queryClient } from '../../../lib/query/client'
import { listItemTypes, listItems, listManufacturers, listItemLabels } from '../../items/api/itemsApi'
import { useSettings } from '../../../composables/useSettings'
import { toRoundedString, formatDisplayWeight } from '../../../lib/units/conversions'
import {
  formatNumber,
  formatValue,
  formatCarryStatus,
  formatType,
  formatText,
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
import type { Item, ItemTypeEntity, Label } from '../../items/types'

const toast = useToast()
const route = useRoute()
const router = useRouter()

type SetsViewMode = 'cards' | 'table'
type TableDetailMode = 'simple' | 'expanded'
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

// Wrapper for formatDisplayWeight that uses current weightInputUnit
const formatWeight = (valueGrams: number): string => {
  return formatDisplayWeight(valueGrams, weightUnit.value)
}

const formatSetValue = (value: number): string => {
  return formatValue(value, currency.value, toRoundedString)
}

const getSetLabels = (setId: string): Label[] => {
  const setItems = setItemsBySetId.value[setId] ?? []
  const labelsMap = new Map<string, Label>()

  for (const setItem of setItems) {
    const itemLabels = itemLabelsMap.value.get(setItem.item_id) ?? []
    for (const label of itemLabels) {
      labelsMap.set(label.id, label)
    }
  }

  return Array.from(labelsMap.values())
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

const itemLabelsQueries = useQuery({
  queryKey: computed(() => ['items-labels', itemsQuery.data.value?.map(i => i.id).sort().join(',') ?? '']),
  queryFn: async () => {
    const items = itemsQuery.data.value ?? []
    if (items.length === 0) return []

    const labelsPromises = items.map(item =>
      listItemLabels(item.id).catch(() => [] as Label[])
    )
    const labelsArrays = await Promise.all(labelsPromises)

    return items.map((item, index) => ({
      itemId: item.id,
      labels: labelsArrays[index],
    }))
  },
  enabled: computed(() => (itemsQuery.data.value?.length ?? 0) > 0),
})

const itemLabelsMap = computed(() => {
  const map = new Map<string, Label[]>()
  for (const entry of itemLabelsQueries.data.value ?? []) {
    map.set(entry.itemId, entry.labels)
  }
  return map
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
const tempCreateSetItems = ref<Array<{ itemId: string; quantity: number; notes: string }>>([])

const setItemsBySetId = ref<Record<string, SetItemWithDetails[]>>({})
const setItemsLoadingBySetId = ref<Record<string, boolean>>({})
const setStatsById = ref<Record<string, { itemCount: number; totalWeightGrams: number; totalValue: number }>>({})

const confirmDialogState = ref<
  | { kind: 'set-delete'; setId: string; setName: string }
  | { kind: 'set-bulk-delete'; setIds: string[] }
  | { kind: 'set-item-remove'; setId: string; itemId: string; itemName: string }
  | { kind: 'category-change'; newCategory: string; mismatchedItems: Array<{ id: string; name: string }> }
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

const formatVolume = (value?: number | null) => {
  if (typeof value !== 'number') {
    return 'Not set'
  }

  if (volumeUnit.value === 'fl_oz') {
    const flOz = value / 29.5735295625
    return `${toRoundedString(flOz)} fl oz`
  }

  return `${toRoundedString(value)} ml`
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
    { label: 'Default quantity', value: formatNumber(item.default_quantity, toRoundedString) },
    { label: 'Is default', value: '', booleanValue: item.is_default },
    {
      label: 'Weight',
      value: typeof item.weight_grams === 'number' ? formatWeight(item.weight_grams) : 'Not set',
    },
    { label: 'Volume', value: formatVolume(item.volume_ml) },
    { label: 'Value', value: formatValue(item.value, currency.value, toRoundedString) },
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
      render: (item: Item) => formatNumber(item.default_quantity, toRoundedString),
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
      render: (item: Item) => formatValue(item.value, currency.value, toRoundedString),
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
    return { itemCount: 0, totalWeightGrams: 0, totalValue: 0 }
  }

  return setStatsById.value[activeSetId.value] ?? { itemCount: 0, totalWeightGrams: 0, totalValue: 0 }
})

const availableItemsForAdd = computed(() => {
  // Use setCategoryInput when details dialog is open to get real-time updates
  const activeCategory = isDetailsDialogOpen.value
    ? setCategoryInput.value?.trim()
    : activeSet.value?.set_category?.trim()
  const existing = new Set(activeSetItems.value.map((entry) => entry.item_id))
  return filterAvailableItems(activeCategory, existing, addItemId.value)
})

const availableItemsForCreate = computed(() => {
  const activeCategory = setCategoryInput.value?.trim()
  const existing = new Set(tempCreateSetItems.value.map((item) => item.itemId))
  return filterAvailableItems(activeCategory, existing, addItemId.value)
})

// Helper function to filter items by category, active status, and excluding existing items
function filterAvailableItems(
  activeCategory: string | undefined,
  existing: Set<string>,
  editingItemId: string
): Item[] {
  return (itemsQuery.data.value ?? []).filter((item) => {
    // Allow currently editing item to be in the list
    if (item.id === editingItemId) {
      return true
    }

    if (existing.has(item.id)) {
      return false
    }

    if (!item.is_active) {
      return false
    }

    if (!activeCategory) {
      return true
    }

    return item.type === activeCategory
  })
}

const tempCreateSetItemsWithDetails = computed(() => {
  const itemsMap = new Map((itemsQuery.data.value ?? []).map((item) => [item.id, item]))

  return tempCreateSetItems.value.map((tempItem) => {
    const item = itemsMap.get(tempItem.itemId)
    return {
      tempItem,
      item,
    }
  }).filter((entry) => entry.item !== undefined)
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

  if (confirmDialogState.value?.kind === 'category-change') {
    const count = confirmDialogState.value.mismatchedItems.length
    const itemNames = confirmDialogState.value.mismatchedItems.map((item) => item.name).join(', ')
    return `Changing the category will remove ${count} item${count === 1 ? '' : 's'} that don't match: ${itemNames}. Continue?`
  }

  return ''
})

const confirmDialogTitle = computed(() => {
  if (confirmDialogState.value?.kind === 'category-change') {
    return 'Confirm category change'
  }
  return 'Confirm delete'
})

const confirmDialogLabel = computed(() => {
  if (confirmDialogState.value?.kind === 'category-change') {
    return 'Change & Remove'
  }
  return 'Delete'
})

const canShowEmptyState = computed(() => {
  return !setsQuery.isPending.value && !setsQuery.isError.value && allSets.value.length === 0
})

const isSubmittingSet = ref(false)
const isAddingItem = ref(false)

const computeSetStats = (setItems: SetItemWithDetails[]) => {
  const itemCount = setItems.length
  const totalWeightGrams = setItems.reduce((sum, entry) => {
    const itemWeight = typeof entry.item.weight_grams === 'number' ? entry.item.weight_grams : 0
    return sum + (itemWeight * entry.quantity)
  }, 0)
  const totalValue = setItems.reduce((sum, entry) => {
    const itemValue = typeof entry.item.value === 'number' ? entry.item.value : 0
    return sum + (itemValue * entry.quantity)
  }, 0)

  return { itemCount, totalWeightGrams, totalValue }
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

  const next: Record<string, { itemCount: number; totalWeightGrams: number; totalValue: number }> = {}
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
        next[set.id] = { itemCount: 0, totalWeightGrams: 0, totalValue: 0 }
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

const openCreateDialog = () => {
  editingSetId.value = null
  setNameInput.value = ''
  setCategoryInput.value = ''
  addItemId.value = ''
  addItemQuantity.value = '1'
  addItemNotes.value = ''
  tempCreateSetItems.value = []
  isFormDialogOpen.value = true
}

const openSetDetailsDialog = async (set: ItemSet) => {
  activeSetId.value = set.id
  setNameInput.value = set.name
  setCategoryInput.value = set.set_category
  isDetailsDialogOpen.value = true
  await loadSetItems(set.id)
}

const onStartEdit = openSetDetailsDialog
const onOpenDetails = openSetDetailsDialog

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
  setNameInput.value = ''
  setCategoryInput.value = ''
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
      const newSet = await createSet({ name, set_category: setCategory })

      // Add temporary items to the newly created set
      if (tempCreateSetItems.value.length > 0) {
        await Promise.all(
          tempCreateSetItems.value.map((item) =>
            addSetItem(newSet.id, {
              item_id: item.itemId,
              quantity: item.quantity,
              ...(item.notes ? { notes: item.notes } : {}),
            })
          )
        )
      }

      const itemCount = tempCreateSetItems.value.length
      const itemWord = itemCount === 1 ? 'item' : 'items'
      const itemsMessage = itemCount > 0 ? ` with ${itemCount} ${itemWord}` : ''

      toast.add({
        severity: 'success',
        summary: 'Set created',
        detail: `New set has been created${itemsMessage}.`,
        life: 2500,
      })
    }

    await queryClient.invalidateQueries({ queryKey: ['sets'] })
    isFormDialogOpen.value = false
    setNameInput.value = ''
    setCategoryInput.value = ''
    editingSetId.value = null
    tempCreateSetItems.value = []
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

const onSaveSetFromDetails = async () => {
  const name = normalizeTitleWords(setNameInput.value.trim())
  if (!name) {
    return
  }
  const setCategory = setCategoryInput.value.trim()
  if (!setCategory) {
    return
  }

  if (!activeSetId.value) {
    return
  }

  isSubmittingSet.value = true
  try {
    await updateSet(activeSetId.value, { name, set_category: setCategory })
    await queryClient.invalidateQueries({ queryKey: ['sets'] })
    toast.add({
      severity: 'success',
      summary: 'Set updated',
      detail: 'Set details were saved.',
      life: 2500,
    })
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

const onRequestCategoryChange = (newCategory: string) => {
  if (!activeSetId.value || !activeSet.value) {
    return
  }

  // Check for mismatched items
  const mismatchedItems = activeSetItems.value
    .filter((entry) => entry.item.type !== newCategory)
    .map((entry) => ({ id: entry.item_id, name: normalizeTitleWords(entry.item.name) }))

  if (mismatchedItems.length > 0) {
    // Show confirmation dialog
    confirmDialogState.value = {
      kind: 'category-change',
      newCategory,
      mismatchedItems,
    }
  } else {
    // No conflicts, proceed with save
    void onSaveSetFromDetails()
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

const populateItemEditForm = (payload: { itemId: string; quantity: number; notes: string }) => {
  addItemId.value = payload.itemId
  addItemQuantity.value = String(payload.quantity)
  addItemNotes.value = payload.notes
}

const onEditSetItem = populateItemEditForm

const closeConfirmDialog = () => {
  confirmDialogState.value = null
}

const handleCategoryChangeConfirmation = async (mismatchedItems: Array<{ id: string; name: string }>) => {
  if (!activeSetId.value) {
    return
  }

  // Remove mismatched items
  await Promise.all(
    mismatchedItems.map((item) => removeSetItem(activeSetId.value!, item.id)),
  )

  // Reload set items
  await loadSetItems(activeSetId.value)

  // Now save the category change
  await onSaveSetFromDetails()

  toast.add({
    severity: 'success',
    summary: 'Category changed',
    detail: `Removed ${mismatchedItems.length} item${mismatchedItems.length === 1 ? '' : 's'} and updated category.`,
    life: 2500,
  })
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

    if (current.kind === 'category-change') {
      await handleCategoryChangeConfirmation(current.mismatchedItems)
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

const onCancelCategoryChange = () => {
  // Revert category to original value if this is a category change
  if (confirmDialogState.value?.kind === 'category-change' && activeSet.value) {
    setCategoryInput.value = activeSet.value.set_category
  }
  closeConfirmDialog()
}

const onAddTempItem = () => {
  if (!addItemId.value) {
    return
  }

  const quantity = Number(addItemQuantity.value)
  if (!Number.isFinite(quantity) || quantity <= 0) {
    return
  }

  // Check if item already exists in temp list
  const existingIndex = tempCreateSetItems.value.findIndex((item) => item.itemId === addItemId.value)

  if (existingIndex >= 0) {
    // Update existing item
    tempCreateSetItems.value[existingIndex] = {
      itemId: addItemId.value,
      quantity,
      notes: addItemNotes.value.trim(),
    }
  } else {
    // Add new item
    tempCreateSetItems.value.push({
      itemId: addItemId.value,
      quantity,
      notes: addItemNotes.value.trim(),
    })
  }

  addItemId.value = ''
  addItemQuantity.value = '1'
  addItemNotes.value = ''
}

const onRemoveTempItem = (itemId: string) => {
  tempCreateSetItems.value = tempCreateSetItems.value.filter((item) => item.itemId !== itemId)
}

const onEditTempItem = populateItemEditForm

const onAddItem = async () => {
  if (!activeSetId.value || !addItemId.value) {
    return
  }

  const quantity = Number(addItemQuantity.value)
  if (!Number.isFinite(quantity) || quantity <= 0) {
    return
  }

  // Check if item already exists in the set
  const existingItem = activeSetItems.value.find((entry) => entry.item_id === addItemId.value)
  const isUpdate = !!existingItem

  isAddingItem.value = true
  try {
    if (isUpdate) {
      // Update existing item
      await updateSetItem(activeSetId.value, addItemId.value, {
        quantity,
        notes: addItemNotes.value.trim() || '',
      })
      toast.add({
        severity: 'success',
        summary: 'Item updated',
        detail: 'Item quantity and notes updated.',
        life: 2200,
      })
    } else {
      // Add new item
      await addSetItem(activeSetId.value, {
        item_id: addItemId.value,
        quantity,
        ...(addItemNotes.value.trim() ? { notes: addItemNotes.value.trim() } : {}),
      })
      toast.add({
        severity: 'success',
        summary: 'Item added',
        detail: 'Item added to set.',
        life: 2200,
      })
    }

    await loadSetItems(activeSetId.value)
    addItemId.value = ''
    addItemQuantity.value = '1'
    addItemNotes.value = ''
  } catch (error) {
    const action = isUpdate ? 'update' : 'add'
    const errorMessage = error instanceof Error ? error.message : `Unable to ${action} item.`
    toast.add({
      severity: 'error',
      summary: isUpdate ? 'Update failed' : 'Add failed',
      detail: errorMessage,
      life: 3500,
    })
  } finally {
    isAddingItem.value = false
  }
}
</script>

<template>
  <section data-component="sets-page" class="flex w-full flex-col gap-4">
    <AppConfirmDialog :open="confirmDialogState !== null" :title="confirmDialogTitle" :message="confirmDialogMessage"
      :confirm-label="confirmDialogLabel" confirm-tone="danger"
      @update:open="(value) => { if (!value) onCancelCategoryChange() }" @cancel="onCancelCategoryChange"
      @confirm="onConfirmDelete" />

    <SetFormDialog :model-value="isFormDialogOpen" :editing-set-id="editingSetId" :set-name-input="setNameInput"
      :set-category-input="setCategoryInput" :item-type-options="itemTypeOptions" :is-submitting-set="isSubmittingSet"
      :available-items-for-add="availableItemsForCreate" :add-item-id="addItemId" :add-item-quantity="addItemQuantity"
      :add-item-notes="addItemNotes" :temp-items="tempCreateSetItemsWithDetails"
      :manufacturers-by-id="manufacturersById" :format-display-weight="formatWeight" :format-value="formatSetValue"
      :volume-input-unit="volumeInputUnit"
      @update:model-value="(value) => { isFormDialogOpen = value; if (!value) { editingSetId = null; setNameInput = ''; setCategoryInput = ''; tempCreateSetItems = [] } }"
      @update:set-name-input="(value) => { setNameInput = value }"
      @update:set-category-input="(value) => { setCategoryInput = value }"
      @update:add-item-id="(value) => { addItemId = value }"
      @update:add-item-quantity="(value) => { addItemQuantity = value }"
      @update:add-item-notes="(value) => { addItemNotes = value }" @add-item="onAddTempItem"
      @remove-item="onRemoveTempItem" @edit-item="onEditTempItem" @submit="onSubmitSet" />

    <ItemDetailsDialog :open="isItemDetailsDialogOpen" :selected-item="selectedItemDetail"
      :get-image-src="getItemImageSrc" :get-detailed-entries="getItemDetailedEntries" :format-type="formatType"
      :manufacturers-by-id="manufacturersById" :is-delete-loading="false" :show-edit-action="false"
      :show-delete-action="false" @update:open="(value) => { if (!value) closeItemDetailsDialog() }"
      @edit="closeItemDetailsDialog" @delete="closeItemDetailsDialog" />

    <SetDetailsDialog :model-value="isDetailsDialogOpen" :active-set="activeSet" :active-set-items="activeSetItems"
      :active-set-stats="activeSetStats"
      :set-items-loading="activeSet ? (setItemsLoadingBySetId[activeSet.id] ?? false) : false"
      :available-items-for-add="availableItemsForAdd" :add-item-id="addItemId" :add-item-quantity="addItemQuantity"
      :add-item-notes="addItemNotes" :is-adding-item="isAddingItem" :get-item-type-label="getItemTypeLabel"
      :format-display-weight="formatWeight" :format-value="formatSetValue" :set-name-input="setNameInput"
      :set-category-input="setCategoryInput" :item-type-options="itemTypeOptions" :is-submitting-set="isSubmittingSet"
      :manufacturers-by-id="manufacturersById" :volume-input-unit="volumeInputUnit"
      @update:model-value="(value) => { if (!value) closeDetailsDialog() }"
      @update:add-item-id="(value) => { addItemId = value }"
      @update:add-item-quantity="(value) => { addItemQuantity = value }"
      @update:add-item-notes="(value) => { addItemNotes = value }"
      @update:set-name-input="(value) => { setNameInput = value }"
      @update:set-category-input="(value) => { setCategoryInput = value }" @add-item="onAddItem"
      @save-set="onSaveSetFromDetails" @request-category-change="onRequestCategoryChange" @edit-set-item="onEditSetItem"
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
      <div class="flex flex-wrap items-center justify-between gap-3">
        <AppToggleGroup name="sets-view-mode" data-element="sets-view-mode" :model-value="setsViewMode"
          :options="viewOptions" fit-content
          @update:model-value="(value) => { setsViewMode = value as SetsViewMode }" />
      </div>

      <SetsCollectionView :sets="allSets" :sets-view-mode="setsViewMode" :table-detail-mode="tableDetailMode"
        :selection-mode="tableSelectionMode" :selected-set-ids="selectedSetIds" :set-stats-by-id="setStatsById"
        :set-items-by-set-id="setItemsBySetId" :item-table-fields="itemTableFields"
        :get-item-type-label="getItemTypeLabel" :format-display-weight="formatWeight" :format-value="formatSetValue"
        :format-date="formatDate" :get-set-labels="getSetLabels" @open-details="onOpenDetails"
        @open-item-details="onOpenItemDetails" @start-edit="onStartEdit" @request-delete="requestDeleteSet"
        @update:table-detail-mode="(mode) => { tableDetailMode = mode }"
        @update:selection-mode="onUpdateTableSelectionMode" @toggle:set-selection="onToggleSetSelection"
        @toggle:select-all="onToggleSelectAllSets" @bulk:delete="onRequestBulkDelete" />
    </div>
  </section>
</template>
