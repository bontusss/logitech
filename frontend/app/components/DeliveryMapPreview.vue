<script setup lang="ts">
import type { Map as MapboxMap, Marker as MapboxMarker } from 'mapbox-gl'
import 'mapbox-gl/dist/mapbox-gl.css'

const props = defineProps<{
  pickupAddress: string
  dropoffAddress: string
}>()

const config = useRuntimeConfig()
const mapElement = ref<HTMLElement | null>(null)
const message = ref('')
const loading = ref(false)

let map: MapboxMap | null = null
let pickupMarker: MapboxMarker | null = null
let dropoffMarker: MapboxMarker | null = null
let mapboxModule: typeof import('mapbox-gl') | null = null
let geocodeTimer: ReturnType<typeof setTimeout> | undefined
let requestController: AbortController | null = null

const mapboxToken = config.public.mapboxAccessToken as string

onMounted(async () => {
  if (!mapboxToken) {
    message.value = 'Add NUXT_PUBLIC_MAPBOX_ACCESS_TOKEN to enable address lookup and the map preview.'
    return
  }
  if (!mapElement.value) return

  try {
    mapboxModule = await import('mapbox-gl')
    mapboxModule.default.accessToken = mapboxToken
    map = new mapboxModule.default.Map({
      container: mapElement.value,
      style: 'mapbox://styles/mapbox/streets-v12',
      center: [7.4951, 9.0579],
      zoom: 5,
      attributionControl: true,
    })
    pickupMarker = new mapboxModule.default.Marker({ color: '#2563eb' })
    dropoffMarker = new mapboxModule.default.Marker({ color: '#e11d48' })
    scheduleGeocode()
  } catch {
    message.value = 'Could not load the Mapbox map. Check your connection and try again.'
  }
})

onBeforeUnmount(() => {
  if (geocodeTimer) clearTimeout(geocodeTimer)
  requestController?.abort()
  pickupMarker?.remove()
  dropoffMarker?.remove()
  map?.remove()
})

watch(
  () => [props.pickupAddress, props.dropoffAddress],
  () => scheduleGeocode(),
)

function scheduleGeocode() {
  if (!map || !pickupMarker || !dropoffMarker) return
  if (geocodeTimer) clearTimeout(geocodeTimer)
  requestController?.abort()

  const pickup = props.pickupAddress.trim()
  const dropoff = props.dropoffAddress.trim()
  if (pickup.length < 5 || dropoff.length < 5) {
    loading.value = false
    message.value = 'Enter both addresses to show their locations on the map.'
    pickupMarker.remove()
    dropoffMarker.remove()
    return
  }

  geocodeTimer = setTimeout(() => {
    void geocodeAddresses(pickup, dropoff)
  }, 500)
}

async function geocodeAddresses(pickup: string, dropoff: string) {
  const controller = new AbortController()
  requestController = controller
  loading.value = true
  message.value = ''

  try {
    const [pickupCoordinates, dropoffCoordinates] = await Promise.all([
      geocode(pickup, controller.signal),
      geocode(dropoff, controller.signal),
    ])
    if (controller.signal.aborted || !map || !pickupMarker || !dropoffMarker) return

    pickupMarker.setLngLat(pickupCoordinates).addTo(map)
    dropoffMarker.setLngLat(dropoffCoordinates).addTo(map)
    if (!mapboxModule) return
    const bounds = new mapboxModule.default.LngLatBounds()
      .extend(pickupCoordinates)
      .extend(dropoffCoordinates)
    map.fitBounds(bounds, { padding: 48, maxZoom: 13, duration: 500 })
  } catch (error) {
    if (controller.signal.aborted) return
    message.value = error instanceof Error ? error.message : 'Could not locate these addresses.'
    pickupMarker?.remove()
    dropoffMarker?.remove()
  } finally {
    if (!controller.signal.aborted) loading.value = false
  }
}

async function geocode(address: string, signal: AbortSignal): Promise<[number, number]> {
  const params = new URLSearchParams({
    q: address,
    country: 'ng',
    limit: '1',
    access_token: mapboxToken,
  })
  const response = await fetch(`https://api.mapbox.com/search/geocode/v6/forward?${params}`, { signal })
  if (!response.ok) throw new Error('Mapbox address lookup failed. Check the token and try again.')
  const result = await response.json() as {
    features?: Array<{ geometry?: { coordinates?: [number, number] } }>
  }
  const coordinates = result.features?.[0]?.geometry?.coordinates
  if (!coordinates) throw new Error(`Could not find “${address}”. Try a more specific address.`)
  return coordinates
}
</script>

<template>
  <section class="overflow-hidden rounded-xl border border-ink-900/10 bg-white" aria-label="Delivery map preview">
    <div class="flex items-center justify-between border-b border-ink-900/8 px-3.5 py-2.5">
      <p class="text-[11px] font-semibold">Map preview</p>
      <span v-if="loading" class="text-[10px] text-ink-900/40">Finding addresses…</span>
      <span v-else class="text-[10px] text-ink-900/40">Pickup and drop-off</span>
    </div>
    <div ref="mapElement" class="h-52 bg-slate-100" />
    <p v-if="message" class="px-3.5 py-2.5 text-[10.5px] leading-relaxed text-ink-900/50">
      {{ message }}
    </p>
  </section>
</template>
