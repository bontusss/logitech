<script setup lang="ts">
import {
  ArrowDownLeft,
  ArrowUpRight,
  Check,
  CircleAlert,
  ClipboardList,
  Copy,
  Loader2,
  MapPin,
  Plus,
  Search,
  Truck,
  UserRound,
  X,
  MessageCircle,
  ExternalLink,
  Clock3,
  Link,
  Share2,
  Sparkles,
} from 'lucide-vue-next'
import type { CompanyDetail } from '~/types/auth'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Deliveries' })

interface Driver {
  id: string
  full_name: string
  phone: string
  vehicle_plate: string
  status: 'active' | 'suspended' | 'inactive'
}

interface Delivery {
  id: string
  sender_name: string
  sender_phone: string
  receiver_name: string
  receiver_phone: string
  pickup_address: string
  dropoff_address: string
  item_description: string
  declared_value_minor: number
  delivery_price_minor: number
  payer: 'sender' | 'receiver'
  is_paid: boolean
  payment_method: 'cash' | 'pos' | 'bank_transfer' | null
  assigned_driver_id: string | null
  assigned_driver_name: string | null
  package_category?: string
  weight_kg?: number
  additional_instructions?: string
  package_photo_url?: string | null
  price_is_estimate?: boolean
  booking_source?: 'dashboard' | 'booking_page'
  delivery_speed?: 'standard' | 'express' | 'same_day'
  distance_km?: number
  status: 'requested' | 'created' | 'assigned' | 'picked_up' | 'in_transit' | 'delivered' | 'cancelled'
  created_at: string
  tracking_code?: string | null
  tracking_url?: string | null
  rider_url?: string | null
  updated_at?: string
}

type DeliveryTab = 'dispatch' | 'all'
type DispatchStage = 'inbox' | 'active' | 'done'

interface BookingRequest {
  id: string
  created_at: string
  sender_name: string
  receiver_name: string
  pickup_address: string
  dropoff_address: string
  item_description: string
  delivery_price_minor: number
  request_status: 'pending' | 'accepted' | 'declined'
}

const { api } = useApi()
const { detail, load, can } = useWorkspace()

const activeTab = ref<DeliveryTab>('dispatch')
const dispatchStage = ref<DispatchStage>('inbox')
const deliveries = ref<Delivery[]>([])
const bookingRequests = ref<BookingRequest[]>([])
const drivers = ref<Driver[]>([])
const loading = ref(true)
const saving = ref(false)
const formOpen = ref(false)
const error = ref('')
const searchQuery = ref('')
const copiedTrackingId = ref('')
const currency = computed(() => detail.value?.settings.currency || 'NGN')

const form = reactive({
  sender_name: '',
  sender_phone: '',
  receiver_name: '',
  receiver_phone: '',
  pickup_address: '',
  dropoff_address: '',
  item_description: '',
  declared_value: '',
  delivery_price: '',
  payer: 'sender' as 'sender' | 'receiver',
  is_paid: false,
  payment_method: '' as '' | 'cash' | 'pos' | 'bank_transfer',
  assigned_driver_id: '',
})

const availableDrivers = computed(() => drivers.value.filter((driver) => driver.status === 'active'))
const pendingRequests = computed(() =>
  bookingRequests.value.filter((request) => request.request_status === 'pending'),
)
const activeDeliveries = computed(() =>
  deliveries.value.filter((delivery) =>
    Boolean(delivery.assigned_driver_id) && !['delivered', 'cancelled'].includes(delivery.status),
  ),
)
const doneTodayDeliveries = computed(() => {
  const today = new Date().toDateString()
  return deliveries.value.filter((delivery) =>
    delivery.status === 'delivered'
    && delivery.updated_at
    && new Date(delivery.updated_at).toDateString() === today,
  )
})
const filteredDeliveries = computed(() => {
  const query = searchQuery.value.trim().toLocaleLowerCase()
  if (!query) return deliveries.value
  return deliveries.value.filter((delivery) =>
    [
      delivery.id,
      delivery.sender_name,
      delivery.sender_phone,
      delivery.receiver_name,
      delivery.receiver_phone,
      delivery.pickup_address,
      delivery.dropoff_address,
      delivery.item_description,
      delivery.assigned_driver_name ?? '',
      delivery.status,
    ].some((value) => value.toLocaleLowerCase().includes(query)),
  )
})

const formPriceMinor = computed(() => toMinorUnits(form.delivery_price))
const formDeclaredValueMinor = computed(() => toMinorUnits(form.declared_value))

onMounted(async () => {
  await load()
  await Promise.all([fetchDeliveries(), fetchDrivers(), fetchBookingRequests()])
  if (detailOpenId.value) {
    try { timeline.value=(await api<{events:Array<{status:string;note:string|null;created_at:string}>}>(`/companies/me/deliveries/${detailOpenId.value}/timeline`)).events } catch { timeline.value=[] }
  }
  if (route.query.create === '1' && can('shipments.create')) formOpen.value = true
  window.addEventListener('keydown', onModalKeydown)
})

const route = useRoute()
const detailOpenId = ref<string>(String(route.query.id || ''))
const timeline = ref<Array<{ status: string; note: string | null; created_at: string }>>([])

watch(detailOpenId, async (id) => {
  if (!id) { timeline.value=[]; return }
  try { timeline.value=(await api<{events:Array<{status:string;note:string|null;created_at:string}>}>(`/companies/me/deliveries/${id}/timeline`)).events }
  catch { timeline.value=[] }
})
onBeforeUnmount(() => window.removeEventListener('keydown', onModalKeydown))

function toMinorUnits(amount: string) {
  const parsed = Number(amount)
  return Number.isFinite(parsed) && parsed >= 0 ? Math.round(parsed * 100) : 0
}

function formatMoney(minor: number) {
  return new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency: currency.value,
    maximumFractionDigits: 2,
  }).format(minor / 100)
}

function closeForm() {
  if (!saving.value) {
    formOpen.value = false
    error.value = ''
  }
}

function onModalKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') closeForm()
}

function whatsappHref(phone: string, message: string) {
  const digits = phone.replace(/[^0-9]/g, '').replace(/^0/, '234')
  return `https://wa.me/${digits}?text=${encodeURIComponent(message)}`
}

async function copyTrackingLink(delivery: Delivery) {
  if (!delivery.tracking_url) return
  try {
    await navigator.clipboard.writeText(delivery.tracking_url)
    copiedTrackingId.value = delivery.id
    window.setTimeout(() => {
      if (copiedTrackingId.value === delivery.id) copiedTrackingId.value = ''
    }, 1800)
  } catch {
    error.value = 'Could not copy the tracking link. Open it and copy from the address bar.'
  }
}

async function shareTrackingLink(delivery: Delivery) {
  if (!delivery.tracking_url) return
  const shareData = {
    title: `Track delivery ${delivery.tracking_code || deliveryReference(delivery.id)}`,
    text: `${delivery.sender_name} → ${delivery.receiver_name}: follow this delivery in real time.`,
    url: delivery.tracking_url,
  }
  if (navigator.share) {
    try {
      await navigator.share(shareData)
      return
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') return
    }
  }
  await copyTrackingLink(delivery)
}

async function reassignDriver(delivery: Delivery, driverId: string) {
  try {
    await api(`/companies/me/deliveries/${delivery.id}/driver`, { method: 'PUT', body: { driver_id: driverId || null } })
    await fetchDeliveries()
  } catch (err) { error.value = apiErrorMessage(err, 'Could not update the assigned driver.') }
}

async function updateDeliveryStatus(delivery: Delivery, status: Delivery['status']) {
  try {
    await api(`/companies/me/deliveries/${delivery.id}/status`, { method: 'PATCH', body: { status } })
    await fetchDeliveries()
  } catch (err) { error.value = apiErrorMessage(err, 'Could not update the delivery status.') }
}

async function fetchDeliveries() {
  loading.value = true
  try {
    deliveries.value = (await api<{ deliveries: Delivery[] }>('/companies/me/deliveries')).deliveries
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not load deliveries.')
  } finally {
    loading.value = false
  }
}

async function fetchDrivers() {
  try {
    drivers.value = (await api<{ drivers: Driver[] }>('/companies/me/drivers')).drivers
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not load drivers for assignment.')
  }
}

async function fetchBookingRequests() {
  try {
    bookingRequests.value = (await api<{ requests: BookingRequest[] }>('/companies/me/requests')).requests
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not load booking requests for the dispatch inbox.')
  }
}

async function createDelivery() {
  error.value = ''
  saving.value = true
  try {
    await api('/companies/me/deliveries', {
      method: 'POST',
      body: {
        sender_name: form.sender_name.trim(),
        sender_phone: form.sender_phone.trim(),
        receiver_name: form.receiver_name.trim(),
        receiver_phone: form.receiver_phone.trim(),
        pickup_address: form.pickup_address.trim(),
        dropoff_address: form.dropoff_address.trim(),
        item_description: form.item_description.trim(),
        declared_value_minor: formDeclaredValueMinor.value,
        delivery_price_minor: formPriceMinor.value,
        payer: form.payer,
        is_paid: form.is_paid,
        payment_method: form.is_paid ? form.payment_method : '',
        assigned_driver_id: form.assigned_driver_id || null,
      },
    })
    formOpen.value = false
    resetForm()
    await fetchDeliveries()
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not create this delivery.')
  } finally {
    saving.value = false
  }
}

function resetForm() {
  Object.assign(form, {
    sender_name: '',
    sender_phone: '',
    receiver_name: '',
    receiver_phone: '',
    pickup_address: '',
    dropoff_address: '',
    item_description: '',
    declared_value: '',
    delivery_price: '',
    payer: 'sender',
    is_paid: false,
    payment_method: '',
    assigned_driver_id: '',
  })
}

function deliveryReference(id: string) {
  return `DEL-${id.slice(0, 8).toUpperCase()}`
}

function statusLabel(status: Delivery['status']) {
  return status.replace('_', ' ')
}

function paymentLabel(delivery: Delivery) {
  return delivery.is_paid
    ? `Paid · ${delivery.payment_method?.replace('_', ' ')}`
    : `Unpaid · ${delivery.payer} pays`
}
</script>

<template>
  <div class="delivery-workspace dash-page">
    <header class="delivery-page-hero relative overflow-hidden rounded-2xl border p-5 sm:p-6">
      <div class="pointer-events-none absolute -top-20 right-12 size-48 rounded-full bg-fuchsia-400/15 blur-3xl" />
      <div class="pointer-events-none absolute -right-10 -bottom-28 size-56 rounded-full bg-cyan-400/15 blur-3xl" />
      <div class="relative flex flex-wrap items-center justify-between gap-4">
        <div>
          <p class="dash-eyebrow inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-semibold tracking-wide">
            <Sparkles :size="12" /> MOVE IT FORWARD
          </p>
          <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Deliveries</h1>
          <p class="mt-1.5 max-w-xl text-[12px] leading-relaxed text-ink-900/65">
            Every pickup, handoff and payment detail, in one place.
          </p>
        </div>
        <button
          v-if="can('shipments.create')"
          type="button"
          class="inline-flex min-h-10 items-center gap-1.5 rounded-xl bg-brand-600 px-4 py-2.5 text-[12px] font-semibold text-white shadow-sm shadow-brand-600/20 transition hover:-translate-y-0.5 hover:bg-brand-700"
          @click="formOpen = true"
        >
          <Plus :size="14" /> Create delivery
        </button>
      </div>
    </header>

    <p v-if="error && !formOpen" role="alert" class="mt-5 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <div class="mt-5 inline-flex rounded-xl border border-ink-900/10 bg-white p-1 dark:bg-white/5" role="tablist" aria-label="Delivery views">
      <button
        type="button"
        role="tab"
        :aria-selected="activeTab === 'dispatch'"
        class="inline-flex items-center gap-2 rounded-lg px-3.5 py-2 text-[12px] font-semibold transition"
        :class="activeTab === 'dispatch' ? 'bg-brand-600 text-white shadow-sm shadow-brand-600/25' : 'text-ink-900/55 hover:text-ink-900'"
        @click="activeTab = 'dispatch'"
      >
        <Truck :size="14" /> Dispatch board
      </button>
      <button
        type="button"
        role="tab"
        :aria-selected="activeTab === 'all'"
        class="inline-flex items-center gap-2 rounded-lg px-3.5 py-2 text-[12px] font-semibold transition"
        :class="activeTab === 'all' ? 'bg-brand-600 text-white shadow-sm shadow-brand-600/25' : 'text-ink-900/55 hover:text-ink-900'"
        @click="activeTab = 'all'"
      >
        <ClipboardList :size="14" /> All deliveries
      </button>
    </div>

    <section v-if="activeTab === 'dispatch'" class="mt-5">
      <div class="grid grid-cols-3 gap-2 sm:gap-3" role="tablist" aria-label="Dispatch stages">
        <button
          v-for="stage in ([
            { id: 'inbox', label: 'Inbox', count: pendingRequests.length, icon: ClipboardList, tone: 'amber' },
            { id: 'active', label: 'Active', count: activeDeliveries.length, icon: Truck, tone: 'blue' },
            { id: 'done', label: 'Done today', count: doneTodayDeliveries.length, icon: Check, tone: 'mint' },
          ] as const)"
          :key="stage.id"
          type="button"
          role="tab"
          :aria-selected="dispatchStage === stage.id"
          class="dash-stat rounded-xl border p-3.5 text-left sm:p-4"
          :class="[`dash-stat--${stage.tone}`, dispatchStage === stage.id ? 'dash-stat--active' : '']"
          @click="dispatchStage = stage.id"
        >
          <span class="flex items-center justify-between gap-2">
            <span class="text-[10px] font-semibold tracking-wide text-ink-900/50 uppercase">{{ stage.label }}</span>
            <span class="flex size-6 items-center justify-center rounded-lg bg-white/80 text-ink-900/45 dark:bg-white/10">
              <component :is="stage.icon" :size="12" />
            </span>
          </span>
          <span class="mt-2 block text-[22px] leading-none font-semibold tracking-tight">{{ stage.count }}</span>
        </button>
      </div>

      <div class="dash-panel mt-4 overflow-hidden rounded-2xl border">
        <div class="flex flex-wrap items-center justify-between gap-3 border-b border-ink-900/8 p-4">
          <div>
            <h2 class="text-[13px] font-semibold">
              {{ dispatchStage === 'inbox' ? 'Booking requests' : dispatchStage === 'active' ? 'Active deliveries' : 'Completed today' }}
            </h2>
            <p class="mt-0.5 text-[10.5px] text-ink-900/40">
              {{ dispatchStage === 'inbox' ? 'New bookings from your provider page' : dispatchStage === 'active' ? 'Deliveries currently assigned to a driver' : 'Deliveries marked delivered today' }}
            </p>
          </div>
          <button
            v-if="dispatchStage === 'inbox' && pendingRequests.length"
            type="button"
            class="rounded-lg border border-brand-200 px-3 py-2 text-[10.5px] font-semibold text-brand-700"
            @click="navigateTo('/dashboard/requests')"
          >
            Manage requests
          </button>
          <button
            v-if="dispatchStage === 'active' || dispatchStage === 'done'"
            type="button"
            class="rounded-lg border border-ink-900/10 px-3 py-2 text-[10.5px] font-semibold text-ink-900/60"
            @click="activeTab = 'all'"
          >
            View all deliveries
          </button>
        </div>

        <div v-if="loading" class="flex items-center justify-center gap-2 p-10 text-[12px] text-ink-900/50">
          <Loader2 :size="15" class="animate-spin" /> Loading dispatch board…
        </div>
        <div
          v-else-if="dispatchStage === 'inbox' && pendingRequests.length === 0"
          class="p-8 text-center"
        >
          <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-gradient-to-br from-amber-100 to-orange-50 text-amber-700 dark:from-amber-400/15 dark:to-orange-400/10 dark:text-amber-200">
            <ClipboardList :size="19" />
          </span>
          <h3 class="mt-3 text-[13px] font-semibold">Inbox is clear</h3>
          <p class="mt-1 text-[11.5px] text-ink-900/45">New bookings from your provider page will appear here.</p>
        </div>
        <div v-else-if="dispatchStage === 'inbox'" class="divide-y divide-ink-900/8">
          <article v-for="request in pendingRequests" :key="request.id" class="p-4 sm:p-5">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p class="text-[9.5px] text-ink-900/40">{{ new Date(request.created_at).toLocaleString() }}</p>
                <h3 class="mt-1 text-[12px] font-semibold">{{ request.sender_name }} → {{ request.receiver_name }}</h3>
              </div>
              <span class="rounded-full bg-amber-50 px-2.5 py-1 text-[9px] font-semibold text-amber-700">New request</span>
            </div>
            <div class="mt-3 grid gap-2 text-[10.5px] sm:grid-cols-2">
              <p class="break-words text-ink-900/55"><span class="font-semibold text-ink-900/75">From:</span> {{ request.pickup_address }}</p>
              <p class="break-words text-ink-900/55"><span class="font-semibold text-ink-900/75">To:</span> {{ request.dropoff_address }}</p>
            </div>
            <div class="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-ink-900/8 pt-3">
              <div>
                <p class="text-[10.5px] text-ink-900/55">{{ request.item_description }}</p>
                <p class="mt-1 text-[11.5px] font-semibold">{{ request.delivery_price_minor ? formatMoney(request.delivery_price_minor) : 'Price to confirm' }}</p>
              </div>
              <button type="button" class="rounded-lg bg-brand-600 px-3 py-2 text-[10.5px] font-semibold text-white" @click="navigateTo('/dashboard/requests')">Review request</button>
            </div>
          </article>
        </div>
        <div
          v-else-if="dispatchStage === 'active' && activeDeliveries.length === 0"
          class="p-8 text-center"
        >
          <Truck :size="22" class="mx-auto text-ink-900/25" />
          <h3 class="mt-3 text-[13px] font-semibold">No active deliveries</h3>
          <p class="mt-1 text-[11.5px] text-ink-900/45">Deliveries assigned to a driver will appear here.</p>
        </div>
        <div v-else-if="dispatchStage === 'active'" class="divide-y divide-ink-900/8">
          <article v-for="delivery in activeDeliveries" :key="delivery.id" class="flex flex-wrap items-center justify-between gap-3 p-4 sm:px-5">
            <div class="min-w-0">
              <p class="text-[9.5px] text-ink-900/40">{{ deliveryReference(delivery.id) }} · {{ new Date(delivery.created_at).toLocaleString() }}</p>
              <h3 class="mt-1 truncate text-[12px] font-semibold">{{ delivery.sender_name }} → {{ delivery.receiver_name }}</h3>
              <p class="mt-1 truncate text-[10.5px] text-ink-900/45">{{ delivery.pickup_address }} → {{ delivery.dropoff_address }}</p>
              <p class="mt-1 text-[10.5px] font-medium text-brand-700">{{ delivery.assigned_driver_name || 'Driver assigned' }} · <span class="capitalize">{{ statusLabel(delivery.status) }}</span></p>
            </div>
            <div class="flex shrink-0 flex-wrap items-center gap-1.5">
              <div v-if="delivery.tracking_url" class="flex items-center gap-1 rounded-lg border border-brand-200 bg-brand-50/50 p-1">
                <span class="px-1 text-[9px] font-semibold text-brand-800 dark:text-indigo-200">{{ delivery.tracking_code || 'Tracking' }}</span>
                <button type="button" class="inline-flex items-center gap-1 rounded-md bg-white px-1.5 py-1 text-[9px] font-semibold text-brand-700" @click="copyTrackingLink(delivery)">
                  <Check v-if="copiedTrackingId === delivery.id" :size="10" />
                  <Copy v-else :size="10" />
                  {{ copiedTrackingId === delivery.id ? 'Copied' : 'Copy' }}
                </button>
                <button type="button" class="inline-flex items-center gap-1 rounded-md bg-white px-1.5 py-1 text-[9px] font-semibold text-brand-700" @click="shareTrackingLink(delivery)">
                  <Share2 :size="10" /> Share
                </button>
              </div>
              <button type="button" class="rounded-lg border border-brand-200 px-3 py-2 text-[10.5px] font-semibold text-brand-700" @click="activeTab = 'all'; detailOpenId = delivery.id">
                Manage delivery
              </button>
            </div>
          </article>
        </div>
        <div
          v-else-if="dispatchStage === 'done' && doneTodayDeliveries.length === 0"
          class="p-8 text-center"
        >
          <Check :size="22" class="mx-auto text-ink-900/25" />
          <h3 class="mt-3 text-[13px] font-semibold">Nothing completed today yet</h3>
          <p class="mt-1 text-[11.5px] text-ink-900/45">Deliveries marked delivered today will be summarized here.</p>
        </div>
        <div v-else class="divide-y divide-ink-900/8">
          <article v-for="delivery in doneTodayDeliveries" :key="delivery.id" class="flex flex-wrap items-center justify-between gap-3 p-4 sm:px-5">
            <div class="min-w-0">
              <p class="text-[9.5px] text-ink-900/40">{{ new Date(delivery.updated_at || delivery.created_at).toLocaleTimeString() }} · {{ deliveryReference(delivery.id) }}</p>
              <h3 class="mt-1 truncate text-[12px] font-semibold">{{ delivery.sender_name }} → {{ delivery.receiver_name }}</h3>
              <p class="mt-1 truncate text-[10.5px] text-ink-900/45">{{ delivery.pickup_address }} → {{ delivery.dropoff_address }}</p>
            </div>
            <span class="rounded-full bg-emerald-50 px-2.5 py-1 text-[9px] font-semibold text-emerald-700">Delivered · {{ formatMoney(delivery.delivery_price_minor) }}</span>
          </article>
        </div>
      </div>
    </section>

    <section v-else class="dash-panel mt-4 overflow-hidden rounded-2xl border">
      <div class="flex flex-col gap-3 border-b border-ink-900/8 p-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-[13px] font-semibold">All deliveries</h2>
          <p class="mt-0.5 text-[10.5px] text-ink-900/40">{{ deliveries.length }} total</p>
        </div>
        <label class="relative block w-full sm:max-w-xs">
          <Search :size="15" class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-ink-900/35" />
          <input
            v-model="searchQuery"
            type="search"
            placeholder="Search deliveries"
            aria-label="Search deliveries"
            class="w-full rounded-lg border border-ink-900/10 py-2.5 pr-3 pl-9 text-[11.5px] outline-none focus:border-brand-500"
          />
        </label>
      </div>
      <div v-if="loading" class="flex items-center gap-2 p-6 text-[12px] text-ink-900/50">
        <Loader2 :size="15" class="animate-spin" /> Loading deliveries…
      </div>
      <div v-else-if="filteredDeliveries.length === 0" class="p-8 text-center">
        <ClipboardList :size="22" class="mx-auto text-ink-900/25" />
        <h3 class="mt-3 text-[13px] font-semibold">{{ deliveries.length ? 'No matching deliveries' : 'No deliveries yet' }}</h3>
        <p class="mt-1 text-[11.5px] text-ink-900/45">
          {{ deliveries.length ? 'Try another search.' : 'Create a delivery to get started.' }}
        </p>
      </div>
      <div v-else class="divide-y divide-ink-900/8">
        <article v-for="delivery in filteredDeliveries" :key="delivery.id" class="p-4 transition hover:bg-brand-50/20 sm:p-5">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="flex items-center gap-2.5">
              <span class="flex size-9 items-center justify-center rounded-xl bg-gradient-to-br from-sky-100 to-cyan-50 text-sky-700 dark:from-sky-400/15 dark:to-cyan-400/10 dark:text-sky-200"><Truck :size="16" /></span>
              <div>
                <h3 class="text-[12px] font-semibold">{{ deliveryReference(delivery.id) }}</h3>
                <p class="mt-0.5 text-[10px] capitalize text-ink-900/40">{{ statusLabel(delivery.status) }} · {{ new Date(delivery.created_at).toLocaleString() }}</p>
              </div>
            </div>
            <p v-if="delivery.price_is_estimate && !delivery.delivery_price_minor" class="rounded-full bg-amber-50 px-2.5 py-1 text-[9px] font-semibold text-amber-800">Price to confirm</p>
            <p v-else class="text-[14px] font-semibold tracking-tight">{{ formatMoney(delivery.delivery_price_minor) }}</p>
          </div>
          <div class="mt-4 grid gap-3 text-[11px] sm:grid-cols-2">
            <div class="flex gap-2">
              <ArrowUpRight :size="14" class="mt-0.5 shrink-0 text-brand-600" />
              <div class="min-w-0">
                <p class="font-medium">{{ delivery.sender_name }} · {{ delivery.sender_phone }}</p>
                <p class="mt-0.5 break-words text-ink-900/45">{{ delivery.pickup_address }}</p>
              </div>
            </div>
            <div class="flex gap-2">
              <ArrowDownLeft :size="14" class="mt-0.5 shrink-0 text-rose-600" />
              <div class="min-w-0">
                <p class="font-medium">{{ delivery.receiver_name }} · {{ delivery.receiver_phone }}</p>
                <p class="mt-0.5 break-words text-ink-900/45">{{ delivery.dropoff_address }}</p>
              </div>
            </div>
          </div>
          <div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 border-t border-ink-900/8 pt-3 text-[10.5px] text-ink-900/50">
            <span class="capitalize">{{ delivery.package_category || delivery.item_description }}<template v-if="delivery.weight_kg"> · {{ delivery.weight_kg }} kg</template><template v-if="delivery.distance_km"> · {{ delivery.distance_km }} km</template></span>
            <span v-if="delivery.delivery_speed" class="rounded-full bg-cyan-50 px-2 py-0.5 font-semibold capitalize text-cyan-700">{{ delivery.delivery_speed.replace('_', ' ') }}</span>
            <span v-if="delivery.booking_source === 'booking_page'" class="rounded-full bg-violet-50 px-2 py-0.5 font-semibold text-violet-700">Online request</span>
            <span>{{ paymentLabel(delivery) }}</span>
            <span class="inline-flex items-center gap-1">
              <UserRound :size="12" />
              {{ delivery.assigned_driver_name ?? 'No driver assigned' }}
            </span>
          </div>
          <p v-if="delivery.additional_instructions" class="mt-2 text-[10.5px] leading-5 text-ink-900/50">{{ delivery.additional_instructions }}</p>
          <div class="mt-3 flex flex-wrap items-center gap-2 border-t border-ink-900/8 pt-3">
            <select :value="delivery.assigned_driver_id || ''" :disabled="!can('shipments.update')" class="rounded-lg border border-ink-900/10 bg-white px-2.5 py-2 text-[10.5px]" @change="reassignDriver(delivery, ($event.target as HTMLSelectElement).value)">
              <option value="">Remove driver</option>
              <option v-for="driver in availableDrivers" :key="driver.id" :value="driver.id">{{ driver.full_name }} · {{ driver.vehicle_plate }}</option>
            </select>
            <select :value="delivery.status" :disabled="!can('shipments.update')" class="rounded-lg border border-ink-900/10 bg-white px-2.5 py-2 text-[10.5px]" @change="updateDeliveryStatus(delivery, ($event.target as HTMLSelectElement).value as Delivery['status'])">
              <option value="created">Created</option><option value="assigned">Assigned</option><option value="picked_up">Picked up</option><option value="in_transit">In transit</option><option value="delivered">Delivered</option><option value="cancelled">Cancelled</option>
            </select>
            <a :href="whatsappHref(delivery.sender_phone, `Hello ${delivery.sender_name}, here is an update for your delivery ${deliveryReference(delivery.id)}.`)" target="_blank" rel="noopener" class="inline-flex items-center gap-1 rounded-lg border border-emerald-200 px-2.5 py-2 text-[10.5px] font-semibold text-emerald-700"><MessageCircle :size="12" />Sender</a>
            <a :href="whatsappHref(delivery.receiver_phone, `Hello ${delivery.receiver_name}, here is an update for your delivery ${deliveryReference(delivery.id)}.`)" target="_blank" rel="noopener" class="inline-flex items-center gap-1 rounded-lg border border-emerald-200 px-2.5 py-2 text-[10.5px] font-semibold text-emerald-700"><MessageCircle :size="12" />Receiver</a>
            <button type="button" class="inline-flex items-center gap-1 rounded-lg border border-ink-900/10 px-2.5 py-2 text-[10.5px] font-semibold" @click="detailOpenId=detailOpenId===delivery.id ? '' : delivery.id"><Clock3 :size="12" />Timeline</button>
            <div v-if="delivery.tracking_url" class="flex w-full flex-wrap items-center gap-1.5 rounded-lg border border-brand-200 bg-brand-50/45 px-2.5 py-2 sm:w-auto dark:border-brand-400/20 dark:bg-brand-400/5">
              <span class="inline-flex items-center gap-1 text-[9.5px] font-semibold text-brand-800 dark:text-indigo-200">
                <Link :size="11" /> {{ delivery.tracking_code || 'Tracking link' }}
              </span>
              <button
                type="button"
                class="inline-flex items-center gap-1 rounded-md border border-brand-200 bg-white px-1.5 py-1 text-[9.5px] font-semibold text-brand-700 hover:bg-brand-50 dark:border-brand-400/20 dark:bg-white/5 dark:text-indigo-200"
                @click="copyTrackingLink(delivery)"
              >
                <Check v-if="copiedTrackingId === delivery.id" :size="10" />
                <Copy v-else :size="10" />
                {{ copiedTrackingId === delivery.id ? 'Copied' : 'Copy' }}
              </button>
              <button
                type="button"
                class="inline-flex items-center gap-1 rounded-md border border-brand-200 bg-white px-1.5 py-1 text-[9.5px] font-semibold text-brand-700 hover:bg-brand-50 dark:border-brand-400/20 dark:bg-white/5 dark:text-indigo-200"
                @click="shareTrackingLink(delivery)"
              >
                <Share2 :size="10" /> Share
              </button>
              <a
                :href="delivery.tracking_url"
                target="_blank"
                rel="noopener"
                class="inline-flex items-center gap-1 rounded-md border border-brand-200 bg-white px-1.5 py-1 text-[9.5px] font-semibold text-brand-700 hover:bg-brand-50 dark:border-brand-400/20 dark:bg-white/5 dark:text-indigo-200"
              >
                <ExternalLink :size="10" /> Open
              </a>
            </div>
          </div>
          <div v-if="detailOpenId===delivery.id" class="mt-3 rounded-xl border border-ink-900/8 bg-canvas p-3">
            <p class="text-[10.5px] font-semibold">Journey timeline</p>
            <div v-if="timeline.length" class="mt-2 space-y-2"><div v-for="event in timeline" :key="event.created_at" class="flex gap-2 text-[10.5px]"><span class="mt-1 size-1.5 shrink-0 rounded-full bg-brand-500"/><div><p class="font-medium capitalize">{{ statusLabel(event.status as Delivery['status']) }}</p><p class="text-ink-900/40">{{ event.note || 'Status updated' }} · {{ new Date(event.created_at).toLocaleString() }}</p></div></div></div>
            <p v-else class="mt-2 text-[10.5px] text-ink-900/40">No timeline events yet.</p>
          </div>
        </article>
      </div>
    </section>

    <div
      v-if="formOpen"
      class="fixed inset-0 z-[80] flex items-center justify-center bg-ink-950/55 px-3 py-4 backdrop-blur-md sm:px-5 sm:py-6"
      role="dialog"
      aria-modal="true"
      aria-labelledby="create-delivery-title"
      @click.self="closeForm"
    >
      <form
        class="delivery-modal max-h-full w-full max-w-6xl overflow-y-auto rounded-2xl border shadow-2xl"
        @submit.prevent="createDelivery"
      >
        <header class="delivery-modal-heading sticky top-0 z-10 flex items-start justify-between gap-4 border-b px-5 py-4 sm:px-6">
          <div class="flex items-start gap-3">
            <span class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-cyan-100 text-cyan-800 dark:bg-cyan-400/15 dark:text-cyan-200"><Truck :size="18" /></span>
            <div>
              <p class="text-[9.5px] font-semibold tracking-[0.16em] text-brand-700 uppercase dark:text-indigo-200">NEW DELIVERY</p>
              <h2 id="create-delivery-title" class="mt-1 text-[15px] font-semibold">Let’s set up this delivery</h2>
              <p class="mt-1 text-[11px] text-ink-900/50">Add the route, people and payment. Your receipt updates as you go.</p>
            </div>
          </div>
          <button type="button" class="rounded-lg p-1.5 text-ink-900/45 hover:bg-ink-900/5" aria-label="Close dialog" :disabled="saving" @click="closeForm">
            <X :size="18" />
          </button>
        </header>

        <p v-if="error" role="alert" class="mx-5 mt-4 rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700 sm:mx-6">
          {{ error }}
        </p>

        <div class="grid gap-4 p-4 sm:p-5 lg:grid-cols-[minmax(0,1.4fr)_minmax(280px,0.8fr)]">
          <div class="space-y-3">
            <section class="delivery-form-section delivery-form-section--route rounded-xl border p-4">
              <h3 class="inline-flex items-center gap-2 text-[12px] font-semibold"><MapPin :size="14" class="text-cyan-700 dark:text-cyan-200" /> Route</h3>
              <div class="mt-3 grid gap-3 sm:grid-cols-2">
                <label class="block">
                  <span class="mb-1.5 block text-[10.5px] font-medium text-ink-900/55">Pickup address</span>
                  <textarea v-model="form.pickup_address" required minlength="5" maxlength="500" rows="2" autocomplete="street-address" class="w-full resize-y rounded-lg border border-slate-300 px-3 py-2.5 text-[12px] focus:border-brand-500" />
                </label>
                <label class="block">
                  <span class="mb-1.5 block text-[10.5px] font-medium text-ink-900/55">Drop-off address</span>
                  <textarea v-model="form.dropoff_address" required minlength="5" maxlength="500" rows="2" autocomplete="street-address" class="w-full resize-y rounded-lg border border-slate-300 px-3 py-2.5 text-[12px] focus:border-brand-500" />
                </label>
              </div>
            </section>

            <section class="delivery-form-section delivery-form-section--contacts rounded-xl border p-4">
              <h3 class="inline-flex items-center gap-2 text-[12px] font-semibold"><UserRound :size="14" class="text-violet-700 dark:text-violet-200" /> People</h3>
              <div class="mt-3 grid gap-4 sm:grid-cols-2">
                <div class="space-y-3">
                  <p class="inline-flex items-center gap-1.5 text-[10px] font-semibold text-brand-700"><ArrowUpRight :size="13" /> Sender</p>
                  <label class="block">
                    <span class="mb-1 block text-[10.5px] text-ink-900/55">Customer name</span>
                    <input v-model="form.sender_name" required minlength="2" maxlength="120" autocomplete="name" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px] focus:border-brand-500" />
                  </label>
                  <label class="block">
                    <span class="mb-1 block text-[10.5px] text-ink-900/55">Phone number</span>
                    <input v-model="form.sender_phone" required minlength="7" maxlength="30" type="tel" autocomplete="tel" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px] focus:border-brand-500" />
                  </label>
                </div>
                <div class="space-y-3">
                  <p class="inline-flex items-center gap-1.5 text-[10px] font-semibold text-rose-700"><ArrowDownLeft :size="13" /> Receiver</p>
                  <label class="block">
                    <span class="mb-1 block text-[10.5px] text-ink-900/55">Receiver name</span>
                    <input v-model="form.receiver_name" required minlength="2" maxlength="120" autocomplete="name" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px] focus:border-brand-500" />
                  </label>
                  <label class="block">
                    <span class="mb-1 block text-[10.5px] text-ink-900/55">Phone number</span>
                    <input v-model="form.receiver_phone" required minlength="7" maxlength="30" type="tel" autocomplete="tel" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px] focus:border-brand-500" />
                  </label>
                </div>
              </div>
            </section>

            <section class="delivery-form-section delivery-form-section--payment rounded-xl border p-4">
              <h3 class="inline-flex items-center gap-2 text-[12px] font-semibold"><ClipboardList :size="14" class="text-amber-700 dark:text-amber-200" /> Parcel & payment</h3>
              <label class="mt-3 block">
                <span class="mb-1.5 block text-[10.5px] font-medium text-ink-900/55">Item description</span>
                <textarea v-model="form.item_description" required minlength="2" maxlength="500" rows="2" class="w-full resize-y rounded-lg border border-slate-300 px-3 py-2.5 text-[12px] focus:border-brand-500" />
              </label>
              <div class="mt-3 grid gap-3 sm:grid-cols-2">
                <label class="block">
                  <span class="mb-1.5 block text-[10.5px] font-medium text-ink-900/55">Declared value ({{ currency }})</span>
                  <input v-model="form.declared_value" required type="number" min="0" step="0.01" inputmode="decimal" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px] focus:border-brand-500" />
                </label>
                <label class="block">
                  <span class="mb-1.5 block text-[10.5px] font-medium text-ink-900/55">Delivery price ({{ currency }})</span>
                  <input v-model="form.delivery_price" required type="number" min="0" step="0.01" inputmode="decimal" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px] focus:border-brand-500" />
                </label>
                <label class="block">
                  <span class="mb-1.5 block text-[10.5px] font-medium text-ink-900/55">Who pays?</span>
                  <select v-model="form.payer" class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-[12px] focus:border-brand-500">
                    <option value="sender">Sender</option>
                    <option value="receiver">Receiver</option>
                  </select>
                </label>
                <div class="flex items-end">
                  <label class="flex min-h-[40px] w-full items-center gap-2 rounded-lg border border-slate-200 px-3 py-2 text-[11px]">
                    <input v-model="form.is_paid" type="checkbox" class="size-4 accent-brand-600" />
                    <span>Already paid</span>
                  </label>
                </div>
                <label v-if="form.is_paid" class="block sm:col-span-2">
                  <span class="mb-1.5 block text-[10.5px] font-medium text-ink-900/55">Payment method</span>
                  <select v-model="form.payment_method" required class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-[12px] focus:border-brand-500">
                    <option disabled value="">Select a method</option>
                    <option value="cash">Cash</option>
                    <option value="pos">POS</option>
                    <option value="bank_transfer">Bank transfer</option>
                  </select>
                </label>
              </div>
            </section>

            <section class="delivery-form-section delivery-form-section--fulfillment rounded-xl border p-4">
              <h3 class="inline-flex items-center gap-2 text-[12px] font-semibold"><Truck :size="14" class="text-emerald-700 dark:text-emerald-200" /> Fulfillment</h3>
              <label class="mt-3 block">
                <span class="mb-1.5 block text-[10.5px] font-medium text-ink-900/55">Assign a driver (optional)</span>
                <select v-model="form.assigned_driver_id" class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-[12px] focus:border-brand-500">
                  <option value="">Leave unassigned</option>
                  <option v-for="driver in availableDrivers" :key="driver.id" :value="driver.id">
                    {{ driver.full_name }} · {{ driver.vehicle_plate }}
                  </option>
                </select>
                <span v-if="availableDrivers.length === 0" class="mt-1.5 block text-[10px] text-ink-900/40">Add an active driver to assign this delivery.</span>
              </label>
            </section>
          </div>

          <aside class="space-y-3 lg:sticky lg:top-[88px] lg:self-start">
            <div class="delivery-map-card overflow-hidden rounded-xl border bg-white">
              <DeliveryMapPreview :pickup-address="form.pickup_address" :dropoff-address="form.dropoff_address" />
            </div>
            <section class="delivery-receipt rounded-xl border p-4" aria-live="polite">
              <div class="flex items-center justify-between border-b border-dashed border-ink-900/15 pb-3">
                <div>
                  <h3 class="text-[12px] font-semibold">Delivery receipt</h3>
                  <p class="mt-0.5 text-[9.5px] text-ink-900/40">Live preview · {{ currency }}</p>
                </div>
                <MapPin :size="16" class="text-brand-600" />
              </div>
              <div class="space-y-3 py-3 text-[10.5px]">
                <div>
                  <p class="text-[9px] font-semibold tracking-wide text-ink-900/35 uppercase">From</p>
                  <p class="mt-0.5 font-medium">{{ form.sender_name || 'Sender name' }} <span class="font-normal text-ink-900/45">{{ form.sender_phone }}</span></p>
                  <p class="mt-0.5 break-words text-ink-900/50">{{ form.pickup_address || 'Pickup address' }}</p>
                </div>
                <div>
                  <p class="text-[9px] font-semibold tracking-wide text-ink-900/35 uppercase">To</p>
                  <p class="mt-0.5 font-medium">{{ form.receiver_name || 'Receiver name' }} <span class="font-normal text-ink-900/45">{{ form.receiver_phone }}</span></p>
                  <p class="mt-0.5 break-words text-ink-900/50">{{ form.dropoff_address || 'Drop-off address' }}</p>
                </div>
                <div class="border-t border-dashed border-ink-900/15 pt-3">
                  <p class="text-ink-900/45">{{ form.item_description || 'Item description' }}</p>
                  <div class="mt-2 flex justify-between gap-3"><span class="text-ink-900/45">Declared value</span><span class="font-medium">{{ formatMoney(formDeclaredValueMinor) }}</span></div>
                  <div class="mt-1 flex justify-between gap-3"><span class="text-ink-900/45">Delivery fee · {{ form.payer }}</span><span class="font-semibold">{{ formatMoney(formPriceMinor) }}</span></div>
                  <div class="mt-1 flex justify-between gap-3"><span class="text-ink-900/45">Payment</span><span class="font-medium">{{ form.is_paid ? `Paid · ${form.payment_method || 'method needed'}` : 'Unpaid' }}</span></div>
                  <div class="mt-1 flex justify-between gap-3"><span class="text-ink-900/45">Driver</span><span class="font-medium">{{ availableDrivers.find((driver) => driver.id === form.assigned_driver_id)?.full_name || 'Unassigned' }}</span></div>
                </div>
              </div>
              <p class="border-t border-dashed border-ink-900/15 pt-3 text-center text-[9px] text-ink-900/35">Delivery record · {{ currency }}</p>
            </section>
          </aside>
        </div>

        <footer class="delivery-modal-footer sticky bottom-0 flex flex-wrap justify-between gap-2 border-t px-5 py-3.5 sm:px-6">
          <p class="mr-auto hidden self-center text-[10px] text-ink-900/45 sm:block">A delivery reference is created when you save.</p>
          <button type="button" :disabled="saving" class="rounded-lg border border-ink-900/10 px-4 py-2.5 text-[11.5px] font-semibold text-ink-900/60 hover:bg-canvas disabled:opacity-50" @click="closeForm">
            Cancel
          </button>
          <button type="submit" :disabled="saving || !can('shipments.create')" class="inline-flex items-center gap-1.5 rounded-lg bg-brand-500 px-4 py-2.5 text-[11.5px] font-semibold text-white hover:bg-brand-600 disabled:opacity-50">
            <Loader2 v-if="saving" :size="14" class="animate-spin" />
            <Check v-else :size="14" />
            Save delivery
          </button>
        </footer>
      </form>
    </div>
  </div>
</template>
