<script setup lang="ts">
import { Comment, Fragment, Text, computed, isVNode, useAttrs, useSlots, type VNode } from 'vue'
import MultiSelect from 'primevue/multiselect'

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

type ParsedSelectOption = {
  label: string
  value: string | number
  disabled?: boolean
}

// Extract text content from VNode children recursively
const toOptionLabel = (children: VNode['children']): string => {
  if (typeof children === 'string') {
    return children
  }

  if (!Array.isArray(children)) {
    return ''
  }

  return children
    .map((child) => {
      if (typeof child === 'string') {
        return child
      }

      if (isVNode(child)) {
        return toOptionLabel(child.children)
      }

      return ''
    })
    .join('')
}

// Parse <option> VNodes from slot into structured option objects
const parseOptionNodes = (nodes: VNode[]): ParsedSelectOption[] => {
  const parsed: ParsedSelectOption[] = []

  const visit = (vnodes: VNode[]) => {
    for (const node of vnodes) {
      if (!isVNode(node) || node.type === Comment || node.type === Text) {
        continue
      }

      if (node.type === Fragment && Array.isArray(node.children)) {
        visit(node.children as VNode[])
        continue
      }

      if (node.type === 'option') {
        const optionProps = (node.props ?? {}) as Record<string, unknown>
        const value = (optionProps.value ?? '') as string | number
        parsed.push({
          value,
          disabled: Boolean(optionProps.disabled),
          label: toOptionLabel(node.children).trim() || String(value),
        })
      }
    }
  }

  visit(nodes)
  return parsed
}

const parsedOptions = computed<ParsedSelectOption[]>(() => {
  return parseOptionNodes((slots.default?.() ?? []) as VNode[])
})

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

const resolvedMessageClass = computed(() => {
  const classes = ['block', 'min-h-4', 'truncate', 'text-xs', 'font-medium']

  if (props.message) {
    classes.push(props.invalid ? 'text-danger-500' : 'text-copy-muted')
  } else if (props.reserveMessageSpace) {
    classes.push('invisible')
  } else {
    classes.push('hidden')
  }

  if (props.messageClass) {
    classes.push(props.messageClass)
  }

  return classes.join(' ')
})

const multiSelectPt = computed(() => ({
  label: { class: 'app-select-label' },
  dropdown: { class: 'app-select-trigger' },
  overlay: { class: 'app-select-overlay' },
  listContainer: { class: 'app-select-list-container' },
  list: { class: 'app-select-list' },
  option: { class: 'app-select-option' },
}))

const onPrimeChange = (event: { value: Array<string | number>; originalEvent: Event }) => {
  emit('update:modelValue', event.value)
  emit('change', event.value, event.originalEvent)
}
</script>

<template>
  <div :class="wrapperClass">
    <MultiSelect v-bind="attrs" :model-value="modelValue" :options="parsedOptions" option-label="label"
      option-value="value" option-disabled="disabled" :placeholder="placeholder" :filter="filter"
      :show-toggle-all="showToggleAll" :max-selected-labels="maxSelectedLabels" :pt="multiSelectPt"
      :class="resolvedSelectClass" fluid @change="onPrimeChange" />
    <span :class="resolvedMessageClass">{{ message || ' ' }}</span>
  </div>
</template>
