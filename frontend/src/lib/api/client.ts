import createClient from 'openapi-fetch'
import type { paths } from './schema'

let authToken: string | null = null
let isRefreshing = false
let refreshPromise: Promise<string | null> | null = null
let refreshTokenHandler: (() => Promise<string | null>) | null = null
let authFailureHandler: ((reason: 'session_expired') => void) | null = null

const RETRY_HEADER = 'x-packing-auth-retry'

// Base URL shared with manual fetch calls (binary downloads, multipart uploads)
// that can't go through the typed openapi-fetch client.
export const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? ''

export const setApiAuthToken = (token: string | null) => {
  authToken = token
}

// getApiAuthToken exposes the current bearer token for manual fetch calls.
export const getApiAuthToken = () => authToken

export const setApiAuthRefreshHandler = (handler: (() => Promise<string | null>) | null) => {
  refreshTokenHandler = handler
}

export const setApiAuthFailureHandler = (handler: ((reason: 'session_expired') => void) | null) => {
  authFailureHandler = handler
}

export const apiClient = createClient<paths>({
  baseUrl: apiBaseUrl,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.use({
  onRequest({ request }) {
    if (authToken) {
      request.headers.set('Authorization', `Bearer ${authToken}`)
    }

    return request
  },
  async onResponse({ request, response }) {
    const isUnauthorized = response.status === 401
    const isRetry = request.headers.get(RETRY_HEADER) === '1'
    const isAuthEndpoint = request.url.includes('/api/v1/auth/')

    if (!isUnauthorized || isRetry || isAuthEndpoint || !refreshTokenHandler) {
      return response
    }

    if (isRefreshing && refreshPromise !== null) {
      const tokenFromQueue = await refreshPromise
      if (!tokenFromQueue) {
        authFailureHandler?.('session_expired')
        return response
      }
    } else {
      isRefreshing = true
      refreshPromise = refreshTokenHandler()
      const refreshedToken = await refreshPromise
      isRefreshing = false
      refreshPromise = null

      if (!refreshedToken) {
        authFailureHandler?.('session_expired')
        return response
      }
    }

    try {
      const retryRequest = request.clone()
      retryRequest.headers.set(RETRY_HEADER, '1')

      if (authToken) {
        retryRequest.headers.set('Authorization', `Bearer ${authToken}`)
      }

      return fetch(retryRequest)
    } catch {
      return response
    }
  },
})
