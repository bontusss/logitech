<script setup lang="ts">
import {
  ArrowLeft,
  ArrowRight,
  Check,
  CircleAlert,
  CircleCheck,
  ClipboardCheck,
  Loader2,
  MailCheck,
  MapPin,
  MessageCircle,
  Palette,
  Send,
  ShieldCheck,
  Building2,
  UserRound,
} from 'lucide-vue-next'
import type { CompanyDetail } from '~/types/auth'
import { isValidNigerianPhone } from '~/utils/nigeria'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Company onboarding' })

const { api } = useApi()
const { user, fetchMe } = useAuth()
const { currentCompany } = useCompany()

const detail = ref<CompanyDetail | null>(null)
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const success = ref('')
const step = ref(0)
const devToken = ref('')
const whatsappCode = ref('')
const sendingWhatsAppCode = ref(false)
const verifyingWhatsAppCode = ref(false)

const steps = [
  { label: 'Business', icon: Building2 },
  { label: 'Contact', icon: UserRound },
  { label: 'Location', icon: MapPin },
  { label: 'Branding', icon: Palette },
  { label: 'CAC', icon: ShieldCheck },
]

const form = reactive({
  name: '',
  cac_business_name: '',
  business_email: '',
  phone: '',
  contact_name: '',
  contact_phone: '',
  country: 'NG',
  state: '',
  city: '',
  address: '',
  primary_color: '#2b4bff',
  secondary_color: '#0a1220',
  logo_url: '',
})

onMounted(async () => {
  if (import.meta.client) {
    devToken.value = localStorage.getItem('bol_dev_verify_token') ?? ''
  }
  await fetchMe()
  await load()
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api<CompanyDetail>('/companies/me')
    if (data.company.status === 'verified') {
      await navigateTo('/dashboard', { replace: true })
      return
    }
    detail.value = data
    Object.assign(form, {
      name: data.company.name,
      cac_business_name: data.company.cac_business_name ?? '',
      business_email: data.company.business_email,
      phone: data.company.phone,
      contact_name: user.value?.full_name ?? '',
      contact_phone: user.value?.phone ?? data.company.phone,
      country: data.company.country,
      state: data.company.state ?? '',
      city: data.company.city ?? '',
      address: data.company.address ?? '',
      primary_color: data.branding.primary_color,
      secondary_color: data.branding.secondary_color,
      logo_url: data.branding.logo_url ?? '',
    })
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not load your company.')
  } finally {
    loading.value = false
  }
}

async function sendWhatsAppCode() {
  if (!isValidNigerianPhone(form.contact_phone)) {
    error.value = 'Enter a valid Nigerian phone number before requesting a WhatsApp code.'
    return
  }
  sendingWhatsAppCode.value = true
  error.value = ''
  success.value = ''
  try {
    const response = await api<{ verified: boolean; message: string }>('/auth/whatsapp/send-code', {
      method: 'POST',
      body: { phone: form.contact_phone },
    })
    await fetchMe()
    success.value = response.message
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not send a WhatsApp verification code.')
  } finally {
    sendingWhatsAppCode.value = false
  }
}

async function verifyWhatsAppCode() {
  verifyingWhatsAppCode.value = true
  error.value = ''
  success.value = ''
  try {
    const response = await api<{ verified: boolean; message: string }>('/auth/whatsapp/verify-code', {
      method: 'POST',
      body: { code: whatsappCode.value },
    })
    await fetchMe()
    form.contact_phone = user.value?.phone ?? form.contact_phone
    whatsappCode.value = ''
    success.value = response.message
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not verify this WhatsApp number.')
  } finally {
    verifyingWhatsAppCode.value = false
  }
}

async function saveCurrentStep(): Promise<boolean> {
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    if (step.value === 0) {
      if (!isValidNigerianPhone(form.phone)) {
        error.value = 'Enter a valid Nigerian company phone number.'
        return false
      }
      await api('/companies/me', {
        method: 'PATCH',
        body: {
          name: form.name,
          phone: form.phone,
        },
      })
    } else if (step.value === 1) {
      if (!isValidNigerianPhone(form.contact_phone)) {
        error.value = 'Enter a valid Nigerian contact phone number.'
        return false
      }
      await api('/users/me', {
        method: 'PATCH',
        body: { full_name: form.contact_name },
      })
      await fetchMe()
    } else if (step.value === 2) {
      if (!form.state || !form.city) {
        error.value = 'Select a state and city or LGA.'
        return false
      }
      await api('/companies/me', {
        method: 'PATCH',
        body: {
          country: form.country,
          state: form.state,
          city: form.city,
          address: form.address,
        },
      })
    } else if (step.value === 4) {
      if (form.cac_business_name.trim().length < 2) {
        error.value = 'Enter the business name exactly as it is registered with the CAC.'
        return false
      }
      await api('/companies/me', {
        method: 'PATCH',
        body: { cac_business_name: form.cac_business_name.trim() },
      })
    } else if (step.value === 3) {
      await api('/companies/me/branding', {
        method: 'PATCH',
        body: {
          primary_color: form.primary_color,
          secondary_color: form.secondary_color,
          logo_url: form.logo_url,
        },
      })
    }
    success.value = 'Saved.'
    return true
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not save this step.')
    return false
  } finally {
    saving.value = false
  }
}

async function next() {
  if (await saveCurrentStep()) {
    step.value = Math.min(step.value + 1, steps.length - 1)
  }
}

function back() {
  error.value = ''
  success.value = ''
  step.value = Math.max(step.value - 1, 0)
}

async function submitForReview() {
  if (step.value === steps.length - 1 && !(await saveCurrentStep())) return

  saving.value = true
  error.value = ''
  success.value = ''
  try {
    const response = await api<{ company: CompanyDetail['company'] }>(
      '/companies/me/submit-verification',
      { method: 'POST' },
    )
    if (detail.value) detail.value.company = response.company
    success.value = 'Your company has been submitted for review.'
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not submit your company for review.')
  } finally {
    saving.value = false
  }
}

const verifyLink = computed(() =>
  devToken.value ? `/verify-email?token=${encodeURIComponent(devToken.value)}` : '/verify-email',
)

const statusLabel = computed(() =>
  (detail.value?.company.status ?? 'pending').replace('_', ' '),
)
const whatsappVerified = computed(() =>
  Boolean(user.value?.whatsapp_verified && user.value.phone === form.contact_phone),
)
</script>

<template>
  <div class="dash-page">
    <section class="dash-hero relative isolate overflow-hidden rounded-3xl border p-5 sm:p-6">
      <div aria-hidden="true" class="pointer-events-none absolute -top-24 right-[-3rem] -z-10 size-56 rounded-full bg-white/40 blur-3xl dark:bg-white/5" />
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="dash-eyebrow inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-bold tracking-[0.18em] uppercase">
            <Building2 :size="13" /> Company onboarding
          </p>
          <h1 class="mt-3 text-[clamp(1.5rem,2.6vw,1.9rem)] leading-tight font-semibold tracking-[-0.035em]">
            Complete {{ currentCompany?.company_name ?? 'your company' }}
          </h1>
          <p class="mt-1.5 max-w-[560px] text-[13px] leading-relaxed text-ink-900/70">
            A complete profile is required before your company can publish rates or receive
            marketplace bookings.
          </p>
          <div class="mt-3 flex flex-wrap items-center gap-2">
            <span class="inline-flex items-center gap-1.5 rounded-full border border-white/80 bg-white/80 px-2.5 py-1 text-[11px] font-semibold text-ink-900/75 capitalize shadow-sm dark:border-white/10 dark:bg-white/10">
              <span class="size-1.5 rounded-full bg-accent-500" />
              {{ statusLabel }}
            </span>
            <span
              class="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] font-semibold shadow-sm"
              :class="whatsappVerified ? 'border-success-200/80 bg-success-50 text-success-700 dark:border-success-500/20 dark:bg-success-500/10 dark:text-success-200' : 'border-accent-200/80 bg-accent-50 text-accent-800 dark:border-accent-400/20 dark:bg-accent-400/10 dark:text-accent-200'"
            >
              <MessageCircle :size="12" />
              WhatsApp {{ whatsappVerified ? 'verified' : 'unverified' }}
            </span>
          </div>
        </div>
        <span class="flex size-12 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-600 to-brand-600 text-white shadow-lg shadow-brand-600/25">
          <ClipboardCheck :size="21" />
        </span>
      </div>
    </section>

    <div v-if="loading" class="dash-panel mt-4 flex items-center gap-2 rounded-2xl border p-6 text-[13px] text-ink-900/60">
      <Loader2 :size="16" class="animate-spin" /> Loading company profile…
    </div>

    <div v-else-if="!detail" class="dash-panel mt-4 rounded-2xl border p-6">
      <span class="flex size-10 items-center justify-center rounded-xl bg-accent-100 text-accent-700 dark:bg-accent-400/15 dark:text-accent-200">
        <CircleAlert :size="19" />
      </span>
      <h2 class="mt-3 text-[15px] font-semibold">This account is not attached to a company</h2>
      <p class="mt-1.5 max-w-md text-[13px] leading-relaxed text-ink-900/65">
        {{ error || 'Marketplace customer accounts do not manage a company workspace.' }}
      </p>
      <div class="mt-4 flex gap-2">
        <AppButton to="/dashboard" variant="secondary" size="sm">Back to dashboard</AppButton>
        <AppButton to="/register" size="sm">Register a company</AppButton>
      </div>
    </div>

    <div v-else class="mt-4 grid gap-4 lg:grid-cols-[minmax(0,1fr)_304px]">
      <div>
        <!-- Progress -->
        <ol class="flex gap-2 overflow-x-auto pb-1">
          <li v-for="(item, index) in steps" :key="item.label" class="min-w-0 flex-1">
            <div
              class="flex items-center gap-2.5 rounded-2xl border px-3 py-2.5 transition"
              :class="
                index === step
                  ? 'border-brand-500/40 bg-gradient-to-br from-brand-50 to-brand-50 shadow-sm dark:border-brand-400/30 dark:from-brand-400/10 dark:to-brand-400/5'
                  : index < step
                    ? 'dash-panel border'
                    : 'border-ink-900/8 bg-white/60 dark:border-white/10 dark:bg-white/5'
              "
            >
              <span
                class="flex size-7 shrink-0 items-center justify-center rounded-xl text-[11px] font-semibold shadow-sm"
                :class="
                  index < step
                    ? 'bg-success-100 text-success-700 dark:bg-success-500/15 dark:text-success-200'
                    : index === step
                      ? 'bg-gradient-to-br from-brand-600 to-brand-600 text-white'
                      : 'bg-ink-900/5 text-ink-900/45 dark:bg-white/10 dark:text-slate-300'
                "
              >
                <Check v-if="index < step" :size="13" />
                <component :is="item.icon" v-else :size="13" />
              </span>
              <span
                class="hidden truncate text-[11.5px] font-semibold sm:block"
                :class="index === step ? 'text-ink-900' : 'text-ink-900/60'"
              >{{ item.label }}</span>
            </div>
          </li>
        </ol>

        <div class="dash-panel mt-3 rounded-2xl border p-5 sm:p-6">
          <div class="flex items-center gap-2.5">
            <span class="flex size-9 items-center justify-center rounded-xl bg-brand-100 text-brand-700 dark:bg-brand-400/15 dark:text-brand-200">
              <component :is="steps[step].icon" :size="16" />
            </span>
            <div>
              <h2 class="text-[15px] font-semibold">{{ steps[step].label }} details</h2>
              <p class="text-[11.5px] font-medium text-ink-900/60">Step {{ step + 1 }} of {{ steps.length }}</p>
            </div>
          </div>

          <!-- Step 1: business -->
          <div v-if="step === 0" class="mt-5 grid gap-3.5 sm:grid-cols-2">
            <label class="block sm:col-span-2">
              <span class="mb-1.5 block text-[11px] font-semibold text-ink-900/65">Company name</span>
              <input v-model="form.name" type="text" class="w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 text-sm transition placeholder:text-ink-900/35 hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-semibold text-ink-900/65">Phone</span>
              <NigeriaPhoneInput v-model="form.phone" required />
            </label>
            <label class="block sm:col-span-2">
              <span class="mb-1.5 block text-[11px] font-semibold text-ink-900/65">Business email</span>
              <input :value="form.business_email" type="email" disabled class="w-full rounded-xl border border-slate-200 bg-slate-50 px-3.5 py-2.5 text-sm text-ink-900/60 dark:border-white/10 dark:bg-white/5" />
              <span class="mt-1.5 block text-[10.5px] text-ink-900/55">Contact support to change the account email.</span>
            </label>
          </div>

          <!-- Step 2: contact -->
          <div v-else-if="step === 1" class="mt-5 grid gap-3.5 sm:grid-cols-2">
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-semibold text-ink-900/65">Contact person</span>
              <input v-model="form.contact_name" type="text" class="w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 text-sm transition placeholder:text-ink-900/35 hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-semibold text-ink-900/65">Contact phone</span>
              <NigeriaPhoneInput v-model="form.contact_phone" required />
            </label>
            <div class="dash-soft sm:col-span-2 rounded-xl border p-3.5">
              <p class="flex items-center gap-1.5 text-[12px] font-semibold">
                <MessageCircle :size="14" :class="whatsappVerified ? 'text-success-700 dark:text-success-200' : 'text-accent-700 dark:text-accent-300'" />
                {{ whatsappVerified ? 'WhatsApp number verified' : 'Verify your WhatsApp number' }}
              </p>
              <p class="mt-1 text-[11.5px] leading-relaxed text-ink-900/60">
                We will send a one-time code to this number. Use international format, including your country code.
              </p>
              <div v-if="!whatsappVerified" class="mt-3 flex flex-wrap gap-2">
                <button
                  type="button"
                  :disabled="sendingWhatsAppCode || !isValidNigerianPhone(form.contact_phone)"
                  class="rounded-xl border border-slate-200 bg-canvas px-3 py-2 text-[11.5px] font-semibold transition hover:border-slate-300 disabled:opacity-50 dark:border-white/10 dark:bg-white/5"
                  @click="sendWhatsAppCode"
                >
                  {{ sendingWhatsAppCode ? 'Sending…' : 'Send WhatsApp code' }}
                </button>
                <form class="flex flex-1 gap-2" @submit.prevent="verifyWhatsAppCode">
                  <input
                    v-model="whatsappCode"
                    inputmode="numeric"
                    autocomplete="one-time-code"
                    maxlength="6"
                    pattern="[0-9]{6}"
                    required
                    aria-label="WhatsApp verification code"
                    placeholder="6-digit code"
                    class="min-w-0 flex-1 rounded-xl border border-slate-200 bg-canvas px-3 py-2 text-[11.5px] transition hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5"
                  />
                  <button
                    type="submit"
                    :disabled="verifyingWhatsAppCode || whatsappCode.length !== 6"
                    class="rounded-xl bg-gradient-to-r from-brand-600 to-brand-600 px-3 py-2 text-[11.5px] font-semibold text-white shadow-sm shadow-brand-600/25 transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-50"
                  >
                    {{ verifyingWhatsAppCode ? 'Verifying…' : 'Verify' }}
                  </button>
                </form>
              </div>
            </div>
            <div class="dash-soft sm:col-span-2 rounded-xl border p-3.5">
              <p class="flex items-center gap-1.5 text-[12px] font-semibold">
                <MailCheck :size="14" :class="user?.email_verified ? 'text-success-700 dark:text-success-200' : 'text-accent-700 dark:text-accent-300'" />
                {{ user?.email_verified ? 'Email verified' : 'Email not verified yet' }}
              </p>
              <p class="mt-1 text-[11.5px] leading-relaxed text-ink-900/60">
                {{ user?.email_verified
                  ? 'Verification is complete; you can submit the company after the remaining steps.'
                  : 'Verify your email before submitting the company for review.' }}
              </p>
              <NuxtLink
                v-if="!user?.email_verified"
                :to="verifyLink"
                class="mt-2 inline-block text-[11.5px] font-semibold text-brand-600 hover:text-brand-700"
              >
                Verify email →
              </NuxtLink>
            </div>
          </div>

          <!-- Step 3: location -->
          <div v-else-if="step === 2" class="mt-5 grid gap-3.5 sm:grid-cols-2">
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-semibold text-ink-900/65">Country</span>
              <input v-model="form.country" type="text" maxlength="2" class="w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 text-sm uppercase transition hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5" />
            </label>
            <NigeriaLocationFields
              class="sm:col-span-2"
              v-model:state="form.state"
              v-model:city="form.city"
              required
            />
            <label class="block sm:col-span-2">
              <span class="mb-1.5 block text-[11px] font-semibold text-ink-900/65">Business address</span>
              <input v-model="form.address" type="text" class="w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 text-sm transition placeholder:text-ink-900/35 hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5" placeholder="12 Port Harcourt Road" />
            </label>
          </div>

          <!-- Step 4: branding -->
          <div v-else-if="step === 3" class="mt-5 grid gap-3.5 sm:grid-cols-2">
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-semibold text-ink-900/65">Primary colour</span>
              <div class="flex items-center gap-2">
                <input v-model="form.primary_color" type="color" class="size-10 cursor-pointer rounded-xl border border-slate-200 bg-white p-1 dark:border-white/10" />
                <input v-model="form.primary_color" type="text" class="w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 font-mono text-[12.5px] uppercase transition hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5" />
              </div>
            </label>
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-semibold text-ink-900/65">Secondary colour</span>
              <div class="flex items-center gap-2">
                <input v-model="form.secondary_color" type="color" class="size-10 cursor-pointer rounded-xl border border-slate-200 bg-white p-1 dark:border-white/10" />
                <input v-model="form.secondary_color" type="text" class="w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 font-mono text-[12.5px] uppercase transition hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5" />
              </div>
            </label>
            <label class="block sm:col-span-2">
              <span class="mb-1.5 block text-[11px] font-semibold text-ink-900/65">Logo URL (optional)</span>
              <input v-model="form.logo_url" type="url" class="w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 text-sm transition placeholder:text-ink-900/35 hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5" placeholder="https://…" />
            </label>

            <div class="sm:col-span-2 rounded-2xl border border-ink-900/10 p-4 shadow-sm" :style="{ backgroundColor: `${form.secondary_color}0d` }">
              <p class="text-[9.5px] font-semibold tracking-[0.16em] uppercase" :style="{ color: form.primary_color }">
                Brand preview
              </p>
              <div class="mt-2 flex items-center gap-2.5">
                <span class="flex size-8 items-center justify-center rounded-lg text-[11px] font-semibold text-white" :style="{ backgroundColor: form.primary_color }">
                  {{ form.name.slice(0, 1).toUpperCase() || 'B' }}
                </span>
                <span class="text-[13px] font-semibold" :style="{ color: form.secondary_color }">
                  {{ form.name || 'Your company' }}
                </span>
              </div>
            </div>
          </div>

          <!-- Step 5: CAC registration -->
          <div v-else class="mt-5 space-y-4">
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-semibold text-ink-900/65">
                Registered CAC business name
              </span>
              <input
                v-model="form.cac_business_name"
                type="text"
                maxlength="200"
                required
                autocomplete="organization"
                class="w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 text-sm transition placeholder:text-ink-900/35 hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5"
                placeholder="e.g. Aba Freight Logistics Limited"
              />
              <span class="mt-1.5 block text-[11px] leading-relaxed text-ink-900/60">
                Enter the business name exactly as it appears on your CAC certificate. Our team
                checks this against the CAC register before approving your company.
              </span>
            </label>

            <div class="dash-soft rounded-xl border p-3.5">
              <p class="flex items-center gap-1.5 text-[12px] font-semibold">
                <ShieldCheck :size="14" class="text-brand-600 dark:text-brand-300" />
                What we review
              </p>
              <p class="mt-1 text-[11.5px] leading-relaxed text-ink-900/60">
                A platform administrator reviews your company profile and CAC business name. Once
                approved, you can publish rates and receive marketplace bookings.
              </p>
            </div>
          </div>

          <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
            <CircleAlert :size="14" class="mt-0.5 shrink-0" />
            {{ error }}
          </p>
          <p v-if="success" role="status" class="mt-4 flex items-start gap-2 rounded-lg border border-success-200 bg-success-50 px-3 py-2.5 text-[12px] text-success-700">
            <CircleCheck :size="14" class="mt-0.5 shrink-0" />
            {{ success }}
          </p>

          <div class="mt-6 flex items-center justify-between gap-3">
            <button
              type="button"
              :disabled="step === 0 || saving"
              class="inline-flex items-center gap-1.5 rounded-xl border border-slate-200 bg-canvas px-3.5 py-2 text-[12.5px] font-semibold transition hover:border-slate-300 disabled:opacity-40 dark:border-white/10 dark:bg-white/5"
              @click="back"
            >
              <ArrowLeft :size="14" /> Back
            </button>

            <button
              v-if="step < steps.length - 1"
              type="button"
              :disabled="saving"
              class="inline-flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-brand-600 to-brand-600 px-4 py-2.5 text-[12.5px] font-semibold text-white shadow-[0_12px_26px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-60"
              @click="next"
            >
              <Loader2 v-if="saving" :size="14" class="animate-spin" />
              Save and continue
              <ArrowRight v-if="!saving" :size="14" />
            </button>

            <button
              v-else
              type="button"
              :disabled="saving || detail.company.status === 'under_review' || detail.company.status === 'verified'"
              class="inline-flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-brand-600 to-brand-600 px-4 py-2.5 text-[12.5px] font-semibold text-white shadow-[0_12px_26px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-60"
              @click="submitForReview"
            >
              <Loader2 v-if="saving" :size="14" class="animate-spin" />
              <Send v-else :size="14" />
              {{ detail.company.status === 'under_review' ? 'Submitted for review' : 'Submit for review' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Sidebar status -->
      <aside class="space-y-3">
        <div class="dash-panel rounded-2xl border p-4">
          <div class="flex items-center gap-2.5">
            <span class="flex size-8 items-center justify-center rounded-xl bg-gradient-to-br from-brand-600 to-brand-600 text-white shadow-sm shadow-brand-600/25">
              <Building2 :size="15" />
            </span>
            <p class="text-[10px] font-bold tracking-[0.16em] text-ink-900/60 uppercase">Company status</p>
          </div>
          <div class="mt-3 flex items-center gap-2">
            <span class="inline-flex items-center gap-1.5 rounded-full border border-accent-200/80 bg-accent-50 px-2.5 py-1 text-[12px] font-semibold capitalize text-accent-800 dark:border-accent-400/20 dark:bg-accent-400/10 dark:text-accent-200">
              <span class="size-1.5 rounded-full bg-accent-500" />
              {{ statusLabel }}
            </span>
          </div>
          <p class="mt-2.5 text-[11.5px] leading-relaxed text-ink-900/65">
            Unverified companies cannot publish public rates or receive unrestricted marketplace
            bookings.
          </p>
          <dl class="mt-4 space-y-2.5 border-t border-ink-900/8 pt-3 text-[11.5px] dark:border-white/10">
            <div class="flex justify-between gap-2">
              <dt class="font-medium text-ink-900/55">Onboarding</dt>
              <dd class="font-semibold capitalize">{{ detail.company.onboarding_status.replace('_', ' ') }}</dd>
            </div>
            <div class="flex justify-between gap-2">
              <dt class="font-medium text-ink-900/55">Owner email</dt>
              <dd class="font-semibold" :class="user?.email_verified ? 'text-success-700 dark:text-success-200' : 'text-accent-700 dark:text-accent-300'">
                {{ user?.email_verified ? 'Verified' : 'Unverified' }}
              </dd>
            </div>
          </dl>
        </div>

        <div v-if="!user?.email_verified" class="rounded-2xl border border-accent-200/80 bg-gradient-to-br from-accent-50 to-accent-50/60 p-4 shadow-sm dark:border-accent-400/20 dark:from-accent-400/10 dark:to-accent-400/5">
          <p class="flex items-center gap-1.5 text-[12px] font-semibold text-accent-800 dark:text-accent-100">
            <MailCheck :size="14" /> Verify your email
          </p>
          <p class="mt-1.5 text-[11.5px] leading-relaxed text-accent-800/80 dark:text-accent-100/70">
            Use the development verification link below to confirm the owner email before
            submitting.
          </p>
          <NuxtLink :to="verifyLink" class="mt-2 inline-block rounded-full bg-white/70 px-2.5 py-1 text-[11.5px] font-semibold text-accent-800 underline underline-offset-2 dark:bg-white/10 dark:text-accent-100">
            Open verification page
          </NuxtLink>
        </div>

        <div class="dash-soft rounded-2xl border p-4">
          <p class="flex items-center gap-1.5 text-[11.5px] font-semibold">
            <CircleCheck :size="14" class="text-success-700 dark:text-success-200" />
            What happens next?
          </p>
          <ol class="mt-2.5 space-y-2 text-[11.5px] leading-relaxed text-ink-900/65">
            <li class="flex gap-2"><span class="flex size-4 shrink-0 items-center justify-center rounded-full bg-brand-100 text-[9px] font-bold text-brand-700 dark:bg-brand-400/15 dark:text-brand-200">1</span> We review your company profile and CAC business name.</li>
            <li class="flex gap-2"><span class="flex size-4 shrink-0 items-center justify-center rounded-full bg-brand-100 text-[9px] font-bold text-brand-700 dark:bg-brand-400/15 dark:text-brand-200">2</span> Once verified, you can publish rates and receive bookings.</li>
            <li class="flex gap-2"><span class="flex size-4 shrink-0 items-center justify-center rounded-full bg-success-100 text-[9px] font-bold text-success-700 dark:bg-success-500/15 dark:text-success-200">3</span> Shipment tools arrive in the next development stage.</li>
          </ol>
        </div>
      </aside>
    </div>
  </div>
</template>
