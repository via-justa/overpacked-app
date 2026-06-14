import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { refreshAuth, type TokenResponse } from '../lib/api/auth'
import { setApiAuthRefreshHandler, setApiAuthToken } from '../lib/api/client'

// Legacy keys from when tokens were persisted in localStorage. The refresh
// token now lives in an HttpOnly cookie and the access token is kept in memory
// only, so any stale values left by older builds are cleared on startup.
const LEGACY_STORAGE_KEYS = [
  'overpacked-app.auth.accessToken',
  'overpacked-app.auth.refreshToken',
  'overpacked-app.auth.expiresAt',
]
const REFRESH_SKEW_MS = 15_000

export type AuthNotice = 'session_expired' | null

type AuthSession = {
  accessToken: string
  expiresAt: number
}

export const useAuthStore = defineStore('auth', () => {
  for (const key of LEGACY_STORAGE_KEYS) {
    localStorage.removeItem(key)
  }

  // Held in memory only: a stolen token can no longer be read from storage, and
  // the session is re-established after reload via the HttpOnly refresh cookie.
  const accessToken = ref<string | null>(null)
  const expiresAt = ref<number>(0)
  const isBootstrapping = ref(true)
  const authNotice = ref<AuthNotice>(null)
  let bootstrapPromise: Promise<void> | null = null

  const isAuthenticated = computed(() => {
    return Boolean(accessToken.value) && Date.now() < expiresAt.value - REFRESH_SKEW_MS
  })

  setApiAuthToken(accessToken.value)
  setApiAuthRefreshHandler(async () => {
    return refreshAccessToken()
  })

  const toSession = (tokens: TokenResponse): AuthSession => {
    return {
      accessToken: tokens.access_token,
      expiresAt: Date.now() + tokens.expires_in * 1000,
    }
  }

  const setSession = (session: AuthSession | null) => {
    accessToken.value = session?.accessToken ?? null
    expiresAt.value = session?.expiresAt ?? 0
    setApiAuthToken(accessToken.value)
  }

  const setSessionFromTokens = (tokens: TokenResponse) => {
    setSession(toSession(tokens))
  }

  // refreshAccessToken mints a new access token from the HttpOnly refresh cookie.
  // Pass { silent: true } during startup so a visitor who was never logged in
  // (no cookie) lands on login without a spurious "session expired" notice.
  const refreshAccessToken = async (options?: { silent?: boolean }): Promise<string | null> => {
    try {
      const refreshed = await refreshAuth()
      setSessionFromTokens(refreshed)
      return refreshed.access_token
    } catch {
      if (!options?.silent) {
        setAuthNotice('session_expired')
      }
      clearSession()
      return null
    }
  }

  const setAuthNotice = (notice: AuthNotice) => {
    authNotice.value = notice
  }

  const consumeAuthNotice = () => {
    const notice = authNotice.value
    authNotice.value = null
    return notice
  }

  const ensureBootstrapped = async () => {
    if (bootstrapPromise !== null) {
      return bootstrapPromise
    }

    bootstrapPromise = (async () => {
      try {
        const hasValidAccessToken = Boolean(accessToken.value) && Date.now() < expiresAt.value - REFRESH_SKEW_MS

        if (hasValidAccessToken) {
          setApiAuthToken(accessToken.value)
          return
        }

        await refreshAccessToken({ silent: true })
      } finally {
        isBootstrapping.value = false
      }
    })()

    await bootstrapPromise
  }

  const clearSession = () => {
    setSession(null)
  }

  return {
    accessToken,
    expiresAt,
    isAuthenticated,
    isBootstrapping,
    authNotice,
    setSession,
    setSessionFromTokens,
    setAuthNotice,
    consumeAuthNotice,
    refreshAccessToken,
    ensureBootstrapped,
    clearSession,
  }
})
