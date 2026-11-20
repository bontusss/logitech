<script setup lang="ts">
import { ArrowLeft, ArrowRight, CheckCircle2, KeyRound } from 'lucide-vue-next'

definePageMeta({ layout: 'auth' })
useSeoMeta({ title: 'Reset password' })

const route = useRoute()
const { api } = useApi()

const token = computed(() => (typeof route.query.token === 'string' ? route.query.token : ''))
const email = ref('')
const password = ref('')
const error = ref('')
const message = ref('')
const devToken = ref('')
const pending = ref(false)
const done = ref(false)

async function requestReset() {
  error.value = ''
  message.value = ''
  pending.value = true
  try {
    const response = await api<{ message: string; dev_reset_token?: string }>('/auth/forgot-password', {
      method: 'POST',
      body: { email: email.value },
    })
    message.value = response.message
    devToken.value = response.dev_reset_token ?? ''
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not request a reset link.')
  } finally {
    pending.value = false
  }
}

async function resetPassword() {
  error.value = ''
  pending.value = true
  try {
    await api('/auth/reset-password', {
      method: 'POST',
      body: { token: token.value, password: password.value },
    })
    done.value = true
  } catch (err) {
    error.value = apiErrorMessage(err, 'This reset link is invalid or has expired.')
  } finally {
    pending.value = false
  }
}

const inputClass =
  'w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 text-sm transition placeholder:text-ink-900/35 hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5'
</script>

<template>
  <div>
    <NuxtLink to="/login" class="inline-flex items-center gap-1.5 rounded-full border border-ink-900/10 bg-white/70 px-3 py-1.5 text-[12px] font-semibold text-ink-900/65 transition hover:-translate-x-0.5 hover:text-ink-900 dark:border-white/10 dark:bg-white/5">
      <ArrowLeft :size="13" />
      Back to sign in
    </NuxtLink>

    <template v-if="done">
      <div class="mt-6 flex size-12 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-500 text-white shadow-lg">
        <CheckCircle2 :size="22" />
      </div>
      <h1 class="mt-4 text-[1.7rem] font-semibold tracking-[-0.03em]">Password updated</h1>
      <p class="mt-3 text-[13.5px] font-medium leading-relaxed text-ink-900/65">
        Your password has been reset and existing sessions were signed out.
      </p>
      <AppButton to="/login" class="mt-6">
        Sign in
        <ArrowRight :size="15" />
      </AppButton>
    </template>

    <template v-else-if="token">
      <div class="mt-6 flex size-11 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-100 to-cyan-50 text-violet-700 shadow-sm dark:from-violet-400/15 dark:to-cyan-400/10 dark:text-violet-200">
        <KeyRound :size="19" />
      </div>
      <h1 class="mt-4 text-[1.7rem] font-semibold tracking-[-0.03em]">Choose a new password</h1>
      <p class="mt-3 text-[13.5px] font-medium leading-relaxed text-ink-900/65">
        Use at least 8 characters. Resetting will sign out your other sessions.
      </p>

      <form class="mt-6 space-y-4" @submit.prevent="resetPassword">
        <label class="block">
          <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">New password</span>
          <PasswordInput v-model="password" required :minlength="8" autocomplete="new-password" :input-class="inputClass" placeholder="At least 8 characters" />
        </label>

        <p v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
          {{ error }}
        </p>

        <button
          type="submit"
          :disabled="pending"
          class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-brand-600 to-violet-600 px-4 py-3 text-sm font-semibold text-white shadow-[0_12px_26px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-60"
        >
          {{ pending ? 'Saving…' : 'Set new password' }}
        </button>
      </form>
    </template>

    <template v-else>
      <div class="mt-6 flex size-11 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-100 to-cyan-50 text-violet-700 shadow-sm dark:from-violet-400/15 dark:to-cyan-400/10 dark:text-violet-200">
        <KeyRound :size="19" />
      </div>
      <h1 class="mt-4 text-[1.7rem] font-semibold tracking-[-0.03em]">Reset your password</h1>
      <p class="mt-3 text-[13.5px] font-medium leading-relaxed text-ink-900/65">
        Enter the email on your account and we will send a reset link.
      </p>

      <form class="mt-6 space-y-4" @submit.prevent="requestReset">
        <label class="block">
          <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Email address</span>
          <input v-model="email" type="email" required autocomplete="email" :class="inputClass" placeholder="you@company.com" />
        </label>

        <p v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
          {{ error }}
        </p>
        <p v-if="message" role="status" class="rounded-lg border border-brand-200 bg-brand-50 px-3 py-2.5 text-[12px] text-brand-700">
          {{ message }}
          <NuxtLink
            v-if="devToken"
            :to="`/forgot-password?token=${devToken}`"
            class="mt-1 block font-semibold underline underline-offset-2"
          >
            Development shortcut: open reset link
          </NuxtLink>
        </p>

        <button
          type="submit"
          :disabled="pending"
          class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-brand-600 to-violet-600 px-4 py-3 text-sm font-semibold text-white shadow-[0_12px_26px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-60"
        >
          {{ pending ? 'Sending…' : 'Send reset link' }}
          <ArrowRight v-if="!pending" :size="15" />
        </button>
      </form>
    </template>
  </div>
</template>
