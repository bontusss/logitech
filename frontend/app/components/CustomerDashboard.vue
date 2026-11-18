<script setup lang="ts">
import {
  ArrowRight,
  Check,
  Clock3,
  Info,
  MapPin,
  Package,
  PackageCheck,
  ReceiptText,
  Send,
  Truck,
  Wallet,
} from 'lucide-vue-next'
import type { BookingStatus } from '~/types/customer'

const { user } = useAuth()
const {
  data,
  activeShipments,
  inTransitCount,
  pendingBookings,
  maxMonthlySpend,
  formatNaira,
  shipmentStatusMeta,
  isSample,
} = useCustomerDashboard()

const selectedId = ref(activeShipments.value[0]?.id ?? '')
const selectedShipment = computed(
  () => activeShipments.value.find((shipment) => shipment.id === selectedId.value) ?? activeShipments.value[0],
)

const recentRoutes = computed(() => {
  const seen = new Set<string>()
  return activeShipments.value
    .filter((shipment) => {
      const route = `${shipment.origin}|${shipment.destination}`
      if (seen.has(route)) return false
      seen.add(route)
      return true
    })
    .slice(0, 2)
    .map((shipment) => ({
      label: `${shipment.origin} → ${shipment.destination}`,
      to: '/#rate-search',
    }))
})

const bookingMeta: Record<BookingStatus, { label: string; tone: string; action: string }> = {
  awaiting_payment: {
    label: 'Payment pending',
    tone: 'border-amber-200 bg-amber-50 text-amber-700',
    action: 'Pay & confirm',
  },
  awaiting_provider: {
    label: 'Awaiting provider',
    tone: 'border-brand-200 bg-brand-50 text-brand-700',
    action: 'View booking',
  },
  awaiting_pickup: {
    label: 'Pickup scheduled',
    tone: 'border-emerald-200 bg-emerald-50 text-emerald-700',
    action: 'View booking',
  },
}

const spendStats = computed(() => [
  { label: 'This month', value: formatNaira(data.spend.thisMonthKobo) },
  { label: 'Last month', value: formatNaira(data.spend.lastMonthKobo) },
  { label: 'Lifetime', value: formatNaira(data.spend.lifetimeKobo) },
])

const firstName = computed(() => user.value?.full_name?.split(' ')[0] || 'there')
</script>

<template>
  <div>
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">
          Marketplace dashboard
        </p>
        <h1 class="mt-2 text-[clamp(1.6rem,3vw,2.1rem)] leading-tight font-semibold tracking-[-0.035em]">
          Welcome back, {{ firstName }}
        </h1>
        <p class="mt-2 max-w-xl text-[13.5px] leading-relaxed text-ink-900/50">
          Your shipments, bookings and shipping spend in one place.
        </p>
      </div>
      <AppButton to="/#rate-search" size="md">
        Send a package
        <ArrowRight :size="15" />
      </AppButton>
    </div>

    <div
      v-if="isSample"
      class="mt-5 flex items-start gap-2 rounded-xl border border-amber-200 bg-amber-50 px-3.5 py-2.5 text-[11px] leading-relaxed text-amber-800"
    >
      <Info :size="14" class="mt-0.5 shrink-0" />
      <p>
        <span class="font-semibold">Sample data.</span>
        Live tracking, bookings and spend appear here once marketplace booking is connected.
      </p>
    </div>

    <div class="mt-5 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <div class="rounded-xl border border-ink-900/10 bg-white p-4">
        <div class="flex items-center justify-between gap-2">
          <p class="text-[9.5px] font-semibold tracking-[0.16em] text-ink-900/40 uppercase">Active shipments</p>
          <Package :size="15" class="text-brand-600" />
        </div>
        <p class="mt-2 text-[22px] font-semibold tracking-[-0.03em]">{{ activeShipments.length }}</p>
        <p class="mt-1 text-[10.5px] text-ink-900/40">Moving right now</p>
      </div>

      <div class="rounded-xl border border-ink-900/10 bg-white p-4">
        <div class="flex items-center justify-between gap-2">
          <p class="text-[9.5px] font-semibold tracking-[0.16em] text-ink-900/40 uppercase">In transit</p>
          <Truck :size="15" class="text-brand-600" />
        </div>
        <p class="mt-2 text-[22px] font-semibold tracking-[-0.03em]">{{ inTransitCount }}</p>
        <p class="mt-1 text-[10.5px] text-ink-900/40">On the road or out for delivery</p>
      </div>

      <div class="rounded-xl border border-ink-900/10 bg-white p-4">
        <div class="flex items-center justify-between gap-2">
          <p class="text-[9.5px] font-semibold tracking-[0.16em] text-ink-900/40 uppercase">Pending bookings</p>
          <ReceiptText :size="15" class="text-brand-600" />
        </div>
        <p class="mt-2 text-[22px] font-semibold tracking-[-0.03em]">{{ pendingBookings.length }}</p>
        <p class="mt-1 text-[10.5px] text-ink-900/40">Waiting on you or the provider</p>
      </div>

      <div class="rounded-xl border border-ink-900/10 bg-white p-4">
        <div class="flex items-center justify-between gap-2">
          <p class="text-[9.5px] font-semibold tracking-[0.16em] text-ink-900/40 uppercase">Spent this month</p>
          <Wallet :size="15" class="text-brand-600" />
        </div>
        <p class="mt-2 text-[22px] font-semibold tracking-[-0.03em]">{{ formatNaira(data.spend.thisMonthKobo) }}</p>
        <p class="mt-1 text-[10.5px] text-ink-900/40">Delivered and in-progress orders</p>
      </div>
    </div>

    <div class="mt-5 grid gap-5 lg:grid-cols-[minmax(0,1.45fr)_minmax(0,1fr)]">
      <div class="min-w-0">
        <section class="rounded-xl border border-ink-900/10 bg-white p-5">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div>
              <h2 class="text-[14px] font-semibold">Your shipments</h2>
              <p class="mt-1 text-[11px] text-ink-900/40">Select a shipment to follow its latest scan events.</p>
            </div>
            <span class="rounded-md bg-ink-900/5 px-2 py-1 text-[10px] font-medium text-ink-900/50">
              {{ activeShipments.length }} active
            </span>
          </div>

          <ul class="mt-4 space-y-2">
            <li v-for="shipment in activeShipments" :key="shipment.id">
              <button
                type="button"
                class="w-full rounded-xl border p-3.5 text-left transition"
                :class="
                  shipment.id === selectedShipment?.id
                    ? 'border-brand-500/40 bg-brand-50/35'
                    : 'border-ink-900/10 hover:border-ink-900/20 hover:bg-canvas/60'
                "
                :aria-pressed="shipment.id === selectedShipment?.id"
                @click="selectedId = shipment.id"
              >
                <div class="flex flex-wrap items-center gap-2">
                  <span class="text-[12.5px] font-semibold">{{ shipment.reference }}</span>
                  <span class="text-[11.5px] text-ink-900/45">
                    {{ shipment.origin }} <span class="px-0.5 text-brand-500">→</span> {{ shipment.destination }}
                  </span>
                  <span
                    class="ml-auto inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[9.5px] font-semibold"
                    :class="shipmentStatusMeta[shipment.status].tone"
                  >
                    <span class="size-1.5 rounded-full" :class="shipmentStatusMeta[shipment.status].dot" />
                    {{ shipmentStatusMeta[shipment.status].label }}
                  </span>
                </div>

                <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-[10.5px] text-ink-900/45">
                  <span class="inline-flex items-center gap-1">
                    <MapPin :size="11" class="text-brand-600" />
                    {{ shipment.currentLocation }}
                  </span>
                  <span class="inline-flex items-center gap-1">
                    <Clock3 :size="11" />
                    {{ shipment.eta }}
                  </span>
                </div>

                <div class="mt-2 flex flex-wrap items-center gap-2 text-[10px] text-ink-900/35">
                  <span>{{ shipment.provider }} · {{ shipment.service }}</span>
                  <span class="text-ink-900/15">|</span>
                  <span>{{ shipment.weightKg }} kg</span>
                  <span class="ml-auto font-semibold text-ink-900/60">{{ formatNaira(shipment.priceKobo) }}</span>
                </div>
              </button>
            </li>
          </ul>

          <p v-if="activeShipments.length === 0" class="mt-4 rounded-lg bg-canvas px-3 py-6 text-center text-[12px] text-ink-900/40">
            No active shipments. Book a delivery to see it here.
          </p>
        </section>

        <section v-if="selectedShipment" class="mt-5 rounded-xl border border-ink-900/10 bg-white p-5">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h2 class="flex items-center gap-2 text-[14px] font-semibold">
                <MapPin :size="15" class="text-brand-600" />
                Where your package is
              </h2>
              <p class="mt-1 text-[11px] text-ink-900/40">
                {{ selectedShipment.reference }} · {{ selectedShipment.provider }}
              </p>
            </div>
            <div class="rounded-lg bg-canvas px-3 py-2 text-right">
              <p class="text-[9px] font-semibold tracking-[0.14em] text-ink-900/35 uppercase">Current location</p>
              <p class="mt-0.5 text-[12px] font-semibold">{{ selectedShipment.currentLocation }}</p>
            </div>
          </div>

          <ol class="mt-5">
            <li
              v-for="(event, index) in selectedShipment.events"
              :key="`${event.label}-${event.at}`"
              class="flex gap-3"
            >
              <div class="flex w-4 shrink-0 flex-col items-center">
                <span
                  class="mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-full border"
                  :class="{
                    'border-brand-500 bg-brand-500 text-white': event.state === 'done',
                    'border-brand-500 bg-white text-brand-600': event.state === 'current',
                    'border-ink-900/15 bg-white': event.state === 'pending',
                  }"
                >
                  <Check v-if="event.state === 'done'" :size="9" :stroke-width="3.5" />
                  <span v-else-if="event.state === 'current'" class="size-1.5 rounded-full bg-brand-500" />
                </span>
                <span
                  v-if="index < selectedShipment.events.length - 1"
                  class="mt-1 w-px flex-1"
                  :class="event.state === 'done' ? 'bg-brand-200' : 'bg-ink-900/10'"
                />
              </div>
              <div class="min-w-0 flex-1 pb-5 last:pb-0">
                <div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-0.5">
                  <p
                    class="text-[12px] font-medium"
                    :class="event.state === 'pending' ? 'text-ink-900/40' : ''"
                  >
                    {{ event.label }}
                  </p>
                  <p class="text-[10px] text-ink-900/35">{{ event.at }}</p>
                </div>
                <p class="mt-0.5 text-[10.5px] text-ink-900/40">{{ event.location }}</p>
              </div>
            </li>
          </ol>
        </section>
      </div>

      <div class="min-w-0 space-y-5">
        <section class="rounded-xl border border-ink-900/10 bg-white p-5">
          <div class="flex items-center justify-between gap-2">
            <h2 class="text-[14px] font-semibold">Pending bookings</h2>
            <span class="rounded-md bg-ink-900/5 px-2 py-1 text-[10px] font-medium text-ink-900/50">
              {{ pendingBookings.length }}
            </span>
          </div>

          <ul v-if="pendingBookings.length" class="mt-4 space-y-3">
            <li
              v-for="booking in pendingBookings"
              :key="booking.id"
              class="rounded-xl border border-ink-900/10 p-3.5"
            >
              <div class="flex flex-wrap items-center gap-2">
                <span class="text-[12px] font-semibold">{{ booking.reference }}</span>
                <span
                  class="ml-auto rounded-full border px-2 py-0.5 text-[9.5px] font-semibold"
                  :class="bookingMeta[booking.status].tone"
                >
                  {{ bookingMeta[booking.status].label }}
                </span>
              </div>
              <p class="mt-1.5 text-[11.5px] text-ink-900/55">
                {{ booking.origin }} <span class="px-0.5 text-brand-500">→</span> {{ booking.destination }}
                <span class="text-ink-900/25">·</span>
                {{ booking.provider }}
              </p>
              <p class="mt-1 text-[10.5px] leading-relaxed text-ink-900/40">{{ booking.note }}</p>
              <div class="mt-3 flex items-center justify-between gap-2 border-t border-ink-900/8 pt-2.5">
                <div>
                  <p class="text-[12.5px] font-semibold">{{ formatNaira(booking.priceKobo) }}</p>
                  <p class="text-[9.5px] text-ink-900/35">{{ booking.createdAt }}</p>
                </div>
                <NuxtLink
                  to="/#rate-search"
                  class="inline-flex items-center gap-1 text-[10.5px] font-semibold text-brand-600 hover:text-brand-700"
                >
                  {{ bookingMeta[booking.status].action }}
                  <ArrowRight :size="11" />
                </NuxtLink>
              </div>
            </li>
          </ul>

          <p v-else class="mt-4 rounded-lg bg-canvas px-3 py-5 text-center text-[11.5px] text-ink-900/40">
            No pending bookings.
          </p>
        </section>

        <section class="rounded-xl border border-brand-100 bg-brand-50/60 p-5">
          <div class="flex items-center gap-2">
            <Send :size="16" class="text-brand-600" />
            <h2 class="text-[14px] font-semibold">Send another package</h2>
          </div>
          <p class="mt-1.5 text-[11.5px] leading-relaxed text-ink-900/50">
            Compare verified providers on price and delivery time, then book in a few steps.
          </p>
          <AppButton to="/#rate-search" size="sm" block class="mt-4">
            Compare shipping rates
            <ArrowRight :size="13" />
          </AppButton>
          <div v-if="recentRoutes.length" class="mt-3">
            <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/35 uppercase">Recent routes</p>
            <div class="mt-1.5 flex flex-wrap gap-1.5">
              <NuxtLink
                v-for="route in recentRoutes"
                :key="route.label"
                :to="route.to"
                class="rounded-md border border-ink-900/10 bg-white px-2 py-1 text-[10.5px] font-medium text-ink-900/60 transition hover:border-brand-200 hover:text-brand-700"
              >
                {{ route.label }}
              </NuxtLink>
            </div>
          </div>
        </section>
      </div>
    </div>

    <section class="mt-5 rounded-xl border border-ink-900/10 bg-white p-5">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 class="flex items-center gap-2 text-[14px] font-semibold">
            <PackageCheck :size="15" class="text-brand-600" />
            Shipping spend
          </h2>
          <p class="mt-1 text-[11px] text-ink-900/40">
            {{ data.spend.deliveredCount }} deliveries completed with BOL.
          </p>
        </div>
        <div class="grid grid-cols-3 gap-4 text-right">
          <div v-for="stat in spendStats" :key="stat.label">
            <p class="text-[9px] font-semibold tracking-[0.12em] text-ink-900/35 uppercase">{{ stat.label }}</p>
            <p class="mt-0.5 text-[13px] font-semibold tracking-[-0.02em]">{{ stat.value }}</p>
          </div>
        </div>
      </div>

      <div class="mt-5 flex h-28 items-end gap-2 sm:gap-4">
        <div
          v-for="(point, index) in data.spend.monthly"
          :key="point.month"
          class="flex h-full flex-1 flex-col items-center justify-end gap-1.5"
          :title="`${point.month}: ${formatNaira(point.amountKobo)}`"
        >
          <div class="flex w-full flex-1 items-end">
            <div
              class="w-full rounded-t-md transition-all"
              :class="index === data.spend.monthly.length - 1 ? 'bg-brand-500' : 'bg-brand-200'"
              :style="{ height: `${Math.max(6, (point.amountKobo / maxMonthlySpend) * 100)}%` }"
            />
          </div>
          <span class="text-[9.5px] font-medium text-ink-900/40">{{ point.month }}</span>
        </div>
      </div>
    </section>
  </div>
</template>
