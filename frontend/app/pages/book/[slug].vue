<script setup lang="ts">
import {
  ArrowLeft,
  ArrowRight,
  BadgeCheck,
  Building2,
  Camera,
  Check,
  CheckCircle2,
  CircleAlert,
  Clock3,
  Copy,
  ExternalLink,
  Facebook,
  Globe,
  Instagram,
  Linkedin,
  Loader2,
  Mail,
  MapPin,
  MessageSquare,
  Package,
  Phone,
  Send,
  ShieldCheck,
  Sparkles,
  Star,
  Twitter,
  Upload,
  UserPlus,
  UserRound,
  Weight,
  X,
} from 'lucide-vue-next'
import type { BookingLocation } from '~/components/BookingLocationField.vue'
import type { OperatingHours } from '~/types/auth'

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
    operating_hours: OperatingHours
    description: string | null
    website_url: string | null
    instagram_url: string | null
    facebook_url: string | null
    twitter_url: string | null
    linkedin_url: string | null
    year_established: number | null
    photo_urls: string[]
    verified_at: string | null
  }
  branding: {
    logo_url: string | null
    primary_color: string
    secondary_color: string
    cover_image_url: string | null
  }
  item_categories: {
    accepted: string[]
    prohibited: string[]
  }
  subdomain: string
  branches: PublicBranch[]
  reviews: PublicReview[]
  review_summary: {
    review_count: number
    average_rating: number
  }
}

interface PublicBranch {
  id: string
  name: string
  address: string
  city: string
  state: string
  phone: string | null
  email: string
  operating_hours: OperatingHours
}

interface PublicReview {
  id: string
  reviewer_name: string
  rating: number
  comment: string
  created_at: string
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

type BookingStep = 0 | 1 | 2 | 3 | 4
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
const branches = computed(() => data.value?.branches ?? [])
const reviews = ref<PublicReview[]>([])
const reviewSummary = ref({ review_count: 0, average_rating: 0 })
watch(
  data,
  (value) => {
    reviews.value = value?.reviews ?? []
    reviewSummary.value = value?.review_summary ?? { review_count: 0, average_rating: 0 }
  },
  { immediate: true },
)
const providerDescription = computed(() => company.value?.description || '')
const websiteHref = computed(() => company.value?.website_url || '')
const companyPhotos = computed(() => (company.value?.photo_urls ?? []).filter(Boolean))
const averageRating = computed(() =>
  reviewSummary.value.review_count > 0
    ? reviewSummary.value.average_rating.toFixed(1)
    : 'New',
)
const socialLinks = computed(() =>
  [
    { label: 'Instagram', href: company.value?.instagram_url ?? '', icon: Instagram },
    { label: 'Facebook', href: company.value?.facebook_url ?? '', icon: Facebook },
    { label: 'X', href: company.value?.twitter_url ?? '', icon: Twitter },
    { label: 'LinkedIn', href: company.value?.linkedin_url ?? '', icon: Linkedin },
  ].filter((link) => Boolean(link.href)),
)
const dayLabels = [
  { key: 'mon', label: 'Mon' },
  { key: 'tue', label: 'Tue' },
  { key: 'wed', label: 'Wed' },
  { key: 'thu', label: 'Thu' },
  { key: 'fri', label: 'Fri' },
  { key: 'sat', label: 'Sat' },
  { key: 'sun', label: 'Sun' },
]
const hasOperatingHours = computed(() => {
  const hours = company.value?.operating_hours
  return Boolean(hours && Object.keys(hours).length)
})
const operatingHoursRows = computed(() =>
  dayLabels
    .map((day) => {
      const entry = company.value?.operating_hours?.[day.key]
      return {
        ...day,
        closed: entry?.closed ?? !entry,
        open: entry?.open || '',
        close: entry?.close || '',
      }
    })
    .filter((row) => row.open || row.close || row.closed),
)
const serviceCoverage = computed(() => {
  const seen = new Set<string>()
  const items: string[] = []
  const add = (city: string | null | undefined, state: string | null | undefined) => {
    const label = [city, state].filter(Boolean).join(', ')
    if (!label || seen.has(label.toLowerCase())) return
    seen.add(label.toLowerCase())
    items.push(label)
  }
  for (const providerRoute of routes.value) {
    add(providerRoute.origin_city, providerRoute.origin_state)
    add(providerRoute.destination_city, providerRoute.destination_state)
  }
  for (const branch of branches.value) {
    add(branch.city, branch.state)
  }
  if (!items.length && company.value) {
    add(company.value.city, company.value.state)
  }
  return items.slice(0, 12)
})
const signedIn = ref(false)
const reviewOpen = ref(false)
const reviewSaving = ref(false)
const reviewError = ref('')
const reviewSuccess = ref(false)
const hoverRating = ref(0)
const reviewForm = reactive({ reviewerName: '', rating: 5, comment: '' })

function phoneHrefFor(value: string | null | undefined) {
  return `tel:${(value || '').replace(/[^+\d]/g, '')}`
}

function formatHour(value: string) {
  const [hourText, minuteText] = value.split(':')
  const hour = Number(hourText)
  if (!Number.isFinite(hour)) return value
  const suffix = hour >= 12 ? 'PM' : 'AM'
  const normalized = hour % 12 === 0 ? 12 : hour % 12
  return `${normalized}:${minuteText || '00'} ${suffix}`
}

function formatReviewDate(value: string) {
  return new Date(value).toLocaleDateString('en-NG', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

async function submitReview() {
  reviewError.value = ''
  reviewSuccess.value = false
  if (reviewForm.reviewerName.trim().length < 2) {
    reviewError.value = 'Add your name so the provider knows who left the review.'
    return
  }
  reviewSaving.value = true
  try {
    const response = await $fetch<{
      review: PublicReview
      review_summary: { review_count: number; average_rating: number }
    }>(`/public/book/${encodeURIComponent(slug.value)}/reviews`, {
      baseURL: apiBase,
      method: 'POST',
      credentials: 'include',
      body: {
        reviewer_name: reviewForm.reviewerName.trim(),
        rating: reviewForm.rating,
        comment: reviewForm.comment.trim(),
      },
    })
    reviews.value = [response.review, ...reviews.value].slice(0, 8)
    reviewSummary.value = response.review_summary
    reviewSuccess.value = true
    reviewForm.comment = ''
    reviewOpen.value = false
  } catch (cause) {
    reviewError.value = apiErrorMessage(cause, 'Could not submit your review. Please try again.')
  } finally {
    reviewSaving.value = false
  }
}

const currency = computed(() => company.value?.currency || 'NGN')
const routeMode = ref<'route' | 'map'>('route')
const currentStep = ref<BookingStep>(0)
const trackingCode = ref('')
const trackingUrl = ref('')
const bookingSavedToAccount = ref(false)
const registering = ref(false)
const registerError = ref('')
const copyError = ref('')
const copiedTracking = ref(false)
const accountForm = reactive({
  fullName: '',
  email: '',
  phone: '',
  password: '',
})
const { user, fetchMe } = useAuth()
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

const availableCategoryOptions = computed(() => {
  const accepted = data.value?.item_categories?.accepted ?? []
  const prohibited = data.value?.item_categories?.prohibited ?? []
  return categoryOptions.filter((option) =>
    !prohibited.includes(option.value) && (accepted.length === 0 || accepted.includes(option.value)),
  )
})
const prohibitedCategoryLabels = computed(() => {
  const prohibited = data.value?.item_categories?.prohibited ?? []
  return categoryOptions.filter((option) => prohibited.includes(option.value)).map((option) => option.label)
})

watch(availableCategoryOptions, (options) => {
  if (form.category && !options.some((option) => option.value === form.category)) {
    form.category = ''
  }
})

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
  [signedIn, () => user.value?.full_name, () => form.senderName],
  () => {
    if (!reviewForm.reviewerName) {
      reviewForm.reviewerName = user.value?.full_name || form.senderName || ''
    }
  },
  { immediate: true },
)

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
        tracking_code?: string
        tracking_url?: string
        customer_linked?: boolean
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
    trackingCode.value = response.booking.tracking_code ?? ''
    trackingUrl.value = response.booking.tracking_url
      || (trackingCode.value ? `${origin}/track/${trackingCode.value}` : '')
    bookingSavedToAccount.value = Boolean(response.booking.customer_linked)
    accountForm.fullName = form.senderName.trim()
    accountForm.phone = form.senderPhone.trim()
    currentStep.value = 4
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not send this booking request. Please try again.')
  } finally {
    saving.value = false
  }
}

async function claimBooking() {
  registerError.value = ''
  if (!bookingReference.value || registering.value) return
  if (!signedIn.value) {
    if (accountForm.fullName.trim().length < 2) {
      registerError.value = 'Add your full name to create an account.'
      return
    }
    if (!accountForm.email.includes('@')) {
      registerError.value = 'Enter the email address you want to use for your account.'
      return
    }
    if (accountForm.password.length < 8) {
      registerError.value = 'Choose a password of at least 8 characters.'
      return
    }
  }

  registering.value = true
  try {
    const response = await $fetch<{
      linked: boolean
      booking?: { tracking_code?: string; tracking_url?: string }
    }>(`/public/bookings/${encodeURIComponent(bookingReference.value)}/claim`, {
      baseURL: apiBase,
      method: 'POST',
      credentials: 'include',
      body: signedIn.value
        ? {}
        : {
            full_name: accountForm.fullName.trim(),
            email: accountForm.email.trim(),
            phone: accountForm.phone.trim(),
            password: accountForm.password,
          },
    })
    if (response.booking?.tracking_code) trackingCode.value = response.booking.tracking_code
    if (response.booking?.tracking_url) trackingUrl.value = response.booking.tracking_url
    bookingSavedToAccount.value = true
    await fetchMe()
    signedIn.value = true
  } catch (cause) {
    registerError.value = apiErrorMessage(
      cause,
      'Could not save this booking to an account. You can still use your tracking link.',
    )
  } finally {
    registering.value = false
  }
}

async function copyTrackingLink() {
  copyError.value = ''
  if (!trackingUrl.value) return
  try {
    await navigator.clipboard.writeText(trackingUrl.value)
    copiedTracking.value = true
    window.setTimeout(() => (copiedTracking.value = false), 1800)
  } catch {
    copyError.value = 'Could not copy automatically. Select the tracking link and copy it.'
  }
}

onMounted(async () => {
  signedIn.value = Boolean(await fetchMe())
})

const bookingCode = computed(() => bookingReference.value
  ? `REQ-${bookingReference.value.slice(0, 8).toUpperCase()}`
  : '')
</script>

<template>
  <div
    class="public-booking booking-shell min-h-dvh text-slate-900"
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
      <div class="pointer-events-none absolute -top-24 right-16 -z-10 size-56 rounded-full bg-white/15 blur-3xl" />
      <div class="pointer-events-none absolute -bottom-28 -left-10 -z-10 size-60 rounded-full bg-brand-300/20 blur-3xl" />
      <div class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-5 py-5 sm:px-8 sm:py-6">
        <div class="flex min-w-0 items-center gap-3">
          <div
            class="flex size-11 shrink-0 items-center justify-center overflow-hidden rounded-2xl bg-white text-lg font-extrabold shadow-lg shadow-black/20 ring-1 ring-white/30"
            :style="{ color: primary }"
          >
            <img v-if="branding?.logo_url" :src="branding.logo_url" :alt="displayName" class="h-full w-full object-cover" />
            <span v-else>{{ firstLetter }}</span>
          </div>
          <div class="min-w-0 text-white">
            <p class="flex items-center gap-1 truncate text-[14px] font-bold tracking-tight">
              {{ displayName }}
              <BadgeCheck v-if="data?.company.verified_at" :size="13" class="shrink-0 text-success-200" />
            </p>
            <p class="mt-0.5 flex items-center gap-1 truncate text-[11px] text-white/65">
              <MapPin :size="11" class="shrink-0" /> {{ location || 'Local delivery, made simple' }}
            </p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <div class="hidden items-center gap-2 rounded-full border border-white/20 bg-white/10 px-3 py-1.5 text-[10px] font-semibold text-white/90 backdrop-blur-sm sm:flex">
            <ShieldCheck :size="13" /> Secure booking
          </div>
          <ThemeToggle class="!size-9 !rounded-xl !border-white/30 !text-white hover:!bg-white/15" />
        </div>
      </div>
    </header>

    <main class="mx-auto max-w-6xl px-4 pb-14 sm:px-7">
      <div
        v-if="status === 'pending'"
        class="booking-panel mx-auto mt-8 flex max-w-3xl items-center gap-2 rounded-2xl border p-8 text-sm text-slate-500"
      >
        <Loader2 :size="17" class="animate-spin" :style="{ color: primary }" /> Loading {{ displayName || 'booking page' }}…
      </div>
      <section
        v-else-if="pageLoadError || !data || !company"
        class="booking-panel mx-auto mt-8 max-w-2xl rounded-3xl border p-8 text-center sm:p-12"
      >
        <span class="mx-auto flex size-14 items-center justify-center rounded-2xl bg-gradient-to-br from-red-100 to-accent-50 text-red-600 shadow-sm"><CircleAlert :size="24" /></span>
        <h1 class="mt-4 text-xl font-bold tracking-tight">This booking page isn’t available</h1>
        <p class="mx-auto mt-2 max-w-sm text-sm leading-6 text-slate-500">The link may have changed. Please check with the business that shared it.</p>
        <NuxtLink to="/" class="mt-6 inline-flex items-center gap-1.5 rounded-full px-4 py-2 text-sm font-semibold text-white" :style="{ backgroundColor: primary }">
          <ArrowLeft :size="14" /> Back to BOL
        </NuxtLink>
      </section>
      <template v-else>
        <div class="grid items-start gap-6 py-7 lg:grid-cols-[minmax(0,1fr)_340px] lg:gap-8 lg:py-10">
          <section class="min-w-0">
            <div class="mb-6">
              <p
                class="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[10px] font-bold tracking-[0.12em] uppercase"
                :style="{ backgroundColor: `${primary}12`, borderColor: `${primary}33`, color: primary }"
              >
                <Sparkles :size="12" /> Book a delivery
              </p>
              <h1 class="mt-3 max-w-2xl text-[clamp(1.7rem,4vw,2.55rem)] leading-[1.07] font-extrabold tracking-[-0.045em]">
                {{ headline }}
              </h1>
              <p class="mt-2 max-w-xl text-[13px] leading-6 text-slate-500">{{ description }}</p>
            </div>

            <section class="booking-panel mb-4 overflow-hidden rounded-3xl border" aria-label="About the provider">
              <div class="grid gap-5 p-4 sm:p-5 lg:grid-cols-[minmax(0,1.35fr)_minmax(0,1fr)]">
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-1.5">
                    <span class="inline-flex items-center gap-1 rounded-full border border-success-200 bg-success-50 px-2 py-0.5 text-[9.5px] font-semibold text-success-700">
                      <BadgeCheck :size="12" /> Verified provider
                    </span>
                    <span class="inline-flex items-center gap-1 rounded-full border border-accent-200 bg-accent-50 px-2 py-0.5 text-[9.5px] font-semibold text-accent-700">
                      <Star :size="11" class="fill-accent-500 text-accent-500" />
                      {{ averageRating }}
                      <span class="font-normal text-accent-700/80">
                        {{ reviewSummary.review_count
                          ? `(${reviewSummary.review_count} review${reviewSummary.review_count === 1 ? '' : 's'})`
                          : 'no reviews yet' }}
                      </span>
                    </span>
                    <span
                      v-if="company.year_established"
                      class="inline-flex items-center gap-1 rounded-full bg-slate-100 px-2 py-0.5 text-[9.5px] font-semibold text-slate-600"
                    >
                      <Building2 :size="11" /> Since {{ company.year_established }}
                    </span>
                  </div>

                  <p v-if="providerDescription" class="mt-2.5 max-w-2xl text-[11.5px] leading-5 text-slate-600">
                    {{ providerDescription }}
                  </p>

                  <div class="mt-3 flex flex-wrap items-center gap-1.5">
                    <a
                      :href="`mailto:${company.business_email}`"
                      class="inline-flex max-w-full items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-[10.5px] font-medium text-slate-600 transition hover:border-slate-300"
                    >
                      <Mail :size="12" :style="{ color: primary }" />
                      <span class="truncate">{{ company.business_email }}</span>
                    </a>
                    <a
                      :href="phoneHref"
                      class="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-[10.5px] font-medium text-slate-600 transition hover:border-slate-300"
                    >
                      <Phone :size="12" :style="{ color: primary }" /> {{ company.phone }}
                    </a>
                    <a
                      v-if="websiteHref"
                      :href="websiteHref"
                      target="_blank"
                      rel="noopener"
                      class="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-[10.5px] font-medium text-slate-600 transition hover:border-slate-300"
                    >
                      <Globe :size="12" :style="{ color: primary }" /> Website <ExternalLink :size="10" />
                    </a>
                  </div>

                  <div v-if="socialLinks.length" class="mt-2.5 flex items-center gap-1.5">
                    <span class="text-[9px] font-semibold tracking-wide text-slate-400 uppercase">Follow</span>
                    <a
                      v-for="link in socialLinks"
                      :key="link.label"
                      :href="link.href"
                      target="_blank"
                      rel="noopener"
                      :aria-label="link.label"
                      class="flex size-7 items-center justify-center rounded-lg border border-slate-200 bg-white text-slate-500 transition hover:text-slate-900"
                    >
                      <component :is="link.icon" :size="13" />
                    </a>
                  </div>
                </div>

                <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-1">
                  <div v-if="hasOperatingHours" class="booking-soft rounded-2xl border p-3">
                    <p class="flex items-center gap-1.5 text-[9.5px] font-bold tracking-wide text-slate-500 uppercase">
                      <Clock3 :size="12" :style="{ color: primary }" /> Operating hours
                    </p>
                    <dl class="mt-2 space-y-1">
                      <div v-for="row in operatingHoursRows" :key="row.key" class="flex items-center justify-between gap-2 text-[10px]">
                        <dt class="text-slate-500">{{ row.label }}</dt>
                        <dd class="font-medium text-slate-700">
                          <span v-if="row.closed" class="text-slate-400">Closed</span>
                          <span v-else>{{ formatHour(row.open) }} – {{ formatHour(row.close) }}</span>
                        </dd>
                      </div>
                    </dl>
                  </div>
                  <div v-if="serviceCoverage.length" class="booking-soft rounded-2xl border p-3">
                    <p class="flex items-center gap-1.5 text-[9.5px] font-bold tracking-wide text-slate-500 uppercase">
                      <MapPin :size="12" :style="{ color: primary }" /> Service coverage
                    </p>
                    <div class="mt-2 flex flex-wrap gap-1">
                      <span
                        v-for="place in serviceCoverage"
                        :key="place"
                        class="rounded-full border border-slate-200 bg-white px-2 py-0.5 text-[9.5px] font-medium text-slate-600"
                      >
                        {{ place }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              <div v-if="branches.length || companyPhotos.length" class="grid gap-4 border-t border-slate-200/70 p-4 sm:p-5 md:grid-cols-2">
                <div v-if="branches.length">
                  <p class="flex items-center gap-1.5 text-[9.5px] font-bold tracking-wide text-slate-500 uppercase">
                    <Building2 :size="12" :style="{ color: primary }" /> Branches
                  </p>
                  <ul class="mt-2 space-y-1.5">
                    <li
                      v-for="branch in branches.slice(0, 4)"
                      :key="branch.id"
                      class="rounded-xl border border-slate-200 bg-white px-2.5 py-2 text-[10.5px]"
                    >
                      <p class="font-semibold text-slate-700">{{ branch.name }}</p>
                      <p class="mt-0.5 text-slate-500">{{ branch.address }}, {{ branch.city }}, {{ branch.state }}</p>
                      <a
                        v-if="branch.phone"
                        :href="phoneHrefFor(branch.phone)"
                        class="mt-0.5 inline-flex items-center gap-1 font-medium"
                        :style="{ color: primary }"
                      >
                        <Phone :size="10" /> {{ branch.phone }}
                      </a>
                    </li>
                  </ul>
                  <p v-if="branches.length > 4" class="mt-1.5 text-[9.5px] text-slate-400">
                    +{{ branches.length - 4 }} more branch{{ branches.length - 4 === 1 ? '' : 'es' }}
                  </p>
                </div>
                <div v-if="companyPhotos.length">
                  <p class="flex items-center gap-1.5 text-[9.5px] font-bold tracking-wide text-slate-500 uppercase">
                    <Camera :size="12" :style="{ color: primary }" /> Company photos
                  </p>
                  <div class="mt-2 grid grid-cols-3 gap-1.5">
                    <a
                      v-for="photo in companyPhotos.slice(0, 6)"
                      :key="photo"
                      :href="photo"
                      target="_blank"
                      rel="noopener"
                      class="group relative aspect-[4/3] overflow-hidden rounded-xl border border-slate-200 bg-slate-100"
                    >
                      <img :src="photo" alt="Company photo" loading="lazy" class="size-full object-cover transition group-hover:scale-105" />
                    </a>
                  </div>
                </div>
              </div>

              <div class="border-t border-slate-200/70 p-4 sm:p-5">
                <div class="flex flex-wrap items-start justify-between gap-3">
                  <div class="min-w-0">
                    <h2 class="flex items-center gap-1.5 text-[13px] font-bold tracking-tight">
                      <MessageSquare :size="14" :style="{ color: primary }" /> Customer reviews
                    </h2>
                    <p class="mt-1 flex flex-wrap items-center gap-1.5 text-[10.5px] text-slate-500">
                      <template v-if="reviewSummary.review_count">
                        <span class="inline-flex items-center gap-0.5">
                          <Star
                            v-for="index in 5"
                            :key="index"
                            :size="11"
                            :class="index <= Math.round(reviewSummary.average_rating) ? 'fill-accent-500 text-accent-500' : 'text-slate-300'"
                          />
                        </span>
                        <span class="font-semibold text-slate-700">{{ averageRating }}</span>
                        from {{ reviewSummary.review_count }} review{{ reviewSummary.review_count === 1 ? '' : 's' }}
                      </template>
                      <template v-else>No reviews yet — be the first to share your experience.</template>
                    </p>
                  </div>
                  <button
                    type="button"
                    class="inline-flex items-center gap-1.5 rounded-xl border bg-white px-3 py-2 text-[10.5px] font-semibold transition hover:bg-slate-50"
                    :style="{ borderColor: `${primary}55`, color: primary }"
                    @click="reviewOpen = !reviewOpen"
                  >
                    <Star :size="12" /> {{ reviewOpen ? 'Close review form' : 'Leave a review' }}
                  </button>
                </div>

                <p v-if="reviewError" role="alert" class="mt-2.5 flex items-start gap-1.5 rounded-lg border border-red-200 bg-red-50 px-2.5 py-2 text-[10.5px] text-red-700">
                  <CircleAlert :size="12" class="mt-0.5 shrink-0" />{{ reviewError }}
                </p>
                <p v-if="reviewSuccess" role="status" class="mt-2.5 flex items-center gap-1.5 rounded-lg border border-success-200 bg-success-50 px-2.5 py-2 text-[10.5px] text-success-700">
                  <CheckCircle2 :size="12" /> Thanks! Your review is now live.
                </p>

                <form
                  v-if="reviewOpen"
                  class="booking-soft mt-3 grid gap-2 rounded-2xl border p-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end"
                  @submit.prevent="submitReview"
                >
                  <label class="block">
                    <span class="mb-1 block text-[10px] font-semibold text-slate-600">Your name</span>
                    <input
                      v-model="reviewForm.reviewerName"
                      required
                      minlength="2"
                      maxlength="120"
                      placeholder="e.g. Ada O."
                      class="w-full rounded-lg border border-slate-200 bg-white px-2.5 py-2 text-[11px] outline-none focus:border-[var(--booking-primary)]"
                    />
                  </label>
                  <div class="block">
                    <span class="mb-1 block text-[10px] font-semibold text-slate-600">Rating</span>
                    <div class="flex items-center gap-0.5" @mouseleave="hoverRating = 0">
                      <button
                        v-for="value in 5"
                        :key="value"
                        type="button"
                        :aria-label="`${value} star${value === 1 ? '' : 's'}`"
                        class="p-0.5"
                        @click="reviewForm.rating = value"
                        @mouseenter="hoverRating = value"
                      >
                        <Star
                          :size="20"
                          :class="value <= (hoverRating || reviewForm.rating) ? 'fill-accent-500 text-accent-500' : 'text-slate-300'"
                        />
                      </button>
                    </div>
                  </div>
                  <label class="block sm:col-span-2">
                    <span class="mb-1 block text-[10px] font-semibold text-slate-600">
                      Comment <span class="font-normal text-slate-400">· optional</span>
                    </span>
                    <textarea
                      v-model="reviewForm.comment"
                      rows="2"
                      maxlength="1000"
                      placeholder="How was the pickup, communication and delivery?"
                      class="w-full resize-y rounded-lg border border-slate-200 bg-white px-2.5 py-2 text-[11px] outline-none focus:border-[var(--booking-primary)]"
                    />
                  </label>
                  <div class="flex justify-end sm:col-span-2">
                    <button
                      type="submit"
                      :disabled="reviewSaving"
                      class="inline-flex min-h-9 items-center gap-1.5 rounded-lg px-3 text-[10.5px] font-bold text-white disabled:opacity-60"
                      :style="{ backgroundColor: primary }"
                    >
                      <Loader2 v-if="reviewSaving" :size="12" class="animate-spin" />
                      <Star v-else :size="12" />
                      {{ reviewSaving ? 'Posting…' : 'Post review' }}
                    </button>
                  </div>
                </form>

                <ul v-if="reviews.length" class="mt-3 grid gap-2 sm:grid-cols-2">
                  <li v-for="review in reviews.slice(0, 4)" :key="review.id" class="rounded-xl border border-slate-200 bg-white p-2.5">
                    <div class="flex items-center justify-between gap-2">
                      <p class="truncate text-[10.5px] font-semibold text-slate-700">{{ review.reviewer_name }}</p>
                      <span class="inline-flex shrink-0 items-center gap-0.5">
                        <Star
                          v-for="index in 5"
                          :key="index"
                          :size="9"
                          :class="index <= review.rating ? 'fill-accent-500 text-accent-500' : 'text-slate-200'"
                        />
                      </span>
                    </div>
                    <p v-if="review.comment" class="mt-1 text-[10px] leading-4 text-slate-500">{{ review.comment }}</p>
                    <p class="mt-1 text-[9px] text-slate-400">{{ formatReviewDate(review.created_at) }}</p>
                  </li>
                </ul>
              </div>
            </section>

            <div class="mb-4 grid grid-cols-5 gap-1 sm:gap-2" aria-label="Booking steps">
              <div
                v-for="(label, index) in ['Route', 'Package', 'People', 'Review', 'Track']"
                :key="label"
                class="booking-step min-w-0 rounded-xl border px-2 py-2 sm:px-2.5"
                :class="index > currentStep ? 'border-slate-200 bg-white/70 dark:border-white/10 dark:bg-white/5' : ''"
                :style="index <= currentStep ? { borderColor: `${primary}45`, backgroundColor: `${primary}0d` } : undefined"
              >
                <div class="flex items-center gap-1.5">
                  <span
                    class="flex size-5 shrink-0 items-center justify-center rounded-lg text-[9px] font-bold"
                    :style="index < currentStep
                      ? { backgroundColor: primary, color: '#fff' }
                      : index === currentStep
                        ? { backgroundColor: `${primary}1f`, color: primary }
                        : { backgroundColor: '#f1f5f9', color: '#64748b' }"
                  >
                    <Check v-if="index < currentStep" :size="11" />
                    <template v-else>{{ index + 1 }}</template>
                  </span>
                  <span class="truncate text-[10px] font-semibold" :class="index <= currentStep ? 'text-slate-800 dark:text-slate-100' : 'text-slate-500'">
                    {{ label }}
                  </span>
                </div>
              </div>
            </div>

            <div v-if="error" role="alert" class="mb-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-3 text-[12px] leading-5 text-red-700">
              <CircleAlert :size="15" class="mt-0.5 shrink-0" /> {{ error }}
            </div>

            <div v-if="currentStep === 0" class="booking-panel rounded-3xl border p-4 sm:p-6">
              <div class="flex items-start justify-between gap-3">
                <div class="flex items-start gap-3">
                  <span class="booking-icon-tile flex size-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-brand-100 to-brand-50 text-brand-700 dark:from-brand-400/15 dark:to-brand-400/10 dark:text-brand-200"><MapPin :size="18" /></span>
                  <div>
                    <h2 class="text-[16px] font-bold tracking-tight">Where are we going?</h2>
                    <p class="mt-1 text-[11.5px] leading-5 text-slate-500">Choose a provider route or pin exact pickup and drop-off places.</p>
                  </div>
                </div>
                <span class="hidden shrink-0 rounded-full bg-slate-100 px-2 py-1 text-[9px] font-semibold tracking-wide text-slate-500 uppercase sm:inline dark:bg-white/10 dark:text-slate-300">Step 1 of 5</span>
              </div>

              <div v-if="routeLoadError" role="status" class="mt-4 rounded-xl bg-accent-50 px-3 py-2.5 text-[11px] text-accent-800">
                We couldn’t load the provider’s routes. You can still enter the locations manually.
              </div>
              <div v-if="routes.length" class="booking-soft mt-5 grid grid-cols-2 gap-1 rounded-xl border p-1">
                <button
                  type="button"
                  class="rounded-lg px-3 py-2.5 text-[11px] font-semibold transition"
                  :class="routeMode === 'route' ? 'bg-white text-slate-900 shadow-sm dark:bg-white/10 dark:text-white' : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'"
                  @click="routeMode = 'route'"
                >
                  Provider routes <span class="ml-1 text-slate-400">{{ routes.length }}</span>
                </button>
                <button
                  type="button"
                  class="rounded-lg px-3 py-2.5 text-[11px] font-semibold transition"
                  :class="routeMode === 'map' ? 'bg-white text-slate-900 shadow-sm dark:bg-white/10 dark:text-white' : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'"
                  @click="routeMode = 'map'"
                >
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
                  <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-brand-100 to-brand-50 text-brand-700 dark:from-brand-400/15 dark:to-brand-400/10 dark:text-brand-200"><MapPin :size="16" /></span>
                  <span class="min-w-0 flex-1">
                    <span class="block text-[12px] font-bold">{{ providerRoute.origin_city }}, {{ providerRoute.origin_state }} <span class="mx-1 text-slate-400">→</span> {{ providerRoute.destination_city }}, {{ providerRoute.destination_state }}</span>
                    <span class="mt-1 block text-[10px] text-slate-600">{{ providerRoute.distance_km }} km · {{ formatCurrency(Math.min(...[providerRoute.standard_price_per_km_minor, providerRoute.express_price_per_km_minor, providerRoute.same_day_price_per_km_minor].filter((price): price is number => price !== null)), providerRoute.currency) }}/km from</span>
                  </span>
                  <span v-if="form.selectedRouteId === providerRoute.id" class="flex size-5 shrink-0 items-center justify-center rounded-full text-white" :style="{ backgroundColor: primary }"><Check :size="12" /></span>
                  <span v-else class="size-5 shrink-0 rounded-full border border-slate-300" />
                </button>
              </div>
              <div v-else class="booking-soft mt-5 space-y-4 rounded-2xl border p-3.5 sm:p-4">
                <BookingLocationField v-model="form.pickup" label="Pickup location" />
                <div class="h-px bg-slate-200 dark:bg-white/10" />
                <BookingLocationField v-model="form.dropoff" label="Drop-off location" />
              </div>
              <div v-if="!routes.length && !routeLoadError" class="mt-4 flex items-start gap-2 rounded-xl bg-brand-50 px-3 py-2.5 text-[10.5px] leading-5 text-brand-700">
                <BadgeCheck :size="14" class="mt-0.5 shrink-0" /> Search and select a Mapbox suggestion to verify each map pin.
              </div>
              <div v-if="routeMode === 'route' && routeSelected" class="booking-soft mt-4 rounded-2xl border p-3.5 sm:p-4">
                <p class="text-[11px] font-semibold text-slate-800 dark:text-slate-100">Choose a delivery speed</p>
                <p class="mt-0.5 text-[10px] text-slate-500">Your price is calculated from this route’s distance.</p>
                <div class="mt-3 grid gap-2 sm:grid-cols-3">
                  <button
                    v-for="service in deliveryServices"
                    :key="service.key"
                    type="button"
                    class="booking-price-card rounded-xl border p-3 text-left"
                    :class="`booking-price-card--${service.key === 'same_day' ? 'same-day' : service.key}`"
                    :style="form.deliverySpeed === service.key ? { borderColor: primary, boxShadow: `0 0 0 1px ${primary}40` } : {}"
                    @click="form.deliverySpeed = service.key"
                  >
                    <span class="flex items-center justify-between gap-1 text-[10px] font-semibold text-slate-600">
                      {{ service.label }}
                      <Check v-if="form.deliverySpeed === service.key" :size="12" :style="{ color: primary }" />
                    </span>
                    <span class="mt-1 block text-[12px] font-extrabold">{{ formatCurrency(Math.max(routeSelected.minimum_charge_minor, Math.ceil((service.price ?? 0) * routeSelected.distance_km)), routeSelected.currency) }}</span>
                    <span class="mt-0.5 block text-[9px] text-slate-500">{{ formatCurrency(service.price ?? 0, routeSelected.currency) }}/km</span>
                  </button>
                </div>
              </div>
            </div>

            <div v-else-if="currentStep === 1" class="booking-panel rounded-3xl border p-4 sm:p-6">
              <div class="flex items-start justify-between gap-3">
                <div class="flex items-start gap-3">
                  <span class="booking-icon-tile flex size-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-accent-100 to-accent-50 text-accent-700 dark:from-accent-400/15 dark:to-accent-400/10 dark:text-accent-200"><Package :size="18" /></span>
                  <div><h2 class="text-[16px] font-bold tracking-tight">Tell us about the package</h2><p class="mt-1 text-[11.5px] text-slate-500">A little detail helps the team handle it with care.</p></div>
                </div>
                <span class="hidden shrink-0 rounded-full bg-slate-100 px-2 py-1 text-[9px] font-semibold tracking-wide text-slate-500 uppercase sm:inline dark:bg-white/10 dark:text-slate-300">Step 2 of 5</span>
              </div>
              <fieldset class="mt-5">
                <legend class="mb-2.5 text-[11px] font-semibold text-slate-700">What are you sending?</legend>
                <p v-if="prohibitedCategoryLabels.length" class="mb-2.5 rounded-lg bg-red-50 px-2.5 py-2 text-[10px] leading-4 text-red-700">
                  This provider cannot move: {{ prohibitedCategoryLabels.join(', ') }}.
                </p>
                <p v-if="!availableCategoryOptions.length" class="mb-2.5 rounded-lg bg-accent-50 px-2.5 py-2 text-[10px] leading-4 text-accent-800">
                  This provider has not published an acceptable item category. Contact them before booking.
                </p>
                <div class="grid grid-cols-2 gap-2 sm:grid-cols-3">
                  <button
                    v-for="item in availableCategoryOptions"
                    :key="item.value"
                    type="button"
                    class="rounded-2xl border border-slate-200/80 bg-white p-3 text-left transition hover:-translate-y-0.5 hover:shadow-md dark:border-white/10 dark:bg-white/5"
                    :style="form.category === item.value ? { borderColor: primary, backgroundColor: `${primary}10`, boxShadow: `0 0 0 1px ${primary}40` } : {}"
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
                <div v-if="form.photoUrl" class="flex items-center gap-3 rounded-xl border border-success-200 bg-success-50 p-2.5">
                  <img :src="form.photoUrl" alt="Package photo preview" class="size-14 rounded-lg object-cover" />
                  <div class="min-w-0 flex-1"><p class="text-[11px] font-semibold text-success-700">Photo added</p><p class="text-[9.5px] text-success-700">Stored securely with Cloudinary</p></div>
                  <button type="button" aria-label="Remove photo" class="rounded-lg p-2 text-slate-500 hover:bg-white" @click="form.photoUrl = ''"><X :size="15" /></button>
                </div>
                <label v-else class="booking-soft flex cursor-pointer items-center gap-3 rounded-xl border border-dashed border-slate-300 px-3.5 py-3 transition hover:border-[var(--booking-primary)]">
                  <span class="flex size-9 items-center justify-center rounded-xl bg-white text-slate-600 shadow-sm"><Loader2 v-if="uploading" :size="16" class="animate-spin" /><Upload v-else :size="16" /></span>
                  <span class="min-w-0 flex-1"><span class="block text-[11px] font-semibold">{{ uploading ? 'Uploading photo…' : 'Add a photo of your item' }}</span><span class="mt-0.5 block text-[9.5px] text-slate-400">{{ photoUploadEnabled ? 'JPG, PNG or WebP · up to 10 MB' : 'Optional · photo upload is not configured' }}</span></span>
                  <input type="file" accept="image/png,image/jpeg,image/webp" class="sr-only" :disabled="uploading || !photoUploadEnabled" @change="uploadPhoto" />
                </label>
              </div>
            </div>

            <div v-else-if="currentStep === 2" class="booking-panel rounded-3xl border p-4 sm:p-6">
              <div class="flex items-start justify-between gap-3">
                <div class="flex items-start gap-3">
                  <span class="booking-icon-tile flex size-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-brand-100 to-accent-50 text-brand-700 dark:from-brand-400/15 dark:to-accent-400/10 dark:text-brand-200"><UserRound :size="18" /></span>
                  <div><h2 class="text-[16px] font-bold tracking-tight">Who’s sending and receiving?</h2><p class="mt-1 text-[11.5px] text-slate-500">We’ll use these details to coordinate your delivery.</p></div>
                </div>
                <span class="hidden shrink-0 rounded-full bg-slate-100 px-2 py-1 text-[9px] font-semibold tracking-wide text-slate-500 uppercase sm:inline dark:bg-white/10 dark:text-slate-300">Step 3 of 5</span>
              </div>
              <div class="mt-5 grid gap-5 md:grid-cols-2">
                <section class="rounded-2xl border p-3.5 sm:p-4" :style="{ background: `${primary}0d`, borderColor: `${primary}26` }">
                  <p class="mb-3 inline-flex items-center gap-1.5 text-[11px] font-bold" :style="{ color: primary }"><Send :size="13" /> Sender</p>
                  <label class="mb-3 block"><span class="mb-1 block text-[10px] font-semibold text-slate-600">Full name</span><input v-model="form.senderName" maxlength="120" autocomplete="name" placeholder="Your name" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-[12px] outline-none focus:border-[var(--booking-primary)]" /></label>
                  <label class="block"><span class="mb-1 block text-[10px] font-semibold text-slate-600">Phone number</span><input v-model="form.senderPhone" maxlength="30" type="tel" autocomplete="tel" placeholder="+234 800 000 0000" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-[12px] outline-none focus:border-[var(--booking-primary)]" /></label>
                </section>
                <section class="rounded-2xl border border-red-200/70 bg-gradient-to-br from-red-50 to-accent-50/50 p-3.5 sm:p-4 dark:border-red-400/20 dark:from-red-400/10 dark:to-accent-400/5">
                  <p class="mb-3 inline-flex items-center gap-1.5 text-[11px] font-bold text-red-700 dark:text-red-200"><MapPin :size="13" /> Receiver</p>
                  <label class="mb-3 block"><span class="mb-1 block text-[10px] font-semibold text-slate-600">Full name</span><input v-model="form.receiverName" maxlength="120" autocomplete="name" placeholder="Who should receive it?" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-[12px] outline-none focus:border-[var(--booking-primary)]" /></label>
                  <label class="block"><span class="mb-1 block text-[10px] font-semibold text-slate-600">Phone number</span><input v-model="form.receiverPhone" maxlength="30" type="tel" placeholder="+234 800 000 0000" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-[12px] outline-none focus:border-[var(--booking-primary)]" /></label>
                </section>
              </div>
            </div>

            <div v-else-if="currentStep === 3 && !bookingReference" class="booking-panel rounded-3xl border p-4 sm:p-6">
              <div class="flex items-start justify-between gap-3">
                <div class="flex items-start gap-3">
                  <span class="booking-icon-tile flex size-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-success-100 to-success-50 text-success-700 dark:from-success-500/15 dark:to-success-500/10 dark:text-success-200"><CheckCircle2 :size="18" /></span>
                  <div><h2 class="text-[16px] font-bold tracking-tight">Everything look right?</h2><p class="mt-1 text-[11.5px] text-slate-500">Review your details before sending this booking request.</p></div>
                </div>
                <span class="hidden shrink-0 rounded-full bg-slate-100 px-2 py-1 text-[9px] font-semibold tracking-wide text-slate-500 uppercase sm:inline dark:bg-white/10 dark:text-slate-300">Step 4 of 5</span>
              </div>
              <div class="mt-5 space-y-3">
                <section class="rounded-2xl border border-brand-200/60 bg-gradient-to-br from-brand-50 to-brand-50/60 p-3.5 dark:border-brand-400/20 dark:from-brand-400/10 dark:to-brand-400/5">
                  <p class="text-[9px] font-bold tracking-[0.14em] text-brand-700/70 uppercase dark:text-brand-200/70">Route</p>
                  <div class="mt-2 grid gap-2 text-[11px] sm:grid-cols-2">
                    <p class="flex gap-2"><MapPin :size="13" class="mt-0.5 shrink-0 text-success-600" /><span><b>Pickup</b><br />{{ pickupText }}</span></p>
                    <p class="flex gap-2"><MapPin :size="13" class="mt-0.5 shrink-0 text-red-600" /><span><b>Drop-off</b><br />{{ dropoffText }}</span></p>
                  </div>
                  <p v-if="routeMode === 'map'" class="mt-2 text-[9px] font-medium text-success-700">Both locations have verified map pins.</p>
                </section>
                <section class="rounded-2xl border border-accent-200/60 bg-gradient-to-br from-accent-50 to-accent-50/60 p-3.5 dark:border-accent-400/20 dark:from-accent-400/10 dark:to-accent-400/5">
                  <p class="text-[9px] font-bold tracking-[0.14em] text-accent-700/70 uppercase dark:text-accent-200/70">Package</p>
                  <div class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-2 text-[11px]">
                    <span class="font-semibold">{{ availableCategoryOptions.find((item) => item.value === form.category)?.label }}</span>
                    <span>{{ form.weightKg }} kg</span>
                    <span>Item value {{ formatMoney(declaredValueMinor) }}</span>
                  </div>
                  <p v-if="form.instructions" class="mt-2 text-[10.5px] leading-5 text-slate-500">{{ form.instructions }}</p>
                  <img v-if="form.photoUrl" :src="form.photoUrl" alt="Package" class="mt-3 h-24 w-28 rounded-xl object-cover" />
                </section>
                <section class="rounded-2xl border border-brand-200/60 bg-gradient-to-br from-brand-50 to-accent-50/50 p-3.5 dark:border-brand-400/20 dark:from-brand-400/10 dark:to-accent-400/5">
                  <p class="text-[9px] font-bold tracking-[0.14em] text-brand-700/70 uppercase dark:text-brand-200/70">People</p>
                  <div class="mt-2 grid gap-2 text-[11px] sm:grid-cols-2">
                    <p><b>{{ form.senderName }}</b><br /><span class="text-slate-500">{{ form.senderPhone }} · Sender</span></p>
                    <p><b>{{ form.receiverName }}</b><br /><span class="text-slate-500">{{ form.receiverPhone }} · Receiver</span></p>
                  </div>
                </section>
              </div>
              <div class="mt-4 flex items-start gap-2 rounded-2xl border border-accent-200/70 bg-gradient-to-br from-accent-50 via-white to-success-50/60 p-3.5 dark:border-accent-400/20 dark:from-accent-400/10 dark:via-transparent dark:to-success-500/5">
                <span class="flex size-6 shrink-0 items-center justify-center rounded-lg bg-white text-accent-700 shadow-sm dark:bg-white/10 dark:text-accent-200">
                  <Sparkles :size="13" />
                </span>
                <p v-if="estimatedDeliveryPrice !== null" class="text-[10.5px] leading-5 text-success-700 dark:text-success-200">
                  <b>{{ serviceLabels[form.deliverySpeed] }} delivery: {{ formatMoney(estimatedDeliveryPrice) }} for {{ routeSelected?.distance_km }} km.</b>
                  Calculated from the provider’s published per-kilometre rate and minimum charge.
                </p>
                <p v-else class="text-[10.5px] leading-5 text-accent-800 dark:text-accent-200">
                  <b>Delivery fee: confirmed by {{ displayName }}.</b>
                  There isn’t a published rate for this package weight. They’ll confirm the amount with you before dispatch.
                </p>
              </div>
            </div>

            <div
              v-else-if="currentStep === 4 && bookingReference"
              class="relative overflow-hidden rounded-3xl border border-success-200/70 bg-gradient-to-br from-success-50 via-white to-brand-50/60 px-4 py-8 sm:px-8 sm:py-10 dark:border-success-500/20 dark:from-success-500/10 dark:via-transparent dark:to-brand-400/5"
            >
              <div class="pointer-events-none absolute -top-16 right-0 size-44 rounded-full bg-success-200/25 blur-3xl" />
              <div class="pointer-events-none absolute -bottom-20 -left-8 size-48 rounded-full bg-brand-300/25 blur-3xl" />
              <div class="relative text-center">
                <span class="booking-icon-tile mx-auto flex size-14 items-center justify-center rounded-2xl bg-gradient-to-br from-success-500 to-success-500 text-white shadow-lg"><CheckCircle2 :size="27" /></span>
                <p class="mt-5 text-[10px] font-bold tracking-[0.16em] text-success-700 uppercase dark:text-success-200">Request received</p>
                <h2 class="mt-2 text-[clamp(1.45rem,3vw,2rem)] font-extrabold tracking-tight">You’re all set.</h2>
                <p class="mx-auto mt-2 max-w-md text-[12px] leading-6 text-slate-500">
                  {{ displayName }} has your delivery request and will coordinate the pickup.
                  <template v-if="submittedPriceIsEstimate">They’ll confirm the delivery fee with you.</template>
                  <template v-else-if="submittedPriceMinor !== null">Your published {{ serviceLabels[form.deliverySpeed] }} price is {{ formatMoney(submittedPriceMinor) }}.</template>
                </p>
                <p class="mx-auto mt-4 inline-flex flex-wrap items-center justify-center gap-2 rounded-full border border-success-200/70 bg-white/80 px-4 py-2 font-mono text-[10.5px] font-bold text-slate-700 shadow-sm dark:border-success-500/20 dark:bg-white/10 dark:text-slate-100">
                  Booking reference <span class="text-success-700 dark:text-success-200">{{ bookingCode }}</span>
                </p>
              </div>

              <div class="relative mt-6 grid gap-3 text-left md:grid-cols-2">
                <section
                  v-if="!bookingSavedToAccount"
                  class="rounded-2xl border border-white/70 bg-white/85 p-4 shadow-sm dark:border-white/10 dark:bg-white/5"
                >
                  <p class="inline-flex items-center gap-1.5 text-[10px] font-bold tracking-[0.14em] text-success-700 uppercase dark:text-success-200">
                    <UserPlus :size="13" /> Optional
                  </p>
                  <h3 class="mt-1.5 text-[13.5px] font-bold">Save this booking</h3>
                  <p class="mt-1 text-[10.5px] leading-5 text-slate-500">
                    Create a free account to follow this parcel in your dashboard and keep bookings from every provider in one place.
                  </p>

                  <p
                    v-if="registerError"
                    role="alert"
                    class="mt-3 flex items-start gap-1.5 rounded-lg border border-red-200 bg-red-50 px-2.5 py-2 text-[10.5px] leading-4 text-red-700"
                  >
                    <CircleAlert :size="12" class="mt-0.5 shrink-0" />{{ registerError }}
                  </p>

                  <button
                    v-if="signedIn"
                    type="button"
                    :disabled="registering"
                    class="mt-3 inline-flex min-h-9 w-full items-center justify-center gap-1.5 rounded-lg px-3 text-[11px] font-bold text-white disabled:opacity-60"
                    :style="{ backgroundColor: primary }"
                    @click="claimBooking"
                  >
                    <Loader2 v-if="registering" :size="13" class="animate-spin" />
                    <UserPlus v-else :size="13" />
                    {{ registering ? 'Saving…' : 'Save to my account' }}
                  </button>
                  <form v-else class="mt-3 space-y-2" @submit.prevent="claimBooking">
                    <input
                      v-model="accountForm.fullName"
                      required
                      maxlength="120"
                      autocomplete="name"
                      placeholder="Full name"
                      class="w-full rounded-lg border border-slate-200 bg-white px-2.5 py-2 text-[11px] outline-none focus:border-[var(--booking-primary)]"
                    />
                    <input
                      v-model="accountForm.email"
                      required
                      type="email"
                      maxlength="254"
                      autocomplete="email"
                      placeholder="Email address"
                      class="w-full rounded-lg border border-slate-200 bg-white px-2.5 py-2 text-[11px] outline-none focus:border-[var(--booking-primary)]"
                    />
                    <input
                      v-model="accountForm.phone"
                      type="tel"
                      maxlength="30"
                      autocomplete="tel"
                      placeholder="Phone number (optional)"
                      class="w-full rounded-lg border border-slate-200 bg-white px-2.5 py-2 text-[11px] outline-none focus:border-[var(--booking-primary)]"
                    />
                    <input
                      v-model="accountForm.password"
                      required
                      type="password"
                      minlength="8"
                      maxlength="128"
                      autocomplete="new-password"
                      placeholder="Password (8+ characters)"
                      class="w-full rounded-lg border border-slate-200 bg-white px-2.5 py-2 text-[11px] outline-none focus:border-[var(--booking-primary)]"
                    />
                    <button
                      type="submit"
                      :disabled="registering"
                      class="inline-flex min-h-9 w-full items-center justify-center gap-1.5 rounded-lg px-3 text-[11px] font-bold text-white disabled:opacity-60"
                      :style="{ backgroundColor: primary }"
                    >
                      <Loader2 v-if="registering" :size="13" class="animate-spin" />
                      <UserPlus v-else :size="13" />
                      {{ registering ? 'Creating account…' : 'Create account & save booking' }}
                    </button>
                  </form>
                  <p class="mt-2 text-[9.5px] text-slate-400">You can skip this and still track the parcel with the link beside it.</p>
                </section>

                <section
                  v-else
                  class="rounded-2xl border border-success-200/70 bg-white/85 p-4 shadow-sm dark:border-success-500/20 dark:bg-white/5"
                >
                  <p class="inline-flex items-center gap-1.5 text-[10px] font-bold tracking-[0.14em] text-success-700 uppercase dark:text-success-200">
                    <CheckCircle2 :size="13" /> Saved
                  </p>
                  <h3 class="mt-1.5 text-[13.5px] font-bold">This booking is in your dashboard</h3>
                  <p class="mt-1 text-[10.5px] leading-5 text-slate-500">
                    Follow it any time from your account, along with every other provider you book with.
                  </p>
                  <NuxtLink
                    to="/dashboard"
                    class="mt-3 inline-flex min-h-9 w-full items-center justify-center gap-1.5 rounded-lg px-3 text-[11px] font-bold text-white"
                    :style="{ backgroundColor: primary }"
                  >
                    Open my dashboard <ArrowRight :size="13" />
                  </NuxtLink>
                </section>

                <section class="rounded-2xl border border-brand-200/70 bg-white/85 p-4 shadow-sm dark:border-brand-400/20 dark:bg-white/5">
                  <p class="inline-flex items-center gap-1.5 text-[10px] font-bold tracking-[0.14em] text-brand-700 uppercase dark:text-brand-200">
                    <MapPin :size="13" /> Track without an account
                  </p>
                  <h3 class="mt-1.5 text-[13.5px] font-bold">Your tracking link</h3>
                  <p class="mt-1 text-[10.5px] leading-5 text-slate-500">
                    Copy this link or keep the code handy. It updates as {{ displayName }} moves the parcel.
                  </p>
                  <div v-if="trackingUrl" class="mt-3">
                    <p class="truncate rounded-lg border border-slate-200 bg-slate-50 px-2.5 py-2 font-mono text-[10px] text-slate-600">
                      {{ trackingUrl }}
                    </p>
                    <div class="mt-2 flex gap-2">
                      <button
                        type="button"
                        class="inline-flex min-h-9 flex-1 items-center justify-center gap-1.5 rounded-lg border border-slate-200 bg-white px-2.5 text-[10.5px] font-semibold text-slate-700 transition hover:bg-slate-50"
                        @click="copyTrackingLink"
                      >
                        <Check v-if="copiedTracking" :size="13" class="text-success-600" />
                        <Copy v-else :size="13" />
                        {{ copiedTracking ? 'Copied' : 'Copy link' }}
                      </button>
                      <NuxtLink
                        :to="`/track/${trackingCode}`"
                        class="inline-flex min-h-9 items-center justify-center gap-1.5 rounded-lg px-2.5 text-[10.5px] font-semibold text-white"
                        :style="{ backgroundColor: primary }"
                      >
                        Open tracker <ArrowRight :size="12" />
                      </NuxtLink>
                    </div>
                    <p v-if="copyError" class="mt-1.5 text-[9.5px] text-red-600">{{ copyError }}</p>
                  </div>
                </section>
              </div>
              <p class="relative mt-4 text-center text-[10px] text-slate-400">
                Keep the reference or tracking link handy if you contact the provider.
              </p>
            </div>

            <div v-if="currentStep < 4" class="mt-4 flex items-center justify-between gap-3">
              <button
                v-if="currentStep > 0"
                type="button"
                class="inline-flex min-h-11 items-center gap-1.5 rounded-xl px-3.5 text-[11px] font-semibold text-slate-500 transition hover:bg-white hover:text-slate-900 dark:hover:bg-white/10 dark:hover:text-white"
                @click="previousStep"
              >
                <ArrowLeft :size="14" /> Back
              </button>
              <span v-else />
              <button
                v-if="currentStep < 3"
                type="button"
                class="inline-flex min-h-11 items-center gap-2 rounded-xl px-5 text-[11px] font-bold text-white shadow-lg transition hover:-translate-y-0.5 hover:shadow-xl"
                :style="{ backgroundImage: `linear-gradient(100deg, ${primary}, ${secondary})`, boxShadow: `0 12px 26px -14px ${primary}cc` }"
                @click="nextStep"
              >
                Continue <ArrowRight :size="14" />
              </button>
              <button
                v-else
                type="button"
                :disabled="saving"
                class="inline-flex min-h-11 items-center gap-2 rounded-xl px-5 text-[11px] font-bold text-white shadow-lg transition hover:-translate-y-0.5 hover:shadow-xl disabled:cursor-wait disabled:opacity-70"
                :style="{ backgroundImage: `linear-gradient(100deg, ${primary}, ${secondary})`, boxShadow: `0 12px 26px -14px ${primary}cc` }"
                @click="createBooking"
              >
                <Loader2 v-if="saving" :size="14" class="animate-spin" />
                <Send v-else :size="14" />
                {{ saving ? 'Sending request…' : 'Request this delivery' }}
              </button>
            </div>
          </section>

          <aside class="space-y-4 lg:sticky lg:top-6">
            <section class="booking-panel overflow-hidden rounded-3xl border">
              <div
                class="px-4 py-3.5 sm:px-5"
                :style="{ backgroundImage: `linear-gradient(110deg, ${primary}, ${secondary})` }"
              >
                <p class="flex items-center gap-1.5 text-[10px] font-bold tracking-[0.14em] text-white/85 uppercase">
                  <Sparkles :size="12" /> Your delivery at a glance
                </p>
              </div>
              <div class="p-4 sm:p-5">
                <div class="space-y-3">
                  <div class="flex gap-2.5">
                    <span class="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-200"><MapPin :size="12" /></span>
                    <div class="min-w-0"><p class="text-[9px] font-bold text-slate-400 uppercase">From</p><p class="mt-0.5 truncate text-[11px] font-semibold">{{ pickupText || 'Pickup location' }}</p></div>
                  </div>
                  <div class="ml-3 h-4 border-l border-dashed border-slate-300 dark:border-white/20" />
                  <div class="flex gap-2.5">
                    <span class="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full bg-red-50 text-red-700 dark:bg-red-400/15 dark:text-red-200"><MapPin :size="12" /></span>
                    <div class="min-w-0"><p class="text-[9px] font-bold text-slate-400 uppercase">To</p><p class="mt-0.5 truncate text-[11px] font-semibold">{{ dropoffText || 'Drop-off location' }}</p></div>
                  </div>
                </div>
                <div class="my-4 h-px bg-slate-100 dark:bg-white/10" />
                <div class="flex items-center justify-between gap-3">
                  <span class="text-[11px] text-slate-500">Estimated delivery fee</span>
                  <span v-if="estimatedDeliveryPrice !== null" class="text-[15px] font-extrabold tracking-tight">{{ formatMoney(estimatedDeliveryPrice) }}</span>
                  <span v-else class="rounded-full bg-accent-50 px-2.5 py-1 text-[9px] font-bold text-accent-700 dark:bg-accent-400/10 dark:text-accent-200">Provider confirms</span>
                </div>
                <p v-if="estimatedDeliveryPrice !== null" class="mt-1 text-[9.5px] text-slate-500">{{ serviceLabels[form.deliverySpeed] }} · {{ routeSelected?.distance_km }} km × published rate</p>
                <p class="mt-2 text-[9.5px] leading-4 text-slate-500">The item value below is for protection and is not your delivery charge.</p>
                <div class="booking-soft mt-3 flex items-center justify-between rounded-xl border px-3 py-2.5">
                  <span class="text-[10px] text-slate-500">Item value</span>
                  <span class="text-[11px] font-bold">{{ formatMoney(declaredValueMinor) }}</span>
                </div>
              </div>
              <div class="booking-soft flex items-center gap-2 border-t px-4 py-3 text-[9.5px] text-slate-500 sm:px-5">
                <ShieldCheck :size="13" class="shrink-0 text-success-600" /> Your details go straight to {{ displayName }}.
              </div>
            </section>

            <section class="booking-soft rounded-2xl border p-4">
              <p class="flex items-center gap-1.5 text-[10px] font-bold text-slate-800 dark:text-slate-100">
                <UserRound :size="12" :style="{ color: primary }" /> Need a hand?
              </p>
              <a v-if="company.phone" :href="phoneHref" class="mt-2 inline-flex items-center gap-1.5 text-[10.5px] font-semibold" :style="{ color: primary }">
                Contact {{ displayName }} <ArrowRight :size="12" />
              </a>
              <p v-else class="mt-1.5 text-[10px] leading-5 text-slate-500">Your booking request will be sent directly to the provider.</p>
            </section>
          </aside>
        </div>
      </template>
      <footer class="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200/70 pt-5 text-[9.5px] text-slate-400 dark:border-white/10">
        <span class="inline-flex items-center gap-1.5"><ShieldCheck :size="12" class="text-success-500" /> {{ displayName }} · {{ location }}</span>
        <NuxtLink to="/" class="font-semibold text-slate-500 transition hover:text-slate-900 dark:hover:text-white">Powered by based on logistics</NuxtLink>
      </footer>
    </main>
  </div>
</template>
