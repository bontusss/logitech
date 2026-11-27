<script setup lang="ts">
import {
  ArrowRight,
  CalendarClock,
  CircleAlert,
  Loader2,
  MapPin,
  PackageSearch,
  Repeat2,
  Share2,
  Star,
  Trash2,
  Truck,
} from 'lucide-vue-next'
import type {
  MarketplaceQuote,
  MarketplaceSearchPayload,
  MarketplaceSearchResponse,
} from '~/types/marketplace'
import type {
  CustomerFavouriteProvider,
  CustomerRecurringShipment,
  CustomerSavedAddress,
  RecurringFrequency,
} from '~/types/customer'

useHead({ title: 'Find Providers · BOL' })

const { api } = useApi()
const { notify, confirm } = useWorkspace()

const addresses = ref<CustomerSavedAddress[]>([])
const favourites = ref<Set<string>>(new Set())
const favouriteProviders = ref<CustomerFavouriteProvider[]>([])
const recurring = ref<CustomerRecurringShipment[]>([])

const searchLoading = ref(false)
const searchError = ref('')
const response = ref<MarketplaceSearchResponse | null>(null)
const searchPayload = ref<MarketplaceSearchPayload | null>(null)
const prefill = ref<MarketplaceSearchPayload | null>(null)
let requestId = 0

const pageError = ref('')
const pendingFavourite = ref('')
const recurringOpen = ref(false)
const recurringLabel = ref('')
const recurringFrequency = ref<RecurringFrequency>('monthly')
const savingRecurring = ref(false)
const recurringError = ref('')
const deletingRecurring = ref('')

const frequencyOptions: { value: RecurringFrequency; label: string }[] = [
  { value: 'weekly', label: 'Every week' },
  { value: 'biweekly', label: 'Every two weeks' },
  { value: 'monthly', label: 'Every month' },
]

const quotes = computed(() => response.value?.quotes ?? [])
const searchMeta = computed(() => response.value?.meta ?? null)

function formatMoney(minor: number, currency = 'NGN') {
  return new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  }).format(minor / 100)
}

function frequencyLabel(value: RecurringFrequency) {
  return frequencyOptions.find((option) => option.value === value)?.label ?? 'Recurring'
}

function routeLabel(payload: MarketplaceSearchPayload) {
  const from = payload.origin_label || [payload.origin_lga, payload.origin_state].filter(Boolean).join(', ')
  const to = payload.destination_label || [payload.destination_lga, payload.destination_state].filter(Boolean).join(', ')
  return `${from || 'Pickup'} → ${to || 'Destination'}`
}

async function loadAddresses() {
  try {
    const result = await api<{ addresses: CustomerSavedAddress[] }>('/customer/addresses')
    addresses.value = result.addresses
  } catch {
    addresses.value = []
  }
}

async function loadFavourites() {
  try {
    const result = await api<{ providers: CustomerFavouriteProvider[] }>('/customer/favourites')
    favouriteProviders.value = result.providers
    favourites.value = new Set(result.providers.map((provider) => provider.id))
  } catch {
    favourites.value = new Set()
    favouriteProviders.value = []
  }
}

async function loadRecurring() {
  try {
    const result = await api<{ recurring_shipments: CustomerRecurringShipment[] }>(
      '/customer/recurring-shipments',
    )
    recurring.value = result.recurring_shipments
  } catch {
    recurring.value = []
  }
}

async function runSearch(payload: MarketplaceSearchPayload) {
  const currentRequest = ++requestId
  searchPayload.value = payload
  searchLoading.value = true
  searchError.value = ''
  response.value = null

  try {
    const result = await api<MarketplaceSearchResponse>('/marketplace/search', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
    if (currentRequest === requestId) response.value = result
  } catch (cause) {
    if (currentRequest === requestId) {
      searchError.value = apiErrorMessage(cause, 'The search could not be completed. Please try again.')
    }
  } finally {
    if (currentRequest === requestId) searchLoading.value = false
  }
}

function savedAddressFor(addressText?: string) {
  if (!addressText) return undefined
  const target = addressText.trim().toLowerCase()
  return addresses.value.find((address) => {
    const composed = [address.street_address, address.area, address.city, address.state]
      .filter(Boolean)
      .join(', ')
      .toLowerCase()
    return composed === target
  })
}

function bookingHref(quote: MarketplaceQuote) {
  if (!quote.booking_slug) return ''
  const params = new URLSearchParams()
  const search = searchPayload.value
  if (search) {
    const zoneParams: [keyof MarketplaceSearchPayload, string][] = [
      ['origin_zone_id', 'origin_zone_id'],
      ['destination_zone_id', 'destination_zone_id'],
      ['origin_lga', 'origin_lga'],
      ['origin_state', 'origin_state'],
      ['destination_lga', 'destination_lga'],
      ['destination_state', 'destination_state'],
      ['category', 'category'],
      ['service_mode', 'service_mode'],
    ]
    for (const [field, key] of zoneParams) {
      const value = search[field]
      if (typeof value === 'string' && value) params.set(key, value)
    }
    if (search.weight_kg) params.set('weight_kg', String(search.weight_kg))
    if (search.origin_address) params.set('pickup_address', search.origin_address)
    if (search.origin_latitude != null) params.set('pickup_latitude', String(search.origin_latitude))
    if (search.origin_longitude != null) params.set('pickup_longitude', String(search.origin_longitude))
    if (search.destination_address) params.set('dropoff_address', search.destination_address)
    if (search.destination_latitude != null) {
      params.set('dropoff_latitude', String(search.destination_latitude))
    }
    if (search.destination_longitude != null) {
      params.set('dropoff_longitude', String(search.destination_longitude))
    }

    const sender = savedAddressFor(search.origin_address)
    if (sender?.contact_name) params.set('sender_name', sender.contact_name)
    if (sender?.phone) params.set('sender_phone', sender.phone)
    const receiver = savedAddressFor(search.destination_address)
    if (receiver?.contact_name) params.set('receiver_name', receiver.contact_name)
    if (receiver?.phone) params.set('receiver_phone', receiver.phone)
  }
  const query = params.toString()
  return `/book/${encodeURIComponent(quote.booking_slug)}${query ? `?${query}` : ''}`
}

function shareOnWhatsApp(quote: MarketplaceQuote) {
  const link = bookingHref(quote)
  const currency = quote.currency || 'NGN'
  const summary = [
    `Shipment quote from ${quote.provider_name}`,
    `${quote.origin_city} → ${quote.destination_city}`,
    `Price: ${formatMoney(quote.total_price_minor, currency)}`,
    `Delivery: ${quote.estimated_delivery || `${quote.estimated_duration_days} day(s)`}`,
    quote.expires_at
      ? `Quote valid until ${new Date(quote.expires_at).toLocaleString('en-NG')}`
      : '',
    link ? `Book: ${window.location.origin}${link}` : '',
  ]
    .filter(Boolean)
    .join('\n')
  window.open(`https://wa.me/?text=${encodeURIComponent(summary)}`, '_blank', 'noopener')
}

async function toggleFavourite(providerId: string, providerName: string) {
  pendingFavourite.value = providerId
  const saved = favourites.value.has(providerId)
  try {
    if (saved) {
      await api(`/customer/favourites/${providerId}`, { method: 'DELETE' })
      favourites.value = new Set([...favourites.value].filter((id) => id !== providerId))
      favouriteProviders.value = favouriteProviders.value.filter((provider) => provider.id !== providerId)
      notify(`Removed ${providerName} from favourites.`)
    } else {
      await api(`/customer/favourites/${providerId}`, { method: 'POST' })
      favourites.value = new Set([...favourites.value, providerId])
      notify(`Saved ${providerName} to favourites.`)
      await loadFavourites()
    }
  } catch (cause) {
    pageError.value = apiErrorMessage(cause, 'Could not update your favourites.')
  } finally {
    pendingFavourite.value = ''
  }
}

function openRecurringForm() {
  recurringError.value = ''
  recurringLabel.value = searchPayload.value ? routeLabel(searchPayload.value) : ''
  recurringOpen.value = true
}

async function saveRecurring() {
  if (!searchPayload.value) return
  savingRecurring.value = true
  recurringError.value = ''
  try {
    const result = await api<{ recurring_shipment: CustomerRecurringShipment }>(
      '/customer/recurring-shipments',
      {
        method: 'POST',
        body: JSON.stringify({
          label: recurringLabel.value.trim(),
          frequency: recurringFrequency.value,
          search_payload: searchPayload.value,
        }),
      },
    )
    recurring.value = [result.recurring_shipment, ...recurring.value]
    recurringOpen.value = false
    notify('Repeating shipment saved.')
  } catch (cause) {
    recurringError.value = apiErrorMessage(cause, 'Could not save this repeating shipment.')
  } finally {
    savingRecurring.value = false
  }
}

function repeatRecurring(shipment: CustomerRecurringShipment) {
  prefill.value = { ...shipment.search_payload }
  void runSearch({ ...shipment.search_payload })
  document.getElementById('find-providers-search')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

async function removeRecurring(shipment: CustomerRecurringShipment) {
  const approved = await confirm({
    title: 'Remove repeating shipment?',
    message: `“${shipment.label}” will be removed from your saved shipments.`,
    confirmLabel: 'Remove',
  })
  if (!approved) return
  deletingRecurring.value = shipment.id
  try {
    await api(`/customer/recurring-shipments/${shipment.id}`, { method: 'DELETE' })
    recurring.value = recurring.value.filter((row) => row.id !== shipment.id)
    notify('Repeating shipment removed.')
  } catch (cause) {
    pageError.value = apiErrorMessage(cause, 'Could not remove this repeating shipment.')
  } finally {
    deletingRecurring.value = ''
  }
}

onMounted(async () => {
  await Promise.all([loadAddresses(), loadFavourites(), loadRecurring()])
})
</script>

<template>
  <div class="dash-page">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Shipments</p>
        <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Find providers</h1>
        <p class="mt-1.5 max-w-2xl text-[12.5px] leading-relaxed text-ink-900/50">
          Compare live rates, pick a saved address instead of typing it again, and keep favourite providers and
          repeating shipments close at hand.
        </p>
      </div>
      <button
        v-if="searchPayload"
        type="button"
        class="inline-flex items-center gap-2 rounded-lg bg-accent-500 px-4 py-2.5 text-sm font-semibold text-white shadow-[0_10px_24px_-14px_rgba(255,122,0,0.75)] transition hover:-translate-y-0.5 hover:bg-accent-400"
        @click="openRecurringForm"
      >
        <Repeat2 :size="14" /> Save as repeating
      </button>
    </div>

    <p v-if="pageError" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ pageError }}
    </p>

    <div id="find-providers-search" class="mt-5 grid scroll-mt-24 gap-4 lg:grid-cols-[minmax(0,520px)_minmax(0,1fr)] lg:items-start">
      <ShippingRateSearch
        :loading="searchLoading"
        :saved-addresses="addresses"
        :prefill="prefill"
        heading="Compare providers"
        @search="runSearch"
      />

      <div class="space-y-3">
        <p v-if="recurringError" class="rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
          {{ recurringError }}
        </p>

        <form v-if="recurringOpen" class="rounded-xl border border-brand-200 bg-brand-50/50 p-3.5" @submit.prevent="saveRecurring">
          <div class="flex items-center justify-between gap-2">
            <p class="text-[12px] font-semibold">Save this shipment to repeat</p>
            <button type="button" class="text-[11px] font-semibold text-ink-900/45 hover:text-ink-900/70" @click="recurringOpen = false">
              Cancel
            </button>
          </div>
          <p class="mt-1 text-[10.5px] leading-relaxed text-ink-900/50">
            We keep this route, saved address and package. Repeating it re-runs the search so prices are current.
          </p>
          <div class="mt-2.5 grid gap-2 sm:grid-cols-[minmax(0,1fr)_150px_auto]">
            <input
              v-model="recurringLabel"
              maxlength="80"
              placeholder="Name, e.g. Weekly documents to Abuja"
              class="w-full rounded-lg border border-ink-900/12 bg-surface px-3 py-2 text-[11.5px] outline-none focus:border-accent-500"
            />
            <select v-model="recurringFrequency" class="rounded-lg border border-ink-900/12 bg-surface px-3 py-2 text-[11.5px] outline-none focus:border-accent-500">
              <option v-for="option in frequencyOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
            <button
              type="submit"
              :disabled="savingRecurring"
              class="inline-flex items-center justify-center gap-1.5 rounded-lg bg-accent-500 px-3.5 py-2 text-[11.5px] font-semibold text-white transition hover:bg-accent-400 disabled:opacity-60"
            >
              <Loader2 v-if="savingRecurring" :size="13" class="animate-spin" />
              Save
            </button>
          </div>
        </form>

        <div v-if="searchLoading" class="flex items-center justify-center gap-2 rounded-xl border border-ink-900/10 bg-surface p-12 text-[12px] text-ink-900/45">
          <Loader2 :size="15" class="animate-spin" /> Checking provider routes…
        </div>

        <div v-else-if="searchError" class="rounded-xl border border-red-200 bg-red-50 px-4 py-5 text-[12px] text-red-700">
          <p class="flex items-start gap-2"><CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ searchError }}</p>
        </div>

        <div v-else-if="!quotes.length" class="rounded-xl border border-dashed border-ink-900/15 bg-surface/60 px-5 py-10 text-center">
          <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
            <PackageSearch :size="20" />
          </span>
          <h2 class="mt-3 text-[13.5px] font-semibold">Live rates appear here</h2>
          <p class="mx-auto mt-1 max-w-sm text-[11.5px] leading-relaxed text-ink-900/45">
            Choose a pickup and destination (or a saved address) to see every provider with an active route,
            their price and delivery estimate.
          </p>
        </div>

        <template v-else>
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div>
              <p class="text-[10px] font-bold tracking-[0.16em] text-ink-900/45 uppercase">Live provider rates</p>
              <p class="mt-0.5 text-[12px] font-semibold">
                {{ quotes.length }} quote{{ quotes.length === 1 ? '' : 's' }} for
                {{ searchPayload ? routeLabel(searchPayload) : 'this route' }}
              </p>
            </div>
            <span v-if="searchMeta?.partial" class="rounded-full border border-amber-200 bg-amber-50 px-2.5 py-1 text-[10px] font-semibold text-amber-800">
              Some providers timed out
            </span>
          </div>

          <ul class="space-y-2.5">
            <li v-for="quote in quotes" :key="quote.quote_reference" class="rounded-xl border border-ink-900/10 bg-surface p-4">
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div class="flex min-w-0 items-start gap-3">
                  <span class="flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-ink-900/8 bg-gradient-to-br from-slate-50 to-white text-[13px] font-bold text-brand-700 dark:border-white/10 dark:from-white/10 dark:to-transparent">
                    <img v-if="quote.provider_logo" :src="quote.provider_logo" :alt="`${quote.provider_name} logo`" class="size-full object-cover" loading="lazy" />
                    <span v-else>{{ quote.provider_name.charAt(0).toUpperCase() }}</span>
                  </span>
                  <div class="min-w-0">
                    <div class="flex flex-wrap items-center gap-1.5">
                      <p class="truncate text-[13px] font-semibold">{{ quote.provider_name }}</p>
                      <span class="rounded-full bg-slate-100 px-2 py-0.5 text-[9.5px] font-bold text-ink-900/60 dark:bg-white/10">
                        {{ quote.service_name }}
                      </span>
                    </div>
                    <p class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-[10.5px] text-ink-900/50">
                      <span class="inline-flex items-center gap-1"><MapPin :size="10" />{{ quote.origin_city }} → {{ quote.destination_city }}</span>
                      <span class="inline-flex items-center gap-1"><Truck :size="10" />{{ quote.estimated_delivery || `${quote.estimated_duration_days} day(s)` }}</span>
                    </p>
                  </div>
                </div>
                <div class="text-right">
                  <p class="text-[18px] font-bold tracking-[-0.03em] tabular-nums">{{ formatMoney(quote.total_price_minor, quote.currency) }}</p>
                  <p class="text-[9.5px] font-medium text-ink-900/40">{{ quote.price_confirmation_type === 'estimated' ? 'Estimated total' : 'Confirmed total' }}</p>
                </div>
              </div>

              <div class="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-ink-900/8 pt-3">
                <div class="flex items-center gap-1">
                  <button
                    type="button"
                    :disabled="pendingFavourite === quote.provider_id"
                    class="rounded-lg p-1.5 transition disabled:opacity-50"
                    :class="favourites.has(quote.provider_id) ? 'text-amber-600 hover:bg-amber-50' : 'text-ink-900/30 hover:bg-canvas hover:text-amber-600'"
                    :aria-label="favourites.has(quote.provider_id) ? `Remove ${quote.provider_name} from favourites` : `Save ${quote.provider_name} to favourites`"
                    @click="toggleFavourite(quote.provider_id, quote.provider_name)"
                  >
                    <Loader2 v-if="pendingFavourite === quote.provider_id" :size="15" class="animate-spin" />
                    <Star v-else :size="15" :class="favourites.has(quote.provider_id) ? 'fill-amber-500' : ''" />
                  </button>
                  <button
                    type="button"
                    class="inline-flex items-center gap-1.5 rounded-lg px-2 py-1.5 text-[11px] font-semibold text-ink-900/55 transition hover:bg-canvas hover:text-emerald-700"
                    @click="shareOnWhatsApp(quote)"
                  >
                    <Share2 :size="13" /> WhatsApp
                  </button>
                </div>
                <NuxtLink
                  v-if="quote.booking_slug"
                  :to="bookingHref(quote)"
                  class="inline-flex items-center gap-1.5 rounded-xl bg-accent-500 px-3.5 py-2 text-[11.5px] font-semibold text-white transition hover:bg-accent-400"
                >
                  Book with {{ quote.provider_name }} <ArrowRight :size="13" />
                </NuxtLink>
                <span v-else class="text-[10.5px] text-ink-900/40">Booking page not published</span>
              </div>
            </li>
          </ul>
        </template>
      </div>
    </div>

    <section v-if="recurring.length" class="mt-6">
      <div class="flex items-center gap-2">
        <Repeat2 :size="15" class="text-brand-700" />
        <h2 class="text-[13px] font-semibold">Repeating shipments</h2>
      </div>
      <p class="mt-1 text-[11.5px] text-ink-900/45">
        Re-run a saved shipment to compare current prices, then book in a couple of clicks.
      </p>
      <ul class="mt-3 grid gap-2.5 sm:grid-cols-2 xl:grid-cols-3">
        <li v-for="shipment in recurring" :key="shipment.id" class="rounded-xl border border-ink-900/10 bg-surface p-3.5">
          <div class="flex items-start justify-between gap-2">
            <div class="min-w-0">
              <p class="truncate text-[12px] font-semibold">{{ shipment.label }}</p>
              <p class="mt-0.5 flex items-center gap-1 text-[10px] font-semibold text-brand-700">
                <CalendarClock :size="10" /> {{ frequencyLabel(shipment.frequency) }}
              </p>
            </div>
            <button
              type="button"
              :disabled="deletingRecurring === shipment.id"
              class="rounded-lg p-1.5 text-ink-900/30 transition hover:bg-red-50 hover:text-red-600 disabled:opacity-50"
              :aria-label="`Remove ${shipment.label}`"
              @click="removeRecurring(shipment)"
            >
              <Loader2 v-if="deletingRecurring === shipment.id" :size="13" class="animate-spin" />
              <Trash2 v-else :size="13" />
            </button>
          </div>
          <p class="mt-2 line-clamp-2 text-[10.5px] leading-relaxed text-ink-900/45">
            {{ routeLabel(shipment.search_payload) }}
          </p>
          <button
            type="button"
            class="mt-2.5 inline-flex items-center gap-1 text-[11px] font-semibold text-brand-700 hover:text-brand-600"
            @click="repeatRecurring(shipment)"
          >
            Compare current rates <ArrowRight :size="12" />
          </button>
        </li>
      </ul>
    </section>

    <section v-if="favouriteProviders.length" class="mt-6">
      <div class="flex items-center gap-2">
        <Star :size="15" class="fill-amber-500 text-amber-500" />
        <h2 class="text-[13px] font-semibold">Favourite providers</h2>
      </div>
      <ul class="mt-3 grid gap-2.5 sm:grid-cols-2 xl:grid-cols-3">
        <li v-for="provider in favouriteProviders" :key="provider.id" class="flex items-center gap-3 rounded-xl border border-ink-900/10 bg-surface p-3.5">
          <span class="flex size-9 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-ink-900/8 bg-gradient-to-br from-slate-50 to-white text-[12px] font-bold text-brand-700 dark:border-white/10 dark:from-white/10 dark:to-transparent">
            <img v-if="provider.logo_url" :src="provider.logo_url" :alt="`${provider.name} logo`" class="size-full object-cover" />
            <span v-else>{{ provider.name.charAt(0).toUpperCase() }}</span>
          </span>
          <div class="min-w-0 flex-1">
            <p class="truncate text-[12px] font-semibold">{{ provider.name }}</p>
            <p class="text-[10px] text-ink-900/45">{{ provider.route_count }} routes</p>
          </div>
          <NuxtLink :to="`/book/${provider.slug}`" class="text-[11px] font-semibold text-brand-700 hover:text-brand-600">
            Book
          </NuxtLink>
        </li>
      </ul>
    </section>
  </div>
</template>
