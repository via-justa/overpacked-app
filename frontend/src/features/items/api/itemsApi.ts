import { apiClient } from '../../../lib/api/client'
import type { Item, ItemCreate, ItemTypeCreate, ItemTypeEntity, ItemTypeField, ItemTypeFieldInput, ItemTypeUpdate, ItemUpdate, Manufacturer, ManufacturerCreate, ManufacturerUpdate, Label, LabelCreate, ItemLabelAdd } from '../types'

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

export const listItems = async (): Promise<Item[]> => {
  const { data, error, response } = await apiClient.GET('/api/v1/items')

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to load items'))
  }

  return data
}

export const createItem = async (payload: ItemCreate): Promise<Item> => {
  const { data, error, response } = await apiClient.POST('/api/v1/items', {
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to create item'))
  }

  return data
}

export const updateItem = async (itemId: string, payload: ItemUpdate): Promise<Item> => {
  const { data, error, response } = await apiClient.PATCH('/api/v1/items/{itemId}', {
    params: {
      path: { itemId },
    },
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to update item'))
  }

  return data
}

export const removeItem = async (itemId: string): Promise<void> => {
  const { error, response } = await apiClient.DELETE('/api/v1/items/{itemId}', {
    params: {
      path: { itemId },
    },
  })

  if (!response.ok) {
    throw new Error(getErrorMessage(error, 'Unable to delete item'))
  }
}

export const listManufacturers = async (): Promise<Manufacturer[]> => {
  const { data, error, response } = await apiClient.GET('/api/v1/manufacturers')

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to load manufacturers'))
  }

  return data
}

export const createManufacturer = async (payload: ManufacturerCreate): Promise<Manufacturer> => {
  const { data, error, response } = await apiClient.POST('/api/v1/manufacturers', {
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to create manufacturer'))
  }

  return data
}

export const updateManufacturer = async (manufacturerId: string, payload: ManufacturerUpdate): Promise<Manufacturer> => {
  const { data, error, response } = await apiClient.PATCH('/api/v1/manufacturers/{manufacturerId}', {
    params: {
      path: { manufacturerId },
    },
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to update manufacturer'))
  }

  return data
}

export const removeManufacturer = async (manufacturerId: string): Promise<void> => {
  const { error, response } = await apiClient.DELETE('/api/v1/manufacturers/{manufacturerId}', {
    params: {
      path: { manufacturerId },
    },
  })

  if (!response.ok) {
    throw new Error(getErrorMessage(error, 'Unable to delete manufacturer'))
  }
}

export const createItemType = async (payload: ItemTypeCreate): Promise<{ id: string; name: string }> => {
  const { data, error, response } = await apiClient.POST('/api/v1/item-types', {
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to create category'))
  }

  return data as { id: string; name: string }
}

export const listItemTypes = async (): Promise<ItemTypeEntity[]> => {
  const { data, error, response } = await apiClient.GET('/api/v1/item-types')

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to load categories'))
  }

  return data
}

export const updateItemType = async (typeId: string, payload: ItemTypeUpdate): Promise<ItemTypeEntity> => {
  const { data, error, response } = await apiClient.PATCH('/api/v1/item-types/{typeId}', {
    params: {
      path: { typeId },
    },
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to update category'))
  }

  return data
}

export const removeItemType = async (typeId: string): Promise<void> => {
  const { error, response } = await apiClient.DELETE('/api/v1/item-types/{typeId}', {
    params: {
      path: { typeId },
    },
  })

  if (!response.ok) {
    throw new Error(getErrorMessage(error, 'Unable to delete category'))
  }
}

export const replaceItemTypeFields = async (typeId: string, fields: ItemTypeFieldInput[]): Promise<void> => {
  const { error, response } = await apiClient.PUT('/api/v1/item-types/{typeId}/fields', {
    params: {
      path: { typeId },
    },
    body: {
      fields,
    },
  })

  if (!response.ok) {
    throw new Error(getErrorMessage(error, 'Unable to save category fields'))
  }
}

export const listItemTypeFields = async (typeId: string): Promise<ItemTypeField[]> => {
  const { data, error, response } = await apiClient.GET('/api/v1/item-types/{typeId}/fields', {
    params: {
      path: { typeId },
    },
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to load category fields'))
  }

  return data
}

export const listLabels = async (): Promise<Label[]> => {
  const { data, error, response } = await apiClient.GET('/api/v1/labels')

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to load labels'))
  }

  return data
}

export const createLabel = async (payload: LabelCreate): Promise<Label> => {
  const { data, error, response } = await apiClient.POST('/api/v1/labels', {
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to create label'))
  }

  return data
}

export const listItemLabels = async (itemId: string): Promise<Label[]> => {
  const { data, error, response } = await apiClient.GET('/api/v1/items/{itemId}/labels', {
    params: {
      path: { itemId },
    },
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to load item labels'))
  }

  return data
}

export const addItemLabel = async (itemId: string, payload: ItemLabelAdd): Promise<Label> => {
  const { data, error, response } = await apiClient.POST('/api/v1/items/{itemId}/labels', {
    params: {
      path: { itemId },
    },
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to add label to item'))
  }

  return data
}

export const removeItemLabel = async (itemId: string, labelId: string): Promise<void> => {
  const { error, response } = await apiClient.DELETE('/api/v1/items/{itemId}/labels/{labelId}', {
    params: {
      path: { itemId, labelId },
    },
  })

  if (!response.ok) {
    throw new Error(getErrorMessage(error, 'Unable to remove label from item'))
  }
}