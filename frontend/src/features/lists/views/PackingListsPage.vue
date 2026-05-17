<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMutation, useQuery } from '@tanstack/vue-query'
import { useToast } from 'primevue/usetoast'
import AppConfirmDialog from '../../../components/dialogs/AppConfirmDialog.vue'
import AppQueryState from '../../../components/feedback/AppQueryState.vue'
import { queryClient } from '../../../lib/query/client'
import { listLabels } from '../../items/api/itemsApi'
import {
  listPackingLists,
  listPackingListLabels,
  createPackingList,
  updatePackingList,
  removePackingList,
  addPackingListLabel,
  removePackingListLabel,
} from '../api/listsApi'
import PackingListCard from '../components/PackingListCard.vue'
import PackingListFormDialog from '../components/PackingListFormDialog.vue'
import type { PackingList, Label } from '../types'

const toast = useToast()
const route = useRoute()
const router = useRouter()

// ─── state ────────────────────────────────────────────────────────────────────

const editingListId = ref<string | null>(null)
const isFormDialogOpen = ref(false)
const confirmDialogState = ref<{ listId: string; listName: string } | null>(null)
const listLabelsById = ref<Record<string, Label[]>>({})

// ─── queries ──────────────────────────────────────────────────────────────────

const packingListsQuery = useQuery({
  queryKey: ['packing-lists'],
  queryFn: listPackingLists,
})

const allLabelsQuery = useQuery({
  queryKey: ['labels'],
  queryFn: listLabels,
})

// Fetch labels for all packing lists when they load
watch(
  () => packingListsQuery.data.value,
  async (lists) => {
    if (!lists) return

    for (const list of lists) {
      try {
        const labels = await listPackingListLabels(list.id)
        listLabelsById.value[list.id] = labels
      } catch (error) {
        console.error(`Failed to load labels for list ${list.id}:`, error)
        listLabelsById.value[list.id] = []
      }
    }
  },
  { immediate: true },
)


// ─── computed ─────────────────────────────────────────────────────────────────

const sortedPackingLists = computed(() => {
  const lists = packingListsQuery.data.value ?? []
  return [...lists].sort((a, b) => a.name.localeCompare(b.name))
})

const isCreateMode = computed(() => editingListId.value === null)

const editingList = computed(() => {
  if (!editingListId.value) return null
  return packingListsQuery.data.value?.find((list) => list.id === editingListId.value) ?? null
})

const selectedLabels = ref<Label[]>([])

const availableLabels = computed(() => {
  const all = allLabelsQuery.data.value ?? []
  const selected = selectedLabels.value
  return all.filter((label) => !selected.some((s) => s.id === label.id))
})

// ─── mutations ────────────────────────────────────────────────────────────────

const createMutation = useMutation({
  mutationFn: createPackingList,
  onSuccess: async (createdList) => {
    // Add selected labels to the newly created list
    if (selectedLabels.value.length > 0) {
      try {
        const addedLabels: Label[] = []
        for (const label of selectedLabels.value) {
          const addedLabel = await addPackingListLabel(createdList.id, { label_id: label.id })
          addedLabels.push(addedLabel)
        }
        // Update the labels cache for this new list
        listLabelsById.value[createdList.id] = addedLabels
      } catch (error) {
        console.error('Failed to add labels:', error)
        toast.add({
          severity: 'warn',
          summary: 'Labels not added',
          detail: 'Packing list created but some labels could not be added.',
          life: 3500,
        })
      }
    } else {
      // No labels selected, set empty array
      listLabelsById.value[createdList.id] = []
    }

    selectedLabels.value = []
    isFormDialogOpen.value = false
    await queryClient.invalidateQueries({ queryKey: ['packing-lists'] })
    toast.add({
      severity: 'success',
      summary: 'Packing list created',
      detail: 'New packing list has been added.',
      life: 3000,
    })
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Create failed',
      detail: error instanceof Error ? error.message : 'Unable to create packing list.',
      life: 3500,
    })
  },
})

const updateMutation = useMutation({
  mutationFn: async (params: { listId: string; name: string; description: string }) => {
    return updatePackingList(params.listId, {
      name: params.name,
      description: params.description || undefined,
    })
  },
  onSuccess: async () => {
    isFormDialogOpen.value = false
    editingListId.value = null
    await queryClient.invalidateQueries({ queryKey: ['packing-lists'] })
    toast.add({
      severity: 'success',
      summary: 'Packing list updated',
      detail: 'Changes have been saved.',
      life: 3000,
    })
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Update failed',
      detail: error instanceof Error ? error.message : 'Unable to update packing list.',
      life: 3500,
    })
  },
})

const deleteMutation = useMutation({
  mutationFn: removePackingList,
  onSuccess: async (_result, listId) => {
    // Clean up labels cache for deleted list
    delete listLabelsById.value[listId]

    confirmDialogState.value = null
    isFormDialogOpen.value = false
    editingListId.value = null
    await queryClient.invalidateQueries({ queryKey: ['packing-lists'] })
    toast.add({
      severity: 'success',
      summary: 'Packing list deleted',
      detail: 'Packing list has been removed.',
      life: 3000,
    })
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Delete failed',
      detail: error instanceof Error ? error.message : 'Unable to delete packing list.',
      life: 3500,
    })
  },
})

const addLabelMutation = useMutation({
  mutationFn: async (params: { listId: string; labelId: string }) => {
    return addPackingListLabel(params.listId, { label_id: params.labelId })
  },
  onSuccess: async (label, variables) => {
    selectedLabels.value = [...selectedLabels.value, label]

    // Update the labels cache for this list
    if (listLabelsById.value[variables.listId]) {
      listLabelsById.value[variables.listId] = [...listLabelsById.value[variables.listId], label]
    }
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Failed to add label',
      detail: error instanceof Error ? error.message : 'Unable to add label.',
      life: 3500,
    })
  },
})

const removeLabelMutation = useMutation({
  mutationFn: async (params: { listId: string; labelId: string }) => {
    return removePackingListLabel(params.listId, params.labelId)
  },
  onSuccess: async (_result, variables) => {
    selectedLabels.value = selectedLabels.value.filter((label) => label.id !== variables.labelId)

    // Update the labels cache for this list
    if (listLabelsById.value[variables.listId]) {
      listLabelsById.value[variables.listId] = listLabelsById.value[variables.listId].filter(
        (label) => label.id !== variables.labelId
      )
    }
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Failed to remove label',
      detail: error instanceof Error ? error.message : 'Unable to remove label.',
      life: 3500,
    })
  },
})

// ─── handlers ─────────────────────────────────────────────────────────────────

const openCreateDialog = () => {
  editingListId.value = null
  selectedLabels.value = []
  isFormDialogOpen.value = true
}

const openEditDialog = async (list: PackingList) => {
  editingListId.value = list.id

  // Fetch labels for this list
  try {
    const labels = await listPackingListLabels(list.id)
    selectedLabels.value = labels
  } catch (error) {
    console.error('Failed to load labels:', error)
    selectedLabels.value = []
  }

  isFormDialogOpen.value = true
}

const handleSubmit = (name: string, description: string) => {
  if (isCreateMode.value) {
    createMutation.mutate({ name, description: description || undefined })
  }
  else if (editingListId.value) {
    updateMutation.mutate({ listId: editingListId.value, name, description })
  }
}

const handleDelete = () => {
  if (!editingList.value) return

  confirmDialogState.value = {
    listId: editingList.value.id,
    listName: editingList.value.name,
  }
}

const handleConfirmDelete = () => {
  if (!confirmDialogState.value) return
  deleteMutation.mutate(confirmDialogState.value.listId)
}

const closeConfirmDialog = () => {
  confirmDialogState.value = null
}

const handleAddLabel = (labelId: string) => {
  // In create mode: add to local array
  if (isCreateMode.value) {
    const label = allLabelsQuery.data.value?.find((l) => l.id === labelId)
    if (label && !selectedLabels.value.some((l) => l.id === labelId)) {
      selectedLabels.value = [...selectedLabels.value, label]
    }
    return
  }

  // In edit mode: make API call
  if (editingListId.value) {
    addLabelMutation.mutate({ listId: editingListId.value, labelId })
  }
}

const handleRemoveLabel = (labelId: string) => {
  // In create mode: remove from local array
  if (isCreateMode.value) {
    selectedLabels.value = selectedLabels.value.filter((label) => label.id !== labelId)
    return
  }

  // In edit mode: make API call
  if (editingListId.value) {
    removeLabelMutation.mutate({ listId: editingListId.value, labelId })
  }
}

const formLoading = computed(() => {
  return isCreateMode.value ? createMutation.isPending.value : updateMutation.isPending.value
})

const confirmDialogMessage = computed(() => {
  if (!confirmDialogState.value) return ''
  return `Are you sure you want to delete "${confirmDialogState.value.listName}"? This action cannot be undone.`
})

// Handle ?create=1 query parameter
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
</script>

<template>
  <section data-component="packing-lists-page" class="flex w-full flex-col gap-4">
    <AppConfirmDialog :open="confirmDialogState !== null" title="Confirm delete" :message="confirmDialogMessage"
      confirm-label="Delete" confirm-tone="danger" @update:open="(value) => { if (!value) closeConfirmDialog() }"
      @cancel="closeConfirmDialog" @confirm="handleConfirmDelete" />

    <PackingListFormDialog :open="isFormDialogOpen" :is-create-mode="isCreateMode" :packing-list="editingList"
      :selected-labels="selectedLabels" :available-labels="availableLabels" :is-loading="formLoading"
      :labels-loading="allLabelsQuery.isPending.value" @update:open="isFormDialogOpen = $event" @submit="handleSubmit"
      @delete="handleDelete" @label:add="handleAddLabel" @label:remove="handleRemoveLabel" />

    <AppQueryState :query="packingListsQuery" loading-message="Loading packing lists..."
      empty-message="Your organizational system is currently powered by memory and hope. Create packing lists for different adventures!"
      error-fallback="Unable to load packing lists.">
      <div data-element="packing-lists-grid" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <PackingListCard v-for="list in sortedPackingLists" :key="list.id" :packing-list="list"
          :labels="listLabelsById[list.id] || []" @open-edit="openEditDialog" />
      </div>
    </AppQueryState>
  </section>
</template>
