/**
 * Safely reads a value from localStorage with SSR compatibility
 */
export function getStoredValue<T extends string>(
  key: string,
  validator: (value: string) => value is T,
  defaultValue: T
): T {
  if (globalThis.window === undefined) {
    return defaultValue
  }

  const stored = globalThis.window.localStorage.getItem(key)
  if (!stored) {
    return defaultValue
  }

  return validator(stored) ? stored : defaultValue
}

/**
 * Safely writes a value to localStorage with SSR compatibility
 */
export function setStoredValue(key: string, value: string): void {
  if (globalThis.window === undefined) {
    return
  }

  globalThis.window.localStorage.setItem(key, value)
}

/**
 * Type guard helpers for common stored value types
 */
export function isViewMode(value: string): value is 'cards' | 'table' {
  return value === 'cards' || value === 'table'
}

export function isDetailMode(value: string): value is 'simple' | 'expanded' {
  return value === 'simple' || value === 'expanded'
}
