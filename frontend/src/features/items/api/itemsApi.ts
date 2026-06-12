import { apiClient } from '../../../lib/api/client'
import { ensureApiResponse, unwrapApiResponse } from '../../../lib/api/request'
import type { Item, ItemCreate, ItemTypeCreate, ItemTypeEntity, ItemTypeField, ItemTypeFieldInput, ItemTypeUpdate, ItemUpdate, Manufacturer, ManufacturerCreate, ManufacturerUpdate, Label, LabelCreate, ItemLabelAdd } from '../types'

export const listItems = async (): Promise<Item[]> =>
  unwrapApiResponse(apiClient.GET('/api/v1/items'), 'Unable to load items')

export const createItem = async (payload: ItemCreate): Promise<Item> =>
  unwrapApiResponse(apiClient.POST('/api/v1/items', { body: payload }), 'Unable to create item')

export const updateItem = async (itemId: string, payload: ItemUpdate): Promise<Item> =>
  unwrapApiResponse(
    apiClient.PATCH('/api/v1/items/{itemId}', { params: { path: { itemId } }, body: payload }),
    'Unable to update item',
  )

export const removeItem = async (itemId: string): Promise<void> =>
  ensureApiResponse(apiClient.DELETE('/api/v1/items/{itemId}', { params: { path: { itemId } } }), 'Unable to delete item')

export const listManufacturers = async (): Promise<Manufacturer[]> =>
  unwrapApiResponse(apiClient.GET('/api/v1/manufacturers'), 'Unable to load manufacturers')

export const createManufacturer = async (payload: ManufacturerCreate): Promise<Manufacturer> =>
  unwrapApiResponse(apiClient.POST('/api/v1/manufacturers', { body: payload }), 'Unable to create manufacturer')

export const updateManufacturer = async (manufacturerId: string, payload: ManufacturerUpdate): Promise<Manufacturer> =>
  unwrapApiResponse(
    apiClient.PATCH('/api/v1/manufacturers/{manufacturerId}', { params: { path: { manufacturerId } }, body: payload }),
    'Unable to update manufacturer',
  )

export const removeManufacturer = async (manufacturerId: string): Promise<void> =>
  ensureApiResponse(
    apiClient.DELETE('/api/v1/manufacturers/{manufacturerId}', { params: { path: { manufacturerId } } }),
    'Unable to delete manufacturer',
  )

export const createItemType = async (payload: ItemTypeCreate): Promise<{ id: string; name: string }> =>
  unwrapApiResponse(apiClient.POST('/api/v1/item-types', { body: payload }), 'Unable to create category') as Promise<{ id: string; name: string }>

export const listItemTypes = async (): Promise<ItemTypeEntity[]> =>
  unwrapApiResponse(apiClient.GET('/api/v1/item-types'), 'Unable to load categories')

export const updateItemType = async (typeId: string, payload: ItemTypeUpdate): Promise<ItemTypeEntity> =>
  unwrapApiResponse(
    apiClient.PATCH('/api/v1/item-types/{typeId}', { params: { path: { typeId } }, body: payload }),
    'Unable to update category',
  )

export const removeItemType = async (typeId: string): Promise<void> =>
  ensureApiResponse(
    apiClient.DELETE('/api/v1/item-types/{typeId}', { params: { path: { typeId } } }),
    'Unable to delete category',
  )

export const replaceItemTypeFields = async (typeId: string, fields: ItemTypeFieldInput[]): Promise<void> =>
  ensureApiResponse(
    apiClient.PUT('/api/v1/item-types/{typeId}/fields', { params: { path: { typeId } }, body: { fields } }),
    'Unable to save category fields',
  )

export const listItemTypeFields = async (typeId: string): Promise<ItemTypeField[]> =>
  unwrapApiResponse(
    apiClient.GET('/api/v1/item-types/{typeId}/fields', { params: { path: { typeId } } }),
    'Unable to load category fields',
  )

export const listLabels = async (): Promise<Label[]> =>
  unwrapApiResponse(apiClient.GET('/api/v1/labels'), 'Unable to load labels')

export const createLabel = async (payload: LabelCreate): Promise<Label> =>
  unwrapApiResponse(apiClient.POST('/api/v1/labels', { body: payload }), 'Unable to create label')

export const listItemLabels = async (itemId: string): Promise<Label[]> =>
  unwrapApiResponse(
    apiClient.GET('/api/v1/items/{itemId}/labels', { params: { path: { itemId } } }),
    'Unable to load item labels',
  )

export const addItemLabel = async (itemId: string, payload: ItemLabelAdd): Promise<Label> =>
  unwrapApiResponse(
    apiClient.POST('/api/v1/items/{itemId}/labels', { params: { path: { itemId } }, body: payload }),
    'Unable to add label to item',
  )

export const removeItemLabel = async (itemId: string, labelId: string): Promise<void> =>
  ensureApiResponse(
    apiClient.DELETE('/api/v1/items/{itemId}/labels/{labelId}', { params: { path: { itemId, labelId } } }),
    'Unable to remove label from item',
  )
