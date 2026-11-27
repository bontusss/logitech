<script setup lang="ts">
import {
  ArrowRight,
  CircleAlert,
  Loader2,
  MapPin,
  Package,
  PackageCheck,
  Search,
  Truck,
} from 'lucide-vue-next'
import type { CustomerBooking } from '~/types/customer'

const { user } = useAuth()
const {
  loading,
  error,
  load,
  overview,
  activeBookings,
  pendingBookings,
  cancelledBookings,
  deliveredBookings,
  recentBookings,
  stateFor,
  formatMoney,
} = useCustomerDashboard()

onMounted(() => load())

const firstName = computed(() => user.value?.full_name?.split(' ')[0] || 'there')

const stats = computed(() => [
  {
    label: 'Active shipments',
    value: overview.value?.active ?? activeBookings.value.length,
    detail: 'Still moving or awaiting pickup',
    tone: 'brand' as const,
  },
  {
    label: 'Pending bookings',
    value: overview.value?.pending ?? pendingBookings.value.length,
    detail: 'Waiting for provider confirmation',
    tone: 'accent' as const,
  },
  {
    label: 'Completed',
    value: overview.value?.delivered ?? deliveredBookings.value.length,
    detail: 'Delivered successfully',
    tone: 'success' as const,
  },
  {
    label: 'Cancelled',
    value: overview.value?.cancelled ?? cancelledBookings.value.length,
    detail: 'Cancelled or declined',
    tone: 'neutral' as const,
  },
])

const quickActions = [
  {
    label: 'Send a package',
    detail: 'Compare providers and book a pickup',
    to: '/#rate-search',
    icon: Search,
  },
  {
    label: 'Track a shipment',
    detail: 'Follow a parcel with its reference',
    to: '/dashboard/customer/track',
    icon: MapPin,
  },
  {
    label: 'Address book',
    detail: 'Save the places you send from most',
    to: '/dashboard/customer/addresses',
    icon: Package,
  },
]

function formatDateTime(value: string) {
  return new Date(value).toLocaleString('en-NG', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function bookingDetail(booking: CustomerBooking) {
  if (booking.request_status === 'pending') return 'Waiting for provider confirmation'
  if (booking.request_status === 'declined') return 'Provider declined this request'
  if (booking.status === 'delivered') return 'Delivered'
  if (booking.status === 'cancelled') return 'Cancelled'
  return stateFor(booking).label
}
</script>

<template>
  <div class="dash-page">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">
          Customer dashboard
        </p>
        <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">
          Welcome back, {{ firstName }}
        </h1>
        <p class="mt-2 max-w-xl text-[12.5px] leading-relaxed text-ink-900/50">
          Everything you send, everyone you ship with, and every update in one place.
        </p>
      </div>
      <AppButton to="/#rate-search" size="md">
        Send a package
        <ArrowRight :size="15" />
      </AppButton>
    </div>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ error }}
    </p>

    <div class="mt-5 grid grid-cols-2 gap-3 xl:grid-cols-4">
      <DashboardStatCard
        v-for="stat in stats"
        :key="stat.label"
        :label="stat.label"
        :value="stat.value"
        :detail="stat.detail"
        :tone="stat.tone"
      />
    </div>

    <div class="mt-5 grid gap-4 lg:grid-cols-[minmax(0,1.55fr)_minmax(260px,1fr)]">
      <section class="rounded-xl border border-ink-900/10 bg-surface">
        <div class="flex items-center justify-between gap-3 border-b border-ink-900/8 px-4 py-3.5">
          <div>
            <h2 class="text-[13.5px] font-semibold">Recent shipment activity</h2>
            <p class="mt-0.5 text-[10.5px] text-ink-900/40">Your latest bookings and updates</p>
          </div>
          <NuxtLink
            to="/dashboard/customer/shipments"
            class="inline-flex items-center gap-1 text-[10.5px] font-semibold text-brand-700 hover:text-brand-600"
          >
            View all <ArrowRight :size="12" />
          </NuxtLink>
        </div>

        <div v-if="loading && !recentBookings.length" class="flex items-center justify-center gap-2 p-10 text-[12px] text-ink-900/45">
          <Loader2 :size="15" class="animate-spin" /> Loading your activity…
        </div>

        <div v-else-if="!recentBookings.length" class="p-8 text-center">
          <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
            <PackageCheck :size="20" />
          </span>
          <h3 class="mt-3 text-[13px] font-semibold">No shipments yet</h3>
          <p class="mx-auto mt-1 max-w-sm text-[11.5px] leading-relaxed text-ink-900/45">
            Search for a provider, compare rates, and your first shipment will show up here.
          </p>
          <AppButton to="/#rate-search" size="sm" class="mt-4">Find a provider</AppButton>
        </div>

        <ul v-else class="divide-y divide-ink-900/8">
          <li v-for="booking in recentBookings" :key="booking.id">
            <NuxtLink
              :to="`/dashboard/customer/shipments?booking=${booking.id}`"
              class="flex items-start gap-3 px-4 py-3.5 transition hover:bg-canvas/70"
            >
              <span
                class="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-lg border"
                :class="stateFor(booking).tone"
              >
                <Truck :size="14" />
              </span>
              <span class="min-w-0 flex-1">
                <span class="flex flex-wrap items-center gap-x-2 gap-y-1">
                  <span class="truncate text-[12px] font-semibold">{{ booking.provider_name }}</span>
                  <span class="text-[10.5px] text-ink-900/35">·</span>
                  <span class="text-[10.5px] text-ink-900/45">{{ formatDateTime(booking.updated_at) }}</span>
                </span>
                <span class="mt-1 block truncate text-[11px] text-ink-900/55">
                  {{ booking.pickup_address }} <span class="px-1 text-brand-500">→</span> {{ booking.dropoff_address }}
                </span>
                <span class="mt-1 block text-[10px] text-ink-900/40">{{ bookingDetail(booking) }}</span>
              </span>
              <span class="shrink-0 text-right">
                <span class="block text-[12px] font-semibold tabular-nums">{{ formatMoney(booking) }}</span>
                <span v-if="booking.tracking_code" class="mt-1 block font-mono text-[9.5px] text-ink-900/35">
                  {{ booking.tracking_code }}
                </span>
              </span>
            </NuxtLink>
          </li>
        </ul>
      </section>

      <aside class="space-y-4">
        <section class="rounded-xl border border-ink-900/10 bg-surface p-4">
          <h2 class="text-[13px] font-semibold">Quick actions</h2>
          <div class="mt-3 space-y-1.5">
            <NuxtLink
              v-for="action in quickActions"
              :key="action.to"
              :to="action.to"
              class="group flex items-start gap-3 rounded-xl border border-transparent px-2.5 py-2.5 transition hover:border-ink-900/8 hover:bg-canvas/70"
            >
              <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
                <component :is="action.icon" :size="14" />
              </span>
              <span class="min-w-0 flex-1">
                <span class="block text-[11.5px] font-semibold">{{ action.label }}</span>
                <span class="mt-0.5 block text-[10px] leading-relaxed text-ink-900/40">{{ action.detail }}</span>
              </span>
              <ArrowRight :size="13" class="mt-2 shrink-0 text-ink-900/25 transition group-hover:translate-x-0.5 group-hover:text-brand-600" />
            </NuxtLink>
          </div>
        </section>

        <section class="rounded-xl border border-accent-200/70 bg-accent-50/60 p-4 dark:border-accent-400/20 dark:bg-accent-400/5">
          <p class="text-[9.5px] font-semibold tracking-[0.14em] text-accent-700 uppercase dark:text-accent-300">Need a hand?</p>
          <h2 class="mt-1.5 text-[13px] font-semibold">We can help with a parcel</h2>
          <p class="mt-1 text-[11px] leading-relaxed text-ink-900/55">
            Report a delayed, missing or damaged shipment, or talk to the provider directly from the shipment details.
          </p>
          <NuxtLink
            to="/dashboard/customer/support"
            class="mt-3 inline-flex items-center gap-1 text-[11px] font-semibold text-accent-700 hover:text-accent-600 dark:text-accent-300"
          >
            Get support <ArrowRight :size="12" />
          </NuxtLink>
        </section>
      </aside>
    </div>
  </div>
</template>
