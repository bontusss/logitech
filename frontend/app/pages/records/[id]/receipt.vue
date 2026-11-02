<script setup lang="ts">
import type { Shipment } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Payment receipt' })
definePageMeta({ layout: 'default' })

const route = useRoute()
const { token } = useAuth()
const { branches, load: loadBranches } = useBranches()

const record = ref<Shipment | null>(null)
const loading = ref(true)
const error = ref('')
const paper = ref<HTMLElement | null>(null)

const downloading = ref(false)
const sharing = ref(false)
const supportsShare = ref(false)
const actionMessage = ref('')

const recordId = computed(() => String(route.params.id))

/** A block of contact lines printed under an office heading. */
interface ReceiptOffice {
  address: string
  phones: string
}

interface ReceiptBranch extends ReceiptOffice {
  name: string
}

/** The parts of the receipt the business can customise. */
interface ReceiptTemplate {
  companyName: string
  services: string[]
  headOffice: ReceiptOffice
  branchOffices: ReceiptBranch[]
  corporateOffice: ReceiptOffice
}

const SERVICE_SLOTS = 6
const TEMPLATE_KEY = 'logitech:receipt-template'
const defaultServices = [
  'Nationwide parcel delivery',
  'Same-day and express courier',
  'Freight forwarding and haulage',
  'Warehousing and storage',
  'Door-to-door delivery',
  'Package tracking and insurance',
]

const template = reactive<ReceiptTemplate>({
  companyName: 'Galopin Logistics',
  services: [...defaultServices],
  headOffice: { address: '', phones: '' },
  branchOffices: [],
  corporateOffice: { address: '', phones: '' },
})
let ready = false

function normaliseTemplate(raw: Partial<ReceiptTemplate>): ReceiptTemplate {
  const services = Array.isArray(raw.services)
    ? raw.services.slice(0, SERVICE_SLOTS).map((service) => String(service ?? ''))
    : []
  while (services.length < SERVICE_SLOTS) services.push('')

  return {
    companyName: typeof raw.companyName === 'string' ? raw.companyName : 'Galopin Logistics',
    services,
    headOffice: {
      address: typeof raw.headOffice?.address === 'string' ? raw.headOffice.address : '',
      phones: typeof raw.headOffice?.phones === 'string' ? raw.headOffice.phones : '',
    },
    branchOffices: Array.isArray(raw.branchOffices)
      ? raw.branchOffices.map((branch) => ({
          name: typeof branch?.name === 'string' ? branch.name : '',
          address: typeof branch?.address === 'string' ? branch.address : '',
          phones: typeof branch?.phones === 'string' ? branch.phones : '',
        }))
      : [],
    corporateOffice: {
      address: typeof raw.corporateOffice?.address === 'string' ? raw.corporateOffice.address : '',
      phones: typeof raw.corporateOffice?.phones === 'string' ? raw.corporateOffice.phones : '',
    },
  }
}

/** The company details the staff member edited last, on this device. */
function storedTemplate(): ReceiptTemplate | null {
  if (!import.meta.client) return null
  try {
    const raw = window.localStorage.getItem(TEMPLATE_KEY)
    if (!raw) return null
    return normaliseTemplate(JSON.parse(raw) as Partial<ReceiptTemplate>)
  } catch {
    return null
  }
}

function persist() {
  if (!import.meta.client || !ready) return
  window.localStorage.setItem(TEMPLATE_KEY, JSON.stringify(template))
}

/** A branch address with its state, e.g. "12 Apapa Road, Lagos". */
function officeAddress(branch: { address?: string; state?: string }): string {
  return [branch.address, branch.state].filter(Boolean).join(', ')
}

/** First-time defaults: the sending branch is the head office and the other
 * active branches are the branch offices. Everything stays editable. */
function applyBranchDefaults() {
  const current = record.value
  if (!current) return
  const origin = branches.value.find((branch) => branch.id === current.origin_branch_id)
  if (origin) {
    template.headOffice = { address: officeAddress(origin), phones: origin.phone ?? '' }
  }
  template.branchOffices = branches.value
    .filter((branch) => branch.id !== current.origin_branch_id && branch.status === 'active')
    .map((branch) => ({ name: branch.name, address: officeAddress(branch), phones: branch.phone ?? '' }))
}

async function load() {
  if (!token.value) return
  loading.value = true
  error.value = ''
  try {
    record.value = await apiRequest<Shipment>(`/api/records/${recordId.value}`, { token: token.value })
    await loadBranches()
    const stored = storedTemplate()
    if (stored) {
      Object.assign(template, stored)
    } else {
      applyBranchDefaults()
    }
    ready = true
    if (!stored) persist()
    invalidatePdf()
  } catch (err) {
    error.value =
      err instanceof ApiError && err.status === 404
        ? 'That record does not exist, or it is outside your branch.'
        : err instanceof ApiError
          ? err.message
          : 'Could not load the receipt.'
  } finally {
    loading.value = false
  }

  // Warm the PDF once the receipt is on screen so Share can use the file
  // immediately, while the click still counts as the user's gesture.
  await nextTick()
  if (record.value) void buildPdf().catch(() => {})
}

watch(
  template,
  () => {
    invalidatePdf()
    persist()
  },
  { deep: true },
)

const payments = computed(() => record.value?.payments ?? [])

/** The payment this receipt is for: the most recent one taken. */
const payment = computed(() => {
  if (!payments.value.length) return null
  return payments.value.reduce((latest, candidate) =>
    new Date(candidate.paid_at).getTime() >= new Date(latest.paid_at).getTime() ? candidate : latest,
  )
})

const currency = computed(() => record.value?.currency ?? 'NGN')
const received = computed(() => payment.value?.amount ?? 0)
const issuedAt = computed(() => payment.value?.paid_at ?? record.value?.created_at)
const receiptNumber = computed(() => {
  const id = payment.value?.id
  if (id) return `RCPT-${id.replaceAll('-', '').slice(0, 8).toUpperCase()}`
  return `RCPT-${record.value?.tracking_number ?? ''}`
})

const services = computed(() => template.services.map((service) => service.trim()).filter(Boolean))

/** Six labelled bullets, editable one slot at a time. */
const serviceSlots = computed(() =>
  Array.from({ length: SERVICE_SLOTS }, (_, index) => ({
    id: `receipt-service-${index}`,
    label: `Bullet ${index + 1}`,
    value: template.services[index] ?? '',
    set: (value: string) => {
      template.services[index] = value
    },
  })),
)

const branchOffices = computed(() =>
  template.branchOffices.filter(
    (branch) => branch.name.trim() || branch.address.trim() || branch.phones.trim(),
  ),
)

const hasHeadOffice = computed(
  () => Boolean(template.headOffice.address.trim() || template.headOffice.phones.trim()),
)
const hasCorporateOffice = computed(
  () => Boolean(template.corporateOffice.address.trim() || template.corporateOffice.phones.trim()),
)

function addBranch() {
  template.branchOffices.push({ name: '', address: '', phones: '' })
}

function removeBranch(index: number) {
  template.branchOffices.splice(index, 1)
}

function print() {
  if (!import.meta.client) return
  window.print()
}

let pdfFile: File | null = null
let pdfPromise: Promise<File> | null = null
let pdfVersion = 0

function invalidatePdf() {
  pdfVersion += 1
  pdfFile = null
  pdfPromise = null
}

/** A PDF of the receipt, rendered once and reused until the details change. */
function buildPdf(): Promise<File> {
  if (pdfFile) return Promise.resolve(pdfFile)
  if (pdfPromise) return pdfPromise
  const version = pdfVersion
  pdfPromise = renderPdf()
    .then((file) => {
      if (version === pdfVersion) pdfFile = file
      return file
    })
    .finally(() => {
      if (version === pdfVersion) pdfPromise = null
    })
  return pdfPromise
}

/** Draws a clean, shadow-free copy of the receipt into an A4 PDF. */
async function renderPdf(): Promise<File> {
  if (!import.meta.client) throw new Error('PDFs can only be created in the browser.')
  const source = paper.value
  if (!source || !record.value) throw new Error('The receipt is not ready yet.')

  const [{ default: html2canvas }, { jsPDF }] = await Promise.all([
    import('html2canvas-pro'),
    import('jspdf'),
  ])
  await document.fonts.ready

  const clone = source.cloneNode(true) as HTMLElement
  clone.classList.add('receipt-light')
  clone.style.position = 'fixed'
  clone.style.top = '0'
  clone.style.left = '-10000px'
  clone.style.width = '794px' // A4 at 96 dpi
  clone.style.maxWidth = 'none'
  clone.style.padding = '32px'
  clone.style.borderRadius = '0'
  clone.style.boxShadow = 'none'
  clone.style.background = '#ffffff'
  document.body.appendChild(clone)

  try {
    const canvas = await html2canvas(clone, {
      scale: 2,
      backgroundColor: '#ffffff',
      logging: false,
    })

    const pdf = new jsPDF({ unit: 'mm', format: 'a4', orientation: 'portrait' })
    const margin = 10
    const pageWidth = pdf.internal.pageSize.getWidth()
    const pageHeight = pdf.internal.pageSize.getHeight()
    const contentWidth = pageWidth - margin * 2
    const sliceHeightPx = Math.max(
      1,
      Math.floor(((pageHeight - margin * 2) * canvas.width) / contentWidth),
    )

    if (canvas.height <= sliceHeightPx) {
      pdf.addImage(
        canvas.toDataURL('image/png'),
        'PNG',
        margin,
        margin,
        contentWidth,
        (canvas.height * contentWidth) / canvas.width,
      )
    } else {
      // Longer receipts continue onto another page, one horizontal slice at a time.
      for (let offset = 0; offset < canvas.height; offset += sliceHeightPx) {
        const height = Math.min(sliceHeightPx, canvas.height - offset)
        const slice = document.createElement('canvas')
        slice.width = canvas.width
        slice.height = height
        const context = slice.getContext('2d')
        if (!context) throw new Error('Could not draw the receipt.')
        context.drawImage(canvas, 0, offset, canvas.width, height, 0, 0, canvas.width, height)
        if (offset > 0) pdf.addPage()
        pdf.addImage(
          slice.toDataURL('image/png'),
          'PNG',
          margin,
          margin,
          contentWidth,
          (height * contentWidth) / canvas.width,
        )
      }
    }

    const blob = pdf.output('blob')
    const name = receiptNumber.value.toLowerCase().replace(/[^a-z0-9-]+/g, '-') || 'receipt'
    return new File([blob], `${name}.pdf`, { type: 'application/pdf' })
  } finally {
    clone.remove()
  }
}

function saveFile(file: File) {
  const url = URL.createObjectURL(file)
  const link = document.createElement('a')
  link.href = url
  link.download = file.name
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

async function downloadPdf() {
  if (downloading.value || !record.value) return
  downloading.value = true
  actionMessage.value = ''
  try {
    saveFile(await buildPdf())
  } catch {
    actionMessage.value = 'Could not create the PDF. You can still print the receipt.'
  } finally {
    downloading.value = false
  }
}

async function shareReceipt() {
  if (sharing.value || !record.value) return
  sharing.value = true
  actionMessage.value = ''
  try {
    const file = await buildPdf()
    const title = `Receipt ${receiptNumber.value}`
    if (navigator.canShare?.({ files: [file] })) {
      await navigator.share({ title, text: `${template.companyName || 'Receipt'} · ${title}`, files: [file] })
      return
    }
    // The browser can share links and text but not files: hand over the PDF instead.
    saveFile(file)
    actionMessage.value = 'Sharing files is not supported on this browser, so the PDF was downloaded.'
  } catch (err) {
    if ((err as Error).name === 'AbortError') return
    actionMessage.value = 'Could not share the receipt. You can download the PDF instead.'
  } finally {
    sharing.value = false
  }
}

onMounted(() => {
  supportsShare.value = typeof navigator.share === 'function'
})
onMounted(load)
watch(recordId, load)
</script>

<template>
  <div class="space-y-5">
    <div class="space-y-3 no-print">
      <NuxtLink
        :to="`/records/${recordId}`"
        class="inline-flex items-center gap-1.5 text-label text-ink-64 transition hover:text-ink"
      >
        <UiIcon name="chevronLeft" :size="16" />
        Record
      </NuxtLink>
      <div class="flex flex-wrap items-end justify-between gap-3">
        <header>
          <h1 class="text-h5 font-medium tracking-tight text-ink">Payment receipt</h1>
          <p class="mt-1 text-label text-ink-64">
            Check the receipt, edit the company details if needed, then print, download or share it
            with the customer.
          </p>
          <p v-if="actionMessage" class="mt-2 text-label text-ink-64" role="status">{{ actionMessage }}</p>
        </header>
        <div class="flex flex-wrap items-center gap-2">
          <UiButton
            v-if="supportsShare"
            variant="tertiary"
            :loading="sharing"
            :disabled="!record"
            @click="shareReceipt"
          >
            <UiIcon name="share" :size="17" />
            Share
          </UiButton>
          <UiButton variant="tertiary" :loading="downloading" :disabled="!record" @click="downloadPdf">
            <UiIcon name="download" :size="17" />
            Download PDF
          </UiButton>
          <UiButton :disabled="!record" @click="print">
            <UiIcon name="receipt" :size="17" />
            Print
          </UiButton>
        </div>
      </div>
    </div>

    <div v-if="loading" class="space-y-3" aria-hidden="true">
      <div class="h-24 animate-pulse rounded-lg bg-surface shadow-card" />
      <div class="h-96 animate-pulse rounded-lg bg-surface shadow-card" />
    </div>

    <div v-else-if="error" class="rounded-lg bg-surface p-6 shadow-card">
      <UiEmptyState title="Receipt unavailable" :description="error">
        <UiButton variant="tertiary" @click="navigateTo('/records')">Back to records</UiButton>
      </UiEmptyState>
    </div>

    <div
      v-else-if="record"
      class="grid items-start gap-4 xl:grid-cols-[minmax(0,1fr)_360px]"
    >
      <div class="space-y-3">
        <p
          v-if="!payments.length"
          class="rounded-lg border border-line bg-surface px-4 py-3 text-label text-ink-64 no-print"
        >
          No payment has been recorded on this record yet, so the receipt prints without an amount
          received. Record a payment first for a complete receipt.
        </p>

        <!-- The printable document. Everything outside .receipt-paper is hidden by print CSS. -->
        <article
          ref="paper"
          class="receipt-paper mx-auto w-full max-w-[760px] rounded-lg bg-surface p-5 text-ink shadow-card sm:p-8"
        >
          <header class="flex flex-wrap items-start justify-between gap-4 border-b-2 border-ink pb-4">
            <div class="min-w-0">
              <h2 class="text-h5 font-semibold tracking-tight text-ink">
                {{ template.companyName || 'Company name' }}
              </h2>
              <p class="mt-1 text-meta font-medium tracking-[0.18em] text-ink-64 uppercase">Payment receipt</p>
            </div>
            <div class="text-right text-meta text-ink-64">
              <p class="text-label font-medium text-ink">{{ receiptNumber }}</p>
              <p class="mt-0.5">{{ shortDate(issuedAt) }}</p>
              <p v-if="payment" class="mt-0.5">{{ humanise(payment.method) }}</p>
            </div>
          </header>

          <section class="mt-4 grid gap-4 sm:grid-cols-2">
            <div>
              <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Received from</p>
              <p class="mt-1 text-p1 font-medium text-ink">{{ record.sender.full_name }}</p>
              <p v-if="record.sender.phone" class="text-label text-ink-64">{{ record.sender.phone }}</p>
            </div>
            <div>
              <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Consignee</p>
              <p class="mt-1 text-p1 font-medium text-ink">{{ record.receiver.full_name }}</p>
              <p v-if="record.receiver.phone" class="text-label text-ink-64">{{ record.receiver.phone }}</p>
            </div>
          </section>

          <section class="receipt-section mt-4">
            <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Shipment</p>
            <table class="mt-1.5 w-full border-collapse text-label">
              <tbody class="divide-y divide-line">
                <tr>
                  <th class="w-[38%] py-1.5 pr-3 text-left font-normal text-ink-64">Waybill number</th>
                  <td class="py-1.5 text-right font-medium text-ink">{{ record.tracking_number }}</td>
                </tr>
                <tr>
                  <th class="py-1.5 pr-3 text-left font-normal text-ink-64">Description</th>
                  <td class="py-1.5 text-right font-medium text-ink">{{ record.description }}</td>
                </tr>
                <tr>
                  <th class="py-1.5 pr-3 text-left font-normal text-ink-64">Quantity</th>
                  <td class="py-1.5 text-right font-medium text-ink">{{ record.quantity }}</td>
                </tr>
                <tr>
                  <th class="py-1.5 pr-3 text-left font-normal text-ink-64">Route</th>
                  <td class="py-1.5 text-right font-medium text-ink">
                    {{ record.origin_branch }} → {{ record.destination_branch }}
                  </td>
                </tr>
              </tbody>
            </table>
          </section>

          <section class="receipt-section mt-4">
            <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Payment</p>
            <table class="mt-1.5 w-full border-collapse text-label">
              <thead>
                <tr class="border-b border-ink/30 text-meta tracking-[0.06em] text-ink-40 uppercase">
                  <th class="py-1.5 text-left font-medium">Details</th>
                  <th class="py-1.5 text-right font-medium">Amount</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-line">
                <tr>
                  <td class="py-1.5 pr-3">
                    Amount received<template v-if="payment"> · {{ humanise(payment.method) }}</template>
                  </td>
                  <td class="py-1.5 text-right font-medium">{{ moneyExact(received, currency) }}</td>
                </tr>
                <tr v-if="payment?.reference">
                  <td class="py-1.5 pr-3">Reference</td>
                  <td class="py-1.5 text-right">{{ payment.reference }}</td>
                </tr>
                <tr v-if="payment">
                  <td class="py-1.5 pr-3">Received by</td>
                  <td class="py-1.5 text-right">{{ payment.recorded_by_name || 'staff' }}</td>
                </tr>
                <tr>
                  <td class="py-1.5 pr-3">Total charged</td>
                  <td class="py-1.5 text-right">{{ moneyExact(record.amount, currency) }}</td>
                </tr>
                <tr>
                  <td class="py-1.5 pr-3">Total paid to date</td>
                  <td class="py-1.5 text-right">{{ moneyExact(record.amount_paid, currency) }}</td>
                </tr>
                <tr>
                  <td class="py-1.5 pr-3 font-medium text-ink">Outstanding balance</td>
                  <td
                    class="py-1.5 text-right font-medium"
                    :class="record.balance > 0 ? 'text-danger' : 'text-ink'"
                  >
                    {{ moneyExact(record.balance, currency) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </section>

          <section v-if="services.length" class="receipt-section mt-4">
            <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Services rendered</p>
            <ul class="mt-1.5 grid gap-x-6 gap-y-1 sm:grid-cols-2">
              <li
                v-for="(service, index) in services"
                :key="`${index}-${service}`"
                class="flex gap-2 text-label text-ink"
              >
                <span aria-hidden="true">•</span>
                <span>{{ service }}</span>
              </li>
            </ul>
          </section>

          <section class="mt-4 grid gap-3 sm:grid-cols-2">
            <div v-if="hasHeadOffice" class="receipt-section">
              <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Head office</p>
              <p v-if="template.headOffice.address" class="mt-1 text-label whitespace-pre-line text-ink">
                {{ template.headOffice.address }}
              </p>
              <p v-if="template.headOffice.phones" class="text-label text-ink-64">
                {{ template.headOffice.phones }}
              </p>
            </div>
            <div v-if="hasCorporateOffice" class="receipt-section">
              <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Corporate office</p>
              <p v-if="template.corporateOffice.address" class="mt-1 text-label whitespace-pre-line text-ink">
                {{ template.corporateOffice.address }}
              </p>
              <p v-if="template.corporateOffice.phones" class="text-label text-ink-64">
                {{ template.corporateOffice.phones }}
              </p>
            </div>
          </section>

          <section v-if="branchOffices.length" class="receipt-section mt-3">
            <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Branch offices</p>
            <ul class="mt-1.5 grid gap-3 sm:grid-cols-2">
              <li v-for="(branch, index) in branchOffices" :key="`branch-${index}`">
                <p v-if="branch.name.trim()" class="text-label font-medium text-ink">{{ branch.name }}</p>
                <p v-if="branch.address.trim()" class="text-label whitespace-pre-line text-ink-64">
                  {{ branch.address }}
                </p>
                <p v-if="branch.phones.trim()" class="text-label text-ink-64">{{ branch.phones }}</p>
              </li>
            </ul>
          </section>

          <footer class="receipt-section mt-5 rounded-lg border border-ink/15 bg-canvas px-3 py-2.5">
            <p class="text-meta font-semibold tracking-[0.08em] text-ink uppercase">Disclaimer</p>
            <p class="mt-1 text-label text-ink">
              Items not picked up after 3 days are at the owner's risk. Payments are not refundable.
            </p>
          </footer>
        </article>

        <p class="text-center text-meta text-ink-40 no-print">Only the receipt is printed.</p>
      </div>

      <!-- Editable company details, kept for the next receipt on this device. -->
      <section class="space-y-4 rounded-lg bg-surface p-4 shadow-card no-print">
        <div>
          <h2 class="text-h5 font-medium text-ink">Receipt details</h2>
          <p class="mt-0.5 text-label text-ink-64">
            These details are saved on this device and reused for every receipt.
          </p>
        </div>

        <UiField id="receipt-company" v-model="template.companyName" label="Company name" />

        <fieldset class="space-y-3 border-t border-line pt-4">
          <legend class="text-label font-medium text-ink">Services rendered (6 bullets)</legend>
          <UiField
            v-for="bullet in serviceSlots"
            :id="bullet.id"
            :key="bullet.id"
            :model-value="bullet.value"
            :label="bullet.label"
            @update:model-value="bullet.set"
          />
        </fieldset>

        <fieldset class="space-y-3 border-t border-line pt-4">
          <legend class="text-label font-medium text-ink">Head office</legend>
          <UiTextarea
            id="receipt-head-address"
            v-model="template.headOffice.address"
            label="Address"
            :rows="2"
          />
          <UiField
            id="receipt-head-phones"
            v-model="template.headOffice.phones"
            label="Phones"
            hint="Separate numbers with a comma."
          />
        </fieldset>

        <fieldset class="space-y-3 border-t border-line pt-4">
          <legend class="text-label font-medium text-ink">Corporate office</legend>
          <UiTextarea
            id="receipt-corporate-address"
            v-model="template.corporateOffice.address"
            label="Address"
            :rows="2"
          />
          <UiField
            id="receipt-corporate-phones"
            v-model="template.corporateOffice.phones"
            label="Phones"
            hint="Separate numbers with a comma."
          />
        </fieldset>

        <fieldset class="space-y-3 border-t border-line pt-4">
          <legend class="text-label font-medium text-ink">Branch offices</legend>
          <div
            v-for="(branch, index) in template.branchOffices"
            :key="`branch-editor-${index}`"
            class="space-y-3 rounded-lg border border-line p-3"
          >
            <div class="flex items-center justify-between gap-3">
              <p class="text-meta font-medium tracking-[0.08em] text-ink-40 uppercase">
                Branch {{ index + 1 }}
              </p>
              <UiButton variant="link" class="text-danger" @click="removeBranch(index)">Remove</UiButton>
            </div>
            <UiField
              :id="`receipt-branch-name-${index}`"
              v-model="branch.name"
              label="Branch name"
            />
            <UiTextarea
              :id="`receipt-branch-address-${index}`"
              v-model="branch.address"
              label="Address"
              :rows="2"
            />
            <UiField
              :id="`receipt-branch-phones-${index}`"
              v-model="branch.phones"
              label="Phones"
            />
          </div>
          <UiButton variant="tertiary" @click="addBranch">
            <UiIcon name="plus" :size="16" />
            Add branch office
          </UiButton>
        </fieldset>
      </section>
    </div>
  </div>
</template>
