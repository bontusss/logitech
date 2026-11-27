<script setup lang="ts">
import {
  ArrowRight,
  CircleAlert,
  Clock3,
  LifeBuoy,
  Loader2,
  Mail,
  MessageSquareWarning,
  Package,
  Phone,
} from 'lucide-vue-next'
import type { CustomerIssue, CustomerIssueType } from '~/types/customer'

useHead({ title: 'Help & Support · BOL' })

const { api } = useApi()
const { whatsappHref } = useSiteConfig()

const issues = ref<CustomerIssue[]>([])
const loading = ref(true)
const error = ref('')

const openCount = computed(() => issues.value.filter((issue) => issue.status === 'open').length)
const resolvedCount = computed(() => issues.value.filter((issue) => issue.status === 'resolved').length)

const faqs = [
  {
    question: 'My parcel is late. What should I do?',
    answer: 'Open the shipment, check its tracking history, then report it as delayed so the provider sees the issue. You can also call the provider number on the shipment.',
  },
  {
    question: 'Who handles a damaged or missing item?',
    answer: 'The logistics provider that carried the parcel investigates and resolves it. BOL records your report and keeps it attached to the shipment.',
  },
  {
    question: 'Can I cancel a booking?',
    answer: 'Yes, while the parcel has not been picked up. Open the shipment details and choose Cancel shipment; after pickup, contact the provider.',
  },
  {
    question: 'Where do I find my receipt?',
    answer: 'Every booking has a receipt on the Payments & receipts page, with print and download options.',
  },
]

function issueLabel(type: CustomerIssueType) {
  return { delayed: 'Delayed', missing: 'Missing', damaged: 'Damaged', other: 'Other' }[type] || 'Issue'
}

function formatDate(value: string) {
  return new Date(value).toLocaleString('en-NG', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const result = await api<{ issues: CustomerIssue[] }>('/customer/issues')
    issues.value = result.issues
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not load your support reports.')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="dash-page">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Support</p>
        <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Help &amp; support</h1>
        <p class="mt-1.5 max-w-2xl text-[12.5px] leading-relaxed text-ink-900/50">
          Report a problem with a parcel, follow up on an issue, or reach the right contact.
        </p>
      </div>
      <AppButton to="/dashboard/customer/shipments" variant="secondary" size="md">
        Report a new issue <ArrowRight :size="14" />
      </AppButton>
    </div>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ error }}
    </p>

    <div class="mt-5 grid gap-4 lg:grid-cols-[minmax(0,1.5fr)_minmax(260px,1fr)]">
      <section class="rounded-xl border border-ink-900/10 bg-surface">
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-ink-900/8 px-4 py-3.5">
          <div>
            <h2 class="text-[13.5px] font-semibold">Your reports</h2>
            <p class="mt-0.5 text-[10.5px] text-ink-900/40">Issues you raised from a shipment</p>
          </div>
          <div class="flex items-center gap-1.5">
            <span class="rounded-full border border-amber-200 bg-amber-50 px-2 py-0.5 text-[9.5px] font-semibold text-amber-700">
              {{ openCount }} open
            </span>
            <span class="rounded-full border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-[9.5px] font-semibold text-emerald-700">
              {{ resolvedCount }} resolved
            </span>
          </div>
        </div>

        <div v-if="loading" class="flex items-center justify-center gap-2 p-10 text-[12px] text-ink-900/45">
          <Loader2 :size="15" class="animate-spin" /> Loading your reports…
        </div>

        <div v-else-if="!issues.length" class="p-8 text-center">
          <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
            <LifeBuoy :size="20" />
          </span>
          <h3 class="mt-3 text-[13px] font-semibold">No reports yet</h3>
          <p class="mx-auto mt-1 max-w-sm text-[11.5px] leading-relaxed text-ink-900/45">
            If something goes wrong with a delivery, open the shipment and choose Report issue.
          </p>
          <AppButton to="/dashboard/customer/shipments" size="sm" class="mt-4">Go to my shipments</AppButton>
        </div>

        <ul v-else class="divide-y divide-ink-900/8">
          <li v-for="issue in issues" :key="issue.id" class="px-4 py-3.5">
            <div class="flex flex-wrap items-center gap-2">
              <span class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[9.5px] font-semibold" :class="issue.issue_type === 'damaged' || issue.issue_type === 'missing'
                ? 'border-rose-200 bg-rose-50 text-rose-700'
                : 'border-amber-200 bg-amber-50 text-amber-700'">
                <MessageSquareWarning :size="10" /> {{ issueLabel(issue.issue_type) }}
              </span>
              <span class="rounded-full border px-2 py-0.5 text-[9.5px] font-semibold" :class="issue.status === 'open'
                ? 'border-sky-200 bg-sky-50 text-sky-700'
                : 'border-emerald-200 bg-emerald-50 text-emerald-700'">
                {{ issue.status === 'open' ? 'Open' : 'Resolved' }}
              </span>
              <span class="text-[10.5px] font-semibold text-ink-900/65">{{ issue.provider_name }}</span>
              <span v-if="issue.tracking_code" class="font-mono text-[9.5px] text-ink-900/35">{{ issue.tracking_code }}</span>
            </div>
            <p class="mt-1.5 text-[11.5px] leading-relaxed text-ink-900/65">{{ issue.description }}</p>
            <p class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[10px] text-ink-900/40">
              <span class="inline-flex items-center gap-1"><Package :size="10" /> {{ issue.pickup_address }} → {{ issue.dropoff_address }}</span>
              <span class="inline-flex items-center gap-1"><Clock3 :size="10" /> {{ formatDate(issue.created_at) }}</span>
            </p>
          </li>
        </ul>
      </section>

      <aside class="space-y-4">
        <section class="rounded-xl border border-ink-900/10 bg-surface p-4">
          <h2 class="text-[13px] font-semibold">Contact us</h2>
          <p class="mt-1 text-[11px] leading-relaxed text-ink-900/50">
            For anything the provider cannot resolve, reach the BOL support team.
          </p>
          <div class="mt-3 space-y-1.5">
            <a
              href="mailto:hello@basedonlogistics.com"
              class="flex items-center gap-2.5 rounded-xl border border-ink-900/8 px-3 py-2.5 text-[11.5px] font-medium transition hover:border-ink-900/20 hover:bg-canvas/60"
            >
              <span class="flex size-8 items-center justify-center rounded-lg bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200"><Mail :size="14" /></span>
              hello@basedonlogistics.com
            </a>
            <a
              v-if="whatsappHref"
              :href="whatsappHref"
              target="_blank"
              rel="noopener"
              class="flex items-center gap-2.5 rounded-xl border border-ink-900/8 px-3 py-2.5 text-[11.5px] font-medium transition hover:border-ink-900/20 hover:bg-canvas/60"
            >
              <span class="flex size-8 items-center justify-center rounded-lg bg-emerald-50 text-emerald-700 dark:bg-emerald-400/10 dark:text-emerald-200"><Phone :size="14" /></span>
              Chat on WhatsApp
            </a>
          </div>
          <p class="mt-3 text-[10px] leading-relaxed text-ink-900/40">
            Have the tracking code ready. For a delay, damage or missing item, report it from the shipment first so the provider is notified.
          </p>
        </section>

        <section class="rounded-xl border border-ink-900/10 bg-surface p-4">
          <h2 class="text-[13px] font-semibold">Quick answers</h2>
          <div class="mt-2 divide-y divide-ink-900/8">
            <details v-for="(faq, index) in faqs" :key="faq.question" class="group py-1" :open="index === 0">
              <summary class="flex cursor-pointer list-none items-start justify-between gap-2 py-2 text-[11.5px] font-semibold marker:hidden [&::-webkit-details-marker]:hidden">
                {{ faq.question }}
                <span class="text-ink-900/35 transition group-open:rotate-45">+</span>
              </summary>
              <p class="pb-2 text-[10.5px] leading-relaxed text-ink-900/55">{{ faq.answer }}</p>
            </details>
          </div>
        </section>
      </aside>
    </div>
  </div>
</template>
