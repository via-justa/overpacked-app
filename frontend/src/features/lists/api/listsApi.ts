import { apiClient } from '../../../lib/api/client'
import { ensureApiResponse, unwrapApiResponse } from '../../../lib/api/request'
import type { PackingList, PackingListCreate, PackingListUpdate, Label, PackingListLabelAdd } from '../types'

// NOTE: the packing-lists paths are cast to `any` because they are not yet in
// the generated OpenAPI types (known drift; see dev/repo-violations.md).

export const listPackingLists = async (): Promise<PackingList[]> =>
  unwrapApiResponse(apiClient.GET('/api/v1/packing-lists' as any, {}), 'Unable to load packing lists')

export const getPackingList = async (listId: string): Promise<PackingList> =>
  unwrapApiResponse(
    apiClient.GET('/api/v1/packing-lists/{listId}' as any, { params: { path: { listId } } }),
    'Unable to load packing list',
  )

export const createPackingList = async (payload: PackingListCreate): Promise<PackingList> =>
  unwrapApiResponse(
    apiClient.POST('/api/v1/packing-lists' as any, { body: payload }),
    'Unable to create packing list',
  )

export const updatePackingList = async (listId: string, payload: PackingListUpdate): Promise<PackingList> =>
  unwrapApiResponse(
    apiClient.PATCH('/api/v1/packing-lists/{listId}' as any, { params: { path: { listId } }, body: payload }),
    'Unable to update packing list',
  )

export const removePackingList = async (listId: string): Promise<void> =>
  ensureApiResponse(
    apiClient.DELETE('/api/v1/packing-lists/{listId}' as any, { params: { path: { listId } } }),
    'Unable to delete packing list',
  )

export const listPackingListLabels = async (listId: string): Promise<Label[]> =>
  unwrapApiResponse(
    apiClient.GET('/api/v1/packing-lists/{listId}/labels' as any, { params: { path: { listId } } }),
    'Unable to load packing list labels',
  )

export const addPackingListLabel = async (listId: string, payload: PackingListLabelAdd): Promise<Label> =>
  unwrapApiResponse(
    apiClient.POST('/api/v1/packing-lists/{listId}/labels' as any, { params: { path: { listId } }, body: payload }),
    'Unable to add label to packing list',
  )

export const removePackingListLabel = async (listId: string, labelId: string): Promise<void> =>
  ensureApiResponse(
    apiClient.DELETE('/api/v1/packing-lists/{listId}/labels/{labelId}' as any, { params: { path: { listId, labelId } } }),
    'Unable to remove label from packing list',
  )
