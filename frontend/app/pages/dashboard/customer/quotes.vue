<script setup lang="ts">
import {
  ArrowRight,
  CircleAlert,
  Clock3,
  FileText,
  Loader2,
  MapPin,
  Package,
} from 'lucide-vue-next'
import type { MarketplaceQuote } from '~/types/marketplace'

useHead({ title: 'Saved Quotes · BOL' })

interface SavedQuoteRow {
  saved_quote_id: string
  selection_status: string
  selected_at: string | null
  quote: MarketplaceQuote
  expired: boolean
}

const { api } = useApi()

const quotes = ref<SavedQuoteRow[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const result = await api<{ quotes: SavedQuoteRow[] }>('/customer/quotes')
    quotes.value = result.quotes
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not load your saved quotes.')
  } finally {
    loading.value = false
  }
}

function formatMoney(minor: number, currency = 'NGN') {
  return new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  }).format(minor / 100)
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString('en-NG', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function routeLabel(quote: MarketplaceQuote) {
  const from = quote.origin_zone_name || quote.origin_city
  const to = quote.destination_zone_name || quote.destination_city
  return `${from} → ${to}`
}

function destinationLabel(quote: MarketplaceQuote) {
  return [quote.destination_state, quote.destination_city].filter(Boolean).join(', ')
}

function bookingLink(quote: MarketplaceQuote) {
  const slug = quote.booking_slug || quote.provider_slug
  return slug ? `/book/${slug}` : ''
}

function statusLabel(row: SavedQuoteRow) {
  if (row.expired) return 'Expired'
  if (row.selection_status === 'selected' || row.quote.status === 'selected') return 'Selected'
  if (row.quote.status === 'active') return 'Available'
  return row.quote.status
}

onMounted(load)
</script>

<template>
  <div class="dash-page">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Shipments</p>
        <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Saved quotes</h1>
        <p class="mt-1.5 max-w-2xl text-[12.5px] leading-relaxed text-ink-900/50">
          Quotes you selected while comparing providers, with the prices and routes you were shown.
        </p>
      </div>
      <AppButton to="/#rate-search" size="md">
        New rate search <ArrowRight :size="14" />
      </AppButton>
    </div>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ error }}
    </p>

    <div v-if="loading" class="mt-4 flex items-center justify-center gap-2 rounded-xl border border-ink-900/10 bg-surface p-12 text-[12px] text-ink-900/45">
      <Loader2 :size="15" class="animate-spin" /> Loading your quotes…
    </div>

    <div v-else-if="!quotes.length" class="mt-4 rounded-xl border border-ink-900/10 bg-surface p-10 text-center">
      <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
        <FileText :size="20" />
      </span>
      <h2 class="mt-3 text-[13.5px] font-semibold">No saved quotes yet</h2>
      <p class="mx-auto mt-1 max-w-md text-[11.5px] leading-relaxed text-ink-900/45">
        Search for a provider and select a quote while booking. Quotes you pick will be listed here.
      </p>
      <AppButton to="/#rate-search" size="sm" class="mt-4">Compare rates</AppButton>
    </div>

    <ul v-else class="mt-4 space-y-2.5">
      <li v-for="row in quotes" :key="row.saved_quote_id" class="rounded-xl border border-ink-900/10 bg-surface p-4">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div class="flex min-w-0 items-start gap-3">
            <span class="flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-ink-900/8 bg-gradient-to-br from-slate-50 to-white text-[13px] font-bold text-brand-700 dark:border-white/10 dark:from-white/10 dark:to-transparent">
              <img v-if="row.quote.provider_logo" :src="row.quote.provider_logo" :alt="`${row.quote.provider_name} logo`" class="size-full object-cover" />
              <span v-else>{{ row.quote.provider_name.charAt(0).toUpperCase() }}</span>
            </span>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <p class="text-[12.5px] font-semibold">{{ row.quote.provider_name }}</p>
                <span
                  class="rounded-full border px-2 py-0.5 text-[9.5px] font-semibold"
                  :class="row.expired
                    ? 'border-slate-200 bg-slate-50 text-slate-600'
                    : row.quote.status === 'selected' || row.selection_status === 'selected'
                      ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                      : 'border-brand-200 bg-brand-50 text-brand-700'"
                >
                  {{ statusLabel(row) }}
                </span>
                <span
                  v-if="row.quote.price_confirmation_type === 'estimated'"
                  class="rounded-full border border-amber-200 bg-amber-50 px-2 py-0.5 text-[9.5px] font-semibold text-amber-700"
                >
                  Estimated
                </span>
              </div>
              <p class="mt-1.5 inline-flex items-center gap-1 text-[11.5px] font-medium text-ink-900/65">
                <MapPin :size="11" class="text-brand-500" /> {{ routeLabel(row.quote) }}
              </p>
              <div class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-[10.5px] text-ink-900/42">
                <span>{{ row.quote.service_name }}</span>
                <span class="inline-flex items-center gap-1"><Package :size="10" /> {{ row.quote.chargeable_weight_kg }} kg</span>
                <span class="inline-flex items-center gap-1"><Clock3 :size="10" /> {{ destinationLabel(row.quote) }}</span>
              </div>
            </div>
          </div>

          <div class="flex shrink-0 items-center justify-between gap-2 sm:flex-col sm:items-end">
            <div class="text-right">
              <p class="text-[14px] font-semibold tabular-nums">{{ formatMoney(row.quote.total_price_minor, row.quote.currency) }}</p>
              <p class="mt-0.5 text-[9.5px] text-ink-900/35">
                <template v-if="row.selected_at">Selected {{ formatDate(row.selected_at) }}</template>
                <template v-else-if="row.expired">Expired {{ formatDate(row.quote.expires_at) }}</template>
                <template v-else>Expires {{ formatDate(row.quote.expires_at) }}</template>
              </p>
            </div>
            <NuxtLink
              v-if="bookingLink(row.quote)"
              :to="bookingLink(row.quote)"
              class="inline-flex items-center gap-1 rounded-lg bg-accent-500 px-2.5 py-1.5 text-[10.5px] font-semibold text-white transition hover:bg-accent-400"
            >
              Book again <ArrowRight :size="12" />
            </NuxtLink>
          </div>
        </div>

        <p v-if="row.expired" class="mt-3 rounded-lg border border-ink-900/8 bg-canvas/60 px-2.5 py-2 text-[10.5px] text-ink-900/50">
          This quote has expired. Search again for the current price before booking.
        </p>
        <p v-else-if="row.quote.price_confirmation_type === 'estimated'" class="mt-3 rounded-lg border border-amber-200 bg-amber-50/60 px-2.5 py-2 text-[10.5px] text-amber-800">
          {{ row.quote.price_confirmation_note || 'Final price will be confirmed by the provider after package inspection.' }}
        </p>
      </li>
    </ul>
  </div>
</template>
