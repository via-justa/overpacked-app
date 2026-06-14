import { getErrorMessage } from './errors'

// Shape returned by every openapi-fetch call.
type ApiResult<T> = {
  data?: T
  error?: unknown
  response: Response
}

// Unwraps a data-returning endpoint: throws with a friendly message when the
// response failed or carried no body, otherwise returns the typed data.
export const unwrapApiResponse = async <T>(call: Promise<ApiResult<T>>, fallback: string): Promise<T> => {
  const { data, error, response } = await call
  if (!response.ok || data === undefined || data === null) {
    throw new Error(getErrorMessage(error, fallback))
  }

  return data
}

// Asserts a no-content endpoint (e.g. DELETE → 204) succeeded; throws otherwise.
export const ensureApiResponse = async (call: Promise<ApiResult<unknown>>, fallback: string): Promise<void> => {
  const { error, response } = await call
  if (!response.ok) {
    throw new Error(getErrorMessage(error, fallback))
  }
}
