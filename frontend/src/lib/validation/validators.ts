/**
 * Validates if a string is a valid HTTP/HTTPS URL
 */
export function isValidUrl(value: string): boolean {
  if (!value.trim()) {
    return true
  }

  try {
    const parsed = new URL(value)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

/**
 * Validates if a string represents a valid float number
 */
export function isFloat(value: string): boolean {
  if (!value.trim()) {
    return true
  }

  const parsed = Number(value)
  return !Number.isNaN(parsed)
}

/**
 * Validates if a string represents a valid integer
 */
export function isInteger(value: string): boolean {
  if (!value.trim()) {
    return true
  }

  return /^-?\d+$/.test(value.trim())
}
