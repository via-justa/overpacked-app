import type { Currency } from '../../features/settings/types'

/**
 * Formats a number value with "Not set" fallback.
 * @param value - The number to format
 * @param formatter - Optional custom formatter function
 * @returns Formatted string or "Not set"
 */
export function formatNumber(value?: number | null, formatter?: (n: number) => string): string {
  if (typeof value !== 'number') {
    return 'Not set'
  }
  return formatter ? formatter(value) : String(value)
}

/**
 * Formats a monetary value with currency symbol.
 * @param value - The numeric value
 * @param currency - The currency type (usd or eur)
 * @param formatter - Formatter function for the number
 * @returns Formatted currency string or "Not set"
 */
export function formatValue(
  value?: number | null,
  currency?: Currency,
  formatter?: (n: number) => string
): string {
  if (typeof value !== 'number') {
    return 'Not set'
  }
  const formatted = formatter ? formatter(value) : String(value)
  const currencySymbol = currency === 'usd' ? '$' : '€'
  return `${formatted} ${currencySymbol}`
}

/**
 * Formats carry status enum value to display text.
 * @param value - The carry status value
 * @returns Formatted status string or "Not set"
 */
export function formatCarryStatus(value?: string | null): string {
  if (!value) {
    return 'Not set'
  }

  if (value === 'packed') {
    return 'Packed'
  }

  if (value === 'worn') {
    return 'Worn'
  }

  return value
}

/**
 * Formats snake_case type strings to Title Case.
 * @param value - The snake_case string
 * @returns Title cased string
 */
export function formatType(value: string): string {
  return value
    .split('_')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}

/**
 * Formats text with "Not set" fallback for empty strings.
 * @param value - The text to format
 * @returns The text or "Not set"
 */
export function formatText(value?: string | null): string {
  if (!value?.trim()) {
    return 'Not set'
  }
  return value
}
