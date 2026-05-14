export const slugifyCategoryId = (value: string): string => {
  const lower = value.trim().toLowerCase()
  let normalized = ''
  let lastWasSeparator = false

  for (const char of lower) {
    const isAlphaNumeric = (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')
    if (isAlphaNumeric) {
      normalized += char
      lastWasSeparator = false
      continue
    }

    if (!lastWasSeparator && normalized.length > 0) {
      normalized += '_'
      lastWasSeparator = true
    }
  }

  if (normalized.endsWith('_')) {
    normalized = normalized.slice(0, -1)
  }

  if (normalized.length > 0) {
    return normalized
  }

  return `category_${Date.now()}`
}
