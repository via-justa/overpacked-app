import { apiClient } from '../../../lib/api/client'
import type { PackingList, PackingListCreate, PackingListUpdate, Label, PackingListLabelAdd } from '../types'

const readString = (value: unknown): string | null => {
  if (typeof value === 'string' && value.trim().length > 0) {
    return value
  }

  return null
}

const getErrorMessage = (error: unknown, fallback: string) => {
  if (!error || typeof error !== 'object') {
    return fallback
  }

  const objectError = error as {
    error?: unknown
    message?: unknown
    detail?: unknown
    details?: unknown
  }

  const directMessage =
    readString(objectError.error)
    ?? readString(objectError.message)
    ?? readString(objectError.detail)
    ?? readString(objectError.details)
  if (directMessage) {
    return directMessage
  }

  if (objectError.error && typeof objectError.error === 'object') {
    const nestedError = objectError.error as { message?: unknown; detail?: unknown; error?: unknown }
    const nestedMessage =
      readString(nestedError.message)
      ?? readString(nestedError.detail)
      ?? readString(nestedError.error)
    if (nestedMessage) {
      return nestedMessage
    }
  }

  return fallback
}

export const listPackingLists = async (): Promise<PackingList[]> => {
  const { data, error, response } = await apiClient.GET('/api/v1/packing-lists' as any, {})

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to load packing lists'))
  }

  return data as PackingList[]
}

export const getPackingList = async (listId: string): Promise<PackingList> => {
  const { data, error, response } = await apiClient.GET('/api/v1/packing-lists/{listId}' as any, {
    params: {
      path: { listId },
    },
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to load packing list'))
  }

  return data as PackingList
}

export const createPackingList = async (payload: PackingListCreate): Promise<PackingList> => {
  const { data, error, response } = await apiClient.POST('/api/v1/packing-lists' as any, {
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to create packing list'))
  }

  return data as PackingList
}

export const updatePackingList = async (listId: string, payload: PackingListUpdate): Promise<PackingList> => {
  const { data, error, response } = await apiClient.PATCH('/api/v1/packing-lists/{listId}' as any, {
    params: {
      path: { listId },
    },
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to update packing list'))
  }

  return data as PackingList
}

export const removePackingList = async (listId: string): Promise<void> => {
  const { error, response } = await apiClient.DELETE('/api/v1/packing-lists/{listId}' as any, {
    params: {
      path: { listId },
    },
  })

  if (!response.ok) {
    throw new Error(getErrorMessage(error, 'Unable to delete packing list'))
  }
}

export const listPackingListLabels = async (listId: string): Promise<Label[]> => {
  const { data, error, response } = await apiClient.GET('/api/v1/packing-lists/{listId}/labels' as any, {
    params: {
      path: { listId },
    },
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to load packing list labels'))
  }

  return data as Label[]
}

export const addPackingListLabel = async (listId: string, payload: PackingListLabelAdd): Promise<Label> => {
  const { data, error, response } = await apiClient.POST('/api/v1/packing-lists/{listId}/labels' as any, {
    params: {
      path: { listId },
    },
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to add label to packing list'))
  }

  return data as Label
}

export const removePackingListLabel = async (listId: string, labelId: string): Promise<void> => {
  const { error, response } = await apiClient.DELETE('/api/v1/packing-lists/{listId}/labels/{labelId}' as any, {
    params: {
      path: { listId, labelId },
    },
  })

  if (!response.ok) {
    throw new Error(getErrorMessage(error, 'Unable to remove label from packing list'))
  }
}
