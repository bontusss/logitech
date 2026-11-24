<script setup lang="ts">
import {
  Building2,
  CircleAlert,
  Loader2,
  RefreshCw,
  ShieldCheck,
  Store,
  UserPlus,
  Users,
} from 'lucide-vue-next'

definePageMeta({ middleware: 'admin', layout: 'admin' })
useSeoMeta({ title: 'Platform administration', robots: 'noindex' })

interface AdminCustomer {
  id: string
  email: string
  full_name: string
  phone: string | null
  status: string
  email_verified_at: string | null
  last_login_at: string | null
  created_at: string
}

interface AdminProvider {
  id: string
  name: string
  business_email: string
  phone: string
  business_type: string
  cac_business_name: string | null
  state: string | null
  city: string | null
  status: string
  onboarding_status: string
  verified_at: string | null
  created_at: string
  owner_name: string | null
  owner_email: string | null
}

interface AdminAccount {
  id: string
  email: string
  full_name: string
  platform_role: string | null
  status: string
  last_login_at: string | null
  created_at: string
}

interface AdminOverview {
  customers: AdminCustomer[]
  providers: AdminProvider[]
  admins: AdminAccount[]
}

const { api } = useApi()
const { user } = useAuth()
const { notify } = useWorkspace()

const overview = ref<AdminOverview | null>(null)
const loading = ref(true)
const loadError = ref('')
const approvingId = ref('')
const tabs = [
  { key: 'providers' as const, label: 'Providers', icon: Store },
  { key: 'customers' as const, label: 'Customers', icon: Users },
  { key: 'admins' as const, label: 'Admins', icon: ShieldCheck },
]
const tab = ref<'providers' | 'customers' | 'admins'>('providers')

const isSuperAdmin = computed(() => user.value?.platform_role === 'super_admin')
const canApprove = computed(() => user.value?.platform_role !== 'support')

const pendingProviders = computed(
  () =>
    overview.value?.providers.filter(
      (provider) => provider.status === 'pending' || provider.status === 'under_review',
    ) ?? [],
)

const adminForm = reactive({
  full_name: '',
  email: '',
  password: '',
  role: 'admin' as 'admin' | 'support',
})
const creatingAdmin = ref(false)
const createError = ref('')

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    overview.value = await api<AdminOverview>('/admin/overview')
  } catch (err) {
    loadError.value = apiErrorMessage(err, 'Could not load the administration overview.')
  } finally {
    loading.value = false
  }
}

async function approve(provider: AdminProvider) {
  if (!canApprove.value) return
  approvingId.value = provider.id
  try {
    await api(`/admin/providers/${provider.id}/approve`, { method: 'POST' })
    provider.status = 'verified'
    provider.onboarding_status = 'complete'
    provider.verified_at = new Date().toISOString()
    notify(`${provider.name} is now a verified provider.`)
  } catch (err) {
    notify(apiErrorMessage(err, 'Could not approve this provider.'), 'error')
  } finally {
    approvingId.value = ''
  }
}

async function createAdmin() {
  createError.value = ''
  creatingAdmin.value = true
  try {
    const result = await api<{ admin: AdminAccount }>('/admin/admins', {
      method: 'POST',
      body: {
        full_name: adminForm.full_name,
        email: adminForm.email,
        password: adminForm.password,
        role: adminForm.role,
      },
    })
    overview.value?.admins.unshift(result.admin)
    notify(`${result.admin.full_name} can now sign in to the admin dashboard.`)
    adminForm.full_name = ''
    adminForm.email = ''
    adminForm.password = ''
    adminForm.role = 'admin'
  } catch (err) {
    createError.value = apiErrorMessage(err, 'Could not create this admin account.')
  } finally {
    creatingAdmin.value = false
  }
}

function formatDate(value: string | null) {
  if (!value) return '—'
  return new Date(value).toLocaleDateString('en-NG', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

function statusClass(status: string) {
  if (status === 'verified' || status === 'active') return 'bg-success-50 text-success-700 border-success-200'
  if (status === 'under_review') return 'bg-brand-50 text-brand-700 border-brand-200'
  if (status === 'pending' || status === 'pending_review') return 'bg-accent-50 text-accent-700 border-accent-200'
  if (status === 'rejected' || status === 'suspended') return 'bg-red-50 text-red-700 border-red-200'
  return 'bg-slate-50 text-slate-600 border-slate-200'
}

onMounted(load)
</script>

<template>
  <div class="dash-page">
    <section class="dash-hero relative isolate overflow-hidden rounded-3xl border p-5 sm:p-6">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="dash-eyebrow inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-semibold tracking-[0.18em] uppercase">
            <ShieldCheck :size="13" /> Platform administration
          </p>
          <h1 class="mt-3 text-[clamp(1.5rem,2.6vw,1.9rem)] leading-tight font-semibold tracking-[-0.035em]">
            Marketplace overview
          </h1>
          <p class="mt-1.5 max-w-[520px] text-[13px] leading-relaxed text-ink-900/55">
            Customers, logistics providers and staff accounts across BOL.
          </p>
        </div>
        <button
          type="button"
          class="inline-flex items-center gap-1.5 rounded-xl border border-ink-900/10 bg-white/75 px-3 py-2 text-[12px] font-semibold shadow-sm backdrop-blur-sm transition hover:-translate-y-0.5 hover:border-brand-200 hover:text-brand-700 disabled:translate-y-0 disabled:opacity-50"
          :disabled="loading"
          @click="load"
        >
          <RefreshCw :size="14" :class="loading ? 'animate-spin' : ''" />
          Refresh
        </button>
      </div>
      <div aria-hidden="true" class="pointer-events-none absolute -top-20 right-[-3rem] -z-10 size-48 rounded-full bg-white/40 blur-3xl dark:bg-white/5" />
    </section>

    <p v-if="loadError" role="alert" class="mt-5 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ loadError }}
    </p>

    <div v-if="loading && !overview" class="mt-8 flex items-center gap-2 text-[13px] text-ink-900/40">
      <Loader2 :size="16" class="animate-spin" /> Loading the marketplace…
    </div>

    <template v-else-if="overview">
      <div class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <div class="dash-stat dash-stat--brand rounded-2xl border p-4">
          <div class="flex items-center justify-between">
            <p class="text-[11px] font-semibold text-ink-900/55">Providers</p>
            <span class="flex size-8 items-center justify-center rounded-xl bg-white/70 text-brand-600 shadow-sm dark:bg-white/10 dark:text-brand-200">
              <Building2 :size="15" />
            </span>
          </div>
          <p class="mt-3 text-[1.7rem] leading-none font-semibold tracking-[-0.03em]">{{ overview.providers.length }}</p>
          <p class="mt-2 text-[10.5px] text-ink-900/40">Companies on the platform</p>
        </div>
        <div class="dash-stat dash-stat--accent rounded-2xl border p-4">
          <div class="flex items-center justify-between">
            <p class="text-[11px] font-semibold text-ink-900/55">Awaiting approval</p>
            <span class="flex size-8 items-center justify-center rounded-xl bg-white/70 text-accent-600 shadow-sm dark:bg-white/10 dark:text-accent-200">
              <CircleAlert :size="15" />
            </span>
          </div>
          <p class="mt-3 text-[1.7rem] leading-none font-semibold tracking-[-0.03em]">{{ pendingProviders.length }}</p>
          <p class="mt-2 text-[10.5px] text-ink-900/40">Needs a review decision</p>
        </div>
        <div class="dash-stat dash-stat--brand rounded-2xl border p-4">
          <div class="flex items-center justify-between">
            <p class="text-[11px] font-semibold text-ink-900/55">Customers</p>
            <span class="flex size-8 items-center justify-center rounded-xl bg-white/70 text-brand-600 shadow-sm dark:bg-white/10 dark:text-brand-200">
              <Users :size="15" />
            </span>
          </div>
          <p class="mt-3 text-[1.7rem] leading-none font-semibold tracking-[-0.03em]">{{ overview.customers.length }}</p>
          <p class="mt-2 text-[10.5px] text-ink-900/40">Marketplace shippers</p>
        </div>
        <div class="dash-stat dash-stat--success rounded-2xl border p-4">
          <div class="flex items-center justify-between">
            <p class="text-[11px] font-semibold text-ink-900/55">Admin accounts</p>
            <span class="flex size-8 items-center justify-center rounded-xl bg-white/70 text-success-600 shadow-sm dark:bg-white/10 dark:text-success-200">
              <ShieldCheck :size="15" />
            </span>
          </div>
          <p class="mt-3 text-[1.7rem] leading-none font-semibold tracking-[-0.03em]">{{ overview.admins.length }}</p>
          <p class="mt-2 text-[10.5px] text-ink-900/40">Dashboard access holders</p>
        </div>
      </div>

      <div class="dash-panel mt-4 flex gap-1 overflow-x-auto rounded-2xl border p-1.5">
        <button
          v-for="item in tabs"
          :key="item.key"
          type="button"
          class="inline-flex shrink-0 items-center gap-1.5 rounded-xl px-3 py-2 text-[12px] font-semibold transition"
          :class="tab === item.key ? 'bg-gradient-to-r from-brand-600 to-brand-600 text-white shadow-sm shadow-brand-600/25' : 'text-ink-900/55 hover:bg-canvas dark:hover:bg-white/5'"
          @click="tab = item.key"
        >
          <component :is="item.icon" :size="14" />
          {{ item.label }}
          <span
            v-if="item.key === 'providers' && pendingProviders.length"
            class="rounded-full px-1.5 text-[10px] font-semibold"
            :class="tab === 'providers' ? 'bg-white/25' : 'bg-accent-100 text-accent-700 dark:bg-accent-400/15 dark:text-accent-200'"
          >
            {{ pendingProviders.length }}
          </span>
        </button>
      </div>

      <!-- Providers -->
      <section v-if="tab === 'providers'" class="dash-panel mt-3 overflow-hidden rounded-2xl border">
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-ink-900/8 px-4 py-3.5 dark:border-white/10">
          <div class="flex items-center gap-2.5">
            <span class="flex size-8 items-center justify-center rounded-xl bg-brand-100 text-brand-700 dark:bg-brand-400/15 dark:text-brand-200">
              <Store :size="15" />
            </span>
            <div>
              <h2 class="text-[13.5px] font-semibold">Logistics providers</h2>
              <p class="text-[11px] text-ink-900/40">Approve a provider to move it into the marketplace.</p>
            </div>
          </div>
          <p v-if="!canApprove" class="text-[11px] text-ink-900/40">Support accounts have read-only access.</p>
        </div>

        <div v-if="!overview.providers.length" class="px-4 py-10 text-center text-[12.5px] text-ink-900/40">
          No provider companies have signed up yet.
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-[820px] text-left">
            <thead class="bg-canvas/70 text-[10.5px] tracking-[0.06em] text-ink-900/40 uppercase dark:bg-white/5">
              <tr>
                <th class="px-4 py-2.5 font-medium">Company</th>
                <th class="px-4 py-2.5 font-medium">CAC name</th>
                <th class="px-4 py-2.5 font-medium">Owner</th>
                <th class="px-4 py-2.5 font-medium">Status</th>
                <th class="px-4 py-2.5 font-medium">Joined</th>
                <th class="px-4 py-2.5 text-right font-medium">Action</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-ink-900/6">
              <tr v-for="provider in overview.providers" :key="provider.id" class="align-top">
                <td class="px-4 py-3">
                  <p class="text-[12.5px] font-semibold">{{ provider.name }}</p>
                  <p class="mt-0.5 text-[11px] text-ink-900/40">{{ provider.business_email }}</p>
                  <p class="text-[11px] text-ink-900/40">
                    {{ [provider.city, provider.state].filter(Boolean).join(', ') || 'No location' }}
                  </p>
                </td>
                <td class="px-4 py-3 text-[12px]">
                  <span v-if="provider.cac_business_name">{{ provider.cac_business_name }}</span>
                  <span v-else class="text-accent-600">Not provided</span>
                </td>
                <td class="px-4 py-3 text-[12px]">
                  <p>{{ provider.owner_name || '—' }}</p>
                  <p class="text-[11px] text-ink-900/40">{{ provider.owner_email }}</p>
                </td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex rounded-full border px-2 py-0.5 text-[10.5px] font-medium capitalize"
                    :class="statusClass(provider.status)"
                  >
                    {{ provider.status.replace('_', ' ') }}
                  </span>
                </td>
                <td class="px-4 py-3 text-[11.5px] text-ink-900/50">{{ formatDate(provider.created_at) }}</td>
                <td class="px-4 py-3 text-right">
                  <button
                    v-if="provider.status !== 'verified'"
                    type="button"
                    class="inline-flex items-center gap-1.5 rounded-lg bg-gradient-to-r from-brand-600 to-brand-600 px-3 py-1.5 text-[11.5px] font-semibold text-white shadow-sm shadow-brand-600/20 transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-50"
                    :disabled="!canApprove || approvingId === provider.id"
                    @click="approve(provider)"
                  >
                    <Loader2 v-if="approvingId === provider.id" :size="12" class="animate-spin" />
                    Approve
                  </button>
                  <span v-else class="text-[11px] font-medium text-success-600">
                    Verified {{ formatDate(provider.verified_at) }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- Customers -->
      <section v-else-if="tab === 'customers'" class="dash-panel mt-3 overflow-hidden rounded-2xl border">
        <div class="flex items-center gap-2.5 border-b border-ink-900/8 px-4 py-3.5 dark:border-white/10">
          <span class="flex size-8 items-center justify-center rounded-xl bg-brand-100 text-brand-700 dark:bg-brand-400/15 dark:text-brand-200">
            <Users :size="15" />
          </span>
          <div>
            <h2 class="text-[13.5px] font-semibold">Customers</h2>
            <p class="text-[11px] text-ink-900/40">People booking deliveries through the marketplace.</p>
          </div>
        </div>

        <div v-if="!overview.customers.length" class="px-4 py-10 text-center text-[12.5px] text-ink-900/40">
          No customer accounts yet.
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-[680px] text-left">
            <thead class="bg-canvas/70 text-[10.5px] tracking-[0.06em] text-ink-900/40 uppercase dark:bg-white/5">
              <tr>
                <th class="px-4 py-2.5 font-medium">Customer</th>
                <th class="px-4 py-2.5 font-medium">Phone</th>
                <th class="px-4 py-2.5 font-medium">Email verified</th>
                <th class="px-4 py-2.5 font-medium">Joined</th>
                <th class="px-4 py-2.5 font-medium">Last seen</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-ink-900/6">
              <tr v-for="customer in overview.customers" :key="customer.id">
                <td class="px-4 py-3">
                  <p class="text-[12.5px] font-semibold">{{ customer.full_name }}</p>
                  <p class="text-[11px] text-ink-900/40">{{ customer.email }}</p>
                </td>
                <td class="px-4 py-3 text-[11.5px] text-ink-900/60">{{ customer.phone || '—' }}</td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex rounded-full border px-2 py-0.5 text-[10.5px] font-medium"
                    :class="customer.email_verified_at ? statusClass('verified') : statusClass('pending')"
                  >
                    {{ customer.email_verified_at ? 'Verified' : 'Pending' }}
                  </span>
                </td>
                <td class="px-4 py-3 text-[11.5px] text-ink-900/50">{{ formatDate(customer.created_at) }}</td>
                <td class="px-4 py-3 text-[11.5px] text-ink-900/50">{{ formatDate(customer.last_login_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- Admins -->
      <section v-else class="mt-3 grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,360px)]">
        <div class="dash-panel overflow-hidden rounded-2xl border">
          <div class="flex items-center gap-2.5 border-b border-ink-900/8 px-4 py-3.5 dark:border-white/10">
            <span class="flex size-8 items-center justify-center rounded-xl bg-success-100 text-success-700 dark:bg-success-500/15 dark:text-success-200">
              <ShieldCheck :size="15" />
            </span>
            <div>
              <h2 class="text-[13.5px] font-semibold">Admin accounts</h2>
              <p class="text-[11px] text-ink-900/40">Everyone who can sign in to this dashboard.</p>
            </div>
          </div>
          <ul class="divide-y divide-ink-900/6">
            <li v-for="admin in overview.admins" :key="admin.id" class="flex items-center gap-3 px-4 py-3">
              <div class="grid size-8 shrink-0 place-items-center rounded-xl bg-gradient-to-br from-brand-100 to-brand-50 text-brand-700 dark:from-brand-400/15 dark:to-brand-400/10 dark:text-brand-200">
                <ShieldCheck :size="14" />
              </div>
              <div class="min-w-0 flex-1">
                <p class="truncate text-[12.5px] font-semibold">{{ admin.full_name }}</p>
                <p class="truncate text-[11px] text-ink-900/40">{{ admin.email }}</p>
              </div>
              <span class="rounded-full border px-2 py-0.5 text-[10.5px] font-medium capitalize" :class="statusClass('active')">
                {{ admin.platform_role?.replace('_', ' ') }}
              </span>
            </li>
          </ul>
        </div>

        <div class="dash-soft rounded-2xl border p-4">
          <h2 class="flex items-center gap-2 text-[13.5px] font-semibold">
            <span class="flex size-8 items-center justify-center rounded-xl bg-gradient-to-br from-brand-600 to-brand-600 text-white shadow-sm shadow-brand-600/25">
              <UserPlus :size="15" />
            </span>
            Add an admin
          </h2>
          <p v-if="!isSuperAdmin" class="mt-2 text-[11.5px] leading-relaxed text-ink-900/45">
            Only the super administrator can create admin accounts.
          </p>
          <form v-else class="mt-4 space-y-3" @submit.prevent="createAdmin">
            <label class="block">
              <span class="mb-1 block text-[11px] font-medium text-ink-900/55">Full name</span>
              <input v-model="adminForm.full_name" type="text" required minlength="2" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-[12.5px] transition hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5" />
            </label>
            <label class="block">
              <span class="mb-1 block text-[11px] font-medium text-ink-900/55">Email</span>
              <input v-model="adminForm.email" type="email" required class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-[12.5px] transition hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5" />
            </label>
            <label class="block">
              <span class="mb-1 block text-[11px] font-medium text-ink-900/55">Role</span>
              <select v-model="adminForm.role" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-[12.5px] transition hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5">
                <option value="admin">Admin — approve and manage</option>
                <option value="support">Support — read only</option>
              </select>
            </label>
            <label class="block">
              <span class="mb-1 block text-[11px] font-medium text-ink-900/55">Temporary password</span>
              <PasswordInput
                v-model="adminForm.password"
                required
                :minlength="8"
                autocomplete="new-password"
                placeholder="At least 8 characters"
                input-class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-[12.5px] transition hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5"
              />
            </label>

            <p v-if="createError" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[11.5px] text-red-700">
              {{ createError }}
            </p>

            <button
              type="submit"
              :disabled="creatingAdmin"
              class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-brand-600 to-brand-600 px-3 py-2.5 text-[12px] font-semibold text-white shadow-[0_10px_24px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-60"
            >
              <Loader2 v-if="creatingAdmin" :size="14" class="animate-spin" />
              Create admin account
            </button>
            <p class="text-[10.5px] leading-relaxed text-ink-900/35">
              Share the password with them directly. They can change it after signing in.
            </p>
          </form>
        </div>
      </section>
    </template>
  </div>
</template>
