import { apiClient } from '../../../lib/api/client'
import { ensureApiResponse, unwrapApiResponse } from '../../../lib/api/request'
import type { Person, PersonCreate, PersonUpdate } from '../types'

export const listPersons = async (): Promise<Person[]> =>
  unwrapApiResponse(apiClient.GET('/api/v1/persons'), 'Unable to load persons')

export const createPerson = async (payload: PersonCreate): Promise<Person> =>
  unwrapApiResponse(apiClient.POST('/api/v1/persons', { body: payload }), 'Unable to create person')

export const updatePerson = async (personId: string, payload: PersonUpdate): Promise<Person> =>
  unwrapApiResponse(
    apiClient.PATCH('/api/v1/persons/{personId}', { params: { path: { personId } }, body: payload }),
    'Unable to update person',
  )

export const removePerson = async (personId: string): Promise<void> =>
  ensureApiResponse(
    apiClient.DELETE('/api/v1/persons/{personId}', { params: { path: { personId } } }),
    'Unable to delete person',
  )
