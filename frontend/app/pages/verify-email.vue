<script setup lang="ts">
import { ArrowRight, CheckCircle2, MailCheck, TriangleAlert } from 'lucide-vue-next'

definePageMeta({ layout: 'auth' })
useSeoMeta({ title: 'Verify email' })

const route = useRoute()
const { api } = useApi()
const { user, fetchMe } = useAuth()

const tokenInput = ref(typeof route.query.token === 'string' ? route.query.token : '')
const state = ref<'idle' | 'pending' | 'success' | 'error'>('idle')
const message = ref('')
const resendMessage = ref('')
const resending = ref(false)

onMounted(async () => {
  await fetchMe()
  if (import.meta.client) {
    const stored = localStorage.getItem('bol_dev_verify_token')
    if (!tokenInput.value && stored) tokenInput.value = stored
  }
  if (tokenInput.value) verify()
})

async function verify() {
  if (!tokenInput.value) return
  state.value = 'pending'
  message.value = ''
  try {
    await api('/auth/verify-email', {
      method: 'POST',
      body: { token: tokenInput.value },
    })
    state.value = 'success'
    if (import.meta.client) localStorage.removeItem('bol_dev_verify_token')
    await fetchMe()
  } catch (err) {
    state.value = 'error'
    message.value = apiErrorMessage(err, 'This verification link is invalid or has expired.')
  }
}

async function resendVerification() {
  resending.value = true
  resendMessage.value = ''
  try {
    const response = await api<{ message: string }>('/auth/email/resend-verification', {
      method: 'POST',
    })
    resendMessage.value = response.message
  } catch (err) {
    resendMessage.value = apiErrorMessage(err, 'Could not send a verification email.')
  } finally {
    resending.value = false
  }
}

const inputClass =
  'w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 text-sm transition placeholder:text-ink-900/35 hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5'
</script>

<template>
  <div>
    <div
      class="flex size-12 items-center justify-center rounded-2xl shadow-sm"
      :class="state === 'success' ? 'bg-gradient-to-br from-emerald-100 to-teal-50 text-emerald-700 dark:from-emerald-400/20 dark:to-teal-400/10 dark:text-emerald-200' : 'bg-gradient-to-br from-violet-100 to-cyan-50 text-violet-700 dark:from-violet-400/15 dark:to-cyan-400/10 dark:text-violet-200'"
    >
      <CheckCircle2 v-if="state === 'success'" :size="21" />
      <MailCheck v-else :size="21" />
    </div>

    <h1 class="mt-4 text-[1.7rem] leading-[1.15] font-semibold tracking-[-0.035em]">
      {{ state === 'success' ? 'Email verified' : 'Verify your email' }}
    </h1>

    <template v-if="state === 'success'">
      <p class="mt-3 text-[13.5px] font-medium leading-relaxed text-ink-900/65">
        Your email is confirmed. You can now complete company onboarding and submit your company
        for review.
      </p>
      <div class="mt-6 flex flex-col gap-2 sm:flex-row">
        <AppButton to="/onboarding">
          Continue onboarding
          <ArrowRight :size="15" />
        </AppButton>
        <AppButton to="/dashboard" variant="secondary">Go to dashboard</AppButton>
      </div>
    </template>

    <template v-else>
      <p class="mt-3 text-[13.5px] font-medium leading-relaxed text-ink-900/65">
        Open the verification link in your inbox. You can also paste the token from the email
        below.
      </p>

      <form class="mt-6 space-y-4" @submit.prevent="verify">
        <label class="block">
          <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Verification token</span>
          <input v-model="tokenInput" type="text" required :class="inputClass" placeholder="Paste the token from your email" />
        </label>

        <p
          v-if="state === 'error'"
          role="alert"
          class="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700"
        >
          <TriangleAlert :size="14" class="mt-0.5 shrink-0" />
          {{ message }}
        </p>

        <button
          type="submit"
          :disabled="state === 'pending'"
          class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-brand-600 to-violet-600 px-4 py-3 text-sm font-semibold text-white shadow-[0_12px_26px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-60"
        >
          {{ state === 'pending' ? 'Verifying…' : 'Verify email' }}
        </button>
      </form>

      <div v-if="user && !user.email_verified" class="mt-4">
        <button
          type="button"
          :disabled="resending"
          class="text-[11.5px] font-semibold text-brand-600 underline underline-offset-2 disabled:opacity-50"
          @click="resendVerification"
        >
          {{ resending ? 'Sending…' : 'Resend verification email' }}
        </button>
        <p v-if="resendMessage" role="status" class="mt-2 text-[11.5px] text-ink-900/70">
          {{ resendMessage }}
        </p>
      </div>

      <p class="mt-5 text-[11.5px] leading-relaxed text-ink-900/55">
        If the email does not arrive shortly, check your spam folder or request another link.
      </p>
    </template>
  </div>
</template>
