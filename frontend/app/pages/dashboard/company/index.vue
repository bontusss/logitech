<script setup lang="ts">
import {
  ArrowRight,
  CalendarCheck,
  CheckCircle2,
  CircleAlert,
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
  { label: 'Branches', description: 'Locations, managers and staff assignments', to: '/dashboard/company/branches', icon: MapPin, permission: 'branches.view' },
  { label: 'Staff', description: 'Invite, assign and suspend team members', to: '/dashboard/company/staff', icon: Users, permission: 'staff.view' },
  { label: 'Roles & permissions', description: 'Granular access for every company function', to: '/dashboard/company/roles', icon: ShieldCheck, permission: 'staff.roles.manage' },
  { label: 'Booking page', description: 'A public page where customers can reach you', to: '/dashboard/company/booking', icon: CalendarCheck, permission: 'bookings.manage' },
  { label: 'Branding', description: 'Colours, logo and company subdomain', to: '/dashboard/company/branding', icon: Palette, permission: 'branding.manage' },
  { label: 'Settings', description: 'Timezone, currency and notifications', to: '/dashboard/company/settings', icon: Settings, permission: 'company.settings.manage' },
].filter((link) => can(link.permission)))

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
  <div>
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Company management</p>
        <h1 class="mt-2 text-[clamp(1.6rem,3vw,2.1rem)] leading-tight font-semibold tracking-[-0.035em]">
          Company profile
        </h1>
        <p class="mt-2 max-w-2xl text-[13.5px] leading-relaxed text-ink-900/50">
          Keep your public identity, contact details and operating hours current. These details are
          used across your branches, marketplace page and installable app.
        </p>
      </div>
      <LookupRefreshButton :loading="loading" @refresh="refresh" />
    </div>

    <div v-if="loading && !detail" class="mt-8 flex items-center gap-2 rounded-xl border border-ink-900/10 bg-white p-6 text-[13px] text-ink-900/50">
      <Loader2 :size="16" class="animate-spin" /> Loading company…
    </div>

    <p v-else-if="error" role="alert" class="mt-8 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 p-4 text-[12.5px] text-red-700">
      <CircleAlert :size="16" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <template v-else-if="detail">
      <div class="mt-6 flex flex-wrap items-center gap-2 rounded-xl border border-ink-900/10 bg-white px-4 py-3">
        <span class="rounded-full border px-2.5 py-1 text-[10.5px] font-semibold capitalize" :class="statusTone">
          {{ detail.company.status.replace('_', ' ') }}
        </span>
        <p class="text-[11.5px] text-ink-900/45">
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
          <div class="rounded-xl border border-ink-900/10 bg-white p-5">
            <div class="flex items-center gap-2">
              <ShieldCheck :size="16" :class="detail.company.status === 'verified' ? 'text-emerald-600' : 'text-amber-600'" />
              <h2 class="text-[13px] font-semibold">Verification</h2>
            </div>
            <p class="mt-2 text-[11.5px] leading-relaxed text-ink-900/45">
              {{ detail.company.verified_at
                ? 'Your company profile has been verified.'
                : 'Submit the completed profile for review when your details and contact information are ready.' }}
            </p>
            <p v-if="verificationError" class="mt-2 text-[11px] text-red-600">{{ verificationError }}</p>
            <button
              v-if="can('company.update') && !['under_review', 'verified'].includes(detail.company.status)"
              type="button"
              :disabled="submitting"
              class="mt-3 inline-flex w-full items-center justify-center gap-1.5 rounded-lg bg-brand-500 px-3 py-2 text-[12px] font-semibold text-white transition hover:bg-brand-600 disabled:opacity-60"
              @click="submitVerification"
            >
              <Loader2 v-if="submitting" :size="13" class="animate-spin" />
              <Send v-else :size="13" />
              Submit for review
            </button>
            <div v-else-if="detail.company.status === 'verified'" class="mt-3 flex items-center gap-1.5 text-[11px] font-medium text-emerald-700">
              <CheckCircle2 :size="14" /> Verified
            </div>
            <div v-else class="mt-3 rounded-lg bg-canvas px-2.5 py-2 text-[10.5px] text-ink-900/40">
              Review is in progress. You can keep editing the profile while you wait.
            </div>
          </div>

          <div class="rounded-xl border border-ink-900/10 bg-white p-5">
            <h2 class="text-[13px] font-semibold">Workspace details</h2>
            <dl class="mt-3 space-y-2 text-[11.5px]">
              <div class="flex justify-between gap-2">
                <dt class="text-ink-900/40">Currency</dt>
                <dd class="font-medium">{{ detail.settings.currency }}</dd>
              </div>
              <div class="flex justify-between gap-2">
                <dt class="text-ink-900/40">Timezone</dt>
                <dd class="font-medium">{{ detail.settings.timezone }}</dd>
              </div>
              <div class="flex justify-between gap-2">
                <dt class="text-ink-900/40">Subdomain</dt>
                <dd class="font-medium">{{ detail.subdomain || 'Not set' }}</dd>
              </div>
              <div class="flex justify-between gap-2">
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
          class="group rounded-xl border border-ink-900/10 bg-white p-4 transition hover:border-brand-200 hover:bg-brand-50/30"
        >
          <component :is="link.icon" :size="16" class="text-brand-600" />
          <p class="mt-2.5 text-[12.5px] font-semibold">{{ link.label }}</p>
          <p class="mt-1 text-[10.5px] leading-relaxed text-ink-900/40">{{ link.description }}</p>
          <span class="mt-2 inline-flex items-center gap-1 text-[10.5px] font-medium text-brand-600">
            Open <ArrowRight :size="11" class="transition group-hover:translate-x-0.5" />
          </span>
        </NuxtLink>
      </div>

      <div v-if="error" class="mt-4 text-[11px] text-red-600">{{ error }}</div>
    </template>
  </div>
</template>
