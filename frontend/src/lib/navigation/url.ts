/**
 * Returns the URL only if it is a safe http(s) URL, otherwise undefined.
 *
 * Defends href bindings against stored XSS: a value like `javascript:alert(1)`
 * or `data:…` is rejected so it is never rendered as a clickable link. This
 * complements the server-side scheme validation and also covers data that can
 * bypass the create/update API (e.g. restored backups, CSV import).
 */
export function safeHttpUrl(value: string | null | undefined): string | undefined {
  if (!value) {
    return undefined
  }
  const trimmed = value.trim()
  if (!trimmed) {
    return undefined
  }
  try {
    const parsed = new URL(trimmed)
    if (parsed.protocol === 'http:' || parsed.protocol === 'https:') {
      return trimmed
    }
  } catch {
    return undefined
  }
  return undefined
}
