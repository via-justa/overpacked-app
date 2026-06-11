import { computed, inject, provide, reactive, ref, type ComputedRef, type InjectionKey, type Ref } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { listPersons } from '../../persons/api/personsApi'
import { listItems, listItemLabels } from '../../items/api/itemsApi'
import { listPackingLists, listPackingListLabels } from '../../lists/api/listsApi'
import { listSetItems } from '../../sets/api/setsApi'
import { getTrip } from '../api/tripsApi'
import { getTripRouteService, getTripRouteUrl, intervalToDays } from '../utils'
import { getPersonRecommendedMaxWeightGrams } from '../../persons/utils'
import type { Item } from '../../items/types'
import type { Person } from '../../persons/types'
import type { Label } from '../../lists/types'
import type {
    PlannerDetails,
    PlannerPerson,
    PlannerPersonStats,
    PlannerPlacement,
    PlannerStats,
} from '../plannerTypes'
import type {
    CarryStatus,
    StagedItem,
    StagedPerson,
    StagedTrip,
    TripPersonItem,
    TripPersonPackItem,
    TripWithDetails,
} from '../types'

// Traffic-light thresholds for packed weight relative to recommended maximum.
const WEIGHT_WARN_RATIO = 0.9
const WEIGHT_OVER_RATIO = 1.1
const DEFAULT_CARRY_STATUS: CarryStatus = 'packed'

let localIdCounter = 0

// Generates a process-unique local id for client-side tracking before persistence.
function nextLocalId(prefix: string): string {
    localIdCounter += 1
    return `${prefix}-${localIdCounter}`
}

function weightOf(placement: PlannerPlacement): number {
    return (placement.item.weight_grams ?? 0) * placement.quantity
}

function valueOf(placement: PlannerPlacement): number {
    return (placement.item.value ?? 0) * placement.quantity
}

// Collapses placements that share an item into one staged item (summing quantities).
function aggregateStagedItems(placements: PlannerPlacement[], carryStatus: CarryStatus): StagedItem[] {
    const byItem = new Map<string, StagedItem>()
    for (const placement of placements) {
        const existing = byItem.get(placement.itemId)
        if (existing) {
            existing.quantity += placement.quantity
            continue
        }
        byItem.set(placement.itemId, {
            itemId: placement.itemId,
            quantity: placement.quantity,
            carryStatus,
            item: placement.item,
        })
    }
    return [...byItem.values()]
}

function placementFromEntry(entry: TripPersonItem | TripPersonPackItem): PlannerPlacement {
    return {
        localId: nextLocalId('placement'),
        itemId: entry.item_id,
        quantity: entry.quantity,
        carryStatus: entry.carry_status,
        item: entry.item,
    }
}

function emptyDetails(): PlannerDetails {
    return {
        name: '',
        tripType: 'day_hike',
        notes: '',
        durationDays: '',
        distanceKm: '',
        routeService: 'komoot',
        routeUrl: '',
    }
}

// The shape exposed to every planner child component via provide/inject.
export type TripPlannerContext = {
    step: Ref<1 | 2>
    isEditMode: ComputedRef<boolean>
    isLoading: Ref<boolean>
    details: PlannerDetails
    persons: Ref<PlannerPerson[]>
    unassigned: Ref<PlannerPlacement[]>
    selectedPackingListId: Ref<string>
    filterLabelId: Ref<string | null>
    originalDetails: Ref<TripWithDetails | null>

    availablePersons: ComputedRef<Person[]>
    availableItems: ComputedRef<Item[]>
    packingLists: ComputedRef<{ id: string; name: string }[]>
    packingListLabels: ComputedRef<Label[]>
    itemLabelsByItemId: ComputedRef<Map<string, Label[]>>
    coveredLabelIds: ComputedRef<Set<string>>

    stats: ComputedRef<PlannerStats>
    canProceedToStep2: ComputedRef<boolean>
    canSave: ComputedRef<boolean>

    addPoolItem: (itemId: string, quantity: number, carryStatus: CarryStatus) => void
    addSet: (setId: string) => Promise<void>
    removePlacement: (placement: PlannerPlacement) => void
    setPlacementQuantity: (placement: PlannerPlacement, quantity: number) => void
    splitPlacement: (placement: PlannerPlacement) => void

    isPersonSelected: (personId: string) => boolean
    togglePerson: (person: Person) => void
    addPack: (personId: string, name: string) => void
    removePack: (personId: string, packLocalId: string) => void
    renamePack: (personId: string, packLocalId: string, name: string) => void
    setMainPack: (personId: string, packLocalId: string) => void

    autoAssignToPerson: (personId: string, placement: PlannerPlacement) => void
    assignToPerson: (personId: string, placements?: PlannerPlacement[]) => void
    personStats: (person: PlannerPerson) => PlannerPersonStats

    buildStagedTrip: () => StagedTrip
    loadExisting: (tripId: string) => Promise<void>
    reset: () => void
}

const TRIP_PLANNER_KEY: InjectionKey<TripPlannerContext> = Symbol('trip-planner')

// Builds the full planner state object plus all mutation/query helpers.
function createTripPlannerContext(): TripPlannerContext {
    const step = ref<1 | 2>(1)
    const isLoading = ref(false)
    const details = reactive<PlannerDetails>(emptyDetails())
    const persons = ref<PlannerPerson[]>([])
    const unassigned = ref<PlannerPlacement[]>([])
    const selectedPackingListId = ref('')
    const filterLabelId = ref<string | null>(null)
    const originalDetails = ref<TripWithDetails | null>(null)

    const isEditMode = computed(() => originalDetails.value !== null)

    // ─── Reference data ──────────────────────────────────────────────────────
    const personsQuery = useQuery({ queryKey: ['persons'], queryFn: listPersons })
    const itemsQuery = useQuery({ queryKey: ['items'], queryFn: listItems })
    const packingListsQuery = useQuery({ queryKey: ['packing-lists'], queryFn: listPackingLists })

    const availablePersons = computed(() => personsQuery.data.value ?? [])
    const availableItems = computed(() => itemsQuery.data.value ?? [])
    const packingLists = computed(() => packingListsQuery.data.value ?? [])

    // Item → labels map powers the packing-list label panel and pool filtering.
    const itemLabelsQuery = useQuery({
        queryKey: computed(() => ['items-labels', availableItems.value.map((item) => item.id).sort().join(',')]),
        queryFn: async () => {
            const list = availableItems.value
            if (list.length === 0) {
                return [] as Array<{ itemId: string; labels: Label[] }>
            }
            const labelArrays = await Promise.all(
                list.map((item) => listItemLabels(item.id).catch(() => [] as Label[])),
            )
            return list.map((item, index) => ({ itemId: item.id, labels: labelArrays[index] }))
        },
        enabled: computed(() => availableItems.value.length > 0),
    })

    const itemLabelsByItemId = computed(() => {
        const map = new Map<string, Label[]>()
        for (const entry of itemLabelsQuery.data.value ?? []) {
            map.set(entry.itemId, entry.labels)
        }
        return map
    })

    const packingListLabelsQuery = useQuery({
        queryKey: computed(() => ['packing-list-labels', selectedPackingListId.value]),
        queryFn: () => listPackingListLabels(selectedPackingListId.value),
        enabled: computed(() => selectedPackingListId.value.length > 0),
    })

    const packingListLabels = computed(() => packingListLabelsQuery.data.value ?? [])

    const allPlacements = computed<PlannerPlacement[]>(() => {
        const result: PlannerPlacement[] = [...unassigned.value]
        for (const person of persons.value) {
            result.push(...person.worn)
            for (const pack of person.packs) {
                result.push(...pack.items)
            }
        }
        return result
    })

    // Label ids already represented by at least one staged item (grays them out).
    const coveredLabelIds = computed(() => {
        const covered = new Set<string>()
        const seenItems = new Set<string>()
        for (const placement of allPlacements.value) {
            if (seenItems.has(placement.itemId)) {
                continue
            }
            seenItems.add(placement.itemId)
            for (const label of itemLabelsByItemId.value.get(placement.itemId) ?? []) {
                covered.add(label.id)
            }
        }
        return covered
    })

    // ─── Stats ───────────────────────────────────────────────────────────────
    const stats = computed<PlannerStats>(() => {
        let itemCount = 0
        let totalWeightGrams = 0
        let totalValue = 0
        for (const placement of allPlacements.value) {
            itemCount += placement.quantity
            totalWeightGrams += weightOf(placement)
            totalValue += valueOf(placement)
        }
        return { itemCount, totalWeightGrams, totalValue }
    })

    const canProceedToStep2 = computed(
        () => details.name.trim().length > 0 && persons.value.length > 0,
    )

    const canSave = computed(() => canProceedToStep2.value)

    // ─── Pool mutations ────────────────────────────────────────────────────────
    function addPoolItem(itemId: string, quantity: number, carryStatus: CarryStatus): void {
        const item = availableItems.value.find((entry) => entry.id === itemId)
        if (!item || quantity < 1) {
            return
        }
        const existing = unassigned.value.find(
            (placement) => placement.itemId === itemId && placement.carryStatus === carryStatus,
        )
        if (existing) {
            existing.quantity += quantity
            return
        }
        unassigned.value.push({
            localId: nextLocalId('placement'),
            itemId,
            quantity,
            carryStatus,
            item,
        })
    }

    async function addSet(setId: string): Promise<void> {
        const setItems = await listSetItems(setId)
        for (const setItem of setItems) {
            const carryStatus = setItem.item.default_carry_status ?? DEFAULT_CARRY_STATUS
            addPoolItem(setItem.item_id, setItem.quantity, carryStatus)
        }
    }

    // Finds and removes a placement from whichever zone currently holds it.
    function removePlacement(placement: PlannerPlacement): void {
        const fromUnassigned = unassigned.value.findIndex((entry) => entry.localId === placement.localId)
        if (fromUnassigned >= 0) {
            unassigned.value.splice(fromUnassigned, 1)
            return
        }
        for (const person of persons.value) {
            const wornIndex = person.worn.findIndex((entry) => entry.localId === placement.localId)
            if (wornIndex >= 0) {
                person.worn.splice(wornIndex, 1)
                return
            }
            for (const pack of person.packs) {
                const packIndex = pack.items.findIndex((entry) => entry.localId === placement.localId)
                if (packIndex >= 0) {
                    pack.items.splice(packIndex, 1)
                    return
                }
            }
        }
    }

    function setPlacementQuantity(placement: PlannerPlacement, quantity: number): void {
        if (quantity < 1) {
            return
        }
        placement.quantity = quantity
    }

    // Splits one unit off a placement into a new sibling card within the same zone.
    function splitPlacement(placement: PlannerPlacement): void {
        if (placement.quantity < 2) {
            return
        }
        const sibling: PlannerPlacement = {
            localId: nextLocalId('placement'),
            itemId: placement.itemId,
            quantity: 1,
            carryStatus: placement.carryStatus,
            item: placement.item,
        }
        placement.quantity -= 1
        insertSiblingBeside(placement, sibling)
    }

    function insertSiblingBeside(reference: PlannerPlacement, sibling: PlannerPlacement): void {
        const zones: PlannerPlacement[][] = [unassigned.value]
        for (const person of persons.value) {
            zones.push(person.worn)
            for (const pack of person.packs) {
                zones.push(pack.items)
            }
        }
        for (const zone of zones) {
            const index = zone.findIndex((entry) => entry.localId === reference.localId)
            if (index >= 0) {
                zone.splice(index + 1, 0, sibling)
                return
            }
        }
    }

    // ─── People & packs ────────────────────────────────────────────────────────
    function isPersonSelected(personId: string): boolean {
        return persons.value.some((entry) => entry.personId === personId)
    }

    function togglePerson(person: Person): void {
        const index = persons.value.findIndex((entry) => entry.personId === person.id)
        if (index >= 0) {
            const [removed] = persons.value.splice(index, 1)
            // Return any placed items to the pool so nothing is silently lost.
            for (const placement of removed.worn) {
                unassigned.value.push(placement)
            }
            for (const pack of removed.packs) {
                for (const placement of pack.items) {
                    unassigned.value.push(placement)
                }
            }
            return
        }
        const firstPackLocalId = nextLocalId('pack')
        persons.value.push({
            personId: person.id,
            person,
            worn: [],
            packs: [{ localId: firstPackLocalId, name: 'Main Pack', items: [] }],
            mainPackLocalId: firstPackLocalId,
        })
    }

    function findPerson(personId: string): PlannerPerson | undefined {
        return persons.value.find((entry) => entry.personId === personId)
    }

    function addPack(personId: string, name: string): void {
        const person = findPerson(personId)
        if (!person) {
            return
        }
        const localId = nextLocalId('pack')
        person.packs.push({ localId, name: name.trim() || `Pack ${person.packs.length + 1}`, items: [] })
        if (person.mainPackLocalId === null) {
            person.mainPackLocalId = localId
        }
    }

    function removePack(personId: string, packLocalId: string): void {
        const person = findPerson(personId)
        if (!person) {
            return
        }
        const index = person.packs.findIndex((pack) => pack.localId === packLocalId)
        if (index < 0) {
            return
        }
        // Move the removed pack's items back to the pool.
        for (const placement of person.packs[index].items) {
            unassigned.value.push(placement)
        }
        person.packs.splice(index, 1)
        if (person.mainPackLocalId === packLocalId) {
            person.mainPackLocalId = person.packs[0]?.localId ?? null
        }
    }

    function renamePack(personId: string, packLocalId: string, name: string): void {
        const pack = findPerson(personId)?.packs.find((entry) => entry.localId === packLocalId)
        if (pack) {
            pack.name = name
        }
    }

    function setMainPack(personId: string, packLocalId: string): void {
        const person = findPerson(personId)
        if (person && person.packs.some((pack) => pack.localId === packLocalId)) {
            person.mainPackLocalId = packLocalId
        }
    }

    // ─── Auto-assignment (drop onto a person) ──────────────────────────────────
    function mainPackOf(person: PlannerPerson) {
        return person.packs.find((pack) => pack.localId === person.mainPackLocalId) ?? person.packs[0]
    }

    function autoAssignToPerson(personId: string, placement: PlannerPlacement): void {
        const person = findPerson(personId)
        if (!person) {
            return
        }
        removePlacement(placement)
        const mainPack = mainPackOf(person)

        if (placement.carryStatus === 'worn') {
            // Worn with extra quantity: one unit packed in the main pack, the rest worn.
            if (placement.quantity > 1 && mainPack) {
                mainPack.items.push({
                    localId: nextLocalId('placement'),
                    itemId: placement.itemId,
                    quantity: 1,
                    carryStatus: 'packed',
                    item: placement.item,
                })
                placement.quantity -= 1
            }
            person.worn.push(placement)
            return
        }

        if (mainPack) {
            mainPack.items.push(placement)
            return
        }
        // No pack available: fall back to worn so the item is not lost.
        person.worn.push(placement)
    }

    // Assign the given placements to a person; defaults to every unassigned item.
    function assignToPerson(personId: string, placements?: PlannerPlacement[]): void {
        if (!findPerson(personId)) {
            return
        }
        // Snapshot first: autoAssignToPerson mutates unassigned as it goes.
        for (const placement of [...(placements ?? unassigned.value)]) {
            autoAssignToPerson(personId, placement)
        }
    }

    // ─── Per-person stats ──────────────────────────────────────────────────────
    function personStats(person: PlannerPerson): PlannerPersonStats {
        let packedWeightGrams = 0
        let wornWeightGrams = 0
        let totalValue = 0
        let itemCount = 0

        for (const placement of person.worn) {
            wornWeightGrams += weightOf(placement)
            totalValue += valueOf(placement)
            itemCount += placement.quantity
        }
        for (const pack of person.packs) {
            for (const placement of pack.items) {
                packedWeightGrams += weightOf(placement)
                totalValue += valueOf(placement)
                itemCount += placement.quantity
            }
        }

        const recommendedMaxGrams = getPersonRecommendedMaxWeightGrams(person.person)
        let weightStatus: PlannerPersonStats['weightStatus'] = 'unknown'
        if (recommendedMaxGrams > 0) {
            if (packedWeightGrams <= recommendedMaxGrams * WEIGHT_WARN_RATIO) {
                weightStatus = 'ok'
            } else if (packedWeightGrams <= recommendedMaxGrams * WEIGHT_OVER_RATIO) {
                weightStatus = 'warn'
            } else {
                weightStatus = 'over'
            }
        }

        return { recommendedMaxGrams, packedWeightGrams, wornWeightGrams, totalValue, itemCount, weightStatus }
    }

    // ─── Persistence mapping ───────────────────────────────────────────────────
    function buildStagedTrip(): StagedTrip {
        const stagedPersons: StagedPerson[] = persons.value.map((person) => ({
            personId: person.personId,
            person: person.person,
            tripPersonId: person.tripPersonId,
            packs: person.packs.map((pack) => ({
                tempId: pack.localId,
                packId: pack.packId,
                name: pack.name,
                items: aggregateStagedItems(pack.items, 'packed'),
            })),
            items: aggregateStagedItems(person.worn, 'worn'),
        }))

        return {
            name: details.name,
            tripType: details.tripType,
            notes: details.notes,
            durationDays: details.durationDays,
            distanceKm: details.distanceKm,
            routeService: details.routeService,
            routeUrl: details.routeUrl,
            persons: stagedPersons,
        }
    }

    function reset(): void {
        Object.assign(details, emptyDetails())
        persons.value = []
        unassigned.value = []
        selectedPackingListId.value = ''
        filterLabelId.value = null
        originalDetails.value = null
        step.value = 1
    }

    async function loadExisting(tripId: string): Promise<void> {
        isLoading.value = true
        try {
            const trip = await getTrip(tripId)
            originalDetails.value = trip

            const days = intervalToDays(trip.duration)
            const routeService = getTripRouteService(trip)
            Object.assign(details, {
                name: trip.name,
                tripType: trip.trip_type,
                notes: trip.notes ?? '',
                durationDays: days === null ? '' : String(days),
                distanceKm: typeof trip.total_distance_km === 'number' ? String(trip.total_distance_km) : '',
                routeService: routeService === 'unknown' ? 'komoot' : routeService,
                routeUrl: getTripRouteUrl(trip),
            } satisfies PlannerDetails)

            persons.value = trip.persons.map((tripPerson) => {
                const packs = tripPerson.packs.map((pack) => ({
                    localId: nextLocalId('pack'),
                    name: pack.pack.name,
                    packId: pack.pack_id,
                    items: pack.items.map(placementFromEntry),
                }))
                return {
                    personId: tripPerson.person_id,
                    person: tripPerson.person,
                    tripPersonId: tripPerson.trip_person_id,
                    worn: tripPerson.items.map(placementFromEntry),
                    packs,
                    mainPackLocalId: packs[0]?.localId ?? null,
                } satisfies PlannerPerson
            })
            unassigned.value = []
        } finally {
            isLoading.value = false
        }
    }

    return {
        step,
        isEditMode,
        isLoading,
        details,
        persons,
        unassigned,
        selectedPackingListId,
        filterLabelId,
        originalDetails,
        availablePersons,
        availableItems,
        packingLists,
        packingListLabels,
        itemLabelsByItemId,
        coveredLabelIds,
        stats,
        canProceedToStep2,
        canSave,
        addPoolItem,
        addSet,
        removePlacement,
        setPlacementQuantity,
        splitPlacement,
        isPersonSelected,
        togglePerson,
        addPack,
        removePack,
        renamePack,
        setMainPack,
        autoAssignToPerson,
        assignToPerson,
        personStats,
        buildStagedTrip,
        loadExisting,
        reset,
    }
}

// Creates the planner context and provides it to descendant components.
export function provideTripPlanner(): TripPlannerContext {
    const context = createTripPlannerContext()
    provide(TRIP_PLANNER_KEY, context)
    return context
}

// Injects the planner context provided by an ancestor TripPlannerPage.
export function useTripPlanner(): TripPlannerContext {
    const context = inject(TRIP_PLANNER_KEY)
    if (!context) {
        throw new Error('useTripPlanner must be used within a TripPlannerPage provider')
    }
    return context
}
