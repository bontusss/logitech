<script setup lang="ts">
import {
  ArrowRight,
  Bell,
  Check,
  CircleAlert,
  Clock3,
  ExternalLink,
  Loader2,
  MapPin,
  Package,
  Send,
  Truck,
  UserRound,
  Users,
} from 'lucide-vue-next'
import type { CustomerBooking } from '~/types/customer'

const { user, fetchMe } = useAuth()
const { api } = useApi()
const {
  bookings,
  loading,
  error,
  load,
  activeBookings,
  inTransitCount,
  deliveredBookings,
  providers,
  addresses,
  stateFor,
  formatMoney,
  isTerminal,
} = useCustomerDashboard()

const activeTab = ref<'bookings' | 'addresses' | 'profile' | 'notifications'>('bookings')
const filter = ref<'all' | 'active' | 'completed'>('all')
const savingProfile = ref(false)
const profileError = ref('')
const profileSaved = ref(false)
const profile = reactive({
  full_name: user.value?.full_name ?? '',
  phone: user.value?.phone ?? '',
})

onMounted(() => load())

const firstName = computed(() => user.value?.full_name?.split(' ')[0] || 'there')

const stats = computed(() => [
  {
    label: 'Active parcels',
    value: String(activeBookings.value.length),
    detail: 'Bookings still in progress',
    icon: Package,
  },
  {
    label: 'In transit',
    value: String(inTransitCount.value),
    detail: 'On the road right now',
    icon: Truck,
  },
  {
    label: 'Delivered',
    value: String(deliveredBookings.value.length),
    detail: 'Completed deliveries',
    icon: Check,
  },
  {
    label: 'Providers used',
    value: String(providers.value.length),
    detail: 'Logistics companies booked',
    icon: Users,
  },
])

const tabs = [
  { key: 'bookings' as const, label: 'Bookings', icon: Package },
  { key: 'addresses' as const, label: 'Addresses', icon: MapPin },
  { key: 'profile' as const, label: 'Profile', icon: UserRound },
  { key: 'notifications' as const, label: 'Notifications', icon: Bell },
]

const filteredBookings = computed<CustomerBooking[]>(() => {
  if (filter.value === 'active') return activeBookings.value
  if (filter.value === 'completed') return bookings.value.filter((booking) => isTerminal(booking))
  return bookings.value
})

function formatDate(value: string) {
  return new Date(value).toLocaleDateString('en-NG', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleString('en-NG', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function speedLabel(speed: CustomerBooking['delivery_speed']) {
  return { standard: 'Standard', express: 'Express', same_day: 'Same-day' }[speed] || speed
}

async function saveProfile() {
  profileError.value = ''
  profileSaved.value = false
  savingProfile.value = true
  try {
    await api('/users/me', {
      method: 'PATCH',
      body: {
        full_name: profile.full_name.trim(),
        phone: profile.phone.trim(),
      },
    })
    await fetchMe()
    profileSaved.value = true
  } catch (cause) {
    profileError.value = apiErrorMessage(cause, 'Could not save your profile.')
  } finally {
    savingProfile.value = false
  }
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">
          Customer dashboard
        </p>
        <h1 class="mt-2 text-[clamp(1.6rem,3vw,2.1rem)] leading-tight font-semibold tracking-[-0.035em]">
          Welcome back, {{ firstName }}
        </h1>
        <p class="mt-2 max-w-xl text-[13.5px] leading-relaxed text-ink-900/50">
          Track parcels, revisit bookings from every provider, and keep your contact details current.
        </p>
      </div>
      <AppButton to="/#rate-search" size="md">
        Send a package
        <ArrowRight :size="15" />
      </AppButton>
    </div>

    <div class="mt-5 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <div v-for="stat in stats" :key="stat.label" class="rounded-xl border border-ink-900/10 bg-surface p-4">
        <div class="flex items-center justify-between gap-2">
          <p class="text-[9.5px] font-semibold tracking-[0.16em] text-ink-900/40 uppercase">{{ stat.label }}</p>
          <component :is="stat.icon" :size="15" class="text-brand-600" />
        </div>
        <p class="mt-2 text-[22px] font-semibold tracking-[-0.03em]">{{ stat.value }}</p>
        <p class="mt-1 text-[10.5px] text-ink-900/40">{{ stat.detail }}</p>
      </div>
    </div>

    <nav class="mt-6 flex flex-wrap gap-1 border-b border-ink-900/10" aria-label="Customer dashboard sections">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        class="-mb-px inline-flex items-center gap-1.5 border-b-2 px-3 py-2.5 text-[12px] font-medium transition"
        :class="activeTab === tab.key
          ? 'border-accent-500 text-ink-900'
          : 'border-transparent text-ink-900/50 hover:text-ink-900'"
        @click="activeTab = tab.key"
      >
        <component :is="tab.icon" :size="14" />
        {{ tab.label }}
        <span
          v-if="tab.key === 'notifications'"
          class="rounded-full border border-accent-200 bg-accent-50 px-1.5 py-0.5 text-[8.5px] font-semibold tracking-wide text-accent-700 uppercase"
        >
          Soon
        </span>
      </button>
    </nav>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ error }}
    </p>

    <section v-if="activeTab === 'bookings'" class="mt-4">
      <div class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h2 class="text-[14px] font-semibold">Your bookings</h2>
          <p class="mt-1 text-[11px] text-ink-900/40">
            {{ bookings.length }} booking{{ bookings.length === 1 ? '' : 's' }} across
            {{ providers.length }} provider{{ providers.length === 1 ? '' : 's' }}
          </p>
        </div>
        <div class="flex gap-1 rounded-lg border border-ink-900/10 bg-surface p-1">
          <button
            v-for="option in ([
              { key: 'all', label: 'All' },
              { key: 'active', label: 'Active' },
              { key: 'completed', label: 'Completed' },
            ] as const)"
            :key="option.key"
            type="button"
            class="rounded-md px-2.5 py-1.5 text-[10.5px] font-semibold transition"
            :class="filter === option.key ? 'bg-accent-500 text-ink-950' : 'text-ink-900/50 hover:text-ink-900'"
            @click="filter = option.key"
          >
            {{ option.label }}
          </button>
        </div>
      </div>

      <div v-if="loading && !bookings.length" class="mt-4 flex items-center justify-center gap-2 rounded-xl border border-ink-900/10 bg-surface p-12 text-[12px] text-ink-900/45">
        <Loader2 :size="15" class="animate-spin" /> Loading your bookings…
      </div>

      <div v-else-if="!filteredBookings.length" class="mt-4 rounded-xl border border-ink-900/10 bg-surface p-10 text-center">
        <IconTile tone="brand" size="lg" rounded="2xl" class="mx-auto">
          <Truck :size="19" />
        </IconTile>
        <h3 class="mt-3 text-[13px] font-semibold">
          {{ filter === 'all' ? 'No bookings yet' : `No ${filter} bookings` }}
        </h3>
        <p class="mt-1 text-[11.5px] text-ink-900/45">
          Bookings made with a provider while signed in, or saved from a booking page, appear here.
        </p>
        <AppButton to="/#rate-search" size="sm" class="mt-4">Compare providers</AppButton>
      </div>

      <ul v-else class="mt-4 space-y-2.5">
        <li
          v-for="booking in filteredBookings"
          :key="booking.id"
          class="rounded-xl border border-ink-900/10 bg-surface p-4"
        >
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <span class="text-[12.5px] font-semibold">{{ booking.provider_name }}</span>
                <span
                  class="inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[9.5px] font-semibold"
                  :class="stateFor(booking).tone"
                >
                  <span class="size-1.5 rounded-full" :class="stateFor(booking).dot" />
                  {{ stateFor(booking).label }}
                </span>
              </div>
              <p class="mt-1.5 text-[11.5px] text-ink-900/60">
                {{ booking.pickup_address }}
                <span class="px-1 text-brand-500">→</span>
                {{ booking.dropoff_address }}
              </p>
              <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-[10.5px] text-ink-900/45">
                <span class="inline-flex items-center gap-1">
                  <Clock3 :size="11" /> {{ formatDateTime(booking.created_at) }}
                </span>
                <span>{{ speedLabel(booking.delivery_speed) }}</span>
                <span>{{ booking.weight_kg }} kg</span>
                <span v-if="booking.tracking_code" class="font-mono text-[10px]">{{ booking.tracking_code }}</span>
              </div>
            </div>
            <div class="flex shrink-0 flex-col items-end gap-2">
              <p class="text-[13px] font-semibold">{{ formatMoney(booking) }}</p>
              <a
                v-if="booking.tracking_url"
                :href="booking.tracking_url"
                target="_blank"
                rel="noopener"
                class="inline-flex items-center gap-1.5 rounded-lg border border-accent-500/30 bg-accent-50/70 px-2.5 py-1.5 text-[10.5px] font-semibold text-accent-700 transition hover:bg-accent-50 dark:text-accent-300"
              >
                Track parcel <ExternalLink :size="12" />
              </a>
            </div>
          </div>
          <p class="mt-2 border-t border-ink-900/8 pt-2 text-[10px] text-ink-900/45">
            <template v-if="booking.request_status === 'pending'">
              {{ booking.provider_name }} is reviewing your request. Tracking will update once they accept.
            </template>
            <template v-else-if="booking.request_status === 'declined'">
              This request was declined. You can book with another provider.
            </template>
            <template v-else-if="booking.price_is_estimate">
              Final price is confirmed by {{ booking.provider_name }} on dispatch.
            </template>
            <template v-else>
              Delivery fee confirmed at booking.
            </template>
          </p>
        </li>
      </ul>
    </section>

    <section v-else-if="activeTab === 'addresses'" class="mt-4 rounded-xl border border-ink-900/10 bg-surface p-5">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 class="text-[14px] font-semibold">Addresses</h2>
          <p class="mt-1 text-[11px] text-ink-900/40">
            Pickup and drop-off places from your bookings. Reuse them when you send another package.
          </p>
        </div>
        <AppButton to="/#rate-search" size="sm" variant="secondary">Send a package</AppButton>
      </div>

      <ul v-if="addresses.length" class="mt-4 divide-y divide-ink-900/8">
        <li v-for="address in addresses" :key="`${address.role}-${address.address}`" class="flex items-start gap-3 py-3">
          <span
            class="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-lg"
            :class="address.role === 'pickup' ? 'bg-success-50 text-success-700' : 'bg-accent-50 text-accent-700 dark:text-accent-300'"
          >
            <MapPin :size="13" />
          </span>
          <div class="min-w-0 flex-1">
            <p class="text-[12px] font-medium">{{ address.address }}</p>
            <p class="mt-0.5 text-[10.5px] text-ink-900/45">
              {{ address.role === 'pickup' ? 'Pickup' : 'Drop-off' }} · {{ address.provider }} ·
              last used {{ formatDate(address.lastUsedAt) }}
            </p>
          </div>
        </li>
      </ul>
      <p v-else class="mt-4 rounded-lg bg-canvas px-3 py-8 text-center text-[12px] text-ink-900/40">
        Addresses you use in bookings will be listed here.
      </p>
    </section>

    <section v-else-if="activeTab === 'profile'" class="mt-4 rounded-xl border border-ink-900/10 bg-surface p-5">
      <h2 class="text-[14px] font-semibold">Personal information</h2>
      <p class="mt-1 text-[11px] text-ink-900/40">
        Providers use these details to coordinate your deliveries.
      </p>

      <p v-if="profileError" role="alert" class="mt-3 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[11px] text-red-700">
        <CircleAlert :size="13" class="mt-0.5 shrink-0" />{{ profileError }}
      </p>
      <p v-if="profileSaved" role="status" class="mt-3 flex items-center gap-2 rounded-lg border border-success-200 bg-success-50 px-3 py-2 text-[11px] text-success-700">
        <Check :size="13" /> Profile updated.
      </p>

      <form class="mt-4 grid max-w-lg gap-3 sm:grid-cols-2" @submit.prevent="saveProfile">
        <label class="block">
          <span class="mb-1.5 block text-[11px] font-medium">Full name</span>
          <input
            v-model="profile.full_name"
            required
            minlength="2"
            maxlength="120"
            autocomplete="name"
            class="w-full rounded-lg border border-ink-900/15 bg-canvas px-3 py-2.5 text-[12px] text-ink-900 outline-none focus:border-accent-500 focus:bg-surface"
          />
        </label>
        <label class="block">
          <span class="mb-1.5 block text-[11px] font-medium">Phone number</span>
          <input
            v-model="profile.phone"
            maxlength="30"
            type="tel"
            autocomplete="tel"
            placeholder="+234 800 000 0000"
            class="w-full rounded-lg border border-ink-900/15 bg-canvas px-3 py-2.5 text-[12px] text-ink-900 outline-none focus:border-accent-500 focus:bg-surface"
          />
        </label>
        <label class="block sm:col-span-2">
          <span class="mb-1.5 block text-[11px] font-medium">Email address</span>
          <input
            :value="user?.email"
            disabled
            class="w-full rounded-lg border border-ink-900/10 bg-canvas px-3 py-2.5 text-[12px] text-ink-900/50"
          />
          <span class="mt-1 block text-[10px] text-ink-900/35">
            {{ user?.email_verified ? 'Email verified' : 'Email verification pending' }}
          </span>
        </label>
        <div class="sm:col-span-2">
          <button
            type="submit"
            :disabled="savingProfile"
            class="inline-flex min-h-10 items-center gap-2 rounded-lg bg-accent-500 px-4 py-2 text-[11.5px] font-semibold text-ink-950 transition hover:bg-accent-400 disabled:opacity-50"
          >
            <Loader2 v-if="savingProfile" :size="14" class="animate-spin" />
            {{ savingProfile ? 'Saving…' : 'Save changes' }}
          </button>
        </div>
      </form>
    </section>

    <section v-else class="mt-4 rounded-xl border border-ink-900/10 bg-surface p-10 text-center">
      <IconTile tone="accent" size="lg" rounded="2xl" class="mx-auto">
        <Bell :size="19" />
      </IconTile>
      <h2 class="mt-3 text-[13.5px] font-semibold">Notifications are coming soon</h2>
      <p class="mx-auto mt-1 max-w-md text-[11.5px] leading-relaxed text-ink-900/45">
        You’ll be able to choose email and WhatsApp updates for pickup, transit and delivery events.
        For now, follow each parcel from its tracking page.
      </p>
      <AppButton to="/#rate-search" size="sm" variant="secondary" class="mt-4">
        <Send :size="13" /> Send a package
      </AppButton>
    </section>
  </div>
</template>
