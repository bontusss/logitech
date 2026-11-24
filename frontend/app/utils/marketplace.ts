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
        pill: 'bg-accent-100 text-accent-800 dark:bg-accent-500/15 dark:text-accent-300',
        tile: 'bg-accent-100 text-accent-800 dark:bg-accent-500/15 dark:text-accent-300',
      }
    case 'same_day':
      return {
        pill: 'bg-brand-100 text-ink-900 dark:bg-white/10 dark:text-brand-200',
        tile: 'bg-brand-100 text-ink-900 dark:bg-white/10 dark:text-brand-200',
      }
    default:
      return {
        pill: 'bg-brand-50 text-ink-900/70 dark:bg-white/8 dark:text-ink-900/70',
        tile: 'bg-brand-50 text-ink-900/70 dark:bg-white/8 dark:text-ink-900/70',
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
