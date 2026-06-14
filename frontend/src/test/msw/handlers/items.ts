import { http, HttpResponse } from 'msw'
import type { components } from '../../../lib/api/schema'

type Item = components['schemas']['Item']
type Label = components['schemas']['Label']

// Bare array bodies (no envelope). Override per-test with server.use(...).
export const itemsHandlers = [
  http.get('*/api/v1/items', () => HttpResponse.json<Item[]>([])),
  http.get('*/api/v1/labels', () => HttpResponse.json<Label[]>([])),
  http.get('*/api/v1/items/:itemId/labels', () => HttpResponse.json<Label[]>([])),
]
