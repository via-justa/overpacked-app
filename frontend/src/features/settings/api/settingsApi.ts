import { apiClient } from '../../../lib/api/client'
import type { Settings, SettingsUpdate } from '../types'

const getErrorMessage = (error: unknown, fallback: string) => {
  if (error && typeof error === 'object' && 'error' in error) {
    const value = (error as { error?: unknown }).error
    if (typeof value === 'string' && value.length > 0) {
      return value
    }
  }

  return fallback
}

export const getSettings = async (): Promise<Settings> => {
  const { data, error, response } = await apiClient.GET('/api/v1/settings')

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to load settings'))
  }

  return data as Settings
}

export const patchSettings = async (payload: SettingsUpdate): Promise<Settings> => {
  const { data, error, response } = await apiClient.PATCH('/api/v1/settings', {
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to save settings'))
  }

  return data as Settings
}

export const startFresh = async (password: string): Promise<void> => {
  const { error, response } = await apiClient.POST('/api/v1/settings/start-fresh', {
    body: { password },
  })

  if (!response.ok) {
    throw new Error(getErrorMessage(error, 'Unable to reset data'))
  }
}
