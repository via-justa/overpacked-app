/**
 * Common option type for select/toggle components
 */
export type SelectOption<T = string> = {
  label: string
  value: T
}

/**
 * Table field definition with render functions
 */
export type TableFieldDefinition<T = unknown> = {
  key: string
  label: string
  render: (item: T) => string
  renderHref?: (item: T) => string | undefined
  renderBoolean?: (item: T) => boolean | null | undefined
}

/**
 * Detail entry for display lists
 */
export type DetailEntry = {
  label: string
  value: string
  href?: string
  booleanValue?: boolean | null
}
