<script setup lang="ts">
import AppToggleGroup from '../../../components/AppToggleGroup.vue'

type Option = {
  label: string
  value: string
}

const props = defineProps<{
  id: string
  label: string
  helper: string
  modelValue: string
  options: Option[]
  disabled?: boolean
  dirty?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <div :data-component="`unit-setting-field-${id}`"
    class="border-line-subtle bg-surface-base flex flex-wrap items-center justify-between gap-2 rounded-xl border px-3 py-2">
    <div class="min-w-0">
      <p class="text-ink text-sm font-medium">{{ label }}</p>
      <p v-if="helper" class="text-copy-muted text-xs">{{ helper }}</p>
    </div>

    <AppToggleGroup :id="id" :name="id" :data-element="`unit-toggle-${id}`" :model-value="modelValue"
      :options="props.options" :disabled="disabled" :dirty="props.dirty"
      @update:model-value="(value) => emit('update:modelValue', value)" />
  </div>
</template>
