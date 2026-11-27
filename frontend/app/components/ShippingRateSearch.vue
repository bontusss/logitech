<script setup lang="ts">
import { ArrowRight, Loader2, PackageCheck, Search, Truck, Warehouse } from 'lucide-vue-next'
import type { MarketplaceSearchPayload } from '~/types/marketplace'
import type { CustomerSavedAddress } from '~/types/customer'
import type { SearchGeoZone, SearchLocationValue } from '~/components/SearchLocationPicker.vue'
import { ITEM_CATEGORIES } from '~/utils/itemCategories'

const props = withDefaults(
  defineProps<{
    loading?: boolean
    savedAddresses?: CustomerSavedAddress[]
    prefill?: MarketplaceSearchPayload | null
    heading?: string
  }>(),
  { loading: false, savedAddresses: () => [], prefill: null, heading: '' },
)

const emit = defineEmits<{ search: [payload: MarketplaceSearchPayload] }>()

type ServiceMode = 'pickup' | 'dropoff'

const { api } = useApi()
const zones = ref<SearchGeoZone[]>([])
const reachableZones = ref<SearchGeoZone[]>([])
const zonesLoading = ref(true)
const reachableLoading = ref(false)
const zoneError = ref('')
const error = ref('')
const pendingDestination = ref<SearchLocationValue | null>(null)
const pendingPrefill = ref<MarketplaceSearchPayload | null>(props.prefill)
const savedNotice = ref('')

const savedPickupAddresses = computed(() =>
  [...props.savedAddresses].sort(
    (a, b) => Number(b.default_pickup) - Number(a.default_pickup) || a.label.localeCompare(b.label),
  ),
)
const savedDeliveryAddresses = computed(() =>
  [...props.savedAddresses].sort(
    (a, b) => Number(b.default_delivery) - Number(a.default_delivery) || a.label.localeCompare(b.label),
  ),
)

function blankLocation(): SearchLocationValue {
  return { state: '', lgaKey: '', zoneId: '', area: '', address: '', latitude: null, longitude: null }
}

function zoneLgaKey(zone: SearchGeoZone) {
  return `${zone.state}|${zone.lga}`
}

function assignZone(location: SearchLocationValue, zone: SearchGeoZone) {
  location.state = zone.state
  location.lgaKey = zoneLgaKey(zone)
  location.zoneId = zone.id
  location.area = zone.parent_id ? zone.name : ''
  resetPlace(location)
}

const form = reactive({
  mode: 'pickup' as ServiceMode,
  origin: blankLocation(),
  destination: blankLocation(),
  weight: 2,
  length: 20,
  width: 15,
  height: 10,
  category: 'other',
})

const categories = ITEM_CATEGORIES.map((category) => ({
  value: category.value,
  label: category.label,
}))
const dimensions = [
  { key: 'length', label: 'Length' },
  { key: 'width', label: 'Width' },
  { key: 'height', label: 'Height' },
] as const

const fieldClass =
  'w-full rounded-xl border border-ink-900/12 bg-canvas px-3.5 py-3 text-[13px] font-medium text-ink-900 transition placeholder:text-ink-900/35 hover:border-ink-900/25 focus:border-accent-500 focus:bg-surface'
const labelClass = 'mb-1.5 block text-[11px] font-semibold text-ink-900/55'

const destinationZones = computed(() => reachableZones.value)
const destinationLoading = computed(() => reachableLoading.value)

function resetPlace(location: SearchLocationValue) {
  location.address = ''
  location.latitude = null
  location.longitude = null
}

function locationLabel(location: SearchLocationValue, list: SearchGeoZone[]) {
  const zone = list.find((item) => item.id === location.zoneId)
  const parts = [location.area.trim(), zone?.lga, zone?.state].filter(Boolean)
  return parts.join(', ')
}

async function assignLocation(location: SearchLocationValue, value: SearchLocationValue) {
  Object.assign(location, {
    state: value.state,
    lgaKey: value.lgaKey,
    zoneId: value.zoneId,
    area: value.area,
    address: '',
    latitude: null,
    longitude: null,
  })
  await nextTick()
  location.address = value.address
  location.latitude = value.latitude
  location.longitude = value.longitude
}

function locationForZone(zone: SearchGeoZone, address = ''): SearchLocationValue {
  return {
    state: zone.state,
    lgaKey: zoneLgaKey(zone),
    zoneId: zone.id,
    area: zone.parent_id ? zone.name : '',
    address,
    latitude: null,
    longitude: null,
  }
}

function savedAddressLocation(address: CustomerSavedAddress, list: SearchGeoZone[]) {
  const state = address.state.trim().toLowerCase()
  const city = address.city.trim().toLowerCase()
  const area = address.area.trim().toLowerCase()
  let matches = list.filter(
    (zone) => zone.state.toLowerCase() === state && zone.lga.toLowerCase() === city,
  )
  if (!matches.length) {
    // Saved addresses use the customer's own city wording, which may be a
    // shorter or longer form of the zone LGA ("Abuja" vs "Abuja Municipal").
    matches = list.filter(
      (zone) =>
        zone.state.toLowerCase() === state &&
        (zone.lga.toLowerCase().includes(city) || city.includes(zone.lga.toLowerCase())),
    )
  }
  if (!matches.length) return null
  const areaZone = area ? matches.find((zone) => zone.name.toLowerCase() === area) : undefined
  const zone = areaZone ?? matches.find((item) => !item.parent_id) ?? matches[0]!
  const location = locationForZone(zone)
  location.area = areaZone?.name ?? ''
  location.zoneId = areaZone?.id ?? zone.id
  const street = [address.street_address, address.area, address.city, address.state]
    .filter(Boolean)
    .join(', ')
  location.address = street
  location.latitude = address.latitude
  location.longitude = address.longitude
  return location
}

async function useSavedAddress(event: Event, role: 'origin' | 'destination') {
  const id = (event.target as HTMLSelectElement).value
  ;(event.target as HTMLSelectElement).value = ''
  if (!id) return
  const address = props.savedAddresses.find((item) => item.id === id)
  if (!address) return
  error.value = ''
  savedNotice.value = ''
  const list = role === 'origin' ? zones.value : destinationZones.value
  const location = savedAddressLocation(address, list)
  if (!location) {
    error.value = role === 'origin'
      ? `We could not match ${address.label} to a service zone yet. Pick the zone manually below.`
      : `${address.label} is not reachable from this pickup zone. Pick another saved address or zone.`
    return
  }
  if (location.latitude === null || location.longitude === null) {
    savedNotice.value = `${address.label} has no map pin saved yet. Confirm the exact spot below so providers can collect it.`
  }
  if (role === 'origin') {
    await assignLocation(form.origin, location)
    return
  }
  if (!list.some((zone) => zone.id === location.zoneId)) {
    if (reachableLoading.value) {
      pendingDestination.value = location
      return
    }
    error.value = `${address.label} is not reachable from this pickup zone. Pick another saved address or zone.`
    return
  }
  await assignLocation(form.destination, location)
}

async function applyPrefill(payload: MarketplaceSearchPayload) {
  if (!zones.value.length) {
    pendingPrefill.value = payload
    return
  }
  pendingPrefill.value = null
  form.mode = payload.service_mode === 'dropoff' ? 'dropoff' : 'pickup'
  const originZone = payload.origin_zone_id
    ? zones.value.find((zone) => zone.id === payload.origin_zone_id)
    : undefined
  const originFallback = !originZone && payload.origin_lga
    ? zones.value.find(
        (zone) =>
          zone.lga.toLowerCase() === payload.origin_lga!.toLowerCase() &&
          zone.state.toLowerCase() === (payload.origin_state ?? '').toLowerCase(),
      )
    : undefined
  const resolvedOrigin = originZone ?? originFallback
  if (resolvedOrigin) {
    const originLocation = locationForZone(resolvedOrigin, payload.origin_address ?? '')
    originLocation.latitude = payload.origin_latitude ?? null
    originLocation.longitude = payload.origin_longitude ?? null
    await assignLocation(form.origin, originLocation)
    await loadReachable()
  }
  const destinationZone = payload.destination_zone_id
    ? destinationZones.value.find((zone) => zone.id === payload.destination_zone_id)
    : undefined
  if (destinationZone) {
    const destinationLocation = locationForZone(
      destinationZone,
      payload.destination_address ?? '',
    )
    destinationLocation.latitude = payload.destination_latitude ?? null
    destinationLocation.longitude = payload.destination_longitude ?? null
    await assignLocation(form.destination, destinationLocation)
  }
  const weight = Number(payload.weight_kg)
  if (Number.isFinite(weight) && weight > 0) form.weight = weight
  if (payload.category) form.category = payload.category
  if (payload.length_cm && payload.width_cm && payload.height_cm) {
    form.length = payload.length_cm
    form.width = payload.width_cm
    form.height = payload.height_cm
  }
}

function ensureDestination() {
  const list = destinationZones.value
  const pending = pendingDestination.value
  if (pending) {
    if (list.some((zone) => zone.id === pending.zoneId)) {
      pendingDestination.value = null
      void assignLocation(form.destination, pending)
      return
    }
    if (reachableLoading.value) return
    pendingDestination.value = null
  }
  const current = list.find((zone) => zone.id === form.destination.zoneId)
  if (current) {
    form.destination.state = current.state
    if (!form.destination.lgaKey) form.destination.lgaKey = zoneLgaKey(current)
    return
  }
  const first = list[0]
  if (first) {
    assignZone(form.destination, first)
  } else {
    form.destination.zoneId = ''
    form.destination.state = ''
    form.destination.lgaKey = ''
    form.destination.area = ''
    resetPlace(form.destination)
  }
}

watch(
  () => form.origin.zoneId,
  (next, previous) => {
    if (previous && next !== previous) resetPlace(form.origin)
    form.destination.zoneId = ''
    form.destination.state = ''
    form.destination.lgaKey = ''
    form.destination.area = ''
    resetPlace(form.destination)
    void loadReachable()
  },
)

watch(
  () => form.destination.zoneId,
  (next, previous) => {
    if (previous && next !== previous) resetPlace(form.destination)
  },
)

watch(
  () => form.mode,
  async () => {
    zoneError.value = ''
    form.destination.zoneId = ''
    form.destination.state = ''
    form.destination.lgaKey = ''
    form.destination.area = ''
    resetPlace(form.destination)
    if (form.origin.zoneId) await loadReachable()
  },
)

onMounted(async () => {
  try {
    const response = await api<{ zones: SearchGeoZone[] }>('/public/zones')
    zones.value = response.zones ?? []
    if (zones.value.length) {
      const first = zones.value.find((zone) => !zone.parent_id) ?? zones.value[0]!
      assignZone(form.origin, first)
    }
  } catch (cause) {
    zoneError.value = apiErrorMessage(cause, 'Could not load service zones.')
  } finally {
    zonesLoading.value = false
  }
  if (pendingPrefill.value) await applyPrefill(pendingPrefill.value)
})

watch(
  () => props.prefill,
  (payload) => {
    if (payload) void applyPrefill(payload)
  },
)

let reachableRequestId = 0

async function loadReachable() {
  const requestId = ++reachableRequestId
  reachableZones.value = []
  if (!form.origin.zoneId) {
    reachableLoading.value = false
    return
  }
  reachableLoading.value = true
  zoneError.value = ''
  try {
    const params = new URLSearchParams({
      origin_zone_id: form.origin.zoneId,
      service_mode: form.mode,
    })
    const response = await api<{ zones: SearchGeoZone[] }>(`/public/zones/reachable?${params}`)
    if (requestId !== reachableRequestId) return
    reachableZones.value = response.zones ?? []
    ensureDestination()
  } catch (cause) {
    if (requestId !== reachableRequestId) return
    zoneError.value = apiErrorMessage(
      cause,
      form.mode === 'dropoff'
        ? 'Could not load destinations reachable from that drop-off zone.'
        : 'Could not load destination zones for that pickup zone.',
    )
  } finally {
    if (requestId === reachableRequestId) reachableLoading.value = false
  }
}

function packagePayload() {
  const dims = [form.length, form.width, form.height].map((value) => Number(value))
  const hasDimensions = dims.every((value) => Number.isFinite(value) && value > 0)
  return {
    weight_kg: Number(form.weight),
    length_cm: hasDimensions ? dims[0] : undefined,
    width_cm: hasDimensions ? dims[1] : undefined,
    height_cm: hasDimensions ? dims[2] : undefined,
    category: form.category,
  }
}

function validatePackage() {
  const weight = Number(form.weight)
  if (!Number.isFinite(weight) || weight <= 0 || weight > 500) {
    error.value = 'Enter a package weight between 0.1 kg and 500 kg.'
    return false
  }
  const dims = [form.length, form.width, form.height].map((value) => Number(value))
  const hasDimensions = dims.every((value) => Number.isFinite(value) && value > 0)
  if (dims.some((value) => Number.isFinite(value) && value > 0) && !hasDimensions) {
    error.value = 'Enter length, width and height together.'
    return false
  }
  return true
}

function locationError(location: SearchLocationValue, label: string, zonesList: SearchGeoZone[]) {
  const zone = zonesList.find((item) => item.id === location.zoneId)
  if (!zone) return `Choose the ${label} state, LGA and area.`
  if (location.latitude === null || location.longitude === null) {
    return `Verify the exact ${label} location from the map suggestions.`
  }
  return ''
}

function submitSearch() {
  error.value = ''
  if (!form.origin.zoneId) {
    error.value = form.mode === 'dropoff'
      ? 'Choose the state and zone where you will drop off the parcel.'
      : 'Choose your pickup state, LGA and area.'
    return
  }
  if (!form.destination.zoneId) {
    error.value = form.mode === 'dropoff'
      ? 'Choose a destination zone reachable from that drop-off hub.'
      : 'Choose a destination zone with an active provider route.'
    return
  }
  if (form.mode === 'pickup') {
    const originProblem = locationError(form.origin, 'pickup', zones.value)
    if (originProblem) {
      error.value = originProblem
      return
    }
  }
  const destinationProblem = locationError(form.destination, 'destination', destinationZones.value)
  if (destinationProblem) {
    error.value = destinationProblem
    return
  }
  if (!validatePackage()) return

  emit('search', {
    service_mode: form.mode,
    origin_zone_id: form.origin.zoneId,
    destination_zone_id: form.destination.zoneId,
    origin_area: form.origin.area.trim() || undefined,
    destination_area: form.destination.area.trim() || undefined,
    origin_address: form.mode === 'pickup' ? form.origin.address : undefined,
    destination_address: form.destination.address,
    origin_latitude: form.mode === 'pickup' ? form.origin.latitude ?? undefined : undefined,
    origin_longitude: form.mode === 'pickup' ? form.origin.longitude ?? undefined : undefined,
    destination_latitude: form.destination.latitude ?? undefined,
    destination_longitude: form.destination.longitude ?? undefined,
    origin_label: locationLabel(form.origin, zones.value),
    origin_lga: form.origin.lgaKey.split('|')[1],
    origin_state: form.origin.state,
    destination_label: locationLabel(form.destination, destinationZones.value),
    destination_lga: form.destination.lgaKey.split('|')[1],
    destination_state: form.destination.state,
    ...packagePayload(),
    sort: 'cheapest',
  })
}
</script>

<template>
  <div
    id="rate-search"
    class="site-panel w-full max-w-[540px] scroll-mt-24 rounded-[1.4rem] border p-5 sm:p-7"
  >
    <div class="flex items-start justify-between gap-4">
      <div>
        <p class="text-[10px] font-bold tracking-[0.18em] text-accent-700 uppercase dark:text-accent-400">where are you sending?</p>
        <h2 class="mt-1.5 text-[19px] font-semibold tracking-[-0.03em]">{{ heading || 'Where are you sending?' }}</h2>
        <p class="mt-1 text-[12.5px] text-ink-900/65">
          Choose how your parcel starts its journey, then the exact destination.
        </p>
      </div>
      <IconTile tone="accent" size="lg" rounded="2xl">
        <PackageCheck :size="20" />
      </IconTile>
    </div>

    <form class="mt-5 space-y-3.5" @submit.prevent="submitSearch">
      <div class="grid grid-cols-2 gap-1 rounded-xl border border-ink-900/10 bg-canvas p-1">
        <button
          type="button"
          class="flex items-center justify-center gap-1.5 rounded-lg px-2 py-2.5 text-[11.5px] font-semibold transition"
          :class="form.mode === 'pickup' ? 'bg-surface text-ink-900 shadow-sm' : 'text-ink-900/50 hover:text-ink-900/75'"
          @click="form.mode = 'pickup'"
        >
          <Truck :size="14" /> Pick up from me
        </button>
        <button
          type="button"
          class="flex items-center justify-center gap-1.5 rounded-lg px-2 py-2.5 text-[11.5px] font-semibold transition"
          :class="form.mode === 'dropoff' ? 'bg-surface text-ink-900 shadow-sm' : 'text-ink-900/50 hover:text-ink-900/75'"
          @click="form.mode = 'dropoff'"
        >
          <Warehouse :size="14" /> Drop off at a hub
        </button>
      </div>

      <p class="-mt-1 text-[11px] leading-relaxed text-ink-900/50">
        <template v-if="form.mode === 'pickup'">
          A provider collects from your verified pickup pin and delivers to the destination pin.
        </template>
        <template v-else>
          Choose the zone where you can hand the parcel to a provider hub, then where it is going.
        </template>
      </p>

      <label v-if="form.mode === 'pickup' && savedPickupAddresses.length" class="block">
        <span :class="labelClass">Use a saved pickup</span>
        <select :class="fieldClass" @change="useSavedAddress($event, 'origin')">
          <option value="">Choose a saved address…</option>
          <option v-for="address in savedPickupAddresses" :key="address.id" :value="address.id">
            {{ address.label }} — {{ address.street_address }}, {{ address.city }}
          </option>
        </select>
      </label>

      <SearchLocationPicker
        v-model="form.origin"
        :label="form.mode === 'dropoff' ? 'Drop-off hub' : 'Pickup'"
        :zones="zones"
        :loading="zonesLoading"
        :show-pin="form.mode === 'pickup'"
      />

      <label v-if="savedDeliveryAddresses.length" class="block">
        <span :class="labelClass">Use a saved destination</span>
        <select
          :class="fieldClass"
          :disabled="!form.origin.zoneId"
          @change="useSavedAddress($event, 'destination')"
        >
          <option value="">Choose a saved address…</option>
          <option v-for="address in savedDeliveryAddresses" :key="address.id" :value="address.id">
            {{ address.label }} — {{ address.street_address }}, {{ address.city }}
          </option>
        </select>
      </label>

      <SearchLocationPicker
        v-model="form.destination"
        label="Destination"
        :zones="destinationZones"
        :loading="destinationLoading"
      />

      <p v-if="savedAddresses.length" class="text-[10.5px] leading-relaxed text-ink-900/45">
        Saved addresses fill the zone and map pin for you; confirm the pin before searching.
        <NuxtLink to="/dashboard/customer/addresses" class="font-semibold text-brand-700 hover:text-brand-600">
          Manage addresses
        </NuxtLink>
      </p>

      <p
        v-if="form.origin.zoneId && !destinationLoading && !destinationZones.length"
        class="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2.5 text-[12px] font-medium text-amber-800"
      >
        {{ form.mode === 'dropoff'
          ? 'No provider has a drop-off hub in that zone yet. Try another zone or pickup instead.'
          : 'No provider has an active route from that pickup zone yet.' }}
      </p>

      <p v-if="zoneError" class="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2.5 text-[12px] font-medium text-amber-800">
        {{ zoneError }}
      </p>

      <div class="grid grid-cols-2 gap-3">
        <label>
          <span :class="labelClass">Package weight</span>
          <span class="relative block">
            <input
              v-model.number="form.weight"
              :class="fieldClass"
              type="number"
              name="weight"
              min="0.1"
              max="500"
              step="0.1"
              inputmode="decimal"
              required
            />
            <span class="pointer-events-none absolute top-1/2 right-3.5 -translate-y-1/2 text-[11px] text-ink-900/40">kg</span>
          </span>
        </label>
        <label>
          <span :class="labelClass">Package type</span>
          <select v-model="form.category" :class="fieldClass" name="category">
            <option v-for="category in categories" :key="category.value" :value="category.value">
              {{ category.label }}
            </option>
          </select>
        </label>
      </div>

      <details class="group">
        <summary class="cursor-pointer list-none text-[11.5px] font-semibold text-ink-900/55 marker:hidden [&::-webkit-details-marker]:hidden">
          Add package dimensions <span class="font-normal text-ink-900/35">(optional)</span>
        </summary>
        <div class="mt-3 grid grid-cols-3 gap-2">
          <label v-for="dimension in dimensions" :key="dimension.key">
            <span :class="labelClass">{{ dimension.label }} (cm)</span>
            <input
              v-model.number="form[dimension.key]"
              :class="fieldClass"
              type="number"
              min="1"
              max="300"
              step="1"
              :name="dimension.key"
            />
          </label>
        </div>
      </details>

      <p
        v-if="savedNotice"
        class="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2.5 text-[12px] font-medium text-amber-800"
      >
        {{ savedNotice }}
      </p>

      <p
        v-if="error"
        role="alert"
        class="rounded-xl border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] font-medium text-red-700 dark:border-red-400/25 dark:bg-red-400/10 dark:text-red-200"
      >
        {{ error }}
      </p>

      <button
        type="submit"
        :disabled="loading || zonesLoading || !form.origin.zoneId || !form.destination.zoneId"
        class="flex w-full items-center justify-center gap-2 rounded-xl bg-accent-500 px-4 py-3.5 text-[14px] font-semibold text-white shadow-[0_14px_28px_-16px_rgba(255,122,0,0.75)] transition hover:-translate-y-0.5 hover:bg-accent-400 hover:shadow-[0_18px_32px_-16px_rgba(255,122,0,0.85)] disabled:translate-y-0 disabled:opacity-60"
      >
        <Loader2 v-if="loading" :size="16" class="animate-spin" />
        <Search v-else :size="16" />
        {{ loading ? 'Checking providers…' : 'Compare provider routes' }}
        <ArrowRight :size="15" class="ml-auto" />
      </button>
    </form>

    <p class="mt-3 text-center text-[10.5px] leading-relaxed text-ink-900/40">
      Only providers with an active route for this service and direction are shown. No account needed.
    </p>
  </div>
</template>
