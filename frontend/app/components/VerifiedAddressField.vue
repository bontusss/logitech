<script setup lang="ts">
import { Check, Crosshair, Loader2, MapPin } from 'lucide-vue-next'

export interface VerifiedAddress {
  address: string
  latitude: number
  longitude: number
  city: string
  state: string
}

const props = defineProps<{
  label: string
  placeholder?: string
  modelValue: VerifiedAddress | null
}>()
const emit = defineEmits<{ 'update:modelValue': [value: VerifiedAddress | null] }>()

interface Suggestion extends VerifiedAddress {}

const input = ref(props.modelValue?.address ?? '')
const suggestions = ref<Suggestion[]>([])
const open = ref(false)
const searching = ref(false)
const locating = ref(false)
const adjusting = ref(false)
const error = ref('')
const { token, loadError: tokenError, load: loadToken } = useMapboxToken()
let timer: ReturnType<typeof setTimeout> | undefined
let controller: AbortController | null = null
let suppressInputWatch = false

onMounted(() => void loadToken())
onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
  controller?.abort()
})

watch(input, (value) => {
  if (suppressInputWatch) return
  emit('update:modelValue', null)
  error.value = ''
  suggestions.value = []
  open.value = false
  controller?.abort()
  if (timer) clearTimeout(timer)
  if (value.trim().length < 3 || !token.value) return
  timer = setTimeout(() => void search(value.trim()), 320)
})

function parseFeature(feature: {
  geometry?: { coordinates?: [number, number] }
  properties?: {
    full_address?: string
    name?: string
    place_formatted?: string
    context?: Record<string, { name?: string }>
  }
}): Suggestion | null {
  const coordinates = feature.geometry?.coordinates
  const address = feature.properties?.full_address
    || [feature.properties?.name, feature.properties?.place_formatted].filter(Boolean).join(', ')
  if (!coordinates || !address) return null
  const context = feature.properties?.context ?? {}
  const city = context.place?.name || context.locality?.name || context.district?.name || ''
  const state = context.region?.name || ''
  return { address, longitude: coordinates[0], latitude: coordinates[1], city, state }
}

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
      access_token: token.value,
    })
    const response = await fetch(`https://api.mapbox.com/search/geocode/v6/forward?${params}`, {
      signal: request.signal,
    })
    if (!response.ok) throw new Error('Address search is unavailable right now.')
    const result = await response.json() as {
      features?: Array<{
        geometry?: { coordinates?: [number, number] }
        properties?: {
          full_address?: string
          name?: string
          place_formatted?: string
          context?: Record<string, { name?: string }>
        }
      }>
    }
    if (request.signal.aborted) return
    suggestions.value = (result.features ?? []).flatMap((feature) => {
      const parsed = parseFeature(feature)
      return parsed ? [parsed] : []
    })
    open.value = true
    if (!suggestions.value.length) error.value = 'No matching places. Add an area or city.'
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
      properties?: {
        full_address?: string
        name?: string
        place_formatted?: string
        context?: Record<string, { name?: string }>
      }
    }>
  }
  const place = result.features?.[0] ? parseFeature(result.features[0]) : null
  if (!place) throw new Error('Mapbox could not find an address at this spot.')
  return place
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
    // Keep the pin exactly where the customer dropped it. Mapbox reverse
    // geocoding can snap its result to the nearest road, which would otherwise
    // pull the marker away from the chosen building or gate.
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
  if (!token.value) {
    error.value = 'Mapbox address search is not configured.'
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
        ? 'Allow location access, or search for an address instead.'
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
    <span class="mb-1.5 flex items-center justify-between gap-2">
      <span class="text-[10.5px] font-semibold text-ink-900/60">{{ label }}</span>
      <button
        type="button"
        class="inline-flex items-center gap-0.5 text-[9.5px] font-semibold text-brand-700 disabled:opacity-50 dark:text-brand-200"
        :disabled="locating"
        @click="useCurrentLocation"
      >
        <Loader2 v-if="locating" :size="10" class="animate-spin" />
        <Crosshair v-else :size="10" />
        Use my location
      </button>
    </span>
    <div class="relative">
      <MapPin :size="14" class="pointer-events-none absolute top-3 left-3 text-ink-900/35" />
      <input
        v-model="input"
        type="search"
        autocomplete="off"
        :aria-label="label"
        :disabled="!token"
        :placeholder="placeholder || 'Search a street, area or landmark'"
        class="w-full rounded-xl border border-slate-200 bg-canvas py-2.5 pr-9 pl-8 text-[12px] font-medium text-ink-900 outline-none transition placeholder:text-ink-900/35 focus:border-brand-500 disabled:opacity-60 dark:border-white/10 dark:bg-white/5"
        @focus="suggestions.length && (open = true)"
        @keydown.esc="open = false"
      />
      <Loader2 v-if="searching" :size="13" class="absolute top-3 right-3 animate-spin text-ink-900/35" />
      <Check v-else-if="modelValue" :size="13" class="absolute top-3 right-3 text-success-600" />
    </div>
    <p v-if="modelValue" class="mt-1 text-[9.5px] font-medium text-success-700 dark:text-success-200">
      Mapbox verified<template v-if="modelValue.city"> · {{ modelValue.city }}<template v-if="modelValue.state">, {{ modelValue.state }}</template></template>
    </p>
    <p v-else-if="tokenError" class="mt-1 text-[9.5px] leading-4 text-accent-700">{{ tokenError }}</p>
    <div v-if="token && modelValue" class="mt-2 space-y-1.5">
      <MapPinAdjuster
        :latitude="modelValue.latitude"
        :longitude="modelValue.longitude"
        :access-token="token"
        height="148px"
        accent="#4f46e5"
        @change="adjustPin"
      />
      <p class="text-[9px] leading-4 text-ink-900/40">
        {{ adjusting ? 'Updating the address from the pin…' : 'Device location can be approximate. Drag the pin to the exact spot, or search again if you are in the wrong area.' }}
      </p>
    </div>
    <p v-else class="mt-1 text-[9.5px] text-ink-900/40">Choose a suggestion to verify the address.</p>
    <ul
      v-if="open && suggestions.length"
      class="absolute z-40 mt-1.5 max-h-52 w-full overflow-y-auto rounded-xl border border-slate-200 bg-white p-1.5 shadow-xl dark:border-white/10 dark:bg-ink-950"
      role="listbox"
    >
      <li v-for="place in suggestions" :key="`${place.longitude}:${place.latitude}`">
        <button
          type="button"
          role="option"
          class="flex w-full items-start gap-2 rounded-lg px-2.5 py-2 text-left text-[11px] leading-relaxed text-ink-900/75 hover:bg-brand-50/70 dark:hover:bg-white/5"
          @click="choose(place)"
        >
          <MapPin :size="13" class="mt-0.5 shrink-0 text-brand-600" />
          <span>
            {{ place.address }}
            <span v-if="place.city" class="mt-0.5 block text-[9.5px] text-ink-900/40">{{ place.city }}<template v-if="place.state">, {{ place.state }}</template></span>
          </span>
        </button>
      </li>
    </ul>
    <p v-if="error" role="status" class="mt-1 text-[9.5px] text-red-600">{{ error }}</p>
  </div>
</template>
