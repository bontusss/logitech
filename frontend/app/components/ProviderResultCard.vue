<script setup lang="ts">
import { ArrowUpRight, ChevronDown, Loader2, MapPin, PackageCheck, Truck } from 'lucide-vue-next'
import type { MarketplaceQuote } from '~/types/marketplace'

const props = withDefaults(
  defineProps<{
    quote: MarketplaceQuote
    busy?: boolean
    selected?: boolean
  }>(),
  { busy: false, selected: false },
)

defineEmits<{ select: [quote: MarketplaceQuote] }>()

const detailsOpen = ref(false)
const tone = computed(() => serviceTone(props.quote.service_key))
const initial = computed(() => props.quote.provider_name?.charAt(0)?.toUpperCase() || 'P')
</script>

<template>
  <article
    class="site-panel overflow-hidden rounded-2xl border transition hover:-translate-y-0.5 hover:shadow-[0_18px_40px_-28px_rgba(15,23,42,0.4)]"
    :class="selected ? 'border-brand-400 ring-2 ring-brand-100 dark:ring-brand-400/20' : ''"
  >
    <div class="p-4 sm:p-5">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div class="flex min-w-0 items-start gap-3">
          <div
            class="flex size-11 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-ink-900/8 bg-gradient-to-br from-slate-50 to-white text-[15px] font-bold text-brand-700 dark:border-white/10 dark:from-white/10 dark:to-transparent"
          >
            <img
              v-if="quote.provider_logo"
              :src="quote.provider_logo"
              :alt="`${quote.provider_name} logo`"
              class="size-full object-cover"
              loading="lazy"
            />
            <span v-else>{{ initial }}</span>
          </div>
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-1.5">
              <NuxtLink
                v-if="quote.booking_slug"
                :to="`/book/${quote.booking_slug}`"
                target="_blank"
                rel="noopener noreferrer"
                class="truncate text-[14.5px] font-semibold tracking-[-0.02em] hover:text-brand-700"
              >
                {{ quote.provider_name }}
              </NuxtLink>
              <span v-else class="truncate text-[14.5px] font-semibold tracking-[-0.02em]">
                {{ quote.provider_name }}
              </span>
              <NuxtLink
                v-if="quote.booking_slug"
                :to="`/book/${quote.booking_slug}`"
                target="_blank"
                rel="noopener noreferrer"
                class="text-ink-900/30 transition hover:text-brand-600"
                :aria-label="`Open ${quote.provider_name} booking page in a new tab`"
              >
                <ArrowUpRight :size="13" />
              </NuxtLink>
            </div>
            <div class="mt-1 flex flex-wrap items-center gap-1.5">
              <span class="rounded-full px-2 py-0.5 text-[10px] font-bold" :class="tone.pill">
                {{ quote.service_name }}
              </span>
              <PriceConfirmationBadge
                :type="quote.price_confirmation_type"
                :note="quote.price_confirmation_note"
              />
              <QuoteExpiryIndicator :expires-at="quote.expires_at" />
            </div>
          </div>
        </div>
        <div class="text-right">
          <p class="text-[21px] font-bold tracking-[-0.04em] tabular-nums">
            {{ formatMinor(quote.total_price_minor, quote.currency) }}
          </p>
          <p class="text-[10px] font-medium text-ink-900/45">Total price</p>
        </div>
      </div>

      <dl class="mt-3.5 grid grid-cols-2 gap-2 sm:grid-cols-4">
        <div class="rounded-xl bg-slate-50/80 px-2.5 py-2 dark:bg-white/5">
          <dt class="flex items-center gap-1 text-[9.5px] font-bold tracking-[0.1em] text-ink-900/40 uppercase">
            <MapPin :size="10" /> Route
          </dt>
          <dd class="mt-0.5 truncate text-[11.5px] font-semibold">
            {{ quote.origin_city }} → {{ quote.destination_city }}
          </dd>
        </div>
        <div class="rounded-xl bg-slate-50/80 px-2.5 py-2 dark:bg-white/5">
          <dt class="flex items-center gap-1 text-[9.5px] font-bold tracking-[0.1em] text-ink-900/40 uppercase">
            <Truck :size="10" /> Delivery
          </dt>
          <dd class="mt-0.5 truncate text-[11.5px] font-semibold">
            {{ quote.estimated_delivery || `${quote.estimated_duration_days} day(s)` }}
          </dd>
        </div>
        <div class="rounded-xl bg-slate-50/80 px-2.5 py-2 dark:bg-white/5">
          <dt class="flex items-center gap-1 text-[9.5px] font-bold tracking-[0.1em] text-ink-900/40 uppercase">
            <PackageCheck :size="10" /> Handling
          </dt>
          <dd class="mt-0.5 truncate text-[11.5px] font-semibold">
            {{ deliveryMethodLabel(quote.delivery_method) }}
          </dd>
        </div>
        <div class="rounded-xl bg-slate-50/80 px-2.5 py-2 dark:bg-white/5">
          <dt class="text-[9.5px] font-bold tracking-[0.1em] text-ink-900/40 uppercase">Distance</dt>
          <dd class="mt-0.5 truncate text-[11.5px] font-semibold">{{ quote.distance_km }} km</dd>
        </div>
      </dl>

      <div class="mt-3 flex flex-wrap items-center justify-between gap-2">
        <button
          type="button"
          class="inline-flex items-center gap-1 text-[11px] font-semibold text-ink-900/55 transition hover:text-brand-700"
          @click="detailsOpen = !detailsOpen"
        >
          Price details
          <ChevronDown :size="13" class="transition" :class="detailsOpen ? 'rotate-180' : ''" />
        </button>
        <button
          type="button"
          :disabled="busy || selected"
          class="inline-flex items-center gap-1.5 rounded-xl bg-accent-500 px-3.5 py-2 text-[11.5px] font-semibold text-white transition hover:bg-accent-400 disabled:opacity-60"
          @click="$emit('select', quote)"
        >
          <Loader2 v-if="busy" :size="13" class="animate-spin" />
          {{ selected ? 'Selected' : busy ? 'Saving…' : 'Select this quote' }}
        </button>
      </div>

      <div v-if="detailsOpen" class="mt-3 space-y-3">
        <QuotePriceBreakdown :quote="quote" />
        <div v-if="quote.conditions" class="rounded-xl border border-ink-900/8 bg-slate-50/70 px-3 py-2 text-[11px] leading-relaxed text-ink-900/55 dark:border-white/10 dark:bg-white/5">
          <span class="font-semibold text-ink-900/70">Provider conditions: </span>{{ quote.conditions }}
        </div>
        <p v-if="quote.price_confirmation_note" class="text-[10.5px] leading-relaxed text-ink-900/45">
          {{ quote.price_confirmation_note }}
        </p>
      </div>
    </div>
  </article>
</template>
