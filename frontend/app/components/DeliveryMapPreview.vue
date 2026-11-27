<script setup lang="ts">
import type { Map as MapLibreMap, Marker as MapLibreMarker } from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'
import type { MapProvider } from '~/utils/maps'

const props = defineProps<{
  pickupAddress: string
  dropoffAddress: string
}>()

const mapElement = ref<HTMLElement | null>(null)
const message = ref('')
const loading = ref(false)

let map: MapLibreMap | null = null
let pickupMarker: MapLibreMarker | null = null
let dropoffMarker: MapLibreMarker | null = null
let maplibreModule: typeof import('maplibre-gl') | null = null
let activeProvider: MapProvider | null = null
let geocodeTimer: ReturnType<typeof setTimeout> | undefined
let requestController: AbortController | null = null

const { provider, loadError: mapConfigError, load: loadMapProvider } = useMapProvider()

onMounted(async () => {
  await loadMapProvider()
  activeProvider = provider.value
  if (!activeProvider) {
    message.value = mapConfigError.value || 'Map previews are not configured for this site yet.'
    return
  }
  if (!mapElement.value) return

  try {
    maplibreModule = await import('maplibre-gl')
    map = new maplibreModule.Map({
      container: mapElement.value,
      style: activeProvider.styleUrl(),
      center: [7.4951, 9.0579],
      zoom: 5,
    })
    pickupMarker = new maplibreModule.Marker({ color: '#2563eb' })
    dropoffMarker = new maplibreModule.Marker({ color: '#e11d48' })
    scheduleGeocode()
  } catch {
    message.value = 'Could not load the map. Check your connection and try again.'
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
  if (!activeProvider) return
  const controller = new AbortController()
  requestController = controller
  loading.value = true
  message.value = ''

  try {
    const [pickupPlace, dropoffPlace] = await Promise.all([
      activeProvider.geocode(pickup, { signal: controller.signal }),
      activeProvider.geocode(dropoff, { signal: controller.signal }),
    ])
    if (controller.signal.aborted || !map || !pickupMarker || !dropoffMarker) return
    if (!pickupPlace || !dropoffPlace) {
      throw new Error('Could not find one of these addresses. Try a more specific address.')
    }

    const pickupCoordinates: [number, number] = [pickupPlace.longitude, pickupPlace.latitude]
    const dropoffCoordinates: [number, number] = [dropoffPlace.longitude, dropoffPlace.latitude]
    pickupMarker.setLngLat(pickupCoordinates).addTo(map)
    dropoffMarker.setLngLat(dropoffCoordinates).addTo(map)
    if (!maplibreModule) return
    const bounds = new maplibreModule.LngLatBounds()
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
