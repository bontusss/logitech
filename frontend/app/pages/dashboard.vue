<script setup lang="ts">
import {
  ArrowRight,
  Building2,
  Check,
  CheckCircle2,
  CircleAlert,
  CircleDot,
  ClipboardList,
  Copy,
  Download,
  ExternalLink,
  Loader2,
  MailCheck,
  Route,
  Share2,
  ShieldCheck,
  Sparkles,
  Users,
} from 'lucide-vue-next'
import * as QRCode from 'qrcode'
import type { CompanyDetail } from '~/types/auth'

definePageMeta({ middleware: 'auth', layout: 'app' })

const { api } = useApi()
const { user } = useAuth()
const { currentCompany } = useCompany()
const route = useRoute()

useSeoMeta({
  title: () => (user.value?.account_type === 'customer' ? 'Marketplace dashboard' : 'Workspace overview'),
})

const isCustomer = computed(() => user.value?.account_type === 'customer')

const detail = ref<CompanyDetail | null>(null)
interface BookingPage {
  slug: string
  is_published: boolean
}

interface Driver {
  id: string
  full_name: string
  status: string
}

interface OverviewBooking {
  id: string
  delivery_price_minor: number
  is_paid: boolean
  paid_at: string | null
  payment_method: string | null
  status: string
  request_status: string
  created_at: string
  updated_at: string
}

interface CompanyOverview {
  financial: {
    earned_today_minor: number
    collected_today_minor: number
    amount_owed_minor: number
    cash_with_drivers_minor: number
  }
  bookings: {
    created_today: number
    in_progress: number
    waiting_for_drivers: number
    requests: number
    total: number
  }
  daily: { date: string; bookings: number }[]
}

const bookingPage = ref<BookingPage | null>(null)
const drivers = ref<Driver[]>([])
const overview = ref<CompanyOverview | null>(null)
const requests = ref<OverviewBooking[]>([])
const metricsError = ref('')
const loading = ref(false)
const error = ref('')
const bookingPageError = ref('')
const driverError = ref('')
const marketplaceZones = ref<number | null>(null)
const marketplaceActiveRoutes = ref<number | null>(null)
const marketplaceSetupError = ref('')
const copiedBookingLink = ref(false)
const bookingShareError = ref('')
const qrCode = ref('')
const generatingQr = ref(false)
const origin = useRequestURL().origin
const bookingPageURL = computed(() =>
  bookingPage.value ? `${origin}/book/${bookingPage.value.slug}` : '',
)

onMounted(() => {
  if (route.path === '/dashboard') load()
})

async function load() {
  if (isCustomer.value || !currentCompany.value) {
    detail.value = null
    return
  }
  loading.value = true
  error.value = ''
  bookingPageError.value = ''
  driverError.value = ''
  marketplaceSetupError.value = ''
  metricsError.value = ''
  marketplaceZones.value = null
  marketplaceActiveRoutes.value = null
  overview.value = null
  requests.value = []
  try {
    detail.value = await api<CompanyDetail>('/companies/me')
    bookingPage.value = null
    if (
      detail.value.company.status === 'verified' &&
      detail.value.membership.permissions.includes('bookings.manage')
    ) {
      try {
        const result = await api<{ booking_page: BookingPage | null }>('/companies/me/booking-page')
        bookingPage.value = result.booking_page
      } catch (err) {
        bookingPageError.value = apiErrorMessage(err, 'Could not load your booking page.')
      }
      if (detail.value.membership.permissions.includes('staff.view')) {
        try {
          const result = await api<{ drivers: Driver[] }>('/companies/me/drivers')
          drivers.value = result.drivers
        } catch (err) {
          driverError.value = apiErrorMessage(err, 'Could not load your driver list.')
        }
      }
    }
    if (
      detail.value.company.status === 'verified' &&
      detail.value.membership.permissions.includes('shipments.view')
    ) {
      try {
        const result = await api<{ overview: CompanyOverview }>('/companies/me/overview')
        overview.value = result.overview
      } catch (err) {
        metricsError.value = apiErrorMessage(err, 'Could not load your delivery overview.')
      }
    }
    // Requests are a bookings-page subset of deliveries. Only fetch them when the
    // member can see bookings but not the wider delivery list.
    if (
      detail.value.company.status === 'verified' &&
      !detail.value.membership.permissions.includes('shipments.view') &&
      detail.value.membership.permissions.includes('bookings.manage')
    ) {
      try {
        const result = await api<{ requests: OverviewBooking[] }>('/companies/me/requests')
        requests.value = result.requests ?? []
      } catch (err) {
        metricsError.value = apiErrorMessage(err, 'Could not load booking requests.')
      }
    }
    if (
      detail.value.company.status === 'verified' &&
      detail.value.membership.permissions.includes('rates.view')
    ) {
      try {
        const [zonesResult, routesResult] = await Promise.all([
          api<{ zones: unknown[] }>('/companies/me/service-zones'),
          api<{ routes: { status: string }[] }>('/companies/me/zone-routes'),
        ])
        marketplaceZones.value = zonesResult.zones?.length ?? 0
        marketplaceActiveRoutes.value = (routesResult.routes ?? []).filter(
          (zoneRoute) => zoneRoute.status === 'active',
        ).length
      } catch (err) {
        marketplaceSetupError.value = apiErrorMessage(
          err,
          'Could not check your marketplace setup.',
        )
      }
    }
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not load your workspace.')
  } finally {
    loading.value = false
  }
}

async function copyBookingLink() {
  bookingShareError.value = ''
  try {
    await navigator.clipboard.writeText(bookingPageURL.value)
    copiedBookingLink.value = true
    window.setTimeout(() => (copiedBookingLink.value = false), 1800)
  } catch {
    bookingShareError.value = 'Could not copy the link. Copy it from the address shown instead.'
  }
}

async function shareBookingLink() {
  bookingShareError.value = ''
  if (navigator.share) {
    try {
      await navigator.share({
        title: `Book with ${detail.value?.company.name ?? 'us'}`,
        text: 'Book your next delivery with us.',
        url: bookingPageURL.value,
      })
      return
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') return
      bookingShareError.value = 'Could not open the share menu. You can copy the link instead.'
      return
    }
  }
  await copyBookingLink()
}

async function generateBookingQrCode() {
  bookingShareError.value = ''
  generatingQr.value = true
  try {
    qrCode.value = await QRCode.toDataURL(bookingPageURL.value, { width: 256, margin: 2 })
  } catch {
    bookingShareError.value = 'Could not create the QR code. Please try again.'
  } finally {
    generatingQr.value = false
  }
}

function downloadBookingQrCode() {
  if (!qrCode.value || !bookingPage.value) return
  const link = document.createElement('a')
  link.href = qrCode.value
  link.download = `${bookingPage.value.slug}-booking-page.png`
  link.click()
}

const checklist = computed(() => {
  const company = detail.value?.company
  const items = [
    {
      label: 'Verify your email',
      body: 'Confirms ownership of the account email.',
      done: Boolean(user.value?.email_verified),
      to: '/verify-email',
    },
    {
      label: 'Add the company profile',
      body: 'Business type, contact and registered location.',
      done: Boolean(company?.address && company?.city && company?.state),
      to: '/onboarding',
    },
    {
      label: 'Configure branding',
      body: 'Company colours and logo for your marketplace profile.',
      done: Boolean(detail.value?.branding),
      to: '/onboarding',
    },
    {
      label: 'Submit for verification',
      body: 'Send the profile for review before publishing rates.',
      done: Boolean(company && ['under_review', 'verified'].includes(company.status)),
      to: '/onboarding',
    },
    {
      label: 'Company verified',
      body: 'Unlocks public rates and marketplace bookings.',
      done: company?.status === 'verified',
      to: '/onboarding',
    },
  ]
  if (detail.value?.membership.permissions.includes('rates.view')) {
    items.push({
      label: 'Set up zones & routes',
      body: 'Required to appear in marketplace searches.',
      done: (marketplaceActiveRoutes.value ?? 0) > 0,
      to: '/dashboard/company/rates',
    })
  }
  return items
})

const isVerified = computed(() => detail.value?.company.status === 'verified')
const canViewFleet = computed(() => detail.value?.membership.permissions.includes('shipments.view') ?? false)
const completed = computed(() => checklist.value.filter((item) => item.done).length)
const progress = computed(() => Math.round((completed.value / checklist.value.length) * 100))

function formatCurrency(amount: number) {
  const currency = detail.value?.settings.currency?.trim() || 'NGN'
  return `${currency} ${amount.toLocaleString('en-NG')}`
}

const inProgressStatuses = ['assigned', 'picked_up', 'in_transit']
const waitingStatuses = ['created', 'requested']

function isToday(value: string) {
  if (!value) return false
  const date = new Date(value)
  const now = new Date()
  return (
    date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth() &&
    date.getDate() === now.getDate()
  )
}

function isCancelled(booking: OverviewBooking) {
  return booking.status === 'cancelled' || booking.request_status === 'declined'
}

const recordCount = computed(
  () => overview.value?.bookings.total ?? requests.value.length,
)

const financialMetrics = computed(() => {
  const financial = overview.value?.financial
  return [
    {
      label: 'Earned today',
      value: formatCurrency((financial?.earned_today_minor ?? 0) / 100),
      detail: 'From deliveries completed today',
    },
    {
      label: 'Collected today',
      value: formatCurrency((financial?.collected_today_minor ?? 0) / 100),
      detail: 'Payments captured today',
    },
    {
      label: 'Amount owed',
      value: formatCurrency((financial?.amount_owed_minor ?? 0) / 100),
      detail: 'Outstanding delivery payments',
    },
    {
      label: 'Cash with drivers',
      value: formatCurrency((financial?.cash_with_drivers_minor ?? 0) / 100),
      detail: 'Cash awaiting reconciliation',
    },
  ]
})

const bookingMetrics = computed(() => {
  if (overview.value) {
    const bookings = overview.value.bookings
    return [
      { label: 'Bookings created today', value: bookings.created_today },
      { label: 'In progress', value: bookings.in_progress },
      { label: 'Waiting for drivers', value: bookings.waiting_for_drivers },
      { label: 'Booking requests', value: bookings.requests },
    ]
  }
  // Members without shipments.view only get the capped requests list.
  const list = requests.value
  return [
    {
      label: 'Bookings created today',
      value: list.filter((booking) => isToday(booking.created_at)).length,
    },
    {
      label: 'In progress',
      value: list.filter(
        (booking) => inProgressStatuses.includes(booking.status) && !isCancelled(booking),
      ).length,
    },
    {
      label: 'Waiting for drivers',
      value: list.filter(
        (booking) => waitingStatuses.includes(booking.status) && !isCancelled(booking),
      ).length,
    },
    {
      label: 'Booking requests',
      value: list.filter((booking) => booking.request_status === 'pending').length,
    },
  ]
})

const financialTones = ['brand', 'accent', 'success', 'neutral'] as const
const bookingTones = ['neutral', 'accent', 'success', 'brand'] as const

const analyticsDays = computed(() => {
  if (overview.value?.daily?.length) {
    return overview.value.daily.map((day) => ({
      label: new Date(`${day.date}T12:00:00`).toLocaleDateString('en-NG', { weekday: 'short' }),
      count: day.bookings,
    }))
  }
  return Array.from({ length: 7 }, (_, index) => {
    const start = new Date()
    start.setHours(0, 0, 0, 0)
    start.setDate(start.getDate() - 6 + index)
    const end = new Date(start)
    end.setDate(end.getDate() + 1)
    const count = requests.value.filter((booking) => {
      if (!booking.created_at) return false
      const created = new Date(booking.created_at)
      return created >= start && created < end
    }).length
    return {
      label: start.toLocaleDateString('en-NG', { weekday: 'short' }),
      count,
    }
  })
})

const analyticsTotal = computed(() =>
  analyticsDays.value.reduce((total, day) => total + day.count, 0),
)

const analyticsPoints = computed(() => {
  const max = Math.max(1, ...analyticsDays.value.map((day) => day.count))
  return analyticsDays.value.map((day, index) => ({
    x: 30 + index * 110,
    y: 78 - Math.round((day.count / max) * 54),
    count: day.count,
  }))
})

const analyticsPath = computed(() =>
  analyticsPoints.value
    .map((point, index) => `${index === 0 ? 'M' : 'L'}${point.x} ${point.y}`)
    .join(' '),
)
</script>

<template>
  <PwaInstallBanner v-if="route.path === '/dashboard'" class="mb-4" />

  <NuxtPage v-if="route.path !== '/dashboard'" />

  <CustomerDashboard v-else-if="isCustomer" />

  <div v-else class="dashboard-overview">
    <div class="overview-hero relative overflow-hidden rounded-2xl border p-5 sm:p-6">
      <div class="pointer-events-none absolute -top-16 right-10 size-44 rounded-full bg-accent-400/15 blur-3xl" />
      <div class="pointer-events-none absolute -right-8 -bottom-24 size-52 rounded-full bg-brand-400/15 blur-3xl" />
      <div class="relative flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="inline-flex items-center gap-1.5 rounded-full bg-surface/75 px-2.5 py-1 text-[10px] font-semibold tracking-wide text-brand-700 dark:bg-white/5">
          <Sparkles :size="12" /> TODAY AT A GLANCE
        </p>
        <div class="mt-3 flex flex-wrap items-center gap-2.5">
          <h1 class="text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">
            Welcome, {{ user?.full_name?.split(' ')[0] || 'there' }}
          </h1>
          <span
            v-if="currentCompany && isVerified"
            class="inline-flex items-center gap-1 rounded-full border border-success-200/60 bg-success-100/70 px-2.5 py-1 text-[10px] font-semibold text-success-700 dark:border-success-500/25 dark:bg-success-500/10 dark:text-success-200"
          >
            <ShieldCheck :size="13" /> Verified
          </span>
        </div>
        <p class="mt-2 max-w-xl text-[12.5px] leading-relaxed text-ink-900/65">
          {{ isVerified
            ? 'A quick snapshot of your delivery business. Let’s make today a good one.'
            : 'Finish setting up your company to unlock your provider workspace.' }}
        </p>
      </div>
      <AppButton v-if="currentCompany && !isVerified" to="/onboarding" size="sm">
        Continue onboarding
        <ArrowRight :size="15" />
      </AppButton>
      </div>
    </div>

    <div v-if="loading" class="overview-card mt-5 flex items-center gap-2 rounded-xl border p-5 text-[12px] text-ink-900/50">
      <Loader2 :size="16" class="animate-spin" /> Loading workspace…
    </div>

    <p v-else-if="error" role="alert" class="mt-5 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 p-4 text-[12px] text-red-700">
      <CircleAlert :size="16" class="mt-0.5 shrink-0" />
      {{ error }}
    </p>

    <template v-else-if="detail">
      <section v-if="isVerified" class="mt-5 space-y-5" aria-label="Company operations overview">
        <section
          v-if="detail.membership.permissions.includes('rates.view') && marketplaceActiveRoutes === 0"
          class="overview-nudge flex flex-col gap-4 rounded-xl border p-4 sm:flex-row sm:items-center sm:justify-between"
          aria-labelledby="marketplace-setup-title"
        >
          <div class="flex items-start gap-3">
            <span class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-surface text-accent-700 dark:bg-white/5 dark:text-accent-300">
              <Route :size="17" />
            </span>
            <div>
              <h2 id="marketplace-setup-title" class="text-[13px] font-semibold">
                Get listed in the marketplace
              </h2>
              <p class="mt-1 max-w-2xl text-[11.5px] leading-relaxed opacity-75">
                Customers only see providers with an active zone route. You have added
                {{ marketplaceZones ?? 0 }} zone{{ (marketplaceZones ?? 0) === 1 ? '' : 's' }} — publish at least
                one direction to appear in landing page searches.
              </p>
            </div>
          </div>
          <NuxtLink
            to="/dashboard/company/rates"
            class="inline-flex min-h-9 shrink-0 items-center justify-center gap-1.5 rounded-lg bg-accent-500 px-3.5 py-2 text-[11px] font-semibold text-white transition hover:bg-accent-400"
          >
            Set up zones & routes <ArrowRight :size="14" />
          </NuxtLink>
        </section>
        <p v-if="marketplaceSetupError" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[11px] text-red-700">
          {{ marketplaceSetupError }}
        </p>
        <section
          v-if="detail.membership.permissions.includes('staff.invite') && detail.membership.permissions.includes('staff.view') && !driverError && drivers.filter((driver) => driver.status === 'active').length === 0"
          class="overview-nudge flex flex-col gap-4 rounded-xl border p-4 sm:flex-row sm:items-center sm:justify-between"
          aria-labelledby="first-driver-title"
        >
          <div class="flex items-start gap-3">
            <span class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-surface text-accent-700 dark:bg-white/5 dark:text-accent-300">
              <Users :size="17" />
            </span>
            <div>
              <h2 id="first-driver-title" class="text-[13px] font-semibold">
                Add your first driver
              </h2>
              <p class="mt-1 max-w-2xl text-[11.5px] leading-relaxed opacity-75">
                Create a driver portal so your team is ready when you start assigning deliveries.
              </p>
            </div>
          </div>
          <NuxtLink
            :to="{ path: '/dashboard/company/drivers', query: { create: '1' } }"
            class="inline-flex min-h-9 shrink-0 items-center justify-center gap-1.5 rounded-lg bg-accent-500 px-3.5 py-2 text-[11px] font-semibold text-white transition hover:bg-accent-400"
          >
            Create first driver <ArrowRight :size="14" />
          </NuxtLink>
        </section>
        <p v-if="driverError" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[11px] text-red-700">
          {{ driverError }}
        </p>

        <section
          v-if="detail.membership.permissions.includes('bookings.manage')"
          class="overview-share overflow-hidden rounded-2xl border"
          aria-labelledby="booking-share-title"
        >
          <div class="grid gap-5 p-4 sm:p-5 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="text-[15px] font-semibold tracking-[-0.02em]">
                  {{ bookingPage?.is_published ? 'Your booking page is live' : 'Your booking page isn’t live yet' }}
                </h2>
                <span
                  v-if="bookingPage?.is_published"
                  class="inline-flex items-center gap-1 rounded-full bg-success-50 px-2 py-1 text-[9.5px] font-semibold text-success-700"
                >
                  <CheckCircle2 :size="12" /> Live
                </span>
              </div>
              <h3 id="booking-share-title" class="mt-2 text-[12px] font-semibold text-brand-700 dark:text-brand-200">
                Get more bookings: share your page
              </h3>
              <p class="mt-1 max-w-2xl text-[11.5px] leading-relaxed text-ink-900/55">
                Post your link to your WhatsApp Status and groups. Every booking lands in Requests and pings your WhatsApp.
              </p>
              <template v-if="bookingPageError">
                <p role="alert" class="mt-3 text-[11px] text-red-700">{{ bookingPageError }}</p>
              </template>
              <template v-else-if="bookingPage?.is_published">
                <a
                  :href="bookingPageURL"
                  target="_blank"
                  rel="noopener"
                  class="mt-3 inline-flex max-w-full items-center gap-1.5 break-all text-[11px] font-medium text-brand-700 underline decoration-brand-300 underline-offset-4"
                >
                  {{ bookingPageURL }} <ExternalLink :size="12" class="shrink-0" />
                </a>
                <p v-if="bookingShareError" role="alert" class="mt-2 text-[11px] text-red-700">
                  {{ bookingShareError }}
                </p>
              </template>
              <template v-else>
                <p class="mt-3 text-[11px] text-accent-800 dark:text-accent-200">
                  Your page isn’t published. Publish it from the booking page settings before sharing.
                </p>
              </template>
            </div>

            <div class="grid grid-cols-2 gap-2 sm:flex sm:flex-wrap md:justify-end">
              <template v-if="bookingPage?.is_published">
                <button
                  type="button"
                  class="inline-flex min-h-10 items-center justify-center gap-1.5 rounded-lg bg-accent-500 px-3 py-2 text-[11.5px] font-semibold text-white transition hover:bg-accent-400"
                  @click="shareBookingLink"
                >
                  <Share2 :size="14" /> Share
                </button>
                <button
                  type="button"
                  class="inline-flex min-h-10 items-center justify-center gap-1.5 rounded-lg border border-ink-900/10 bg-surface px-3 py-2 text-[11.5px] font-semibold text-ink-900/70 transition hover:border-accent-500/40"
                  @click="copyBookingLink"
                >
                  <Check v-if="copiedBookingLink" :size="14" class="text-success-600" />
                  <Copy v-else :size="14" />
                  {{ copiedBookingLink ? 'Copied' : 'Copy link' }}
                </button>
                <button
                  type="button"
                  :disabled="generatingQr"
                  class="col-span-2 inline-flex min-h-10 items-center justify-center gap-1.5 rounded-lg border border-ink-900/10 bg-surface px-3 py-2 text-[11.5px] font-semibold text-ink-900/70 transition hover:border-accent-500/40 disabled:opacity-50 sm:col-span-1"
                  @click="generateBookingQrCode"
                >
                  <Loader2 v-if="generatingQr" :size="14" class="animate-spin" />
                  <Download v-else :size="14" />
                  QR code
                </button>
              </template>
            </div>
          </div>

          <div v-if="qrCode && bookingPage?.is_published" class="border-t border-brand-100 bg-surface/70 px-5 py-4 sm:px-6 dark:bg-white/5">
            <div class="flex flex-wrap items-center gap-4">
              <img :src="qrCode" alt="QR code for your booking page" class="size-32 rounded-lg border border-ink-900/10 bg-surface p-2" />
              <div>
                <p class="text-[12px] font-semibold">Share your page in person</p>
                <p class="mt-1 text-[11px] text-ink-900/45">Customers can scan this code to open your booking page.</p>
                <button
                  type="button"
                  class="mt-3 inline-flex items-center gap-1.5 text-[11px] font-semibold text-brand-700 hover:text-brand-600"
                  @click="downloadBookingQrCode"
                >
                  <Download :size="13" /> Download QR code
                </button>
              </div>
            </div>
          </div>
        </section>

        <p v-if="metricsError" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[11px] text-red-700">
          {{ metricsError }}
        </p>

        <div>
          <div class="mb-3 flex flex-wrap items-end justify-between gap-2">
            <div>
              <h2 class="text-[14px] font-semibold">Financial overview</h2>
              <p class="mt-1 text-[11px] text-ink-900/40">Daily delivery earnings and cash position</p>
            </div>
            <span class="text-[10px] text-ink-900/35">
              {{ recordCount ? `${recordCount} delivery records` : 'No financial records yet' }}
            </span>
          </div>
          <div class="grid grid-cols-2 gap-3 xl:grid-cols-4">
            <DashboardStatCard
              v-for="(metric, index) in financialMetrics"
              :key="metric.label"
              :label="metric.label"
              :value="metric.value"
              :detail="metric.detail"
              :tone="financialTones[index % financialTones.length]"
            />
          </div>
        </div>

        <div>
          <div class="mb-3 flex flex-wrap items-end justify-between gap-2">
            <div>
              <h2 class="text-[14px] font-semibold">Booking activity</h2>
              <p class="mt-1 text-[11px] text-ink-900/40">Today’s workload and requests</p>
            </div>
            <span class="text-[10px] text-ink-900/35">
              {{ recordCount ? `${recordCount} booking records` : 'No booking records yet' }}
            </span>
          </div>
          <div class="grid grid-cols-2 gap-3 xl:grid-cols-4">
            <DashboardStatCard
              v-for="(metric, index) in bookingMetrics"
              :key="metric.label"
              :label="metric.label"
              :value="metric.value"
              :tone="bookingTones[index % bookingTones.length]"
            />
          </div>
        </div>

        <DriversOnRoad v-if="canViewFleet" />

        <section class="overview-card rounded-xl border p-4 sm:p-5" aria-labelledby="analytics-title">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h2 id="analytics-title" class="text-[14px] font-semibold">Last 7 days</h2>
              <p class="mt-1 text-[11px] text-ink-900/40">Booking activity over the past week</p>
            </div>
            <span class="rounded-md bg-canvas px-2 py-1 text-[10px] font-medium text-ink-900/50">
              {{ analyticsTotal }} {{ analyticsTotal === 1 ? 'booking' : 'bookings' }}
            </span>
          </div>
          <div class="mt-4">
            <svg
              viewBox="0 0 700 110"
              class="h-24 w-full"
              role="img"
              :aria-label="`${analyticsTotal} bookings created in the last seven days`"
              preserveAspectRatio="none"
            >
              <line x1="30" y1="24" x2="690" y2="24" stroke="currentColor" class="text-ink-900/5" />
              <line x1="30" y1="51" x2="690" y2="51" stroke="currentColor" class="text-ink-900/5" />
              <line x1="30" y1="78" x2="690" y2="78" stroke="currentColor" class="text-ink-900/10" />
              <path :d="analyticsPath" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="text-brand-500" />
              <circle
                v-for="(point, index) in analyticsPoints"
                :key="index"
                :cx="point.x"
                :cy="point.y"
                r="3"
                fill="currentColor"
                class="text-brand-500"
              >
                <title>{{ point.count }} on {{ analyticsDays[index]?.label }}</title>
              </circle>
            </svg>
            <div class="grid grid-cols-7 text-center text-[9.5px] text-ink-900/35">
              <span v-for="(day, index) in analyticsDays" :key="`${day.label}-${index}`">{{ day.label }}</span>
            </div>
          </div>
          <p class="mt-3 text-center text-[10.5px] text-ink-900/35">
            {{ analyticsTotal
              ? 'Bookings created per day over the last week.'
              : 'Analytics will populate when booking records are available.' }}
          </p>
        </section>
      </section>

      <div v-if="!isVerified" class="mt-5 max-w-3xl">
        <div class="overview-card rounded-2xl border p-5 sm:p-6">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-[14px] font-semibold">A few steps to get rolling</h2>
                <p class="mt-1 text-[11px] text-ink-900/45">Your setup progress</p>
              </div>
              <span class="rounded-full bg-brand-50 px-2.5 py-1 text-[10px] font-semibold text-brand-700 dark:bg-brand-100 dark:text-brand-700">
                {{ progress }}%
              </span>
            </div>
            <div class="mt-4 h-2 overflow-hidden rounded-full bg-ink-900/8">
              <div class="h-full rounded-full bg-gradient-to-r from-accent-500 to-accent-400 transition-all" :style="{ width: `${progress}%` }" />
            </div>

            <ul class="mt-3 divide-y divide-ink-900/8">
              <li v-for="item in checklist" :key="item.label" class="flex items-start gap-3 py-2.5">
                <span
                  class="mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full border"
                  :class="item.done ? 'border-success-200 bg-success-50 text-success-600' : 'border-ink-900/15 text-ink-900/25'"
                >
                  <Check v-if="item.done" :size="11" :stroke-width="3" />
                  <CircleDot v-else :size="9" />
                </span>
                <div class="min-w-0 flex-1">
                  <p class="text-[12px] font-medium" :class="item.done ? 'text-ink-900/45' : ''">
                    {{ item.label }}
                  </p>
                  <p class="mt-0.5 text-[11px] leading-relaxed text-ink-900/40">{{ item.body }}</p>
                </div>
                <NuxtLink
                  v-if="!item.done"
                  :to="item.to"
                  class="mt-0.5 shrink-0 text-[10.5px] font-semibold text-brand-600 hover:text-brand-700"
                >
                  Start
                </NuxtLink>
              </li>
            </ul>
        </div>
      </div>
    </template>

    <template v-else>
      <div class="mt-8 grid gap-4 sm:grid-cols-2">
        <div class="rounded-xl border border-ink-900/10 bg-surface p-5">
          <Sparkles :size="18" class="text-brand-600" />
          <h2 class="mt-3 text-[14px] font-semibold">Marketplace account ready</h2>
          <p class="mt-1.5 text-[12.5px] leading-relaxed text-ink-900/50">
            Rate comparison and shipment booking for customers arrive in the marketplace stage.
            Your account already has secure login, email verification and password reset.
          </p>
          <div class="mt-4 flex flex-wrap gap-2">
            <AppButton to="/" size="sm" variant="secondary">Explore the public site</AppButton>
          </div>
        </div>

        <div class="rounded-xl border border-ink-900/10 bg-surface p-5">
          <Building2 :size="18" class="text-brand-600" />
          <h2 class="mt-3 text-[14px] font-semibold">Operate a logistics company?</h2>
          <p class="mt-1.5 text-[12.5px] leading-relaxed text-ink-900/50">
            A company must be registered as a separate account with its own owner email. Sign out
            and choose the logistics company option on the registration page.
          </p>
          <div class="mt-4">
            <AppButton to="/register" size="sm">Register a company</AppButton>
          </div>
        </div>
      </div>

      <div class="mt-5 rounded-xl border border-ink-900/10 bg-surface p-5">
        <h2 class="text-[13px] font-semibold">Account privacy</h2>
        <p class="mt-1.5 text-[11.5px] leading-relaxed text-ink-900/45">
          Your customer account has no company permissions. Company staff roles are always
          attached to a specific company membership.
        </p>
        <div class="mt-4 flex items-center gap-2 text-[11px] text-ink-900/40">
          <MailCheck :size="13" />
          {{ user?.email_verified ? 'Email verified' : 'Email verification pending' }}
          <span class="text-ink-900/20">·</span>
          <ClipboardList :size="13" />
          {{ user?.account_type }} account
        </div>
      </div>
    </template>
  </div>
</template>
