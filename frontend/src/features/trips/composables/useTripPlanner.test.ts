import { http, HttpResponse } from 'msw'
import { provideTripPlanner } from './useTripPlanner'
import { withSetup } from '../../../test/withSetup'
import { server } from '../../../test/msw/server'
import { personFixture, itemFixture } from '../../../test/fixtures'

// Seed the reference-data queries (persons/items) that the planner loads on
// creation. Returns the mounted context once those queries have resolved.
async function mountPlanner(opts: {
  persons?: ReturnType<typeof personFixture>[]
  items?: ReturnType<typeof itemFixture>[]
} = {}) {
  const persons = opts.persons ?? [personFixture({ id: 'p-1' })]
  const items = opts.items ?? [itemFixture({ id: 'i-1', weight_grams: 200, value: 5 })]
  server.use(
    http.get('*/api/v1/persons', () => HttpResponse.json(persons)),
    http.get('*/api/v1/items', () => HttpResponse.json(items)),
  )

  const handle = withSetup(() => provideTripPlanner())
  await vi.waitFor(() => {
    expect(handle.result.availablePersons.value).toHaveLength(persons.length)
    expect(handle.result.availableItems.value).toHaveLength(items.length)
  })
  return handle
}

describe('useTripPlanner', () => {
  it('assigns pool items to a person, then aggregates stats and the staged trip', async () => {
    const { result, unmount } = await mountPlanner()

    result.details.name = 'Weekend Trek'
    result.togglePerson(result.availablePersons.value[0])
    result.addPoolItem('i-1', 2, 'packed')

    expect(result.unassigned.value).toHaveLength(1)
    expect(result.stats.value).toMatchObject({ itemCount: 2, totalWeightGrams: 400, totalValue: 10 })
    expect(result.canSave.value).toBe(true)

    result.assignToPerson('p-1')
    expect(result.unassigned.value).toHaveLength(0)

    const stats = result.personStats(result.persons.value[0])
    expect(stats.packedWeightGrams).toBe(400)
    expect(stats.itemCount).toBe(2)
    expect(stats.weightStatus).toBe('ok') // 400g well under the recommended max

    const staged = result.buildStagedTrip()
    expect(staged.name).toBe('Weekend Trek')
    expect(staged.persons).toHaveLength(1)
    expect(staged.persons[0].packs[0].items).toEqual([
      expect.objectContaining({ itemId: 'i-1', quantity: 2, carryStatus: 'packed' }),
    ])
    expect(staged.persons[0].items).toEqual([]) // nothing worn

    unmount()
  })

  it('flags an over-weight pack', async () => {
    const { result, unmount } = await mountPlanner({
      items: [itemFixture({ id: 'heavy', weight_grams: 50_000, value: 0 })],
    })

    result.togglePerson(result.availablePersons.value[0])
    result.addPoolItem('heavy', 1, 'packed')
    result.assignToPerson('p-1')

    expect(result.personStats(result.persons.value[0]).weightStatus).toBe('over')
    unmount()
  })

  it('merges, splits, resizes and removes pool placements', async () => {
    const { result, unmount } = await mountPlanner()

    result.addPoolItem('i-1', 1, 'packed')
    result.addPoolItem('i-1', 2, 'packed') // same item + status → merges
    expect(result.unassigned.value).toHaveLength(1)
    expect(result.unassigned.value[0].quantity).toBe(3)

    result.splitPlacement(result.unassigned.value[0]) // peel one unit off
    expect(result.unassigned.value).toHaveLength(2)
    expect(result.unassigned.value.map((p) => p.quantity).sort()).toEqual([1, 2])

    result.setPlacementQuantity(result.unassigned.value[0], 5)
    expect(result.unassigned.value[0].quantity).toBe(5)
    result.setPlacementQuantity(result.unassigned.value[0], 0) // ignored (< 1)
    expect(result.unassigned.value[0].quantity).toBe(5)

    result.removePlacement(result.unassigned.value[0])
    expect(result.unassigned.value).toHaveLength(1)
    unmount()
  })

  it('returns placed items to the pool when a person is de-selected', async () => {
    const { result, unmount } = await mountPlanner()

    result.togglePerson(result.availablePersons.value[0])
    result.addPoolItem('i-1', 1, 'packed')
    result.assignToPerson('p-1')
    expect(result.unassigned.value).toHaveLength(0)

    result.togglePerson(result.availablePersons.value[0]) // de-select
    expect(result.persons.value).toHaveLength(0)
    expect(result.unassigned.value).toHaveLength(1) // item came back
    unmount()
  })

  it('loads an existing trip into edit mode with mapped details', async () => {
    server.use(
      http.get('*/api/v1/trips/:tripId', () =>
        HttpResponse.json({
          name: 'Existing Trip',
          trip_type: 'multi_day',
          notes: 'bring poles',
          duration: '4 days',
          total_distance_km: 42,
          persons: [],
        }),
      ),
    )

    const { result, unmount } = await mountPlanner()
    await result.loadExisting('trip-123')

    expect(result.isEditMode.value).toBe(true)
    expect(result.isLoading.value).toBe(false)
    expect(result.details.name).toBe('Existing Trip')
    expect(result.details.durationDays).toBe('4')
    expect(result.details.distanceKm).toBe('42')
    unmount()
  })
})
