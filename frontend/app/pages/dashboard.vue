<script setup lang="ts">
import {
  ArrowRight,
  Building2,
  Check,
  CircleAlert,
  CircleDot,
  ClipboardList,
  Loader2,
  MailCheck,
  ShieldCheck,
  Sparkles,
} from 'lucide-vue-next'
import type { CompanyDetail, Membership } from '~/types/auth'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Workspace overview' })

const { api } = useApi()
const { user } = useAuth()
const { currentCompany, companies, selectCompany } = useCompany()

const detail = ref<CompanyDetail | null>(null)
const loading = ref(false)
const error = ref('')

onMounted(load)

async function load() {
  if (!currentCompany.value) {
    detail.value = null
    return
  }
  loading.value = true
  error.value = ''
  try {
    detail.value = await api<CompanyDetail>('/companies/me')
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not load your workspace.')
  } finally {
    loading.value = false
  }
}

function chooseCompany(company: Membership) {
  selectCompany(company.company_id)
  window.location.reload()
}

const checklist = computed(() => {
  const company = detail.value?.company
  return [
    {
      label: 'Verify your email',
      body: 'Confirms ownership of the account email.',
      done: Boolean(user.value?.email_verified),
      to: '/verify-email',
    },
    {
      label: 'Add the company profile',
      body: 'Business type, contact and registered location.',
      done: Boolean(company?.address && company?.city && company?.state),
      to: '/onboarding',
    },
    {
      label: 'Configure branding',
      body: 'Company colours and logo for your marketplace profile.',
      done: Boolean(detail.value?.branding),
      to: '/onboarding',
    },
    {
      label: 'Submit for verification',
      body: 'Send the profile for review before publishing rates.',
      done: Boolean(company && ['under_review', 'verified'].includes(company.status)),
      to: '/onboarding',
    },
    {
      label: 'Company verified',
      body: 'Unlocks public rates and marketplace bookings.',
      done: company?.status === 'verified',
      to: '/onboarding',
    },
  ]
})

const completed = computed(() => checklist.value.filter((item) => item.done).length)
const progress = computed(() => Math.round((completed.value / checklist.value.length) * 100))

const statusTone = computed(() => {
  switch (detail.value?.company.status) {
    case 'verified':
      return 'border-emerald-200 bg-emerald-50 text-emerald-700'
    case 'under_review':
      return 'border-brand-200 bg-brand-50 text-brand-700'
    case 'rejected':
    case 'suspended':
      return 'border-red-200 bg-red-50 text-red-700'
    default:
      return 'border-amber-200 bg-amber-50 text-amber-700'
  }
})
</script>

<template>
  <div>
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">
          {{ user?.account_type === 'customer' ? 'Marketplace account' : 'Company workspace' }}
        </p>
        <h1 class="mt-2 text-[clamp(1.6rem,3vw,2.1rem)] leading-tight font-semibold tracking-[-0.035em]">
          Welcome, {{ user?.full_name?.split(' ')[0] || 'there' }}
        </h1>
        <p class="mt-2 max-w-xl text-[13.5px] leading-relaxed text-ink-900/50">
          {{ currentCompany
            ? 'Your account is set up. Finish onboarding to unlock marketplace visibility.'
            : 'Your BOL account is ready. Marketplace booking tools are the next stage.' }}
        </p>
      </div>
      <AppButton v-if="currentCompany" to="/onboarding" size="md">
        Continue onboarding
        <ArrowRight :size="15" />
      </AppButton>
    </div>

    <div v-if="loading" class="mt-8 flex items-center gap-2 rounded-xl border border-ink-900/10 bg-white p-6 text-[13px] text-ink-900/50">
      <Loader2 :size="16" class="animate-spin" /> Loading workspace…
    </div>

    <p v-else-if="error" role="alert" class="mt-8 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 p-4 text-[12.5px] text-red-700">
      <CircleAlert :size="16" class="mt-0.5 shrink-0" />
      {{ error }}
    </p>

    <template v-else-if="detail">
      <div class="mt-7 grid gap-4 sm:grid-cols-3">
        <div class="rounded-xl border border-ink-900/10 bg-white p-4">
          <p class="text-[9.5px] font-semibold tracking-[0.16em] text-ink-900/40 uppercase">Company status</p>
          <div class="mt-2 flex items-center gap-2">
            <span class="rounded-md border px-2 py-1 text-[11px] font-semibold capitalize" :class="statusTone">
              {{ detail.company.status.replace('_', ' ') }}
            </span>
          </div>
          <p class="mt-2 text-[10.5px] leading-relaxed text-ink-900/40">
            {{ detail.company.status === 'verified'
              ? 'Marketplace publishing unlocked.'
              : 'Profile review is required before publishing rates.' }}
          </p>
        </div>

        <div class="rounded-xl border border-ink-900/10 bg-white p-4">
          <p class="text-[9.5px] font-semibold tracking-[0.16em] text-ink-900/40 uppercase">Onboarding</p>
          <p class="mt-2 text-[20px] font-semibold tracking-[-0.03em]">{{ progress }}%</p>
          <div class="mt-2 h-1.5 overflow-hidden rounded-full bg-ink-900/8">
            <div class="h-full rounded-full bg-brand-500 transition-all" :style="{ width: `${progress}%` }" />
          </div>
          <p class="mt-2 text-[10.5px] text-ink-900/40">{{ completed }} of {{ checklist.length }} steps complete</p>
        </div>

        <div class="rounded-xl border border-ink-900/10 bg-white p-4">
          <p class="text-[9.5px] font-semibold tracking-[0.16em] text-ink-900/40 uppercase">Branches</p>
          <p class="mt-2 text-[20px] font-semibold tracking-[-0.03em]">1</p>
          <p class="mt-2 text-[10.5px] leading-relaxed text-ink-900/40">
            Branch management arrives with the operations module.
          </p>
        </div>
      </div>

      <div class="mt-5 grid gap-5 lg:grid-cols-[minmax(0,1.35fr)_minmax(0,1fr)]">
        <div class="rounded-xl border border-ink-900/10 bg-white p-5">
          <div class="flex items-center justify-between gap-3">
            <h2 class="text-[14px] font-semibold">Onboarding checklist</h2>
            <span class="rounded-md bg-ink-900/5 px-2 py-1 text-[10px] font-medium text-ink-900/50">
              {{ progress }}%
            </span>
          </div>

          <ul class="mt-4 divide-y divide-ink-900/8">
            <li v-for="item in checklist" :key="item.label" class="flex items-start gap-3 py-3">
              <span
                class="mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full border"
                :class="item.done ? 'border-emerald-200 bg-emerald-50 text-emerald-600' : 'border-ink-900/15 text-ink-900/25'"
              >
                <Check v-if="item.done" :size="11" :stroke-width="3" />
                <CircleDot v-else :size="9" />
              </span>
              <div class="min-w-0 flex-1">
                <p class="text-[12.5px] font-medium" :class="item.done ? 'text-ink-900/45' : ''">
                  {{ item.label }}
                </p>
                <p class="mt-0.5 text-[11px] leading-relaxed text-ink-900/40">{{ item.body }}</p>
              </div>
              <NuxtLink
                v-if="!item.done"
                :to="item.to"
                class="mt-0.5 shrink-0 text-[10.5px] font-semibold text-brand-600 hover:text-brand-700"
              >
                Start
              </NuxtLink>
            </li>
          </ul>
        </div>

        <div class="space-y-4">
          <div class="rounded-xl border border-ink-900/10 bg-white p-5">
            <h2 class="text-[13px] font-semibold">Company</h2>
            <div class="mt-3 flex items-center gap-2.5">
              <span
                class="flex size-9 items-center justify-center rounded-lg text-[13px] font-semibold text-white"
                :style="{ backgroundColor: detail.branding.primary_color }"
              >
                {{ detail.company.name.slice(0, 1).toUpperCase() }}
              </span>
              <div class="min-w-0">
                <p class="truncate text-[12.5px] font-semibold">{{ detail.company.name }}</p>
                <p class="truncate text-[10.5px] text-ink-900/40">{{ detail.company.business_email }}</p>
              </div>
            </div>
            <dl class="mt-4 space-y-2 border-t border-ink-900/8 pt-3 text-[11.5px]">
              <div class="flex justify-between gap-2">
                <dt class="text-ink-900/40">Business type</dt>
                <dd class="text-right font-medium">{{ detail.company.business_type }}</dd>
              </div>
              <div class="flex justify-between gap-2">
                <dt class="text-ink-900/40">Location</dt>
                <dd class="text-right font-medium">
                  {{ [detail.company.city, detail.company.state].filter(Boolean).join(', ') || 'Not set' }}
                </dd>
              </div>
              <div class="flex justify-between gap-2">
                <dt class="text-ink-900/40">Currency</dt>
                <dd class="font-medium">{{ detail.settings.currency }}</dd>
              </div>
              <div class="flex justify-between gap-2">
                <dt class="text-ink-900/40">Timezone</dt>
                <dd class="font-medium">{{ detail.settings.timezone }}</dd>
              </div>
            </dl>
          </div>

          <div v-if="companies.length > 1" class="rounded-xl border border-ink-900/10 bg-white p-5">
            <h2 class="text-[13px] font-semibold">Switch company</h2>
            <div class="mt-3 space-y-1.5">
              <button
                v-for="company in companies"
                :key="company.company_id"
                type="button"
                class="flex w-full items-center justify-between gap-2 rounded-lg border px-3 py-2 text-left text-[11.5px] transition"
                :class="company.company_id === currentCompany?.company_id ? 'border-brand-200 bg-brand-50' : 'border-ink-900/10 hover:bg-canvas'"
                @click="chooseCompany(company)"
              >
                <span class="truncate">{{ company.company_name }}</span>
                <span class="text-[9.5px] text-ink-900/35 capitalize">{{ company.status.replace('_', ' ') }}</span>
              </button>
            </div>
          </div>

          <div class="rounded-xl border border-brand-100 bg-brand-50/60 p-5">
            <ShieldCheck :size="17" class="text-brand-600" />
            <p class="mt-2 text-[12px] font-semibold">Your data stays in your workspace</p>
            <p class="mt-1 text-[11px] leading-relaxed text-ink-900/50">
              Company access is always checked against your active memberships, not a company id
              supplied by the browser.
            </p>
          </div>
        </div>
      </div>
    </template>

    <template v-else>
      <div class="mt-8 grid gap-4 sm:grid-cols-2">
        <div class="rounded-xl border border-ink-900/10 bg-white p-5">
          <Sparkles :size="18" class="text-brand-600" />
          <h2 class="mt-3 text-[14px] font-semibold">Marketplace account ready</h2>
          <p class="mt-1.5 text-[12.5px] leading-relaxed text-ink-900/50">
            Rate comparison and shipment booking for customers arrive in the marketplace stage.
            Your account already has secure login, email verification and password reset.
          </p>
          <div class="mt-4 flex flex-wrap gap-2">
            <AppButton to="/" size="sm" variant="secondary">Explore the public site</AppButton>
          </div>
        </div>

        <div class="rounded-xl border border-ink-900/10 bg-white p-5">
          <Building2 :size="18" class="text-brand-600" />
          <h2 class="mt-3 text-[14px] font-semibold">Operate a logistics company?</h2>
          <p class="mt-1.5 text-[12.5px] leading-relaxed text-ink-900/50">
            A company must be registered as a separate account with its own owner email. Sign out
            and choose the logistics company option on the registration page.
          </p>
          <div class="mt-4">
            <AppButton to="/register" size="sm">Register a company</AppButton>
          </div>
        </div>
      </div>

      <div class="mt-5 rounded-xl border border-ink-900/10 bg-white p-5">
        <h2 class="text-[13px] font-semibold">Account privacy</h2>
        <p class="mt-1.5 text-[11.5px] leading-relaxed text-ink-900/45">
          Your customer account has no company permissions. Company staff roles are always
          attached to a specific company membership.
        </p>
        <div class="mt-4 flex items-center gap-2 text-[11px] text-ink-900/40">
          <MailCheck :size="13" />
          {{ user?.email_verified ? 'Email verified' : 'Email verification pending' }}
          <span class="text-ink-900/20">·</span>
          <ClipboardList :size="13" />
          {{ user?.account_type }} account
        </div>
      </div>
    </template>
  </div>
</template>
