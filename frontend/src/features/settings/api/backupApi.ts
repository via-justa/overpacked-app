import { apiBaseUrl, apiClient, getApiAuthToken } from '../../../lib/api/client'
import { unwrapApiResponse } from '../../../lib/api/request'
import type {
  BackupConfig,
  BackupConfigUpdate,
  BackupImportMode,
  BackupImportResult,
  BackupRunResult,
} from '../types'

export const getBackupConfig = async (): Promise<BackupConfig> =>
  unwrapApiResponse(apiClient.GET('/api/v1/backup/config'), 'Unable to load backup settings')

export const updateBackupConfig = async (payload: BackupConfigUpdate): Promise<BackupConfig> =>
  unwrapApiResponse(apiClient.PUT('/api/v1/backup/config', { body: payload }), 'Unable to save backup settings')

export const runBackupNow = async (): Promise<BackupRunResult> =>
  unwrapApiResponse(apiClient.POST('/api/v1/backup/run'), 'Unable to run backup')

// authedFetch performs a manual fetch with the bearer token for the binary/multipart
// endpoints that can't go through the typed openapi-fetch client.
const authedFetch = (path: string, init?: RequestInit): Promise<Response> => {
  const headers = new Headers(init?.headers)
  const token = getApiAuthToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  return fetch(`${apiBaseUrl}${path}`, { ...init, headers })
}

const filenameFromDisposition = (header: string | null, fallback: string): string => {
  if (!header) {
    return fallback
  }
  const match = /filename\*?=(?:UTF-8'')?"?([^";]+)"?/i.exec(header)
  return match ? decodeURIComponent(match[1]) : fallback
}

const triggerDownload = (blob: Blob, filename: string): void => {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

const downloadFrom = async (path: string, fallbackName: string, errorMessage: string): Promise<void> => {
  const response = await authedFetch(path)
  if (!response.ok) {
    throw new Error(errorMessage)
  }
  const blob = await response.blob()
  triggerDownload(blob, filenameFromDisposition(response.headers.get('Content-Disposition'), fallbackName))
}

export const downloadBackup = (): Promise<void> =>
  downloadFrom('/api/v1/backup/export', 'overpacked-backup.zip', 'Unable to download backup')

export const exportItems = (includeImages: boolean): Promise<void> =>
  downloadFrom(
    `/api/v1/export/items?include_images=${includeImages ? 'true' : 'false'}`,
    includeImages ? 'items-export.zip' : 'items.csv',
    'Unable to export items',
  )

export interface ImportBackupParams {
  file: File
  mode: BackupImportMode
  password: string
}

export const importBackup = async ({ file, mode, password }: ImportBackupParams): Promise<BackupImportResult> => {
  const form = new FormData()
  form.append('file', file)
  form.append('mode', mode)
  form.append('password', password)

  const response = await authedFetch('/api/v1/backup/import', { method: 'POST', body: form })
  const payload = (await response.json().catch(() => null)) as BackupImportResult & { error?: string } | null
  if (!response.ok) {
    throw new Error(payload?.error ?? 'Unable to import backup')
  }
  return payload as BackupImportResult
}
