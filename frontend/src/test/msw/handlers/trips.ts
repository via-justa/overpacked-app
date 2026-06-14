import { http, HttpResponse } from 'msw'
import type { components } from '../../../lib/api/schema'

type Trip = components['schemas']['Trip']

// Bare array body (no envelope). Override per-test with server.use(...);
// detail (`/trips/:id`) intentionally has no default so tests must opt in.
export const tripsHandlers = [
  http.get('*/api/v1/trips', () => HttpResponse.json<Trip[]>([])),
]
