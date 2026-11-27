<script setup lang="ts">
import {
  ArrowRight,
  CircleAlert,
  Loader2,
  MapPin,
  Package,
  Search,
  Truck,
} from 'lucide-vue-next'
import type { CustomerBooking, DeliveryStatus, TrackingResult } from '~/types/customer'

useHead({ title: 'Track Shipment · BOL' })

const route = useRoute()
const { api } = useApi()
const { bookings, load: loadBookings } = useCustomerDashboard()

const reference = ref(typeof route.query.code === 'string' ? route.query.code : '')
const loading = ref(false)
const error = ref('')
const result = ref<TrackingResult | null>(null)
const matchedBooking = ref<CustomerBooking | null>(null)
const searched = ref(false)

onMounted(async () => {
  await loadBookings()
  if (reference.value) void track()
})

const tracked = computed(() =>
  bookings.value.filter((booking) => booking.tracking_code && !['delivered', 'cancelled'].includes(booking.status)).slice(0, 4),
)

function statusTone(status: DeliveryStatus | string) {
  if (status === 'delivered') return 'border-emerald-200 bg-emerald-50 text-emerald-700'
  if (status === 'cancelled') return 'border-rose-200 bg-rose-50 text-rose-700'
  if (status === 'in_transit' || status === 'picked_up') return 'border-brand-200 bg-brand-50 text-brand-700'
  if (status === 'assigned') return 'border-sky-200 bg-sky-50 text-sky-700'
  return 'border-slate-200 bg-slate-50 text-slate-600'
}

function statusLabel(status: string) {
  return status.replaceAll('_', ' ').replace(/\b\w/g, (letter) => letter.toUpperCase())
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleString('en-NG', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function track() {
  const raw = reference.value.trim()
  if (!raw) {
    error.value = 'Enter a tracking code or booking reference.'
    return
  }
  loading.value = true
  searched.value = true
  error.value = ''
  result.value = null
  matchedBooking.value = null
  try {
    const bookingMatch = bookings.value.find(
      (booking) => booking.id.toLowerCase() === raw.toLowerCase() || booking.tracking_code?.toLowerCase() === raw.toLowerCase(),
    )
    const code = bookingMatch?.tracking_code || raw
    matchedBooking.value = bookingMatch || null
    result.value = await api<TrackingResult>(`/public/tracking/${encodeURIComponent(code)}`)
    if (import.meta.client) {
      void navigateTo(`/dashboard/customer/track?code=${encodeURIComponent(code)}`, { replace: true })
    }
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'We could not find a shipment with that reference.')
  } finally {
    loading.value = false
  }
}

function trackBooking(booking: CustomerBooking) {
  if (!booking.tracking_code) return
  reference.value = booking.tracking_code
  void track()
}
</script>

<template>
  <div class="dash-page">
    <div>
      <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Shipments</p>
      <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Track shipment</h1>
      <p class="mt-1.5 max-w-2xl text-[12.5px] leading-relaxed text-ink-900/50">
        Use the tracking code or booking reference from your receipt to see every update.
      </p>
    </div>

    <form class="mt-5 rounded-2xl border border-ink-900/10 bg-surface p-4 sm:p-5" @submit.prevent="track">
      <label class="block">
        <span class="mb-1.5 block text-[10.5px] font-semibold tracking-[0.08em] text-ink-900/50 uppercase">Tracking code or booking reference</span>
        <div class="flex flex-col gap-2 sm:flex-row">
          <div class="relative flex-1">
            <Search :size="15" class="pointer-events-none absolute top-3.5 left-3 text-ink-900/35" />
            <input
              v-model="reference"
              type="search"
              placeholder="e.g. TRK-9F2A81 or your booking reference"
              class="w-full rounded-xl border border-ink-900/12 bg-canvas/50 py-3 pr-3 pl-9 text-[12px] outline-none focus:border-accent-500 focus:bg-surface"
            />
          </div>
          <button
            type="submit"
            :disabled="loading"
            class="inline-flex min-h-11 items-center justify-center gap-1.5 rounded-xl bg-accent-500 px-5 py-2.5 text-[12px] font-semibold text-white transition hover:bg-accent-400 disabled:opacity-50"
          >
            <Loader2 v-if="loading" :size="14" class="animate-spin" />
            <Search v-else :size="14" />
            Track
          </button>
        </div>
      </label>
      <p class="mt-2 text-[10.5px] text-ink-900/40">
        You can also open a shipment from your bookings and tap “Track shipment”.
      </p>
    </form>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ error }}
    </p>

    <section v-if="result" class="mt-4 overflow-hidden rounded-2xl border border-ink-900/10 bg-surface">
      <div class="flex flex-wrap items-start justify-between gap-3 border-b border-ink-900/8 px-4 py-4 sm:px-5">
        <div>
          <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">Tracking code</p>
          <p class="mt-1 font-mono text-[14px] font-semibold">{{ result.tracking_code }}</p>
          <p class="mt-1 text-[10.5px] text-ink-900/45">Last updated {{ formatDateTime(result.updated_at) }}</p>
        </div>
        <span class="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[10px] font-semibold" :class="statusTone(result.status)">
          <span class="size-1.5 rounded-full bg-current" />
          {{ result.request_status === 'pending' ? 'Awaiting provider' : statusLabel(result.status) }}
        </span>
      </div>

      <div class="p-4 sm:p-5">
        <div class="grid gap-3 sm:grid-cols-2">
          <div class="rounded-xl border border-ink-900/8 bg-canvas/50 p-3.5">
            <div class="flex items-center gap-1.5 text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">
              <MapPin :size="11" /> Pickup
            </div>
            <p class="mt-1.5 text-[11.5px] font-medium">{{ result.pickup_address }}</p>
          </div>
          <div class="rounded-xl border border-ink-900/8 bg-canvas/50 p-3.5">
            <div class="flex items-center gap-1.5 text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">
              <MapPin :size="11" /> Destination
            </div>
            <p class="mt-1.5 text-[11.5px] font-medium">{{ result.dropoff_address }}</p>
          </div>
        </div>

        <div class="mt-4 flex items-center justify-between gap-2">
          <h2 class="text-[12.5px] font-semibold">Shipment timeline</h2>
          <span class="text-[10px] text-ink-900/40">{{ result.timeline.length }} update{{ result.timeline.length === 1 ? '' : 's' }}</span>
        </div>
        <ol v-if="result.timeline.length" class="mt-3 space-y-4">
          <li v-for="(event, index) in result.timeline.slice().reverse()" :key="`${event.created_at}-${index}`" class="relative flex gap-3">
            <span class="relative mt-1 flex size-5 shrink-0 items-center justify-center rounded-full border" :class="index === 0 ? 'border-brand-300 bg-brand-50 text-brand-700' : 'border-ink-900/10 bg-canvas text-ink-900/35'">
              <span class="size-1.5 rounded-full bg-current" />
            </span>
            <div class="min-w-0 border-b border-ink-900/8 pb-3 last:border-0 last:pb-0">
              <div class="flex flex-wrap items-center gap-2">
                <p class="text-[11.5px] font-semibold">{{ statusLabel(event.status) }}</p>
                <span v-if="index === 0" class="rounded-full bg-brand-50 px-1.5 py-0.5 text-[8.5px] font-semibold tracking-wide text-brand-700 uppercase">Latest</span>
              </div>
              <p v-if="event.note" class="mt-1 text-[11px] leading-relaxed text-ink-900/55">{{ event.note }}</p>
              <p class="mt-1 text-[9.5px] text-ink-900/35">{{ formatDateTime(event.created_at) }}</p>
            </div>
          </li>
        </ol>
        <p v-else class="mt-3 rounded-lg bg-canvas px-3 py-5 text-center text-[11.5px] text-ink-900/45">
          No timeline updates yet. Check back once the provider starts the pickup.
        </p>

        <div v-if="matchedBooking" class="mt-4 flex flex-wrap items-center justify-between gap-2 rounded-xl border border-ink-900/8 bg-canvas/50 px-3.5 py-3">
          <div>
            <p class="text-[10.5px] font-semibold">{{ matchedBooking.provider_name }}</p>
            <p class="mt-0.5 text-[10px] text-ink-900/45">{{ matchedBooking.item_description }} · {{ matchedBooking.weight_kg }} kg</p>
          </div>
          <NuxtLink
            :to="`/dashboard/customer/shipments?booking=${matchedBooking.id}`"
            class="inline-flex items-center gap-1 text-[10.5px] font-semibold text-brand-700 hover:text-brand-600"
          >
            Shipment details <ArrowRight :size="12" />
          </NuxtLink>
        </div>
      </div>
    </section>

    <section v-else-if="searched && !loading && !error" class="mt-4 rounded-xl border border-ink-900/10 bg-surface p-10 text-center">
      <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
        <Package :size="20" />
      </span>
      <h2 class="mt-3 text-[13.5px] font-semibold">No shipment found</h2>
      <p class="mx-auto mt-1 max-w-md text-[11.5px] leading-relaxed text-ink-900/45">
        Check the reference and try again, or open the shipment from your bookings.
      </p>
    </section>

    <section v-if="!result && tracked.length" class="mt-4 rounded-xl border border-ink-900/10 bg-surface p-4">
      <div class="flex items-center justify-between gap-2">
        <h2 class="text-[12.5px] font-semibold">In-progress shipments</h2>
        <NuxtLink to="/dashboard/customer/shipments" class="text-[10.5px] font-semibold text-brand-700 hover:text-brand-600">View all</NuxtLink>
      </div>
      <div class="mt-2 divide-y divide-ink-900/8">
        <button
          v-for="booking in tracked"
          :key="booking.id"
          type="button"
          class="flex w-full items-center gap-3 py-2.5 text-left transition hover:text-brand-700"
          @click="trackBooking(booking)"
        >
          <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
            <Truck :size="14" />
          </span>
          <span class="min-w-0 flex-1">
            <span class="block truncate text-[11.5px] font-semibold">{{ booking.provider_name }}</span>
            <span class="mt-0.5 block truncate text-[10px] text-ink-900/45">{{ booking.pickup_address }} → {{ booking.dropoff_address }}</span>
          </span>
          <span class="shrink-0 font-mono text-[9.5px] text-ink-900/35">{{ booking.tracking_code }}</span>
        </button>
      </div>
    </section>
  </div>
</template>
