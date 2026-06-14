import { http, HttpResponse } from 'msw'
import type { components } from '../../../lib/api/schema'

type TokenResponse = components['schemas']['TokenResponse']

const token: TokenResponse = { access_token: 'test-token', token_type: 'Bearer', expires_in: 900 }

// Defaults model a logged-OUT visitor: refresh fails (no cookie) so the auth
// store bootstraps to unauthenticated. Tests that need a session override
// `*/api/v1/auth/refresh` with a 200. Login defaults to success.
export const authHandlers = [
  http.post('*/api/v1/auth/login', () => HttpResponse.json<TokenResponse>(token)),
  http.post('*/api/v1/auth/refresh', () => new HttpResponse(null, { status: 401 })),
  http.post('*/api/v1/auth/logout', () => new HttpResponse(null, { status: 204 })),
]
