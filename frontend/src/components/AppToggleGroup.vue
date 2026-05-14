<script setup lang="ts">
type ToggleOption = {
  label: string
  value: string
}

const props = defineProps<{
  name: string
  modelValue: string
  options: ToggleOption[]
  id?: string
  dataElement?: string
  disabled?: boolean
  dirty?: boolean
  fitContent?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const onSelect = (value: string) => {
  emit('update:modelValue', value)
}
</script>

<template>
  <fieldset :id="id" :data-element="dataElement" class="toggle-group" :class="[
    props.dirty ? 'is-dirty' : '',
    props.fitContent ? 'w-fit' : '',
  ]" :disabled="disabled">
    <label v-for="option in props.options" :key="option.value" class="toggle-option">
      <input class="toggle-option-input sr-only" type="radio" :name="name" :value="option.value"
        :checked="modelValue === option.value" :disabled="disabled" @change="onSelect(option.value)" />
      <span class="toggle-option-label">
        {{ option.label }}
      </span>
    </label>
  </fieldset>
</template>