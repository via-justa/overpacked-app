import { http, HttpResponse } from 'msw'
import type { components } from '../../../lib/api/schema'

type Person = components['schemas']['Person']

// Bare array body (no envelope). Override per-test with server.use(...).
export const personsHandlers = [
  http.get('*/api/v1/persons', () => HttpResponse.json<Person[]>([])),
]
