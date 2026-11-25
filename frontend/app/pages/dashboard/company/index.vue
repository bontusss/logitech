<script setup lang="ts">
import {
  ArrowRight,
  Building2,
  CalendarCheck,
  CheckCircle2,
  CircleAlert,
  ExternalLink,
  Loader2,
  MapPin,
  Palette,
  Send,
  Settings,
  ShieldCheck,
  Users,
} from 'lucide-vue-next'
import type { Company } from '~/types/auth'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Company profile' })

const { api } = useApi()
const { detail, loading, error, load, refresh, can, notify } = useWorkspace()

const submitting = ref(false)
const verificationError = ref('')

onMounted(() => load())

const links = computed(() => [
  { label: 'Branches', description: 'Locations, managers and staff assignments', to: '/dashboard/company/branches', icon: MapPin, permission: 'branches.view', iconClass: 'bg-brand-100 text-brand-700 dark:bg-brand-400/15 dark:text-brand-200' },
  { label: 'Staff', description: 'Invite, assign and suspend team members', to: '/dashboard/company/staff', icon: Users, permission: 'staff.view', iconClass: 'bg-brand-100 text-brand-700 dark:bg-brand-400/15 dark:text-brand-200' },
  { label: 'Roles & permissions', description: 'Granular access for every company function', to: '/dashboard/company/roles', icon: ShieldCheck, permission: 'staff.roles.manage', iconClass: 'bg-success-100 text-success-700 dark:bg-success-500/15 dark:text-success-200' },
  { label: 'Booking page', description: 'A public page where customers can reach you', to: '/dashboard/company/booking', icon: CalendarCheck, permission: 'bookings.manage', iconClass: 'bg-brand-100 text-brand-700 dark:bg-brand-400/15 dark:text-brand-200' },
  { label: 'Branding', description: 'Colours, logo and company subdomain', to: '/dashboard/company/branding', icon: Palette, permission: 'branding.manage', iconClass: 'bg-red-100 text-red-700 dark:bg-red-400/15 dark:text-red-200' },
  { label: 'Settings', description: 'Timezone, currency and notifications', to: '/dashboard/company/settings', icon: Settings, permission: 'company.settings.manage', iconClass: 'bg-accent-100 text-accent-700 dark:bg-accent-400/15 dark:text-accent-200' },
].filter((link) => can(link.permission)))

const statusTone = computed(() => {
  switch (detail.value?.company.status) {
    case 'verified':
      return 'border-success-200 bg-success-50 text-success-700'
    case 'under_review':
      return 'border-brand-200 bg-brand-50 text-brand-700'
    case 'rejected':
    case 'suspended':
      return 'border-red-200 bg-red-50 text-red-700'
    default:
      return 'border-accent-200 bg-accent-50 text-accent-700'
  }
})

function updateCompany(company: Company) {
  if (detail.value) detail.value.company = company
}

async function submitVerification() {
  verificationError.value = ''
  submitting.value = true
  try {
    const data = await api<{ company: Company }>('/companies/me/submit-verification', {
      method: 'POST',
      body: {},
    })
    if (detail.value) detail.value.company = data.company
    notify('Company submitted for review.')
  } catch (err) {
    verificationError.value = apiErrorMessage(err, 'Could not submit the company for review.')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="dash-page">
    <header class="dash-hero relative overflow-hidden rounded-2xl border p-5 sm:p-6">
      <div class="pointer-events-none absolute -top-16 right-10 size-44 rounded-full bg-accent-400/15 blur-3xl" />
      <div class="pointer-events-none absolute -right-8 -bottom-24 size-52 rounded-full bg-brand-400/15 blur-3xl" />
      <div class="relative flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="dash-eyebrow inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-semibold tracking-wide">
            <Building2 :size="12" /> COMPANY MANAGEMENT
          </p>
          <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">
            Company profile
          </h1>
          <p class="mt-1.5 max-w-xl text-[12px] leading-relaxed text-ink-900/65">
            Keep your public identity, contact details and operating hours current across your
            branches, marketplace page and installable app.
          </p>
        </div>
        <LookupRefreshButton :loading="loading" @refresh="refresh" />
      </div>
    </header>

    <div v-if="loading && !detail" class="dash-panel mt-4 flex items-center gap-2 rounded-2xl border p-6 text-[13px] text-ink-900/50">
      <Loader2 :size="16" class="animate-spin" /> Loading company…
    </div>

    <p v-else-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 p-4 text-[12.5px] text-red-700">
      <CircleAlert :size="16" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <template v-else-if="detail">
      <div class="mt-4 grid gap-3 sm:grid-cols-3">
        <article class="dash-stat dash-stat--brand rounded-xl border p-4">
          <div class="flex items-center justify-between gap-2">
            <p class="text-[10px] font-semibold tracking-wide text-ink-900/50 uppercase">Status</p>
            <span class="flex size-6 items-center justify-center rounded-lg bg-white/80 text-brand-700 dark:bg-white/10 dark:text-brand-200">
              <ShieldCheck :size="12" />
            </span>
          </div>
          <p class="mt-2 text-[15px] font-semibold capitalize">{{ detail.company.status.replace('_', ' ') }}</p>
          <p class="mt-0.5 text-[10px] text-ink-900/45">
            {{ detail.company.status === 'verified' ? 'Public and bookable' : 'Verification required' }}
          </p>
        </article>
        <article class="dash-stat dash-stat--success rounded-xl border p-4">
          <div class="flex items-center justify-between gap-2">
            <p class="text-[10px] font-semibold tracking-wide text-ink-900/50 uppercase">Your role</p>
            <span class="flex size-6 items-center justify-center rounded-lg bg-white/80 text-success-700 dark:bg-white/10 dark:text-success-200">
              <Users :size="12" />
            </span>
          </div>
          <p class="mt-2 truncate text-[15px] font-semibold capitalize">
            {{ detail.membership.roles.join(', ').replaceAll('_', ' ') }}
          </p>
          <p class="mt-0.5 text-[10px] text-ink-900/45">Company membership</p>
        </article>
        <article class="dash-stat dash-stat--brand rounded-xl border p-4">
          <div class="flex items-center justify-between gap-2">
            <p class="text-[10px] font-semibold tracking-wide text-ink-900/50 uppercase">Public address</p>
            <span class="flex size-6 items-center justify-center rounded-lg bg-white/80 text-brand-700 dark:bg-white/10 dark:text-brand-200">
              <ExternalLink :size="12" />
            </span>
          </div>
          <p class="mt-2 truncate text-[15px] font-semibold">{{ detail.subdomain || 'Not set' }}</p>
          <p class="mt-0.5 text-[10px] text-ink-900/45">{{ detail.settings.currency }} · {{ detail.settings.timezone }}</p>
        </article>
      </div>

      <div class="mt-4 flex flex-wrap items-center gap-2 rounded-xl border border-ink-900/10 bg-gradient-to-r from-success-50/70 via-white to-brand-50/60 px-4 py-3 dark:border-white/10 dark:from-success-500/5 dark:via-transparent dark:to-brand-400/5">
        <span class="rounded-full border px-2.5 py-1 text-[10.5px] font-semibold capitalize" :class="statusTone">
          {{ detail.company.status.replace('_', ' ') }}
        </span>
        <p class="text-[11.5px] text-ink-900/50">
          {{ detail.company.status === 'verified'
            ? 'Verified companies can publish rates and receive marketplace bookings.'
            : 'Verification is required before public rates and unrestricted bookings.' }}
        </p>
        <NuxtLink
          v-if="detail.company.status !== 'verified'"
          to="/onboarding"
          class="ml-auto inline-flex items-center gap-1 text-[11.5px] font-semibold text-brand-600 hover:text-brand-700"
        >
          Onboarding <ArrowRight :size="13" />
        </NuxtLink>
      </div>

      <div class="mt-5 grid gap-5 lg:grid-cols-[minmax(0,1fr)_280px]">
        <CompanyProfileForm :company="detail.company" :can-edit="can('company.update')" @updated="updateCompany" />

        <aside class="space-y-4">
          <div class="dash-panel rounded-2xl border p-5">
            <div class="flex items-center gap-2">
              <span class="flex size-7 items-center justify-center rounded-lg bg-gradient-to-br from-accent-100 to-accent-50 text-accent-700 dark:from-accent-400/15 dark:to-accent-400/10 dark:text-accent-200">
                <ShieldCheck :size="14" />
              </span>
              <h2 class="text-[13px] font-semibold">Verification</h2>
            </div>
            <p class="mt-2.5 text-[11.5px] leading-relaxed text-ink-900/45">
              {{ detail.company.verified_at
                ? 'Your company profile has been verified.'
                : 'Submit the completed profile for review when your details and contact information are ready.' }}
            </p>
            <p v-if="verificationError" class="mt-2 text-[11px] text-red-600">{{ verificationError }}</p>
            <button
              v-if="can('company.update') && !['under_review', 'verified'].includes(detail.company.status)"
              type="button"
              :disabled="submitting"
              class="mt-3 inline-flex w-full items-center justify-center gap-1.5 rounded-xl bg-gradient-to-r from-brand-600 to-brand-600 px-3 py-2.5 text-[12px] font-semibold text-white shadow-sm shadow-brand-600/25 transition hover:-translate-y-0.5 disabled:opacity-60"
              @click="submitVerification"
            >
              <Loader2 v-if="submitting" :size="13" class="animate-spin" />
              <Send v-else :size="13" />
              Submit for review
            </button>
            <div v-else-if="detail.company.status === 'verified'" class="mt-3 flex items-center gap-1.5 rounded-lg bg-success-50 px-2.5 py-2 text-[11px] font-medium text-success-700 dark:bg-success-500/10 dark:text-success-200">
              <CheckCircle2 :size="14" /> Verified
            </div>
            <div v-else class="mt-3 rounded-lg bg-canvas px-2.5 py-2 text-[10.5px] text-ink-900/40">
              Review is in progress. You can keep editing the profile while you wait.
            </div>
          </div>

          <div class="dash-panel rounded-2xl border p-5">
            <h2 class="text-[13px] font-semibold">Workspace details</h2>
            <dl class="mt-3 divide-y divide-ink-900/8 text-[11.5px]">
              <div class="flex justify-between gap-2 py-2">
                <dt class="text-ink-900/40">Currency</dt>
                <dd class="font-medium">{{ detail.settings.currency }}</dd>
              </div>
              <div class="flex justify-between gap-2 py-2">
                <dt class="text-ink-900/40">Timezone</dt>
                <dd class="font-medium">{{ detail.settings.timezone }}</dd>
              </div>
              <div class="flex justify-between gap-2 py-2">
                <dt class="text-ink-900/40">Subdomain</dt>
                <dd class="font-medium">{{ detail.subdomain || 'Not set' }}</dd>
              </div>
              <div class="flex justify-between gap-2 py-2">
                <dt class="text-ink-900/40">Your role</dt>
                <dd class="font-medium capitalize">{{ detail.membership.roles.join(', ').replaceAll('_', ' ') }}</dd>
              </div>
            </dl>
          </div>
        </aside>
      </div>

      <div v-if="links.length" class="mt-5 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <NuxtLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="dash-link-tile group rounded-2xl border p-4"
        >
          <span class="flex size-9 items-center justify-center rounded-xl" :class="link.iconClass">
            <component :is="link.icon" :size="16" />
          </span>
          <p class="mt-3 text-[12.5px] font-semibold">{{ link.label }}</p>
          <p class="mt-1 text-[10.5px] leading-relaxed text-ink-900/40">{{ link.description }}</p>
          <span class="mt-2 inline-flex items-center gap-1 text-[10.5px] font-semibold text-brand-600">
            Open <ArrowRight :size="11" class="transition group-hover:translate-x-0.5" />
          </span>
        </NuxtLink>
      </div>

      <div v-if="error" class="mt-4 text-[11px] text-red-600">{{ error }}</div>
    </template>
  </div>
</template>
