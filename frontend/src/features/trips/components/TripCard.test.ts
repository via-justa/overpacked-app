import { http, HttpResponse } from 'msw'
import { fireEvent } from '@testing-library/vue'
import { renderWithProviders } from '../../../test/renderWithProviders'
import { server } from '../../../test/msw/server'
import TripCard from './TripCard.vue'
import type { Trip, TripStats } from '../types'

const baseTrip = (over: Partial<Trip> = {}): Trip =>
  ({
    id: 't-1',
    name: 'weekend hike',
    trip_type: 'day_hike',
    duration: '3 days',
    total_distance_km: 30,
    notes: '',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...over,
  }) as unknown as Trip

const stats: TripStats = {
  totalWeightGrams: 1500,
  packedWeightGrams: 1000,
  wornWeightGrams: 500,
  totalValue: 100,
  packsCount: 2,
  travelersCount: 3,
}

describe('TripCard', () => {
  it('renders derived trip labels (title-cased name, type, distance, avg/day)', () => {
    const { getByText, getByRole } = renderWithProviders(TripCard, {
      props: { trip: baseTrip(), stats },
    })
    expect(getByText('Weekend Hike')).toBeInTheDocument() // normalizeTitleWords
    // The type/duration/distance line joins to one text run (the "/" are in spans).
    expect(getByText(/Day Hike\s+3 days\s+30 km\s+10 km\/day/)).toBeInTheDocument()
    expect(getByRole('button', { name: 'Delete trip' })).toBeInTheDocument()
  })

  it('emits edit when the card is activated and delete from the delete button', async () => {
    const { getByText, getByRole, emitted } = renderWithProviders(TripCard, {
      props: { trip: baseTrip(), stats },
    })

    await fireEvent.click(getByText('Weekend Hike').closest('button') as HTMLButtonElement)
    await fireEvent.click(getByRole('button', { name: 'Delete trip' }))

    const events = emitted() as Record<string, unknown[][]>
    expect(events.edit[0][0]).toMatchObject({ id: 't-1' })
    expect(events.delete[0][0]).toMatchObject({ id: 't-1' })
  })

  describe('route link (stored-XSS guard)', () => {
    // The route service triggers an OG-preview fetch; stub it.
    beforeEach(() =>
      server.use(
        http.get('*/api/v1/trips/route-preview/:service', () => HttpResponse.json({ image_url: null })),
      ),
    )

    it('renders the link for a safe http(s) URL', () => {
      const { getByRole } = renderWithProviders(TripCard, {
        props: { trip: baseTrip({ trip_strava_url: 'https://strava.com/activities/1' }), stats },
      })
      expect(getByRole('link', { name: /route/i })).toHaveAttribute(
        'href',
        'https://strava.com/activities/1',
      )
    })

    it('suppresses the link for an unsafe (javascript:) URL', () => {
      const { queryByRole } = renderWithProviders(TripCard, {
        props: { trip: baseTrip({ trip_komoot_url: 'javascript:alert(1)' }), stats },
      })
      expect(queryByRole('link', { name: /route/i })).toBeNull()
    })
  })
})
