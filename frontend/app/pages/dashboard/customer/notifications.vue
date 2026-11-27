<script setup lang="ts">
import {
  CheckCircle2,
  CircleAlert,
  Clock3,
  Loader2,
  MessageSquareWarning,
  Package,
  RefreshCw,
  Truck,
  XCircle,
} from 'lucide-vue-next'
import type { Component } from 'vue'
import type { CustomerBooking, CustomerIssue } from '~/types/customer'

useHead({ title: 'Notifications · BOL' })

interface ActivityItem {
  id: string
  title: string
  detail: string
  at: string
  href: string
  icon: Component
  tone: string
}

const { api } = useApi()
const {
  bookings,
  loaded,
  error: bookingError,
  load: loadBookings,
} = useCustomerDashboard()

const issues = ref<CustomerIssue[]>([])
const loadingIssues = ref(true)
const refreshing = ref(false)
const error = ref('')

const items = computed<ActivityItem[]>(() => {
  const result: ActivityItem[] = []

  for (const booking of bookings.value) {
    const route = `${booking.pickup_address} → ${booking.dropoff_address}`
    const href = `/dashboard/customer/shipments?booking=${booking.id}`
    const base = { href, detail: `${booking.provider_name} · ${route}` }

    result.push({
      ...base,
      id: `booked-${booking.id}`,
      title: 'Booking created',
      at: booking.created_at,
      icon: Package,
      tone: 'bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200',
    })

    if (booking.request_status === 'pending') {
      result.push({
        ...base,
        id: `pending-${booking.id}`,
        title: 'Waiting for provider confirmation',
        at: booking.updated_at,
        icon: Clock3,
        tone: 'bg-amber-50 text-amber-700 dark:bg-amber-400/10 dark:text-amber-200',
      })
    }

    if (booking.request_status === 'declined') {
      result.push({
        ...base,
        id: `declined-${booking.id}`,
        title: 'Provider declined the request',
        at: booking.updated_at,
        icon: XCircle,
        tone: 'bg-rose-50 text-rose-700 dark:bg-rose-400/10 dark:text-rose-200',
      })
      continue
    }

    if (booking.status === 'cancelled') {
      result.push({
        ...base,
        id: `cancelled-${booking.id}`,
        title: 'Shipment cancelled',
        at: booking.updated_at,
        icon: XCircle,
        tone: 'bg-rose-50 text-rose-700 dark:bg-rose-400/10 dark:text-rose-200',
      })
      continue
    }

    const statusCopy: Partial<Record<CustomerBooking['status'], { title: string; icon: Component; tone: string }>> = {
      assigned: {
        title: 'Rider assigned',
        icon: Truck,
        tone: 'bg-sky-50 text-sky-700 dark:bg-sky-400/10 dark:text-sky-200',
      },
      picked_up: {
        title: 'Parcel picked up',
        icon: Truck,
        tone: 'bg-sky-50 text-sky-700 dark:bg-sky-400/10 dark:text-sky-200',
      },
      in_transit: {
        title: 'Parcel is in transit',
        icon: Truck,
        tone: 'bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200',
      },
      delivered: {
        title: 'Parcel delivered',
        icon: CheckCircle2,
        tone: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-400/10 dark:text-emerald-200',
      },
    }

    const status = statusCopy[booking.status]
    if (status) {
      result.push({
        ...base,
        id: `status-${booking.id}`,
        title: status.title,
        at: booking.updated_at,
        icon: status.icon,
        tone: status.tone,
      })
    }
  }

  for (const issue of issues.value) {
    const href = issue.delivery_id
      ? `/dashboard/customer/shipments?booking=${issue.delivery_id}`
      : '/dashboard/customer/support'
    const detail = `${issue.provider_name} · ${issue.pickup_address} → ${issue.dropoff_address}`
    result.push({
      id: `issue-${issue.id}`,
      title: issue.status === 'resolved' ? 'Support report resolved' : `Support report: ${issue.issue_type}`,
      detail,
      at: issue.created_at,
      href,
      icon: MessageSquareWarning,
      tone: issue.status === 'resolved'
        ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-400/10 dark:text-emerald-200'
        : 'bg-amber-50 text-amber-700 dark:bg-amber-400/10 dark:text-amber-200',
    })
  }

  return result
    .filter((item) => item.at)
    .sort((a, b) => new Date(b.at).getTime() - new Date(a.at).getTime())
    .slice(0, 40)
})

const grouped = computed(() => {
  const groups: { label: string; items: ActivityItem[] }[] = []
  for (const item of items.value) {
    const label = dayLabel(item.at)
    const last = groups[groups.length - 1]
    if (last && last.label === label) last.items.push(item)
    else groups.push({ label, items: [item] })
  }
  return groups
})

const displayError = computed(() => error.value || bookingError.value)
const loading = computed(
  () => (loadingIssues.value || !loaded.value) && !refreshing.value && !displayError.value,
)

function dayLabel(value: string) {
  const date = new Date(value)
  const today = new Date()
  const startOfToday = new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime()
  const startOfDate = new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime()
  const diffDays = Math.round((startOfToday - startOfDate) / 86_400_000)
  if (diffDays === 0) return 'Today'
  if (diffDays === 1) return 'Yesterday'
  if (diffDays < 7) return 'Earlier this week'
  return date.toLocaleDateString('en-NG', { month: 'long', year: 'numeric' })
}

function formatTime(value: string) {
  return new Date(value).toLocaleString('en-NG', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function loadIssues() {
  loadingIssues.value = true
  try {
    const result = await api<{ issues: CustomerIssue[] }>('/customer/issues')
    issues.value = result.issues
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not load your support updates.')
  } finally {
    loadingIssues.value = false
  }
}

async function refresh() {
  refreshing.value = true
  error.value = ''
  try {
    await Promise.all([loadBookings(true), loadIssues()])
  } finally {
    refreshing.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadBookings(), loadIssues()])
})
</script>

<template>
  <div class="dash-page">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Support</p>
        <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Notifications</h1>
        <p class="mt-1.5 max-w-2xl text-[12.5px] leading-relaxed text-ink-900/50">
          Shipment and support activity from across your account, newest first.
        </p>
      </div>
      <button
        type="button"
        :disabled="refreshing"
        class="inline-flex items-center gap-1.5 rounded-lg border border-ink-900/15 bg-surface px-3.5 py-2 text-[11.5px] font-semibold transition hover:bg-canvas disabled:opacity-50"
        @click="refresh"
      >
        <RefreshCw :size="13" :class="refreshing ? 'animate-spin' : ''" /> Refresh
      </button>
    </div>

    <p v-if="displayError" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ displayError }}
    </p>

    <div v-if="loading" class="mt-4 flex items-center justify-center gap-2 rounded-xl border border-ink-900/10 bg-surface p-12 text-[12px] text-ink-900/45">
      <Loader2 :size="15" class="animate-spin" /> Loading your updates…
    </div>

    <div v-else-if="!items.length" class="mt-4 rounded-xl border border-ink-900/10 bg-surface p-10 text-center">
      <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
        <Clock3 :size="20" />
      </span>
      <h2 class="mt-3 text-[13.5px] font-semibold">Nothing to catch up on</h2>
      <p class="mx-auto mt-1 max-w-md text-[11.5px] leading-relaxed text-ink-900/45">
        Booking, tracking and support updates will appear here as they happen.
      </p>
    </div>

    <div v-else class="mt-4 space-y-5">
      <section v-for="group in grouped" :key="group.label">
        <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">{{ group.label }}</p>
        <ul class="mt-2 space-y-2">
          <li v-for="item in group.items" :key="item.id">
            <NuxtLink
              :to="item.href"
              class="flex items-start gap-3 rounded-xl border border-ink-900/10 bg-surface px-3.5 py-3 transition hover:border-ink-900/20 hover:bg-canvas/40"
            >
              <span class="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-lg" :class="item.tone">
                <component :is="item.icon" :size="14" />
              </span>
              <span class="min-w-0 flex-1">
                <span class="block text-[11.5px] font-semibold">{{ item.title }}</span>
                <span class="mt-0.5 block truncate text-[10.5px] text-ink-900/50">{{ item.detail }}</span>
              </span>
              <span class="shrink-0 text-[9.5px] text-ink-900/35">{{ formatTime(item.at) }}</span>
            </NuxtLink>
          </li>
        </ul>
      </section>
    </div>

    <p class="mt-4 text-[10.5px] leading-relaxed text-ink-900/40">
      This feed reflects shipment and support records in your account. Email and WhatsApp alerts are sent according to your provider's settings.
    </p>
  </div>
</template>
