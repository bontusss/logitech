<script setup lang="ts">
import { ApiError } from '~/utils/api'

useHead({ title: 'Records' })
definePageMeta({ layout: 'default' })

const { user, token } = useAuth()

interface RecordItem {
  id: string
  waybill_number: string
  description: string
  quantity: number
  sender_name: string
  sender_phone: string
  receiver_name: string
  receiver_phone: string
  destination_branch: string
}

interface RecordManifest {
  id: string
  manifest_number: string
  status: string
  manifest_date: string
  origin_branch: string
  created_at: string
  items: RecordItem[]
}

interface RecordReceipt {
  id: string
  received_at: string
  received_quantity: number
  condition: string
  received_by: string
}

interface RecordPayment {
  id: string
  reference: string
  method: string
  amount: number
  currency: string
  paid_at: string
  collected_by: string
}

interface RecordDetail extends Omit<RecordManifest, 'items'> {
  completed_at: string | null
  items: Array<RecordItem & { receipts: RecordReceipt[]; payments: RecordPayment[] }>
}

const searchText = ref('')
const records = ref<RecordManifest[]>([])
const loading = ref(false)
const error = ref('')
const selectedRecordId = ref<string | null>(null)
const recordDetail = ref<RecordDetail | null>(null)
const detailLoading = ref(false)
const detailError = ref('')
let searchTimeout: ReturnType<typeof setTimeout> | undefined

async function loadRecords() {
  if (!token.value) return
  loading.value = true
  error.value = ''
  try {
    const query = new URLSearchParams({ q: searchText.value.trim() })
    records.value = await apiRequest<RecordManifest[]>(`/api/records?${query}`, { token: token.value })
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not load records. Try again.'
  } finally {
    loading.value = false
  }
}

function scheduleSearch() {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(loadRecords, 250)
}

function formatDate(value: string): string {
  const date = new Date(`${value}T00:00:00`)
  return new Intl.DateTimeFormat('en', { day: 'numeric', month: 'short', year: 'numeric' }).format(date)
}

function formatDateTime(value: string): string {
  return new Intl.DateTimeFormat('en', {
    day: 'numeric', month: 'short', year: 'numeric', hour: 'numeric', minute: '2-digit',
  }).format(new Date(value))
}

function formatAmount(value: number, currency: string): string {
  return `${currency} ${value.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

async function toggleRecord(manifest: RecordManifest) {
  if (selectedRecordId.value === manifest.id) {
    selectedRecordId.value = null
    recordDetail.value = null
    return
  }

  selectedRecordId.value = manifest.id
  recordDetail.value = null
  detailError.value = ''
  detailLoading.value = true
  try {
    const detail = await apiRequest<RecordDetail>(`/api/records/${manifest.id}`, { token: token.value })
    if (selectedRecordId.value === manifest.id) recordDetail.value = detail
  } catch (err) {
    if (selectedRecordId.value === manifest.id) {
      detailError.value = err instanceof ApiError ? err.message : 'Could not load this manifest record.'
    }
  } finally {
    if (selectedRecordId.value === manifest.id) detailLoading.value = false
  }
}

onMounted(loadRecords)
watch(searchText, scheduleSearch)
watch(token, () => { if (token.value) loadRecords() })
onBeforeUnmount(() => {
  if (searchTimeout) clearTimeout(searchTimeout)
})
</script>

<template>
  <div v-if="user">
    <div class="min-h-dvh bg-canvas md:hidden">
      <AppNavbar :user-name="user.full_name" />

      <main class="mx-auto w-full max-w-[480px] px-4 pb-8 pt-5">
        <NuxtLink to="/" class="inline-flex items-center gap-2 text-label text-ink-64 transition hover:text-blue">
          <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="m15 18-6-6 6-6"/>
          </svg>
          Dashboard
        </NuxtLink>

        <header class="mt-5">
          <h1 class="text-h3 text-ink">Records</h1>
          <p class="mt-1 text-label text-ink-64">{{ records.length }} manifest{{ records.length === 1 ? '' : 's' }}</p>
        </header>

        <div class="sticky top-14 z-20 -mx-4 bg-canvas px-4 pb-3 pt-2">
          <UiSearchBar v-model="searchText" />
        </div>

        <p v-if="error" class="py-4 text-label text-danger" role="alert">{{ error }}</p>
        <p v-else-if="loading && records.length === 0" class="py-4 text-label text-ink-64">Loading records…</p>
        <p v-else-if="records.length === 0" class="py-4 text-label text-ink-64">
          {{ searchText ? 'No records match that search.' : 'No records yet.' }}
        </p>

        <div v-else class="divide-y divide-line">
          <section v-for="manifest in records" :key="manifest.id" class="py-4">
            <button
              type="button"
              class="flex w-full items-start justify-between gap-3 text-left focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
              :aria-expanded="selectedRecordId === manifest.id"
              @click="toggleRecord(manifest)"
            >
              <div class="min-w-0">
                <h2 class="truncate text-p1 font-semibold text-ink">Manifest {{ manifest.manifest_number }}</h2>
                <p class="mt-1 text-meta text-ink-64">
                  Created {{ formatDateTime(manifest.created_at) }}
                </p>
                <p class="mt-1 text-meta text-ink-64">
                  Manifest date {{ formatDate(manifest.manifest_date) }}
                  <span v-if="manifest.origin_branch"> · {{ manifest.origin_branch }}</span>
                </p>
              </div>
              <span class="mt-1 shrink-0 text-meta font-medium text-blue">
                {{ selectedRecordId === manifest.id ? 'Close' : 'Details' }}
              </span>
            </button>

            <section v-if="selectedRecordId === manifest.id" class="mt-3 border-t border-line pt-3">
              <p v-if="detailLoading" class="text-label text-ink-64">Loading record details…</p>
              <p v-else-if="detailError" class="text-label text-danger" role="alert">{{ detailError }}</p>
              <template v-else-if="recordDetail">
                <dl class="grid grid-cols-1 gap-2 text-meta text-ink-64">
                  <div>
                    <dt class="inline text-ink-40">Created: </dt>
                    <dd class="inline">{{ formatDateTime(recordDetail.created_at) }}</dd>
                  </div>
                  <div>
                    <dt class="inline text-ink-40">Completed: </dt>
                    <dd class="inline">{{ recordDetail.completed_at ? formatDateTime(recordDetail.completed_at) : 'Time not recorded' }}</dd>
                  </div>
                </dl>

                <div v-if="recordDetail.items.length" class="mt-3 divide-y divide-line">
                  <article v-for="item in recordDetail.items" :key="item.id" class="py-3 first:pt-0">
                    <div class="flex items-start justify-between gap-3">
                      <div class="min-w-0">
                        <p class="break-words text-label font-medium text-ink">{{ item.description }}</p>
                        <p class="mt-1 text-meta text-ink-64">
                          <span v-if="item.waybill_number">Waybill {{ item.waybill_number }}</span>
                          <span v-else>No waybill number</span>
                          · Qty {{ item.quantity }}
                        </p>
                      </div>
                    </div>
                    <dl class="mt-2 grid grid-cols-2 gap-3 text-meta">
                      <div class="min-w-0">
                        <dt class="text-ink-40">Sender</dt>
                        <dd class="break-words text-ink-64">{{ item.sender_name }}<span v-if="item.sender_phone"> · {{ item.sender_phone }}</span></dd>
                      </div>
                      <div class="min-w-0">
                        <dt class="text-ink-40">Receiver</dt>
                        <dd class="break-words text-ink-64">{{ item.receiver_name }}<span v-if="item.receiver_phone"> · {{ item.receiver_phone }}</span></dd>
                      </div>
                    </dl>
                    <p v-if="item.destination_branch" class="mt-2 text-meta text-ink-64">To {{ item.destination_branch }}</p>

                    <section class="mt-3" aria-label="Receipts">
                      <h3 class="text-meta font-semibold text-ink">Receipts</h3>
                      <p v-if="!item.receipts.length" class="mt-1 text-meta text-ink-64">No receipt recorded.</p>
                      <div v-for="receipt in item.receipts" :key="receipt.id" class="mt-2 border-l-2 border-line pl-3 text-meta text-ink-64">
                        <p>{{ formatDateTime(receipt.received_at) }} · {{ receipt.condition.replaceAll('_', ' ') }}</p>
                        <p>Received {{ receipt.received_quantity }} by {{ receipt.received_by }}</p>
                      </div>
                    </section>

                    <section class="mt-3" aria-label="Payments">
                      <h3 class="text-meta font-semibold text-ink">Payments</h3>
                      <p v-if="!item.payments.length" class="mt-1 text-meta text-ink-64">No payment recorded.</p>
                      <div v-for="payment in item.payments" :key="payment.id" class="mt-2 border-l-2 border-line pl-3 text-meta text-ink-64">
                        <p>{{ formatAmount(payment.amount, payment.currency) }} · {{ payment.method.replaceAll('_', ' ') }}</p>
                        <p>Reference: {{ payment.reference || 'Not provided' }}</p>
                        <p>Collected by {{ payment.collected_by }} · {{ formatDateTime(payment.paid_at) }}</p>
                      </div>
                    </section>
                  </article>
                </div>
                <p v-else class="mt-3 text-meta text-ink-64">No items recorded.</p>
              </template>
            </section>
          </section>
        </div>
      </main>
    </div>

    <div class="hidden min-h-dvh items-center justify-center bg-canvas px-6 text-center md:flex">
      <p class="text-label text-ink-64">Galopin is available on mobile screens.</p>
    </div>
  </div>
</template>
