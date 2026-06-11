import { apiClient } from '../../../lib/api/client'
import { ensureApiResponse, unwrapApiResponse } from '../../../lib/api/request'
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

// ─── Trip CRUD ───────────────────────────────────────────────────────────────

export const listTrips = async (): Promise<Trip[]> =>
    unwrapApiResponse(apiClient.GET('/api/v1/trips'), 'Unable to load trips')

export const getTrip = async (tripId: string): Promise<TripWithDetails> =>
    unwrapApiResponse(
        apiClient.GET('/api/v1/trips/{tripId}', { params: { path: { tripId } } }),
        'Unable to load trip',
    )

export const createTrip = async (payload: TripCreate): Promise<Trip> =>
    unwrapApiResponse(apiClient.POST('/api/v1/trips', { body: payload }), 'Unable to create trip')

export const updateTrip = async (tripId: string, payload: TripUpdate): Promise<Trip> =>
    unwrapApiResponse(
        apiClient.PATCH('/api/v1/trips/{tripId}', { params: { path: { tripId } }, body: payload }),
        'Unable to update trip',
    )

export const removeTrip = async (tripId: string): Promise<void> =>
    ensureApiResponse(apiClient.DELETE('/api/v1/trips/{tripId}', { params: { path: { tripId } } }), 'Unable to delete trip')

// ─── Route preview ───────────────────────────────────────────────────────────

export const getRoutePreview = async (
    service: Exclude<RouteService, 'unknown'>,
    url: string,
): Promise<TripRoutePreview> =>
    unwrapApiResponse(
        apiClient.GET('/api/v1/trips/route-preview/{service}', { params: { path: { service }, query: { url } } }),
        'Unable to load route preview',
    )

// ─── Trip persons ────────────────────────────────────────────────────────────

export const addTripPerson = async (tripId: string, payload: TripPersonCreate): Promise<void> =>
    ensureApiResponse(
        apiClient.POST('/api/v1/trips/{tripId}/persons', { params: { path: { tripId } }, body: payload }),
        'Unable to add person to trip',
    )

export const removeTripPerson = async (tripId: string, personId: string): Promise<void> =>
    ensureApiResponse(
        apiClient.DELETE('/api/v1/trips/{tripId}/persons/{personId}', { params: { path: { tripId, personId } } }),
        'Unable to remove person from trip',
    )

// ─── Trip person packs ───────────────────────────────────────────────────────

export const addTripPersonPack = async (
    tripId: string,
    personId: string,
    payload: TripPersonPackCreate,
): Promise<TripPersonPackCreated> =>
    unwrapApiResponse(
        apiClient.POST('/api/v1/trips/{tripId}/persons/{personId}/packs', {
            params: { path: { tripId, personId } },
            body: payload,
        }),
        'Unable to create pack',
    )

export const removeTripPersonPack = async (
    tripId: string,
    personId: string,
    packId: string,
): Promise<void> =>
    ensureApiResponse(
        apiClient.DELETE('/api/v1/trips/{tripId}/persons/{personId}/packs/{packId}', {
            params: { path: { tripId, personId, packId } },
        }),
        'Unable to remove pack',
    )

// ─── Trip person pack items ──────────────────────────────────────────────────

export const addTripPersonPackItem = async (
    tripId: string,
    personId: string,
    packId: string,
    payload: TripPersonPackItemCreate,
): Promise<TripPersonPackItem> =>
    unwrapApiResponse(
        apiClient.POST('/api/v1/trips/{tripId}/persons/{personId}/packs/{packId}/items', {
            params: { path: { tripId, personId, packId } },
            body: payload,
        }),
        'Unable to add item to pack',
    )

export const removeTripPersonPackItem = async (
    tripId: string,
    personId: string,
    packId: string,
    itemId: string,
): Promise<void> =>
    ensureApiResponse(
        apiClient.DELETE('/api/v1/trips/{tripId}/persons/{personId}/packs/{packId}/items/{itemId}', {
            params: { path: { tripId, personId, packId, itemId } },
        }),
        'Unable to remove item from pack',
    )

// ─── Trip person direct items ────────────────────────────────────────────────

export const addTripPersonItem = async (
    tripId: string,
    personId: string,
    payload: TripPersonItemCreate,
): Promise<TripPersonItem> =>
    unwrapApiResponse(
        apiClient.POST('/api/v1/trips/{tripId}/persons/{personId}/items', {
            params: { path: { tripId, personId } },
            body: payload,
        }),
        'Unable to add item to person',
    )

export const removeTripPersonItem = async (
    tripId: string,
    personId: string,
    itemId: string,
): Promise<void> =>
    ensureApiResponse(
        apiClient.DELETE('/api/v1/trips/{tripId}/persons/{personId}/items/{itemId}', {
            params: { path: { tripId, personId, itemId } },
        }),
        'Unable to remove item from person',
    )
