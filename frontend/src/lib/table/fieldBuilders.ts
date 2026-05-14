import type { Item } from '../../features/items/types'

/**
 * Reusable table field configuration builder.
 * Creates standardized field definitions with render/renderHref/renderBoolean support.
 */
export type TableFieldDefinition = {
  key: string
  label: string
  render: (item: Item) => string
  renderHref?: (item: Item) => string | undefined
  renderBoolean?: (item: Item) => boolean | null | undefined
}

export type FieldBuilder = {
  key: string
  label: string
  render?: (item: Item) => string
  renderHref?: (item: Item) => string | undefined
  renderBoolean?: (item: Item) => boolean | null | undefined
}

/**
 * Creates a table field definition with optional custom render functions.
 * If no render function is provided, returns 'Not set' by default.
 */
export function createTableField(builder: FieldBuilder): TableFieldDefinition {
  return {
    key: builder.key,
    label: builder.label,
    render: builder.render ?? (() => 'Not set'),
    renderHref: builder.renderHref,
    renderBoolean: builder.renderBoolean,
  }
}

/**
 * Creates multiple table field definitions from an array of builders.
 */
export function createTableFields(builders: FieldBuilder[]): TableFieldDefinition[] {
  return builders.map(createTableField)
}

/**
 * Common field render helpers
 */
export const fieldRenderers = {
  boolean: (value: boolean | null | undefined): string => (value ? 'Yes' : 'No'),
  notSet: (value: string | null | undefined): string => (value && value.trim().length > 0 ? value : 'Not set'),
  number: (value: number | null | undefined): string =>
    typeof value === 'number' ? String(value) : 'Not set',
}
