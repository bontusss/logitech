<script setup lang="ts">
import {
  ArrowRight,
  CircleAlert,
  CreditCard,
  FileText,
  Loader2,
  Package,
} from 'lucide-vue-next'
import type { CustomerBooking } from '~/types/customer'

useHead({ title: 'Payments & Receipts · BOL' })

const {
  bookings,
  overview,
  loading,
  error,
  load,
  formatMoney,
} = useCustomerDashboard()

type FilterKey = 'all' | 'paid' | 'unpaid'
const filter = ref<FilterKey>('all')
const search = ref('')
const receiptOpen = ref(false)
const receiptBooking = ref<CustomerBooking | null>(null)

onMounted(() => load())

const billable = computed(() =>
  bookings.value.filter((booking) => booking.status !== 'cancelled' && booking.request_status !== 'declined'),
)

const paidBookings = computed(() => billable.value.filter((booking) => booking.is_paid))
const unpaidBookings = computed(() => billable.value.filter((booking) => !booking.is_paid))

const billsCurrency = computed(() => bookings.value[0]?.currency || 'NGN')

const totals = computed(() => {
  const summary = overview.value
  const paid = summary
    ? summary.paid_minor
    : paidBookings.value.reduce((sum, booking) => sum + booking.delivery_price_minor, 0)
  const outstanding = summary
    ? summary.outstanding_minor
    : unpaidBookings.value.reduce((sum, booking) => sum + booking.delivery_price_minor, 0)
  const currency = summary?.currency || billsCurrency.value
  return {
    paid: formatAmount(paid, currency),
    outstanding: formatAmount(outstanding, currency),
    paidCount: summary?.paid_count ?? paidBookings.value.length,
    unpaidCount: summary?.unpaid_count ?? unpaidBookings.value.length,
  }
})

const filtered = computed(() => {
  const term = search.value.trim().toLowerCase()
  return billable.value
    .filter((booking) => {
      if (filter.value === 'paid') return booking.is_paid
      if (filter.value === 'unpaid') return !booking.is_paid
      return true
    })
    .filter((booking) => {
      if (!term) return true
      return [booking.provider_name, booking.tracking_code || '', booking.item_description, booking.id]
        .some((value) => value.toLowerCase().includes(term))
    })
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
})

function formatAmount(minor: number, currency = 'NGN') {
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

const paymentLabels: Record<string, string> = {
  cash: 'Cash',
  pos: 'Card / POS',
  bank_transfer: 'Bank transfer',
}

function paymentLabel(booking: CustomerBooking) {
  if (!booking.is_paid) return 'Payment pending'
  return paymentLabels[booking.payment_method || ''] || 'Paid'
}

function openReceipt(booking: CustomerBooking) {
  receiptBooking.value = booking
  receiptOpen.value = true
}
</script>

<template>
  <div class="dash-page">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Personal</p>
        <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Payments &amp; receipts</h1>
        <p class="mt-1.5 max-w-2xl text-[12.5px] leading-relaxed text-ink-900/50">
          Every booking fee, what has been settled, and a receipt you can print or download.
        </p>
      </div>
      <AppButton to="/dashboard/customer/shipments" variant="secondary" size="md">
        My shipments <ArrowRight :size="14" />
      </AppButton>
    </div>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ error }}
    </p>

    <div class="mt-5 grid grid-cols-2 gap-3 lg:grid-cols-3">
      <DashboardStatCard
        label="Total paid"
        :value="totals.paid"
        :detail="`${totals.paidCount} settled bookings`"
        tone="success"
      />
      <DashboardStatCard
        label="Outstanding"
        :value="totals.outstanding"
        :detail="`${totals.unpaidCount} awaiting payment`"
        tone="accent"
      />
      <DashboardStatCard
        class="col-span-2 lg:col-span-1"
        label="Bookings"
        :value="overview?.billable_count ?? billable.length"
        detail="Excludes cancelled requests"
        tone="brand"
      />
    </div>

    <div class="mt-5 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
      <div class="flex flex-wrap gap-1.5">
        <button
          v-for="option in [
            { key: 'all' as const, label: 'All bookings' },
            { key: 'paid' as const, label: 'Paid' },
            { key: 'unpaid' as const, label: 'Unpaid' },
          ]"
          :key="option.key"
          type="button"
          class="rounded-full border px-3 py-1.5 text-[11px] font-semibold transition"
          :class="filter === option.key
            ? 'border-accent-500 bg-accent-500 text-white'
            : 'border-ink-900/10 bg-surface text-ink-900/55 hover:border-ink-900/25 hover:text-ink-900'"
          @click="filter = option.key"
        >
          {{ option.label }}
        </button>
      </div>
      <label class="relative block w-full lg:w-72">
        <input
          v-model="search"
          type="search"
          placeholder="Reference, provider, package…"
          class="w-full rounded-xl border border-ink-900/12 bg-surface px-3 py-2 text-[11.5px] outline-none focus:border-accent-500"
        />
      </label>
    </div>

    <div v-if="loading && !bookings.length" class="mt-4 flex items-center justify-center gap-2 rounded-xl border border-ink-900/10 bg-surface p-12 text-[12px] text-ink-900/45">
      <Loader2 :size="15" class="animate-spin" /> Loading your payments…
    </div>

    <div v-else-if="!filtered.length" class="mt-4 rounded-xl border border-ink-900/10 bg-surface p-10 text-center">
      <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
        <CreditCard :size="20" />
      </span>
      <h2 class="mt-3 text-[13.5px] font-semibold">
        {{ search || filter !== 'all' ? 'No bookings match this view' : 'No payments yet' }}
      </h2>
      <p class="mx-auto mt-1 max-w-md text-[11.5px] leading-relaxed text-ink-900/45">
        {{ search || filter !== 'all' ? 'Try a different reference or filter.' : 'Once you book a provider, the fee and its receipt will appear here.' }}
      </p>
    </div>

    <ul v-else class="mt-4 space-y-2.5">
      <li
        v-for="booking in filtered"
        :key="booking.id"
        class="flex flex-col gap-3 rounded-xl border border-ink-900/10 bg-surface p-4 sm:flex-row sm:items-center sm:justify-between"
      >
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <p class="text-[12.5px] font-semibold">{{ booking.provider_name }}</p>
            <span
              class="rounded-full border px-2 py-0.5 text-[9.5px] font-semibold"
              :class="booking.is_paid
                ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                : 'border-amber-200 bg-amber-50 text-amber-700'"
            >
              {{ paymentLabel(booking) }}
            </span>
            <span v-if="booking.tracking_code" class="font-mono text-[9.5px] text-ink-900/35">{{ booking.tracking_code }}</span>
          </div>
          <p class="mt-1 truncate text-[11px] text-ink-900/55">{{ booking.item_description }} · {{ booking.pickup_address }} → {{ booking.dropoff_address }}</p>
          <p class="mt-1 text-[10.5px] text-ink-900/40">
            Booked {{ formatDate(booking.created_at) }}
            <template v-if="booking.is_paid && booking.paid_at">
              <span class="px-1 text-ink-900/20">·</span> Paid {{ formatDate(booking.paid_at) }}
            </template>
          </p>
        </div>
        <div class="flex shrink-0 items-center justify-between gap-3 sm:justify-end">
          <p class="text-[14px] font-semibold tabular-nums">{{ formatMoney(booking) }}</p>
          <button
            type="button"
            class="inline-flex items-center gap-1 rounded-lg border border-ink-900/12 bg-surface px-2.5 py-1.5 text-[10.5px] font-semibold transition hover:bg-canvas"
            @click="openReceipt(booking)"
          >
            <FileText :size="12" /> Receipt
          </button>
        </div>
      </li>
    </ul>

    <p class="mt-4 text-[10.5px] leading-relaxed text-ink-900/40">
      BOL does not process payments. The amount and payment method shown are those recorded by the provider for your booking.
    </p>

    <CustomerReceiptDialog :open="receiptOpen" :booking="receiptBooking" @close="receiptOpen = false" />
  </div>
</template>
