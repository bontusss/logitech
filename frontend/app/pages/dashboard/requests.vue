<script setup lang="ts">
import {
  ArrowDownLeft,
  ArrowUpRight,
  Check,
  CheckCircle2,
  CircleAlert,
  Clock3,
  Loader2,
  Search,
  Sparkles,
  Truck,
  X,
  XCircle,
} from 'lucide-vue-next'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Requests' })

interface RequestItem {
  id: string
  created_at: string
  sender_name: string
  sender_phone: string
  receiver_name: string
  receiver_phone: string
  pickup_address: string
  dropoff_address: string
  item_description: string
  package_category?: string
  weight_kg?: number
  delivery_price_minor: number
  price_is_estimate?: boolean
  request_status: 'pending' | 'accepted' | 'declined'
}
interface Driver { id: string; full_name: string; phone: string; vehicle_plate: string; status: 'active' | 'suspended' }

const { api } = useApi()
const { detail, load, can, notify, confirm } = useWorkspace()
const requests = ref<RequestItem[]>([])
const drivers = ref<Driver[]>([])
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const searchQuery = ref('')
const filter = ref<'pending' | 'accepted' | 'declined'>('pending')
const acceptOpen = ref(false)
const accepting = ref<RequestItem | null>(null)
const selectedDriver = ref('')
const currency = computed(() => detail.value?.settings.currency || 'NGN')

onMounted(async () => { await load(); await Promise.all([fetchRequests(), fetchDrivers()]) })

async function fetchRequests() {
  loading.value=true
  try { requests.value=(await api<{requests:RequestItem[]}>('/companies/me/requests')).requests }
  catch(err){ error.value=apiErrorMessage(err,'Could not load booking requests.') }
  finally{loading.value=false}
}
async function fetchDrivers(){
  try{drivers.value=(await api<{drivers:Driver[]}>('/companies/me/drivers')).drivers}catch{drivers.value=[]}
}
const filtered = computed(() => {
  const q=searchQuery.value.trim().toLowerCase()
  return requests.value.filter(r => r.request_status === filter.value && (!q || [r.sender_name,r.sender_phone,r.receiver_name,r.receiver_phone,r.pickup_address,r.dropoff_address,r.package_category,r.item_description].some(v => (v||'').toLowerCase().includes(q))))
})
const counts=computed(()=>({pending:requests.value.filter(r=>r.request_status==='pending').length,accepted:requests.value.filter(r=>r.request_status==='accepted').length,declined:requests.value.filter(r=>r.request_status==='declined').length}))

function formatMoney(minor:number){return new Intl.NumberFormat('en-NG',{style:'currency',currency:currency.value,maximumFractionDigits:2}).format(minor/100)}
function openAccept(request:RequestItem){accepting.value=request;selectedDriver.value='';acceptOpen.value=true}
async function acceptRequest(){
  if(!accepting.value||!selectedDriver.value){error.value='Choose a driver to accept this request.';return}
  saving.value=true
  try{
    await api(`/companies/me/requests/${accepting.value.id}/accept`,{method:'POST',body:{driver_id:selectedDriver.value}})
    acceptOpen.value=false;notify('Request accepted and delivery created.');await fetchRequests();filter.value='accepted'
  }catch(err){error.value=apiErrorMessage(err,'Could not accept this request.')}
  finally{saving.value=false}
}
async function declineRequest(request:RequestItem){
  const approved=await confirm({title:'Decline request?',message:`Decline the request from ${request.sender_name}?`,confirmLabel:'Decline request',danger:true});if(!approved)return
  try{await api(`/companies/me/requests/${request.id}/decline`,{method:'POST'});notify('Request declined.');await fetchRequests()}catch(err){error.value=apiErrorMessage(err,'Could not decline this request.')}
}
function openDelivery(id:string){navigateTo(`/dashboard/deliveries?id=${encodeURIComponent(id)}`)}
</script>

<template>
  <div class="dash-page">
    <header class="dash-hero relative overflow-hidden rounded-2xl border p-5 sm:p-6">
      <div class="pointer-events-none absolute -top-16 right-10 size-44 rounded-full bg-fuchsia-400/15 blur-3xl" />
      <div class="pointer-events-none absolute -right-8 -bottom-24 size-52 rounded-full bg-cyan-400/15 blur-3xl" />
      <div class="relative flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="dash-eyebrow inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-semibold tracking-wide">
            <Sparkles :size="12" /> BOOKING ACTIVITY
          </p>
          <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Requests</h1>
          <p class="mt-1.5 max-w-xl text-[12px] leading-relaxed text-ink-900/65">
            Review bookings submitted through your public booking page.
          </p>
        </div>
        <span class="inline-flex items-center gap-1.5 rounded-full border border-white/70 bg-white/70 px-3 py-1.5 text-[11px] font-semibold text-ink-900/60 dark:border-white/10 dark:bg-white/5">
          <Clock3 :size="13" /> Pending first
        </span>
      </div>
    </header>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ error }}
    </p>

    <div class="mt-4 grid grid-cols-3 gap-2 sm:gap-3">
      <button
        v-for="tab in ([
          { key: 'pending', label: 'Pending', icon: Clock3, tone: 'amber' },
          { key: 'accepted', label: 'Accepted', icon: CheckCircle2, tone: 'mint' },
          { key: 'declined', label: 'Declined', icon: XCircle, tone: 'rose' },
        ] as const)"
        :key="tab.key"
        type="button"
        class="dash-stat rounded-xl border p-3.5 text-left sm:p-4"
        :class="[`dash-stat--${tab.tone}`, filter === tab.key ? 'dash-stat--active' : '']"
        @click="filter = tab.key"
      >
        <span class="flex items-center justify-between gap-2">
          <span class="text-[10px] font-semibold tracking-wide text-ink-900/50 uppercase">{{ tab.label }}</span>
          <span class="flex size-6 items-center justify-center rounded-lg bg-white/80 text-ink-900/45 dark:bg-white/10">
            <component :is="tab.icon" :size="12" />
          </span>
        </span>
        <span class="mt-2 block text-[22px] leading-none font-semibold tracking-tight">{{ counts[tab.key] }}</span>
      </button>
    </div>

    <section class="dash-panel mt-4 overflow-hidden rounded-2xl border">
      <div class="flex flex-col gap-3 border-b border-ink-900/8 p-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-[13px] font-semibold capitalize">{{ filter }} requests</h2>
          <p class="mt-0.5 text-[10.5px] text-ink-900/40">{{ filtered.length }} shown</p>
        </div>
        <label class="relative block w-full sm:max-w-xs">
          <Search :size="15" class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-ink-900/35" />
          <input
            v-model="searchQuery"
            type="search"
            placeholder="Search name, phone or route"
            class="w-full rounded-lg border border-ink-900/10 py-2.5 pr-3 pl-9 text-[11.5px] outline-none focus:border-brand-500"
          />
        </label>
      </div>
      <div v-if="loading" class="flex items-center justify-center gap-2 p-12 text-[12px] text-ink-900/45">
        <Loader2 :size="15" class="animate-spin" /> Loading requests…
      </div>
      <div v-else-if="!filtered.length" class="p-10 text-center">
        <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-100 to-cyan-50 text-brand-700 dark:from-violet-400/15 dark:to-cyan-400/10 dark:text-indigo-200">
          <Truck :size="19" />
        </span>
        <h3 class="mt-3 text-[13px] font-semibold">No {{ filter }} requests</h3>
        <p class="mt-1 text-[11.5px] text-ink-900/45">
          {{ searchQuery ? 'Try another search.' : 'Requests submitted through your booking page will appear here.' }}
        </p>
      </div>
      <div v-else class="divide-y divide-ink-900/8">
        <article v-for="request in filtered" :key="request.id" class="p-4 transition hover:bg-brand-50/20 sm:p-5">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="flex min-w-0 items-center gap-2.5">
              <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-violet-50 text-violet-700 dark:bg-violet-400/15 dark:text-violet-200">
                <Truck :size="16" />
              </span>
              <div class="min-w-0">
                <p class="text-[10px] text-ink-900/40">{{ new Date(request.created_at).toLocaleString() }}</p>
                <p class="mt-0.5 truncate text-[12.5px] font-semibold">
                  {{ request.sender_name }} <span class="text-ink-900/30">→</span> {{ request.receiver_name }}
                </p>
              </div>
            </div>
            <span
              class="rounded-full px-2.5 py-1 text-[9px] font-semibold capitalize"
              :class="request.request_status === 'pending'
                ? 'bg-amber-50 text-amber-700 dark:bg-amber-400/10 dark:text-amber-200'
                : request.request_status === 'accepted'
                  ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-400/10 dark:text-emerald-200'
                  : 'bg-slate-100 text-slate-600 dark:bg-white/10 dark:text-slate-300'"
            >
              {{ request.request_status }}
            </span>
          </div>
          <div class="mt-4 grid gap-3 text-[11px] sm:grid-cols-2">
            <div class="flex gap-2 rounded-xl bg-cyan-50/50 p-2.5 dark:bg-cyan-400/5">
              <ArrowUpRight :size="14" class="mt-0.5 shrink-0 text-cyan-700 dark:text-cyan-200" />
              <div class="min-w-0">
                <p class="font-medium">{{ request.sender_name }} · {{ request.sender_phone }}</p>
                <p class="mt-0.5 break-words text-ink-900/45">{{ request.pickup_address }}</p>
              </div>
            </div>
            <div class="flex gap-2 rounded-xl bg-rose-50/50 p-2.5 dark:bg-rose-400/5">
              <ArrowDownLeft :size="14" class="mt-0.5 shrink-0 text-rose-600" />
              <div class="min-w-0">
                <p class="font-medium">{{ request.receiver_name }} · {{ request.receiver_phone }}</p>
                <p class="mt-0.5 break-words text-ink-900/45">{{ request.dropoff_address }}</p>
              </div>
            </div>
          </div>
          <div class="mt-3 flex flex-wrap items-center justify-between gap-3 border-t border-ink-900/8 pt-3">
            <div>
              <p class="text-[10.5px] text-ink-900/55">
                {{ request.package_category || request.item_description }}
                <template v-if="request.weight_kg"> · {{ request.weight_kg }} kg</template>
              </p>
              <p class="mt-1 text-[13px] font-semibold">
                {{ request.delivery_price_minor ? formatMoney(request.delivery_price_minor) : 'Price to confirm' }}
              </p>
            </div>
            <div v-if="request.request_status === 'pending'" class="flex gap-2">
              <button
                type="button"
                class="rounded-lg border border-rose-200 px-3 py-2 text-[10.5px] font-semibold text-rose-700 transition hover:bg-rose-50"
                @click="declineRequest(request)"
              >
                Decline
              </button>
              <button
                v-if="can('bookings.manage')"
                type="button"
                class="inline-flex items-center gap-1 rounded-lg bg-brand-600 px-3 py-2 text-[10.5px] font-semibold text-white transition hover:bg-brand-700"
                @click="openAccept(request)"
              >
                <Check :size="12" /> Accept
              </button>
            </div>
            <button
              v-else-if="request.request_status === 'accepted'"
              type="button"
              class="rounded-lg border border-brand-200 bg-brand-50/60 px-3 py-2 text-[10.5px] font-semibold text-brand-700 transition hover:bg-brand-50"
              @click="openDelivery(request.id)"
            >
              View delivery
            </button>
          </div>
        </article>
      </div>
    </section>

    <div
      v-if="acceptOpen"
      class="fixed inset-0 z-[80] flex items-center justify-center bg-ink-950/55 px-3 backdrop-blur-md"
      role="dialog"
      aria-modal="true"
      @click.self="acceptOpen = false"
    >
      <form class="dash-modal w-full max-w-md overflow-hidden rounded-2xl border shadow-2xl" @submit.prevent="acceptRequest">
        <div class="flex items-start justify-between gap-3 border-b border-ink-900/8 bg-white p-5 dark:bg-white/5">
          <div class="flex items-start gap-3">
            <span class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-violet-100 text-violet-700 dark:bg-violet-400/15 dark:text-violet-200">
              <CheckCircle2 :size="18" />
            </span>
            <div>
              <p class="text-[9.5px] font-semibold tracking-[0.16em] text-brand-700 uppercase dark:text-indigo-200">ACCEPT REQUEST</p>
              <h2 class="mt-1 text-[15px] font-semibold">Assign a driver</h2>
              <p class="mt-1 text-[11px] text-ink-900/50">A delivery is created as soon as you assign it.</p>
            </div>
          </div>
          <button type="button" class="rounded-lg p-1.5 text-ink-900/45 hover:bg-ink-900/5" aria-label="Close" @click="acceptOpen = false">
            <X :size="18" />
          </button>
        </div>
        <div class="p-5">
          <label class="block">
            <span class="mb-1.5 block text-[11px] font-medium">Driver</span>
            <select v-model="selectedDriver" required class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-[12px]">
              <option value="">Select active driver</option>
              <option v-for="driver in drivers.filter((d) => d.status === 'active')" :key="driver.id" :value="driver.id">
                {{ driver.full_name }} · {{ driver.vehicle_plate }}
              </option>
            </select>
          </label>
          <p v-if="!drivers.filter((d) => d.status === 'active').length" class="mt-2 text-[10.5px] text-amber-700">
            Create an active driver before accepting the request.
          </p>
        </div>
        <div class="flex justify-end gap-2 border-t border-ink-900/8 bg-white px-5 py-4 dark:bg-white/5">
          <button type="button" class="rounded-lg border px-3 py-2 text-[11px] font-semibold" @click="acceptOpen = false">Cancel</button>
          <button
            type="submit"
            :disabled="saving || !selectedDriver"
            class="rounded-lg bg-brand-600 px-4 py-2 text-[11px] font-semibold text-white transition hover:bg-brand-700 disabled:opacity-50"
          >
            {{ saving ? 'Accepting…' : 'Accept & assign' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
