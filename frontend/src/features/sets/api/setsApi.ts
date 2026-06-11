import { apiClient } from '../../../lib/api/client'
import { ensureApiResponse, unwrapApiResponse } from '../../../lib/api/request'
import type { ItemSet, ItemSetCreate, ItemSetUpdate, SetItemCreate, SetItemUpdate, SetItemWithDetails } from '../types'

export const listSets = async (): Promise<ItemSet[]> =>
  unwrapApiResponse(apiClient.GET('/api/v1/sets'), 'Unable to load sets')

export const createSet = async (payload: ItemSetCreate): Promise<ItemSet> =>
  unwrapApiResponse(apiClient.POST('/api/v1/sets', { body: payload }), 'Unable to create set')

export const updateSet = async (setId: string, payload: ItemSetUpdate): Promise<ItemSet> =>
  unwrapApiResponse(
    apiClient.PATCH('/api/v1/sets/{setId}', { params: { path: { setId } }, body: payload }),
    'Unable to update set',
  )

export const removeSet = async (setId: string): Promise<void> =>
  ensureApiResponse(apiClient.DELETE('/api/v1/sets/{setId}', { params: { path: { setId } } }), 'Unable to delete set')

export const listSetItems = async (setId: string): Promise<SetItemWithDetails[]> =>
  unwrapApiResponse(
    apiClient.GET('/api/v1/sets/{setId}/items', { params: { path: { setId } } }),
    'Unable to load set items',
  )

export const addSetItem = async (setId: string, payload: SetItemCreate): Promise<SetItemWithDetails> =>
  unwrapApiResponse(
    apiClient.POST('/api/v1/sets/{setId}/items', { params: { path: { setId } }, body: payload }),
    'Unable to add item to set',
  )

export const updateSetItem = async (setId: string, itemId: string, payload: SetItemUpdate): Promise<SetItemWithDetails> =>
  unwrapApiResponse(
    apiClient.PATCH('/api/v1/sets/{setId}/items/{itemId}', { params: { path: { setId, itemId } }, body: payload }),
    'Unable to update set item',
  )

export const removeSetItem = async (setId: string, itemId: string): Promise<void> =>
  ensureApiResponse(
    apiClient.DELETE('/api/v1/sets/{setId}/items/{itemId}', { params: { path: { setId, itemId } } }),
    'Unable to remove item from set',
  )
