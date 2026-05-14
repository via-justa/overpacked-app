export const normalizeTitleWords = (value: string): string => {
  const trimmed = value.trim()
  if (!trimmed) {
    return ''
  }

  return trimmed
    .split(/\s+/)
    .map((word) => {
      if (!word) {
        return ''
      }

      const lower = word.toLowerCase()
      return `${lower.charAt(0).toUpperCase()}${lower.slice(1)}`
    })
    .join(' ')
}
