import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { refreshAuth, type TokenResponse } from '../lib/api/auth'
import { setApiAuthRefreshHandler, setApiAuthToken } from '../lib/api/client'

const ACCESS_TOKEN_KEY = 'overpacked-app.auth.accessToken'
const REFRESH_TOKEN_KEY = 'overpacked-app.auth.refreshToken'
const EXPIRES_AT_KEY = 'overpacked-app.auth.expiresAt'
const REFRESH_SKEW_MS = 15_000

export type AuthNotice = 'session_expired' | null

type AuthSession = {
  accessToken: string
  refreshToken: string
  expiresAt: number
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string | null>(localStorage.getItem(ACCESS_TOKEN_KEY))
  const refreshToken = ref<string | null>(localStorage.getItem(REFRESH_TOKEN_KEY))
  const expiresAt = ref<number>(Number(localStorage.getItem(EXPIRES_AT_KEY)) || 0)
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
      refreshToken: tokens.refresh_token,
      expiresAt: Date.now() + tokens.expires_in * 1000,
    }
  }

  const setSession = (session: AuthSession | null) => {
    accessToken.value = session?.accessToken ?? null
    refreshToken.value = session?.refreshToken ?? null
    expiresAt.value = session?.expiresAt ?? 0
    setApiAuthToken(accessToken.value)

    if (session) {
      localStorage.setItem(ACCESS_TOKEN_KEY, session.accessToken)
      localStorage.setItem(REFRESH_TOKEN_KEY, session.refreshToken)
      localStorage.setItem(EXPIRES_AT_KEY, String(session.expiresAt))
      return
    }

    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    localStorage.removeItem(EXPIRES_AT_KEY)
  }

  const setSessionFromTokens = (tokens: TokenResponse) => {
    setSession(toSession(tokens))
  }

  const refreshAccessToken = async (): Promise<string | null> => {
    if (!refreshToken.value) {
      clearSession()
      return null
    }

    try {
      const refreshed = await refreshAuth({
        refresh_token: refreshToken.value,
      })
      setSessionFromTokens(refreshed)
      return refreshed.access_token
    } catch {
      setAuthNotice('session_expired')
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
        if (!refreshToken.value) {
          clearSession()
          return
        }

        const hasValidAccessToken = Boolean(accessToken.value) && Date.now() < expiresAt.value - REFRESH_SKEW_MS

        if (hasValidAccessToken) {
          setApiAuthToken(accessToken.value)
          return
        }

        await refreshAccessToken()
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
    refreshToken,
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
