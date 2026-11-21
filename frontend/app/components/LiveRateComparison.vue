<script setup lang="ts">
import {
  ArrowUpRight,
  Clock3,
  Loader2,
  MapPin,
  Search,
  ShieldCheck,
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

const visibleQuotes = computed<MarketplaceQuote[]>(() => props.response?.quotes.slice(0, 3) ?? [])
const totalQuotes = computed(() => props.response?.quotes.length ?? 0)

const categories: Record<string, string> = {
  document: 'Documents',
  cloth: 'Clothing',
  electronics: 'Electronics',
  food: 'Food',
  jewelry: 'Jewelry',
  other: 'Other',
}
const isLastMile = computed(() => props.search?.service_type === 'last_mile')

function bookingHref(quote: MarketplaceQuote) {
  return quote.booking_slug ? `/book/${encodeURIComponent(quote.booking_slug)}` : undefined
}

function etaLabel(quote: MarketplaceQuote) {
  return quote.estimated_delivery || `${quote.estimated_duration_days} day(s)`
}

function handlingLabel(quote: MarketplaceQuote) {
  return deliveryMethodLabel(quote.delivery_method)
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
            ? 'border-amber-200/80 bg-amber-50 text-amber-800 dark:border-amber-400/20 dark:bg-amber-400/10 dark:text-amber-200'
            : 'border-emerald-200/80 bg-emerald-50 text-emerald-700 dark:border-emerald-400/20 dark:bg-emerald-400/10 dark:text-emerald-200'
        "
      >
        <span
          class="size-1.5 rounded-full"
          :class="loading ? 'animate-pulse bg-amber-500' : 'bg-emerald-500'"
        />
        {{ loading ? 'Checking' : 'Live' }}
      </span>
    </div>

    <div v-if="loading" class="flex flex-col items-center gap-3 py-12 text-center">
      <Loader2 :size="26" class="animate-spin text-brand-600" />
      <p class="text-[13px] font-semibold">Checking available providers…</p>
      <p v-if="search" class="max-w-[280px] text-[11.5px] leading-relaxed text-ink-900/50">
        Contacting {{ isLastMile ? 'last-mile providers covering' : 'providers serving' }}
        {{ search.origin_city }} → {{ search.destination_city }}.
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
          <p class="text-[9px] font-semibold tracking-[0.12em] text-ink-900/60 uppercase">{{ isLastMile ? 'Local route' : 'Route' }}</p>
          <p class="mt-0.5 truncate text-[13px] font-semibold">
            {{ search.origin_city }}
            <span class="bg-gradient-to-r from-brand-500 to-violet-500 bg-clip-text px-1 text-transparent">→</span>
            {{ search.destination_city }}
          </p>
          <p v-if="isLastMile && search.distance_km" class="mt-0.5 text-[10px] text-ink-900/45">
            {{ search.distance_km.toLocaleString() }} km verified distance
          </p>
        </div>
        <div class="shrink-0 text-right">
          <p class="text-[9px] font-semibold tracking-[0.12em] text-ink-900/60 uppercase">Package</p>
          <p class="mt-0.5 text-[12px] font-medium text-ink-900/70">
            {{ search.weight_kg }} kg · {{ categories[search.category ?? 'other'] ?? 'Other' }}
          </p>
        </div>
      </div>

      <div v-if="totalQuotes === 0" class="flex flex-col items-center gap-2.5 py-10 text-center">
        <span class="flex size-10 items-center justify-center rounded-2xl bg-gradient-to-br from-slate-50 to-white text-ink-900/35 dark:from-white/10 dark:to-transparent">
          <Truck :size="18" />
        </span>
        <p class="text-[13px] font-semibold">
          {{ isLastMile ? 'No last-mile providers cover this route yet' : 'No providers found for this route yet' }}
        </p>
        <p class="max-w-[300px] text-[11.5px] leading-relaxed text-ink-900/50">
          {{ response.meta.providers_considered }} provider(s) were checked.
          {{ isLastMile
            ? 'Try another nearby drop-off, or check back after providers activate coverage in this area.'
            : 'Try a nearby city pair or check back after providers publish this lane.' }}
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
              class="block rounded-xl border border-ink-900/8 bg-gradient-to-br from-white to-slate-50/60 p-3.5 transition dark:border-white/10 dark:from-white/5 dark:to-transparent"
              :class="
                quote.booking_slug
                  ? 'cursor-pointer hover:-translate-y-0.5 hover:border-brand-300 hover:shadow-[0_12px_26px_-20px_rgba(43,75,255,0.7)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-500'
                  : ''
              "
            >
              <div class="flex items-start justify-between gap-3">
                <div class="flex min-w-0 items-start gap-2.5">
                  <span class="flex size-9 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-white text-[12px] font-bold text-brand-700 shadow-sm ring-1 ring-ink-900/8 dark:bg-white/10 dark:text-indigo-200 dark:ring-white/10">
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
                    <p class="truncate text-[13px] font-semibold">{{ quote.provider_name }}</p>
                    <div class="mt-1 flex flex-wrap items-center gap-1.5">
                      <span
                        class="rounded-full px-1.5 py-0.5 text-[9.5px] font-semibold"
                        :class="serviceTone(quote.service_key).pill"
                      >
                        {{ quote.service_name }}
                      </span>
                      <span
                        v-if="quote.price_confirmation_type === 'estimated'"
                        class="rounded-full border border-amber-200 bg-amber-50 px-1.5 py-0.5 text-[9.5px] font-semibold text-amber-800 dark:border-amber-400/25 dark:bg-amber-400/10 dark:text-amber-200"
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
                <span class="inline-flex items-center gap-1">
                  <Clock3 :size="11" /> {{ etaLabel(quote) }}
                </span>
                <span class="inline-flex items-center gap-1">
                  <MapPin :size="11" /> {{ handlingLabel(quote) }}
                </span>
                <span
                  v-if="quote.booking_slug"
                  class="ml-auto inline-flex items-center gap-1 font-semibold text-brand-700 dark:text-indigo-200"
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
      <span class="flex size-10 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-50 to-violet-50 text-brand-700 dark:from-brand-400/15 dark:to-violet-400/10 dark:text-indigo-200">
        <Search :size="18" />
      </span>
      <p class="text-[13px] font-semibold">Search a route to see live prices</p>
      <p class="max-w-[280px] text-[11.5px] leading-relaxed text-ink-900/50">
        We'll check the rates published by logistics providers serving your route. No sample
        prices, no made-up estimates.
      </p>
    </div>

    <p class="mt-4 flex items-start gap-2 border-t border-ink-900/8 pt-3 text-[10px] leading-relaxed text-ink-900/50 dark:border-white/10">
      <ShieldCheck :size="13" class="mt-0.5 shrink-0 text-emerald-700 dark:text-emerald-300" />
      <span v-if="response && totalQuotes > 0">
        {{ isLastMile ? 'Local prices use the Mapbox road distance and each provider’s active coverage.' : 'Prices come from each provider’s published rates.' }}
        Selecting a provider opens their booking page in a new tab.
      </span>
      <span v-else>
        {{ isLastMile ? 'Local prices use verified Mapbox addresses and provider coverage, never sample data.' : 'Prices come from each provider’s published rates, never from sample data.' }}
      </span>
    </p>
  </div>
</template>
