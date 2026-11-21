<script setup lang="ts">
import {
  ChevronDown,
  CircleAlert,
  ExternalLink,
  Loader2,
  Mail,
  Package,
  Phone,
  Search,
  Truck,
  Users,
  Wallet,
} from 'lucide-vue-next'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Customers' })

interface ProviderCustomer {
  customer_id: string
  full_name: string
  email: string
  phone: string | null
  joined_at: string
  total_bookings: number
  delivered_bookings: number
  active_bookings: number
  first_booking_at: string
  last_booking_at: string
}

interface CustomerShipment {
  id: string
  status: string
  request_status: 'pending' | 'accepted' | 'declined'
  pickup_address: string
  dropoff_address: string
  item_description: string
  package_category: string
  weight_kg: number
  delivery_price_minor: number
  tracking_code: string | null
  tracking_url: string | null
  created_at: string
}

const { api } = useApi()
const { detail, load } = useWorkspace()
const customers = ref<ProviderCustomer[]>([])
const loading = ref(true)
const error = ref('')
const searchQuery = ref('')
const expandedId = ref('')
const shipments = ref<Record<string, CustomerShipment[]>>({})
const shipmentsLoading = ref('')

const currency = computed(() => detail.value?.settings.currency || 'NGN')

onMounted(async () => {
  await load()
  await fetchCustomers()
})

async function fetchCustomers() {
  loading.value = true
  error.value = ''
  try {
    customers.value = (
      await api<{ customers: ProviderCustomer[] }>('/companies/me/customers')
    ).customers
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not load your customers.')
  } finally {
    loading.value = false
  }
}

const filtered = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return customers.value
  return customers.value.filter((customer) =>
    [customer.full_name, customer.email, customer.phone].some((value) =>
      (value || '').toLowerCase().includes(query),
    ),
  )
})

const stats = computed(() => [
  {
    label: 'Customers',
    value: String(customers.value.length),
    detail: 'Registered accounts that booked with you',
    icon: Users,
  },
  {
    label: 'Bookings',
    value: String(customers.value.reduce((total, customer) => total + Number(customer.total_bookings), 0)),
    detail: 'Requests and deliveries from these customers',
    icon: Package,
  },
  {
    label: 'Active',
    value: String(customers.value.reduce((total, customer) => total + Number(customer.active_bookings), 0)),
    detail: 'In progress with this provider',
    icon: Truck,
  },
  {
    label: 'Delivered',
    value: String(customers.value.reduce((total, customer) => total + Number(customer.delivered_bookings), 0)),
    detail: 'Completed for these customers',
    icon: Wallet,
  },
])

function formatMoney(minor: number) {
  return new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency: currency.value,
    maximumFractionDigits: 0,
  }).format(minor / 100)
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString('en-NG', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function shipmentLabel(shipment: CustomerShipment) {
  if (shipment.request_status === 'pending') return 'Awaiting review'
  if (shipment.request_status === 'declined') return 'Declined'
  return shipment.status.replaceAll('_', ' ')
}

function shipmentTone(shipment: CustomerShipment) {
  if (shipment.request_status === 'pending') return 'bg-amber-50 text-amber-700'
  if (shipment.request_status === 'declined' || shipment.status === 'cancelled') return 'bg-rose-50 text-rose-700'
  if (shipment.status === 'delivered') return 'bg-emerald-50 text-emerald-700'
  return 'bg-brand-50 text-brand-700'
}

async function toggleCustomer(customer: ProviderCustomer) {
  if (expandedId.value === customer.customer_id) {
    expandedId.value = ''
    return
  }
  expandedId.value = customer.customer_id
  if (shipments.value[customer.customer_id]) return

  shipmentsLoading.value = customer.customer_id
  try {
    const result = await api<{ deliveries: CustomerShipment[] }>(
      `/companies/me/customers/${encodeURIComponent(customer.customer_id)}/bookings`,
    )
    shipments.value = { ...shipments.value, [customer.customer_id]: result.deliveries }
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not load this customer’s shipment history.')
  } finally {
    shipmentsLoading.value = ''
  }
}
</script>

<template>
  <div class="dash-page">
    <header class="dash-hero relative overflow-hidden rounded-2xl border p-5 sm:p-6">
      <div class="pointer-events-none absolute -top-16 right-10 size-44 rounded-full bg-rose-400/15 blur-3xl" />
      <div class="pointer-events-none absolute -right-8 -bottom-24 size-52 rounded-full bg-cyan-400/15 blur-3xl" />
      <div class="relative flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="dash-eyebrow inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-semibold tracking-wide">
            <Users :size="12" /> CUSTOMER DIRECTORY
          </p>
          <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Customers</h1>
          <p class="mt-1.5 max-w-xl text-[12px] leading-relaxed text-ink-900/65">
            People who created an account when booking with you. Open a customer to see their contact
            details and shipment history with your company.
          </p>
        </div>
        <span class="inline-flex items-center gap-1.5 rounded-full border border-white/70 bg-white/70 px-3 py-1.5 text-[11px] font-semibold text-ink-900/60 dark:border-white/10 dark:bg-white/5">
          <Package :size="13" /> {{ customers.length }} registered
        </span>
      </div>
    </header>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ error }}
    </p>

    <div class="mt-4 grid grid-cols-2 gap-2 sm:gap-3 lg:grid-cols-4">
      <div v-for="stat in stats" :key="stat.label" class="dash-stat dash-stat--violet rounded-xl border p-3.5 sm:p-4">
        <span class="flex items-center justify-between gap-2">
          <span class="text-[10px] font-semibold tracking-wide text-ink-900/50 uppercase">{{ stat.label }}</span>
          <span class="flex size-6 items-center justify-center rounded-lg bg-white/80 text-ink-900/45 dark:bg-white/10">
            <component :is="stat.icon" :size="12" />
          </span>
        </span>
        <span class="mt-2 block text-[22px] leading-none font-semibold tracking-tight">{{ stat.value }}</span>
        <span class="mt-1.5 block text-[10px] leading-relaxed text-ink-900/40">{{ stat.detail }}</span>
      </div>
    </div>

    <section class="dash-panel mt-4 overflow-hidden rounded-2xl border">
      <div class="flex flex-col gap-3 border-b border-ink-900/8 p-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-[13px] font-semibold">Customer accounts</h2>
          <p class="mt-0.5 text-[10.5px] text-ink-900/40">{{ filtered.length }} shown</p>
        </div>
        <label class="relative block w-full sm:max-w-xs">
          <Search :size="15" class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-ink-900/35" />
          <input
            v-model="searchQuery"
            type="search"
            placeholder="Search name, email or phone"
            class="w-full rounded-lg border border-ink-900/10 py-2.5 pr-3 pl-9 text-[11.5px] outline-none focus:border-brand-500"
          />
        </label>
      </div>

      <div v-if="loading" class="flex items-center justify-center gap-2 p-12 text-[12px] text-ink-900/45">
        <Loader2 :size="15" class="animate-spin" /> Loading customers…
      </div>
      <div v-else-if="!filtered.length" class="p-10 text-center">
        <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-gradient-to-br from-rose-100 to-cyan-50 text-brand-700 dark:from-rose-400/15 dark:to-cyan-400/10 dark:text-indigo-200">
          <Users :size="19" />
        </span>
        <h3 class="mt-3 text-[13px] font-semibold">
          {{ searchQuery ? 'No matching customers' : 'No registered customers yet' }}
        </h3>
        <p class="mt-1 text-[11.5px] text-ink-900/45">
          {{ searchQuery
            ? 'Try a different name, email or phone number.'
            : 'Customers who choose to save a booking to an account will appear here with their details.' }}
        </p>
      </div>

      <div v-else class="divide-y divide-ink-900/8">
        <article v-for="customer in filtered" :key="customer.customer_id" class="transition hover:bg-brand-50/20">
          <button
            type="button"
            class="flex w-full flex-wrap items-start justify-between gap-3 p-4 text-left sm:p-5"
            :aria-expanded="expandedId === customer.customer_id"
            @click="toggleCustomer(customer)"
          >
            <div class="flex min-w-0 items-start gap-3">
              <span class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-brand-500 to-violet-500 text-[12px] font-semibold text-white shadow-sm shadow-brand-600/25">
                {{ customer.full_name.slice(0, 1).toUpperCase() }}
              </span>
              <div class="min-w-0">
                <p class="truncate text-[12.5px] font-semibold">{{ customer.full_name }}</p>
                <div class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-[10.5px] text-ink-900/50">
                  <span class="inline-flex items-center gap-1"><Mail :size="11" /> {{ customer.email }}</span>
                  <span v-if="customer.phone" class="inline-flex items-center gap-1"><Phone :size="11" /> {{ customer.phone }}</span>
                </div>
                <p class="mt-1 text-[10px] text-ink-900/40">
                  {{ customer.total_bookings }} booking{{ Number(customer.total_bookings) === 1 ? '' : 's' }}
                  · {{ customer.delivered_bookings }} delivered
                  · last booking {{ formatDate(customer.last_booking_at) }}
                </p>
              </div>
            </div>
            <span class="flex shrink-0 items-center gap-2">
              <span class="rounded-full bg-ink-900/5 px-2.5 py-1 text-[9.5px] font-semibold text-ink-900/55">
                Since {{ formatDate(customer.joined_at) }}
              </span>
              <ChevronDown
                :size="15"
                class="text-ink-900/35 transition"
                :class="expandedId === customer.customer_id ? 'rotate-180' : ''"
              />
            </span>
          </button>

          <div v-if="expandedId === customer.customer_id" class="border-t border-ink-900/8 bg-canvas/60 px-4 py-4 sm:px-5">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <h3 class="text-[11.5px] font-semibold">Shipment history with your company</h3>
              <span class="text-[10px] text-ink-900/40">{{ customer.first_booking_at ? `First booking ${formatDate(customer.first_booking_at)}` : '' }}</span>
            </div>

            <div v-if="shipmentsLoading === customer.customer_id" class="mt-3 flex items-center gap-2 text-[11px] text-ink-900/45">
              <Loader2 :size="13" class="animate-spin" /> Loading shipments…
            </div>
            <div v-else-if="shipments[customer.customer_id]?.length" class="mt-3 space-y-2">
              <div
                v-for="shipment in shipments[customer.customer_id]"
                :key="shipment.id"
                class="rounded-xl border border-ink-900/8 bg-white p-3"
              >
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <p class="text-[11.5px] font-medium">
                    {{ shipment.pickup_address }} <span class="px-1 text-brand-500">→</span> {{ shipment.dropoff_address }}
                  </p>
                  <span class="rounded-full px-2 py-0.5 text-[9px] font-semibold capitalize" :class="shipmentTone(shipment)">
                    {{ shipmentLabel(shipment) }}
                  </span>
                </div>
                <div class="mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-[10px] text-ink-900/45">
                  <span>{{ formatDate(shipment.created_at) }}</span>
                  <span class="capitalize">{{ shipment.package_category }}</span>
                  <span>{{ shipment.weight_kg }} kg</span>
                  <span class="font-semibold text-ink-900/60">{{ formatMoney(shipment.delivery_price_minor) }}</span>
                  <a
                    v-if="shipment.tracking_url"
                    :href="shipment.tracking_url"
                    target="_blank"
                    rel="noopener"
                    class="ml-auto inline-flex items-center gap-1 font-semibold text-brand-700 hover:text-brand-800"
                  >
                    Track <ExternalLink :size="11" />
                  </a>
                </div>
              </div>
            </div>
            <p v-else class="mt-3 rounded-lg bg-white px-3 py-6 text-center text-[11px] text-ink-900/40">
              No shipments recorded for this customer yet.
            </p>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>
