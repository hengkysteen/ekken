/**
 * Truncate a string to a specific length and add ellipsis if needed.
 */
export function truncate(str: string, length: number = 30): string {
  if (!str) return ''
  if (str.length <= length) return str
  return str.slice(0, length) + '...'
}
