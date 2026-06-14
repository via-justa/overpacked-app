import { http, HttpResponse } from 'msw'
import type { components } from '../../../lib/api/schema'

type PackingList = components['schemas']['PackingList']
type Label = components['schemas']['Label']

// Bare bodies (no envelope). Override per-test with server.use(...).
export const listsHandlers = [
  http.get('*/api/v1/packing-lists', () => HttpResponse.json<PackingList[]>([])),
  http.get('*/api/v1/packing-lists/:listId/labels', () => HttpResponse.json<Label[]>([])),
]
