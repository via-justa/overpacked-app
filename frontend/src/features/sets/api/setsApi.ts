import { apiClient } from '../../../lib/api/client'
import type { ItemSet, ItemSetCreate, ItemSetUpdate, SetItemCreate, SetItemUpdate, SetItemWithDetails } from '../types'

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

  return fallback
}

export const listSets = async (): Promise<ItemSet[]> => {
  const { data, error, response } = await apiClient.GET('/api/v1/sets')

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to load sets'))
  }

  return data as ItemSet[]
}

export const createSet = async (payload: ItemSetCreate): Promise<ItemSet> => {
  const { data, error, response } = await apiClient.POST('/api/v1/sets', {
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to create set'))
  }

  return data as ItemSet
}

export const updateSet = async (setId: string, payload: ItemSetUpdate): Promise<ItemSet> => {
  const { data, error, response } = await apiClient.PATCH('/api/v1/sets/{setId}', {
    params: {
      path: { setId },
    },
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to update set'))
  }

  return data as ItemSet
}

export const removeSet = async (setId: string): Promise<void> => {
  const { error, response } = await apiClient.DELETE('/api/v1/sets/{setId}', {
    params: {
      path: { setId },
    },
  })

  if (!response.ok) {
    throw new Error(getErrorMessage(error, 'Unable to delete set'))
  }
}

export const listSetItems = async (setId: string): Promise<SetItemWithDetails[]> => {
  const { data, error, response } = await apiClient.GET('/api/v1/sets/{setId}/items', {
    params: {
      path: { setId },
    },
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to load set items'))
  }

  return data as SetItemWithDetails[]
}

export const addSetItem = async (setId: string, payload: SetItemCreate): Promise<SetItemWithDetails> => {
  const { data, error, response } = await apiClient.POST('/api/v1/sets/{setId}/items', {
    params: {
      path: { setId },
    },
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to add item to set'))
  }

  return data as SetItemWithDetails
}

export const updateSetItem = async (setId: string, itemId: string, payload: SetItemUpdate): Promise<SetItemWithDetails> => {
  const { data, error, response } = await apiClient.PATCH('/api/v1/sets/{setId}/items/{itemId}', {
    params: {
      path: { setId, itemId },
    },
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to update set item'))
  }

  return data as SetItemWithDetails
}

export const removeSetItem = async (setId: string, itemId: string): Promise<void> => {
  const { error, response } = await apiClient.DELETE('/api/v1/sets/{setId}/items/{itemId}', {
    params: {
      path: { setId, itemId },
    },
  })

  if (!response.ok) {
    throw new Error(getErrorMessage(error, 'Unable to remove item from set'))
  }
}
