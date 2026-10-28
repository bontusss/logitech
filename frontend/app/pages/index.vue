<script setup lang="ts">
import type { BranchReportRow, DailyReportRow, ReportOverview, Shipment, ShipmentException } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Dashboard' })
definePageMeta({ layout: 'default' })

const { token, user } = useAuth()
const { branchId, isOwner, isManager, isStaff, isSystemAdmin, canViewReports, canSend, canReceiveGoods } = usePermissions()
const { name: branchName, load: loadBranches } = useBranches()

const loading = ref(true)
const error = ref('')
const overview = ref<ReportOverview | null>(null)
const daily = ref<DailyReportRow[]>([])
const branchRows = ref<BranchReportRow[]>([])
const recent = ref<Shipment[]>([])
const incoming = ref<Shipment[]>([])
const balances = ref<Shipment[]>([])
const exceptions = ref<ShipmentException[]>([])

const branch = computed(() => branchName(branchId.value))

const roleLine = computed(() => {
  if (isOwner.value) return 'Owner · every branch'
  if (isManager.value) return branch.value ? `Branch manager · ${branch.value}` : 'Branch manager'
  if (isStaff.value) {
    const job = humanise(user.value?.job_role ?? 'staff')
    return branch.value ? `${job} · ${branch.value}` : job
  }
  return 'System admin'
})

const periodLabel = computed(() =>
  overview.value ? `${shortDate(overview.value.period.from)} – ${shortDate(overview.value.period.to)}` : 'Last 30 days',
)

const openExceptions = computed(() =>
  exceptions.value.filter((item) => item.status === 'open' || item.status === 'investigating'),
)

const attentionCount = computed(() =>
  overview.value ? overview.value.exceptions.open + overview.value.exceptions.investigating : openExceptions.value.length,
)

const receiveLink = computed(() =>
  branchId.value ? `/records?status=pending,in_transit&destination_branch_id=${branchId.value}` : '/records?status=pending,in_transit',
)

/** One failing panel must not blank the whole dashboard. */
async function safe<T>(run: () => Promise<T>, assign: (value: T) => void) {
  try {
    assign(await run())
  } catch {
    // Panel stays empty; the rest of the dashboard still renders.
  }
}

async function load() {
  if (!token.value || !user.value) return
  loading.value = true
  error.value = ''

  const t = token.value
  const period = lastDays(30)
  const range = `from=${period.from}&to=${period.to}`
  const optional: Promise<unknown>[] = []

  try {
    if (canViewReports.value) {
      optional.push(
        safe(
          () => apiRequest<ReportOverview>(`/api/reports/overview?${range}`, { token: t }),
          (value) => (overview.value = value),
        ),
        safe(
          () => apiRequest<DailyReportRow[]>(`/api/reports/daily?${range}`, { token: t }),
          (value) => (daily.value = value),
        ),
        safe(
          () => apiRequest<Shipment[]>('/api/records?outstanding=true&limit=5', { token: t }),
          (value) => (balances.value = value),
        ),
      )
      if (isOwner.value) {
        optional.push(
          safe(
            () => apiRequest<BranchReportRow[]>(`/api/reports/branches?${range}`, { token: t }),
            (value) => (branchRows.value = value),
          ),
        )
      }
    }

    if (branchId.value && !isOwner.value) {
      optional.push(
        safe(
          () =>
            apiRequest<Shipment[]>(
              `/api/records?destination_branch_id=${branchId.value}&status=pending,in_transit&limit=5`,
              { token: t },
            ),
          (value) => (incoming.value = value),
        ),
      )
    }

    optional.push(
      safe(
        () => apiRequest<ShipmentException[]>('/api/exceptions?limit=20', { token: t }),
        (value) => (exceptions.value = value),
      ),
    )

    // The record feed is the one call that has to work: if it fails the
    // session or the API is down, so is the dashboard.
    recent.value = await apiRequest<Shipment[]>('/api/records?limit=6', { token: t })
    await Promise.all(optional)
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not load the dashboard. Please try again.'
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadBranches()
  await load()
})
</script>

<template>
  <div class="space-y-5">
    <header class="flex flex-wrap items-end justify-between gap-3">
      <div class="min-w-0">
        <h1 class="text-h5 font-medium tracking-tight text-ink">Dashboard</h1>
        <p class="mt-1 text-label text-ink-64">{{ roleLine }}</p>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <UiButton v-if="canReceiveGoods" variant="tertiary" @click="navigateTo(receiveLink)">
          <UiIcon name="receive" :size="17" />
          Receive
        </UiButton>
        <UiButton v-if="canSend" @click="navigateTo('/records/new')">
          <UiIcon name="plus" :size="17" />
          Send parcel
        </UiButton>
      </div>
    </header>

    <div v-if="error" class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-danger/25 bg-danger/5 px-4 py-3">
      <p class="text-label text-danger">{{ error }}</p>
      <UiButton variant="tertiary" :loading="loading" @click="load">Try again</UiButton>
    </div>

    <div v-if="loading" class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4" aria-hidden="true">
      <div v-for="index in 4" :key="index" class="h-[108px] animate-pulse rounded-lg bg-surface shadow-card" />
    </div>

    <template v-else>
      <!-- Metrics for the owner and branch managers -->
      <section v-if="canViewReports && overview" class="space-y-2">
        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <UiStat
            label="Shipments"
            :value="String(overview.shipments.total)"
            :hint="`${overview.shipments.quantity} items · ${overview.shipments.received} received`"
            icon="package"
          />
          <UiStat
            label="Collected"
            :value="money(overview.money.collected)"
            :hint="`${overview.money.collection_rate.toFixed(1)}% of ${money(overview.money.billed)} billed`"
            icon="wallet"
            tone="info"
          />
          <UiStat
            label="Outstanding"
            :value="money(overview.money.outstanding)"
            hint="Unpaid balances on live records"
            icon="receipt"
            :tone="overview.money.outstanding > 0 ? 'danger' : 'default'"
          />
          <UiStat
            label="Open exceptions"
            :value="String(attentionCount)"
            :hint="`${overview.exceptions.investigating} investigating · ${overview.exceptions.resolved} resolved`"
            icon="alert"
            :tone="attentionCount > 0 ? 'danger' : 'default'"
          />
        </div>
        <p class="text-meta text-ink-40">
          {{ periodLabel }} · collected today {{ money(overview.money.collected_today) }}
        </p>
      </section>

      <!-- Staff see their queue first -->
      <section v-if="!canViewReports && !isSystemAdmin" class="grid gap-3 sm:grid-cols-2">
        <UiStat
          :label="`Incoming to ${branch || 'your branch'}`"
          :value="String(incoming.length)"
          hint="Pending and in transit"
          icon="receive"
          :tone="incoming.length > 0 ? 'info' : 'default'"
        />
        <UiStat
          label="Open exceptions"
          :value="String(openExceptions.length)"
          hint="At your branch"
          icon="alert"
          :tone="openExceptions.length > 0 ? 'danger' : 'default'"
        />
      </section>

      <!-- Admin setup shortcuts -->
      <section v-if="isSystemAdmin" class="rounded-lg bg-surface p-4 shadow-card">
        <h2 class="text-h5 font-medium text-ink">Set up the business</h2>
        <p class="mt-0.5 text-label text-ink-64">Branches, accounts and the fleet behind every shipment.</p>
        <div class="mt-3 grid gap-2 sm:grid-cols-2">
          <NuxtLink
            v-for="item in [
              { to: '/branches', label: 'Branches', icon: 'building' },
              { to: '/staff', label: 'Staff', icon: 'users' },
              { to: '/vehicles', label: 'Vehicles', icon: 'truck' },
              { to: '/drivers', label: 'Drivers', icon: 'user' },
            ]"
            :key="item.to"
            :to="item.to"
            class="flex items-center gap-2.5 rounded-lg border border-line px-3 py-2.5 text-label font-medium text-ink transition hover:border-blue/40 hover:text-blue"
          >
            <UiIcon :name="item.icon" :size="17" />
            {{ item.label }}
          </NuxtLink>
        </div>
      </section>

      <div class="grid gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
        <div class="space-y-4">
          <!-- Receiving queue -->
          <section v-if="incoming.length" class="rounded-lg bg-surface p-4 shadow-card">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-h5 font-medium text-ink">Incoming to {{ branch || 'your branch' }}</h2>
                <p class="mt-0.5 text-label text-ink-64">Confirm delivery or raise an exception.</p>
              </div>
              <NuxtLink :to="receiveLink" class="shrink-0 text-label font-medium text-blue hover:underline">
                View all
              </NuxtLink>
            </div>
            <ul class="mt-3 divide-y divide-line">
              <li v-for="record in incoming" :key="record.id">
                <NuxtLink
                  :to="`/records/${record.id}`"
                  class="flex items-center justify-between gap-3 py-2.5 transition hover:text-blue"
                >
                  <div class="min-w-0">
                    <p class="truncate text-p1 font-medium text-ink">{{ record.description }}</p>
                    <p class="truncate text-meta text-ink-40">
                      {{ record.tracking_number }} · from {{ record.origin_branch }} · {{ record.receiver.full_name }}
                    </p>
                  </div>
                  <StatusPill :status="record.status" />
                </NuxtLink>
              </li>
            </ul>
          </section>

          <!-- Money still owed -->
          <section v-if="canViewReports && balances.length" class="rounded-lg bg-surface p-4 shadow-card">
            <div class="flex items-center justify-between gap-3">
              <h2 class="text-h5 font-medium text-ink">Balances to collect</h2>
              <NuxtLink to="/records?outstanding=true" class="shrink-0 text-label font-medium text-blue hover:underline">
                View all
              </NuxtLink>
            </div>
            <ul class="mt-3 divide-y divide-line">
              <li v-for="record in balances" :key="record.id">
                <NuxtLink
                  :to="`/records/${record.id}`"
                  class="flex items-center justify-between gap-3 py-2.5 transition hover:text-blue"
                >
                  <div class="min-w-0">
                    <p class="truncate text-p1 font-medium text-ink">{{ record.description }}</p>
                    <p class="truncate text-meta text-ink-40">
                      {{ record.tracking_number }} · {{ record.receiver.full_name }} → {{ record.destination_branch }}
                    </p>
                  </div>
                  <div class="shrink-0 text-right">
                    <p class="text-label font-medium text-danger">{{ money(record.balance, record.currency) }}</p>
                    <p class="text-meta text-ink-40">of {{ money(record.amount, record.currency) }}</p>
                  </div>
                </NuxtLink>
              </li>
            </ul>
          </section>

          <!-- Latest activity -->
          <section class="rounded-lg bg-surface p-4 shadow-card">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-h5 font-medium text-ink">Recent records</h2>
                <p class="mt-0.5 text-label text-ink-64">The newest shipments in your view.</p>
              </div>
              <NuxtLink to="/records" class="shrink-0 text-label font-medium text-blue hover:underline">
                All records
              </NuxtLink>
            </div>
            <div v-if="recent.length" class="mt-3 grid gap-3 sm:grid-cols-2">
              <RecordCard v-for="record in recent" :key="record.id" :record="record" />
            </div>
            <div v-else class="mt-3">
              <UiEmptyState
                title="No records yet"
                description="Records you create or receive will show up here."
              >
                <UiButton v-if="canSend" @click="navigateTo('/records/new')">Send the first parcel</UiButton>
              </UiEmptyState>
            </div>
          </section>
        </div>

        <div class="space-y-4">
          <!-- Daily collections -->
          <section v-if="canViewReports" class="rounded-lg bg-surface p-4 shadow-card">
            <div class="flex items-start justify-between gap-3">
              <div>
                <h2 class="text-h5 font-medium text-ink">Collections</h2>
                <p class="mt-0.5 text-label text-ink-64">Payments taken, last 14 days.</p>
              </div>
              <NuxtLink to="/reports" class="shrink-0 text-label font-medium text-blue hover:underline">Reports</NuxtLink>
            </div>
            <div class="mt-3">
              <DailyBars :rows="daily" metric="collected" :days="14" />
            </div>
          </section>

          <!-- Exceptions -->
          <section class="rounded-lg bg-surface p-4 shadow-card">
            <div class="flex items-center justify-between gap-3">
              <h2 class="text-h5 font-medium text-ink">Needs attention</h2>
              <NuxtLink to="/exceptions" class="shrink-0 text-label font-medium text-blue hover:underline">
                All
              </NuxtLink>
            </div>
            <ul v-if="openExceptions.length" class="mt-2 divide-y divide-line">
              <li v-for="item in openExceptions.slice(0, 5)" :key="item.id">
                <NuxtLink :to="`/exceptions/${item.id}`" class="block py-2.5 transition hover:text-blue">
                  <div class="flex items-start justify-between gap-2">
                    <p class="min-w-0 truncate text-label font-medium text-ink">{{ item.title }}</p>
                    <StatusPill :status="item.status" kind="exception" />
                  </div>
                  <p class="mt-0.5 truncate text-meta text-ink-40">
                    {{ item.exception_number }} · {{ item.tracking_number }}
                  </p>
                </NuxtLink>
              </li>
            </ul>
            <p v-else class="mt-3 text-label text-ink-64">Nothing reported. Every shipment is moving cleanly.</p>
          </section>
        </div>
      </div>

      <!-- Branch comparison for the owner -->
      <section v-if="isOwner && branchRows.length" class="rounded-lg bg-surface p-4 shadow-card">
        <div class="flex items-end justify-between gap-3">
          <div>
            <h2 class="text-h5 font-medium text-ink">Branches</h2>
            <p class="mt-0.5 text-label text-ink-64">Sent, received and collected per branch.</p>
          </div>
          <NuxtLink to="/reports" class="shrink-0 text-label font-medium text-blue hover:underline">Full report</NuxtLink>
        </div>
        <div class="mt-3">
          <BranchTable :rows="branchRows" />
        </div>
      </section>
    </template>
  </div>
</template>
