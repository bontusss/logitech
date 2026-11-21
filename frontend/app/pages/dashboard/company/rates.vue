<script setup lang="ts">
import { parseProviderBusinessTypes } from '~/utils/providerBusinessTypes'
import {
  ArrowDownLeft,
  ArrowUpRight,
  Check,
  CircleAlert,
  Clock3,
  Loader2,
  MapPin,
  Plus,
  Route,
  Save,
  Sparkles,
  Truck,
  X,
  Zap,
  Copy,
  Trash2,
} from 'lucide-vue-next'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Rates & Coverage' })

interface BookingRoute {
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
  status: 'active' | 'inactive'
}

interface RouteForm {
  origin_state: string
  origin_city: string
  destination_state: string
  destination_city: string
  distance_km: string | number
  standard_price_per_km: string | number
  express_price_per_km: string | number
  same_day_price_per_km: string | number
  minimum_charge: string | number
}

const { api } = useApi()
const { detail, load, can, confirm } = useWorkspace()
const routes = ref<BookingRoute[]>([])
interface LastMileArea {
  id: string
  name: string
  state: string | null
  city: string | null
  coverage_status: 'draft' | 'active' | 'inactive'
  pricing_model: 'flat' | 'zone' | 'distance' | 'manual'
  flat_price_minor: number | null
  base_delivery_fee_minor: number | null
  price_per_km_minor: number | null
  maximum_distance_km: number
  express_available: boolean
}
const areas = ref<LastMileArea[]>([])
const areaModalOpen = ref(false)
const areaEditingId = ref('')
const areaForm = reactive({
  name: '', state: '', city: '', coverage_status: 'draft' as 'draft' | 'active' | 'inactive',
  pricing_model: 'flat' as 'flat' | 'zone' | 'distance' | 'manual',
  flat_price_minor: '' as string | number,
  base_delivery_fee_minor: '' as string | number,
  price_per_km_minor: '' as string | number,
  maximum_distance_km: '' as string | number,
  express_available: false, pickup_available: true, delivery_available: true,
  price_is_estimate: true,
})
const loading = ref(true)
const saving = ref(false)
const modalOpen = ref(false)
const error = ref('')
const search = ref('')
const editingRouteId = ref('')
const currency = computed(() => detail.value?.settings.currency || 'NGN')
const canCreate = computed(() => can('rates.create'))
const canUpdate = computed(() => can('rates.update'))

const emptyForm = (): RouteForm => ({
  origin_state: '',
  origin_city: '',
  destination_state: '',
  destination_city: '',
  distance_km: '',
  standard_price_per_km: '',
  express_price_per_km: '',
  same_day_price_per_km: '',
  minimum_charge: '',
})
const form = reactive<RouteForm>(emptyForm())

const filteredRoutes = computed(() => {
  const query = search.value.trim().toLocaleLowerCase()
  if (!query) return routes.value
  return routes.value.filter((route) =>
    [route.origin_state, route.origin_city, route.destination_state, route.destination_city]
      .some((value) => value.toLocaleLowerCase().includes(query)),
  )
})
const offeredTypes = computed(() => parseProviderBusinessTypes(detail.value?.company.business_type ?? ''))
const offersIntercity = computed(() => offeredTypes.value.includes('intercity'))
const offersLastMile = computed(() => offeredTypes.value.includes('lastmile'))
const activeTab = ref<'intercity' | 'lastmile'>('intercity')

watch(detail, (value) => {
  if (!value) return
  if (!offersIntercity.value && offersLastMile.value) activeTab.value = 'lastmile'
  if (offersIntercity.value && !offersLastMile.value) activeTab.value = 'intercity'
}, { immediate: true })
const activeRoutes = computed(() => routes.value.filter((route) => route.status === 'active').length)
const serviceOptions = computed(() =>
  routes.value.reduce(
    (count, route) =>
      count
      + Number(route.standard_price_per_km_minor !== null)
      + Number(route.express_price_per_km_minor !== null)
      + Number(route.same_day_price_per_km_minor !== null),
    0,
  ),
)

onMounted(async () => {
  await load()
  await Promise.all([fetchRoutes(), fetchAreas()])
  window.addEventListener('keydown', onKeydown)
})
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))

async function fetchAreas() {
  try {
    areas.value = (await api<{ areas: LastMileArea[] }>('/companies/me/last-mile-areas')).areas
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not load last-mile service areas.')
  }
}

async function fetchRoutes() {
  loading.value = true
  error.value = ''
  try {
    routes.value = (await api<{ routes: BookingRoute[] }>('/companies/me/booking-routes')).routes
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not load your routes and rates.')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingRouteId.value = ''
  Object.assign(form, emptyForm())
  error.value = ''
  modalOpen.value = true
}

function editRoute(route: BookingRoute) {
  editingRouteId.value = route.id
  Object.assign(form, {
    origin_state: route.origin_state,
    origin_city: route.origin_city,
    destination_state: route.destination_state,
    destination_city: route.destination_city,
    distance_km: String(route.distance_km),
    standard_price_per_km: route.standard_price_per_km_minor === null ? '' : String(route.standard_price_per_km_minor / 100),
    express_price_per_km: route.express_price_per_km_minor === null ? '' : String(route.express_price_per_km_minor / 100),
    same_day_price_per_km: route.same_day_price_per_km_minor === null ? '' : String(route.same_day_price_per_km_minor / 100),
    minimum_charge: route.minimum_charge_minor ? String(route.minimum_charge_minor / 100) : '',
  })
  error.value = ''
  modalOpen.value = true
}

function closeModal() {
  if (!saving.value) {
    modalOpen.value = false
    error.value = ''
  }
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') closeModal()
}

function isBlank(value: string | number) {
  return String(value).trim() === ''
}

function toPrice(value: string | number) {
  if (isBlank(value)) return null
  const amount = Number(value)
  return Number.isFinite(amount) && amount >= 0 ? amount : null
}

function pricePerKm(rateMinor: number | null) {
  if (rateMinor === null) return 'Not offered'
  return `${formatMoney(rateMinor)}/km`
}

function estimateForRoute(route: BookingRoute, rateMinor: number | null) {
  if (rateMinor === null) return null
  return Math.max(route.minimum_charge_minor, Math.ceil(rateMinor * route.distance_km))
}

function formatMoney(minor: number) {
  return new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency: currency.value,
    maximumFractionDigits: 2,
  }).format(minor / 100)
}

function requestBody() {
  return {
    origin_state: form.origin_state.trim(),
    origin_city: form.origin_city.trim(),
    destination_state: form.destination_state.trim(),
    destination_city: form.destination_city.trim(),
    distance_km: Number(form.distance_km),
    standard_price_per_km: toPrice(form.standard_price_per_km),
    express_price_per_km: toPrice(form.express_price_per_km),
    same_day_price_per_km: toPrice(form.same_day_price_per_km),
    minimum_charge: toPrice(form.minimum_charge) ?? 0,
  }
}

async function saveRoute() {
  error.value = ''
  const prices = [
    form.standard_price_per_km,
    form.express_price_per_km,
    form.same_day_price_per_km,
  ].filter((value) => !isBlank(value))
  if (
    !form.origin_state.trim()
    || !form.origin_city.trim()
    || !form.destination_state.trim()
    || !form.destination_city.trim()
    || !Number.isFinite(Number(form.distance_km))
    || Number(form.distance_km) <= 0
    || prices.length === 0
    || prices.some((value) => !Number.isFinite(Number(value)) || Number(value) <= 0)
    || (!isBlank(form.minimum_charge) && (!Number.isFinite(Number(form.minimum_charge)) || Number(form.minimum_charge) < 0))
  ) {
    error.value = 'Add both locations, a route distance, and at least one valid price per kilometre.'
    return
  }
  saving.value = true
  try {
    const body = requestBody()
    if (editingRouteId.value) {
      await api(`/companies/me/booking-routes/${editingRouteId.value}`, { method: 'PUT', body })
    } else {
      await api('/companies/me/booking-routes', { method: 'POST', body })
    }
    modalOpen.value = false
    await fetchRoutes()
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not save this route.')
  } finally {
    saving.value = false
  }
}

function openAreaCreate() {
  areaEditingId.value = ''
  Object.assign(areaForm, { name: '', state: '', city: '', coverage_status: 'draft', pricing_model: 'flat', flat_price_minor: '', base_delivery_fee_minor: '', price_per_km_minor: '', maximum_distance_km: '', express_available: false, pickup_available: true, delivery_available: true, price_is_estimate: true })
  areaModalOpen.value = true
}

function editArea(area: LastMileArea) {
  areaEditingId.value = area.id
  Object.assign(areaForm, { name: area.name, state: area.state ?? '', city: area.city ?? '', coverage_status: area.coverage_status, pricing_model: area.pricing_model, flat_price_minor: area.flat_price_minor ? String(area.flat_price_minor / 100) : '', base_delivery_fee_minor: area.base_delivery_fee_minor ? String(area.base_delivery_fee_minor / 100) : '', price_per_km_minor: area.price_per_km_minor ? String(area.price_per_km_minor / 100) : '', maximum_distance_km: area.maximum_distance_km ? String(area.maximum_distance_km) : '', express_available: area.express_available, price_is_estimate: true })
  areaModalOpen.value = true
}

async function saveArea() {
  if (!areaForm.name.trim()) { error.value = 'Enter a service area name.'; return }
  if (areaForm.pricing_model === 'flat' && !areaForm.flat_price_minor) { error.value = 'Enter a flat delivery price.'; return }
  if (areaForm.pricing_model === 'distance' && (!areaForm.base_delivery_fee_minor || !areaForm.price_per_km_minor)) { error.value = 'Enter a base fee and price per kilometre.'; return }
  saving.value = true
  try {
    const money = (value: string | number) => String(value).trim() ? Math.round(Number(value) * 100) : null
    const body = {
      name: areaForm.name.trim(), state: areaForm.state.trim() || null, city: areaForm.city.trim() || null,
      coverage_status: areaForm.coverage_status, pricing_model: areaForm.pricing_model,
      flat_price_minor: areaForm.pricing_model === 'flat' ? money(areaForm.flat_price_minor) : null,
      base_delivery_fee_minor: areaForm.pricing_model === 'distance' ? money(areaForm.base_delivery_fee_minor) : null,
      price_per_km_minor: areaForm.pricing_model === 'distance' ? money(areaForm.price_per_km_minor) : null,
      maximum_distance_km: areaForm.maximum_distance_km ? Number(areaForm.maximum_distance_km) : null,
      express_available: areaForm.express_available, pickup_available: areaForm.pickup_available,
      delivery_available: areaForm.delivery_available, price_is_estimate: areaForm.price_is_estimate,
    }
    if (areaEditingId.value) await api(`/companies/me/last-mile-areas/${areaEditingId.value}`, { method: 'PUT', body })
    else await api('/companies/me/last-mile-areas', { method: 'POST', body })
    areaModalOpen.value = false
    await fetchAreas()
  } catch (cause) { error.value = apiErrorMessage(cause, 'Could not save the service area.') }
  finally { saving.value = false }
}

async function deleteArea(area: LastMileArea) {
  const approved = await confirm({ title: 'Delete service area?', message: `${area.name} and its zones will be permanently deleted.`, confirmLabel: 'Delete area', danger: true })
  if (!approved) return
  try { await api(`/companies/me/last-mile-areas/${area.id}`, { method: 'DELETE' }); await fetchAreas() }
  catch (cause) { error.value = apiErrorMessage(cause, 'Could not delete this service area.') }
}

async function duplicateRoute(route: BookingRoute) {
  error.value = ''
  try {
    await api(`/companies/me/booking-routes/${route.id}/duplicate`, { method: 'POST' })
    await fetchRoutes()
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not duplicate this route.')
  }
}

async function deleteRoute(route: BookingRoute) {
  const approved = await confirm({
    title: 'Delete route?',
    message: `${route.origin_city} to ${route.destination_city} will be permanently deleted.`,
    confirmLabel: 'Delete route',
    danger: true,
  })
  if (!approved) return
  try {
    await api(`/companies/me/booking-routes/${route.id}`, { method: 'DELETE' })
    await fetchRoutes()
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not delete this route.')
  }
}

async function toggleStatus(route: BookingRoute) {
  error.value = ''
  try {
    await api(`/companies/me/booking-routes/${route.id}/status`, {
      method: 'PATCH',
      body: { status: route.status === 'active' ? 'inactive' : 'active' },
    })
    await fetchRoutes()
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not update this route.')
  }
}
</script>

<template>
  <div class="rates-workspace dash-page">
    <header class="rates-page-hero relative overflow-hidden rounded-2xl border p-5 sm:p-6">
      <div class="pointer-events-none absolute -top-20 right-12 size-48 rounded-full bg-fuchsia-400/15 blur-3xl" />
      <div class="pointer-events-none absolute -right-10 -bottom-28 size-56 rounded-full bg-cyan-400/15 blur-3xl" />
      <div class="relative flex flex-wrap items-center justify-between gap-4">
        <div>
          <p class="inline-flex items-center gap-1.5 rounded-full bg-white/75 px-2.5 py-1 text-[10px] font-semibold tracking-wide text-brand-700 dark:bg-slate-950/30 dark:text-indigo-200">
            <Sparkles :size="12" /> YOUR DELIVERY NETWORK
          </p>
          <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Rates & Coverage</h1>
          <p class="mt-1.5 max-w-xl text-[12px] leading-relaxed text-ink-900/65">
            Set the lanes you serve and what each delivery speed costs per kilometre.
          </p>
        </div>
        <button
          v-if="canCreate && activeTab === 'intercity'"
          type="button"
          class="inline-flex min-h-10 items-center gap-1.5 rounded-xl bg-brand-600 px-4 py-2.5 text-[12px] font-semibold text-white shadow-sm shadow-brand-600/20 transition hover:-translate-y-0.5 hover:bg-brand-700"
          @click="openCreate"
        >
          <Plus :size="14" /> Add route
        </button>
      </div>
    </header>

    <div v-if="offersIntercity && offersLastMile" class="mt-5 inline-flex rounded-xl border border-ink-900/10 bg-white p-1">
      <button type="button" class="rounded-lg px-4 py-2 text-[12px] font-semibold" :class="activeTab === 'intercity' ? 'bg-brand-600 text-white' : 'text-ink-900/55'" @click="activeTab = 'intercity'">Inter-city</button>
      <button type="button" class="rounded-lg px-4 py-2 text-[12px] font-semibold" :class="activeTab === 'lastmile' ? 'bg-brand-600 text-white' : 'text-ink-900/55'" @click="activeTab = 'lastmile'">Last-mile</button>
    </div>

    <p v-if="error && !modalOpen && !areaModalOpen" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <ItemCategoryPolicy />

    <section v-if="activeTab === 'intercity'" class="mt-5 grid gap-3 sm:grid-cols-3" aria-label="Rate overview">
      <article class="rates-stat rates-stat--blue rounded-2xl border p-4">
        <span class="flex size-8 items-center justify-center rounded-xl bg-white/70 text-blue-700 dark:bg-white/10 dark:text-blue-200"><Route :size="16" /></span>
        <p class="mt-3 text-[10px] font-semibold text-ink-900/50">Total routes</p>
        <p class="mt-0.5 text-[22px] leading-none font-semibold tracking-tight">{{ routes.length }}</p>
      </article>
      <article class="rates-stat rates-stat--mint rounded-2xl border p-4">
        <span class="flex size-8 items-center justify-center rounded-xl bg-white/70 text-emerald-700 dark:bg-white/10 dark:text-emerald-200"><MapPin :size="16" /></span>
        <p class="mt-3 text-[10px] font-semibold text-ink-900/50">Active lanes</p>
        <p class="mt-0.5 text-[22px] leading-none font-semibold tracking-tight">{{ activeRoutes }}</p>
      </article>
      <article class="rates-stat rates-stat--peach rounded-2xl border p-4">
        <span class="flex size-8 items-center justify-center rounded-xl bg-white/70 text-orange-700 dark:bg-white/10 dark:text-orange-200"><Zap :size="16" /></span>
        <p class="mt-3 text-[10px] font-semibold text-ink-900/50">Available service options</p>
        <p class="mt-0.5 text-[22px] leading-none font-semibold tracking-tight">{{ serviceOptions }}</p>
      </article>
    </section>

    <section v-if="activeTab === 'intercity'" class="mt-5 overflow-hidden rounded-2xl border border-ink-900/10 bg-white">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-ink-900/8 px-4 py-3.5 sm:px-5">
        <div>
          <h2 class="text-[13px] font-semibold">Your routes</h2>
          <p class="mt-0.5 text-[10.5px] text-ink-900/45">Customers see active routes and available prices on your booking page.</p>
        </div>
        <label class="relative w-full sm:max-w-56">
          <MapPin :size="13" class="absolute top-2.5 left-2.5 text-ink-900/35" />
          <input v-model="search" type="search" placeholder="Search a city or state" class="w-full rounded-lg border border-ink-900/10 bg-canvas py-2 pl-8 pr-3 text-[10.5px] outline-none focus:border-brand-500" />
        </label>
      </div>

      <div v-if="loading" class="flex items-center justify-center gap-2 p-12 text-[12px] text-ink-900/45">
        <Loader2 :size="15" class="animate-spin" /> Loading routes…
      </div>
      <div v-else-if="!filteredRoutes.length" class="px-5 py-12 text-center">
        <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-brand-50 text-brand-700"><Route :size="19" /></span>
        <h3 class="mt-3 text-[13px] font-semibold">{{ routes.length ? 'No routes match that search' : 'Start with your first route' }}</h3>
        <p class="mx-auto mt-1.5 max-w-sm text-[11px] leading-5 text-ink-900/45">
          {{ routes.length ? 'Try another city or state name.' : 'Add the states and cities you connect, then publish prices per kilometre for each delivery speed.' }}
        </p>
        <button v-if="canCreate && !routes.length" type="button" class="mt-4 inline-flex items-center gap-1.5 rounded-lg bg-brand-600 px-3.5 py-2.5 text-[11px] font-semibold text-white" @click="openCreate">
          <Plus :size="13" /> Add your first route
        </button>
      </div>
      <div v-else class="divide-y divide-ink-900/8">
        <article v-for="route in filteredRoutes" :key="route.id" class="p-4 sm:p-5">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="flex min-w-0 items-center gap-3">
              <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-cyan-50 text-cyan-700 dark:bg-cyan-400/10 dark:text-cyan-200"><Route :size="16" /></span>
              <div class="min-w-0">
                <h3 class="text-[12px] font-semibold">
                  {{ route.origin_city }}, {{ route.origin_state }}
                  <ArrowRight class="mx-1 inline text-ink-900/35" :size="13" />
                  {{ route.destination_city }}, {{ route.destination_state }}
                </h3>
                <p class="mt-1 inline-flex items-center gap-1 text-[10px] text-ink-900/45"><MapPin :size="11" /> {{ route.distance_km.toLocaleString() }} km route distance</p>
              </div>
            </div>
            <span class="rounded-full px-2.5 py-1 text-[9px] font-semibold" :class="route.status === 'active' ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-400/10 dark:text-emerald-200' : 'bg-slate-100 text-slate-500'">
              {{ route.status === 'active' ? 'Live on booking page' : 'Hidden from customers' }}
            </span>
          </div>
          <div class="mt-4 grid gap-2 sm:grid-cols-3">
            <div class="rates-price-card rates-price-card--standard rounded-xl border p-3">
              <p class="inline-flex items-center gap-1 text-[9.5px] font-semibold text-ink-900/50"><Truck :size="12" /> Standard</p>
              <p class="mt-1.5 text-[12px] font-bold">{{ pricePerKm(route.standard_price_per_km_minor) }}</p>
              <p v-if="estimateForRoute(route, route.standard_price_per_km_minor) !== null" class="mt-1 text-[9px] text-ink-900/40">
                ≈ {{ formatMoney(estimateForRoute(route, route.standard_price_per_km_minor)!) }} for this lane
              </p>
            </div>
            <div class="rates-price-card rates-price-card--express rounded-xl border p-3">
              <p class="inline-flex items-center gap-1 text-[9.5px] font-semibold text-ink-900/50"><Zap :size="12" /> Express</p>
              <p class="mt-1.5 text-[12px] font-bold">{{ pricePerKm(route.express_price_per_km_minor) }}</p>
              <p v-if="estimateForRoute(route, route.express_price_per_km_minor) !== null" class="mt-1 text-[9px] text-ink-900/40">
                ≈ {{ formatMoney(estimateForRoute(route, route.express_price_per_km_minor)!) }} for this lane
              </p>
            </div>
            <div class="rates-price-card rates-price-card--same-day rounded-xl border p-3">
              <p class="inline-flex items-center gap-1 text-[9.5px] font-semibold text-ink-900/50"><Clock3 :size="12" /> Same-day</p>
              <p class="mt-1.5 text-[12px] font-bold">{{ pricePerKm(route.same_day_price_per_km_minor) }}</p>
              <p v-if="estimateForRoute(route, route.same_day_price_per_km_minor) !== null" class="mt-1 text-[9px] text-ink-900/40">
                ≈ {{ formatMoney(estimateForRoute(route, route.same_day_price_per_km_minor)!) }} for this lane
              </p>
            </div>
          </div>
          <div v-if="route.minimum_charge_minor" class="mt-2 text-[9.5px] text-ink-900/45">
            Minimum charge {{ formatMoney(route.minimum_charge_minor) }} · {{ currency }} {{ route.distance_km }} km × price/km
          </div>
          <div v-if="canUpdate || canCreate" class="mt-3 flex justify-end gap-2">
            <button v-if="canCreate" type="button" class="rounded-lg border border-ink-900/10 px-3 py-1.5 text-[10px] font-semibold text-ink-900/65 hover:bg-canvas" @click="duplicateRoute(route)"><Copy :size="11" class="mr-1 inline" />Duplicate</button>
            <button v-if="canUpdate" type="button" class="rounded-lg border border-ink-900/10 px-3 py-1.5 text-[10px] font-semibold text-ink-900/65 hover:bg-canvas" @click="editRoute(route)">Edit route</button>
            <button v-if="canUpdate" type="button" class="rounded-lg px-3 py-1.5 text-[10px] font-semibold text-rose-700 hover:bg-rose-50" @click="deleteRoute(route)"><Trash2 :size="11" class="mr-1 inline" />Delete</button>
            <button v-if="canUpdate" type="button" class="rounded-lg px-3 py-1.5 text-[10px] font-semibold" :class="route.status === 'active' ? 'text-rose-700 hover:bg-rose-50' : 'text-emerald-700 hover:bg-emerald-50'" @click="toggleStatus(route)">
              {{ route.status === 'active' ? 'Pause route' : 'Publish route' }}
            </button>
          </div>
        </article>
      </div>
    </section>

    <section v-if="activeTab === 'lastmile'" class="mt-5 overflow-hidden rounded-2xl border border-ink-900/10 bg-white">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-ink-900/8 px-4 py-3.5 sm:px-5">
        <div><h2 class="text-[13px] font-semibold">Last-mile service areas</h2><p class="mt-0.5 text-[10.5px] text-ink-900/45">Configure coverage zones and local delivery prices.</p></div>
        <button v-if="canCreate" type="button" class="inline-flex items-center gap-1.5 rounded-lg bg-brand-600 px-3.5 py-2.5 text-[11.5px] font-semibold text-white" @click="openAreaCreate"><Plus :size="13" /> Add service area</button>
      </div>
      <div v-if="!areas.length" class="px-5 py-12 text-center"><MapPin class="mx-auto text-brand-600" :size="20" /><h3 class="mt-3 text-[13px] font-semibold">No last-mile coverage configured</h3><p class="mt-1 text-[11px] text-ink-900/45">Create a service area, add zones and choose a pricing model.</p></div>
      <div v-else class="divide-y divide-ink-900/8">
        <article v-for="area in areas" :key="area.id" class="p-4 sm:p-5">
          <div class="flex flex-wrap items-start justify-between gap-3"><div><h3 class="text-[12px] font-semibold">{{ area.name }}</h3><p class="mt-1 text-[10.5px] text-ink-900/45">{{ [area.city, area.state].filter(Boolean).join(', ') || 'Custom coverage area' }}</p></div><span class="rounded-full px-2.5 py-1 text-[9px] font-semibold" :class="area.coverage_status === 'active' ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'">{{ area.coverage_status }}</span></div>
          <p class="mt-3 text-[11px]"><span class="text-ink-900/45">Pricing:</span> <strong class="capitalize">{{ area.pricing_model }}</strong><span v-if="area.flat_price_minor" class="ml-2">{{ formatMoney(area.flat_price_minor) }} flat</span></p>
          <div v-if="canUpdate || canCreate" class="mt-3 flex justify-end gap-2"><button v-if="canUpdate" type="button" class="rounded-lg border border-ink-900/10 px-3 py-1.5 text-[10px] font-semibold" @click="editArea(area)">Edit area</button><button v-if="canUpdate" type="button" class="rounded-lg px-3 py-1.5 text-[10px] font-semibold text-rose-700 hover:bg-rose-50" @click="deleteArea(area)">Delete</button></div>
        </article>
      </div>
    </section>

    <div v-if="areaModalOpen" class="fixed inset-0 z-[80] flex items-center justify-center bg-ink-950/55 px-3 py-4 backdrop-blur-md" role="dialog" aria-modal="true" @click.self="areaModalOpen = false">
      <form class="max-h-full w-full max-w-xl overflow-y-auto rounded-2xl border bg-white shadow-2xl" @submit.prevent="saveArea">
        <header class="flex items-start justify-between border-b px-5 py-4"><div><p class="text-[9.5px] font-semibold tracking-[0.16em] text-brand-700 uppercase">LAST-MILE COVERAGE</p><h2 class="mt-1 text-[15px] font-semibold">{{ areaEditingId ? 'Edit service area' : 'Add service area' }}</h2></div><button type="button" class="p-1.5" aria-label="Close" @click="areaModalOpen=false"><X :size="18" /></button></header>
        <p v-if="error" role="alert" class="mx-5 mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[12px] text-red-700">{{ error }}</p>
        <div class="grid gap-3 p-5 sm:grid-cols-2"><label class="block sm:col-span-2"><span class="mb-1.5 block text-[10.5px] font-medium">Service area name</span><input v-model="areaForm.name" required class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px]" placeholder="e.g. Lagos mainland" /></label><label class="block"><span class="mb-1.5 block text-[10.5px] font-medium">State</span><input v-model="areaForm.state" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px]" /></label><label class="block"><span class="mb-1.5 block text-[10.5px] font-medium">City</span><input v-model="areaForm.city" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px]" /></label><label class="block sm:col-span-2"><span class="mb-1.5 block text-[10.5px] font-medium">Pricing model</span><select v-model="areaForm.pricing_model" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px]"><option value="flat">Flat pricing</option><option value="zone">Zone-based pricing</option><option value="distance">Distance-based pricing</option><option value="manual">Manual quotation</option></select></label><label v-if="areaForm.pricing_model === 'flat'" class="block sm:col-span-2"><span class="mb-1.5 block text-[10.5px] font-medium">Flat price</span><input v-model="areaForm.flat_price_minor" type="number" min="0.01" step="0.01" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px]" /></label><template v-if="areaForm.pricing_model === 'distance'"><label class="block"><span class="mb-1.5 block text-[10.5px] font-medium">Base fee</span><input v-model="areaForm.base_delivery_fee_minor" type="number" min="0.01" step="0.01" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px]" /></label><label class="block"><span class="mb-1.5 block text-[10.5px] font-medium">Price per km</span><input v-model="areaForm.price_per_km_minor" type="number" min="0.01" step="0.01" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px]" /></label><label class="block sm:col-span-2"><span class="mb-1.5 block text-[10.5px] font-medium">Maximum distance (optional)</span><input v-model="areaForm.maximum_distance_km" type="number" min="0.1" step="0.1" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px]" /></label></template><label class="block"><span class="mb-1.5 block text-[10.5px] font-medium">Coverage status</span><select v-model="areaForm.coverage_status" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px]"><option value="draft">Draft</option><option value="active">Active</option><option value="inactive">Inactive</option></select></label><label class="flex items-center gap-2 text-[11px]"><input v-model="areaForm.express_available" type="checkbox" /> Express delivery available</label></div>
        <footer class="flex justify-end gap-2 border-t px-5 py-4"><button type="button" class="rounded-lg border px-3 py-2 text-[11px] font-semibold" @click="areaModalOpen=false">Cancel</button><button type="submit" :disabled="saving" class="rounded-lg bg-brand-600 px-4 py-2 text-[11px] font-semibold text-white"><Loader2 v-if="saving" :size="13" class="inline animate-spin" /> Save service area</button></footer>
      </form>
    </div>

    <div v-if="modalOpen" class="fixed inset-0 z-[80] flex items-center justify-center bg-ink-950/55 px-3 py-4 backdrop-blur-md sm:px-5 sm:py-6" role="dialog" aria-modal="true" aria-labelledby="route-modal-title" @click.self="closeModal">
      <form class="rates-modal max-h-full w-full max-w-2xl overflow-y-auto rounded-2xl border shadow-2xl" @submit.prevent="saveRoute">
        <header class="rates-modal-heading sticky top-0 z-10 flex items-start justify-between gap-4 border-b px-5 py-4 sm:px-6">
          <div class="flex items-start gap-3">
            <span class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-cyan-100 text-cyan-800 dark:bg-cyan-400/15 dark:text-cyan-200"><Route :size="18" /></span>
            <div>
              <p class="text-[9.5px] font-semibold tracking-[0.16em] text-brand-700 uppercase dark:text-indigo-200">ROUTE PRICING</p>
              <h2 id="route-modal-title" class="mt-1 text-[15px] font-semibold">{{ editingRouteId ? 'Tune this delivery lane' : 'Add a delivery lane' }}</h2>
              <p class="mt-1 text-[11px] text-ink-900/50">Set the distance and per-kilometre price for each service level.</p>
            </div>
          </div>
          <button type="button" class="rounded-lg p-1.5 text-ink-900/45 hover:bg-ink-900/5" aria-label="Close dialog" :disabled="saving" @click="closeModal"><X :size="18" /></button>
        </header>

        <p v-if="error" role="alert" class="mx-5 mt-4 rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700 sm:mx-6">{{ error }}</p>
        <div class="space-y-4 p-4 sm:p-5">
          <section class="rates-form-section rates-form-section--lane rounded-xl border p-4">
            <h3 class="inline-flex items-center gap-2 text-[12px] font-semibold"><MapPin :size="14" class="text-cyan-700 dark:text-cyan-200" /> Route lane</h3>
            <div class="mt-3 grid gap-3 sm:grid-cols-2">
              <label class="block"><span class="mb-1.5 block text-[10.5px] font-medium text-ink-900/55">From state</span><input v-model="form.origin_state" required maxlength="100" placeholder="e.g. Abia" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px] focus:border-brand-500" /></label>
              <label class="block"><span class="mb-1.5 block text-[10.5px] font-medium text-ink-900/55">From city</span><input v-model="form.origin_city" required maxlength="100" placeholder="e.g. Aba" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px] focus:border-brand-500" /></label>
              <label class="block"><span class="mb-1.5 block text-[10.5px] font-medium text-ink-900/55">To state</span><input v-model="form.destination_state" required maxlength="100" placeholder="e.g. Lagos" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px] focus:border-brand-500" /></label>
              <label class="block"><span class="mb-1.5 block text-[10.5px] font-medium text-ink-900/55">To city</span><input v-model="form.destination_city" required maxlength="100" placeholder="e.g. Ikeja" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px] focus:border-brand-500" /></label>
              <label class="block sm:col-span-2"><span class="mb-1.5 block text-[10.5px] font-medium text-ink-900/55">Route distance</span><div class="relative"><input v-model="form.distance_km" required type="number" min="0.1" max="10000" step="0.1" placeholder="e.g. 580" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 pr-12 text-[12px] focus:border-brand-500" /><span class="absolute top-2.5 right-3 text-[11px] text-ink-900/40">km</span></div><span class="mt-1 block text-[9.5px] text-ink-900/40">Use your usual road distance for this city-to-city lane.</span></label>
            </div>
          </section>

          <section class="rates-form-section rates-form-section--prices rounded-xl border p-4">
            <h3 class="inline-flex items-center gap-2 text-[12px] font-semibold"><Zap :size="14" class="text-amber-700 dark:text-amber-200" /> Service prices <span class="font-normal text-ink-900/45">· {{ currency }} per km</span></h3>
            <div class="mt-3 grid gap-3 sm:grid-cols-3">
              <label class="block"><span class="mb-1.5 flex items-center gap-1 text-[10px] font-semibold text-ink-900/60"><Truck :size="12" /> Standard</span><div class="relative"><input v-model="form.standard_price_per_km" type="number" min="0.01" step="0.01" placeholder="Optional" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 pr-12 text-[11px] focus:border-brand-500" /><span class="absolute top-2.5 right-3 text-[10px] text-ink-900/40">/km</span></div></label>
              <label class="block"><span class="mb-1.5 flex items-center gap-1 text-[10px] font-semibold text-ink-900/60"><Zap :size="12" /> Express</span><div class="relative"><input v-model="form.express_price_per_km" type="number" min="0.01" step="0.01" placeholder="Optional" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 pr-12 text-[11px] focus:border-brand-500" /><span class="absolute top-2.5 right-3 text-[10px] text-ink-900/40">/km</span></div></label>
              <label class="block"><span class="mb-1.5 flex items-center gap-1 text-[10px] font-semibold text-ink-900/60"><Clock3 :size="12" /> Same-day</span><div class="relative"><input v-model="form.same_day_price_per_km" type="number" min="0.01" step="0.01" placeholder="Optional" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 pr-12 text-[11px] focus:border-brand-500" /><span class="absolute top-2.5 right-3 text-[10px] text-ink-900/40">/km</span></div></label>
            </div>
            <label class="mt-3 block max-w-xs"><span class="mb-1.5 block text-[10px] font-semibold text-ink-900/60">Minimum delivery charge <span class="font-normal text-ink-900/40">· optional</span></span><div class="relative"><span class="absolute top-2.5 left-3 text-[10px] text-ink-900/40">{{ currency }}</span><input v-model="form.minimum_charge" type="number" min="0" step="0.01" placeholder="0.00" class="w-full rounded-lg border border-slate-300 py-2.5 pr-3 pl-14 text-[11px] focus:border-brand-500" /></div></label>
            <p class="mt-3 rounded-lg bg-white/70 px-3 py-2.5 text-[9.5px] leading-5 text-ink-900/50 dark:bg-slate-900/35">
              Customer estimate = route km × price/km. If the total is below the minimum charge, the minimum is used. Leave a service blank if you don’t offer it on this lane.
            </p>
          </section>
        </div>
        <footer class="rates-modal-footer sticky bottom-0 flex flex-wrap items-center justify-end gap-2 border-t px-4 py-3.5 sm:px-5">
          <button type="button" class="rounded-lg border border-ink-900/15 px-3.5 py-2.5 text-[11px] font-semibold text-ink-900/65 hover:bg-canvas" :disabled="saving" @click="closeModal">Cancel</button>
          <button type="submit" class="inline-flex items-center gap-1.5 rounded-lg bg-brand-600 px-4 py-2.5 text-[11px] font-semibold text-white hover:bg-brand-700 disabled:opacity-70" :disabled="saving">
            <Loader2 v-if="saving" :size="13" class="animate-spin" /><Save v-else :size="13" /> {{ saving ? 'Saving…' : editingRouteId ? 'Save changes' : 'Add route' }}
          </button>
        </footer>
      </form>
    </div>
  </div>
</template>
