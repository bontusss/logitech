<script setup lang="ts">
import {
  ArrowLeft,
  ArrowRight,
  BadgeCheck,
  Check,
  CheckCircle2,
  CircleAlert,
  Loader2,
  MapPin,
  Package,
  Send,
  ShieldCheck,
  Sparkles,
  Upload,
  UserRound,
  Weight,
  X,
} from 'lucide-vue-next'
import type { BookingLocation } from '~/components/BookingLocationField.vue'

definePageMeta({ layout: false })

interface PublicBookingPage {
  booking_page: {
    slug: string
    headline: string | null
    description: string | null
  }
  company: {
    id: string
    name: string
    slug: string
    public_name: string
    business_type: string
    business_email: string
    phone: string
    city: string | null
    state: string | null
    country: string
    currency: string
  }
  branding: {
    logo_url: string | null
    primary_color: string
    secondary_color: string
    cover_image_url: string | null
  }
  subdomain: string
}

interface PublicRoute {
  id: string
  origin_state: string
  origin_city: string
  destination_state: string
  destination_city: string
  distance_km: number
  standard_price_per_km_minor: number | null
  express_price_per_km_minor: number | null
  same_day_price_per_km_minor: number | null
  minimum_charge_minor: number
  currency: string
}

type BookingStep = 0 | 1 | 2 | 3
type PackageCategory = 'cloth' | 'electronics' | 'document' | 'food' | 'jewelry' | 'other'

const route = useRoute()
const config = useRuntimeConfig()
const origin = useRequestURL().origin
const slug = computed(() => String(route.params.slug ?? ''))
const apiBase = (config.public.apiBase as string) || '/api/v1'

const { data, error: pageLoadError, status } = await useAsyncData(
  () => `booking-page:${slug.value}`,
  () =>
    $fetch<PublicBookingPage>(`/public/book/${encodeURIComponent(slug.value)}`, {
      baseURL: apiBase,
    }),
  { watch: [slug] },
)
const { data: routeData, error: routeLoadError } = await useAsyncData(
  () => `booking-routes:${slug.value}`,
  () =>
    $fetch<{ routes: PublicRoute[] }>(`/public/book/${encodeURIComponent(slug.value)}/routes`, {
      baseURL: apiBase,
    }),
  { watch: [slug], default: () => ({ routes: [] }) },
)

const company = computed(() => data.value?.company ?? null)
const branding = computed(() => data.value?.branding ?? null)
const primary = computed(() => /^#[0-9a-f]{6}$/i.test(branding.value?.primary_color ?? '')
  ? branding.value!.primary_color
  : '#2b4bff')
const secondary = computed(() => /^#[0-9a-f]{6}$/i.test(branding.value?.secondary_color ?? '')
  ? branding.value!.secondary_color
  : '#10172b')
const displayName = computed(() => company.value?.public_name || company.value?.name || '')
const firstLetter = computed(() => Array.from(displayName.value.trim())[0]?.toLocaleUpperCase() || 'B')
const headline = computed(
  () => data.value?.booking_page.headline || `Book a delivery with ${displayName.value}`,
)
const description = computed(
  () => data.value?.booking_page.description || 'Choose a route, tell us about your package, and share where it needs to go.',
)
const location = computed(() =>
  [company.value?.city, company.value?.state, company.value?.country].filter(Boolean).join(', '),
)
const pageURL = computed(() => `${origin}/book/${data.value?.booking_page.slug || slug.value}`)
const phoneHref = computed(() => `tel:${(company.value?.phone || '').replace(/[^+\d]/g, '')}`)
const routes = computed(() => routeData.value?.routes ?? [])
const currency = computed(() => company.value?.currency || 'NGN')
const routeMode = ref<'route' | 'map'>('route')
const currentStep = ref<BookingStep>(0)
const saving = ref(false)
const uploading = ref(false)
const error = ref('')
const bookingReference = ref('')
const submittedPriceMinor = ref<number | null>(null)
const submittedPriceIsEstimate = ref(true)
const configPublic = config.public as Record<string, unknown>
const cloudinaryCloudName = String(configPublic.cloudinaryCloudName || '')
const cloudinaryUploadPreset = String(configPublic.cloudinaryUploadPreset || '')
const photoUploadEnabled = computed(() => Boolean(cloudinaryCloudName && cloudinaryUploadPreset))

const categoryOptions: { value: PackageCategory; label: string; emoji: string; hint: string }[] = [
  { value: 'cloth', label: 'Clothing', emoji: '👕', hint: 'Fashion & fabric' },
  { value: 'electronics', label: 'Electronics', emoji: '🎧', hint: 'Devices & accessories' },
  { value: 'document', label: 'Documents', emoji: '📄', hint: 'Paperwork & files' },
  { value: 'food', label: 'Food', emoji: '🍱', hint: 'Meals & groceries' },
  { value: 'jewelry', label: 'Jewelry', emoji: '✨', hint: 'Watches & valuables' },
  { value: 'other', label: 'Other', emoji: '📦', hint: 'Something else' },
]

const form = reactive({
  selectedRouteId: '',
  deliverySpeed: 'standard' as 'standard' | 'express' | 'same_day',
  pickup: null as BookingLocation | null,
  dropoff: null as BookingLocation | null,
  category: '' as PackageCategory | '',
  declaredValue: '',
  weightKg: '',
  instructions: '',
  photoUrl: '',
  senderName: '',
  senderPhone: '',
  receiverName: '',
  receiverPhone: '',
})

const routeSelected = computed(() => routes.value.find((item) => item.id === form.selectedRouteId) ?? null)

watch(
  routes,
  (availableRoutes) => {
    routeMode.value = availableRoutes.length ? 'route' : 'map'
  },
  { immediate: true },
)

const faviconHref = computed(() => {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><rect width="64" height="64" rx="18" fill="${primary.value}"/><text x="50%" y="54%" dominant-baseline="middle" text-anchor="middle" font-family="Arial,sans-serif" font-size="38" font-weight="700" fill="#fff">${firstLetter.value}</text></svg>`
  return `data:image/svg+xml,${encodeURIComponent(svg)}`
})

useHead(() => ({
  link: [{ key: 'booking-favicon', rel: 'icon', href: faviconHref.value }],
  htmlAttrs: { style: `--booking-primary:${primary.value};--booking-primary-soft:${primary.value}18` },
}))

useSeoMeta({
  title: () => data.value ? `${headline.value} · ${displayName.value}` : 'Book a delivery',
  description: () => description.value,
  ogTitle: () => headline.value,
  ogDescription: () => description.value,
  ogType: 'website',
  ogUrl: () => pageURL.value,
  twitterCard: 'summary',
})

function formatMoney(minor: number) {
  return formatCurrency(minor, currency.value)
}

function formatCurrency(minor: number, currencyCode: string) {
  return new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency: currencyCode,
    maximumFractionDigits: 2,
  }).format(minor / 100)
}

const declaredValueMinor = computed(() => {
  const amount = Number(form.declaredValue)
  return Number.isFinite(amount) && amount >= 0 ? Math.round(amount * 100) : 0
})
const pickedRoute = computed(() => routeMode.value === 'route' ? routeSelected.value : null)
const pickupText = computed(() => pickedRoute.value
  ? `${pickedRoute.value.origin_city}, ${pickedRoute.value.origin_state}`
  : form.pickup?.address || '')
const dropoffText = computed(() => pickedRoute.value
  ? `${pickedRoute.value.destination_city}, ${pickedRoute.value.destination_state}`
  : form.dropoff?.address || '')
const servicePricePerKm = computed(() => {
  if (!routeSelected.value) return null
  return {
    standard: routeSelected.value.standard_price_per_km_minor,
    express: routeSelected.value.express_price_per_km_minor,
    same_day: routeSelected.value.same_day_price_per_km_minor,
  }[form.deliverySpeed]
})
const estimatedDeliveryPrice = computed(() => {
  if (!routeSelected.value || servicePricePerKm.value === null) return null
  return Math.max(
    routeSelected.value.minimum_charge_minor,
    Math.ceil(servicePricePerKm.value * routeSelected.value.distance_km),
  )
})
const serviceLabels = {
  standard: 'Standard',
  express: 'Express',
  same_day: 'Same-day',
} as const
const deliveryServices = computed(() => {
  const providerRoute = routeSelected.value
  if (!providerRoute) return []
  return [
    { key: 'standard' as const, label: 'Standard', price: providerRoute.standard_price_per_km_minor },
    { key: 'express' as const, label: 'Express', price: providerRoute.express_price_per_km_minor },
    { key: 'same_day' as const, label: 'Same-day', price: providerRoute.same_day_price_per_km_minor },
  ].filter((service) => service.price !== null)
})

function nextStep() {
  error.value = ''
  if (currentStep.value === 0) {
    if (routeMode.value === 'route' && !form.selectedRouteId) {
      error.value = 'Choose one of the available routes, or enter pickup and drop-off locations.'
      return
    }
    if (routeMode.value === 'map' && (!form.pickup || !form.dropoff)) {
      error.value = 'Choose both locations from the Mapbox suggestions to confirm their pins.'
      return
    }
    if (routeMode.value === 'route' && routeSelected.value && servicePricePerKm.value === null) {
      error.value = 'Choose a delivery speed offered on this route.'
      return
    }
  } else if (currentStep.value === 1) {
    if (!form.category) {
      error.value = 'Choose a package type to continue.'
      return
    }
    if (!form.declaredValue || declaredValueMinor.value < 0) {
      error.value = 'Add the approximate value of the item.'
      return
    }
    const weight = Number(form.weightKg)
    if (!Number.isFinite(weight) || weight <= 0 || weight > 500) {
      error.value = 'Enter a package weight between 0.1 and 500 kg.'
      return
    }
  } else if (currentStep.value === 2) {
    if (form.senderName.trim().length < 2 || form.senderPhone.trim().length < 7) {
      error.value = 'Add the sender’s name and a valid phone number.'
      return
    }
    if (form.receiverName.trim().length < 2 || form.receiverPhone.trim().length < 7) {
      error.value = 'Add the receiver’s name and a valid phone number.'
      return
    }
  }
  currentStep.value = Math.min(3, currentStep.value + 1) as BookingStep
}

function previousStep() {
  error.value = ''
  currentStep.value = Math.max(0, currentStep.value - 1) as BookingStep
}

function selectRoute(providerRoute: PublicRoute) {
  form.selectedRouteId = providerRoute.id
  const availableSpeeds = [
    ['standard', providerRoute.standard_price_per_km_minor],
    ['express', providerRoute.express_price_per_km_minor],
    ['same_day', providerRoute.same_day_price_per_km_minor],
  ] as const
  if (!availableSpeeds.some(([speed, price]) => speed === form.deliverySpeed && price !== null)) {
    const firstAvailable = availableSpeeds.find(([, price]) => price !== null)
    if (firstAvailable) form.deliverySpeed = firstAvailable[0]
  }
}

async function uploadPhoto(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  error.value = ''
  if (!file.type.startsWith('image/')) {
    error.value = 'Choose an image file such as JPG, PNG or WebP.'
    return
  }
  if (file.size > 10 * 1024 * 1024) {
    error.value = 'Package photos must be 10 MB or smaller.'
    return
  }
  if (!photoUploadEnabled.value) {
    error.value = 'Package photo upload is not configured yet. You can continue without a photo.'
    return
  }
  uploading.value = true
  try {
    const body = new FormData()
    body.append('file', file)
    body.append('upload_preset', cloudinaryUploadPreset)
    body.append('folder', 'bol-booking-items')
    const result = await $fetch<{ secure_url: string }>(
      `https://api.cloudinary.com/v1_1/${encodeURIComponent(cloudinaryCloudName)}/image/upload`,
      { method: 'POST', body },
    )
    if (!result.secure_url) throw new Error('The image host did not return a photo link.')
    form.photoUrl = result.secure_url
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not upload the photo. Try a smaller image or continue without one.')
  } finally {
    uploading.value = false
    ;(event.target as HTMLInputElement).value = ''
  }
}

async function createBooking() {
  error.value = ''
  saving.value = true
  try {
    const response = await $fetch<{
      booking: {
        reference: string
        delivery_price_minor?: number
        price_is_estimate?: boolean
      }
    }>(
      `/public/book/${encodeURIComponent(slug.value)}/bookings`,
      {
        baseURL: apiBase,
        method: 'POST',
        body: {
          sender_name: form.senderName.trim(),
          sender_phone: form.senderPhone.trim(),
          receiver_name: form.receiverName.trim(),
          receiver_phone: form.receiverPhone.trim(),
          pickup_address: pickupText.value,
          dropoff_address: dropoffText.value,
          package_category: form.category,
          declared_value_minor: declaredValueMinor.value,
          weight_kg: Number(form.weightKg),
          additional_instructions: form.instructions.trim(),
          package_photo_url: form.photoUrl,
          pickup_latitude: routeMode.value === 'map' ? form.pickup?.latitude : null,
          pickup_longitude: routeMode.value === 'map' ? form.pickup?.longitude : null,
          dropoff_latitude: routeMode.value === 'map' ? form.dropoff?.latitude : null,
          dropoff_longitude: routeMode.value === 'map' ? form.dropoff?.longitude : null,
          selected_route_id: routeMode.value === 'route' ? form.selectedRouteId : null,
          delivery_speed: form.deliverySpeed,
        },
      },
    )
    bookingReference.value = response.booking.reference
    submittedPriceMinor.value = response.booking.delivery_price_minor ?? null
    submittedPriceIsEstimate.value = response.booking.price_is_estimate ?? true
    currentStep.value = 3
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not send this booking request. Please try again.')
  } finally {
    saving.value = false
  }
}

const bookingCode = computed(() => bookingReference.value
  ? `REQ-${bookingReference.value.slice(0, 8).toUpperCase()}`
  : '')
</script>

<template>
  <div
    class="public-booking min-h-dvh bg-[#f7f8fc] text-slate-900"
    :style="{ '--booking-primary': primary, '--booking-primary-soft': `${primary}18` }"
  >
    <header class="relative isolate overflow-hidden" :style="{ backgroundColor: secondary }">
      <img
        v-if="branding?.cover_image_url"
        :src="branding.cover_image_url"
        alt=""
        class="absolute inset-0 -z-20 h-full w-full object-cover opacity-30"
      />
      <div class="absolute inset-0 -z-10" :style="{ background: `linear-gradient(115deg, ${secondary} 5%, ${primary} 145%)`, opacity: 0.96 }" />
      <div class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-5 py-4 sm:px-8 sm:py-5">
        <div class="flex min-w-0 items-center gap-3">
          <div
            class="flex size-11 shrink-0 items-center justify-center overflow-hidden rounded-2xl bg-white text-lg font-extrabold shadow-lg shadow-black/10"
            :style="{ color: primary }"
          >
            <img v-if="branding?.logo_url" :src="branding.logo_url" :alt="displayName" class="h-full w-full object-cover" />
            <span v-else>{{ firstLetter }}</span>
          </div>
          <div class="min-w-0 text-white">
            <p class="truncate text-[14px] font-bold tracking-tight">{{ displayName }}</p>
            <p class="mt-0.5 truncate text-[11px] text-white/65">{{ location || 'Local delivery, made simple' }}</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <div class="hidden items-center gap-2 rounded-full border border-white/15 bg-white/10 px-3 py-1.5 text-[10px] font-semibold text-white/90 sm:flex">
            <ShieldCheck :size="13" /> Secure booking
          </div>
          <ThemeToggle class="!size-9 !rounded-xl !border-white/30 !text-white hover:!bg-white/15" />
        </div>
      </div>
    </header>

    <main class="mx-auto max-w-6xl px-4 pb-14 sm:px-7">
      <div
        v-if="status === 'pending'"
        class="mx-auto mt-10 flex max-w-3xl items-center gap-2 rounded-2xl bg-white p-8 text-sm text-slate-500 shadow-sm"
      >
        <Loader2 :size="17" class="animate-spin" /> Loading {{ displayName || 'booking page' }}…
      </div>
      <section
        v-else-if="pageLoadError || !data || !company"
        class="mx-auto mt-10 max-w-2xl rounded-3xl border border-slate-200 bg-white p-8 text-center shadow-sm sm:p-12"
      >
        <span class="mx-auto flex size-12 items-center justify-center rounded-2xl bg-rose-50 text-rose-600"><CircleAlert :size="22" /></span>
        <h1 class="mt-4 text-xl font-bold tracking-tight">This booking page isn’t available</h1>
        <p class="mx-auto mt-2 max-w-sm text-sm leading-6 text-slate-500">The link may have changed. Please check with the business that shared it.</p>
        <NuxtLink to="/" class="mt-6 inline-flex items-center gap-1.5 text-sm font-semibold" :style="{ color: primary }">
          <ArrowLeft :size="14" /> Back to BOL
        </NuxtLink>
      </section>
      <template v-else>
        <div class="grid items-start gap-6 py-7 lg:grid-cols-[minmax(0,1fr)_340px] lg:gap-8 lg:py-10">
          <section class="min-w-0">
            <div class="mb-6">
              <p class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-[0.12em]" :style="{ backgroundColor: `${primary}14`, color: primary }">
                <Sparkles :size="12" /> Book a delivery
              </p>
              <h1 class="mt-3 max-w-2xl text-[clamp(1.7rem,4vw,2.55rem)] leading-[1.07] font-extrabold tracking-[-0.045em]">
                {{ headline }}
              </h1>
              <p class="mt-2 max-w-xl text-[13px] leading-6 text-slate-500">{{ description }}</p>
            </div>

            <div class="mb-4 grid grid-cols-4 gap-2" aria-label="Booking steps">
              <div v-for="(label, index) in ['Route', 'Package', 'People', 'Review']" :key="label" class="min-w-0">
                <div class="h-1 rounded-full transition" :style="{ backgroundColor: index <= currentStep ? primary : '#e5e7eb' }" />
                <p class="mt-2 flex items-center gap-1.5 truncate text-[10.5px] font-semibold" :class="index <= currentStep ? 'text-slate-800' : 'text-slate-500'">
                  <span class="flex size-4 shrink-0 items-center justify-center rounded-full text-[8px] font-bold" :style="index <= currentStep ? { backgroundColor: `${primary}16`, color: primary } : { backgroundColor: '#f1f5f9', color: '#64748b' }">
                    {{ index + 1 }}
                  </span>
                  {{ label }}
                </p>
              </div>
            </div>

            <div v-if="error" role="alert" class="mb-4 flex items-start gap-2 rounded-xl border border-rose-200 bg-rose-50 px-3.5 py-3 text-[12px] leading-5 text-rose-700">
              <CircleAlert :size="15" class="mt-0.5 shrink-0" /> {{ error }}
            </div>

            <div v-if="currentStep === 0" class="rounded-3xl border border-slate-200/80 bg-white p-4 shadow-[0_16px_50px_-40px_rgba(15,23,42,.35)] sm:p-6">
              <div class="flex items-start gap-3">
                <span class="flex size-9 shrink-0 items-center justify-center rounded-xl" :style="{ backgroundColor: `${primary}14`, color: primary }"><MapPin :size="17" /></span>
                <div>
                  <h2 class="text-[16px] font-bold tracking-tight">Where are we going?</h2>
                  <p class="mt-1 text-[11.5px] leading-5 text-slate-500">Choose a provider route or pin exact pickup and drop-off places.</p>
                </div>
              </div>

              <div v-if="routeLoadError" role="status" class="mt-4 rounded-xl bg-amber-50 px-3 py-2.5 text-[11px] text-amber-800">
                We couldn’t load the provider’s routes. You can still enter the locations manually.
              </div>
              <div v-if="routes.length" class="mt-5 grid grid-cols-2 gap-2 rounded-xl bg-slate-100 p-1">
                <button type="button" class="rounded-lg px-3 py-2.5 text-[11px] font-semibold transition" :class="routeMode === 'route' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500'" @click="routeMode = 'route'">
                  Provider routes <span class="ml-1 text-slate-400">{{ routes.length }}</span>
                </button>
                <button type="button" class="rounded-lg px-3 py-2.5 text-[11px] font-semibold transition" :class="routeMode === 'map' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500'" @click="routeMode = 'map'">
                  Enter locations
                </button>
              </div>

              <div v-if="routeMode === 'route' && routes.length" class="mt-4 space-y-2.5">
                <button
                  v-for="providerRoute in routes"
                  :key="providerRoute.id"
                  type="button"
                  class="flex w-full items-center gap-3 rounded-2xl border p-3.5 text-left transition hover:-translate-y-0.5 hover:shadow-md sm:p-4"
                  :style="form.selectedRouteId === providerRoute.id ? { borderColor: primary, backgroundColor: `${primary}08`, boxShadow: `0 0 0 1px ${primary}` } : {}"
                  @click="selectRoute(providerRoute)"
                >
                  <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-slate-100 text-slate-600"><MapPin :size="16" /></span>
                  <span class="min-w-0 flex-1">
                    <span class="block text-[12px] font-bold">{{ providerRoute.origin_city }}, {{ providerRoute.origin_state }} <span class="mx-1 text-slate-400">→</span> {{ providerRoute.destination_city }}, {{ providerRoute.destination_state }}</span>
                    <span class="mt-1 block text-[10px] text-slate-600">{{ providerRoute.distance_km }} km · {{ formatCurrency(Math.min(...[providerRoute.standard_price_per_km_minor, providerRoute.express_price_per_km_minor, providerRoute.same_day_price_per_km_minor].filter((price): price is number => price !== null)), providerRoute.currency) }}/km from</span>
                  </span>
                  <span v-if="form.selectedRouteId === providerRoute.id" class="flex size-5 shrink-0 items-center justify-center rounded-full text-white" :style="{ backgroundColor: primary }"><Check :size="12" /></span>
                  <span v-else class="size-5 shrink-0 rounded-full border border-slate-300" />
                </button>
              </div>
              <div v-else class="mt-5 space-y-4 rounded-2xl bg-slate-50 p-3.5 sm:p-4">
                <BookingLocationField v-model="form.pickup" label="Pickup location" />
                <div class="h-px bg-slate-200" />
                <BookingLocationField v-model="form.dropoff" label="Drop-off location" />
              </div>
              <div v-if="!routes.length && !routeLoadError" class="mt-4 flex items-start gap-2 rounded-xl bg-blue-50 px-3 py-2.5 text-[10.5px] leading-5 text-blue-800">
                <BadgeCheck :size="14" class="mt-0.5 shrink-0" /> Search and select a Mapbox suggestion to verify each map pin.
              </div>
              <div v-if="routeMode === 'route' && routeSelected" class="mt-4 rounded-2xl bg-slate-50 p-3.5 sm:p-4">
                <p class="text-[11px] font-semibold text-slate-800">Choose a delivery speed</p>
                <p class="mt-0.5 text-[10px] text-slate-500">Your price is calculated from this route’s distance.</p>
                <div class="mt-3 grid gap-2 sm:grid-cols-3">
                  <button
                    v-for="service in deliveryServices"
                    :key="service.key"
                    type="button"
                    class="rounded-xl border bg-white p-3 text-left transition hover:border-slate-300"
                    :style="form.deliverySpeed === service.key ? { borderColor: primary, backgroundColor: `${primary}08` } : {}"
                    @click="form.deliverySpeed = service.key"
                  >
                    <span class="block text-[10px] font-semibold text-slate-600">{{ service.label }}</span>
                    <span class="mt-1 block text-[12px] font-extrabold">{{ formatCurrency(Math.max(routeSelected.minimum_charge_minor, Math.ceil((service.price ?? 0) * routeSelected.distance_km)), routeSelected.currency) }}</span>
                    <span class="mt-0.5 block text-[9px] text-slate-500">{{ formatCurrency(service.price ?? 0, routeSelected.currency) }}/km</span>
                  </button>
                </div>
              </div>
            </div>

            <div v-else-if="currentStep === 1" class="rounded-3xl border border-slate-200/80 bg-white p-4 shadow-[0_16px_50px_-40px_rgba(15,23,42,.35)] sm:p-6">
              <div class="flex items-start gap-3">
                <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-amber-50 text-amber-700"><Package :size="17" /></span>
                <div><h2 class="text-[16px] font-bold tracking-tight">Tell us about the package</h2><p class="mt-1 text-[11.5px] text-slate-500">A little detail helps the team handle it with care.</p></div>
              </div>
              <fieldset class="mt-5">
                <legend class="mb-2.5 text-[11px] font-semibold text-slate-700">What are you sending?</legend>
                <div class="grid grid-cols-2 gap-2 sm:grid-cols-3">
                  <button
                    v-for="item in categoryOptions"
                    :key="item.value"
                    type="button"
                    class="rounded-2xl border p-3 text-left transition hover:-translate-y-0.5"
                    :style="form.category === item.value ? { borderColor: primary, backgroundColor: `${primary}08` } : {}"
                    @click="form.category = item.value"
                  >
                    <span class="text-lg">{{ item.emoji }}</span>
                    <span class="mt-1.5 block text-[11px] font-bold">{{ item.label }}</span>
                    <span class="mt-0.5 block text-[9.5px] text-slate-400">{{ item.hint }}</span>
                  </button>
                </div>
              </fieldset>
              <div class="mt-4 grid gap-3 sm:grid-cols-2">
                <label class="block">
                  <span class="mb-1.5 block text-[11px] font-semibold text-slate-700">Approximate item value</span>
                  <div class="relative">
                    <span class="absolute top-3 left-3 text-[11px] font-semibold text-slate-400">{{ currency }}</span>
                    <input v-model="form.declaredValue" type="number" min="0" step="0.01" inputmode="decimal" placeholder="0.00" class="w-full rounded-xl border border-slate-200 bg-white py-3 pr-3 pl-14 text-[13px] outline-none focus:border-[var(--booking-primary)] focus:ring-4 focus:ring-[var(--booking-primary-soft)]" />
                  </div>
                  <span class="mt-1 block text-[9.5px] text-slate-400">For handling and protection, not the delivery fee.</span>
                </label>
                <label class="block">
                  <span class="mb-1.5 inline-flex items-center gap-1 text-[11px] font-semibold text-slate-700"><Weight :size="12" /> Package weight</span>
                  <div class="relative">
                    <input v-model="form.weightKg" type="number" min="0.1" max="500" step="0.1" inputmode="decimal" placeholder="e.g. 2.5" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-3 pr-10 text-[13px] outline-none focus:border-[var(--booking-primary)] focus:ring-4 focus:ring-[var(--booking-primary-soft)]" />
                    <span class="absolute top-3 right-3 text-[11px] text-slate-400">kg</span>
                  </div>
                </label>
              </div>
              <label class="mt-4 block">
                <span class="mb-1.5 block text-[11px] font-semibold text-slate-700">Extra delivery instructions <span class="font-normal text-slate-400">· optional</span></span>
                <textarea v-model="form.instructions" maxlength="1000" rows="3" placeholder="Fragile, call before arrival, gate code…" class="w-full resize-y rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-[12px] leading-5 outline-none focus:border-[var(--booking-primary)] focus:ring-4 focus:ring-[var(--booking-primary-soft)]" />
              </label>
              <div class="mt-4">
                <p class="mb-1.5 text-[11px] font-semibold text-slate-700">Package photo <span class="font-normal text-slate-400">· optional</span></p>
                <div v-if="form.photoUrl" class="flex items-center gap-3 rounded-xl border border-emerald-200 bg-emerald-50 p-2.5">
                  <img :src="form.photoUrl" alt="Package photo preview" class="size-14 rounded-lg object-cover" />
                  <div class="min-w-0 flex-1"><p class="text-[11px] font-semibold text-emerald-800">Photo added</p><p class="text-[9.5px] text-emerald-700">Stored securely with Cloudinary</p></div>
                  <button type="button" aria-label="Remove photo" class="rounded-lg p-2 text-slate-500 hover:bg-white" @click="form.photoUrl = ''"><X :size="15" /></button>
                </div>
                <label v-else class="flex cursor-pointer items-center gap-3 rounded-xl border border-dashed border-slate-300 bg-slate-50 px-3.5 py-3 transition hover:border-[var(--booking-primary)]">
                  <span class="flex size-9 items-center justify-center rounded-xl bg-white text-slate-600 shadow-sm"><Loader2 v-if="uploading" :size="16" class="animate-spin" /><Upload v-else :size="16" /></span>
                  <span class="min-w-0 flex-1"><span class="block text-[11px] font-semibold">{{ uploading ? 'Uploading photo…' : 'Add a photo of your item' }}</span><span class="mt-0.5 block text-[9.5px] text-slate-400">{{ photoUploadEnabled ? 'JPG, PNG or WebP · up to 10 MB' : 'Optional · photo upload is not configured' }}</span></span>
                  <input type="file" accept="image/png,image/jpeg,image/webp" class="sr-only" :disabled="uploading || !photoUploadEnabled" @change="uploadPhoto" />
                </label>
              </div>
            </div>

            <div v-else-if="currentStep === 2" class="rounded-3xl border border-slate-200/80 bg-white p-4 shadow-[0_16px_50px_-40px_rgba(15,23,42,.35)] sm:p-6">
              <div class="flex items-start gap-3">
                <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-violet-50 text-violet-700"><UserRound :size="17" /></span>
                <div><h2 class="text-[16px] font-bold tracking-tight">Who’s sending and receiving?</h2><p class="mt-1 text-[11.5px] text-slate-500">We’ll use these details to coordinate your delivery.</p></div>
              </div>
              <div class="mt-5 grid gap-5 md:grid-cols-2">
                <section class="rounded-2xl bg-slate-50 p-3.5 sm:p-4">
                  <p class="mb-3 inline-flex items-center gap-1.5 text-[11px] font-bold" :style="{ color: primary }"><Send :size="13" /> Sender</p>
                  <label class="mb-3 block"><span class="mb-1 block text-[10px] font-semibold text-slate-600">Full name</span><input v-model="form.senderName" maxlength="120" autocomplete="name" placeholder="Your name" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-[12px] outline-none focus:border-[var(--booking-primary)]" /></label>
                  <label class="block"><span class="mb-1 block text-[10px] font-semibold text-slate-600">Phone number</span><input v-model="form.senderPhone" maxlength="30" type="tel" autocomplete="tel" placeholder="+234 800 000 0000" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-[12px] outline-none focus:border-[var(--booking-primary)]" /></label>
                </section>
                <section class="rounded-2xl bg-slate-50 p-3.5 sm:p-4">
                  <p class="mb-3 inline-flex items-center gap-1.5 text-[11px] font-bold text-rose-700"><MapPin :size="13" /> Receiver</p>
                  <label class="mb-3 block"><span class="mb-1 block text-[10px] font-semibold text-slate-600">Full name</span><input v-model="form.receiverName" maxlength="120" autocomplete="name" placeholder="Who should receive it?" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-[12px] outline-none focus:border-[var(--booking-primary)]" /></label>
                  <label class="block"><span class="mb-1 block text-[10px] font-semibold text-slate-600">Phone number</span><input v-model="form.receiverPhone" maxlength="30" type="tel" placeholder="+234 800 000 0000" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-[12px] outline-none focus:border-[var(--booking-primary)]" /></label>
                </section>
              </div>
            </div>

            <div v-else-if="currentStep === 3 && !bookingReference" class="rounded-3xl border border-slate-200/80 bg-white p-4 shadow-[0_16px_50px_-40px_rgba(15,23,42,.35)] sm:p-6">
              <div class="flex items-start gap-3">
                <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-emerald-50 text-emerald-700"><CheckCircle2 :size="17" /></span>
                <div><h2 class="text-[16px] font-bold tracking-tight">Everything look right?</h2><p class="mt-1 text-[11.5px] text-slate-500">Review your details before sending this booking request.</p></div>
              </div>
              <div class="mt-5 space-y-3">
                <section class="rounded-2xl bg-slate-50 p-3.5">
                  <p class="text-[9px] font-bold tracking-[0.14em] text-slate-400 uppercase">Route</p>
                  <div class="mt-2 grid gap-2 text-[11px] sm:grid-cols-2">
                    <p class="flex gap-2"><MapPin :size="13" class="mt-0.5 shrink-0 text-emerald-600" /><span><b>Pickup</b><br />{{ pickupText }}</span></p>
                    <p class="flex gap-2"><MapPin :size="13" class="mt-0.5 shrink-0 text-rose-600" /><span><b>Drop-off</b><br />{{ dropoffText }}</span></p>
                  </div>
                  <p v-if="routeMode === 'map'" class="mt-2 text-[9px] font-medium text-emerald-700">Both locations have verified map pins.</p>
                </section>
                <section class="rounded-2xl bg-slate-50 p-3.5">
                  <p class="text-[9px] font-bold tracking-[0.14em] text-slate-400 uppercase">Package</p>
                  <div class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-2 text-[11px]">
                    <span class="font-semibold">{{ categoryOptions.find((item) => item.value === form.category)?.label }}</span>
                    <span>{{ form.weightKg }} kg</span>
                    <span>Item value {{ formatMoney(declaredValueMinor) }}</span>
                  </div>
                  <p v-if="form.instructions" class="mt-2 text-[10.5px] leading-5 text-slate-500">{{ form.instructions }}</p>
                  <img v-if="form.photoUrl" :src="form.photoUrl" alt="Package" class="mt-3 h-24 w-28 rounded-xl object-cover" />
                </section>
                <section class="rounded-2xl bg-slate-50 p-3.5">
                  <p class="text-[9px] font-bold tracking-[0.14em] text-slate-400 uppercase">People</p>
                  <div class="mt-2 grid gap-2 text-[11px] sm:grid-cols-2">
                    <p><b>{{ form.senderName }}</b><br /><span class="text-slate-500">{{ form.senderPhone }} · Sender</span></p>
                    <p><b>{{ form.receiverName }}</b><br /><span class="text-slate-500">{{ form.receiverPhone }} · Receiver</span></p>
                  </div>
                </section>
              </div>
              <div class="mt-4 flex items-start gap-2 rounded-2xl border border-amber-200 bg-amber-50 p-3.5">
                <Sparkles :size="14" class="mt-0.5 shrink-0 text-amber-700" />
                <p v-if="estimatedDeliveryPrice !== null" class="text-[10.5px] leading-5 text-emerald-900">
                  <b>{{ serviceLabels[form.deliverySpeed] }} delivery: {{ formatMoney(estimatedDeliveryPrice) }} for {{ routeSelected?.distance_km }} km.</b>
                  Calculated from the provider’s published per-kilometre rate and minimum charge.
                </p>
                <p v-else class="text-[10.5px] leading-5 text-amber-900">
                  <b>Delivery fee: confirmed by {{ displayName }}.</b>
                  There isn’t a published rate for this package weight. They’ll confirm the amount with you before dispatch.
                </p>
              </div>
            </div>

            <div v-else-if="bookingReference" class="rounded-3xl border border-emerald-200 bg-white px-5 py-10 text-center shadow-sm sm:px-10 sm:py-14">
              <span class="mx-auto flex size-14 items-center justify-center rounded-2xl bg-emerald-100 text-emerald-700"><CheckCircle2 :size="27" /></span>
              <p class="mt-5 text-[10px] font-bold tracking-[0.16em] text-emerald-700 uppercase">Request received</p>
              <h2 class="mt-2 text-[clamp(1.45rem,3vw,2rem)] font-extrabold tracking-tight">You’re all set.</h2>
              <p class="mx-auto mt-2 max-w-md text-[12px] leading-6 text-slate-500">
                {{ displayName }} has your delivery request and will coordinate the pickup.
                <template v-if="submittedPriceIsEstimate">They’ll confirm the delivery fee with you.</template>
                <template v-else-if="submittedPriceMinor !== null">Your published {{ serviceLabels[form.deliverySpeed] }} price is {{ formatMoney(submittedPriceMinor) }}.</template>
              </p>
              <p class="mx-auto mt-5 inline-flex items-center gap-2 rounded-full bg-slate-100 px-4 py-2 font-mono text-[11px] font-bold text-slate-700">
                Booking reference <span>{{ bookingCode }}</span>
              </p>
              <p class="mt-5 text-[10px] text-slate-400">Keep this reference handy if you contact the provider.</p>
            </div>

            <div v-if="!bookingReference" class="mt-4 flex items-center justify-between gap-3">
              <button
                v-if="currentStep > 0"
                type="button"
                class="inline-flex min-h-11 items-center gap-1.5 rounded-xl px-3.5 text-[11px] font-semibold text-slate-500 transition hover:bg-white hover:text-slate-900"
                @click="previousStep"
              >
                <ArrowLeft :size="14" /> Back
              </button>
              <span v-else />
              <button
                v-if="currentStep < 3"
                type="button"
                class="inline-flex min-h-11 items-center gap-2 rounded-xl px-5 text-[11px] font-bold text-white shadow-lg transition hover:-translate-y-0.5 hover:shadow-xl"
                :style="{ backgroundColor: primary, boxShadow: `0 8px 20px -9px ${primary}99` }"
                @click="nextStep"
              >
                Continue <ArrowRight :size="14" />
              </button>
              <button
                v-else
                type="button"
                :disabled="saving"
                class="inline-flex min-h-11 items-center gap-2 rounded-xl px-5 text-[11px] font-bold text-white shadow-lg transition hover:-translate-y-0.5 hover:shadow-xl disabled:cursor-wait disabled:opacity-70"
                :style="{ backgroundColor: primary, boxShadow: `0 8px 20px -9px ${primary}99` }"
                @click="createBooking"
              >
                <Loader2 v-if="saving" :size="14" class="animate-spin" />
                <Send v-else :size="14" />
                {{ saving ? 'Sending request…' : 'Request this delivery' }}
              </button>
            </div>
          </section>

          <aside class="space-y-4 lg:sticky lg:top-6">
            <section class="overflow-hidden rounded-3xl bg-white shadow-[0_16px_50px_-40px_rgba(15,23,42,.4)]">
              <div class="p-4 sm:p-5">
                <p class="text-[10px] font-bold tracking-[0.14em] text-slate-400 uppercase">Your delivery at a glance</p>
                <div class="mt-4 space-y-3">
                  <div class="flex gap-2.5">
                    <span class="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full bg-emerald-50 text-emerald-700"><MapPin :size="12" /></span>
                    <div class="min-w-0"><p class="text-[9px] font-bold text-slate-400 uppercase">From</p><p class="mt-0.5 truncate text-[11px] font-semibold">{{ pickupText || 'Pickup location' }}</p></div>
                  </div>
                  <div class="ml-3 h-4 border-l border-dashed border-slate-300" />
                  <div class="flex gap-2.5">
                    <span class="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full bg-rose-50 text-rose-700"><MapPin :size="12" /></span>
                    <div class="min-w-0"><p class="text-[9px] font-bold text-slate-400 uppercase">To</p><p class="mt-0.5 truncate text-[11px] font-semibold">{{ dropoffText || 'Drop-off location' }}</p></div>
                  </div>
                </div>
                <div class="my-4 h-px bg-slate-100" />
                <div class="flex items-center justify-between gap-3">
                  <span class="text-[11px] text-slate-500">Estimated delivery fee</span>
                  <span v-if="estimatedDeliveryPrice !== null" class="text-[12px] font-extrabold">{{ formatMoney(estimatedDeliveryPrice) }}</span>
                  <span v-else class="rounded-full bg-amber-50 px-2.5 py-1 text-[9px] font-bold text-amber-700">Provider confirms</span>
                </div>
                <p v-if="estimatedDeliveryPrice !== null" class="mt-1 text-[9.5px] text-slate-500">{{ serviceLabels[form.deliverySpeed] }} · {{ routeSelected?.distance_km }} km × published rate</p>
                <p class="mt-2 text-[9.5px] leading-4 text-slate-500">The item value below is for protection and is not your delivery charge.</p>
                <div class="mt-3 flex items-center justify-between rounded-xl bg-slate-50 px-3 py-2.5">
                  <span class="text-[10px] text-slate-500">Item value</span>
                  <span class="text-[11px] font-bold">{{ formatMoney(declaredValueMinor) }}</span>
                </div>
              </div>
              <div class="flex items-center gap-2 border-t border-slate-100 bg-slate-50/80 px-4 py-3 text-[9.5px] text-slate-500 sm:px-5">
                <ShieldCheck :size="13" class="shrink-0 text-emerald-600" /> Your details go straight to {{ displayName }}.
              </div>
            </section>

            <section class="rounded-2xl border border-white bg-white/70 p-4">
              <p class="text-[10px] font-bold text-slate-800">Need a hand?</p>
              <a v-if="company.phone" :href="phoneHref" class="mt-2 inline-flex items-center gap-1.5 text-[10.5px] font-semibold" :style="{ color: primary }">
                Contact {{ displayName }} <ArrowRight :size="12" />
              </a>
              <p v-else class="mt-1.5 text-[10px] leading-5 text-slate-500">Your booking request will be sent directly to the provider.</p>
            </section>
          </aside>
        </div>
      </template>
      <footer class="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200/70 pt-5 text-[9.5px] text-slate-400">
        <span>{{ displayName }} · {{ location }}</span>
        <NuxtLink to="/" class="font-semibold text-slate-500 hover:text-slate-900">Powered by based on logistics</NuxtLink>
      </footer>
    </main>
  </div>
</template>
