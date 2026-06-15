<script setup lang="ts">
import { computed, ref } from 'vue'
import { AppIcon } from '../../../components/icons'
import ItemLabel from './ItemLabel.vue'
import type { Label } from '../types'

const props = defineProps<{
  selectedLabels: Label[]
  availableLabels: Label[]
  loading?: boolean
}>()

const emit = defineEmits<{
  'add': [label: Label]
  'remove': [labelId: string]
  'create': [name: string]
}>()

const searchInput = ref('')
const showDropdown = ref(false)

const filteredLabels = computed(() => {
  if (!searchInput.value.trim()) {
    return []
  }

  const search = searchInput.value.toLowerCase()
  const selectedIds = new Set(props.selectedLabels.map(l => l.id))

  return props.availableLabels
    .filter(label => !selectedIds.has(label.id) && label.name.toLowerCase().includes(search))
    .slice(0, 5)
})

const showCreateOption = computed(() => {
  if (!searchInput.value.trim()) {
    return false
  }

  const exactMatch = props.availableLabels.some(
    label => label.name.toLowerCase() === searchInput.value.toLowerCase()
  )

  return !exactMatch
})

const onInputFocus = () => {
  showDropdown.value = true
}

const onInputBlur = () => {
  setTimeout(() => {
    showDropdown.value = false
  }, 200)
}

const onInputKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Enter') {
    event.preventDefault()

    if (filteredLabels.value.length === 1) {
      selectLabel(filteredLabels.value[0])
      return
    }

    if (showCreateOption.value && searchInput.value.trim()) {
      createLabel()
    }
  }
}

const selectLabel = (label: Label) => {
  emit('add', label)
  searchInput.value = ''
  showDropdown.value = false
}

const createLabel = () => {
  if (!searchInput.value.trim()) {
    return
  }

  emit('create', searchInput.value.trim())
  searchInput.value = ''
  showDropdown.value = false
}

const removeLabel = (labelId: string) => {
  emit('remove', labelId)
}
</script>

<template>
  <div data-component="item-labels-selector" class="grid gap-2">
    <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Labels</span>

    <div v-if="selectedLabels.length > 0" class="flex flex-wrap gap-1.5">
      <ItemLabel v-for="label in selectedLabels" :key="label.id" :label="label" size="sm" :removable="true"
        @remove="removeLabel(label.id)" />
    </div>

    <div class="relative">
      <input v-model="searchInput" data-element="item-labels-search" aria-label="Search or create label" class="input-shell w-full" type="text"
        placeholder="Search or create label..." :disabled="loading" @focus="onInputFocus" @blur="onInputBlur"
        @keydown="onInputKeydown" />

      <div v-if="showDropdown && (filteredLabels.length > 0 || showCreateOption)"
        class="border-line-subtle bg-surface-elevated absolute top-full left-0 right-0 z-10 mt-1 max-h-48 overflow-y-auto rounded-lg border shadow-sm">
        <button v-for="label in filteredLabels" :key="label.id" type="button"
          class="text-copy-subtle hover:bg-surface-soft flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition"
          @click="selectLabel(label)">
          <ItemLabel :label="label" size="sm" />
        </button>

        <button v-if="showCreateOption" type="button"
          class="text-copy-subtle hover:bg-surface-soft flex w-full items-center gap-2 border-t border-line-subtle px-3 py-2 text-left text-sm font-medium transition"
          @click="createLabel">
          <AppIcon category="action" name="create" size="xs" />
          <span>Create "{{ searchInput }}"</span>
        </button>
      </div>
    </div>
  </div>
</template>
