<script setup lang="ts">
import type { ActivityLog } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Activity' })
definePageMeta({ layout: 'default', roles: ['owner', 'manager', 'system_admin'] })

const { token } = useAuth()
const { canViewAllBranches } = usePermissions()
const { options: branchOptions, load: loadBranches } = useBranches()

const logs = ref<ActivityLog[]>([])
const loading = ref(true)
const error = ref('')
const entityType = ref('')
const branchFilter = ref('')

const entityOptions = [
  { value: '', label: 'All activity' },
  { value: 'manifest_item', label: 'Shipments' },
  { value: 'exception', label: 'Exceptions' },
]

/** Plain-English wording for every action written to the log. */
const ACTION_LABELS: Record<string, string> = {
  record_created: 'recorded a shipment',
  record_updated: 'edited a shipment',
  record_dispatched: 'dispatched a shipment',
  record_received: 'received a shipment',
  record_delivered: 'handed a shipment over',
  record_cancelled: 'cancelled a shipment',
  payment_recorded: 'recorded a payment',
  exception_raised: 'raised an exception',
  exception_updated: 'updated an exception',
}

function actionLabel(action: string): string {
  return ACTION_LABELS[action] ?? humanise(action)
}

function detailValue(log: ActivityLog, key: string): string {
  const value = log.details?.[key]
  return value === undefined || value === null ? '' : String(value)
}

async function load() {
  if (!token.value) return
  loading.value = true
  error.value = ''
  const params = new URLSearchParams({ limit: '100' })
  if (entityType.value) params.set('entity_type', entityType.value)
  if (branchFilter.value) params.set('branch_id', branchFilter.value)
  try {
    logs.value = await apiRequest<ActivityLog[]>(`/api/activity?${params.toString()}`, {
      token: token.value,
    })
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not load the activity log.'
  } finally {
    loading.value = false
  }
}

watch([entityType, branchFilter], load)

onMounted(async () => {
  await loadBranches()
  await load()
})
</script>

<template>
  <div class="space-y-4">
    <header>
      <h1 class="text-h5 font-medium tracking-tight text-ink">Activity</h1>
      <p class="mt-1 text-label text-ink-64">
        {{
          canViewAllBranches
            ? 'Everything staff have done across the branches, newest first.'
            : 'Everything your branch staff have done, newest first.'
        }}
      </p>
    </header>

    <div class="flex flex-wrap items-center gap-2">
      <UiSelect
        id="activity-entity"
        v-model="entityType"
        label="Activity type"
        hide-label
        :options="entityOptions"
        class="w-[180px]"
      />
      <UiSelect
        v-if="canViewAllBranches"
        id="activity-branch"
        v-model="branchFilter"
        label="Branch filter"
        hide-label
        :options="[{ value: '', label: 'All branches' }, ...branchOptions]"
        class="w-[190px]"
      />
    </div>

    <div
      v-if="error"
      class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-danger/25 bg-danger/5 px-4 py-3"
    >
      <p class="text-label text-danger">{{ error }}</p>
      <UiButton variant="tertiary" :loading="loading" @click="load">Try again</UiButton>
    </div>

    <div v-if="loading" class="space-y-2" aria-hidden="true">
      <div v-for="index in 4" :key="index" class="h-[76px] animate-pulse rounded-lg bg-surface shadow-card" />
    </div>

    <template v-else>
      <ul v-if="logs.length" class="space-y-2">
        <li
          v-for="log in logs"
          :key="log.id"
          class="flex items-start gap-3 rounded-lg bg-surface p-4 shadow-card"
        >
          <span
            class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-canvas text-blue"
            aria-hidden="true"
          >
            <UiIcon :name="log.entity_type === 'exception' ? 'alert' : 'package'" :size="17" />
          </span>
          <div class="min-w-0 flex-1">
            <p class="text-p1 text-ink">
              <span class="font-medium">{{ log.user_name || 'Staff' }}</span>
              {{ actionLabel(log.action) }}
            </p>
            <p class="mt-0.5 flex flex-wrap items-center gap-x-1.5 text-meta text-ink-64">
              <span v-if="detailValue(log, 'tracking_number')">{{ detailValue(log, 'tracking_number') }}</span>
              <span
                v-if="detailValue(log, 'tracking_number') && log.branch_name"
                class="text-ink-40"
                aria-hidden="true"
              >·</span>
              <span v-if="log.branch_name">{{ log.branch_name }}</span>
            </p>
            <p class="mt-0.5 text-meta text-ink-40">{{ shortDateTime(log.created_at) }}</p>
          </div>
        </li>
      </ul>

      <UiEmptyState
        v-else
        icon="clock"
        title="No activity yet"
        description="Recording, dispatching and receiving shipments all show up here."
      />
    </template>
  </div>
</template>
