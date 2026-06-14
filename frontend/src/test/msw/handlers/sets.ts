import { http, HttpResponse } from 'msw'
import type { components } from '../../../lib/api/schema'

type ItemSet = components['schemas']['ItemSet']
type SetItemWithDetails = components['schemas']['SetItemWithDetails']

// Bare array bodies (no envelope). Override per-test with server.use(...).
export const setsHandlers = [
  http.get('*/api/v1/sets', () => HttpResponse.json<ItemSet[]>([])),
  http.get('*/api/v1/sets/:setId/items', () => HttpResponse.json<SetItemWithDetails[]>([])),
]
