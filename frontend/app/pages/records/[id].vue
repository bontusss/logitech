<script setup lang="ts">
import type { Shipment } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Record' })
definePageMeta({ layout: 'default' })

const route = useRoute()
const { token } = useAuth()
const { canReceive, canDispatch, canEdit, canCancel } = usePermissions()

const record = ref<Shipment | null>(null)
const loading = ref(true)
const error = ref('')

const recordId = computed(() => String(route.params.id))

const receiveAllowed = computed(() => Boolean(record.value && canReceive(record.value)))
const dispatchAllowed = computed(() => Boolean(record.value && canDispatch(record.value)))
const editAllowed = computed(() => Boolean(record.value && canEdit(record.value)))
const cancelAllowed = computed(() => Boolean(record.value && canCancel(record.value)))

const paidRatio = computed(() => {
  if (!record.value || record.value.amount <= 0) return 0
  return Math.min(100, Math.round((record.value.amount_paid / record.value.amount) * 100))
})

const steps = computed(() => {
  const current = record.value
  if (!current) return []
  return [
    {
      label: 'Recorded',
      detail: `${current.origin_branch} · ${relativeDay(current.created_at)}`,
      done: true,
      description: `By ${current.recorded_by_name || 'staff'}`,
    },
    {
      label: 'Dispatched',
      detail: current.dispatched_at ? shortDateTime(current.dispatched_at) : 'Not dispatched yet',
      done: Boolean(current.dispatched_at),
      description: current.dispatched_at ? `In transit to ${current.destination_branch}` : 'Waiting at the origin branch',
    },
    {
      label: current.condition && current.condition !== 'good' ? `Received · ${humanise(current.condition)}` : 'Received',
      detail: current.received_at ? shortDateTime(current.received_at) : 'Not received yet',
      done: Boolean(current.received_at),
      description: current.received_at
        ? `By ${current.received_by_name || 'staff'} · ${current.received_quantity ?? current.quantity} of ${current.quantity} items`
        : `Waiting at ${current.destination_branch}`,
    },
  ]
})

async function load() {
  if (!token.value) return
  loading.value = true
  error.value = ''
  try {
    record.value = await apiRequest<Shipment>(`/api/records/${recordId.value}`, { token: token.value })
  } catch (err) {
    error.value =
      err instanceof ApiError && err.status === 404
        ? 'That record does not exist, or it is outside your branch.'
        : err instanceof ApiError
          ? err.message
          : 'Could not load the record.'
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(recordId, load)
</script>

<template>
  <div class="space-y-5">
    <NuxtLink to="/records" class="inline-flex items-center gap-1.5 text-label text-ink-64 transition hover:text-ink">
      <UiIcon name="chevronLeft" :size="16" />
      Records
    </NuxtLink>

    <div v-if="loading" class="space-y-3" aria-hidden="true">
      <div class="h-24 animate-pulse rounded-lg bg-surface shadow-card" />
      <div class="h-64 animate-pulse rounded-lg bg-surface shadow-card" />
    </div>

    <div v-else-if="error" class="rounded-lg bg-surface p-6 shadow-card">
      <UiEmptyState title="Record unavailable" :description="error">
        <UiButton variant="tertiary" @click="navigateTo('/records')">Back to records</UiButton>
      </UiEmptyState>
    </div>

    <template v-else-if="record">
      <header class="space-y-3">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-meta tracking-wide text-ink-40">{{ record.tracking_number }}</p>
            <h1 class="mt-0.5 text-h5 font-medium tracking-tight text-ink">{{ record.description }}</h1>
            <p class="mt-1 flex flex-wrap items-center gap-x-1.5 text-label text-ink-64">
              <span>{{ record.origin_branch }}</span>
              <UiIcon name="arrowRight" :size="14" class="text-ink-40" />
              <span>{{ record.destination_branch }}</span>
              <span class="text-ink-40">·</span>
              <span>Qty {{ record.quantity }}</span>
            </p>
          </div>
          <div class="flex flex-wrap items-center gap-1.5">
            <StatusPill :status="record.status" />
            <StatusPill :status="record.payment_status" kind="payment" />
            <UiPill v-if="record.condition && record.condition !== 'good'" tone="danger">{{ humanise(record.condition) }}</UiPill>
          </div>
        </div>

        <RecordActions
          :record="record"
          :can-receive="receiveAllowed"
          :can-dispatch="dispatchAllowed"
          :can-edit="editAllowed"
          :can-cancel="cancelAllowed"
          @updated="load"
        />
      </header>

      <div class="grid gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
        <div class="space-y-4">
          <!-- Progress -->
          <section class="rounded-lg bg-surface p-4 shadow-card">
            <h2 class="text-h5 font-medium text-ink">Journey</h2>
            <ol class="mt-3 space-y-3">
              <li v-for="(step, index) in steps" :key="step.label" class="flex gap-3">
                <div class="flex flex-col items-center">
                  <span
                    class="flex size-6 shrink-0 items-center justify-center rounded-full text-white"
                    :class="step.done ? 'bg-blue' : 'bg-muted text-ink-40'"
                  >
                    <UiIcon v-if="step.done" name="check" :size="13" />
                    <span v-else class="text-meta">{{ index + 1 }}</span>
                  </span>
                  <span v-if="index < steps.length - 1" class="mt-1 w-px flex-1" :class="step.done ? 'bg-blue/30' : 'bg-line'" />
                </div>
                <div class="min-w-0 pb-1">
                  <div class="flex flex-wrap items-baseline justify-between gap-x-2">
                    <p class="text-p1 font-medium" :class="step.done ? 'text-ink' : 'text-ink-40'">{{ step.label }}</p>
                    <p class="text-meta text-ink-40">{{ step.detail }}</p>
                  </div>
                  <p class="mt-0.5 text-label text-ink-64">{{ step.description }}</p>
                </div>
              </li>
            </ol>
          </section>

          <!-- Customers -->
          <section class="grid gap-4 sm:grid-cols-2">
            <div class="rounded-lg bg-surface p-4 shadow-card">
              <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Sender</p>
              <p class="mt-1 text-p1 font-medium text-ink">{{ record.sender.full_name }}</p>
              <a v-if="record.sender.phone" :href="`tel:${record.sender.phone}`" class="mt-0.5 inline-flex items-center gap-1.5 text-label text-blue">
                <UiIcon name="phone" :size="14" />
                {{ record.sender.phone }}
              </a>
              <p v-else class="mt-0.5 text-label text-ink-40">No phone recorded</p>
            </div>
            <div class="rounded-lg bg-surface p-4 shadow-card">
              <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Receiver</p>
              <p class="mt-1 text-p1 font-medium text-ink">{{ record.receiver.full_name }}</p>
              <a v-if="record.receiver.phone" :href="`tel:${record.receiver.phone}`" class="mt-0.5 inline-flex items-center gap-1.5 text-label text-blue">
                <UiIcon name="phone" :size="14" />
                {{ record.receiver.phone }}
              </a>
              <p v-else class="mt-0.5 text-label text-ink-40">No phone recorded</p>
            </div>
          </section>

          <!-- Payments -->
          <section class="rounded-lg bg-surface p-4 shadow-card">
            <div class="flex items-end justify-between gap-3">
              <h2 class="text-h5 font-medium text-ink">Payments</h2>
              <p class="text-label text-ink-64">{{ record.payments?.length ?? 0 }} recorded</p>
            </div>

            <ul v-if="record.payments?.length" class="mt-3 divide-y divide-line">
              <li v-for="payment in record.payments" :key="payment.id" class="flex items-start justify-between gap-3 py-3">
                <div class="min-w-0">
                  <p class="text-p1 font-medium text-ink">{{ money(payment.amount, payment.currency) }}</p>
                  <p class="mt-0.5 text-meta text-ink-40">
                    {{ humanise(payment.method) }} · {{ shortDateTime(payment.paid_at) }}
                    <template v-if="payment.reference"> · {{ payment.reference }}</template>
                  </p>
                  <p class="mt-0.5 text-meta text-ink-40">By {{ payment.recorded_by_name || 'staff' }}</p>
                </div>
                <p v-if="payment.notes" class="max-w-[45%] text-right text-meta text-ink-64">{{ payment.notes }}</p>
              </li>
            </ul>
            <p v-else class="mt-3 text-label text-ink-64">Nothing paid yet. The full amount is due on delivery.</p>
          </section>

          <!-- Exceptions -->
          <section class="rounded-lg bg-surface p-4 shadow-card">
            <div class="flex items-end justify-between gap-3">
              <h2 class="text-h5 font-medium text-ink">Exceptions</h2>
              <NuxtLink to="/exceptions" class="text-label font-medium text-blue hover:underline">All exceptions</NuxtLink>
            </div>

            <ul v-if="record.exceptions?.length" class="mt-3 divide-y divide-line">
              <li v-for="item in record.exceptions" :key="item.id">
                <NuxtLink :to="`/exceptions/${item.id}`" class="block py-3 transition hover:text-blue">
                  <div class="flex items-start justify-between gap-2">
                    <p class="min-w-0 text-p1 font-medium text-ink">{{ item.title }}</p>
                    <StatusPill :status="item.status" kind="exception" />
                  </div>
                  <p class="mt-0.5 text-meta text-ink-40">
                    {{ item.exception_number }} · {{ humanise(item.type) }} · {{ humanise(item.severity) }} ·
                    {{ relativeDay(item.created_at) }}
                  </p>
                </NuxtLink>
              </li>
            </ul>
            <p v-else class="mt-3 text-label text-ink-64">No exceptions on this record.</p>
          </section>
        </div>

        <div class="space-y-4">
          <!-- Money -->
          <section class="rounded-lg bg-surface p-4 shadow-card">
            <h2 class="text-h5 font-medium text-ink">Money</h2>
            <p class="mt-3 text-h3 leading-none font-medium tracking-tight text-ink">{{ money(record.amount, record.currency) }}</p>
            <p class="mt-1 text-label text-ink-64">Charged by {{ record.origin_branch }}</p>

            <div class="mt-4 h-1.5 overflow-hidden rounded-full bg-muted" role="presentation">
              <div class="h-full rounded-full bg-blue" :style="{ width: `${paidRatio}%` }" />
            </div>
            <p class="mt-1.5 text-meta text-ink-40">{{ paidRatio }}% paid</p>

            <dl class="mt-4 divide-y divide-line">
              <div class="flex items-center justify-between gap-3 py-2">
                <dt class="text-label text-ink-64">Paid</dt>
                <dd class="text-label font-medium text-ink">{{ money(record.amount_paid, record.currency) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3 py-2">
                <dt class="text-label text-ink-64">Balance</dt>
                <dd class="text-label font-medium" :class="record.balance > 0 ? 'text-danger' : 'text-ink'">
                  {{ money(record.balance, record.currency) }}
                </dd>
              </div>
              <div class="flex items-center justify-between gap-3 py-2">
                <dt class="text-label text-ink-64">Payment</dt>
                <dd><StatusPill :status="record.payment_status" kind="payment" /></dd>
              </div>
            </dl>
          </section>

          <!-- Details -->
          <section class="rounded-lg bg-surface p-4 shadow-card">
            <h2 class="text-h5 font-medium text-ink">Details</h2>
            <dl class="mt-2 divide-y divide-line">
              <div class="flex items-start justify-between gap-3 py-2">
                <dt class="text-label text-ink-64">Serial number</dt>
                <dd class="text-right text-label font-medium text-ink">{{ record.serial_number || '—' }}</dd>
              </div>
              <div class="flex items-start justify-between gap-3 py-2">
                <dt class="text-label text-ink-64">Recorded by</dt>
                <dd class="text-right text-label font-medium text-ink">{{ record.recorded_by_name || '—' }}</dd>
              </div>
              <div class="flex items-start justify-between gap-3 py-2">
                <dt class="text-label text-ink-64">Created</dt>
                <dd class="text-right text-label font-medium text-ink">{{ shortDateTime(record.created_at) }}</dd>
              </div>
              <div class="flex items-start justify-between gap-3 py-2">
                <dt class="text-label text-ink-64">Updated</dt>
                <dd class="text-right text-label font-medium text-ink">{{ shortDateTime(record.updated_at) }}</dd>
              </div>
              <div v-if="record.condition" class="flex items-start justify-between gap-3 py-2">
                <dt class="text-label text-ink-64">Condition</dt>
                <dd class="text-right text-label font-medium text-ink">{{ humanise(record.condition) }}</dd>
              </div>
            </dl>
            <div v-if="record.notes" class="mt-3 rounded-lg bg-canvas p-3">
              <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Notes</p>
              <p class="mt-1 text-label text-ink">{{ record.notes }}</p>
            </div>
          </section>
        </div>
      </div>
    </template>
  </div>
</template>
