<script setup lang="ts">
import { computed } from 'vue'
import { actionOptions, type ActionOption, type ActionTarget } from '../actions/actionOptions'
import { useActionRail } from '../../composables/useActionRail'
import { useActionRouter } from '../actions/useActionRouter'

const emit = defineEmits<{
  logout: []
}>()

const { pinned } = useActionRail()
const { selectAction } = useActionRouter(() => emit('logout'))

const optionByValue = (value: ActionTarget): ActionOption => {
  const option = actionOptions.find((item) => item.value === value)
  if (!option) {
    throw new Error(`Unknown action option: ${value}`)
  }
  return option
}

// Group the shared action list into the rail's visual sections.
const createGroup = (['add-trip', 'add-item', 'add-set', 'add-person', 'add-packing-list'] as ActionTarget[]).map(optionByValue)
const manageGroup = (['manage-manufacturers', 'manage-categories', 'import-csv'] as ActionTarget[]).map(optionByValue)
const footerGroup = (['settings', 'logout'] as ActionTarget[]).map(optionByValue)

// Labels and group headers are hidden when collapsed and fade in once the rail
// expands (pinned, hovered, or focused into).
const labelVisibilityClass = computed(() =>
  pinned.value
    ? 'opacity-100'
    : 'opacity-0 group-hover:opacity-100 group-focus-within:opacity-100',
)
</script>

<template>
  <nav data-component="app-action-rail" aria-label="Actions"
    class="group border-line-subtle bg-surface-elevated fixed bottom-0 left-0 top-16 z-30 hidden flex-col gap-1 overflow-hidden border-r py-3 backdrop-blur transition-[width,box-shadow] duration-200 motion-reduce:transition-none md:flex"
    :class="pinned
      ? 'w-60 shadow-panel'
      : 'w-16 hover:w-60 focus-within:w-60 hover:shadow-panel focus-within:shadow-panel'">
    <p data-element="rail-group-header"
      class="text-copy-subtle whitespace-nowrap px-5 pb-1 pt-1 text-xs font-semibold uppercase tracking-[0.08em] transition-opacity duration-200 motion-reduce:transition-none"
      :class="labelVisibilityClass">
      Create
    </p>
    <button v-for="option in createGroup" :key="option.value" type="button" :data-action-option="option.value"
      class="text-copy hover:bg-surface-soft hover:text-ink focus-visible:bg-surface-soft focus-visible:text-ink focus-visible:ring-brand-500 mx-2 flex h-10 items-center gap-3 rounded-lg px-3 text-left outline-none ring-2 ring-transparent transition-colors"
      @click="selectAction(option.value)">
      <span class="flex w-6 shrink-0 items-center justify-center" aria-hidden="true">
        <i :class="option.icon" class="text-lg" />
      </span>
      <span class="whitespace-nowrap text-sm font-medium transition-opacity duration-200 motion-reduce:transition-none"
        :class="labelVisibilityClass">{{ option.label }}</span>
    </button>

    <div class="border-line-subtle mx-3 my-1 border-t"></div>
    <p data-element="rail-group-header"
      class="text-copy-subtle whitespace-nowrap px-5 pb-1 pt-1 text-xs font-semibold uppercase tracking-[0.08em] transition-opacity duration-200 motion-reduce:transition-none"
      :class="labelVisibilityClass">
      Manage
    </p>
    <button v-for="option in manageGroup" :key="option.value" type="button" :data-action-option="option.value"
      class="text-copy hover:bg-surface-soft hover:text-ink focus-visible:bg-surface-soft focus-visible:text-ink focus-visible:ring-brand-500 mx-2 flex h-10 items-center gap-3 rounded-lg px-3 text-left outline-none ring-2 ring-transparent transition-colors"
      @click="selectAction(option.value)">
      <span class="flex w-6 shrink-0 items-center justify-center" aria-hidden="true">
        <i :class="option.icon" class="text-lg" />
      </span>
      <span class="whitespace-nowrap text-sm font-medium transition-opacity duration-200 motion-reduce:transition-none"
        :class="labelVisibilityClass">{{ option.label }}</span>
    </button>

    <div class="flex-1"></div>
    <div class="border-line-subtle mx-3 my-1 border-t"></div>
    <button v-for="option in footerGroup" :key="option.value" type="button" :data-action-option="option.value"
      class="text-copy hover:bg-surface-soft hover:text-ink focus-visible:bg-surface-soft focus-visible:text-ink focus-visible:ring-brand-500 mx-2 flex h-10 items-center gap-3 rounded-lg px-3 text-left outline-none ring-2 ring-transparent transition-colors"
      @click="selectAction(option.value)">
      <span class="flex w-6 shrink-0 items-center justify-center" aria-hidden="true">
        <i :class="option.icon" class="text-lg" />
      </span>
      <span class="whitespace-nowrap text-sm font-medium transition-opacity duration-200 motion-reduce:transition-none"
        :class="labelVisibilityClass">{{ option.label }}</span>
    </button>
  </nav>
</template>
