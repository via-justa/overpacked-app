import { http, HttpResponse } from 'msw'
import type { components } from '../../../lib/api/schema'
import { settingsFixture } from '../../fixtures'

type Settings = components['schemas']['Settings']

// Endpoints return the bare schema body (no { data } envelope) — see
// src/lib/api/request.ts (unwrapApiResponse returns openapi-fetch's `data`).
export const settingsHandlers = [
  http.get('*/api/v1/settings', () => HttpResponse.json<Settings>(settingsFixture())),
]
