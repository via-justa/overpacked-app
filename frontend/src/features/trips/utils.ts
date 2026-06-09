import type {
    RouteService,
    Trip,
    TripStats,
    TripType,
    TripWithDetails,
} from './types'

export const TRIP_TYPE_OPTIONS: ReadonlyArray<{ value: TripType; label: string }> = [
    { value: 'day_hike', label: 'Day Hike' },
    { value: 'overnight', label: 'Overnight' },
    { value: 'multi_day', label: 'Multi Day' },
    { value: 'thru_hike', label: 'Thru Hike' },
]

export function formatTripType(type: TripType): string {
    return TRIP_TYPE_OPTIONS.find((option) => option.value === type)?.label ?? type
}

// Selectable route services for the trip form (excludes the 'unknown' fallback).
export const ROUTE_SERVICE_OPTIONS: ReadonlyArray<{ value: Exclude<RouteService, 'unknown'>; label: string }> = [
    { value: 'komoot', label: 'Komoot' },
    { value: 'strava', label: 'Strava' },
    { value: 'wanderer', label: 'Wanderer' },
]

// ─── Route URL service detection ─────────────────────────────────────────────

/**
 * Detects the route service from a share URL host. Mirrors the backend allowlist.
 */
export function detectRouteService(rawUrl: string): RouteService {
    const trimmed = rawUrl.trim()
    if (!trimmed) {
        return 'unknown'
    }

    let host: string
    try {
        host = new URL(trimmed).hostname.toLowerCase()
    } catch {
        return 'unknown'
    }

    if (host === 'komoot.com' || host.endsWith('.komoot.com') || host === 'komoot.de' || host.endsWith('.komoot.de')) {
        return 'komoot'
    }
    if (host === 'strava.com' || host.endsWith('.strava.com') || host === 'strava.app.link' || host.endsWith('.strava.app.link')) {
        return 'strava'
    }
    if (host === 'wanderer.to' || host.endsWith('.wanderer.to')) {
        return 'wanderer'
    }

    return 'unknown'
}

/**
 * Maps a route service to its icon name within the `content` icon category.
 */
export const ROUTE_SERVICE_ICONS: Record<RouteService, string> = {
    komoot: 'routeKomoot',
    strava: 'routeStrava',
    wanderer: 'routeWanderer',
    unknown: 'routeLink',
}

/**
 * Maps a route URL onto the backend's per-service field, based on the
 * user-selected service rather than the URL host (self-hosted wanderer
 * instances cannot be detected from the host alone).
 */
export function serviceToUrlField(service: RouteService, rawUrl: string): {
    trip_komoot_url?: string
    trip_strava_url?: string
    trip_wanderer_url?: string
} {
    const url = rawUrl.trim()
    if (!url) {
        return {}
    }

    switch (service) {
        case 'komoot':
            return { trip_komoot_url: url }
        case 'strava':
            return { trip_strava_url: url }
        case 'wanderer':
            return { trip_wanderer_url: url }
        default:
            return {}
    }
}

/**
 * Returns the first populated service URL on a trip.
 */
export function getTripRouteUrl(trip: Trip): string {
    return trip.trip_komoot_url ?? trip.trip_strava_url ?? trip.trip_wanderer_url ?? ''
}

/**
 * Returns the route service of a trip based on which per-service URL field is populated.
 */
export function getTripRouteService(trip: Trip): RouteService {
    if (trip.trip_komoot_url) {
        return 'komoot'
    }
    if (trip.trip_strava_url) {
        return 'strava'
    }
    if (trip.trip_wanderer_url) {
        return 'wanderer'
    }
    return 'unknown'
}

/**
 * Cleans a route page title (e.g. og:title) into a usable trip name.
 * Route services append " | sport | Komoot"-style suffixes, so keep the first segment.
 */
export function cleanRouteTitle(rawTitle: string | null | undefined): string {
    if (!rawTitle) {
        return ''
    }
    return rawTitle.split('|')[0].trim()
}

// ─── Duration (days) <-> PostgreSQL interval string ──────────────────────────

/**
 * Converts a number of days into a PostgreSQL interval string (e.g. "3 days").
 */
export function daysToInterval(days: string | number): string | undefined {
    const value = typeof days === 'number' ? days : Number.parseInt(days.trim(), 10)
    if (!Number.isFinite(value) || value <= 0) {
        return undefined
    }
    return `${value} days`
}

/**
 * Parses a PostgreSQL interval string back into a whole number of days.
 */
export function intervalToDays(interval?: string | null): number | null {
    if (!interval) {
        return null
    }

    const dayMatch = /(\d+)\s*day/i.exec(interval)
    if (dayMatch) {
        return Number.parseInt(dayMatch[1], 10)
    }

    // Fall back to a leading integer (e.g. "3" or "3:00:00").
    const leading = /^(\d+)/.exec(interval.trim())
    return leading ? Number.parseInt(leading[1], 10) : null
}

export function formatDurationDays(interval?: string | null): string {
    const days = intervalToDays(interval)
    if (days === null) {
        return 'Not set'
    }
    return days === 1 ? '1 day' : `${days} days`
}

// ─── Distance helpers ────────────────────────────────────────────────────────

/**
 * Calculates the average distance per day from total distance and duration.
 */
export function averageDistancePerDay(distanceKm?: number | null, interval?: string | null): number | null {
    const days = intervalToDays(interval)
    if (distanceKm == null || days == null || days <= 0) {
        return null
    }
    return distanceKm / days
}

// ─── Stats aggregation from a fully nested trip ──────────────────────────────

/**
 * Aggregates pack + worn weight, value, and counts from a nested trip.
 */
export function computeTripStats(trip: TripWithDetails): TripStats {
    let packedWeightGrams = 0
    let wornWeightGrams = 0
    let totalValue = 0
    let packsCount = 0

    const addItem = (quantity: number, carryStatus: string, weightGrams?: number | null, value?: number | null) => {
        const weight = (typeof weightGrams === 'number' ? weightGrams : 0) * quantity
        if (carryStatus === 'worn') {
            wornWeightGrams += weight
        } else {
            packedWeightGrams += weight
        }
        totalValue += (typeof value === 'number' ? value : 0) * quantity
    }

    for (const person of trip.persons) {
        packsCount += person.packs.length
        for (const pack of person.packs) {
            for (const entry of pack.items) {
                addItem(entry.quantity, entry.carry_status, entry.item.weight_grams, entry.item.value)
            }
        }
        for (const entry of person.items) {
            addItem(entry.quantity, entry.carry_status, entry.item.weight_grams, entry.item.value)
        }
    }

    return {
        packsCount,
        travelersCount: trip.persons.length,
        packedWeightGrams,
        wornWeightGrams,
        totalWeightGrams: packedWeightGrams + wornWeightGrams,
        totalValue,
    }
}
