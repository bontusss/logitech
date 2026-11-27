<script setup lang="ts">
import {
  BadgeCheck,
  CircleAlert,
  Loader2,
  Mail,
  Phone,
  ShieldCheck,
  UserRound,
} from 'lucide-vue-next'
import type { AuthUser } from '~/types/auth'

useHead({ title: 'Settings · BOL' })

const { api } = useApi()
const { user, loaded, fetchMe } = useAuth()
const { notify } = useWorkspace()

const form = reactive({ full_name: '', phone: '' })
const saving = ref(false)
const error = ref('')
const resending = ref(false)

const accountTypeLabel = computed(() => {
  const type = user.value?.account_type
  if (type === 'customer') return 'Customer account'
  if (type === 'company') return 'Company account'
  return 'BOL account'
})

function syncForm() {
  form.full_name = user.value?.full_name || ''
  form.phone = user.value?.phone || ''
}

onMounted(async () => {
  if (!loaded.value || !user.value) {
    await fetchMe()
  }
  syncForm()
})

async function save() {
  if (!form.full_name.trim() || form.full_name.trim().length < 2) {
    error.value = 'Enter your full name.'
    return
  }
  saving.value = true
  error.value = ''
  try {
    const result = await api<{ user: AuthUser }>('/users/me', {
      method: 'PATCH',
      body: {
        full_name: form.full_name.trim(),
        phone: form.phone.trim(),
      },
    })
    user.value = result.user
    syncForm()
    notify('Profile updated.')
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not save your profile.')
  } finally {
    saving.value = false
  }
}

async function resendVerification() {
  resending.value = true
  error.value = ''
  try {
    await api('/auth/email/resend-verification', { method: 'POST' })
    notify('Verification email sent. Check your inbox.')
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not send the verification email.')
  } finally {
    resending.value = false
  }
}

function formatDate(value: string | null | undefined) {
  if (!value) return '—'
  return new Date(value).toLocaleDateString('en-NG', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}
</script>

<template>
  <div class="dash-page">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Account</p>
        <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Settings</h1>
        <p class="mt-1.5 max-w-2xl text-[12.5px] leading-relaxed text-ink-900/50">
          Keep your personal details current so providers know who is sending the parcel.
        </p>
      </div>
    </div>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ error }}
    </p>

    <div class="mt-5 grid gap-4 lg:grid-cols-[minmax(0,1.4fr)_minmax(260px,1fr)]">
      <form class="rounded-xl border border-ink-900/10 bg-surface" @submit.prevent="save">
        <div class="border-b border-ink-900/8 px-4 py-3.5">
          <h2 class="text-[13.5px] font-semibold">Personal details</h2>
          <p class="mt-0.5 text-[10.5px] text-ink-900/40">Used on new bookings and provider communication.</p>
        </div>

        <div class="space-y-3 p-4">
          <label class="block">
            <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Full name</span>
            <div class="relative">
              <UserRound :size="14" class="pointer-events-none absolute top-3 left-3 text-ink-900/35" />
              <input
                v-model="form.full_name"
                type="text"
                required
                maxlength="120"
                autocomplete="name"
                class="w-full rounded-xl border border-ink-900/15 bg-canvas py-2.5 pr-3.5 pl-9 text-sm outline-none focus:border-accent-500"
              />
            </div>
          </label>

          <label class="block">
            <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Phone number</span>
            <div class="relative">
              <Phone :size="14" class="pointer-events-none absolute top-3 left-3 text-ink-900/35" />
              <input
                v-model="form.phone"
                type="tel"
                maxlength="30"
                autocomplete="tel"
                placeholder="e.g. 0801 234 5678"
                class="w-full rounded-xl border border-ink-900/15 bg-canvas py-2.5 pr-3.5 pl-9 text-sm outline-none focus:border-accent-500"
              />
            </div>
          </label>

          <label class="block">
            <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Email address</span>
            <div class="relative">
              <Mail :size="14" class="pointer-events-none absolute top-3 left-3 text-ink-900/35" />
              <input
                :value="user?.email || ''"
                type="email"
                disabled
                class="w-full rounded-xl border border-ink-900/12 bg-canvas/60 py-2.5 pr-3.5 pl-9 text-sm text-ink-900/55"
              />
            </div>
            <span class="mt-1.5 flex flex-wrap items-center gap-1.5 text-[10.5px]">
              <span
                class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 font-semibold"
                :class="user?.email_verified
                  ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                  : 'border-amber-200 bg-amber-50 text-amber-700'"
              >
                <BadgeCheck :size="10" /> {{ user?.email_verified ? 'Verified' : 'Not verified' }}
              </span>
              <button
                v-if="user && !user.email_verified"
                type="button"
                :disabled="resending"
                class="font-semibold text-brand-700 hover:text-brand-600 disabled:opacity-50"
                @click="resendVerification"
              >
                {{ resending ? 'Sending…' : 'Resend verification email' }}
              </button>
              <span class="text-ink-900/35">Contact support to change your email.</span>
            </span>
          </label>
        </div>

        <div class="flex justify-end border-t border-ink-900/8 px-4 py-3.5">
          <button
            type="submit"
            :disabled="saving"
            class="inline-flex items-center gap-1.5 rounded-lg bg-accent-500 px-3.5 py-2 text-[11.5px] font-semibold text-white transition hover:bg-accent-400 disabled:opacity-50"
          >
            <Loader2 v-if="saving" :size="13" class="animate-spin" />
            Save changes
          </button>
        </div>
      </form>

      <aside class="space-y-4">
        <section class="rounded-xl border border-ink-900/10 bg-surface p-4">
          <h2 class="text-[13px] font-semibold">Account</h2>
          <dl class="mt-3 space-y-2.5 text-[11px]">
            <div class="flex items-start justify-between gap-3">
              <dt class="text-ink-900/40">Type</dt>
              <dd class="font-medium">{{ accountTypeLabel }}</dd>
            </div>
            <div class="flex items-start justify-between gap-3">
              <dt class="text-ink-900/40">Member since</dt>
              <dd class="font-medium">{{ formatDate(user?.created_at) }}</dd>
            </div>
            <div class="flex items-start justify-between gap-3">
              <dt class="text-ink-900/40">Last sign-in</dt>
              <dd class="font-medium">{{ formatDate(user?.last_login_at) }}</dd>
            </div>
            <div class="flex items-start justify-between gap-3">
              <dt class="text-ink-900/40">WhatsApp</dt>
              <dd>
                <span
                  class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[9.5px] font-semibold"
                  :class="user?.whatsapp_verified
                    ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                    : 'border-ink-900/12 bg-canvas text-ink-900/50'"
                >
                  {{ user?.whatsapp_verified ? 'Verified' : 'Not linked' }}
                </span>
              </dd>
            </div>
          </dl>
        </section>

        <section class="rounded-xl border border-ink-900/10 bg-surface p-4">
          <span class="flex size-8 items-center justify-center rounded-lg bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
            <ShieldCheck :size="14" />
          </span>
          <h2 class="mt-2 text-[13px] font-semibold">Account security</h2>
          <p class="mt-1 text-[11px] leading-relaxed text-ink-900/50">
            Your session is protected by a secure cookie. If you think someone else has access to your account, change your password from the sign-in page.
          </p>
          <NuxtLink
            to="/forgot-password"
            class="mt-2.5 inline-flex text-[11px] font-semibold text-brand-700 hover:text-brand-600"
          >
            Reset my password
          </NuxtLink>
        </section>
      </aside>
    </div>
  </div>
</template>
