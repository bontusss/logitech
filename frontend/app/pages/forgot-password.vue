<script setup lang="ts">
import { ArrowLeft, ArrowRight, CheckCircle2 } from 'lucide-vue-next'

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
  'w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-sm transition hover:border-slate-400 focus:border-brand-500'
</script>

<template>
  <div>
    <NuxtLink to="/login" class="inline-flex items-center gap-1.5 text-[12px] font-medium text-ink-900/45 hover:text-ink-900">
      <ArrowLeft :size="13" />
      Back to sign in
    </NuxtLink>

    <template v-if="done">
      <div class="mt-6 flex size-11 items-center justify-center rounded-full bg-brand-50 text-brand-600">
        <CheckCircle2 :size="22" />
      </div>
      <h1 class="mt-4 text-[1.7rem] font-semibold tracking-[-0.03em]">Password updated</h1>
      <p class="mt-3 text-[13.5px] leading-relaxed text-ink-900/50">
        Your password has been reset and existing sessions were signed out.
      </p>
      <AppButton to="/login" class="mt-6">
        Sign in
        <ArrowRight :size="15" />
      </AppButton>
    </template>

    <template v-else-if="token">
      <h1 class="mt-5 text-[1.7rem] font-semibold tracking-[-0.03em]">Choose a new password</h1>
      <p class="mt-3 text-[13.5px] leading-relaxed text-ink-900/50">
        Use at least 8 characters. Resetting will sign out your other sessions.
      </p>

      <form class="mt-6 space-y-4" @submit.prevent="resetPassword">
        <label class="block">
          <span class="mb-1.5 block text-[11.5px] font-medium text-ink-900/60">New password</span>
          <input v-model="password" type="password" required minlength="8" autocomplete="new-password" :class="inputClass" placeholder="At least 8 characters" />
        </label>

        <p v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
          {{ error }}
        </p>

        <button
          type="submit"
          :disabled="pending"
          class="flex w-full items-center justify-center gap-2 rounded-lg bg-brand-500 px-4 py-3 text-sm font-semibold text-white transition hover:bg-brand-600 disabled:opacity-60"
        >
          {{ pending ? 'Saving…' : 'Set new password' }}
        </button>
      </form>
    </template>

    <template v-else>
      <h1 class="mt-5 text-[1.7rem] font-semibold tracking-[-0.03em]">Reset your password</h1>
      <p class="mt-3 text-[13.5px] leading-relaxed text-ink-900/50">
        Enter the email on your account and we will send a reset link.
      </p>

      <form class="mt-6 space-y-4" @submit.prevent="requestReset">
        <label class="block">
          <span class="mb-1.5 block text-[11.5px] font-medium text-ink-900/60">Email address</span>
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
          class="flex w-full items-center justify-center gap-2 rounded-lg bg-brand-500 px-4 py-3 text-sm font-semibold text-white transition hover:bg-brand-600 disabled:opacity-60"
        >
          {{ pending ? 'Sending…' : 'Send reset link' }}
          <ArrowRight v-if="!pending" :size="15" />
        </button>
      </form>
    </template>
  </div>
</template>
