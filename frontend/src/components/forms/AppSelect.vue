<script setup lang="ts">
import { Comment, Fragment, Text, computed, isVNode, useAttrs, useSlots, type VNode } from 'vue'
import Select from 'primevue/select'

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

  if (props.compact) {
    classes.push('is-compact')
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

const selectPt = computed(() => ({
  label: { class: 'app-select-label' },
  dropdown: { class: 'app-select-trigger' },
  overlay: { class: 'app-select-overlay' },
  listContainer: { class: 'app-select-list-container' },
  list: { class: 'app-select-list' },
  option: { class: 'app-select-option' },
}))

const onPrimeChange = (event: { value: string; originalEvent: Event }) => {
  emit('update:modelValue', event.value)
  emit('change', event.value, event.originalEvent)
}
</script>

<template>
  <div :class="wrapperClass">
    <Select v-bind="attrs" :model-value="modelValue ?? ''" :options="parsedOptions" option-label="label"
      option-value="value" option-disabled="disabled" :pt="selectPt" :class="resolvedSelectClass" fluid
      @change="onPrimeChange" />
    <span :class="resolvedMessageClass">{{ message || ' ' }}</span>
  </div>
</template>
