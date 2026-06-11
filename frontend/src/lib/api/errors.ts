// Shared helpers for turning an API error payload into a human-readable message.
// Previously each feature's api/*.ts redefined its own copy of these.

export const readString = (value: unknown): string | null => {
  if (typeof value === 'string' && value.trim().length > 0) {
    return value
  }

  return null
}

export const getErrorMessage = (error: unknown, fallback: string): string => {
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
