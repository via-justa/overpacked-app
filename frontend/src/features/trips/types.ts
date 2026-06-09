import type { components } from '../../lib/api/schema'
import type { Item } from '../items/types'
import type { Person } from '../persons/types'

// Server types are sourced from the generated OpenAPI schema (single source of truth).
export type Trip = components['schemas']['Trip']
export type TripCreate = components['schemas']['TripCreate']
export type TripUpdate = components['schemas']['TripUpdate']
export type TripWithDetails = components['schemas']['TripWithDetails']
export type TripPersonDetails = components['schemas']['TripPersonDetailsNested']
export type TripPersonPack = components['schemas']['TripPersonPackDetailsNested']
// The create-pack endpoint returns the pack record without its (still-empty) items.
export type TripPersonPackCreated = components['schemas']['TripPersonPackWithDetails']
export type TripPersonPackItem = components['schemas']['PackItemWithDetails']
export type TripPersonItem = components['schemas']['TripPersonItemWithDetails']
export type TripPersonCreate = components['schemas']['TripPersonCreate']
export type TripPersonPackCreate = components['schemas']['TripPersonPackCreate']
export type TripPersonPackItemCreate = components['schemas']['PackItemCreate']
export type TripPersonItemCreate = components['schemas']['TripPersonItemCreate']
export type TripRoutePreview = components['schemas']['TripRoutePreview']

// Named enums derived from the schema for use in forms/selects.
export type TripType = Trip['trip_type']
export type CarryStatus = TripPersonItem['carry_status']
export type RouteService = TripRoutePreview['service']

// ─── Aggregated stats used by cards ──────────────────────────────────────────

export type TripStats = {
    packsCount: number
    travelersCount: number
    packedWeightGrams: number
    wornWeightGrams: number
    totalWeightGrams: number
    totalValue: number
}

// ─── Staged (dialog-local) structures for the create/update wizard ───────────

export type StagedItem = {
    itemId: string
    quantity: number
    carryStatus: CarryStatus
    notes?: string
    item: Item
}

export type StagedPack = {
    // Local-only identifier used to track packs before persistence.
    tempId: string
    name: string
    // Present when the pack already exists on the server (edit mode).
    packId?: string
    items: StagedItem[]
}

export type StagedPerson = {
    personId: string
    person: Person
    // Present when the person is already attached to the trip (edit mode).
    tripPersonId?: string
    packs: StagedPack[]
    // Direct (worn / loose) items not in a pack.
    items: StagedItem[]
}

export type StagedTrip = {
    name: string
    tripType: TripType
    notes: string
    durationDays: string
    distanceKm: string
    routeService: RouteService
    routeUrl: string
    persons: StagedPerson[]
}
