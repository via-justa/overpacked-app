// Format an ISO date string as DD-MM-YYYY; returns the input unchanged when unparseable.
export const formatDate = (value: string): string => {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) {
    return value
  }

  const day = String(parsed.getDate()).padStart(2, '0')
  const month = String(parsed.getMonth() + 1).padStart(2, '0')
  const year = parsed.getFullYear()
  return `${day}-${month}-${year}`
}

// Format an ISO date string as a localized medium date, e.g. "Jun 11, 2026".
export const formatDateMedium = (value: string): string =>
  new Date(value).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
