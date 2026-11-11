/**
 * Landing page pricing — the single source of truth for the plan grid.
 *
 * These are proposed prices, not a billing contract: plans and their limits are
 * meant to be configured per tenant (see the white-label section), so anything
 * shown on the landing page should be read from here rather than typed into a
 * card by hand.
 */

export type BillingPeriod = 'month' | 'year'

export interface Plan {
  /** Stable key, e.g. for a future checkout link. */
  id: string
  name: string
  /** Price in the smallest sensible unit of `currency`. 0 means free. */
  price: number
  currency: 'NGN'
  period: BillingPeriod
  tagline: string
  description: string
  features: string[]
  cta: { label: string; to: string }
  /** The recommended plan — drawn with the inverted (navy) treatment. */
  featured?: boolean
}

export const CURRENCY_SYMBOL = '₦'

export const PLANS: Plan[] = [
  {
    id: 'free',
    name: 'Free',
    price: 0,
    currency: 'NGN',
    period: 'month',
    tagline: 'One counter, on the record',
    description: 'Put a single branch on the system before you commit to anything.',
    features: [
      '1 branch, 1 counter',
      'Waybills with a full custody trail',
      'Customer receipts and balances',
      'Daily collection summary',
    ],
    cta: { label: 'Start free', to: '/sign-in' },
  },
  {
    id: 'starter',
    name: 'Starter',
    price: 25_000,
    currency: 'NGN',
    period: 'month',
    tagline: 'For growing counters',
    description: 'The first few routes, the first few staff, one shared record.',
    features: [
      'Up to 3 branches',
      'Unlimited staff accounts and roles',
      'Bags and container manifests',
      'Collections, balances and receipts',
      'Exception logging and follow-up',
    ],
    cta: { label: 'Choose Starter', to: '/sign-in' },
  },
  {
    id: 'professional',
    name: 'Professional',
    price: 65_000,
    currency: 'NGN',
    period: 'month',
    tagline: 'For daily dispatch',
    description: 'Multi-branch operations moving parcels, bags and trucks every day.',
    features: [
      'Unlimited branches and counters',
      'Partial dispatch and split manifests',
      'Branch performance reports',
      'Fleet, drivers and vehicles',
      'Activity log for every action',
      'Priority support',
    ],
    cta: { label: 'Choose Professional', to: '/sign-in' },
    featured: true,
  },
  {
    id: 'business',
    name: 'Business',
    price: 120_000,
    currency: 'NGN',
    period: 'month',
    tagline: 'For networks and white-label',
    description: 'Run the platform under your own brand, across your own network.',
    features: [
      'Everything in Professional',
      'White-label branding and domain',
      'Tenant-level plans and price books',
      'Guided onboarding and data import',
      'Named support contact',
    ],
    cta: { label: 'Talk to us', to: 'mailto:hello@basedonlogistics.com' },
  },
]

/** "Free" for a zero-price plan, otherwise "₦65,000". */
export function formatPrice(plan: Plan): string {
  if (plan.price === 0) return 'Free'
  return `${CURRENCY_SYMBOL}${plan.price.toLocaleString('en-NG')}`
}

/** Small print shown under the grid. */
export const PRICING_NOTE =
  'Indicative pricing. Amounts, branches and limits are configured per workspace, so what you pay is agreed before you switch anything on.'
