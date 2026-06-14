import type { components } from '../../lib/api/schema'

// Server types are sourced from the generated OpenAPI schema (single source of truth).
export type PackingList = components['schemas']['PackingList']
export type PackingListCreate = components['schemas']['PackingListCreate']
export type PackingListUpdate = components['schemas']['PackingListUpdate']
export type Label = components['schemas']['Label']
export type LabelCreate = components['schemas']['LabelCreate']
export type PackingListLabelAdd = components['schemas']['PackingListLabelAdd']
