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
</script>

<template>
  <div>
    <IconTile
      :tone="state === 'success' ? 'success' : 'accent-solid'"
      size="lg"
      rounded="2xl"
    >
      <CheckCircle2 v-if="state === 'success'" :size="21" />
      <MailCheck v-else :size="21" />
    </IconTile>

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
        <FormTextField
          v-model="tokenInput"
          label="Verification token"
          placeholder="Paste the token from your email"
          required
        />

        <p
          v-if="state === 'error'"
          role="alert"
          class="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700 dark:border-red-400/25 dark:bg-red-400/10 dark:text-red-200"
        >
          <TriangleAlert :size="14" class="mt-0.5 shrink-0" />
          {{ message }}
        </p>

        <SubmitButton
          :pending="state === 'pending'"
          label="Verify email"
          pending-label="Verifying…"
          :icon="false"
        />
      </form>

      <div v-if="user && !user.email_verified" class="mt-4">
        <button
          type="button"
          :disabled="resending"
          class="text-[11.5px] font-semibold text-accent-700 underline underline-offset-2 disabled:opacity-50 dark:text-accent-400"
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
