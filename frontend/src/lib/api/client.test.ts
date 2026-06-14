import { http, HttpResponse } from 'msw'
import {
  apiClient,
  getApiAuthToken,
  setApiAuthFailureHandler,
  setApiAuthRefreshHandler,
  setApiAuthToken,
} from './client'
import { server } from '../../test/msw/server'

afterEach(() => {
  setApiAuthToken(null)
  setApiAuthRefreshHandler(null)
  setApiAuthFailureHandler(null)
})

describe('apiClient auth middleware', () => {
  it('attaches the bearer token to requests', async () => {
    setApiAuthToken('tok-123')
    expect(getApiAuthToken()).toBe('tok-123')

    let seenAuth: string | null = null
    server.use(
      http.get('*/api/v1/items', ({ request }) => {
        seenAuth = request.headers.get('Authorization')
        return HttpResponse.json([])
      }),
    )

    await apiClient.GET('/api/v1/items')
    expect(seenAuth).toBe('Bearer tok-123')
  })

  it('refreshes the token on 401 and retries the request once', async () => {
    setApiAuthRefreshHandler(async () => {
      setApiAuthToken('refreshed')
      return 'refreshed'
    })

    server.use(
      http.get('*/api/v1/items', ({ request }) => {
        if (request.headers.get('x-packing-auth-retry') === '1') {
          return HttpResponse.json([{ id: 'i-1' }])
        }
        return new HttpResponse(null, { status: 401 })
      }),
    )

    const { data, response } = await apiClient.GET('/api/v1/items')
    expect(response.status).toBe(200)
    expect(data).toEqual([{ id: 'i-1' }])
  })

  it('invokes the failure handler when the refresh yields no token', async () => {
    const onFailure = vi.fn()
    setApiAuthFailureHandler(onFailure)
    setApiAuthRefreshHandler(async () => null)

    server.use(http.get('*/api/v1/items', () => new HttpResponse(null, { status: 401 })))

    const { response } = await apiClient.GET('/api/v1/items')
    expect(response.status).toBe(401)
    expect(onFailure).toHaveBeenCalledWith('session_expired')
  })

  it('does not attempt a refresh when no refresh handler is registered', async () => {
    server.use(http.get('*/api/v1/items', () => new HttpResponse(null, { status: 401 })))
    const { response } = await apiClient.GET('/api/v1/items')
    expect(response.status).toBe(401)
  })

  it('never refreshes on the auth endpoints themselves', async () => {
    const refresh = vi.fn(async () => 'tok')
    setApiAuthRefreshHandler(refresh)
    server.use(http.post('*/api/v1/auth/login', () => new HttpResponse(null, { status: 401 })))

    const { response } = await apiClient.POST('/api/v1/auth/login', {
      body: { username: 'a', password: 'b' },
    })
    expect(response.status).toBe(401)
    expect(refresh).not.toHaveBeenCalled()
  })
})
