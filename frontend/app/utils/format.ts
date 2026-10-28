/** Formatting helpers shared by every screen. */

/** Local 24-hour clock time, e.g. "09:12". */
export function clockTime(date: Date = new Date()): string {
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${hours}:${minutes}`
}

/** Money with no decimals when the amount is whole, e.g. "₦45,000". */
export function money(value: number, currency = 'NGN'): string {
  const decimals = Number.isInteger(value) ? 0 : 2
  return `${currencySymbol(currency)}${value.toLocaleString('en', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  })}`
}

/** Money with the ISO code, used where the code matters (tables, receipts). */
export function moneyExact(value: number, currency = 'NGN'): string {
  return `${currency} ${value.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

export function currencySymbol(currency: string): string {
  return currency === 'NGN' ? '₦' : `${currency} `
}

/** "12 Oct 2026" — dates coming from the API are ISO strings. */
export function shortDate(value?: string): string {
  if (!value) return '—'
  return new Intl.DateTimeFormat('en', { day: 'numeric', month: 'short', year: 'numeric' }).format(asDate(value))
}

/** "12 Oct, 14:03". */
export function shortDateTime(value?: string): string {
  if (!value) return '—'
  return new Intl.DateTimeFormat('en', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  }).format(asDate(value))
}

/** "Today", "Yesterday" or "12 Oct" — used in lists. */
export function relativeDay(value?: string): string {
  if (!value) return '—'
  const date = asDate(value)
  const today = new Date()
  const startOf = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime()
  const days = Math.round((startOf(today) - startOf(date)) / 86_400_000)
  if (days === 0) return 'Today'
  if (days === 1) return 'Yesterday'
  return shortDate(value)
}

/** "in transit" → "In transit". */
export function humanise(value?: string): string {
  if (!value) return ''
  const spaced = value.replaceAll('_', ' ')
  return spaced.charAt(0).toUpperCase() + spaced.slice(1)
}

/** YYYY-MM-DD in local time. */
export function isoDay(date: Date = new Date()): string {
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}

/** Inclusive period of the last N days, for report queries. */
export function lastDays(days: number, today: Date = new Date()): { from: string; to: string } {
  const from = new Date(today)
  from.setDate(from.getDate() - (days - 1))
  return { from: isoDay(from), to: isoDay(today) }
}

/** Days between two dates, for ageing balances. */
export function daysSince(value?: string): number {
  if (!value) return 0
  return Math.max(0, Math.floor((Date.now() - asDate(value).getTime()) / 86_400_000))
}

function asDate(value: string): Date {
  // Date-only strings would otherwise be parsed as UTC and shift a day.
  return new Date(value.length === 10 ? `${value}T00:00:00` : value)
}
