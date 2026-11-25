<script setup lang="ts">
import {
  ArrowRight,
  CircleAlert,
  Loader2,
  MapPin,
  Pencil,
  Plus,
  Route,
  Save,
  Trash2,
  X,
} from 'lucide-vue-next'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Zones & Routes' })

interface GeoZone {
  id: string
  state: string
  lga: string
  name: string
}

interface ServiceZone {
  id: string
  name: string
  geo_zone_id: string
  state: string
  lga: string
  geo_name: string
  parent_zone_id: string | null
}

interface ZoneRoute {
  id: string
  from_zone_id: string
  from_zone_name: string
  from_state: string
  from_lga: string
  to_zone_id: string
  to_zone_name: string
  to_state: string
  to_lga: string
  price_minor: number
  included_weight_kg: number
  additional_weight_fee_minor: number
  status: 'active' | 'inactive'
}

const { api } = useApi()
const { can, confirm } = useWorkspace()

const loading = ref(true)
const savingZone = ref(false)
const savingRoute = ref(false)
const error = ref('')
const notice = ref('')
const states = ref<string[]>([])
const lgas = ref<string[]>([])
const geoZones = ref<GeoZone[]>([])
const zones = ref<ServiceZone[]>([])
const routes = ref<ZoneRoute[]>([])

const zoneForm = reactive({ state: '', lga: '', geoZoneId: '' })
const routeForm = reactive({
  id: '',
  from: '',
  to: '',
  basePrice: '',
  includedWeight: '5',
  additionalFee: '',
  status: 'active' as 'active' | 'inactive',
})

const canCreate = computed(() => can('rates.create'))
const canUpdate = computed(() => can('rates.update'))
const activeRoutes = computed(() => routes.value.filter((route) => route.status === 'active').length)
const sortedRoutes = computed(() =>
  [...routes.value].sort((a, b) =>
    `${a.from_state} ${a.from_lga} ${a.to_state} ${a.to_lga}`.localeCompare(
      `${b.from_state} ${b.from_lga} ${b.to_state} ${b.to_lga}`,
    ),
  ),
)
const zoneOptions = computed(() =>
  [...zones.value].sort((a, b) => `${a.state} ${a.lga}`.localeCompare(`${b.state} ${b.lga}`)),
)

const labelClass = 'mb-1.5 block text-[11px] font-semibold text-ink-900/60'
const fieldClass =
  'w-full rounded-xl border border-ink-900/12 bg-canvas px-3.5 py-2.5 text-[13px] font-medium text-ink-900 transition placeholder:text-ink-900/35 hover:border-ink-900/25 focus:border-accent-500 disabled:cursor-not-allowed disabled:opacity-60'

function zoneLabel(zone: ServiceZone) {
  return `${zone.lga}, ${zone.state}`
}

function nairaToMinor(value: string | number) {
  const amount = Number(value)
  if (!Number.isFinite(amount) || amount < 0) return null
  return Math.round(amount * 100)
}

function minorToNaira(value: number) {
  return (value / 100).toFixed(2).replace(/\.00$/, '')
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [statesResponse, zonesResponse, routesResponse] = await Promise.all([
      api<{ states: string[] }>('/companies/me/geo/states'),
      api<{ zones: ServiceZone[] }>('/companies/me/service-zones'),
      api<{ routes: ZoneRoute[] }>('/companies/me/zone-routes'),
    ])
    states.value = statesResponse.states ?? []
    zoneForm.state = zoneForm.state || states.value[0] || ''
    zones.value = zonesResponse.zones ?? []
    routes.value = routesResponse.routes ?? []
    if (zoneForm.state) await loadLgas(zoneForm.state)
    if (!routeForm.from) routeForm.from = zones.value[0]?.id ?? ''
    if (!routeForm.to) routeForm.to = zones.value[1]?.id ?? zones.value[0]?.id ?? ''
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not load your zones and routes.')
  } finally {
    loading.value = false
  }
}

async function loadLgas(state: string) {
  lgas.value = []
  geoZones.value = []
  zoneForm.lga = ''
  zoneForm.geoZoneId = ''
  if (!state) return
  try {
    const params = new URLSearchParams({ state })
    const response = await api<{ zones: GeoZone[] }>(`/companies/me/geo/zones?${params}`)
    geoZones.value = response.zones ?? []
    lgas.value = [...new Set(geoZones.value.map((zone) => zone.lga))]
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not load LGAs for that state.')
  }
}

function selectLga(lga: string) {
  zoneForm.lga = lga
  zoneForm.geoZoneId = geoZones.value.find((zone) => zone.lga === lga)?.id ?? ''
}

async function addZone() {
  error.value = ''
  notice.value = ''
  if (!zoneForm.geoZoneId) {
    error.value = 'Choose a state and LGA for this zone.'
    return
  }
  savingZone.value = true
  try {
    await api('/companies/me/service-zones', {
      method: 'POST',
      body: { geo_zone_id: zoneForm.geoZoneId },
    })
    notice.value = 'Service zone added.'
    const lga = zoneForm.lga
    zoneForm.lga = ''
    zoneForm.geoZoneId = ''
    const response = await api<{ zones: ServiceZone[] }>('/companies/me/service-zones')
    zones.value = response.zones ?? []
    if (!routeForm.from) routeForm.from = zones.value[0]?.id ?? ''
    if (!routeForm.to) routeForm.to = zones.value[0]?.id ?? ''
    if (!lga) await loadLgas(zoneForm.state)
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not add that zone.')
  } finally {
    savingZone.value = false
  }
}

async function removeZone(zone: ServiceZone) {
  if (!(await confirm({
    title: `Remove ${zoneLabel(zone)}?`,
    message: 'Routes that start or end in this zone will also be removed.',
    confirmLabel: 'Remove zone',
    danger: true,
  }))) return
  error.value = ''
  try {
    await api(`/companies/me/service-zones/${zone.id}`, { method: 'DELETE' })
    zones.value = zones.value.filter((item) => item.id !== zone.id)
    routes.value = routes.value.filter(
      (route) => route.from_zone_id !== zone.id && route.to_zone_id !== zone.id,
    )
    if (routeForm.from === zone.id) routeForm.from = zones.value[0]?.id ?? ''
    if (routeForm.to === zone.id) routeForm.to = zones.value[0]?.id ?? ''
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not remove that zone.')
  }
}

function resetRouteForm() {
  Object.assign(routeForm, {
    id: '',
    from: zones.value[0]?.id ?? '',
    to: zones.value[0]?.id ?? '',
    basePrice: '',
    includedWeight: '5',
    additionalFee: '',
    status: 'active',
  })
}

function editRoute(route: ZoneRoute) {
  Object.assign(routeForm, {
    id: route.id,
    from: route.from_zone_id,
    to: route.to_zone_id,
    basePrice: minorToNaira(route.price_minor),
    includedWeight: String(route.included_weight_kg),
    additionalFee: minorToNaira(route.additional_weight_fee_minor),
    status: route.status,
  })
  notice.value = ''
  error.value = ''
}

async function saveRoute() {
  error.value = ''
  notice.value = ''
  const price = nairaToMinor(routeForm.basePrice)
  const additional = nairaToMinor(routeForm.additionalFee || 0)
  const included = Number(routeForm.includedWeight)
  if (!routeForm.from || !routeForm.to) {
    error.value = 'Choose both the from and to zones.'
    return
  }
  if (price === null || additional === null || !Number.isFinite(included) || included < 0) {
    error.value = 'Enter a valid route price, included weight and additional weight fee.'
    return
  }
  savingRoute.value = true
  try {
    const body = {
      from_zone_id: routeForm.from,
      to_zone_id: routeForm.to,
      price_minor: price,
      included_weight_kg: included,
      additional_weight_fee_minor: additional,
      status: routeForm.status,
    }
    if (routeForm.id) {
      await api(`/companies/me/zone-routes/${routeForm.id}`, { method: 'PUT', body })
      notice.value = 'Route updated.'
    } else {
      await api('/companies/me/zone-routes', { method: 'POST', body })
      notice.value = 'Route added.'
    }
    const response = await api<{ routes: ZoneRoute[] }>('/companies/me/zone-routes')
    routes.value = response.routes ?? []
    resetRouteForm()
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not save that route.')
  } finally {
    savingRoute.value = false
  }
}

async function toggleRoute(route: ZoneRoute) {
  error.value = ''
  try {
    const status = route.status === 'active' ? 'inactive' : 'active'
    await api(`/companies/me/zone-routes/${route.id}/status`, {
      method: 'PATCH',
      body: { status },
    })
    route.status = status
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not update that route.')
  }
}

async function removeRoute(route: ZoneRoute) {
  if (!(await confirm({
    title: 'Delete this route?',
    message: `${route.from_lga}, ${route.from_state} → ${route.to_lga}, ${route.to_state} will stop appearing in the marketplace.`,
    confirmLabel: 'Delete route',
    danger: true,
  }))) return
  error.value = ''
  try {
    await api(`/companies/me/zone-routes/${route.id}`, { method: 'DELETE' })
    routes.value = routes.value.filter((item) => item.id !== route.id)
    if (routeForm.id === route.id) resetRouteForm()
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not delete that route.')
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-5">
    <header class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <p class="text-[10px] font-bold tracking-[0.18em] text-accent-700 uppercase dark:text-accent-400">Coverage & pricing</p>
        <h1 class="mt-1 text-[24px] font-semibold tracking-[-0.04em]">Zones & Routes</h1>
        <p class="mt-1 max-w-[640px] text-[13px] text-ink-900/60">
          Add the LGAs you serve, then publish a price for each direction you cover. The marketplace only matches these active routes.
        </p>
      </div>
      <div class="flex gap-2 text-[11px] font-semibold">
        <span class="rounded-full border border-ink-900/10 bg-white px-3 py-1.5">{{ zones.length }} zones</span>
        <span class="rounded-full border border-ink-900/10 bg-white px-3 py-1.5">{{ activeRoutes }} active routes</span>
      </div>
    </header>

    <p v-if="error" class="flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12.5px] font-medium text-red-700">
      <CircleAlert :size="15" class="mt-0.5 shrink-0" /> {{ error }}
    </p>
    <p v-if="notice" class="rounded-xl border border-success-200 bg-success-50 px-3.5 py-2.5 text-[12.5px] font-medium text-success-700">
      {{ notice }}
    </p>

    <div class="flex items-center gap-2 py-10 text-[13px] text-ink-900/50" v-if="loading">
      <Loader2 :size="16" class="animate-spin" /> Loading your coverage…
    </div>

    <div v-else class="grid gap-4 lg:grid-cols-[0.9fr_1.1fr]">
      <section class="rounded-2xl border border-ink-900/10 bg-white p-4">
        <div class="flex items-center gap-2">
          <span class="flex size-8 items-center justify-center rounded-xl bg-brand-100 text-brand-700"><MapPin :size="15" /></span>
          <div>
            <h2 class="text-[14px] font-semibold">Service zones</h2>
            <p class="text-[11px] text-ink-900/50">LGAs where you pick up or deliver.</p>
          </div>
        </div>

        <div class="mt-4 grid gap-2 sm:grid-cols-2">
          <label>
            <span :class="labelClass">State</span>
            <select v-model="zoneForm.state" :class="fieldClass" @change="loadLgas(zoneForm.state)">
              <option value="" disabled>Choose a state</option>
              <option v-for="state in states" :key="state" :value="state">{{ state }}</option>
            </select>
          </label>
          <label>
            <span :class="labelClass">LGA</span>
            <select :value="zoneForm.lga" :class="fieldClass" :disabled="!lgas.length" @change="selectLga(($event.target as HTMLSelectElement).value)">
              <option value="" disabled>Choose an LGA</option>
              <option v-for="lga in lgas" :key="lga" :value="lga">{{ lga }}</option>
            </select>
          </label>
        </div>
        <button
          type="button"
          class="mt-3 inline-flex w-full items-center justify-center gap-1.5 rounded-xl bg-ink-900 px-3.5 py-2.5 text-[12px] font-semibold text-white transition hover:bg-ink-900/90 disabled:opacity-50"
          :disabled="!canCreate || savingZone || !zoneForm.geoZoneId"
          @click="addZone"
        >
          <Loader2 v-if="savingZone" :size="14" class="animate-spin" />
          <Plus v-else :size="14" />
          Add zone
        </button>

        <div v-if="!zones.length" class="mt-4 rounded-xl border border-dashed border-ink-900/15 px-3 py-6 text-center text-[12px] text-ink-900/45">
          No zones yet. Add your first LGA to start publishing routes.
        </div>
        <ul v-else class="mt-4 space-y-2">
          <li v-for="zone in zoneOptions" :key="zone.id" class="flex items-center justify-between gap-2 rounded-xl border border-ink-900/8 bg-canvas/60 px-3 py-2.5">
            <div class="min-w-0">
              <p class="truncate text-[12.5px] font-semibold">{{ zone.lga }}</p>
              <p class="text-[10.5px] text-ink-900/45">{{ zone.state }}</p>
            </div>
            <button
              type="button"
              class="flex size-7 shrink-0 items-center justify-center rounded-lg text-ink-900/40 transition hover:bg-red-50 hover:text-red-600 disabled:opacity-30"
              :disabled="!canUpdate"
              aria-label="Remove zone"
              @click="removeZone(zone)"
            >
              <Trash2 :size="13" />
            </button>
          </li>
        </ul>
      </section>

      <section class="rounded-2xl border border-ink-900/10 bg-white p-4">
        <div class="flex items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <span class="flex size-8 items-center justify-center rounded-xl bg-accent-100 text-accent-800"><Route :size="15" /></span>
            <div>
              <h2 class="text-[14px] font-semibold">{{ routeForm.id ? 'Edit route' : 'New directional route' }}</h2>
              <p class="text-[11px] text-ink-900/50">Reverse routes are priced separately.</p>
            </div>
          </div>
          <button v-if="routeForm.id" type="button" class="flex size-7 items-center justify-center rounded-lg text-ink-900/45 hover:bg-canvas" aria-label="Cancel editing" @click="resetRouteForm">
            <X :size="14" />
          </button>
        </div>

        <div class="mt-4 grid gap-2.5 sm:grid-cols-[1fr_auto_1fr] sm:items-end">
          <label>
            <span :class="labelClass">From zone</span>
            <select v-model="routeForm.from" :class="fieldClass">
              <option value="" disabled>From</option>
              <option v-for="zone in zoneOptions" :key="zone.id" :value="zone.id">{{ zoneLabel(zone) }}</option>
            </select>
          </label>
          <span class="hidden pb-2.5 text-ink-900/30 sm:block"><ArrowRight :size="15" /></span>
          <label>
            <span :class="labelClass">To zone</span>
            <select v-model="routeForm.to" :class="fieldClass">
              <option value="" disabled>To</option>
              <option v-for="zone in zoneOptions" :key="zone.id" :value="zone.id">{{ zoneLabel(zone) }}</option>
            </select>
          </label>
        </div>

        <div class="mt-2.5 grid gap-2.5 sm:grid-cols-3">
          <label>
            <span :class="labelClass">Base price (₦)</span>
            <input v-model="routeForm.basePrice" :class="fieldClass" type="number" min="0" step="0.01" placeholder="5000" />
          </label>
          <label>
            <span :class="labelClass">Included weight (kg)</span>
            <input v-model="routeForm.includedWeight" :class="fieldClass" type="number" min="0" max="500" step="0.1" placeholder="5" />
          </label>
          <label>
            <span :class="labelClass">Extra per kg (₦)</span>
            <input v-model="routeForm.additionalFee" :class="fieldClass" type="number" min="0" step="0.01" placeholder="250" />
          </label>
        </div>

        <div class="mt-2.5 flex flex-wrap items-end gap-2.5">
          <label class="min-w-[140px] flex-1">
            <span :class="labelClass">Status</span>
            <select v-model="routeForm.status" :class="fieldClass">
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </select>
          </label>
          <button
            type="button"
            class="inline-flex items-center justify-center gap-1.5 rounded-xl bg-accent-500 px-4 py-2.5 text-[12px] font-semibold text-ink-950 transition hover:bg-accent-400 disabled:opacity-50"
            :disabled="!canCreate || savingRoute"
            @click="saveRoute"
          >
            <Loader2 v-if="savingRoute" :size="14" class="animate-spin" />
            <Save v-else :size="14" />
            {{ routeForm.id ? 'Update route' : 'Add route' }}
          </button>
        </div>
      </section>
    </div>

    <section v-if="!loading" class="overflow-hidden rounded-2xl border border-ink-900/10 bg-white">
      <div class="flex items-center justify-between gap-2 border-b border-ink-900/8 px-4 py-3">
        <div>
          <h2 class="text-[14px] font-semibold">Your routes</h2>
          <p class="text-[11px] text-ink-900/50">Customers see only active routes in the matching direction.</p>
        </div>
        <span class="rounded-full bg-canvas px-2.5 py-1 text-[10.5px] font-semibold text-ink-900/55">{{ routes.length }}</span>
      </div>

      <div v-if="!sortedRoutes.length" class="px-4 py-10 text-center text-[12.5px] text-ink-900/45">
        No routes yet. Add at least one direction to appear in marketplace searches.
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full min-w-[680px] text-left text-[12px]">
          <thead class="bg-canvas/60 text-[10px] font-bold tracking-[0.08em] text-ink-900/45 uppercase">
            <tr>
              <th class="px-4 py-2.5">Direction</th>
              <th class="px-4 py-2.5">Base price</th>
              <th class="px-4 py-2.5">Weight</th>
              <th class="px-4 py-2.5">Status</th>
              <th class="px-4 py-2.5 text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-ink-900/6">
            <tr v-for="route in sortedRoutes" :key="route.id" class="align-middle">
              <td class="px-4 py-3">
                <p class="font-semibold">{{ route.from_lga }} <span class="px-1 text-ink-900/30">→</span> {{ route.to_lga }}</p>
                <p class="text-[10.5px] text-ink-900/45">{{ route.from_state }} → {{ route.to_state }}</p>
              </td>
              <td class="px-4 py-3 font-semibold tabular-nums">{{ formatMinor(route.price_minor) }}</td>
              <td class="px-4 py-3 text-ink-900/65">
                {{ route.included_weight_kg }} kg included
                <span class="block text-[10.5px] text-ink-900/45">{{ formatMinor(route.additional_weight_fee_minor) }}/extra kg</span>
              </td>
              <td class="px-4 py-3">
                <button
                  type="button"
                  class="rounded-full px-2 py-0.5 text-[10px] font-semibold disabled:opacity-50"
                  :class="route.status === 'active' ? 'bg-success-100 text-success-700' : 'bg-canvas text-ink-900/50'"
                  :disabled="!canUpdate"
                  @click="toggleRoute(route)"
                >
                  {{ route.status }}
                </button>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center justify-end gap-1">
                  <button type="button" class="flex size-7 items-center justify-center rounded-lg text-ink-900/45 hover:bg-canvas hover:text-ink-900 disabled:opacity-30" :disabled="!canUpdate" aria-label="Edit route" @click="editRoute(route)">
                    <Pencil :size="13" />
                  </button>
                  <button type="button" class="flex size-7 items-center justify-center rounded-lg text-ink-900/40 hover:bg-red-50 hover:text-red-600 disabled:opacity-30" :disabled="!canUpdate" aria-label="Delete route" @click="removeRoute(route)">
                    <Trash2 :size="13" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>
