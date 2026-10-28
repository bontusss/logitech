<script setup lang="ts">
import type { ShipmentException } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Exception' })
definePageMeta({ layout: 'default' })

const route = useRoute()
const { token } = useAuth()

const exception = ref<ShipmentException | null>(null)
const loading = ref(true)
const error = ref('')

const comment = ref('')
const nextStatus = ref('')
const resolutionNotes = ref('')
const saving = ref(false)
const updateError = ref('')

const exceptionId = computed(() => String(route.params.id))
const isOpen = computed(() => exception.value?.status === 'open' || exception.value?.status === 'investigating')

const statusOptions = [
  { value: '', label: 'No status change' },
  { value: 'investigating', label: 'Investigating' },
  { value: 'resolved', label: 'Resolved' },
  { value: 'closed', label: 'Closed' },
]

const severityTone = computed(() => {
  const severity = exception.value?.severity
  return severity === 'critical' || severity === 'high' ? 'danger' : severity === 'medium' ? 'info' : 'neutral'
})

async function load() {
  if (!token.value) return
  loading.value = true
  error.value = ''
  try {
    exception.value = await apiRequest<ShipmentException>(`/api/exceptions/${exceptionId.value}`, {
      token: token.value,
    })
  } catch (err) {
    error.value =
      err instanceof ApiError && err.status === 404
        ? 'That exception does not exist, or it is outside your branch.'
        : err instanceof ApiError
          ? err.message
          : 'Could not load the exception.'
  } finally {
    loading.value = false
  }
}

async function postUpdate(status?: string) {
  if (!token.value || !exception.value) return
  const target = status ?? nextStatus.value
  if (!target && !comment.value.trim()) {
    updateError.value = 'Write a comment or choose a new status.'
    return
  }

  saving.value = true
  updateError.value = ''
  try {
    exception.value = await apiRequest<ShipmentException>(`/api/exceptions/${exception.value.id}/updates`, {
      method: 'POST',
      token: token.value,
      body: {
        status: target || undefined,
        comment: comment.value.trim() || undefined,
        resolution_notes: resolutionNotes.value.trim() || undefined,
      },
    })
    comment.value = ''
    nextStatus.value = ''
    resolutionNotes.value = ''
  } catch (err) {
    updateError.value = err instanceof ApiError ? err.message : 'Could not save the update.'
  } finally {
    saving.value = false
  }
}

function startInvestigating() {
  nextStatus.value = 'investigating'
  void postUpdate('investigating')
}

onMounted(load)
watch(exceptionId, load)
</script>

<template>
  <div class="space-y-5">
    <NuxtLink to="/exceptions" class="inline-flex items-center gap-1.5 text-label text-ink-64 transition hover:text-ink">
      <UiIcon name="chevronLeft" :size="16" />
      Exceptions
    </NuxtLink>

    <div v-if="loading" class="space-y-3" aria-hidden="true">
      <div class="h-24 animate-pulse rounded-lg bg-surface shadow-card" />
      <div class="h-64 animate-pulse rounded-lg bg-surface shadow-card" />
    </div>

    <div v-else-if="error" class="rounded-lg bg-surface p-6 shadow-card">
      <UiEmptyState title="Exception unavailable" :description="error">
        <UiButton variant="tertiary" @click="navigateTo('/exceptions')">Back to exceptions</UiButton>
      </UiEmptyState>
    </div>

    <template v-else-if="exception">
      <header class="space-y-2">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-meta tracking-wide text-ink-40">{{ exception.exception_number }}</p>
            <h1 class="mt-0.5 text-h5 font-medium tracking-tight text-ink">{{ exception.title }}</h1>
            <p class="mt-1 text-label text-ink-64">
              {{ humanise(exception.type) }} · {{ exception.branch_name }} · reported {{ relativeDay(exception.created_at) }}
            </p>
          </div>
          <div class="flex flex-wrap items-center gap-1.5">
            <UiPill :tone="severityTone">{{ humanise(exception.severity) }}</UiPill>
            <StatusPill :status="exception.status" kind="exception" />
          </div>
        </div>
        <NuxtLink
          :to="`/records/${exception.manifest_item_id}`"
          class="inline-flex items-center gap-1.5 text-label font-medium text-blue hover:underline"
        >
          {{ exception.tracking_number }}
          <UiIcon name="arrow" :size="13" />
        </NuxtLink>
      </header>

      <div class="grid gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
        <div class="space-y-4">
          <!-- Description -->
          <section class="rounded-lg bg-surface p-4 shadow-card">
            <h2 class="text-h5 font-medium text-ink">What happened</h2>
            <p v-if="exception.description" class="mt-2 text-p1 text-ink">{{ exception.description }}</p>
            <p v-else class="mt-2 text-label text-ink-64">No description was added.</p>

            <dl class="mt-4 divide-y divide-line">
              <div class="flex items-start justify-between gap-3 py-2">
                <dt class="text-label text-ink-64">Reported by</dt>
                <dd class="text-right text-label font-medium text-ink">{{ exception.reported_by_name || '—' }}</dd>
              </div>
              <div class="flex items-start justify-between gap-3 py-2">
                <dt class="text-label text-ink-64">Assigned to</dt>
                <dd class="text-right text-label font-medium text-ink">{{ exception.assigned_to_name || 'Unassigned' }}</dd>
              </div>
              <div v-if="exception.resolved_at" class="flex items-start justify-between gap-3 py-2">
                <dt class="text-label text-ink-64">Resolved</dt>
                <dd class="text-right text-label font-medium text-ink">
                  {{ shortDateTime(exception.resolved_at) }}
                  <template v-if="exception.resolved_by_name"> · {{ exception.resolved_by_name }}</template>
                </dd>
              </div>
            </dl>

            <div v-if="exception.resolution_notes" class="mt-3 rounded-lg bg-canvas p-3">
              <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Resolution</p>
              <p class="mt-1 text-label text-ink">{{ exception.resolution_notes }}</p>
            </div>
          </section>

          <!-- History -->
          <section class="rounded-lg bg-surface p-4 shadow-card">
            <h2 class="text-h5 font-medium text-ink">History</h2>
            <ol v-if="exception.updates?.length" class="mt-3 space-y-3">
              <li v-for="update in exception.updates" :key="update.id" class="flex gap-3">
                <span class="mt-1.5 size-2 shrink-0 rounded-full bg-blue/60" />
                <div class="min-w-0 flex-1 border-b border-line pb-3 last:border-0 last:pb-0">
                  <p class="text-p1 text-ink">{{ update.comment }}</p>
                  <p class="mt-0.5 text-meta text-ink-40">
                    {{ update.updated_by_name || 'Staff' }} · {{ shortDateTime(update.created_at) }}
                    <template v-if="update.new_status">
                      · {{ humanise(update.previous_status || 'open') }} → {{ humanise(update.new_status) }}
                    </template>
                  </p>
                </div>
              </li>
            </ol>
            <p v-else class="mt-3 text-label text-ink-64">No comments yet.</p>
          </section>
        </div>

        <!-- Update -->
        <section class="h-fit rounded-lg bg-surface p-4 shadow-card">
          <h2 class="text-h5 font-medium text-ink">Add an update</h2>
          <p class="mt-0.5 text-label text-ink-64">
            {{ isOpen ? 'Comment, investigate or resolve this exception.' : 'This exception is closed to new decisions, but you can still comment.' }}
          </p>

          <div v-if="updateError" class="mt-3 rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
            {{ updateError }}
          </div>

          <div class="mt-3 space-y-3">
            <UiTextarea
              id="exception-comment"
              v-model="comment"
              label="Comment"
              :rows="3"
              placeholder="What did you find or do?"
            />
            <UiSelect
              v-if="isOpen"
              id="exception-status"
              v-model="nextStatus"
              label="Move to"
              :options="statusOptions"
            />
            <UiTextarea
              v-if="nextStatus === 'resolved' || nextStatus === 'closed'"
              id="exception-resolution"
              v-model="resolutionNotes"
              label="Resolution notes"
              :rows="2"
              hint="What fixed it, or why it is being closed."
            />

            <div class="flex flex-col gap-2">
              <UiButton block :loading="saving" @click="postUpdate()">Post update</UiButton>
              <UiButton
                v-if="exception.status === 'open'"
                variant="tertiary"
                block
                :loading="saving"
                @click="startInvestigating"
              >
                Start investigating
              </UiButton>
            </div>
          </div>
        </section>
      </div>
    </template>
  </div>
</template>
