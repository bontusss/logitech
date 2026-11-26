<script setup lang="ts">
import { ArrowRight, Loader2, PackageCheck, Search, Truck, Warehouse } from 'lucide-vue-next'
import type { MarketplaceSearchPayload } from '~/types/marketplace'
import type { SearchGeoZone, SearchLocationValue } from '~/components/SearchLocationPicker.vue'
import { ITEM_CATEGORIES } from '~/utils/itemCategories'

withDefaults(defineProps<{ loading?: boolean }>(), { loading: false })

const emit = defineEmits<{ search: [payload: MarketplaceSearchPayload] }>()

type ServiceMode = 'pickup' | 'dropoff'

const { api } = useApi()
const zones = ref<SearchGeoZone[]>([])
const reachableZones = ref<SearchGeoZone[]>([])
const zonesLoading = ref(true)
const reachableLoading = ref(false)
const zoneError = ref('')
const error = ref('')

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

function ensureDestination() {
  const list = destinationZones.value
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
})

async function loadReachable() {
  reachableZones.value = []
  if (!form.origin.zoneId) return
  reachableLoading.value = true
  zoneError.value = ''
  try {
    const params = new URLSearchParams({
      origin_zone_id: form.origin.zoneId,
      service_mode: form.mode,
    })
    const response = await api<{ zones: SearchGeoZone[] }>(`/public/zones/reachable?${params}`)
    reachableZones.value = response.zones ?? []
    ensureDestination()
  } catch (cause) {
    zoneError.value = apiErrorMessage(
      cause,
      form.mode === 'dropoff'
        ? 'Could not load destinations reachable from that drop-off zone.'
        : 'Could not load destination zones for that pickup zone.',
    )
  } finally {
    reachableLoading.value = false
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
        <h2 class="mt-1.5 text-[19px] font-semibold tracking-[-0.03em]">Where are you sending?</h2>
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

      <SearchLocationPicker
        v-model="form.origin"
        :label="form.mode === 'dropoff' ? 'Drop-off hub' : 'Pickup'"
        :zones="zones"
        :loading="zonesLoading"
        :show-pin="form.mode === 'pickup'"
      />

      <SearchLocationPicker
        v-model="form.destination"
        label="Destination"
        :zones="destinationZones"
        :loading="destinationLoading"
      />

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
