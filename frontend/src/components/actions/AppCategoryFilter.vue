<script setup lang="ts">
/**
 * Reusable filter/category navigation pattern.
 * Renders a horizontal list of filter buttons with separator slashes.
 */
interface FilterOption {
  value: string
  label: string
}

defineProps<{
  modelValue: string
  options: FilterOption[]
  label?: string
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <nav :aria-label="label ?? 'Filter'"
    class="text-copy-subtle flex flex-wrap items-center text-xs font-semibold uppercase tracking-[0.08em]">
    <template v-for="(option, index) in options" :key="option.value">
      <span v-if="index > 0" class="text-line mx-0.5">/</span>
      <button type="button" class="rounded px-2 py-1 transition"
        :class="modelValue === option.value ? 'bg-brand-50 text-brand-800' : 'text-copy-subtle hover:bg-surface-soft hover:text-copy'"
        @click="$emit('update:modelValue', option.value)">
        {{ option.label }}
      </button>
    </template>
  </nav>
</template>
