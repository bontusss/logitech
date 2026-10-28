<script setup lang="ts">
import type { ShipmentException } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Exceptions' })
definePageMeta({ layout: 'default' })

const { token } = useAuth()
const { canViewAllBranches } = usePermissions()
const { options: branchOptions, load: loadBranches } = useBranches()

const exceptions = ref<ShipmentException[]>([])
const loading = ref(true)
const error = ref('')
const status = ref('open')
const severity = ref('')
const branchFilter = ref('')

const statusOptions = [
  { value: 'open', label: 'Open' },
  { value: 'investigating', label: 'Investigating' },
  { value: 'resolved', label: 'Resolved' },
  { value: 'closed', label: 'Closed' },
  { value: '', label: 'All' },
]

const severityOptions = [
  { value: '', label: 'Any severity' },
  { value: 'critical', label: 'Critical' },
  { value: 'high', label: 'High' },
  { value: 'medium', label: 'Medium' },
  { value: 'low', label: 'Low' },
]

const severityTone = (value: string) =>
  value === 'critical' || value === 'high' ? 'danger' : value === 'medium' ? 'info' : 'neutral'

async function load() {
  if (!token.value) return
  loading.value = true
  error.value = ''
  const params = new URLSearchParams({ limit: '100' })
  if (status.value) params.set('status', status.value)
  if (severity.value) params.set('severity', severity.value)
  if (branchFilter.value) params.set('branch_id', branchFilter.value)
  try {
    exceptions.value = await apiRequest<ShipmentException[]>(`/api/exceptions?${params.toString()}`, {
      token: token.value,
    })
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not load exceptions.'
  } finally {
    loading.value = false
  }
}

watch([status, severity, branchFilter], load)

onMounted(async () => {
  await loadBranches()
  await load()
})
</script>

<template>
  <div class="space-y-4">
    <header>
      <h1 class="text-h5 font-medium tracking-tight text-ink">Exceptions</h1>
      <p class="mt-1 text-label text-ink-64">Goods that arrived late, wrong, incomplete or damaged.</p>
    </header>

    <div class="space-y-3">
      <div class="-mx-1 flex gap-1.5 overflow-x-auto px-1 pb-1">
        <button
          v-for="option in statusOptions"
          :key="option.value"
          type="button"
          class="h-8 shrink-0 rounded-full px-3 text-label transition"
          :class="status === option.value ? 'bg-ink font-medium text-white' : 'bg-surface text-ink-64 shadow-light hover:text-ink'"
          :aria-pressed="status === option.value"
          @click="status = option.value"
        >
          {{ option.label }}
        </button>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <UiSelect
          id="severity-filter"
          v-model="severity"
          label="Severity filter"
          hide-label
          :options="severityOptions"
          class="w-[170px]"
        />
        <UiSelect
          v-if="canViewAllBranches"
          id="exception-branch"
          v-model="branchFilter"
          label="Branch filter"
          hide-label
          :options="[{ value: '', label: 'All branches' }, ...branchOptions]"
          class="w-[190px]"
        />
      </div>
    </div>

    <div v-if="error" class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-danger/25 bg-danger/5 px-4 py-3">
      <p class="text-label text-danger">{{ error }}</p>
      <UiButton variant="tertiary" :loading="loading" @click="load">Try again</UiButton>
    </div>

    <div v-if="loading" class="space-y-2" aria-hidden="true">
      <div v-for="index in 3" :key="index" class="h-[92px] animate-pulse rounded-lg bg-surface shadow-card" />
    </div>

    <template v-else>
      <ul v-if="exceptions.length" class="space-y-2">
        <li v-for="item in exceptions" :key="item.id">
          <NuxtLink
            :to="`/exceptions/${item.id}`"
            class="block rounded-lg bg-surface p-4 shadow-card transition hover:shadow-halo"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <p class="text-meta tracking-wide text-ink-40">{{ item.exception_number }}</p>
                <p class="mt-0.5 truncate text-p1 font-medium text-ink">{{ item.title }}</p>
              </div>
              <div class="flex shrink-0 flex-wrap items-center justify-end gap-1.5">
                <UiPill :tone="severityTone(item.severity)">{{ humanise(item.severity) }}</UiPill>
                <StatusPill :status="item.status" kind="exception" />
              </div>
            </div>
            <p class="mt-1.5 flex flex-wrap items-center gap-x-1.5 text-meta text-ink-64">
              <span>{{ item.tracking_number }}</span>
              <span class="text-ink-40">·</span>
              <span>{{ humanise(item.type) }}</span>
              <span class="text-ink-40">·</span>
              <span>{{ item.branch_name }}</span>
              <span class="text-ink-40">·</span>
              <span>{{ relativeDay(item.created_at) }}</span>
            </p>
            <p class="mt-0.5 truncate text-meta text-ink-40">Reported by {{ item.reported_by_name || 'staff' }}</p>
          </NuxtLink>
        </li>
      </ul>

      <UiEmptyState
        v-else
        icon="alert"
        :title="status ? `No ${status} exceptions` : 'No exceptions'"
        description="When receiving staff flag missing, damaged or wrong goods, it appears here."
      />
    </template>
  </div>
</template>
