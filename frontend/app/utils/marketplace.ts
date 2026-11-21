const naira = new Intl.NumberFormat('en-NG', {
  style: 'currency',
  currency: 'NGN',
  maximumFractionDigits: 0,
})

/** Formats an integer minor-unit amount (kobo) for display. */
export function formatMinor(minor: number | null | undefined, currency = 'NGN') {
  const value = (minor ?? 0) / 100
  if (currency === 'NGN') return naira.format(value)
  return new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency,
    maximumFractionDigits: 2,
  }).format(value)
}

export function serviceTone(key: string) {
  switch (key) {
    case 'express':
      return {
        pill: 'bg-sky-100 text-sky-800 dark:bg-sky-400/15 dark:text-sky-200',
        tile: 'bg-sky-100 text-sky-800 dark:bg-sky-400/15 dark:text-sky-200',
      }
    case 'same_day':
      return {
        pill: 'bg-amber-100 text-amber-800 dark:bg-amber-400/15 dark:text-amber-200',
        tile: 'bg-amber-100 text-amber-800 dark:bg-amber-400/15 dark:text-amber-200',
      }
    default:
      return {
        pill: 'bg-violet-100 text-violet-800 dark:bg-violet-400/15 dark:text-violet-200',
        tile: 'bg-violet-100 text-violet-800 dark:bg-violet-400/15 dark:text-violet-200',
      }
  }
}

export function deliveryMethodLabel(method: string) {
  const labels: Record<string, string> = {
    'door-to-door': 'Door to door',
    'door-to-terminal': 'Door to terminal',
    'terminal-to-door': 'Terminal to door',
    'terminal-to-terminal': 'Terminal to terminal',
  }
  return labels[method] ?? method
}
