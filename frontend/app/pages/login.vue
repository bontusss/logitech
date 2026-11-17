<script setup lang="ts">
import { ArrowRight } from 'lucide-vue-next'

definePageMeta({ layout: 'auth' })
useSeoMeta({ title: 'Sign in', description: 'Sign in to your BOL logistics workspace.' })

const route = useRoute()
const { login } = useAuth()

const email = ref('')
const password = ref('')
const error = ref('')
const pending = ref(false)

async function submit() {
  error.value = ''
  pending.value = true
  try {
    await login(email.value, password.value)
    const next = typeof route.query.next === 'string' ? route.query.next : '/dashboard'
    await navigateTo(next)
  } catch (err) {
    error.value = apiErrorMessage(err, 'Email or password is incorrect.')
  } finally {
    pending.value = false
  }
}
</script>

<template>
  <div>
    <p class="text-[11px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Welcome back</p>
    <h1 class="mt-2 text-[1.8rem] leading-[1.1] font-semibold tracking-[-0.035em]">
      Sign in to your workspace
    </h1>
    <p class="mt-3 text-[13.5px] leading-relaxed text-ink-900/50">
      Manage your company, branches and onboarding.
    </p>

    <form class="mt-7 space-y-4" @submit.prevent="submit">
      <label class="block">
        <span class="mb-1.5 block text-[11.5px] font-medium text-ink-900/60">Email address</span>
        <input
          v-model="email"
          type="email"
          required
          autocomplete="email"
          class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-sm transition hover:border-slate-400 focus:border-brand-500"
          placeholder="you@company.com"
        />
      </label>

      <label class="block">
        <span class="mb-1.5 flex items-center justify-between text-[11.5px] font-medium text-ink-900/60">
          Password
          <NuxtLink to="/forgot-password" class="font-normal text-brand-600 hover:text-brand-700">
            Forgot password?
          </NuxtLink>
        </span>
        <input
          v-model="password"
          type="password"
          required
          autocomplete="current-password"
          class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-sm transition hover:border-slate-400 focus:border-brand-500"
          placeholder="Your password"
        />
      </label>

      <p v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
        {{ error }}
      </p>

      <button
        type="submit"
        :disabled="pending"
        class="flex w-full items-center justify-center gap-2 rounded-lg bg-brand-500 px-4 py-3 text-sm font-semibold text-white transition hover:bg-brand-600 disabled:opacity-60"
      >
        {{ pending ? 'Signing in…' : 'Sign in' }}
        <ArrowRight v-if="!pending" :size="15" />
      </button>
    </form>

    <p class="mt-6 text-center text-[12.5px] text-ink-900/50">
      New to BOL?
      <NuxtLink to="/register" class="font-semibold text-brand-600 hover:text-brand-700">
        Create an account
      </NuxtLink>
    </p>
  </div>
</template>
