<script setup lang="ts">
import {
  ArrowLeftRight,
  ArrowRight,
  Loader2,
  PackageCheck,
  Search,
} from 'lucide-vue-next'
import type { MarketplaceSearchPayload } from '~/types/marketplace'
import { ITEM_CATEGORIES } from '~/utils/itemCategories'

withDefaults(defineProps<{ loading?: boolean }>(), { loading: false })

const emit = defineEmits<{ search: [payload: MarketplaceSearchPayload] }>()

interface GeoZone {
  id: string
  state: string
  lga: string
  name: string
}

const { api } = useApi()
const zones = ref<GeoZone[]>([])
const reachableZones = ref<GeoZone[]>([])
const zonesLoading = ref(true)
const reachableLoading = ref(false)
const zoneError = ref('')
const error = ref('')

const form = reactive({
  originZoneId: '',
  destinationZoneId: '',
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
const selectClass =
  'w-full rounded-xl border border-ink-900/12 bg-canvas px-3 py-2.5 text-[12.5px] font-medium text-ink-900 transition hover:border-ink-900/25 focus:border-accent-500 disabled:cursor-not-allowed disabled:opacity-60'
const labelClass = 'mb-1.5 block text-[11px] font-semibold text-ink-900/55'

function zoneLabel(zone: GeoZone) {
  return `${zone.lga}, ${zone.state}`
}

function groupZones(list: GeoZone[]) {
  const grouped = new Map<string, GeoZone[]>()
  for (const zone of list) {
    const entries = grouped.get(zone.state) ?? []
    entries.push(zone)
    grouped.set(zone.state, entries)
  }
  return [...grouped.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([state, entries]) => ({
      state,
      zones: entries.sort((a, b) => a.lga.localeCompare(b.lga)),
    }))
}

const originGroups = computed(() => groupZones(zones.value))
const destinationGroups = computed(() => groupZones(reachableZones.value))

onMounted(async () => {
  try {
    const response = await api<{ zones: GeoZone[] }>('/public/zones')
    zones.value = response.zones ?? []
    if (zones.value.length) {
      form.originZoneId = zones.value[0]!.id
      await loadReachable()
    }
  } catch (cause) {
    zoneError.value = apiErrorMessage(cause, 'Could not load service zones.')
  } finally {
    zonesLoading.value = false
  }
})

async function loadReachable() {
  reachableZones.value = []
  form.destinationZoneId = ''
  if (!form.originZoneId) return
  reachableLoading.value = true
  zoneError.value = ''
  try {
    const params = new URLSearchParams({ origin_zone_id: form.originZoneId })
    const response = await api<{ zones: GeoZone[] }>(`/public/zones/reachable?${params}`)
    reachableZones.value = response.zones ?? []
    const current = reachableZones.value.find((zone) => zone.id === form.destinationZoneId)
    form.destinationZoneId = current?.id ?? reachableZones.value[0]?.id ?? ''
  } catch (cause) {
    zoneError.value = apiErrorMessage(cause, 'Could not load destination zones for that pickup zone.')
  } finally {
    reachableLoading.value = false
  }
}

function swapZones() {
  const previousOrigin = form.originZoneId
  const previousDestination = form.destinationZoneId
  if (!previousDestination) return
  form.originZoneId = previousDestination
  void loadReachable().then(() => {
    if (reachableZones.value.some((zone) => zone.id === previousOrigin)) {
      form.destinationZoneId = previousOrigin
    }
  })
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
  if (!form.originZoneId) {
    error.value = 'Choose a pickup zone.'
    return
  }
  if (!form.destinationZoneId) {
    error.value = 'Choose a destination zone with an active provider route.'
    return
  }
  if (!validatePackage()) return
  const origin = zones.value.find((zone) => zone.id === form.originZoneId)
  const destination = reachableZones.value.find((zone) => zone.id === form.destinationZoneId)
  emit('search', {
    origin_zone_id: form.originZoneId,
    destination_zone_id: form.destinationZoneId,
    origin_label: origin ? zoneLabel(origin) : undefined,
    destination_label: destination ? zoneLabel(destination) : undefined,
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
          Choose the pickup zone, then a destination served by a provider route.
        </p>
      </div>
      <IconTile tone="accent" size="lg" rounded="2xl">
        <PackageCheck :size="20" />
      </IconTile>
    </div>

    <form class="mt-5 space-y-3.5" @submit.prevent="submitSearch">
      <div class="grid grid-cols-[1fr_auto_1fr] items-end gap-2">
        <div class="min-w-0">
          <span :class="labelClass">Pickup zone</span>
          <select
            v-model="form.originZoneId"
            :class="selectClass"
            name="origin_zone_id"
            aria-label="Pickup zone"
            :disabled="zonesLoading || !zones.length"
            required
            @change="loadReachable"
          >
            <option value="" disabled>{{ zonesLoading ? 'Loading zones…' : 'Choose a zone' }}</option>
            <optgroup v-for="group in originGroups" :key="group.state" :label="group.state">
              <option v-for="zone in group.zones" :key="zone.id" :value="zone.id">
                {{ zone.lga }}
              </option>
            </optgroup>
          </select>
        </div>
        <button
          type="button"
          class="mb-1 flex size-9 shrink-0 items-center justify-center rounded-xl border border-ink-900/10 bg-canvas text-ink-900/50 transition hover:-translate-y-0.5 hover:border-accent-500/40 hover:text-accent-700 disabled:opacity-40 dark:hover:text-accent-400"
          aria-label="Swap pickup and destination zones"
          :disabled="!form.destinationZoneId"
          @click="swapZones"
        >
          <ArrowLeftRight :size="15" />
        </button>
        <div class="min-w-0">
          <span :class="labelClass">Destination zone</span>
          <select
            v-model="form.destinationZoneId"
            :class="selectClass"
            name="destination_zone_id"
            aria-label="Destination zone"
            :disabled="reachableLoading || !reachableZones.length"
            required
          >
            <option value="" disabled>
              {{ reachableLoading ? 'Finding routes…' : reachableZones.length ? 'Choose a zone' : 'No routes yet' }}
            </option>
            <optgroup v-for="group in destinationGroups" :key="group.state" :label="group.state">
              <option v-for="zone in group.zones" :key="zone.id" :value="zone.id">
                {{ zone.lga }}
              </option>
            </optgroup>
          </select>
        </div>
      </div>

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
        :disabled="loading || zonesLoading || !form.destinationZoneId"
        class="flex w-full items-center justify-center gap-2 rounded-xl bg-accent-500 px-4 py-3.5 text-[14px] font-semibold text-ink-950 shadow-[0_14px_28px_-16px_rgba(255,122,0,0.75)] transition hover:-translate-y-0.5 hover:bg-accent-400 hover:shadow-[0_18px_32px_-16px_rgba(255,122,0,0.85)] disabled:translate-y-0 disabled:opacity-60"
      >
        <Loader2 v-if="loading" :size="16" class="animate-spin" />
        <Search v-else :size="16" />
        {{ loading ? 'Checking providers…' : 'Compare provider routes' }}
        <ArrowRight :size="15" class="ml-auto" />
      </button>
    </form>

    <p class="mt-3 text-center text-[10.5px] leading-relaxed text-ink-900/40">
      Only providers with an active route in that direction are shown. No account needed.
    </p>
  </div>
</template>
