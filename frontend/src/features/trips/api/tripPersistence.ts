import {
    addTripPerson,
    addTripPersonItem,
    addTripPersonPack,
    addTripPersonPackItem,
    createTrip,
    removeTripPerson,
    removeTripPersonItem,
    removeTripPersonPack,
    removeTripPersonPackItem,
    updateTrip,
} from './tripsApi'
import { daysToInterval, serviceToUrlField } from '../utils'
import type {
    StagedItem,
    StagedPerson,
    TripCreate,
    TripPersonDetails,
    TripWithDetails,
    StagedTrip,
} from '../types'

// Builds the canonical trip payload from staged form values.
function buildTripPayload(staged: StagedTrip): TripCreate {
    const distance = Number.parseFloat(staged.distanceKm)
    return {
        name: staged.name.trim(),
        trip_type: staged.tripType,
        notes: staged.notes.trim() || undefined,
        duration: daysToInterval(staged.durationDays),
        total_distance_km: Number.isFinite(distance) && distance > 0 ? distance : undefined,
        ...serviceToUrlField(staged.routeService, staged.routeUrl),
    }
}

// Persists every pack (and its items) plus direct items for a freshly added person.
async function persistPersonContents(
    tripId: string,
    person: StagedPerson,
    tripType: StagedTrip['tripType'],
): Promise<void> {
    for (const pack of person.packs) {
        const created = await addTripPersonPack(tripId, person.personId, {
            name: pack.name.trim() || 'Pack',
            trip_type: tripType,
        })
        for (const item of pack.items) {
            await addTripPersonPackItem(tripId, person.personId, created.pack_id, {
                item_id: item.itemId,
                quantity: item.quantity,
                carry_status: item.carryStatus,
                notes: item.notes,
            })
        }
    }

    for (const item of person.items) {
        await addTripPersonItem(tripId, person.personId, {
            item_id: item.itemId,
            quantity: item.quantity,
            carry_status: item.carryStatus,
            notes: item.notes,
        })
    }
}

// Creates a new trip together with all staged persons, packs, and items.
export async function persistNewTrip(staged: StagedTrip): Promise<string> {
    const trip = await createTrip(buildTripPayload(staged))

    for (const person of staged.persons) {
        await addTripPerson(trip.id, { person_id: person.personId })
        await persistPersonContents(trip.id, person, staged.tripType)
    }

    return trip.id
}

type OriginalItem = { item_id: string; quantity: number; carry_status: string }

// True when a staged item differs from its original counterpart (or has no counterpart).
function itemChanged(staged: StagedItem, original?: OriginalItem): boolean {
    return (
        !original
        || original.quantity !== staged.quantity
        || original.carry_status !== staged.carryStatus
    )
}

// Reconciles a single set of items by removing stale/changed entries and adding new/changed ones.
async function reconcileItemSet(
    originalItems: OriginalItem[],
    stagedItems: StagedItem[],
    addFn: (item: StagedItem) => Promise<unknown>,
    removeFn: (itemId: string) => Promise<unknown>,
): Promise<void> {
    const stagedByItem = new Map(stagedItems.map((item) => [item.itemId, item]))
    const originalByItem = new Map(originalItems.map((item) => [item.item_id, item]))

    for (const original of originalItems) {
        const staged = stagedByItem.get(original.item_id)
        if (!staged || itemChanged(staged, original)) {
            await removeFn(original.item_id)
        }
    }

    for (const staged of stagedItems) {
        if (itemChanged(staged, originalByItem.get(staged.itemId))) {
            await addFn(staged)
        }
    }
}

// Reconciles all packs for an already-attached person (add new, remove deleted, sync items).
async function reconcilePersonPacks(
    tripId: string,
    person: StagedPerson,
    original: TripPersonDetails,
    tripType: StagedTrip['tripType'],
): Promise<void> {
    const stagedPackIds = new Set(person.packs.filter((pack) => pack.packId).map((pack) => pack.packId))

    for (const originalPack of original.packs) {
        if (!stagedPackIds.has(originalPack.pack_id)) {
            await removeTripPersonPack(tripId, person.personId, originalPack.pack_id)
        }
    }

    const originalPackById = new Map(original.packs.map((pack) => [pack.pack_id, pack]))

    for (const pack of person.packs) {
        if (!pack.packId) {
            const created = await addTripPersonPack(tripId, person.personId, {
                name: pack.name.trim() || 'Pack',
                trip_type: tripType,
            })
            for (const item of pack.items) {
                await addTripPersonPackItem(tripId, person.personId, created.pack_id, {
                    item_id: item.itemId,
                    quantity: item.quantity,
                    carry_status: item.carryStatus,
                    notes: item.notes,
                })
            }
            continue
        }

        const originalPack = originalPackById.get(pack.packId)
        const packId = pack.packId
        await reconcileItemSet(
            originalPack?.items ?? [],
            pack.items,
            (item) =>
                addTripPersonPackItem(tripId, person.personId, packId, {
                    item_id: item.itemId,
                    quantity: item.quantity,
                    carry_status: item.carryStatus,
                    notes: item.notes,
                }),
            (itemId) => removeTripPersonPackItem(tripId, person.personId, packId, itemId),
        )
    }
}

// Reconciles a person's direct (non-pack) items.
async function reconcilePersonItems(
    tripId: string,
    person: StagedPerson,
    original: TripPersonDetails,
): Promise<void> {
    await reconcileItemSet(
        original.items,
        person.items,
        (item) =>
            addTripPersonItem(tripId, person.personId, {
                item_id: item.itemId,
                quantity: item.quantity,
                carry_status: item.carryStatus,
                notes: item.notes,
            }),
        (itemId) => removeTripPersonItem(tripId, person.personId, itemId),
    )
}

// Updates an existing trip, diffing staged state against the originally loaded details.
export async function persistTripUpdate(
    original: TripWithDetails,
    staged: StagedTrip,
): Promise<string> {
    await updateTrip(original.id, buildTripPayload(staged))

    const stagedPersonIds = new Set(staged.persons.map((person) => person.personId))
    for (const originalPerson of original.persons) {
        if (!stagedPersonIds.has(originalPerson.person_id)) {
            await removeTripPerson(original.id, originalPerson.person_id)
        }
    }

    const originalByPerson = new Map(original.persons.map((person) => [person.person_id, person]))

    for (const person of staged.persons) {
        const existing = originalByPerson.get(person.personId)
        if (!existing) {
            await addTripPerson(original.id, { person_id: person.personId })
            await persistPersonContents(original.id, person, staged.tripType)
            continue
        }

        await reconcilePersonPacks(original.id, person, existing, staged.tripType)
        await reconcilePersonItems(original.id, person, existing)
    }

    return original.id
}
