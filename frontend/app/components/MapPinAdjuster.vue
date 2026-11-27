<script setup lang="ts">
import type { Map as MapLibreMap, Marker as MapLibreMarker } from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'

const props = withDefaults(defineProps<{
  latitude: number
  longitude: number
  styleUrl: string
  height?: string
  accent?: string
}>(), {
  height: '152px',
  accent: '#4f46e5',
})

const emit = defineEmits<{
  change: [coordinates: { latitude: number; longitude: number }]
}>()

const container = ref<HTMLElement | null>(null)
const message = ref('')
let map: MapLibreMap | null = null
let marker: MapLibreMarker | null = null
let cancelled = false

onMounted(async () => {
  if (!container.value || !props.styleUrl) return
  try {
    const maplibre = await import('maplibre-gl')
    if (cancelled || !container.value) return
    map = new maplibre.Map({
      container: container.value,
      style: props.styleUrl,
      center: [props.longitude, props.latitude],
      zoom: 15,
      scrollZoom: false,
      dragRotate: false,
      pitchWithRotate: false,
    })
    map.addControl(new maplibre.NavigationControl({ showCompass: false }), 'top-right')
    marker = new maplibre.Marker({ color: props.accent, draggable: true })
      .setLngLat([props.longitude, props.latitude])
      .addTo(map)
    marker.on('dragend', () => {
      const position = marker?.getLngLat()
      if (!position) return
      emit('change', { latitude: position.lat, longitude: position.lng })
    })
    // Tapping a compound, gate or side street is often easier than trying to
    // drag the pin precisely. Keep the same exact-coordinate behaviour.
    map.on('click', (event) => {
      const { lat, lng } = event.lngLat
      marker?.setLngLat([lng, lat])
      emit('change', { latitude: lat, longitude: lng })
    })
  } catch {
    message.value = 'The map preview could not load. You can still search for the address.'
  }
})

watch(
  () => [props.latitude, props.longitude] as const,
  ([latitude, longitude]) => {
    marker?.setLngLat([longitude, latitude])
    map?.easeTo({ center: [longitude, latitude], duration: 350 })
  },
)

onBeforeUnmount(() => {
  cancelled = true
  marker?.remove()
  map?.remove()
})
</script>

<template>
  <div class="relative overflow-hidden rounded-xl border border-slate-200 bg-slate-100 dark:border-white/10">
    <div ref="container" class="w-full" :style="{ height }" />
    <span class="pointer-events-none absolute right-2 bottom-2 left-2 flex justify-center">
      <span class="rounded-full bg-white/95 px-2.5 py-1 text-[9px] font-semibold text-slate-600 shadow-sm">
        Drag the pin to fine-tune the exact spot
      </span>
    </span>
    <p v-if="message" class="absolute inset-x-2 top-2 rounded-lg bg-white/95 px-2.5 py-1.5 text-[9.5px] leading-4 text-red-600 shadow-sm">
      {{ message }}
    </p>
  </div>
</template>
