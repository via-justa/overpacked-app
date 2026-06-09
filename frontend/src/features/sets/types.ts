import type { components } from '../../lib/api/schema'

// Server types are sourced from the generated OpenAPI schema (single source of truth).
export type ItemSet = components['schemas']['ItemSet']
export type ItemSetCreate = components['schemas']['ItemSetCreate']
export type ItemSetUpdate = components['schemas']['ItemSetUpdate']
export type SetItemCreate = components['schemas']['SetItemCreate']
export type SetItemUpdate = components['schemas']['SetItemUpdate']
export type SetItemWithDetails = components['schemas']['SetItemWithDetails']
