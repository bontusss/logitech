<script setup lang="ts">
import {
  ArrowLeftRight,
  ArrowRight,
  Building2,
  Car,
  Loader2,
  MapPin,
  PackageCheck,
  Search,
} from 'lucide-vue-next'
import type { MarketplaceSearchPayload } from '~/types/marketplace'
import type { VerifiedAddress } from '~/components/VerifiedAddressField.vue'
import { NIGERIAN_LOCATIONS } from '~/utils/nigeria'
import { ITEM_CATEGORIES } from '~/utils/itemCategories'

withDefaults(defineProps<{ loading?: boolean }>(), { loading: false })

const emit = defineEmits<{ search: [payload: MarketplaceSearchPayload] }>()

interface MarketplaceLocation {
  state: string
  city: string
}

const fallbackLocations: MarketplaceLocation[] = Object.entries(NIGERIAN_LOCATIONS).flatMap(
  ([state, cities]) => cities.map((city) => ({ state, city })),
)
if (!fallbackLocations.some((location) => location.state === 'Lagos' && location.city === 'Lagos')) {
  fallbackLocations.push({ state: 'Lagos', city: 'Lagos' })
}

const { api } = useApi()
const { token: mapboxToken, load: loadMapboxToken } = useMapboxToken()
const locations = ref<MarketplaceLocation[]>(fallbackLocations)
const liveLocations = ref(false)
const serviceMode = ref<'intercity' | 'last_mile'>('intercity')
const pickup = ref<VerifiedAddress | null>(null)
const dropoff = ref<VerifiedAddress | null>(null)
const distanceKm = ref(0)
const distanceLoading = ref(false)
const distanceError = ref('')
const error = ref('')
let distanceTimer: ReturnType<typeof setTimeout> | undefined
let distanceController: AbortController | null = null

const cityMap = computed(() => {
  const map: Record<string, string[]> = {}
  for (const location of locations.value) {
    const cities = map[location.state] ?? (map[location.state] = [])
    if (!cities.includes(location.city)) cities.push(location.city)
  }
  for (const state of Object.keys(map)) {
    map[state]!.sort((a, b) => a.localeCompare(b))
  }
  return map
})

const states = computed(() => Object.keys(cityMap.value).sort((a, b) => a.localeCompare(b)))
const originCities = computed(() => cityMap.value[form.originState] ?? [])
const destinationCities = computed(() => cityMap.value[form.destinationState] ?? [])

const form = reactive({
  originState: 'Abia',
  originCity: 'Aba',
  destinationState: 'Lagos',
  destinationCity: 'Lagos',
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
  'w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-3 text-[13px] font-medium text-ink-900 transition placeholder:text-ink-900/35 hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5'
const selectClass =
  'w-full rounded-xl border border-slate-200 bg-canvas px-3 py-2.5 text-[12.5px] font-medium text-ink-900 transition hover:border-slate-300 focus:border-brand-500 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/10 dark:bg-white/5'
const labelClass = 'mb-1.5 block text-[11px] font-semibold text-ink-900/55'

watch(
  () => form.originState,
  () => {
    if (!originCities.value.includes(form.originCity)) {
      form.originCity = originCities.value[0] ?? ''
    }
  },
)
watch(
  () => form.destinationState,
  () => {
    if (!destinationCities.value.includes(form.destinationCity)) {
      form.destinationCity = destinationCities.value[0] ?? ''
    }
  },
)
watch([pickup, dropoff], () => scheduleDistance())

onMounted(async () => {
  void loadMapboxToken()
  try {
    const response = await api<{ locations: MarketplaceLocation[] }>('/marketplace/locations')
    if (response.locations?.length) {
      locations.value = response.locations
      liveLocations.value = true
      ensureSelection()
    }
  } catch {
    // Keep the built-in Nigerian location list so the search remains usable.
  }
})

onBeforeUnmount(() => {
  if (distanceTimer) clearTimeout(distanceTimer)
  distanceController?.abort()
})

function ensureSelection() {
  if (!originCities.value.includes(form.originCity)) {
    const first = locations.value[0]
    if (first) {
      form.originState = first.state
      form.originCity = first.city
    }
  }
  if (!destinationCities.value.includes(form.destinationCity)) {
    const next = locations.value.find(
      (location) =>
        location.state !== form.originState || location.city !== form.originCity,
    )
    if (next) {
      form.destinationState = next.state
      form.destinationCity = next.city
    }
  }
}

function swapCities() {
  const originState = form.originState
  const originCity = form.originCity
  form.originState = form.destinationState
  form.originCity = form.destinationCity
  form.destinationState = originState
  form.destinationCity = originCity
}

function setMode(mode: 'intercity' | 'last_mile') {
  serviceMode.value = mode
  error.value = ''
}

function scheduleDistance() {
  if (distanceTimer) clearTimeout(distanceTimer)
  distanceController?.abort()
  distanceKm.value = 0
  distanceError.value = ''
  if (serviceMode.value !== 'last_mile' || !pickup.value || !dropoff.value) {
    distanceLoading.value = false
    return
  }
  distanceTimer = setTimeout(() => void calculateDistance(), 350)
}

async function calculateDistance() {
  if (!pickup.value || !dropoff.value) return
  if (!mapboxToken.value) {
    distanceError.value = 'Mapbox distance calculation is not configured.'
    return
  }
  if (
    pickup.value.latitude === dropoff.value.latitude
    && pickup.value.longitude === dropoff.value.longitude
  ) {
    distanceError.value = 'Pickup and drop-off must be different addresses.'
    return
  }
  const controller = new AbortController()
  distanceController = controller
  distanceLoading.value = true
  distanceError.value = ''
  try {
    const coordinates = `${pickup.value.longitude},${pickup.value.latitude};${dropoff.value.longitude},${dropoff.value.latitude}`
    const params = new URLSearchParams({
      overview: 'false',
      access_token: mapboxToken.value,
    })
    const response = await fetch(
      `https://api.mapbox.com/directions/v5/mapbox/driving/${coordinates}?${params}`,
      { signal: controller.signal },
    )
    if (!response.ok) throw new Error('Could not calculate the road distance right now.')
    const result = await response.json() as { routes?: Array<{ distance?: number }> }
    const meters = result.routes?.[0]?.distance
    if (controller.signal.aborted) return
    if (!meters || meters <= 0) throw new Error('Mapbox could not calculate a route between these addresses.')
    distanceKm.value = Math.round((meters / 1000) * 100) / 100
  } catch (cause) {
    if (!controller.signal.aborted) {
      distanceError.value = cause instanceof Error ? cause.message : 'Could not calculate the road distance.'
    }
  } finally {
    if (!controller.signal.aborted) distanceLoading.value = false
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

function submitSearch() {
  error.value = ''
  if (!validatePackage()) return

  if (serviceMode.value === 'last_mile') {
    if (!pickup.value || !dropoff.value) {
      error.value = 'Choose both addresses from the Mapbox suggestions so we can verify the route.'
      return
    }
    if (!pickup.value.city || !pickup.value.state || !dropoff.value.city || !dropoff.value.state) {
      error.value = 'Those addresses are missing a city or state. Choose a more specific suggestion.'
      return
    }
    if (distanceLoading.value) {
      error.value = 'Wait for the road distance to finish calculating.'
      return
    }
    if (!distanceKm.value || distanceError.value) {
      error.value = distanceError.value || 'Choose verified pickup and drop-off addresses first.'
      return
    }
    emit('search', {
      service_type: 'last_mile',
      origin_state: pickup.value.state,
      origin_city: pickup.value.city,
      destination_state: dropoff.value.state,
      destination_city: dropoff.value.city,
      pickup_address: pickup.value.address,
      dropoff_address: dropoff.value.address,
      pickup_latitude: pickup.value.latitude,
      pickup_longitude: pickup.value.longitude,
      dropoff_latitude: dropoff.value.latitude,
      dropoff_longitude: dropoff.value.longitude,
      distance_km: distanceKm.value,
      ...packagePayload(),
      sort: 'cheapest',
    })
    return
  }

  if (!form.originState || !form.originCity || !form.destinationState || !form.destinationCity) {
    error.value = 'Choose the origin and destination states and cities.'
    return
  }
  if (form.originState === form.destinationState && form.originCity === form.destinationCity) {
    error.value = 'Origin and destination must be different.'
    return
  }
  emit('search', {
    service_type: 'intercity',
    origin_state: form.originState,
    origin_city: form.originCity,
    destination_state: form.destinationState,
    destination_city: form.destinationCity,
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
        <p class="text-[10px] font-bold tracking-[0.18em] text-brand-600 uppercase">Start here</p>
        <h2 class="mt-1.5 text-[19px] font-semibold tracking-[-0.03em]">Where are you sending?</h2>
        <p class="mt-1 text-[12.5px] text-ink-900/65">
          Enter your shipment to compare live provider rates.
        </p>
      </div>
      <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-100 to-cyan-50 text-violet-700 shadow-sm dark:from-violet-400/15 dark:to-cyan-400/10 dark:text-violet-200">
        <PackageCheck :size="20" />
      </div>
    </div>

    <div class="mt-4 grid grid-cols-2 gap-1 rounded-xl border border-slate-200 bg-canvas p-1 dark:border-white/10 dark:bg-white/5" role="tablist" aria-label="Delivery type">
      <button
        v-for="mode in ([
          { id: 'intercity', label: 'Between cities', icon: Building2 },
          { id: 'last_mile', label: 'Within a city', icon: Car },
        ] as const)"
        :key="mode.id"
        type="button"
        role="tab"
        :aria-selected="serviceMode === mode.id"
        class="inline-flex items-center justify-center gap-1.5 rounded-lg px-2.5 py-2.5 text-[11.5px] font-semibold transition"
        :class="serviceMode === mode.id
          ? 'bg-white text-ink-900 shadow-sm dark:bg-white/10 dark:text-white'
          : 'text-ink-900/50 hover:text-ink-900'"
        @click="setMode(mode.id)"
      >
        <component :is="mode.icon" :size="13" />
        {{ mode.label }}
      </button>
    </div>

    <form class="mt-4 space-y-3.5" @submit.prevent="submitSearch">
      <template v-if="serviceMode === 'intercity'">
        <div class="grid grid-cols-[1fr_auto_1fr] items-end gap-2">
          <div class="min-w-0">
            <span :class="labelClass">Sending from</span>
            <div class="space-y-1.5">
              <select
                v-model="form.originState"
                :class="selectClass"
                name="origin_state"
                aria-label="Origin state"
                required
              >
                <option value="" disabled>State</option>
                <option v-for="state in states" :key="state" :value="state">{{ state }}</option>
              </select>
              <select
                v-model="form.originCity"
                :class="selectClass"
                name="origin_city"
                aria-label="Origin city"
                :disabled="!form.originState"
                required
              >
                <option value="" disabled>City</option>
                <option v-for="city in originCities" :key="city" :value="city">{{ city }}</option>
              </select>
            </div>
          </div>
          <button
            type="button"
            class="mb-1 flex size-9 shrink-0 items-center justify-center rounded-xl border border-slate-200 bg-canvas text-ink-900/50 transition hover:-translate-y-0.5 hover:border-brand-200 hover:text-brand-600 dark:border-white/10 dark:bg-white/5"
            aria-label="Swap origin and destination"
            @click="swapCities"
          >
            <ArrowLeftRight :size="15" />
          </button>
          <div class="min-w-0">
            <span :class="labelClass">Sending to</span>
            <div class="space-y-1.5">
              <select
                v-model="form.destinationState"
                :class="selectClass"
                name="destination_state"
                aria-label="Destination state"
                required
              >
                <option value="" disabled>State</option>
                <option v-for="state in states" :key="state" :value="state">{{ state }}</option>
              </select>
              <select
                v-model="form.destinationCity"
                :class="selectClass"
                name="destination_city"
                aria-label="Destination city"
                :disabled="!form.destinationState"
                required
              >
                <option value="" disabled>City</option>
                <option v-for="city in destinationCities" :key="city" :value="city">{{ city }}</option>
              </select>
            </div>
          </div>
        </div>
      </template>

      <template v-else>
        <div class="space-y-3 rounded-xl border border-slate-200/80 bg-canvas/60 p-3 dark:border-white/10 dark:bg-white/5">
          <VerifiedAddressField v-model="pickup" label="Pickup address" placeholder="Street, estate or landmark" />
          <VerifiedAddressField v-model="dropoff" label="Drop-off address" placeholder="Street, estate or landmark" />
          <div class="flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-2.5 py-2 text-[10.5px] dark:border-white/10 dark:bg-white/5">
            <Loader2 v-if="distanceLoading" :size="13" class="animate-spin text-brand-600" />
            <MapPin v-else :size="13" class="text-brand-600" />
            <span v-if="distanceLoading" class="text-ink-900/55">Calculating the Mapbox road distance…</span>
            <span v-else-if="distanceKm" class="font-semibold text-emerald-700 dark:text-emerald-300">
              {{ distanceKm.toLocaleString() }} km verified route
              <span class="ml-1 font-normal text-ink-900/45">· used for local rates</span>
            </span>
            <span v-else-if="distanceError" class="text-rose-600">{{ distanceError }}</span>
            <span v-else class="text-ink-900/45">Choose both verified addresses to calculate distance.</span>
          </div>
        </div>
      </template>

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
        v-if="error"
        role="alert"
        class="rounded-xl border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] font-medium text-red-700 dark:border-red-400/25 dark:bg-red-400/10 dark:text-red-200"
      >
        {{ error }}
      </p>

      <button
        type="submit"
        :disabled="loading"
        class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-brand-600 to-violet-600 px-4 py-3.5 text-[14px] font-semibold text-white shadow-[0_14px_28px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 hover:shadow-[0_18px_32px_-16px_rgba(43,75,255,0.95)] disabled:translate-y-0 disabled:opacity-60"
      >
        <Loader2 v-if="loading" :size="16" class="animate-spin" />
        <Search v-else :size="16" />
        {{ loading ? 'Checking providers…' : serviceMode === 'last_mile' ? 'Compare local delivery options' : 'Compare delivery options' }}
        <ArrowRight :size="15" class="ml-auto" />
      </button>
    </form>

    <p class="mt-3 text-center text-[10.5px] leading-relaxed text-ink-900/40">
      <template v-if="serviceMode === 'last_mile'">Mapbox-verified addresses · live last-mile coverage.</template>
      <template v-else-if="liveLocations">Live provider coverage. No account needed.</template>
      <template v-else>No commitment. Review provider rates before you choose.</template>
    </p>
  </div>
</template>
