<script setup lang="ts">
import type { Map as MapLibreMap, Marker as MapLibreMarker } from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'
import { MapPin } from 'lucide-vue-next'
import type { LiveDriver } from '~/types/fleet'

const props = defineProps<{
  drivers: LiveDriver[]
}>()

const mapElement = ref<HTMLElement | null>(null)
const message = ref('')
const locatedCount = computed(() => props.drivers.filter((driver) => driver.location).length)

const { provider, loadError, load } = useMapProvider()

let maplibreModule: typeof import('maplibre-gl') | null = null
let map: MapLibreMap | null = null
let resizeObserver: ResizeObserver | null = null
let hasFitted = false
const markers = new Map<string, { marker: MapLibreMarker; status: string }>()

const STATUS_COLORS: Record<string, string> = {
  assigned: '#0B1F44',
  picked_up: '#FF7A00',
  in_transit: '#10B981',
}

onMounted(async () => {
  await load()
  const mapProvider = provider.value
  if (!mapProvider) {
    message.value = loadError.value || 'Live map is unavailable right now.'
    return
  }
  if (!mapElement.value) return

  try {
    maplibreModule = await import('maplibre-gl')
    map = new maplibreModule.Map({
      container: mapElement.value,
      style: mapProvider.styleUrl(),
      center: [8.6753, 9.082],
      zoom: 5.2,
    })
    map.addControl(new maplibreModule.NavigationControl({ showCompass: false }), 'top-right')
    map.on('load', syncMarkers)

    resizeObserver = new ResizeObserver(() => map?.resize())
    resizeObserver.observe(mapElement.value)
  } catch {
    message.value = 'Could not load the live map. Check your connection and try again.'
  }
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  markers.forEach(({ marker }) => marker.remove())
  markers.clear()
  map?.remove()
  map = null
})

watch(
  () => props.drivers,
  () => syncMarkers(),
  { deep: true },
)

function syncMarkers() {
  if (!map || !maplibreModule) return

  const activeIds = new Set<string>()
  for (const driver of props.drivers) {
    const location = driver.location
    if (!location) continue
    activeIds.add(driver.driver_id)

    const coordinates: [number, number] = [location.longitude, location.latitude]
    const existing = markers.get(driver.driver_id)
    if (existing && existing.status === driver.delivery_status) {
      existing.marker.setLngLat(coordinates)
      continue
    }
    if (existing) {
      existing.marker.remove()
      markers.delete(driver.driver_id)
    }
    markers.set(driver.driver_id, {
      marker: createMarker(driver, coordinates),
      status: driver.delivery_status,
    })
  }

  for (const [id, entry] of markers) {
    if (!activeIds.has(id)) {
      entry.marker.remove()
      markers.delete(id)
    }
  }

  if (!hasFitted && activeIds.size > 0) {
    fitToDrivers()
  }
  if (activeIds.size === 0) {
    hasFitted = false
  }
}

function createMarker(driver: LiveDriver, coordinates: [number, number]) {
  const element = document.createElement('div')
  element.style.cssText = 'display:flex;align-items:center;gap:4px;cursor:pointer'

  const dot = document.createElement('span')
  dot.style.cssText = `width:11px;height:11px;border-radius:9999px;background:${STATUS_COLORS[driver.delivery_status] ?? '#0B1F44'};border:2px solid #fff;box-shadow:0 0 0 3px rgba(11,31,68,0.16)`

  const label = document.createElement('span')
  label.textContent = driver.vehicle_plate
  label.style.cssText = 'font:600 9px/1 Switzer,system-ui,sans-serif;letter-spacing:0.02em;background:#0B1F44;color:#fff;padding:3px 5px;border-radius:5px;white-space:nowrap;box-shadow:0 1px 3px rgba(0,0,0,0.25)'

  element.append(dot, label)

  const marker = new maplibreModule!.Marker({ element, anchor: 'left' })
    .setLngLat(coordinates)
    .addTo(map!)

  const popup = new maplibreModule!.Popup({ offset: 18, closeButton: false }).setDOMContent(
    buildPopup(driver),
  )
  element.addEventListener('click', () => {
    popup.setLngLat(coordinates).addTo(map!)
  })

  return marker
}

function buildPopup(driver: LiveDriver) {
  const wrapper = document.createElement('div')
  wrapper.style.cssText = 'font-family:Switzer,system-ui,sans-serif;padding:2px 1px'

  const name = document.createElement('p')
  name.textContent = driver.full_name
  name.style.cssText = 'margin:0;font-size:12px;font-weight:600;color:#0B1F44'

  const meta = document.createElement('p')
  meta.textContent = `${driver.vehicle_plate} · ${driver.dropoff_address}`
  meta.style.cssText = 'margin:3px 0 0;font-size:10.5px;line-height:1.45;color:rgba(11,31,68,0.6);max-width:190px'

  wrapper.append(name, meta)
  return wrapper
}

function fitToDrivers() {
  if (!map || !maplibreModule) return
  const bounds = new maplibreModule.LngLatBounds()
  let hasCoordinates = false
  for (const driver of props.drivers) {
    if (!driver.location) continue
    bounds.extend([driver.location.longitude, driver.location.latitude])
    hasCoordinates = true
  }
  if (!hasCoordinates) return
  map.fitBounds(bounds, { padding: 56, maxZoom: 12, duration: 600 })
  hasFitted = true
}
</script>

<template>
  <div class="relative h-[260px] overflow-hidden rounded-xl border border-ink-900/10 bg-canvas sm:h-[320px]">
    <div ref="mapElement" class="absolute inset-0" aria-label="Live map of drivers on the road" />

    <div
      v-if="message"
      class="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-canvas px-6 text-center"
    >
      <MapPin :size="18" class="text-brand-500" />
      <p class="max-w-[260px] text-[11px] leading-relaxed text-ink-900/50">{{ message }}</p>
    </div>

    <div
      v-else-if="locatedCount === 0"
      class="pointer-events-none absolute inset-x-3 bottom-3 flex items-center justify-center gap-2 rounded-lg border border-ink-900/8 bg-surface/95 px-3 py-2 text-center shadow-sm backdrop-blur"
    >
      <span class="size-1.5 rounded-full bg-ink-900/20" />
      <p class="text-[10.5px] font-medium text-ink-900/55">No live GPS yet — drivers appear here when they start sharing.</p>
    </div>

    <span class="pointer-events-none absolute top-3 left-3 inline-flex items-center gap-1.5 rounded-full bg-surface/95 px-2 py-1 text-[9px] font-semibold tracking-wide text-ink-900/55 shadow-sm backdrop-blur">
      <span class="relative flex size-1.5">
        <span class="absolute inline-flex size-full animate-ping rounded-full bg-accent-500 opacity-60" />
        <span class="relative inline-flex size-1.5 rounded-full bg-accent-500" />
      </span>
      LIVE FLEET
    </span>
  </div>
</template>
