<script setup lang="ts">
import { AlertCircle, Loader2, MapPin, Navigation, Square, Truck, UserRound } from 'lucide-vue-next'

definePageMeta({ layout: false })
useSeoMeta({
  title: 'Driver portal',
  robots: 'noindex, nofollow, noarchive',
})
useHead({ meta: [{ name: 'referrer', content: 'no-referrer' }] })

const route = useRoute()
const { api } = useApi()

interface ActiveAssignment {
  id: string
  status: string
  tracking_code: string | null
  pickup_address: string
  dropoff_address: string
  updated_at: string
}

interface DriverPortal {
  driver: {
    full_name: string
    vehicle_plate: string
    company_name: string
  }
  active_assignment: ActiveAssignment | null
  location_tracking: string
}

const portal = ref<DriverPortal | null>(null)
const loading = ref(true)
const error = ref('')

const sharing = ref(false)
const starting = ref(false)
const shareError = ref('')
const lastSentAt = ref('')
let watchId: number | null = null
let lastPostAt = 0

const slug = computed(() => String(route.params.slug ?? ''))
const assignmentStatus = computed(() => {
  const status = portal.value?.active_assignment?.status
  if (status === 'picked_up') return { label: 'Picked up', class: 'bg-accent-50 text-accent-700' }
  if (status === 'in_transit') return { label: 'In transit', class: 'bg-success-50 text-success-700' }
  return { label: 'Assigned', class: 'bg-brand-50 text-brand-700' }
})

const lastSentLabel = computed(() =>
  lastSentAt.value
    ? new Date(lastSentAt.value).toLocaleTimeString('en-NG', { hour: '2-digit', minute: '2-digit' })
    : '',
)

onMounted(async () => {
  try {
    portal.value = await api<DriverPortal>(`/public/riders/${encodeURIComponent(slug.value)}`)
  } catch (err) {
    error.value = apiErrorMessage(err, 'This driver portal link is invalid or inactive.')
  } finally {
    loading.value = false
  }
})

onBeforeUnmount(() => {
  if (watchId !== null) {
    navigator.geolocation.clearWatch(watchId)
    watchId = null
  }
  if (sharing.value) {
    void api(`/public/riders/${encodeURIComponent(slug.value)}/location`, {
      method: 'DELETE',
      keepalive: true,
    }).catch(() => {})
  }
})

function startSharing() {
  if (!portal.value?.active_assignment) return
  if (!('geolocation' in navigator)) {
    shareError.value = 'This browser does not support location sharing. Open the portal in a recent version of Chrome or Safari.'
    return
  }

  starting.value = true
  shareError.value = ''

  watchId = navigator.geolocation.watchPosition(
    (position) => {
      if (!sharing.value) {
        sharing.value = true
        starting.value = false
      }
      void pushLocation(position)
    },
    (geoError) => {
      starting.value = false
      if (!sharing.value) {
        if (watchId !== null) {
          navigator.geolocation.clearWatch(watchId)
          watchId = null
        }
        shareError.value =
          geoError.code === geoError.PERMISSION_DENIED
            ? 'Location permission was denied. Allow location access for this site in your browser settings, then try again.'
            : 'Could not get a GPS fix. Move somewhere with a clearer view of the sky and try again.'
        return
      }
      shareError.value = 'Location updates paused. Check your connection or GPS signal — sharing resumes automatically.'
    },
    { enableHighAccuracy: true, maximumAge: 5000, timeout: 20000 },
  )
}

async function pushLocation(position: GeolocationPosition) {
  const now = Date.now()
  if (now - lastPostAt < 8000) return
  lastPostAt = now

  try {
    const result = await api<{ location: { updated_at: string } }>(
      `/public/riders/${encodeURIComponent(slug.value)}/location`,
      {
        method: 'POST',
        body: {
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
          accuracy_m: Number.isFinite(position.coords.accuracy) ? position.coords.accuracy : undefined,
          heading: Number.isFinite(position.coords.heading) ? position.coords.heading : undefined,
          speed_kph: Number.isFinite(position.coords.speed) && position.coords.speed !== null
            ? position.coords.speed * 3.6
            : undefined,
        },
      },
    )
    lastSentAt.value = result.location.updated_at
    shareError.value = ''
  } catch (err) {
    const code = (err as { data?: { error?: { code?: string } } })?.data?.error?.code
    if (code === 'no_active_delivery') {
      if (portal.value) portal.value.active_assignment = null
      void stopSharing(false)
      shareError.value = 'This delivery is no longer active, so location sharing stopped.'
      return
    }
    shareError.value = apiErrorMessage(err, 'Could not send your location. Check your connection and try again.')
  }
}

async function stopSharing(notifyServer = true) {
  if (watchId !== null) {
    navigator.geolocation.clearWatch(watchId)
    watchId = null
  }
  sharing.value = false
  starting.value = false
  lastSentAt.value = ''
  lastPostAt = 0

  if (notifyServer) {
    try {
      await api(`/public/riders/${encodeURIComponent(slug.value)}/location`, { method: 'DELETE' })
    } catch {
      // The dashboard stops showing the position within five minutes regardless.
    }
  }
}
</script>

<template>
  <main class="flex min-h-dvh items-center justify-center bg-canvas px-4 py-8 text-ink-900">
    <section class="w-full max-w-lg rounded-2xl border border-ink-900/10 bg-white p-6 shadow-sm sm:p-8">
      <AppLogo :size="34" />

      <div v-if="loading" class="mt-8 flex items-center gap-2 text-[13px] text-ink-900/50">
        <Loader2 :size="16" class="animate-spin" /> Opening your driver portal…
      </div>
      <div v-else-if="error" role="alert" class="mt-8 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 p-4 text-[12.5px] text-red-700">
        <AlertCircle :size="16" class="mt-0.5 shrink-0" /> {{ error }}
      </div>
      <template v-else-if="portal">
        <p class="mt-8 text-[10px] font-semibold tracking-[0.18em] text-brand-600 uppercase">Driver portal</p>
        <h1 class="mt-2 text-[24px] font-semibold tracking-[-0.035em]">Hello, {{ portal.driver.full_name }}</h1>
        <p class="mt-1 text-[13px] text-ink-900/50">{{ portal.driver.company_name }}</p>

        <div class="mt-6 grid gap-3 sm:grid-cols-2">
          <div class="rounded-xl border border-ink-900/10 bg-canvas p-4">
            <UserRound :size="16" class="text-brand-600" />
            <p class="mt-2 text-[10px] text-ink-900/40">Driver</p>
            <p class="mt-0.5 text-[12px] font-semibold">{{ portal.driver.full_name }}</p>
          </div>
          <div class="rounded-xl border border-ink-900/10 bg-canvas p-4">
            <Truck :size="16" class="text-brand-600" />
            <p class="mt-2 text-[10px] text-ink-900/40">Vehicle plate</p>
            <p class="mt-0.5 text-[12px] font-semibold uppercase">{{ portal.driver.vehicle_plate }}</p>
          </div>
        </div>

        <div v-if="portal.active_assignment" class="mt-5 rounded-xl border border-ink-900/10 bg-surface p-4">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="text-[10px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">Active delivery</p>
              <p class="mt-1.5 text-[12.5px] leading-relaxed font-semibold">
                {{ portal.active_assignment.pickup_address }}
                <span class="text-ink-900/35">→</span>
                {{ portal.active_assignment.dropoff_address }}
              </p>
              <p v-if="portal.active_assignment.tracking_code" class="mt-1 text-[10.5px] text-ink-900/45">
                Tracking {{ portal.active_assignment.tracking_code }}
              </p>
            </div>
            <span class="shrink-0 rounded-full px-2 py-1 text-[10px] font-semibold" :class="assignmentStatus.class">
              {{ assignmentStatus.label }}
            </span>
          </div>

          <div class="mt-4 border-t border-ink-900/8 pt-4">
            <p class="text-[11px] leading-relaxed text-ink-900/55">
              Your company only sees your live position while this delivery is active.
            </p>
            <button
              v-if="!sharing"
              type="button"
              :disabled="starting"
              class="mt-3 inline-flex min-h-10 w-full items-center justify-center gap-2 rounded-xl bg-accent-500 px-4 py-2.5 text-[12.5px] font-semibold text-white transition hover:bg-accent-400 disabled:opacity-60"
              @click="startSharing"
            >
              <Navigation :size="15" />
              {{ starting ? 'Waiting for GPS permission…' : 'Share live location' }}
            </button>
            <button
              v-else
              type="button"
              class="mt-3 inline-flex min-h-10 w-full items-center justify-center gap-2 rounded-xl border border-ink-900/15 bg-surface px-4 py-2.5 text-[12.5px] font-semibold transition hover:bg-brand-50"
              @click="stopSharing()"
            >
              <Square :size="13" />
              Stop sharing location
            </button>

            <p v-if="sharing" class="mt-2 flex items-center gap-1.5 text-[11px] font-medium text-success-700">
              <span class="relative flex size-1.5">
                <span class="absolute inline-flex size-full animate-ping rounded-full bg-success-500 opacity-70" />
                <span class="relative inline-flex size-1.5 rounded-full bg-success-500" />
              </span>
              Live{{ lastSentLabel ? ` · updated ${lastSentLabel}` : '' }}
            </p>
            <p v-if="shareError" role="alert" class="mt-2 rounded-lg border border-red-200 bg-red-50 px-2.5 py-2 text-[10.5px] leading-relaxed text-red-700">
              {{ shareError }}
            </p>
            <p v-if="sharing" class="mt-2 text-[10px] leading-relaxed text-ink-900/35">
              Keep this page open while sharing. Sharing stops when you close it or the delivery ends.
            </p>
          </div>
        </div>

        <div v-else class="mt-5 rounded-xl border border-brand-100 bg-brand-50/60 p-4">
          <div class="flex items-center gap-2 text-brand-700">
            <MapPin :size="16" />
            <h2 class="text-[12.5px] font-semibold">Location privacy</h2>
          </div>
          <p class="mt-2 text-[11.5px] leading-relaxed text-ink-900/60">
            {{ portal.location_tracking }}
          </p>
          <p class="mt-2 text-[10.5px] leading-relaxed text-ink-900/45">
            There are no active delivery assignments yet. Your browser will not request or share location until a delivery is assigned to you.
          </p>
        </div>

        <p class="mt-5 text-[10px] leading-relaxed text-ink-900/35">
          This private link identifies your portal. Don’t forward it. Contact {{ portal.driver.company_name }} if you think the link has been shared by mistake.
        </p>
      </template>
    </section>
  </main>
</template>
