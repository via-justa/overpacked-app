import type { components } from '../../lib/api/schema'

// Server types are sourced from the generated OpenAPI schema (single source of truth).
export type Person = components['schemas']['Person']
export type PersonCreate = components['schemas']['PersonCreate']
export type PersonUpdate = components['schemas']['PersonUpdate']

// Named enums derived from the schema for use in forms/selects.
export type Gender = NonNullable<Person['gender']>
export type ConditioningLevel = NonNullable<Person['conditioning_level']>

// UI-only form state — no server/spec equivalent.
export type PersonFormValues = {
  name: string
  gender: Gender
  birthdate: string
  body_weight_value: string
  conditioning_level: ConditioningLevel
}
