import { Comment, Fragment, Text, computed, isVNode, type ComputedRef, type Slots, type VNode } from 'vue'

export type ParsedSelectOption = {
  label: string
  value: string | number
  disabled?: boolean
}

// Extract text content from VNode children recursively.
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

// Parse <option> VNodes from a default slot into structured option objects.
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

// Shared PrimeVue passthrough wiring for the themed select / multiselect skins.
export const selectPassThrough = {
  label: { class: 'app-select-label' },
  dropdown: { class: 'app-select-trigger' },
  overlay: { class: 'app-select-overlay' },
  listContainer: { class: 'app-select-list-container' },
  list: { class: 'app-select-list' },
  option: { class: 'app-select-option' },
}

// Turns a component's <option> default slot into PrimeVue option objects.
export const useSelectOptions = (slots: Slots): { parsedOptions: ComputedRef<ParsedSelectOption[]> } => {
  const parsedOptions = computed<ParsedSelectOption[]>(() =>
    parseOptionNodes((slots.default?.() ?? []) as VNode[]),
  )

  return { parsedOptions }
}

// Builds the class list for a field's message line (error / muted / reserved space).
export const useFieldMessageClass = (params: {
  message?: string
  invalid?: boolean
  reserveMessageSpace?: boolean
  messageClass?: string
}): ComputedRef<string> =>
  computed(() => {
    const classes = ['block', 'min-h-4', 'truncate', 'text-xs', 'font-medium']

    if (params.message) {
      classes.push(params.invalid ? 'text-danger-500' : 'text-copy-muted')
    } else if (params.reserveMessageSpace) {
      classes.push('invisible')
    } else {
      classes.push('hidden')
    }

    if (params.messageClass) {
      classes.push(params.messageClass)
    }

    return classes.join(' ')
  })
