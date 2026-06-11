<script setup lang="ts">
import { computed, useAttrs, useSlots } from 'vue'
import MultiSelect from 'primevue/multiselect'
import { selectPassThrough, useFieldMessageClass, useSelectOptions } from '../../composables/useSelectOptions'

defineOptions({
  inheritAttrs: false,
})

const props = withDefaults(defineProps<{
  modelValue?: Array<string | number>
  placeholder?: string
  invalid?: boolean
  message?: string
  reserveMessageSpace?: boolean
  filter?: boolean
  showToggleAll?: boolean
  maxSelectedLabels?: number
  selectClass?: string
  wrapperClass?: string
  messageClass?: string
}>(), {
  modelValue: () => [],
  placeholder: '',
  invalid: false,
  message: '',
  reserveMessageSpace: false,
  filter: false,
  showToggleAll: false,
  maxSelectedLabels: 2,
  selectClass: '',
  wrapperClass: '',
  messageClass: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: Array<string | number>]
  change: [value: Array<string | number>, event: Event]
}>()

const attrs = useAttrs()
const slots = useSlots()

const { parsedOptions } = useSelectOptions(slots)
const resolvedMessageClass = useFieldMessageClass(props)

const resolvedSelectClass = computed(() => {
  const classes = ['app-select-field']

  if (props.invalid) {
    classes.push('is-invalid')
  }

  if (props.selectClass) {
    classes.push(props.selectClass)
  }

  return classes.join(' ')
})

const onPrimeChange = (event: { value: Array<string | number>; originalEvent: Event }) => {
  emit('update:modelValue', event.value)
  emit('change', event.value, event.originalEvent)
}
</script>

<template>
  <div :class="wrapperClass">
    <MultiSelect v-bind="attrs" :model-value="modelValue" :options="parsedOptions" option-label="label"
      option-value="value" option-disabled="disabled" :placeholder="placeholder" :filter="filter"
      :show-toggle-all="showToggleAll" :max-selected-labels="maxSelectedLabels" :pt="selectPassThrough"
      :class="resolvedSelectClass" fluid @change="onPrimeChange" />
    <span :class="resolvedMessageClass">{{ message || ' ' }}</span>
  </div>
</template>
