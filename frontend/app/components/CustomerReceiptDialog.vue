<script setup lang="ts">
import { Download, FileText, Printer, X } from 'lucide-vue-next'
import type { CustomerBooking } from '~/types/customer'

const props = defineProps<{
  open: boolean
  booking: CustomerBooking | null
}>()
const emit = defineEmits<{ close: [] }>()

const amount = computed(() =>
  props.booking ? formatBookingMoney(props.booking.delivery_price_minor, props.booking.currency) : '',
)

function formatDate(value: string) {
  return new Date(value).toLocaleString('en-NG', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function escapeHtml(value: string) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;')
}

function paymentLabel(booking: CustomerBooking) {
  if (!booking.is_paid) return 'Payment pending'
  if (booking.payment_method === 'bank_transfer') return 'Paid by bank transfer'
  if (booking.payment_method === 'pos') return 'Paid by card / POS'
  if (booking.payment_method === 'cash') return 'Paid in cash'
  return 'Paid'
}

function receiptHtml(booking: CustomerBooking) {
  const rows = [
    ['Booking reference', booking.tracking_code || booking.id],
    ['Provider', booking.provider_name],
    ['Booked', formatDate(booking.created_at)],
    ['Pickup', booking.pickup_address],
    ['Delivery', booking.dropoff_address],
    ['Sender', `${booking.sender_name} · ${booking.sender_phone}`],
    ['Receiver', `${booking.receiver_name} · ${booking.receiver_phone}`],
    ['Package', `${booking.item_description} · ${booking.weight_kg} kg`],
    ['Payment', paymentLabel(booking)],
  ]
  return `<!doctype html>
<html><head><meta charset="utf-8"><title>Receipt ${escapeHtml(booking.tracking_code || booking.id)}</title>
<style>
  body{font-family:Inter,ui-sans-serif,system-ui,sans-serif;color:#0f172a;margin:0;padding:40px;background:#f8fafc}
  .sheet{max-width:680px;margin:0 auto;background:#fff;border:1px solid #e2e8f0;border-radius:16px;padding:32px}
  .brand{font-size:13px;letter-spacing:.16em;text-transform:uppercase;color:#0e7490;font-weight:700}
  h1{font-size:24px;margin:8px 0 4px;letter-spacing:-.03em}
  .muted{color:#64748b;font-size:12px}
  table{width:100%;border-collapse:collapse;margin-top:24px}
  td{padding:10px 0;border-bottom:1px solid #e2e8f0;font-size:12px;vertical-align:top}
  td:first-child{color:#64748b;width:160px}
  .total{display:flex;justify-content:space-between;align-items:center;margin-top:24px;padding-top:16px;border-top:2px solid #0f172a}
  .total strong{font-size:22px}
  @media print{body{background:#fff;padding:0}.sheet{border:0}}
</style></head>
<body><div class="sheet">
  <p class="brand">BOL · Based on Logistics</p>
  <h1>Booking receipt</h1>
  <p class="muted">${escapeHtml(booking.provider_name)} · ${escapeHtml(paymentLabel(booking))}</p>
  <table>${rows.map(([label, value]) => `<tr><td>${escapeHtml(label)}</td><td>${escapeHtml(value)}</td></tr>`).join('')}</table>
  <div class="total"><span>Delivery fee</span><strong>${escapeHtml(amount.value)}</strong></div>
  <p class="muted" style="margin-top:24px">This receipt was generated from your BOL customer dashboard. Final provider pricing may be confirmed on dispatch.</p>
</div></body></html>`
}

function printReceipt() {
  if (!props.booking || !import.meta.client) return
  const printWindow = window.open('', '_blank', 'width=760,height=900')
  if (!printWindow) return
  printWindow.document.write(receiptHtml(props.booking))
  printWindow.document.close()
  printWindow.focus()
  printWindow.print()
}

function downloadReceipt() {
  if (!props.booking || !import.meta.client) return
  const blob = new Blob([receiptHtml(props.booking)], { type: 'text/html;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `receipt-${props.booking.tracking_code || props.booking.id}.html`
  link.click()
  URL.revokeObjectURL(url)
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open && booking"
      class="fixed inset-0 z-[80] flex items-center justify-center bg-ink-950/45 px-4 py-6 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      aria-label="Booking receipt"
      @click.self="emit('close')"
    >
      <div class="w-full max-w-lg overflow-hidden rounded-2xl border border-ink-900/10 bg-surface shadow-2xl">
        <div class="flex items-center justify-between gap-3 border-b border-ink-900/8 px-5 py-4">
          <div class="flex items-center gap-2.5">
            <span class="flex size-8 items-center justify-center rounded-lg bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
              <FileText :size="15" />
            </span>
            <div>
              <p class="text-[13px] font-semibold">Booking receipt</p>
              <p class="text-[10.5px] text-ink-900/40">{{ booking.provider_name }}</p>
            </div>
          </div>
          <button type="button" class="text-ink-900/35 hover:text-ink-900" aria-label="Close receipt" @click="emit('close')">
            <X :size="17" />
          </button>
        </div>

        <div class="p-5">
          <div class="rounded-xl border border-ink-900/8 bg-canvas/60 p-4">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">Amount</p>
                <p class="mt-1 text-[24px] font-semibold tracking-[-0.04em] tabular-nums">{{ amount }}</p>
              </div>
              <span
                class="inline-flex items-center rounded-full border px-2.5 py-1 text-[10px] font-semibold"
                :class="booking.is_paid
                  ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                  : 'border-amber-200 bg-amber-50 text-amber-700'"
              >
                {{ paymentLabel(booking) }}
              </span>
            </div>
            <dl class="mt-3 grid gap-2 text-[11px] sm:grid-cols-2">
              <div>
                <dt class="text-ink-900/40">Reference</dt>
                <dd class="mt-0.5 font-mono font-medium">{{ booking.tracking_code || booking.id }}</dd>
              </div>
              <div>
                <dt class="text-ink-900/40">Booked</dt>
                <dd class="mt-0.5 font-medium">{{ formatDate(booking.created_at) }}</dd>
              </div>
              <div class="sm:col-span-2">
                <dt class="text-ink-900/40">Route</dt>
                <dd class="mt-0.5 font-medium">{{ booking.pickup_address }} → {{ booking.dropoff_address }}</dd>
              </div>
              <div>
                <dt class="text-ink-900/40">Sender</dt>
                <dd class="mt-0.5 font-medium">{{ booking.sender_name }} · {{ booking.sender_phone }}</dd>
              </div>
              <div>
                <dt class="text-ink-900/40">Receiver</dt>
                <dd class="mt-0.5 font-medium">{{ booking.receiver_name }} · {{ booking.receiver_phone }}</dd>
              </div>
            </dl>
          </div>

          <div class="mt-4 flex flex-wrap justify-end gap-2">
            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-lg border border-ink-900/15 bg-surface px-3 py-2 text-[11.5px] font-semibold transition hover:bg-canvas"
              @click="emit('close')"
            >
              Close
            </button>
            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-lg border border-ink-900/15 bg-surface px-3 py-2 text-[11.5px] font-semibold transition hover:bg-canvas"
              @click="printReceipt"
            >
              <Printer :size="13" /> Print / PDF
            </button>
            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-lg bg-accent-500 px-3 py-2 text-[11.5px] font-semibold text-white transition hover:bg-accent-400"
              @click="downloadReceipt"
            >
              <Download :size="13" /> Download
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
