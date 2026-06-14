import type { Item } from '../items/types'
import type { Person } from '../persons/types'
import type { CarryStatus, RouteService, TripType } from './types'

// Shared drag-and-drop group name so every zone (pool, worn, packs) can exchange cards.
export const TRIP_PLANNER_DND_GROUP = 'trip-planner-items'

// A single placement card. The same item may appear in several placements (split quantities).
export type PlannerPlacement = {
    // Local-only identifier used for keying and drag tracking.
    localId: string
    itemId: string
    quantity: number
    carryStatus: CarryStatus
    item: Item
}

// A named pack belonging to a planner person; holds the cards dropped into it on page 2.
export type PlannerPack = {
    localId: string
    name: string
    // Present when the pack already exists on the server (edit mode).
    packId?: string
    items: PlannerPlacement[]
}

// A person joining the trip, the packs they carry, and their worn (direct) items.
export type PlannerPerson = {
    personId: string
    person: Person
    // Present when the person is already attached to the trip (edit mode).
    tripPersonId?: string
    worn: PlannerPlacement[]
    packs: PlannerPack[]
    // The pack auto-assignment targets first; non-persistent, ordering only.
    mainPackLocalId: string | null
}

// Trip metadata captured on page 1 (mirrors the fields persisted via StagedTrip).
export type PlannerDetails = {
    name: string
    tripType: TripType
    notes: string
    durationDays: string
    distanceKm: string
    routeService: RouteService
    routeUrl: string
}

// Aggregated figures shown in the page-1 stats bar.
export type PlannerStats = {
    itemCount: number
    totalWeightGrams: number
    totalValue: number
}

// Per-person figures shown in the page-2 person column header.
export type PlannerPersonStats = {
    recommendedMaxGrams: number
    packedWeightGrams: number
    wornWeightGrams: number
    totalValue: number
    itemCount: number
    // Traffic-light status of packed weight versus the recommended maximum.
    weightStatus: 'ok' | 'warn' | 'over' | 'unknown'
}
