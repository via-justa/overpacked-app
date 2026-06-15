<script setup lang="ts">
import { computed, useAttrs, useSlots } from 'vue'
import Select from 'primevue/select'
import { selectPassThrough, useFieldMessageClass, useSelectOptions } from '../../composables/useSelectOptions'

defineOptions({
  inheritAttrs: false,
})

const props = withDefaults(defineProps<{
  modelValue?: string | number | null
  compact?: boolean
  invalid?: boolean
  message?: string
  reserveMessageSpace?: boolean
  selectClass?: string
  wrapperClass?: string
  messageClass?: string
}>(), {
  modelValue: '',
  compact: false,
  invalid: false,
  message: '',
  reserveMessageSpace: false,
  selectClass: '',
  wrapperClass: '',
  messageClass: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  change: [value: string, event: Event]
}>()

const attrs = useAttrs()
const slots = useSlots()

const { parsedOptions } = useSelectOptions(slots)
const resolvedMessageClass = useFieldMessageClass(props)

// Give the underlying <select> an accessible name. Consumers can override via an
// explicit aria-label or placeholder; otherwise fall back to a generic label.
const resolvedAriaLabel = computed(
  () => (attrs['aria-label'] as string) || (attrs.placeholder as string) || 'Select option',
)

const resolvedSelectClass = computed(() => {
  const classes = ['app-select-field']

  if (props.invalid) {
    classes.push('is-invalid')
  }

  if (props.compact) {
    classes.push('is-compact')
  }

  if (props.selectClass) {
    classes.push(props.selectClass)
  }

  return classes.join(' ')
})

const onPrimeChange = (event: { value: string; originalEvent: Event }) => {
  emit('update:modelValue', event.value)
  emit('change', event.value, event.originalEvent)
}
</script>

<template>
  <div :class="wrapperClass">
    <Select v-bind="attrs" :aria-label="resolvedAriaLabel" :model-value="modelValue ?? ''" :options="parsedOptions" option-label="label"
      option-value="value" option-disabled="disabled" :pt="selectPassThrough" :class="resolvedSelectClass" fluid
      @change="onPrimeChange" />
    <span :class="resolvedMessageClass">{{ message || ' ' }}</span>
  </div>
</template>
