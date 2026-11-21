<script setup lang="ts">
import { Check, Crosshair, Loader2, MapPin } from 'lucide-vue-next'

export interface BookingLocation {
  address: string
  latitude: number
  longitude: number
}

const props = defineProps<{
  label: string
  modelValue: BookingLocation | null
}>()
const emit = defineEmits<{
  'update:modelValue': [value: BookingLocation | null]
}>()

interface Suggestion {
  address: string
  latitude: number
  longitude: number
}

const input = ref(props.modelValue?.address ?? '')
const suggestions = ref<Suggestion[]>([])
const searching = ref(false)
const locating = ref(false)
const adjusting = ref(false)
const error = ref('')
const open = ref(false)
const { token, loadError: mapboxConfigError, load: loadMapboxToken } = useMapboxToken()
let timer: ReturnType<typeof setTimeout> | undefined
let controller: AbortController | null = null
let suppressInputWatch = false

onMounted(() => void loadMapboxToken())

watch(input, (value) => {
  if (suppressInputWatch) return
  emit('update:modelValue', null)
  error.value = ''
  suggestions.value = []
  open.value = false
  if (timer) clearTimeout(timer)
  controller?.abort()
  if (value.trim().length < 2 || !token) return
  timer = setTimeout(() => void search(value.trim()), 300)
})

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
  controller?.abort()
})

async function search(query: string) {
  const request = new AbortController()
  controller = request
  searching.value = true
  try {
    const params = new URLSearchParams({
      q: query,
      country: 'ng',
      autocomplete: 'true',
      limit: '5',
      access_token: token,
    })
    const response = await fetch(`https://api.mapbox.com/search/geocode/v6/forward?${params}`, {
      signal: request.signal,
    })
    if (!response.ok) throw new Error('Address search is unavailable right now.')
    const result = await response.json() as {
      features?: Array<{
        geometry?: { coordinates?: [number, number] }
        properties?: { full_address?: string; name?: string; place_formatted?: string }
      }>
    }
    if (request.signal.aborted) return
    suggestions.value = (result.features ?? []).flatMap((feature) => {
      const coordinates = feature.geometry?.coordinates
      const address = feature.properties?.full_address
        || [feature.properties?.name, feature.properties?.place_formatted].filter(Boolean).join(', ')
      return coordinates && address
        ? [{ address, longitude: coordinates[0], latitude: coordinates[1] }]
        : []
    })
    open.value = true
    if (!suggestions.value.length) error.value = 'No matching places. Try adding a city or state.'
  } catch (cause) {
    if (!request.signal.aborted) {
      error.value = cause instanceof Error ? cause.message : 'Could not search for this address.'
    }
  } finally {
    if (!request.signal.aborted) searching.value = false
  }
}

async function choose(place: Suggestion) {
  suppressInputWatch = true
  input.value = place.address
  suggestions.value = []
  open.value = false
  error.value = ''
  await nextTick()
  emit('update:modelValue', place)
  suppressInputWatch = false
}

async function reverseGeocode(latitude: number, longitude: number): Promise<Suggestion> {
  const params = new URLSearchParams({
    longitude: String(longitude),
    latitude: String(latitude),
    access_token: token.value,
  })
  const response = await fetch(`https://api.mapbox.com/search/geocode/v6/reverse?${params}`)
  if (!response.ok) throw new Error('Could not verify this spot on the map.')
  const result = await response.json() as {
    features?: Array<{
      geometry?: { coordinates?: [number, number] }
      properties?: { full_address?: string; name?: string; place_formatted?: string }
    }>
  }
  const feature = result.features?.[0]
  const address = feature?.properties?.full_address
    || [feature?.properties?.name, feature?.properties?.place_formatted].filter(Boolean).join(', ')
  const coordinates = feature?.geometry?.coordinates
  if (!address || !coordinates) throw new Error('Mapbox could not find an address at this spot.')
  return { address, longitude: coordinates[0], latitude: coordinates[1] }
}

async function adjustPin(coordinates: { latitude: number; longitude: number }) {
  if (!token.value) return
  adjusting.value = true
  error.value = ''
  try {
    const place = await reverseGeocode(coordinates.latitude, coordinates.longitude)
    suppressInputWatch = true
    input.value = place.address
    await nextTick()
    // Preserve the exact dragged position instead of snapping the pin back to
    // the nearest road returned by reverse geocoding.
    emit('update:modelValue', {
      ...place,
      latitude: coordinates.latitude,
      longitude: coordinates.longitude,
    })
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : 'Could not verify this spot. Move the pin again.'
  } finally {
    suppressInputWatch = false
    adjusting.value = false
  }
}

function useCurrentLocation() {
  error.value = ''
  if (!navigator.geolocation) {
    error.value = 'This browser does not support location access.'
    return
  }
  if (!token) {
    error.value = 'Mapbox address search is not configured for this booking page.'
    return
  }
  locating.value = true
  navigator.geolocation.getCurrentPosition(
    async ({ coords }) => {
      try {
        const place = await reverseGeocode(coords.latitude, coords.longitude)
        await choose(place)
      } catch (cause) {
        error.value = cause instanceof Error ? cause.message : 'Could not verify your current location.'
      } finally {
        locating.value = false
      }
    },
    (cause) => {
      error.value = cause.code === cause.PERMISSION_DENIED
        ? 'Allow location access in your browser, or search for an address instead.'
        : 'Could not get your location. Search for an address instead.'
      locating.value = false
    },
    // Do not reuse a cached fix: a stale location is worse than a fresh one.
    { enableHighAccuracy: true, timeout: 15000, maximumAge: 0 },
  )
}
</script>

<template>
  <div class="relative">
    <label class="mb-1.5 block text-[11px] font-semibold text-slate-600">{{ label }}</label>
    <div class="relative">
      <MapPin :size="15" class="pointer-events-none absolute top-3.5 left-3 text-slate-400" />
      <input
        v-model="input"
        type="search"
        autocomplete="off"
        :aria-label="label"
        :disabled="!token"
        placeholder="Search a street, area, or city"
        class="w-full rounded-xl border border-slate-200 bg-white py-3 pl-9 pr-10 text-[13px] text-slate-800 outline-none transition placeholder:text-slate-400 focus:border-[var(--booking-primary)] focus:ring-4 focus:ring-[var(--booking-primary-soft)]"
        @focus="suggestions.length && (open = true)"
        @keydown.esc="open = false"
      />
      <Loader2 v-if="searching" :size="15" class="absolute top-3.5 right-3 animate-spin text-slate-400" />
      <Check v-else-if="modelValue" :size="15" class="absolute top-3.5 right-3 text-emerald-600" />
    </div>
    <div class="mt-1.5 flex items-center justify-between gap-2">
      <p v-if="modelValue" class="text-[10px] font-medium text-emerald-700">Mapbox verified · pin selected</p>
      <p v-else class="text-[10px] text-slate-400">Choose a suggested place to set the map pin.</p>
      <button
        type="button"
        class="inline-flex shrink-0 items-center gap-1 text-[10px] font-semibold text-[var(--booking-primary)] disabled:opacity-50"
        :disabled="locating"
        @click="useCurrentLocation"
      >
        <Loader2 v-if="locating" :size="12" class="animate-spin" />
        <Crosshair v-else :size="12" />
        Use my location
      </button>
    </div>
    <div v-if="token && modelValue" class="mt-2 space-y-1.5">
      <MapPinAdjuster
        :latitude="modelValue.latitude"
        :longitude="modelValue.longitude"
        :access-token="token"
        height="140px"
        @change="adjustPin"
      />
      <p class="text-[9.5px] leading-4 text-slate-400">
        {{ adjusting ? 'Updating the address from the pin…' : 'Device location can be approximate. Drag the pin to the exact spot, or search again if you are in the wrong area.' }}
      </p>
    </div>
    <p v-if="!token" class="mt-2 text-[10px] leading-4 text-amber-700">
      {{ mapboxConfigError || 'Mapbox is not configured for this page yet. Add its public access token to enable verified location search.' }}
    </p>
    <ul
      v-if="open && suggestions.length"
      class="absolute z-30 mt-1.5 max-h-56 w-full overflow-y-auto rounded-xl border border-slate-200 bg-white p-1.5 shadow-xl"
      role="listbox"
    >
      <li v-for="place in suggestions" :key="`${place.longitude}:${place.latitude}`">
        <button
          type="button"
          role="option"
          class="flex w-full items-start gap-2 rounded-lg px-2.5 py-2.5 text-left text-[12px] leading-relaxed text-slate-700 hover:bg-slate-50"
          @click="choose(place)"
        >
          <MapPin :size="14" class="mt-0.5 shrink-0 text-[var(--booking-primary)]" />
          <span>{{ place.address }}</span>
        </button>
      </li>
    </ul>
    <p v-if="error" role="status" class="mt-1.5 text-[10px] text-rose-600">{{ error }}</p>
  </div>
</template>
