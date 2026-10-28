<script setup lang="ts">
import type { Shipment } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Records' })
definePageMeta({ layout: 'default' })

const route = useRoute()
const { token } = useAuth()
const { canSend, canViewAllBranches, branchId } = usePermissions()
const { options: branchOptions, name: branchName, load: loadBranches } = useBranches()

const pageSize = 24

const records = ref<Shipment[]>([])
const loading = ref(true)
const loadingMore = ref(false)
const error = ref('')
const reachedEnd = ref(false)

const search = ref(typeof route.query.q === 'string' ? route.query.q : '')
const status = ref(typeof route.query.status === 'string' ? route.query.status : '')
const paymentStatus = ref(typeof route.query.payment_status === 'string' ? route.query.payment_status : '')
const branchFilter = ref(typeof route.query.branch_id === 'string' ? route.query.branch_id : '')
const destination = ref(
  typeof route.query.destination_branch_id === 'string' ? route.query.destination_branch_id : '',
)
const outstanding = ref(route.query.outstanding === 'true')

const statusOptions = [
  { value: '', label: 'All' },
  { value: 'pending', label: 'Pending' },
  { value: 'in_transit', label: 'In transit' },
  { value: 'pending,in_transit', label: 'In motion' },
  { value: 'received', label: 'Received' },
  { value: 'exception', label: 'Exception' },
  { value: 'cancelled', label: 'Cancelled' },
]

const paymentOptions = [
  { value: '', label: 'Any payment' },
  { value: 'pending', label: 'Unpaid' },
  { value: 'partial', label: 'Part paid' },
  { value: 'paid', label: 'Paid' },
]

const destinationName = computed(() => branchName(destination.value))
const hasFilters = computed(
  () =>
    Boolean(search.value.trim() || status.value || paymentStatus.value || branchFilter.value || destination.value) ||
    outstanding.value,
)
const emptyTitle = computed(() => (hasFilters.value ? 'No records match those filters' : 'No records yet'))

function queryString(offset: number): string {
  const params = new URLSearchParams()
  if (search.value.trim()) params.set('q', search.value.trim())
  if (status.value) params.set('status', status.value)
  if (paymentStatus.value) params.set('payment_status', paymentStatus.value)
  if (branchFilter.value) params.set('branch_id', branchFilter.value)
  if (destination.value) params.set('destination_branch_id', destination.value)
  if (outstanding.value) params.set('outstanding', 'true')
  params.set('limit', String(pageSize))
  params.set('offset', String(offset))
  return params.toString()
}

async function fetchPage(offset: number) {
  if (!token.value) return
  const page = await apiRequest<Shipment[]>(`/api/records?${queryString(offset)}`, { token: token.value })
  if (offset === 0) records.value = page
  else records.value = [...records.value, ...page]
  reachedEnd.value = page.length < pageSize
}

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    await fetchPage(0)
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not load records. Please try again.'
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadingMore.value || reachedEnd.value) return
  loadingMore.value = true
  try {
    await fetchPage(records.value.length)
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not load more records.'
  } finally {
    loadingMore.value = false
  }
}

function clearFilters() {
  search.value = ''
  status.value = ''
  paymentStatus.value = ''
  branchFilter.value = ''
  destination.value = ''
  outstanding.value = false
}

let searchTimer: ReturnType<typeof setTimeout> | undefined
watch(search, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(refresh, 300)
})
watch([status, paymentStatus, branchFilter, destination, outstanding], refresh)
onBeforeUnmount(() => clearTimeout(searchTimer))

onMounted(async () => {
  await loadBranches()
  await refresh()
})
</script>

<template>
  <div class="space-y-4">
    <header class="flex flex-wrap items-end justify-between gap-3">
      <div class="min-w-0">
        <h1 class="text-h5 font-medium tracking-tight text-ink">Records</h1>
        <p class="mt-1 text-label text-ink-64">
          Every item, its money and where it is on the route.
        </p>
      </div>
      <UiButton v-if="canSend" @click="navigateTo('/records/new')">
        <UiIcon name="plus" :size="17" />
        Send parcel
      </UiButton>
    </header>

    <div class="space-y-3">
      <UiSearchBar v-model="search" />

      <div class="-mx-1 flex gap-1.5 overflow-x-auto px-1 pb-1">
        <button
          v-for="option in statusOptions"
          :key="option.value"
          type="button"
          class="h-8 shrink-0 rounded-full px-3 text-label transition"
          :class="
            status === option.value
              ? 'bg-ink font-medium text-white'
              : 'bg-surface text-ink-64 shadow-light hover:text-ink'
          "
          :aria-pressed="status === option.value"
          @click="status = option.value"
        >
          {{ option.label }}
        </button>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <UiSelect
          id="payment-filter"
          v-model="paymentStatus"
          label="Payment filter"
          hide-label
          :options="paymentOptions"
          class="w-[160px]"
        />
        <UiSelect
          v-if="canViewAllBranches"
          id="branch-filter"
          v-model="branchFilter"
          label="Branch filter"
          hide-label
          :options="[{ value: '', label: 'All branches' }, ...branchOptions]"
          class="w-[190px]"
        />
        <button
          type="button"
          class="h-10 rounded-lg px-3 text-label shadow-light transition"
          :class="outstanding ? 'bg-ink font-medium text-white' : 'bg-surface text-ink-64 hover:text-ink'"
          :aria-pressed="outstanding"
          @click="outstanding = !outstanding"
        >
          Has balance
        </button>
        <UiButton v-if="hasFilters" variant="link" @click="clearFilters">Clear</UiButton>
      </div>

      <div v-if="destination" class="flex items-center gap-2">
        <span class="inline-flex items-center gap-1.5 rounded-full bg-chip/50 px-3 py-1 text-meta text-ink">
          Destined for {{ destinationName || 'your branch' }}
          <button type="button" class="text-ink-64 hover:text-ink" aria-label="Clear destination filter" @click="destination = ''">
            <UiIcon name="close" :size="12" />
          </button>
        </span>
      </div>
    </div>

    <div v-if="error" class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-danger/25 bg-danger/5 px-4 py-3">
      <p class="text-label text-danger">{{ error }}</p>
      <UiButton variant="tertiary" :loading="loading" @click="refresh">Try again</UiButton>
    </div>

    <div v-if="loading" class="grid gap-3 sm:grid-cols-2" aria-hidden="true">
      <div v-for="index in 4" :key="index" class="h-[168px] animate-pulse rounded-lg bg-surface shadow-card" />
    </div>

    <template v-else>
      <div v-if="records.length" class="grid gap-3 sm:grid-cols-2">
        <RecordCard v-for="record in records" :key="record.id" :record="record" />
      </div>

      <UiEmptyState v-else :title="emptyTitle" :description="hasFilters ? 'Try a different search or clear the filters.' : 'Send the first parcel and it will appear here.'">
        <UiButton v-if="hasFilters" variant="tertiary" @click="clearFilters">Clear filters</UiButton>
        <UiButton v-else-if="canSend" @click="navigateTo('/records/new')">Send parcel</UiButton>
      </UiEmptyState>

      <div v-if="records.length" class="flex flex-col items-center gap-2 pt-1">
        <p class="text-meta text-ink-40">
          Showing {{ records.length }} record{{ records.length === 1 ? '' : 's' }}{{ reachedEnd ? '' : ' so far' }}
        </p>
        <UiButton v-if="!reachedEnd" variant="tertiary" :loading="loadingMore" @click="loadMore">Load more</UiButton>
      </div>
    </template>
  </div>
</template>
