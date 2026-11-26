<script setup lang="ts">
import {
  ArrowUpRight,
  Clock3,
  Loader2,
  MapPin,
  PackageCheck,
  Search,
  ShieldCheck,
  Star,
  Truck,
  Wifi,
} from 'lucide-vue-next'
import type {
  MarketplaceQuote,
  MarketplaceSearchPayload,
  MarketplaceSearchResponse,
} from '~/types/marketplace'

const props = defineProps<{
  loading: boolean
  error: string
  response: MarketplaceSearchResponse | null
  search: MarketplaceSearchPayload | null
}>()

const sortMode = ref<'cheapest' | 'rating' | 'pickup' | 'delivery'>('cheapest')

function deliveryHours(quote: MarketplaceQuote) {
  return quote.estimated_delivery_hours ?? quote.estimated_duration_days * 24
}

function pickupHours(quote: MarketplaceQuote) {
  return quote.estimated_pickup_hours ?? Number.POSITIVE_INFINITY
}

const visibleQuotes = computed<MarketplaceQuote[]>(() => {
  const quotes = [...(props.response?.quotes ?? [])]
  if (sortMode.value === 'rating') {
    quotes.sort((a, b) => b.provider_rating - a.provider_rating || a.total_price_minor - b.total_price_minor)
  } else if (sortMode.value === 'pickup') {
    quotes.sort(
      (a, b) =>
        pickupHours(a) - pickupHours(b) ||
        deliveryHours(a) - deliveryHours(b) ||
        a.total_price_minor - b.total_price_minor,
    )
  } else if (sortMode.value === 'delivery') {
    quotes.sort(
      (a, b) =>
        deliveryHours(a) - deliveryHours(b) || a.total_price_minor - b.total_price_minor,
    )
  } else {
    quotes.sort((a, b) => a.total_price_minor - b.total_price_minor)
  }
  return quotes.slice(0, 3)
})
const totalQuotes = computed(() => props.response?.quotes.length ?? 0)

const categories: Record<string, string> = {
  document: 'Documents',
  cloth: 'Clothing',
  electronics: 'Electronics',
  food: 'Food',
  jewelry: 'Jewelry',
  other: 'Other',
}

function bookingHref(quote: MarketplaceQuote) {
  if (!quote.booking_slug) return undefined
  const params = new URLSearchParams()
  const search = props.search
  if (search) {
    if (search.destination_zone_id) params.set('destination_zone_id', search.destination_zone_id)
    if (search.origin_zone_id) params.set('origin_zone_id', search.origin_zone_id)
    if (search.destination_lga) params.set('destination_lga', search.destination_lga)
    if (search.destination_state) params.set('destination_state', search.destination_state)
    if (search.origin_lga) params.set('origin_lga', search.origin_lga)
    if (search.origin_state) params.set('origin_state', search.origin_state)
    if (search.weight_kg) params.set('weight_kg', String(search.weight_kg))
    if (search.category) params.set('category', search.category)
    if (search.service_mode) params.set('service_mode', search.service_mode)
  }
  const query = params.toString()
  return `/book/${encodeURIComponent(quote.booking_slug)}${query ? `?${query}` : ''}`
}

function etaLabel(quote: MarketplaceQuote) {
  return quote.estimated_delivery || `${quote.estimated_duration_days} day(s)`
}

function handlingLabel(quote: MarketplaceQuote) {
  return quote.service_mode === 'dropoff' ? 'Drop off at hub' : deliveryMethodLabel(quote.delivery_method)
}

function pickupLabel(quote: MarketplaceQuote) {
  return quote.estimated_pickup || ''
}

function providerInitial(quote: MarketplaceQuote) {
  return quote.provider_name?.charAt(0)?.toUpperCase() || 'P'
}
</script>

<template>
  <div class="site-panel w-full max-w-[500px] rounded-[1.4rem] border p-5 sm:p-6">
    <div class="flex items-start justify-between gap-3">
      <div>
        <p class="text-[10px] font-bold tracking-[0.17em] text-ink-900/60 uppercase">Live provider rates</p>
        <h2 class="mt-1.5 text-[18px] font-semibold tracking-[-0.03em]">Compare at a glance</h2>
      </div>
      <span
        class="inline-flex items-center gap-1 rounded-full border px-2.5 py-1 text-[9px] font-bold tracking-[0.12em] uppercase"
        :class="
          loading
            ? 'border-accent-200 bg-accent-50 text-accent-800 dark:border-accent-500/25 dark:bg-accent-500/10 dark:text-accent-300'
            : 'border-success-200 bg-success-50 text-success-700 dark:border-success-500/25 dark:bg-success-500/10 dark:text-success-700'
        "
      >
        <span
          class="size-1.5 rounded-full"
          :class="loading ? 'animate-pulse bg-accent-500' : 'bg-success-500'"
        />
        {{ loading ? 'Checking' : 'Live' }}
      </span>
    </div>

    <div v-if="loading" class="flex flex-col items-center gap-3 py-12 text-center">
      <Loader2 :size="26" class="animate-spin text-accent-600 dark:text-accent-400" />
      <p class="text-[13px] font-semibold">Checking available providers…</p>
      <p v-if="search" class="max-w-[280px] text-[11.5px] leading-relaxed text-ink-900/50">
        Contacting providers with active routes
        {{ search.origin_label || 'your pickup zone' }} → {{ search.destination_label || 'your destination zone' }}.
        This usually takes a few seconds.
      </p>
    </div>

    <div v-else-if="error" class="flex flex-col items-center gap-3 py-10 text-center">
      <span class="flex size-10 items-center justify-center rounded-2xl border border-red-200 bg-red-50 dark:border-red-400/25 dark:bg-red-400/10">
        <Wifi :size="18" class="text-red-600 dark:text-red-300" />
      </span>
      <p class="text-[13px] font-semibold">Search failed</p>
      <p class="max-w-[320px] text-[11.5px] leading-relaxed text-ink-900/55">{{ error }}</p>
    </div>

    <template v-else-if="response && search">
      <div class="mt-4 flex items-center justify-between gap-3 rounded-xl border border-ink-900/8 bg-canvas px-3.5 py-3 dark:border-white/10 dark:bg-white/5">
        <div class="min-w-0">
          <p class="text-[9px] font-semibold tracking-[0.12em] text-ink-900/60 uppercase">
            {{ search.service_mode === 'dropoff' ? 'Drop-off route' : 'Pickup route' }}
          </p>
          <p class="mt-0.5 truncate text-[13px] font-semibold">
            {{ search.origin_label || (search.service_mode === 'dropoff' ? 'Drop-off hub zone' : 'Pickup zone') }}
            <span class="px-1 text-accent-600 dark:text-accent-400">→</span>
            {{ search.destination_label || 'Destination zone' }}
          </p>
        </div>
        <div class="shrink-0 text-right">
          <p class="text-[9px] font-semibold tracking-[0.12em] text-ink-900/60 uppercase">Package</p>
          <p class="mt-0.5 text-[12px] font-medium text-ink-900/70">
            {{ search.weight_kg }} kg · {{ categories[search.category ?? 'other'] ?? 'Other' }}
          </p>
        </div>
      </div>

      <div v-if="totalQuotes" class="mt-4 flex items-center gap-1 rounded-xl border border-ink-900/8 bg-canvas p-1 text-[10.5px] font-semibold">
        <button
          v-for="option in ([
            { id: 'cheapest', label: 'Cheapest' },
            { id: 'rating', label: 'Top rated' },
            { id: 'pickup', label: 'Pickup' },
            { id: 'delivery', label: 'Delivery' },
          ] as const)"
          :key="option.id"
          type="button"
          class="flex-1 rounded-lg px-2 py-1.5 transition"
          :class="sortMode === option.id ? 'bg-surface text-ink-900 shadow-sm' : 'text-ink-900/50'"
          @click="sortMode = option.id"
        >
          {{ option.label }}
        </button>
      </div>

      <div v-if="totalQuotes === 0" class="flex flex-col items-center gap-2.5 py-10 text-center">
        <IconTile tone="neutral" rounded="2xl">
          <Truck :size="18" />
        </IconTile>
        <p class="text-[13px] font-semibold">
          No provider routes this lane yet
        </p>
        <p class="max-w-[300px] text-[11.5px] leading-relaxed text-ink-900/50">
          {{ response.meta.providers_considered }} provider(s) were checked.
          Try a nearby destination zone or check back after providers publish this route.
        </p>
      </div>

      <template v-else>
        <ul class="mt-3 space-y-2">
          <li v-for="quote in visibleQuotes" :key="quote.quote_reference">
            <component
              :is="quote.booking_slug ? 'a' : 'div'"
              :href="bookingHref(quote)"
              :target="quote.booking_slug ? '_blank' : undefined"
              :rel="quote.booking_slug ? 'noopener noreferrer' : undefined"
              :aria-label="
                quote.booking_slug
                  ? `Book ${quote.provider_name} in a new tab`
                  : undefined
              "
              class="block rounded-xl border border-ink-900/8 bg-canvas/60 p-3.5 transition"
              :class="
                quote.booking_slug
                  ? 'cursor-pointer hover:-translate-y-0.5 hover:border-accent-500/50 hover:bg-accent-50/40 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent-500 dark:hover:bg-white/5'
                  : ''
              "
            >
              <div class="flex items-start justify-between gap-3">
                <div class="flex min-w-0 items-start gap-2.5">
                  <span class="flex size-9 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-surface text-[12px] font-bold text-ink-900 shadow-sm ring-1 ring-ink-900/8">
                    <img
                      v-if="quote.provider_logo"
                      :src="quote.provider_logo"
                      :alt="`${quote.provider_name} logo`"
                      class="size-full object-cover"
                      loading="lazy"
                    />
                    <span v-else>{{ providerInitial(quote) }}</span>
                  </span>
                  <div class="min-w-0">
                    <div class="flex items-center gap-1.5">
                      <p class="truncate text-[13px] font-semibold">{{ quote.provider_name }}</p>
                      <span
                        v-if="quote.provider_rating > 0"
                        class="inline-flex shrink-0 items-center gap-0.5 text-[9.5px] font-semibold text-accent-700 dark:text-accent-400"
                      >
                        <Star :size="9" class="fill-accent-500 text-accent-500" />
                        {{ quote.provider_rating.toFixed(1) }}
                      </span>
                    </div>
                    <div class="mt-1 flex flex-wrap items-center gap-1.5">
                      <span
                        class="rounded-full px-1.5 py-0.5 text-[9.5px] font-semibold"
                        :class="serviceTone(quote.service_key).pill"
                      >
                        {{ quote.service_name }}
                      </span>
                      <span
                        v-if="quote.price_confirmation_type === 'estimated'"
                        class="rounded-full border border-accent-200 bg-accent-50 px-1.5 py-0.5 text-[9.5px] font-semibold text-accent-800 dark:border-accent-500/25 dark:bg-accent-500/10 dark:text-accent-300"
                      >
                        Estimate
                      </span>
                    </div>
                  </div>
                </div>
                <p class="whitespace-nowrap text-[15px] font-bold tracking-[-0.03em] tabular-nums">
                  {{ formatMinor(quote.total_price_minor, quote.currency) }}
                </p>
              </div>

              <div class="mt-2.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-[10.5px] font-medium text-ink-900/60">
                <span v-if="pickupLabel(quote)" class="inline-flex items-center gap-1">
                  <PackageCheck :size="11" /> {{ pickupLabel(quote) }}
                </span>
                <span class="inline-flex items-center gap-1">
                  <Clock3 :size="11" /> {{ etaLabel(quote) }}
                </span>
                <span v-if="quote.service_mode === 'dropoff'" class="inline-flex min-w-0 items-center gap-1">
                  <MapPin :size="11" class="shrink-0" />
                  <span class="truncate">
                    Hub:
                    {{ quote.dropoff_branch_name || quote.origin_zone_name || quote.origin_city }}, {{ quote.origin_state }}<template v-if="quote.dropoff_branch_address"> — {{ quote.dropoff_branch_address }}</template>
                  </span>
                </span>
                <span v-else class="inline-flex items-center gap-1">
                  <MapPin :size="11" /> {{ handlingLabel(quote) }}
                </span>
                <span
                  v-if="quote.booking_slug"
                  class="ml-auto inline-flex items-center gap-1 font-semibold text-accent-700 dark:text-accent-400"
                >
                  Book
                  <ArrowUpRight :size="11" />
                </span>
              </div>
            </component>
          </li>
        </ul>

        <div class="mt-3 flex items-center justify-between gap-2 text-[10px] text-ink-900/45">
          <span>
            {{ response.meta.providers_quoted }}
            {{ response.meta.providers_quoted === 1 ? 'provider' : 'providers' }} responded
          </span>
          <span v-if="totalQuotes > visibleQuotes.length">
            Showing {{ visibleQuotes.length }} of {{ totalQuotes }} options
          </span>
        </div>
      </template>
    </template>

    <div v-else class="flex flex-col items-center gap-2.5 py-10 text-center">
      <IconTile tone="neutral" rounded="2xl">
        <Search :size="18" />
      </IconTile>
      <p class="text-[13px] font-semibold">Search a route to see live prices</p>
      <p class="max-w-[280px] text-[11.5px] leading-relaxed text-ink-900/50">
        We'll check the rates published by logistics providers serving your route. No sample
        prices, no made-up estimates.
      </p>
    </div>

    <p class="mt-4 flex items-start gap-2 border-t border-ink-900/8 pt-3 text-[10px] leading-relaxed text-ink-900/50 dark:border-white/10">
      <ShieldCheck :size="13" class="mt-0.5 shrink-0 text-success-600" />
      <span v-if="response && totalQuotes > 0">
        Prices come from each provider’s directional zone route.
        Selecting a provider opens their booking page in a new tab.
      </span>
      <span v-else>
        Prices come from each provider’s published zone routes, never from sample data.
      </span>
    </p>
  </div>
</template>
