<script setup lang="ts">
import { ArrowRight, LogIn } from 'lucide-vue-next'

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
    const session = await login(email.value, password.value)
    if (session.user.platform_role) {
      await navigateTo('/admin')
      return
    }
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
    <div class="flex items-center gap-3">
      <span class="flex size-11 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-100 to-cyan-50 text-violet-700 shadow-sm dark:from-violet-400/15 dark:to-cyan-400/10 dark:text-violet-200">
        <LogIn :size="19" />
      </span>
      <p class="text-[10px] font-bold tracking-[0.2em] text-brand-600 uppercase">Welcome back</p>
    </div>
    <h1 class="mt-4 text-[1.75rem] leading-[1.1] font-semibold tracking-[-0.035em]">
      Sign in to your workspace
    </h1>
    <p class="mt-2.5 text-[13.5px] font-medium leading-relaxed text-ink-900/65">
      Manage your company, branches and onboarding.
    </p>

    <form class="mt-7 space-y-4" @submit.prevent="submit">
      <label class="block">
        <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Email address</span>
        <input
          v-model="email"
          type="email"
          required
          autocomplete="email"
          class="w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 text-sm transition placeholder:text-ink-900/35 hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5"
          placeholder="you@company.com"
        />
      </label>

      <label class="block">
        <span class="mb-1.5 flex items-center justify-between text-[11.5px] font-semibold text-ink-900/75">
          Password
          <NuxtLink to="/forgot-password" class="font-normal text-brand-600 hover:text-brand-700">
            Forgot password?
          </NuxtLink>
        </span>
        <PasswordInput
          v-model="password"
          required
          autocomplete="current-password"
          placeholder="Your password"
        />
      </label>

      <p v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
        {{ error }}
      </p>

      <button
        type="submit"
        :disabled="pending"
        class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-brand-600 to-violet-600 px-4 py-3 text-sm font-semibold text-white shadow-[0_12px_26px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-60"
      >
        {{ pending ? 'Signing in…' : 'Sign in' }}
        <ArrowRight v-if="!pending" :size="15" />
      </button>
    </form>

    <p class="mt-6 text-center text-[12.5px] font-medium text-ink-900/65">
      New to BOL?
      <NuxtLink to="/register" class="font-semibold text-brand-600 hover:text-brand-700">
        Create an account
      </NuxtLink>
    </p>
  </div>
</template>
