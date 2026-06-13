const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

type ApiErrorPayload = {
  error?: string
}

type LoginRequest = {
  username: string
  password: string
}

export type TokenResponse = {
  access_token: string
  token_type: string
  expires_in: number
}

const buildUrl = (path: string) => {
  if (API_BASE_URL.endsWith('/')) {
    return `${API_BASE_URL.slice(0, -1)}${path}`
  }

  return `${API_BASE_URL}${path}`
}

const extractErrorMessage = async (response: Response) => {
  try {
    const payload = (await response.json()) as ApiErrorPayload
    return payload.error ?? `Request failed with status ${response.status}`
  } catch {
    return `Request failed with status ${response.status}`
  }
}

// credentials: 'include' ensures the HttpOnly refresh cookie is stored on login
// and sent back on refresh/logout (the refresh token is never held in JS).
export const loginAuth = async (payload: LoginRequest): Promise<TokenResponse> => {
  const response = await fetch(buildUrl('/api/v1/auth/login'), {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    throw new Error(await extractErrorMessage(response))
  }

  return (await response.json()) as TokenResponse
}

// refreshAuth carries no body: the refresh token travels in the HttpOnly cookie.
export const refreshAuth = async (): Promise<TokenResponse> => {
  const response = await fetch(buildUrl('/api/v1/auth/refresh'), {
    method: 'POST',
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error(await extractErrorMessage(response))
  }

  return (await response.json()) as TokenResponse
}

export const logoutAuth = async (accessToken: string) => {
  const response = await fetch(buildUrl('/api/v1/auth/logout'), {
    method: 'POST',
    credentials: 'include',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })

  if (!response.ok && response.status !== 401) {
    throw new Error(await extractErrorMessage(response))
  }
}
