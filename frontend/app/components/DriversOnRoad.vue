<script setup lang="ts">
import { MapPin, RefreshCw, Satellite, Truck } from 'lucide-vue-next'
import type { LiveDriver } from '~/types/fleet'

const { api } = useApi()

const drivers = ref<LiveDriver[]>([])
const loading = ref(true)
const refreshing = ref(false)
const error = ref('')
const now = ref(Date.now())

let pollTimer: ReturnType<typeof setInterval> | null = null
let clockTimer: ReturnType<typeof setInterval> | null = null
let inFlight = false

const broadcastingCount = computed(() => drivers.value.filter((driver) => driver.location).length)

const STATUS_META: Record<string, { label: string; class: string }> = {
  assigned: { label: 'Assigned', class: 'bg-brand-50 text-brand-700 dark:bg-white/5 dark:text-brand-200' },
  picked_up: { label: 'Picked up', class: 'bg-accent-50 text-accent-700 dark:bg-accent-500/10 dark:text-accent-300' },
  in_transit: { label: 'In transit', class: 'bg-success-50 text-success-700 dark:bg-success-500/10 dark:text-success-300' },
}

function statusMeta(status: string) {
  return STATUS_META[status] ?? { label: status, class: 'bg-canvas text-ink-900/55' }
}

function initials(name: string) {
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')
}

function freshness(updatedAt: string) {
  const seconds = Math.max(0, Math.floor((now.value - new Date(updatedAt).getTime()) / 1000))
  if (seconds < 15) return 'just now'
  if (seconds < 60) return `${seconds}s ago`
  const minutes = Math.floor(seconds / 60)
  return `${minutes}m ago`
}

async function load(silent = false) {
  if (inFlight) return
  inFlight = true
  if (silent) {
    refreshing.value = true
  } else {
    loading.value = true
  }
  try {
    const result = await api<{ drivers: LiveDriver[] }>('/companies/me/drivers/live')
    drivers.value = result.drivers
    error.value = ''
  } catch (err) {
    if (!silent || drivers.value.length === 0) {
      error.value = apiErrorMessage(err, 'Could not load drivers on the road.')
    }
  } finally {
    inFlight = false
    loading.value = false
    refreshing.value = false
  }
}

function onVisibilityChange() {
  if (document.visibilityState === 'visible') void load(true)
}

onMounted(() => {
  void load()
  pollTimer = setInterval(() => {
    if (document.visibilityState === 'visible') void load(true)
  }, 20000)
  clockTimer = setInterval(() => {
    now.value = Date.now()
  }, 10000)
  document.addEventListener('visibilitychange', onVisibilityChange)
})

onBeforeUnmount(() => {
  if (pollTimer) clearInterval(pollTimer)
  if (clockTimer) clearInterval(clockTimer)
  document.removeEventListener('visibilitychange', onVisibilityChange)
})
</script>

<template>
  <section aria-labelledby="drivers-on-road-title">
    <div class="mb-3 flex flex-wrap items-end justify-between gap-2">
      <div>
        <h2 id="drivers-on-road-title" class="text-[14px] font-semibold">Drivers on the road</h2>
        <p class="mt-1 text-[11px] text-ink-900/40">Active deliveries and live GPS shared from the driver portal</p>
      </div>
      <div class="flex items-center gap-2">
        <span
          v-if="broadcastingCount > 0"
          class="inline-flex items-center gap-1.5 rounded-full border border-success-200 bg-success-50 px-2 py-1 text-[10px] font-semibold text-success-700"
        >
          <span class="relative flex size-1.5">
            <span class="absolute inline-flex size-full animate-ping rounded-full bg-success-500 opacity-60" />
            <span class="relative inline-flex size-1.5 rounded-full bg-success-500" />
          </span>
          {{ broadcastingCount }} live
        </span>
        <span v-else class="rounded-full bg-canvas px-2 py-1 text-[10px] font-medium text-ink-900/45">
          Waiting for GPS
        </span>
        <button
          type="button"
          class="rounded-md p-1.5 text-ink-900/40 transition hover:bg-brand-50 hover:text-brand-700 disabled:opacity-50"
          aria-label="Refresh drivers on the road"
          :disabled="refreshing"
          @click="load(true)"
        >
          <RefreshCw :size="14" :class="refreshing ? 'animate-spin' : ''" />
        </button>
      </div>
    </div>

    <div class="overview-card rounded-xl border p-3.5 sm:p-4">
      <div v-if="loading && drivers.length === 0" class="space-y-2" aria-hidden="true">
        <div v-for="index in 3" :key="index" class="h-[68px] animate-pulse rounded-xl bg-canvas" />
      </div>

      <div
        v-else-if="error && drivers.length === 0"
        role="alert"
        class="flex flex-col items-center gap-2 rounded-xl border border-red-200 bg-red-50 px-4 py-8 text-center"
      >
        <p class="text-[12px] font-medium text-red-700">{{ error }}</p>
        <button type="button" class="text-[11px] font-semibold text-red-700 underline underline-offset-2" @click="load()">
          Try again
        </button>
      </div>

      <div v-else-if="drivers.length === 0" class="flex flex-col items-center px-4 py-8 text-center">
        <IconTile tone="brand" size="lg" rounded="2xl">
          <Truck :size="19" />
        </IconTile>
        <p class="mt-3 text-[12.5px] font-semibold">No drivers are on the road</p>
        <p class="mt-1 max-w-xs text-[11px] leading-relaxed text-ink-900/45">
          Assign a driver to a delivery, then they can start sharing GPS from their rider portal link.
        </p>
        <NuxtLink
          to="/dashboard/deliveries"
          class="mt-3 inline-flex items-center gap-1 text-[11px] font-semibold text-brand-700 hover:text-brand-600"
        >
          Manage deliveries
        </NuxtLink>
      </div>

      <div v-else class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.35fr)]">
        <ul class="max-h-[360px] space-y-2 overflow-y-auto pr-0.5">
          <li
            v-for="driver in drivers"
            :key="driver.driver_id"
            class="rounded-xl border border-ink-900/8 bg-canvas/60 p-3"
          >
            <div class="flex items-start gap-2.5">
              <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-brand-600 text-[10.5px] font-semibold text-white">
                {{ initials(driver.full_name) }}
              </span>
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
                  <p class="truncate text-[12px] font-semibold">{{ driver.full_name }}</p>
                  <span class="rounded bg-brand-50 px-1.5 py-0.5 text-[9px] font-semibold tracking-wide text-brand-700 uppercase">
                    {{ driver.vehicle_plate }}
                  </span>
                  <span class="rounded-full px-1.5 py-0.5 text-[9px] font-semibold" :class="statusMeta(driver.delivery_status).class">
                    {{ statusMeta(driver.delivery_status).label }}
                  </span>
                </div>
                <p class="mt-1 flex items-center gap-1 text-[10.5px] text-ink-900/45">
                  <MapPin :size="11" class="shrink-0" />
                  <span class="truncate">To {{ driver.dropoff_address }}</span>
                </p>
                <p
                  v-if="driver.location"
                  class="mt-1.5 inline-flex items-center gap-1 text-[9.5px] font-medium text-success-700"
                >
                  <Satellite :size="10" />
                  Live GPS · {{ freshness(driver.location.updated_at) }}
                </p>
                <p v-else class="mt-1.5 inline-flex items-center gap-1 text-[9.5px] text-ink-900/35">
                  <span class="size-1.5 rounded-full bg-ink-900/20" />
                  No live GPS
                </p>
              </div>
            </div>
          </li>
        </ul>

        <FleetLiveMap :drivers="drivers" />
      </div>
    </div>
  </section>
</template>
