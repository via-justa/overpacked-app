import { persistNewTrip, persistTripUpdate } from './tripPersistence'
import type { StagedTrip, TripWithDetails } from '../types'

vi.mock('./tripsApi', () => ({
  createTrip: vi.fn(async () => ({ id: 'trip-1' })),
  updateTrip: vi.fn(async () => ({ id: 'trip-9' })),
  addTripPerson: vi.fn(async () => ({})),
  removeTripPerson: vi.fn(async () => ({})),
  addTripPersonPack: vi.fn(async () => ({ pack_id: 'new-pack' })),
  removeTripPersonPack: vi.fn(async () => ({})),
  addTripPersonPackItem: vi.fn(async () => ({})),
  removeTripPersonPackItem: vi.fn(async () => ({})),
  addTripPersonItem: vi.fn(async () => ({})),
  removeTripPersonItem: vi.fn(async () => ({})),
}))

import * as api from './tripsApi'

describe('persistNewTrip', () => {
  it('creates the trip and all staged persons, packs, and items', async () => {
    const staged = {
      name: '  Trek  ',
      tripType: 'multi_day',
      notes: ' bring poles ',
      durationDays: '3',
      distanceKm: '12.5',
      routeService: 'komoot',
      routeUrl: 'https://komoot.com/x',
      persons: [
        {
          personId: 'p-1',
          person: {},
          packs: [{ tempId: 't1', name: 'Main', items: [{ itemId: 'i-1', quantity: 2, carryStatus: 'packed' }] }],
          items: [{ itemId: 'i-2', quantity: 1, carryStatus: 'worn' }],
        },
      ],
    } as unknown as StagedTrip

    const id = await persistNewTrip(staged)
    expect(id).toBe('trip-1')

    expect(api.createTrip).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'Trek',
        trip_type: 'multi_day',
        notes: 'bring poles',
        duration: '3 days',
        total_distance_km: 12.5,
        trip_komoot_url: 'https://komoot.com/x',
      }),
    )
    expect(api.addTripPerson).toHaveBeenCalledWith('trip-1', { person_id: 'p-1' })
    expect(api.addTripPersonPack).toHaveBeenCalledWith('trip-1', 'p-1', expect.objectContaining({ name: 'Main' }))
    expect(api.addTripPersonPackItem).toHaveBeenCalledWith(
      'trip-1',
      'p-1',
      'new-pack',
      expect.objectContaining({ item_id: 'i-1', quantity: 2, carry_status: 'packed' }),
    )
    expect(api.addTripPersonItem).toHaveBeenCalledWith(
      'trip-1',
      'p-1',
      expect.objectContaining({ item_id: 'i-2', quantity: 1, carry_status: 'worn' }),
    )
  })

  it('omits empty optional fields from the payload', async () => {
    const staged = {
      name: 'T',
      tripType: 'day_hike',
      notes: '',
      durationDays: '',
      distanceKm: '',
      routeService: 'komoot',
      routeUrl: '',
      persons: [],
    } as unknown as StagedTrip

    await persistNewTrip(staged)
    expect(api.createTrip).toHaveBeenCalledWith({
      name: 'T',
      trip_type: 'day_hike',
      notes: undefined,
      duration: undefined,
      total_distance_km: undefined,
    })
  })
})

describe('persistTripUpdate', () => {
  it('diffs staged state against the original: drop person, change item, add person/pack', async () => {
    const original = {
      id: 'trip-9',
      persons: [
        { person_id: 'keep', packs: [{ pack_id: 'pk1', items: [{ item_id: 'i-1', quantity: 1, carry_status: 'packed' }] }], items: [] },
        { person_id: 'drop', packs: [], items: [] },
      ],
    } as unknown as TripWithDetails

    const staged = {
      name: 'Updated',
      tripType: 'day_hike',
      notes: '',
      durationDays: '',
      distanceKm: '',
      routeService: 'komoot',
      routeUrl: '',
      persons: [
        {
          personId: 'keep',
          person: {},
          packs: [
            // existing pack, item qty changed 1 -> 3 (remove + re-add)
            { tempId: 'a', packId: 'pk1', name: 'Main', items: [{ itemId: 'i-1', quantity: 3, carryStatus: 'packed' }] },
            // brand-new pack (no packId) with an item
            { tempId: 'b', name: 'Daypack', items: [{ itemId: 'i-9', quantity: 1, carryStatus: 'packed' }] },
          ],
          items: [],
        },
        // brand-new person
        { personId: 'added', person: {}, packs: [], items: [{ itemId: 'i-2', quantity: 1, carryStatus: 'worn' }] },
      ],
    } as unknown as StagedTrip

    const id = await persistTripUpdate(original, staged)
    expect(id).toBe('trip-9')

    expect(api.updateTrip).toHaveBeenCalledWith('trip-9', expect.objectContaining({ name: 'Updated' }))
    expect(api.removeTripPerson).toHaveBeenCalledWith('trip-9', 'drop')

    // changed item removed then re-added
    expect(api.removeTripPersonPackItem).toHaveBeenCalledWith('trip-9', 'keep', 'pk1', 'i-1')
    expect(api.addTripPersonPackItem).toHaveBeenCalledWith(
      'trip-9',
      'keep',
      'pk1',
      expect.objectContaining({ item_id: 'i-1', quantity: 3 }),
    )

    // new pack created and its item added
    expect(api.addTripPersonPack).toHaveBeenCalledWith('trip-9', 'keep', expect.objectContaining({ name: 'Daypack' }))
    expect(api.addTripPersonPackItem).toHaveBeenCalledWith(
      'trip-9',
      'keep',
      'new-pack',
      expect.objectContaining({ item_id: 'i-9' }),
    )

    // new person attached with their direct item
    expect(api.addTripPerson).toHaveBeenCalledWith('trip-9', { person_id: 'added' })
    expect(api.addTripPersonItem).toHaveBeenCalledWith('trip-9', 'added', expect.objectContaining({ item_id: 'i-2' }))
  })
})
