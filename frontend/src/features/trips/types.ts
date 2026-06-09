import type { Item } from '../items/types'
import type { Person } from '../persons/types'

export type TripType = 'day_hike' | 'overnight' | 'multi_day' | 'thru_hike'

export type CarryStatus = 'packed' | 'worn'

export type RouteService = 'komoot' | 'strava' | 'wanderer' | 'unknown'

export type Trip = {
    id: string
    name: string
    trip_type: TripType
    duration?: string | null
    notes?: string | null
    trip_komoot_url?: string | null
    trip_strava_url?: string | null
    trip_wanderer_url?: string | null
    total_distance_km?: number | null
    created_at: string
    updated_at: string
}

export type TripCreate = {
    name: string
    trip_type: TripType
    duration?: string
    notes?: string
    trip_komoot_url?: string
    trip_strava_url?: string
    trip_wanderer_url?: string
    total_distance_km?: number
}

export type TripUpdate = Partial<TripCreate>

export type TripPersonPackItem = {
    id: string
    item_id: string
    quantity: number
    carry_status: CarryStatus
    notes?: string | null
    item: Item
}

export type TripPersonPack = {
    id: string
    trip_person_id: string
    pack_id: string
    pack: {
        id: string
        name: string
        trip_type: TripType
        person_id?: string | null
        created_at: string
        updated_at: string
    }
    items: TripPersonPackItem[]
    created_at: string
    updated_at: string
}

export type TripPersonItem = {
    id: string
    trip_person_id: string
    item_id: string
    quantity: number
    carry_status: CarryStatus
    notes?: string | null
    item: Item
    created_at: string
    updated_at: string
}

export type TripPersonDetails = {
    trip_person_id: string
    person_id: string
    person: Person
    packs: TripPersonPack[]
    items: TripPersonItem[]
}

export type TripWithDetails = Trip & {
    persons: TripPersonDetails[]
}

export type TripPersonCreate = {
    person_id: string
}

export type TripPersonPackCreate = {
    name: string
    trip_type: TripType
    notes?: string | null
}

export type TripPersonPackItemCreate = {
    item_id: string
    quantity: number
    carry_status: CarryStatus
    notes?: string | null
}

export type TripPersonItemCreate = {
    item_id: string
    quantity: number
    carry_status: CarryStatus
    notes?: string
}

export type TripRoutePreview = {
    service: RouteService
    image_url?: string | null
    title?: string | null
}

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
