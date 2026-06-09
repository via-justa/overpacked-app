import { apiClient } from '../../../lib/api/client'
import type {
    RouteService,
    Trip,
    TripCreate,
    TripUpdate,
    TripWithDetails,
    TripPersonCreate,
    TripPersonItem,
    TripPersonItemCreate,
    TripPersonPackCreated,
    TripPersonPackCreate,
    TripPersonPackItem,
    TripPersonPackItemCreate,
    TripRoutePreview,
} from '../types'

const readString = (value: unknown): string | null => {
    if (typeof value === 'string' && value.trim().length > 0) {
        return value
    }
    return null
}

const getErrorMessage = (error: unknown, fallback: string): string => {
    if (!error || typeof error !== 'object') {
        return fallback
    }

    const objectError = error as {
        error?: unknown
        message?: unknown
        detail?: unknown
        details?: unknown
    }

    return (
        readString(objectError.error)
        ?? readString(objectError.message)
        ?? readString(objectError.detail)
        ?? readString(objectError.details)
        ?? fallback
    )
}

// ─── Trip CRUD ───────────────────────────────────────────────────────────────

export const listTrips = async (): Promise<Trip[]> => {
    const { data, error, response } = await apiClient.GET('/api/v1/trips')

    if (!response.ok || !data) {
        throw new Error(getErrorMessage(error, 'Unable to load trips'))
    }

    return data
}

export const getTrip = async (tripId: string): Promise<TripWithDetails> => {
    const { data, error, response } = await apiClient.GET('/api/v1/trips/{tripId}', {
        params: { path: { tripId } },
    })

    if (!response.ok || !data) {
        throw new Error(getErrorMessage(error, 'Unable to load trip'))
    }

    return data
}

export const createTrip = async (payload: TripCreate): Promise<Trip> => {
    const { data, error, response } = await apiClient.POST('/api/v1/trips', {
        body: payload,
    })

    if (!response.ok || !data) {
        throw new Error(getErrorMessage(error, 'Unable to create trip'))
    }

    return data
}

export const updateTrip = async (tripId: string, payload: TripUpdate): Promise<Trip> => {
    const { data, error, response } = await apiClient.PATCH('/api/v1/trips/{tripId}', {
        params: { path: { tripId } },
        body: payload,
    })

    if (!response.ok || !data) {
        throw new Error(getErrorMessage(error, 'Unable to update trip'))
    }

    return data
}

export const removeTrip = async (tripId: string): Promise<void> => {
    const { error, response } = await apiClient.DELETE('/api/v1/trips/{tripId}', {
        params: { path: { tripId } },
    })

    if (!response.ok) {
        throw new Error(getErrorMessage(error, 'Unable to delete trip'))
    }
}

// ─── Route preview ───────────────────────────────────────────────────────────

export const getRoutePreview = async (
    service: Exclude<RouteService, 'unknown'>,
    url: string,
): Promise<TripRoutePreview> => {
    const { data, error, response } = await apiClient.GET('/api/v1/trips/route-preview/{service}', {
        params: { path: { service }, query: { url } },
    })

    if (!response.ok || !data) {
        throw new Error(getErrorMessage(error, 'Unable to load route preview'))
    }

    return data
}

// ─── Trip persons ────────────────────────────────────────────────────────────

export const addTripPerson = async (tripId: string, payload: TripPersonCreate): Promise<void> => {
    const { error, response } = await apiClient.POST('/api/v1/trips/{tripId}/persons', {
        params: { path: { tripId } },
        body: payload,
    })

    if (!response.ok) {
        throw new Error(getErrorMessage(error, 'Unable to add person to trip'))
    }
}

export const removeTripPerson = async (tripId: string, personId: string): Promise<void> => {
    const { error, response } = await apiClient.DELETE('/api/v1/trips/{tripId}/persons/{personId}', {
        params: { path: { tripId, personId } },
    })

    if (!response.ok) {
        throw new Error(getErrorMessage(error, 'Unable to remove person from trip'))
    }
}

// ─── Trip person packs ───────────────────────────────────────────────────────

export const addTripPersonPack = async (
    tripId: string,
    personId: string,
    payload: TripPersonPackCreate,
): Promise<TripPersonPackCreated> => {
    const { data, error, response } = await apiClient.POST(
        '/api/v1/trips/{tripId}/persons/{personId}/packs',
        {
            params: { path: { tripId, personId } },
            body: payload,
        },
    )

    if (!response.ok || !data) {
        throw new Error(getErrorMessage(error, 'Unable to create pack'))
    }

    return data
}

export const removeTripPersonPack = async (
    tripId: string,
    personId: string,
    packId: string,
): Promise<void> => {
    const { error, response } = await apiClient.DELETE(
        '/api/v1/trips/{tripId}/persons/{personId}/packs/{packId}',
        {
            params: { path: { tripId, personId, packId } },
        },
    )

    if (!response.ok) {
        throw new Error(getErrorMessage(error, 'Unable to remove pack'))
    }
}

// ─── Trip person pack items ──────────────────────────────────────────────────

export const addTripPersonPackItem = async (
    tripId: string,
    personId: string,
    packId: string,
    payload: TripPersonPackItemCreate,
): Promise<TripPersonPackItem> => {
    const { data, error, response } = await apiClient.POST(
        '/api/v1/trips/{tripId}/persons/{personId}/packs/{packId}/items',
        {
            params: { path: { tripId, personId, packId } },
            body: payload,
        },
    )

    if (!response.ok || !data) {
        throw new Error(getErrorMessage(error, 'Unable to add item to pack'))
    }

    return data
}

export const removeTripPersonPackItem = async (
    tripId: string,
    personId: string,
    packId: string,
    itemId: string,
): Promise<void> => {
    const { error, response } = await apiClient.DELETE(
        '/api/v1/trips/{tripId}/persons/{personId}/packs/{packId}/items/{itemId}',
        {
            params: { path: { tripId, personId, packId, itemId } },
        },
    )

    if (!response.ok) {
        throw new Error(getErrorMessage(error, 'Unable to remove item from pack'))
    }
}

// ─── Trip person direct items ────────────────────────────────────────────────

export const addTripPersonItem = async (
    tripId: string,
    personId: string,
    payload: TripPersonItemCreate,
): Promise<TripPersonItem> => {
    const { data, error, response } = await apiClient.POST(
        '/api/v1/trips/{tripId}/persons/{personId}/items',
        {
            params: { path: { tripId, personId } },
            body: payload,
        },
    )

    if (!response.ok || !data) {
        throw new Error(getErrorMessage(error, 'Unable to add item to person'))
    }

    return data
}

export const removeTripPersonItem = async (
    tripId: string,
    personId: string,
    itemId: string,
): Promise<void> => {
    const { error, response } = await apiClient.DELETE(
        '/api/v1/trips/{tripId}/persons/{personId}/items/{itemId}',
        {
            params: { path: { tripId, personId, itemId } },
        },
    )

    if (!response.ok) {
        throw new Error(getErrorMessage(error, 'Unable to remove item from person'))
    }
}
