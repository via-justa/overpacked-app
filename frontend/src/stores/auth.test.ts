import { createPinia, setActivePinia } from 'pinia'
import { http, HttpResponse } from 'msw'
import { useAuthStore } from './auth'
import { server } from '../test/msw/server'

const tokenResponse = { access_token: 'fresh', token_type: 'Bearer', expires_in: 900 }

const refresh = (status: number) =>
  server.use(
    http.post('*/api/v1/auth/refresh', () =>
      status === 200 ? HttpResponse.json(tokenResponse) : new HttpResponse(null, { status }),
    ),
  )

beforeEach(() => setActivePinia(createPinia()))

describe('auth store', () => {
  it('bootstraps a session when the refresh cookie is valid (silent, no notice)', async () => {
    refresh(200)
    const store = useAuthStore()
    await store.ensureBootstrapped()

    expect(store.isAuthenticated).toBe(true)
    expect(store.authNotice).toBeNull()
    expect(store.isBootstrapping).toBe(false)
  })

  it('bootstraps logged-out without a notice when there is no valid cookie', async () => {
    refresh(401) // default is already 401, set explicitly for clarity
    const store = useAuthStore()
    await store.ensureBootstrapped()

    expect(store.isAuthenticated).toBe(false)
    expect(store.authNotice).toBeNull() // silent: no spurious "session expired"
  })

  it('memoizes the bootstrap so concurrent calls share one refresh', async () => {
    refresh(200)
    const store = useAuthStore()
    await Promise.all([store.ensureBootstrapped(), store.ensureBootstrapped()])
    expect(store.isAuthenticated).toBe(true)
  })

  it('skips refresh when a valid access token is already held', async () => {
    const store = useAuthStore()
    store.setSessionFromTokens(tokenResponse)
    expect(store.isAuthenticated).toBe(true)

    await store.ensureBootstrapped() // hasValidAccessToken === true → no refresh
    expect(store.isAuthenticated).toBe(true)
  })

  it('raises a session_expired notice on a non-silent refresh failure', async () => {
    refresh(401)
    const store = useAuthStore()
    const token = await store.refreshAccessToken() // not silent
    expect(token).toBeNull()
    expect(store.authNotice).toBe('session_expired')
    expect(store.isAuthenticated).toBe(false)
  })

  it('clears the session and consumes the notice', () => {
    const store = useAuthStore()
    store.setSessionFromTokens(tokenResponse)
    store.setAuthNotice('session_expired')

    expect(store.consumeAuthNotice()).toBe('session_expired')
    expect(store.authNotice).toBeNull()

    store.clearSession()
    expect(store.isAuthenticated).toBe(false)
    expect(store.accessToken).toBeNull()
  })
})
