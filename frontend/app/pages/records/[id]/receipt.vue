<script setup lang="ts">
import type { Shipment } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Payment receipt' })
definePageMeta({ layout: 'default' })

const route = useRoute()
const { token } = useAuth()

const record = ref<Shipment | null>(null)
const loading = ref(true)
const error = ref('')
const paper = ref<HTMLElement | null>(null)

const downloading = ref(false)
const sharing = ref(false)
const supportsShare = ref(false)
const actionMessage = ref('')

const recordId = computed(() => String(route.params.id))

/**
 * Company details printed on every receipt. They are fixed in the app, so
 * staff never edit them on the printed document.
 */
const template = {
  companyName: 'Galopin Logistics',
  services: [
    'NURTW rep',
    'Commission agent',
    'Auto dealers',
    'Transports',
    'Revenue consultants',
    'General merchandise',
  ],
  headOffice: {
    address: "#1 Milverton Road, Aba\n#10 Asa Road by St. Michael's Road, Aba, Abia State",
    phones: '08165074444, 070385494444, 080338399414',
  },
  branchOffices: [
    {
      name: 'Port Harcourt Branch',
      address: '#36 Aba Road by Market Junction, PH',
      phones: '09137716414',
    },
  ],
  corporateOffice: {
    address: "#40 St Michael's Road, Aba",
    phones: '',
  },
}

async function load() {
  if (!token.value) return
  loading.value = true
  error.value = ''
  try {
    record.value = await apiRequest<Shipment>(`/api/records/${recordId.value}`, { token: token.value })
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

const services = computed(() => template.services)
const branchOffices = computed(() => template.branchOffices)
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
            Check the receipt, then print, download or share it with the customer.
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

    <div v-else-if="record">
      <div class="space-y-3">
        <p
          v-if="!payments.length"
          class="rounded-lg border border-line bg-surface px-4 py-3 text-label text-ink-64 no-print"
        >
          Nothing has been paid at the counter, so this receipt prints with nothing received and the
          full amount as the balance due. The consignee pays on collection.
        </p>

        <!-- The printable document. Everything outside .receipt-paper is hidden by print CSS. -->
        <article
          ref="paper"
          class="receipt-paper mx-auto w-full max-w-[760px] rounded-lg bg-surface p-5 text-ink shadow-card sm:p-8"
        >
          <header class="flex flex-wrap items-start justify-between gap-4 border-b-2 border-ink pb-4">
            <div class="flex min-w-0 items-center gap-3">
              <img
                src="/logo.png"
                alt=""
                width="56"
                height="56"
                class="size-14 shrink-0"
              />
              <div class="min-w-0">
                <h2 class="text-h5 font-semibold tracking-tight text-ink">
                  {{ template.companyName || 'Company name' }}
                </h2>
                <p class="mt-1 text-meta font-medium tracking-[0.18em] text-ink-64 uppercase">Payment receipt</p>
              </div>
            </div>
            <div class="text-right text-meta text-ink-64">
              <p class="text-label font-medium text-ink">{{ receiptNumber }}</p>
              <p class="mt-0.5">{{ shortDate(issuedAt) }}</p>
              <p v-if="payment" class="mt-0.5">{{ humanise(payment.method) }}</p>
            </div>
          </header>

          <!-- The company's own offices sit under the masthead, as a business receipt prints them. -->
          <section class="mt-3 grid gap-3 sm:grid-cols-2">
            <div class="receipt-section">
              <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Head office</p>
              <p v-if="template.headOffice.address" class="mt-1 text-label whitespace-pre-line text-ink">
                {{ template.headOffice.address }}
              </p>
              <p v-if="template.headOffice.phones" class="text-label text-ink-64">
                {{ template.headOffice.phones }}
              </p>
            </div>
            <div class="receipt-section">
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

          <section class="mt-4 grid gap-4 sm:grid-cols-2">
            <div>
              <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Received from</p>
              <p class="mt-1 text-p1 font-medium text-ink">{{ record.sender.full_name || '—' }}</p>
              <p v-if="record.sender.phone" class="text-label text-ink-64">{{ record.sender.phone }}</p>
            </div>
            <div>
              <p class="text-meta tracking-[0.08em] text-ink-40 uppercase">Consignee</p>
              <p class="mt-1 text-p1 font-medium text-ink">{{ record.receiver.full_name || '—' }}</p>
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

          <footer class="receipt-section mt-5 rounded-lg border border-ink/15 bg-canvas px-3 py-2.5">
            <p class="text-meta font-semibold tracking-[0.08em] text-ink uppercase">Disclaimer</p>
            <p class="mt-1 text-label text-ink">
              Items not picked up after 3 days are at the owner's risk. Payments are not refundable.
            </p>
          </footer>
        </article>

        <p class="text-center text-meta text-ink-40 no-print">Only the receipt is printed.</p>
      </div>
    </div>
  </div>
</template>
