<script setup lang="ts">
import {
  ArrowRight,
  CircleAlert,
  Clock3,
  FileText,
  Loader2,
  MapPin,
  MessageSquareWarning,
  Package,
  Phone,
  Search,
  Truck,
  X,
} from 'lucide-vue-next'
import type { CustomerBooking, CustomerIssueType, TrackingResult } from '~/types/customer'

useHead({ title: 'My Shipments · BOL' })

const route = useRoute()
const { api } = useApi()
const { notify, confirm } = useWorkspace()
const {
  bookings,
  loading,
  error,
  load,
  stateFor,
  formatMoney,
} = useCustomerDashboard()

type FilterKey = 'all' | 'active' | 'pending' | 'delivered' | 'cancelled'
const filter = ref<FilterKey>('all')
const search = ref('')
const selected = ref<CustomerBooking | null>(null)
const timeline = ref<TrackingResult['timeline']>([])
const timelineLoading = ref(false)
const cancelBusy = ref(false)
const actionError = ref('')
const reportOpen = ref(false)
const reportBusy = ref(false)
const receiptOpen = ref(false)
const receiptBooking = ref<CustomerBooking | null>(null)
const report = reactive<{ issue_type: CustomerIssueType; description: string }>({
  issue_type: 'delayed',
  description: '',
})

onMounted(async () => {
  await load()
  const bookingID = typeof route.query.booking === 'string' ? route.query.booking : ''
  if (bookingID) {
    const match = bookings.value.find((booking) => booking.id === bookingID)
    if (match) void openBooking(match)
  }
})

const filters = computed(() => [
  { key: 'all' as const, label: 'All', count: bookings.value.length },
  {
    key: 'active' as const,
    label: 'Active',
    count: bookings.value.filter((booking) =>
      booking.request_status === 'accepted' &&
      !['delivered', 'cancelled'].includes(booking.status),
    ).length,
  },
  {
    key: 'pending' as const,
    label: 'Pending',
    count: bookings.value.filter((booking) => booking.request_status === 'pending' || booking.status === 'requested').length,
  },
  {
    key: 'delivered' as const,
    label: 'Delivered',
    count: bookings.value.filter((booking) => booking.status === 'delivered').length,
  },
  {
    key: 'cancelled' as const,
    label: 'Cancelled',
    count: bookings.value.filter((booking) => booking.status === 'cancelled' || booking.request_status === 'declined').length,
  },
])

const filtered = computed(() => {
  const term = search.value.trim().toLowerCase()
  return bookings.value.filter((booking) => {
    const matchesFilter = (() => {
      if (filter.value === 'active') {
        return booking.request_status === 'accepted' && !['delivered', 'cancelled'].includes(booking.status)
      }
      if (filter.value === 'pending') {
        return booking.request_status === 'pending' || booking.status === 'requested'
      }
      if (filter.value === 'delivered') return booking.status === 'delivered'
      if (filter.value === 'cancelled') {
        return booking.status === 'cancelled' || booking.request_status === 'declined'
      }
      return true
    })()
    if (!matchesFilter) return false
    if (!term) return true
    return [
      booking.provider_name,
      booking.tracking_code || '',
      booking.pickup_address,
      booking.dropoff_address,
      booking.item_description,
      booking.id,
    ].some((value) => value.toLowerCase().includes(term))
  })
})

function formatDate(value: string) {
  return new Date(value).toLocaleDateString('en-NG', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleString('en-NG', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function speedLabel(speed: CustomerBooking['delivery_speed']) {
  return { standard: 'Standard', express: 'Express', same_day: 'Same-day' }[speed] || speed
}

function paymentLabel(booking: CustomerBooking) {
  if (!booking.is_paid) return 'Payment pending'
  return { cash: 'Cash', pos: 'Card / POS', bank_transfer: 'Bank transfer' }[booking.payment_method || ''] || 'Paid'
}

async function openBooking(booking: CustomerBooking) {
  selected.value = booking
  actionError.value = ''
  reportOpen.value = false
  timeline.value = []
  if (!booking.tracking_code) return
  timelineLoading.value = true
  try {
    const result = await api<TrackingResult>(`/public/tracking/${booking.tracking_code}`)
    timeline.value = result.timeline ?? []
  } catch {
    timeline.value = []
  } finally {
    timelineLoading.value = false
  }
}

function closeBooking() {
  selected.value = null
  reportOpen.value = false
  actionError.value = ''
}

async function cancelBooking() {
  if (!selected.value) return
  const confirmed = await confirm({
    title: 'Cancel this shipment?',
    message: `Cancel the booking with ${selected.value.provider_name}? This only works before the parcel is picked up.`,
    confirmLabel: 'Cancel shipment',
    cancelLabel: 'Keep shipment',
    danger: true,
  })
  if (!confirmed || !selected.value) return
  cancelBusy.value = true
  actionError.value = ''
  try {
    await api(`/customer/bookings/${selected.value.id}/cancel`, { method: 'POST' })
    const match = bookings.value.find((booking) => booking.id === selected.value?.id)
    if (match) {
      match.status = 'cancelled'
      match.can_cancel = false
      match.updated_at = new Date().toISOString()
      selected.value = match
    }
    notify('Shipment cancelled.')
  } catch (cause) {
    actionError.value = apiErrorMessage(cause, 'Could not cancel this shipment.')
  } finally {
    cancelBusy.value = false
  }
}

async function submitReport() {
  if (!selected.value) return
  if (report.description.trim().length < 5) {
    actionError.value = 'Add a short description so the provider can investigate.'
    return
  }
  reportBusy.value = true
  actionError.value = ''
  try {
    await api(`/customer/bookings/${selected.value.id}/issues`, {
      method: 'POST',
      body: { issue_type: report.issue_type, description: report.description.trim() },
    })
    reportOpen.value = false
    report.description = ''
    notify('Issue reported. The provider can now pick it up.')
  } catch (cause) {
    actionError.value = apiErrorMessage(cause, 'Could not report this issue.')
  } finally {
    reportBusy.value = false
  }
}

function openReceipt(booking: CustomerBooking) {
  receiptBooking.value = booking
  receiptOpen.value = true
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Shipments</p>
        <h1 class="mt-2 text-[clamp(1.5rem,2.6vw,1.9rem)] font-semibold tracking-[-0.035em]">My shipments</h1>
        <p class="mt-1.5 max-w-2xl text-[12.5px] leading-relaxed text-ink-900/50">
          Every booking across every provider, with the details you need to follow up.
        </p>
      </div>
      <AppButton to="/#rate-search" size="md">
        Send a package <ArrowRight :size="14" />
      </AppButton>
    </div>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ error }}
    </p>

    <div class="mt-5 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
      <div class="flex flex-wrap gap-1.5">
        <button
          v-for="option in filters"
          :key="option.key"
          type="button"
          class="inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-[11px] font-semibold transition"
          :class="filter === option.key
            ? 'border-accent-500 bg-accent-500 text-white'
            : 'border-ink-900/10 bg-surface text-ink-900/55 hover:border-ink-900/25 hover:text-ink-900'"
          @click="filter = option.key"
        >
          {{ option.label }}
          <span
            class="rounded-full px-1.5 py-0.5 text-[9px] tabular-nums"
            :class="filter === option.key ? 'bg-white/15 text-white dark:bg-ink-900/10 dark:text-ink-900' : 'bg-ink-900/5 text-ink-900/45'"
          >
            {{ option.count }}
          </span>
        </button>
      </div>

      <label class="relative block w-full lg:w-72">
        <Search :size="14" class="pointer-events-none absolute top-2.5 left-3 text-ink-900/35" />
        <input
          v-model="search"
          type="search"
          placeholder="Reference, provider, route…"
          class="w-full rounded-xl border border-ink-900/12 bg-surface py-2 pr-3 pl-9 text-[11.5px] outline-none focus:border-accent-500"
        />
      </label>
    </div>

    <div v-if="loading && !bookings.length" class="mt-4 flex items-center justify-center gap-2 rounded-xl border border-ink-900/10 bg-surface p-12 text-[12px] text-ink-900/45">
      <Loader2 :size="15" class="animate-spin" /> Loading your shipments…
    </div>

    <div v-else-if="!filtered.length" class="mt-4 rounded-xl border border-ink-900/10 bg-surface p-10 text-center">
      <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
        <Package :size="20" />
      </span>
      <h2 class="mt-3 text-[13.5px] font-semibold">
        {{ search ? 'No shipments match that search' : 'No shipments in this view' }}
      </h2>
      <p class="mx-auto mt-1 max-w-md text-[11.5px] leading-relaxed text-ink-900/45">
        {{ search ? 'Try a different reference, provider or route.' : 'Book a provider and your shipment will appear here.' }}
      </p>
      <AppButton v-if="!search" to="/#rate-search" size="sm" class="mt-4">Find a provider</AppButton>
    </div>

    <ul v-else class="mt-4 space-y-2.5">
      <li
        v-for="booking in filtered"
        :key="booking.id"
        class="rounded-xl border border-ink-900/10 bg-surface transition hover:border-ink-900/20"
      >
        <div class="flex flex-col gap-3 p-4 sm:flex-row sm:items-start sm:justify-between">
          <div class="flex min-w-0 items-start gap-3">
            <span
              class="mt-0.5 flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-ink-900/8 bg-gradient-to-br from-slate-50 to-white text-[12px] font-bold text-brand-700 dark:border-white/10 dark:from-white/10 dark:to-transparent"
            >
              <img
                v-if="booking.provider_logo_url"
                :src="booking.provider_logo_url"
                :alt="`${booking.provider_name} logo`"
                class="size-full object-cover"
              />
              <span v-else>{{ booking.provider_name.charAt(0).toUpperCase() }}</span>
            </span>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <p class="text-[12.5px] font-semibold">{{ booking.provider_name }}</p>
                <span class="inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[9.5px] font-semibold" :class="stateFor(booking).tone">
                  <span class="size-1.5 rounded-full" :class="stateFor(booking).dot" />
                  {{ stateFor(booking).label }}
                </span>
                <span v-if="booking.tracking_code" class="font-mono text-[9.5px] text-ink-900/35">{{ booking.tracking_code }}</span>
              </div>
              <p class="mt-1.5 truncate text-[11.5px] text-ink-900/60">
                {{ booking.pickup_address }} <span class="px-1 text-brand-500">→</span> {{ booking.dropoff_address }}
              </p>
              <div class="mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-[10.5px] text-ink-900/42">
                <span class="inline-flex items-center gap-1"><Clock3 :size="11" /> {{ formatDateTime(booking.created_at) }}</span>
                <span>{{ speedLabel(booking.delivery_speed) }}</span>
                <span>{{ booking.item_description }} · {{ booking.weight_kg }} kg</span>
              </div>
            </div>
          </div>
          <div class="flex shrink-0 items-center justify-between gap-2 sm:flex-col sm:items-end">
            <p class="text-[14px] font-semibold tabular-nums">{{ formatMoney(booking) }}</p>
            <div class="flex items-center gap-1.5">
              <button
                type="button"
                class="inline-flex items-center gap-1 rounded-lg border border-ink-900/12 bg-surface px-2.5 py-1.5 text-[10.5px] font-semibold transition hover:bg-canvas"
                @click="openReceipt(booking)"
              >
                <FileText :size="12" /> Receipt
              </button>
              <button
                type="button"
                class="inline-flex items-center gap-1 rounded-lg bg-accent-500 px-2.5 py-1.5 text-[10.5px] font-semibold text-white transition hover:bg-accent-400"
                @click="openBooking(booking)"
              >
                Details <ArrowRight :size="12" />
              </button>
            </div>
          </div>
        </div>
      </li>
    </ul>

    <!-- Shipment details -->
    <Teleport to="body">
      <div
        v-if="selected"
        class="fixed inset-0 z-[70] flex items-start justify-center overflow-y-auto bg-ink-950/45 px-3 py-6 backdrop-blur-sm sm:px-5"
        role="dialog"
        aria-modal="true"
        aria-label="Shipment details"
        @click.self="closeBooking"
      >
        <div class="my-auto w-full max-w-2xl overflow-hidden rounded-2xl border border-ink-900/10 bg-surface shadow-2xl">
          <div class="flex items-start justify-between gap-3 border-b border-ink-900/8 px-5 py-4">
            <div>
              <div class="flex flex-wrap items-center gap-2">
                <p class="text-[14px] font-semibold">{{ selected.provider_name }}</p>
                <span class="inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[9.5px] font-semibold" :class="stateFor(selected).tone">
                  <span class="size-1.5 rounded-full" :class="stateFor(selected).dot" />
                  {{ stateFor(selected).label }}
                </span>
              </div>
              <p class="mt-1 font-mono text-[10.5px] text-ink-900/45">{{ selected.tracking_code || selected.id }}</p>
            </div>
            <button type="button" class="text-ink-900/35 hover:text-ink-900" aria-label="Close details" @click="closeBooking">
              <X :size="18" />
            </button>
          </div>

          <div class="max-h-[calc(100dvh-10rem)] overflow-y-auto p-5">
            <p v-if="actionError" role="alert" class="mb-3 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[11px] text-red-700">
              <CircleAlert :size="13" class="mt-0.5 shrink-0" />{{ actionError }}
            </p>

            <div class="grid gap-3 sm:grid-cols-2">
              <section class="rounded-xl border border-ink-900/8 bg-canvas/50 p-3.5 sm:col-span-2">
                <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">Route</p>
                <div class="mt-2 space-y-2">
                  <div class="flex items-start gap-2">
                    <span class="mt-1 size-2 shrink-0 rounded-full bg-emerald-500" />
                    <div>
                      <p class="text-[10px] font-semibold text-ink-900/40 uppercase">Pickup</p>
                      <p class="text-[11.5px] font-medium">{{ selected.pickup_address }}</p>
                    </div>
                  </div>
                  <div class="ml-[3px] h-3 border-l border-dashed border-ink-900/20" />
                  <div class="flex items-start gap-2">
                    <span class="mt-1 size-2 shrink-0 rounded-full bg-brand-500" />
                    <div>
                      <p class="text-[10px] font-semibold text-ink-900/40 uppercase">Destination</p>
                      <p class="text-[11.5px] font-medium">{{ selected.dropoff_address }}</p>
                    </div>
                  </div>
                </div>
              </section>

              <section class="rounded-xl border border-ink-900/8 p-3.5">
                <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">Sender</p>
                <p class="mt-1.5 text-[11.5px] font-semibold">{{ selected.sender_name }}</p>
                <a :href="`tel:${selected.sender_phone}`" class="mt-0.5 inline-flex items-center gap-1 text-[10.5px] text-ink-900/50 hover:text-brand-700">
                  <Phone :size="11" /> {{ selected.sender_phone }}
                </a>
              </section>
              <section class="rounded-xl border border-ink-900/8 p-3.5">
                <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">Receiver</p>
                <p class="mt-1.5 text-[11.5px] font-semibold">{{ selected.receiver_name }}</p>
                <a :href="`tel:${selected.receiver_phone}`" class="mt-0.5 inline-flex items-center gap-1 text-[10.5px] text-ink-900/50 hover:text-brand-700">
                  <Phone :size="11" /> {{ selected.receiver_phone }}
                </a>
              </section>

              <section class="rounded-xl border border-ink-900/8 p-3.5">
                <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">Provider</p>
                <p class="mt-1.5 text-[11.5px] font-semibold">{{ selected.provider_name }}</p>
                <a v-if="selected.provider_phone" :href="`tel:${selected.provider_phone}`" class="mt-0.5 inline-flex items-center gap-1 text-[10.5px] text-ink-900/50 hover:text-brand-700">
                  <Phone :size="11" /> {{ selected.provider_phone }}
                </a>
                <NuxtLink
                  v-if="selected.provider_slug"
                  :to="`/book/${selected.provider_slug}`"
                  target="_blank"
                  class="mt-1.5 inline-flex items-center gap-1 text-[10px] font-semibold text-brand-700 hover:text-brand-600"
                >
                  Open booking page <ArrowRight :size="11" />
                </NuxtLink>
              </section>
              <section class="rounded-xl border border-ink-900/8 p-3.5">
                <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">Payment</p>
                <p class="mt-1.5 text-[13px] font-semibold tabular-nums">{{ formatMoney(selected) }}</p>
                <p class="mt-0.5 text-[10.5px]" :class="selected.is_paid ? 'text-emerald-700' : 'text-amber-700'">{{ paymentLabel(selected) }}</p>
              </section>

              <section class="rounded-xl border border-ink-900/8 p-3.5 sm:col-span-2">
                <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">Package</p>
                <p class="mt-1.5 text-[11.5px] font-medium">{{ selected.item_description }}</p>
                <p class="mt-0.5 text-[10.5px] text-ink-900/45">
                  {{ selected.package_category }} · {{ selected.weight_kg }} kg · {{ speedLabel(selected.delivery_speed) }} · {{ selected.distance_km.toFixed(1) }} km
                </p>
                <p v-if="selected.additional_instructions" class="mt-2 rounded-lg bg-canvas px-2.5 py-2 text-[10.5px] leading-relaxed text-ink-900/55">
                  {{ selected.additional_instructions }}
                </p>
                <a v-if="selected.package_photo_url" :href="selected.package_photo_url" target="_blank" rel="noopener" class="mt-2 inline-flex text-[10.5px] font-semibold text-brand-700">
                  View package photo
                </a>
              </section>
            </div>

            <section class="mt-3 rounded-xl border border-ink-900/8 p-3.5">
              <div class="flex items-center justify-between gap-2">
                <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">Tracking history</p>
                <NuxtLink
                  v-if="selected.tracking_code"
                  :to="`/dashboard/customer/track?code=${selected.tracking_code}`"
                  class="text-[10px] font-semibold text-brand-700 hover:text-brand-600"
                >
                  Full tracking
                </NuxtLink>
              </div>
              <div v-if="timelineLoading" class="mt-3 flex items-center gap-2 text-[11px] text-ink-900/45">
                <Loader2 :size="13" class="animate-spin" /> Loading updates…
              </div>
              <ol v-else-if="timeline.length" class="mt-3 space-y-3">
                <li v-for="(event, index) in timeline.slice().reverse()" :key="`${event.created_at}-${index}`" class="flex gap-2.5">
                  <span class="mt-1.5 size-1.5 shrink-0 rounded-full" :class="index === 0 ? 'bg-brand-500' : 'bg-ink-900/20'" />
                  <div class="min-w-0">
                    <p class="text-[11px] font-semibold capitalize">{{ event.status.replaceAll('_', ' ') }}</p>
                    <p v-if="event.note" class="text-[10.5px] leading-relaxed text-ink-900/50">{{ event.note }}</p>
                    <p class="text-[9.5px] text-ink-900/35">{{ formatDateTime(event.created_at) }}</p>
                  </div>
                </li>
              </ol>
              <p v-else class="mt-2 text-[11px] text-ink-900/45">Updates will appear once the provider starts moving the parcel.</p>
            </section>

            <section v-if="reportOpen" class="mt-3 rounded-xl border border-amber-200 bg-amber-50/60 p-3.5">
              <div class="flex items-start justify-between gap-2">
                <div>
                  <p class="text-[11.5px] font-semibold">Report a problem</p>
                  <p class="mt-0.5 text-[10px] text-ink-900/50">The provider will see this on the shipment.</p>
                </div>
                <button type="button" class="text-ink-900/35 hover:text-ink-900" aria-label="Close report form" @click="reportOpen = false"><X :size="14" /></button>
              </div>
              <div class="mt-2.5 grid gap-2 sm:grid-cols-[160px_minmax(0,1fr)]">
                <select v-model="report.issue_type" class="rounded-lg border border-ink-900/12 bg-surface px-2.5 py-2 text-[11px]">
                  <option value="delayed">Delayed</option>
                  <option value="missing">Missing</option>
                  <option value="damaged">Damaged</option>
                  <option value="other">Other</option>
                </select>
                <textarea
                  v-model="report.description"
                  rows="3"
                  maxlength="2000"
                  placeholder="What happened?"
                  class="rounded-lg border border-ink-900/12 bg-surface px-2.5 py-2 text-[11px] outline-none focus:border-accent-500"
                />
              </div>
              <button
                type="button"
                :disabled="reportBusy"
                class="mt-2 inline-flex items-center gap-1.5 rounded-lg bg-accent-500 px-3 py-2 text-[11px] font-semibold text-white transition hover:bg-accent-400 disabled:opacity-50"
                @click="submitReport"
              >
                <Loader2 v-if="reportBusy" :size="12" class="animate-spin" />
                Submit report
              </button>
            </section>

            <div class="mt-4 flex flex-wrap justify-end gap-2">
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-lg border border-ink-900/15 bg-surface px-3 py-2 text-[11.5px] font-semibold transition hover:bg-canvas"
                @click="openReceipt(selected)"
              >
                <FileText :size="13" /> Receipt
              </button>
              <button
                v-if="!reportOpen"
                type="button"
                class="inline-flex items-center gap-1.5 rounded-lg border border-amber-300 bg-amber-50 px-3 py-2 text-[11.5px] font-semibold text-amber-800 transition hover:bg-amber-100"
                @click="reportOpen = true"
              >
                <MessageSquareWarning :size="13" /> Report issue
              </button>
              <button
                v-if="selected.can_cancel"
                type="button"
                :disabled="cancelBusy"
                class="inline-flex items-center gap-1.5 rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-[11.5px] font-semibold text-rose-700 transition hover:bg-rose-100 disabled:opacity-50"
                @click="cancelBooking"
              >
                <Loader2 v-if="cancelBusy" :size="13" class="animate-spin" />
                Cancel shipment
              </button>
              <button
                v-if="selected.tracking_code"
                type="button"
                class="inline-flex items-center gap-1.5 rounded-lg bg-accent-500 px-3 py-2 text-[11.5px] font-semibold text-white transition hover:bg-accent-400"
                @click="navigateTo(`/dashboard/customer/track?code=${selected.tracking_code}`)"
              >
                <MapPin :size="13" /> Track shipment
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <CustomerReceiptDialog :open="receiptOpen" :booking="receiptBooking" @close="receiptOpen = false" />
  </div>
</template>
