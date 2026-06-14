import { apiClient } from '../../../lib/api/client'
import { ensureApiResponse, unwrapApiResponse } from '../../../lib/api/request'
import type { Settings, SettingsUpdate } from '../types'

export const getSettings = async (): Promise<Settings> =>
  unwrapApiResponse(apiClient.GET('/api/v1/settings'), 'Unable to load settings')

export const patchSettings = async (payload: SettingsUpdate): Promise<Settings> =>
  unwrapApiResponse(apiClient.PATCH('/api/v1/settings', { body: payload }), 'Unable to save settings')

export const startFresh = async (password: string): Promise<void> =>
  ensureApiResponse(
    apiClient.POST('/api/v1/settings/start-fresh', { body: { password } }),
    'Unable to reset data',
  )
