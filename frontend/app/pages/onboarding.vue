<script setup lang="ts">
import {
  ArrowLeft,
  ArrowRight,
  Check,
  CircleAlert,
  CircleCheck,
  FileText,
  Loader2,
  MailCheck,
  MapPin,
  Palette,
  Send,
  Building2,
  UserRound,
} from 'lucide-vue-next'
import type { CompanyDetail } from '~/types/auth'

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

const steps = [
  { label: 'Business', icon: Building2 },
  { label: 'Contact', icon: UserRound },
  { label: 'Location', icon: MapPin },
  { label: 'Branding', icon: Palette },
  { label: 'Documents', icon: FileText },
]

const form = reactive({
  name: '',
  business_type: '',
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
  if (!user.value) await fetchMe()
  await load()
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api<CompanyDetail>('/companies/me')
    detail.value = data
    Object.assign(form, {
      name: data.company.name,
      business_type: data.company.business_type,
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

async function saveCurrentStep(): Promise<boolean> {
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    if (step.value === 0) {
      await api('/companies/me', {
        method: 'PATCH',
        body: {
          name: form.name,
          business_type: form.business_type,
          phone: form.phone,
        },
      })
    } else if (step.value === 1) {
      await api('/users/me', {
        method: 'PATCH',
        body: { full_name: form.contact_name, phone: form.contact_phone },
      })
      await fetchMe()
    } else if (step.value === 2) {
      await api('/companies/me', {
        method: 'PATCH',
        body: {
          country: form.country,
          state: form.state,
          city: form.city,
          address: form.address,
        },
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
</script>

<template>
  <div>
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">
          Company onboarding
        </p>
        <h1 class="mt-2 text-[clamp(1.6rem,3vw,2.1rem)] leading-tight font-semibold tracking-[-0.035em]">
          Complete {{ currentCompany?.company_name ?? 'your company' }}
        </h1>
        <p class="mt-2 max-w-xl text-[13.5px] leading-relaxed text-ink-900/50">
          A complete profile is required before your company can publish rates or receive
          marketplace bookings.
        </p>
      </div>
      <span class="rounded-md border border-ink-900/10 bg-white px-2.5 py-1.5 text-[10.5px] font-semibold tracking-[0.08em] capitalize">
        {{ statusLabel }}
      </span>
    </div>

    <div v-if="loading" class="mt-8 flex items-center gap-2 rounded-xl border border-ink-900/10 bg-white p-6 text-[13px] text-ink-900/50">
      <Loader2 :size="16" class="animate-spin" /> Loading company profile…
    </div>

    <div v-else-if="!detail" class="mt-8 rounded-xl border border-ink-900/10 bg-white p-6">
      <CircleAlert :size="20" class="text-amber-600" />
      <h2 class="mt-3 text-[15px] font-semibold">This account is not attached to a company</h2>
      <p class="mt-1.5 max-w-md text-[13px] leading-relaxed text-ink-900/55">
        {{ error || 'Marketplace customer accounts do not manage a company workspace.' }}
      </p>
      <div class="mt-4 flex gap-2">
        <AppButton to="/dashboard" variant="secondary" size="sm">Back to dashboard</AppButton>
        <AppButton to="/register" size="sm">Register a company</AppButton>
      </div>
    </div>

    <div v-else class="mt-7 grid gap-5 lg:grid-cols-[minmax(0,1fr)_300px]">
      <div>
        <!-- Progress -->
        <ol class="flex gap-1 overflow-x-auto pb-1">
          <li v-for="(item, index) in steps" :key="item.label" class="min-w-0 flex-1">
            <div
              class="flex items-center gap-2 rounded-lg border px-2.5 py-2"
              :class="
                index === step
                  ? 'border-brand-500/40 bg-brand-50'
                  : index < step
                    ? 'border-ink-900/10 bg-white'
                    : 'border-ink-900/8 bg-transparent'
              "
            >
              <span
                class="flex size-6 shrink-0 items-center justify-center rounded-md text-[10px] font-semibold"
                :class="
                  index < step
                    ? 'bg-emerald-50 text-emerald-600'
                    : index === step
                      ? 'bg-brand-500 text-white'
                      : 'bg-ink-900/5 text-ink-900/40'
                "
              >
                <Check v-if="index < step" :size="12" />
                <component :is="item.icon" v-else :size="12" />
              </span>
              <span class="hidden truncate text-[11px] font-medium sm:block">{{ item.label }}</span>
            </div>
          </li>
        </ol>

        <div class="mt-4 rounded-xl border border-ink-900/10 bg-white p-5 sm:p-6">
          <h2 class="text-[15px] font-semibold">{{ steps[step].label }} details</h2>
          <p class="mt-1 text-[12px] text-ink-900/45">
            Step {{ step + 1 }} of {{ steps.length }}
          </p>

          <!-- Step 1: business -->
          <div v-if="step === 0" class="mt-5 grid gap-3.5 sm:grid-cols-2">
            <label class="block sm:col-span-2">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Company name</span>
              <input v-model="form.name" type="text" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Business type</span>
              <input v-model="form.business_type" type="text" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Phone</span>
              <input v-model="form.phone" type="tel" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" />
            </label>
            <label class="block sm:col-span-2">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Business email</span>
              <input :value="form.business_email" type="email" disabled class="w-full rounded-lg border border-slate-200 bg-slate-50 px-3 py-2.5 text-sm text-ink-900/50" />
              <span class="mt-1 block text-[10.5px] text-ink-900/35">Contact support to change the account email.</span>
            </label>
          </div>

          <!-- Step 2: contact -->
          <div v-else-if="step === 1" class="mt-5 grid gap-3.5 sm:grid-cols-2">
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Contact person</span>
              <input v-model="form.contact_name" type="text" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Contact phone</span>
              <input v-model="form.contact_phone" type="tel" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" />
            </label>
            <div class="sm:col-span-2 rounded-lg border border-ink-900/8 bg-canvas p-3">
              <p class="flex items-center gap-1.5 text-[12px] font-medium">
                <MailCheck :size="14" :class="user?.email_verified ? 'text-emerald-600' : 'text-amber-600'" />
                {{ user?.email_verified ? 'Email verified' : 'Email not verified yet' }}
              </p>
              <p class="mt-1 text-[11.5px] leading-relaxed text-ink-900/45">
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
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Country</span>
              <input v-model="form.country" type="text" maxlength="2" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm uppercase focus:border-brand-500" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">State</span>
              <input v-model="form.state" type="text" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" placeholder="Abia" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">City</span>
              <input v-model="form.city" type="text" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" placeholder="Aba" />
            </label>
            <label class="block sm:col-span-2">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Business address</span>
              <input v-model="form.address" type="text" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" placeholder="12 Port Harcourt Road" />
            </label>
          </div>

          <!-- Step 4: branding -->
          <div v-else-if="step === 3" class="mt-5 grid gap-3.5 sm:grid-cols-2">
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Primary colour</span>
              <div class="flex items-center gap-2">
                <input v-model="form.primary_color" type="color" class="size-10 cursor-pointer rounded border border-slate-300 bg-white p-1" />
                <input v-model="form.primary_color" type="text" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-mono text-[12.5px] uppercase focus:border-brand-500" />
              </div>
            </label>
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Secondary colour</span>
              <div class="flex items-center gap-2">
                <input v-model="form.secondary_color" type="color" class="size-10 cursor-pointer rounded border border-slate-300 bg-white p-1" />
                <input v-model="form.secondary_color" type="text" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-mono text-[12.5px] uppercase focus:border-brand-500" />
              </div>
            </label>
            <label class="block sm:col-span-2">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Logo URL (optional)</span>
              <input v-model="form.logo_url" type="url" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" placeholder="https://…" />
            </label>

            <div class="sm:col-span-2 rounded-xl border border-ink-900/8 p-4" :style="{ backgroundColor: `${form.secondary_color}0d` }">
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

          <!-- Step 5: documents -->
          <div v-else class="mt-5 space-y-4">
            <div class="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2.5 text-[12px] leading-relaxed text-amber-800">
              Document upload and review are not enabled yet. Prepare the interface now; the
              verification workflow will be connected in a later task.
            </div>

            <div
              v-for="doc in ['CAC / business registration', 'Proof of business address', 'Owner identification']"
              :key="doc"
              class="flex items-center justify-between gap-3 rounded-lg border border-dashed border-ink-900/15 px-3 py-3"
            >
              <div>
                <p class="text-[12.5px] font-medium">{{ doc }}</p>
                <p class="mt-0.5 text-[10.5px] text-ink-900/40">PDF or image · upload coming next</p>
              </div>
              <button type="button" disabled class="rounded-md border border-ink-900/10 px-2.5 py-1.5 text-[10.5px] text-ink-900/35">
                Upload
              </button>
            </div>

            <div class="rounded-lg border border-ink-900/8 bg-canvas p-3">
              <p class="text-[12px] font-medium">Ready to submit?</p>
              <p class="mt-1 text-[11.5px] leading-relaxed text-ink-900/45">
                We will review the company profile. You can update details after submitting.
              </p>
            </div>
          </div>

          <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
            <CircleAlert :size="14" class="mt-0.5 shrink-0" />
            {{ error }}
          </p>
          <p v-if="success" role="status" class="mt-4 flex items-start gap-2 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2.5 text-[12px] text-emerald-700">
            <CircleCheck :size="14" class="mt-0.5 shrink-0" />
            {{ success }}
          </p>

          <div class="mt-6 flex items-center justify-between gap-3">
            <button
              type="button"
              :disabled="step === 0 || saving"
              class="inline-flex items-center gap-1.5 rounded-lg border border-slate-300 px-3.5 py-2 text-[12.5px] font-medium transition hover:bg-slate-50 disabled:opacity-40"
              @click="back"
            >
              <ArrowLeft :size="14" /> Back
            </button>

            <button
              v-if="step < steps.length - 1"
              type="button"
              :disabled="saving"
              class="inline-flex items-center gap-1.5 rounded-lg bg-brand-500 px-4 py-2.5 text-[12.5px] font-semibold text-white transition hover:bg-brand-600 disabled:opacity-60"
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
              class="inline-flex items-center gap-1.5 rounded-lg bg-brand-500 px-4 py-2.5 text-[12.5px] font-semibold text-white transition hover:bg-brand-600 disabled:opacity-60"
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
      <aside class="space-y-4">
        <div class="rounded-xl border border-ink-900/10 bg-white p-4">
          <p class="text-[9.5px] font-semibold tracking-[0.16em] text-ink-900/40 uppercase">Company status</p>
          <p class="mt-2 text-[16px] font-semibold capitalize">{{ statusLabel }}</p>
          <p class="mt-1 text-[11.5px] leading-relaxed text-ink-900/45">
            Unverified companies cannot publish public rates or receive unrestricted marketplace
            bookings.
          </p>
          <dl class="mt-4 space-y-2 border-t border-ink-900/8 pt-3 text-[11.5px]">
            <div class="flex justify-between gap-2">
              <dt class="text-ink-900/40">Onboarding</dt>
              <dd class="font-medium capitalize">{{ detail.company.onboarding_status.replace('_', ' ') }}</dd>
            </div>
            <div class="flex justify-between gap-2">
              <dt class="text-ink-900/40">Owner email</dt>
              <dd class="font-medium" :class="user?.email_verified ? 'text-emerald-600' : 'text-amber-600'">
                {{ user?.email_verified ? 'Verified' : 'Unverified' }}
              </dd>
            </div>
          </dl>
        </div>

        <div v-if="!user?.email_verified" class="rounded-xl border border-amber-200 bg-amber-50 p-4">
          <p class="flex items-center gap-1.5 text-[12px] font-semibold text-amber-800">
            <MailCheck :size="14" /> Verify your email
          </p>
          <p class="mt-1.5 text-[11.5px] leading-relaxed text-amber-800/80">
            Use the development verification link below to confirm the owner email before
            submitting.
          </p>
          <NuxtLink :to="verifyLink" class="mt-2 inline-block text-[11.5px] font-semibold text-amber-900 underline underline-offset-2">
            Open verification page
          </NuxtLink>
        </div>

        <div class="rounded-xl border border-ink-900/10 bg-canvas p-4">
          <p class="text-[11.5px] font-semibold">What happens next?</p>
          <ol class="mt-2 space-y-2 text-[11.5px] leading-relaxed text-ink-900/50">
            <li>1. We review your company profile and documents.</li>
            <li>2. Once verified, you can publish rates and receive bookings.</li>
            <li>3. Shipment tools arrive in the next development stage.</li>
          </ol>
        </div>
      </aside>
    </div>
  </div>
</template>
