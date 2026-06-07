import createClient from 'openapi-fetch'

type ApiPaths = {
  '/api/v1/settings': {
    get: {
      responses: {
        200: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
    patch: {
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
  }
  '/api/v1/settings/start-fresh': {
    post: {
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        204: {
          content: never
        }
      }
    }
  }
  '/api/v1/persons': {
    get: {
      responses: {
        200: {
          content: {
            'application/json': Array<Record<string, unknown>>
          }
        }
      }
    }
    post: {
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        201: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
  }
  '/api/v1/persons/{personId}': {
    patch: {
      parameters: {
        path: {
          personId: string
        }
      }
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
    delete: {
      parameters: {
        path: {
          personId: string
        }
      }
      responses: {
        204: {
          content: never
        }
      }
    }
  }
  '/api/v1/manufacturers': {
    get: {
      responses: {
        200: {
          content: {
            'application/json': Array<Record<string, unknown>>
          }
        }
      }
    }
    post: {
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        201: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
  }
  '/api/v1/manufacturers/{manufacturerId}': {
    delete: {
      parameters: {
        path: {
          manufacturerId: string
        }
      }
      responses: {
        204: {
          content: never
        }
      }
    }
    patch: {
      parameters: {
        path: {
          manufacturerId: string
        }
      }
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
  }
  '/api/v1/search': {
    get: {
      parameters: {
        query: {
          q: string
          types?: string[]
          limit?: number
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Array<Record<string, unknown>>
          }
        }
      }
    }
  }
  '/api/v1/items': {
    get: {
      responses: {
        200: {
          content: {
            'application/json': Array<Record<string, unknown>>
          }
        }
      }
    }
    post: {
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        201: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
  }
  '/api/v1/items/{itemId}': {
    patch: {
      parameters: {
        path: {
          itemId: string
        }
      }
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
    delete: {
      parameters: {
        path: {
          itemId: string
        }
      }
      responses: {
        204: {
          content: never
        }
      }
    }
  }
  '/api/v1/labels': {
    get: {
      responses: {
        200: {
          content: {
            'application/json': Array<Record<string, unknown>>
          }
        }
      }
    }
    post: {
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        201: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
  }
  '/api/v1/labels/{labelId}': {
    get: {
      parameters: {
        path: {
          labelId: string
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
    patch: {
      parameters: {
        path: {
          labelId: string
        }
      }
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
    delete: {
      parameters: {
        path: {
          labelId: string
        }
      }
      responses: {
        204: {
          content: never
        }
      }
    }
  }
  '/api/v1/items/{itemId}/labels': {
    get: {
      parameters: {
        path: {
          itemId: string
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Array<Record<string, unknown>>
          }
        }
      }
    }
    post: {
      parameters: {
        path: {
          itemId: string
        }
      }
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        201: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
  }
  '/api/v1/items/{itemId}/labels/{labelId}': {
    delete: {
      parameters: {
        path: {
          itemId: string
          labelId: string
        }
      }
      responses: {
        204: {
          content: never
        }
      }
    }
  }
  '/api/v1/item-types': {
    get: {
      responses: {
        200: {
          content: {
            'application/json': Array<Record<string, unknown>>
          }
        }
      }
    }
    post: {
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        201: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
  }
  '/api/v1/item-types/{typeId}': {
    delete: {
      parameters: {
        path: {
          typeId: string
        }
      }
      responses: {
        204: {
          content: never
        }
      }
    }
    patch: {
      parameters: {
        path: {
          typeId: string
        }
      }
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
  }
  '/api/v1/item-types/{typeId}/fields': {
    get: {
      parameters: {
        path: {
          typeId: string
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Array<Record<string, unknown>>
          }
        }
      }
    }
    put: {
      parameters: {
        path: {
          typeId: string
        }
      }
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Array<Record<string, unknown>>
          }
        }
      }
    }
  }
  '/api/v1/sets': {
    get: {
      responses: {
        200: {
          content: {
            'application/json': Array<Record<string, unknown>>
          }
        }
      }
    }
    post: {
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        201: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
  }
  '/api/v1/sets/{setId}': {
    get: {
      parameters: {
        path: {
          setId: string
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
    patch: {
      parameters: {
        path: {
          setId: string
        }
      }
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
    delete: {
      parameters: {
        path: {
          setId: string
        }
      }
      responses: {
        204: {
          content: never
        }
      }
    }
  }
  '/api/v1/sets/{setId}/items': {
    get: {
      parameters: {
        path: {
          setId: string
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Array<Record<string, unknown>>
          }
        }
      }
    }
    post: {
      parameters: {
        path: {
          setId: string
        }
      }
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        201: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
  }
  '/api/v1/sets/{setId}/items/{itemId}': {
    patch: {
      parameters: {
        path: {
          setId: string
          itemId: string
        }
      }
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
    delete: {
      parameters: {
        path: {
          setId: string
          itemId: string
        }
      }
      responses: {
        204: {
          content: never
        }
      }
    }
  }
  '/api/v1/packing-lists': {
    get: {
      responses: {
        200: {
          content: {
            'application/json': Array<Record<string, unknown>>
          }
        }
      }
    }
    post: {
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        201: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
  }
  '/api/v1/packing-lists/{listId}': {
    get: {
      parameters: {
        path: {
          listId: string
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
    patch: {
      parameters: {
        path: {
          listId: string
        }
      }
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
    delete: {
      parameters: {
        path: {
          listId: string
        }
      }
      responses: {
        204: {
          content: never
        }
      }
    }
  }
  '/api/v1/packing-lists/{listId}/labels': {
    get: {
      parameters: {
        path: {
          listId: string
        }
      }
      responses: {
        200: {
          content: {
            'application/json': Array<Record<string, unknown>>
          }
        }
      }
    }
    post: {
      parameters: {
        path: {
          listId: string
        }
      }
      requestBody: {
        content: {
          'application/json': Record<string, unknown>
        }
      }
      responses: {
        201: {
          content: {
            'application/json': Record<string, unknown>
          }
        }
      }
    }
  }
  '/api/v1/packing-lists/{listId}/labels/{labelId}': {
    delete: {
      parameters: {
        path: {
          listId: string
          labelId: string
        }
      }
      responses: {
        204: {
          content: never
        }
      }
    }
  }
}

let authToken: string | null = null
let isRefreshing = false
let refreshPromise: Promise<string | null> | null = null
let refreshTokenHandler: (() => Promise<string | null>) | null = null
let authFailureHandler: ((reason: 'session_expired') => void) | null = null

const RETRY_HEADER = 'x-packing-auth-retry'

export const setApiAuthToken = (token: string | null) => {
  authToken = token
}

export const setApiAuthRefreshHandler = (handler: (() => Promise<string | null>) | null) => {
  refreshTokenHandler = handler
}

export const setApiAuthFailureHandler = (handler: ((reason: 'session_expired') => void) | null) => {
  authFailureHandler = handler
}

export const apiClient = createClient<ApiPaths>({
  baseUrl: import.meta.env.VITE_API_BASE_URL ?? '',
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