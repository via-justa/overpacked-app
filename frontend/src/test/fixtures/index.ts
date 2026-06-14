import type { components } from '../../lib/api/schema'

// Domain fixtures typed from the generated OpenAPI schema (single source of
// truth). Each factory returns a valid default and accepts a Partial override,
// so tests only specify the fields they care about.

type Schemas = components['schemas']
type Person = Schemas['Person']
type Item = Schemas['Item']
type Settings = Schemas['Settings']

const ISO = '2026-01-01T00:00:00Z'

export const settingsFixture = (overrides: Partial<Settings> = {}): Settings => ({
  id: 1,
  weight_unit: 'g',
  distance_unit: 'km',
  temperature_unit: 'c',
  volume_unit: 'ml',
  currency: 'usd',
  ...overrides,
})

export const personFixture = (overrides: Partial<Person> = {}): Person => ({
  id: 'p-1',
  name: 'Alice',
  gender: 'female',
  birthdate: '1990-01-01',
  body_weight_grams: 70000,
  conditioning_level: 'average',
  created_at: ISO,
  updated_at: ISO,
  ...overrides,
})

export const itemFixture = (overrides: Partial<Item> = {}): Item => ({
  id: 'i-1',
  name: 'Tent',
  type: 'shelter',
  is_active: true,
  manufacturer_id: 'm-1',
  weight_grams: 1200,
  value: 350,
  default_quantity: 1,
  default_carry_status: 'packed',
  created_at: ISO,
  updated_at: ISO,
  ...overrides,
})
