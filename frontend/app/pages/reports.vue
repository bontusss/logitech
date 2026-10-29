<script setup lang="ts">
import type { BranchReportRow, DailyReportRow, ReportOverview } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Reports' })
definePageMeta({ layout: 'default' })

const { token } = useAuth()
const { branchId, canViewAllBranches } = usePermissions()
const { options: branchOptions, name: branchName, load: loadBranches } = useBranches()

const overview = ref<ReportOverview | null>(null)
const daily = ref<DailyReportRow[]>([])
const branchRows = ref<BranchReportRow[]>([])
const loading = ref(true)
const error = ref('')

const periods = [
  { days: 7, label: '7 days' },
  { days: 30, label: '30 days' },
  { days: 90, label: '90 days' },
]
const days = ref(30)
/** Empty means every branch (owner / system admin only). */
const branchFilter = ref('')

const scopeLabel = computed(() => {
  if (canViewAllBranches.value) {
    return branchName(branchFilter.value) || 'All branches'
  }
  return branchName(branchId.value) || 'Your branch'
})
const range = computed(() => lastDays(days.value))
const maxMethod = computed(() => Math.max(1, ...(overview.value?.payments_by_method.map((m) => m.amount) ?? [1])))
const exceptionRows = computed(() => {
  const stats = overview.value?.exceptions
  if (!stats) return []
  return [
    { label: 'Open', value: stats.open, tone: 'danger' as const },
    { label: 'Investigating', value: stats.investigating, tone: 'info' as const },
    { label: 'Resolved', value: stats.resolved, tone: 'neutral' as const },
    { label: 'Closed', value: stats.closed, tone: 'neutral' as const },
  ]
})

async function load() {
  if (!token.value) return
  loading.value = true
  error.value = ''
  const params = new URLSearchParams({ from: range.value.from, to: range.value.to })
  if (branchFilter.value) params.set('branch_id', branchFilter.value)
  const query = params.toString()
  try {
    const [summary, rows, branches] = await Promise.all([
      apiRequest<ReportOverview>(`/api/reports/overview?${query}`, { token: token.value }),
      apiRequest<DailyReportRow[]>(`/api/reports/daily?${query}`, { token: token.value }),
      apiRequest<BranchReportRow[]>(`/api/reports/branches?${query}`, { token: token.value }),
    ])
    overview.value = summary
    daily.value = rows
    branchRows.value = branches
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not load the reports.'
  } finally {
    loading.value = false
  }
}

watch([days, branchFilter], load)

onMounted(async () => {
  await loadBranches()
  await load()
})
</script>

<template>
  <div class="space-y-5">
    <header class="flex flex-wrap items-end justify-between gap-3">
      <div class="min-w-0">
        <h1 class="text-h5 font-medium tracking-tight text-ink">Reports</h1>
        <p class="mt-1 text-label text-ink-64">{{ scopeLabel }} · {{ shortDate(range.from) }} – {{ shortDate(range.to) }}</p>
      </div>
      <div class="flex flex-wrap items-center gap-1.5">
        <UiSelect
          v-if="canViewAllBranches"
          id="report-branch"
          v-model="branchFilter"
          label="Branch"
          hide-label
          :options="[{ value: '', label: 'All branches' }, ...branchOptions]"
          class="w-[210px]"
        />
        <button
          v-for="period in periods"
          :key="period.days"
          type="button"
          class="h-8 rounded-full px-3 text-label transition"
          :class="days === period.days ? 'bg-ink font-medium text-white' : 'bg-surface text-ink-64 shadow-light hover:text-ink'"
          :aria-pressed="days === period.days"
          @click="days = period.days"
        >
          {{ period.label }}
        </button>
      </div>
    </header>

    <div v-if="error" class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-danger/25 bg-danger/5 px-4 py-3">
      <p class="text-label text-danger">{{ error }}</p>
      <UiButton variant="tertiary" :loading="loading" @click="load">Try again</UiButton>
    </div>

    <div v-if="loading" class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4" aria-hidden="true">
      <div v-for="index in 4" :key="index" class="h-[108px] animate-pulse rounded-lg bg-surface shadow-card" />
    </div>

    <template v-else-if="overview">
      <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <UiStat
          label="Waybills"
          :value="String(overview.waybills?.total ?? overview.shipments.total)"
          :hint="`${overview.waybills?.quantity ?? overview.shipments.quantity} items · ${overview.waybills?.received ?? overview.shipments.received} received`"
          icon="package"
        />
        <UiStat label="Billed" :value="money(overview.money.billed)" :hint="`${overview.money.payments_count} payments received`" icon="receipt" />
        <UiStat
          label="Collected"
          :value="money(overview.money.collected)"
          :hint="`${overview.money.collection_rate.toFixed(1)}% collection rate`"
          icon="wallet"
          tone="info"
        />
        <UiStat
          label="Outstanding"
          :value="money(overview.money.outstanding)"
          hint="Unpaid balances, all time"
          icon="alert"
          :tone="overview.money.outstanding > 0 ? 'danger' : 'default'"
        />
      </section>

      <div class="grid gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
        <section class="rounded-lg bg-surface p-4 shadow-card">
          <div class="flex items-end justify-between gap-3">
            <div>
              <h2 class="text-h5 font-medium text-ink">Daily collections</h2>
              <p class="mt-0.5 text-label text-ink-64">Payments taken each day in this period.</p>
            </div>
            <p class="text-label text-ink-64">Today {{ money(overview.money.collected_today) }}</p>
          </div>
          <div class="mt-4">
            <DailyBars :rows="daily" metric="collected" :days="days" />
          </div>

          <div class="mt-6 border-t border-line pt-4">
            <h3 class="text-h5 font-medium text-ink">Payment methods</h3>
            <ul v-if="overview.payments_by_method.length" class="mt-3 space-y-3">
              <li v-for="method in overview.payments_by_method" :key="method.method">
                <div class="flex items-baseline justify-between gap-3">
                  <p class="text-label text-ink">{{ humanise(method.method) }}</p>
                  <p class="text-label font-medium text-ink">{{ money(method.amount) }}</p>
                </div>
                <div class="mt-1.5 h-1.5 overflow-hidden rounded-full bg-muted">
                  <div class="h-full rounded-full bg-blue" :style="{ width: `${Math.round((method.amount / maxMethod) * 100)}%` }" />
                </div>
                <p class="mt-1 text-meta text-ink-40">{{ method.count }} payment{{ method.count === 1 ? '' : 's' }}</p>
              </li>
            </ul>
            <p v-else class="mt-2 text-label text-ink-64">No payments in this period.</p>
          </div>
        </section>

        <div class="space-y-4">
          <section class="rounded-lg bg-surface p-4 shadow-card">
            <h2 class="text-h5 font-medium text-ink">Exceptions</h2>
            <dl class="mt-2 divide-y divide-line">
              <div v-for="row in exceptionRows" :key="row.label" class="flex items-center justify-between gap-3 py-2.5">
                <dt class="text-label text-ink-64">{{ row.label }}</dt>
                <dd class="text-label font-medium" :class="row.tone === 'danger' && row.value > 0 ? 'text-danger' : 'text-ink'">
                  {{ row.value }}
                </dd>
              </div>
            </dl>
            <NuxtLink to="/exceptions" class="mt-3 inline-block text-label font-medium text-blue hover:underline">
              Open the exception queue
            </NuxtLink>
          </section>

          <section class="rounded-lg bg-surface p-4 shadow-card">
            <h2 class="text-h5 font-medium text-ink">Waybill mix</h2>
            <dl class="mt-2 divide-y divide-line">
              <div v-for="(count, status) in overview.shipments.by_status" :key="status" class="flex items-center justify-between gap-3 py-2.5">
                <dt class="text-label text-ink-64">{{ humanise(status) }}</dt>
                <dd class="text-label font-medium text-ink">{{ count }}</dd>
              </div>
            </dl>
          </section>
        </div>
      </div>

      <section class="rounded-lg bg-surface p-4 shadow-card">
        <div class="flex items-end justify-between gap-3">
          <div>
            <h2 class="text-h5 font-medium text-ink">Branch performance</h2>
            <p class="mt-0.5 text-label text-ink-64">Sent, received, billed and collected per branch.</p>
          </div>
        </div>
        <div class="mt-3">
          <BranchTable :rows="branchRows" />
        </div>
      </section>
    </template>
  </div>
</template>
