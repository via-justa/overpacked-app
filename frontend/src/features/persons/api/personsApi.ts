import { apiClient } from '../../../lib/api/client'
import type { Person, PersonCreate, PersonUpdate } from '../types'

const getErrorMessage = (error: unknown, fallback: string) => {
  if (error && typeof error === 'object' && 'error' in error) {
    const value = (error as { error?: unknown }).error
    if (typeof value === 'string' && value.length > 0) {
      return value
    }
  }

  return fallback
}

export const listPersons = async (): Promise<Person[]> => {
  const { data, error, response } = await apiClient.GET('/api/v1/persons')

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to load persons'))
  }

  return data
}

export const createPerson = async (payload: PersonCreate): Promise<Person> => {
  const { data, error, response } = await apiClient.POST('/api/v1/persons', {
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to create person'))
  }

  return data
}

export const updatePerson = async (
  personId: string,
  payload: PersonUpdate,
): Promise<Person> => {
  const { data, error, response } = await apiClient.PATCH('/api/v1/persons/{personId}', {
    params: {
      path: { personId },
    },
    body: payload,
  })

  if (!response.ok || !data) {
    throw new Error(getErrorMessage(error, 'Unable to update person'))
  }

  return data
}

export const removePerson = async (personId: string): Promise<void> => {
  const { error, response } = await apiClient.DELETE('/api/v1/persons/{personId}', {
    params: {
      path: { personId },
    },
  })

  if (!response.ok) {
    throw new Error(getErrorMessage(error, 'Unable to delete person'))
  }
}
