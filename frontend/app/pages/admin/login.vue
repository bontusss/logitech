<script setup lang="ts">
import { ArrowRight, CircleAlert, Loader2, ShieldCheck, UserPlus } from 'lucide-vue-next'

definePageMeta({ layout: 'auth' })
useSeoMeta({ title: 'Admin sign in', robots: 'noindex' })

const route = useRoute()
const { login, logout } = useAuth()

const email = ref('')
const password = ref('')
const error = ref('')
const pending = ref(false)

async function submit() {
  error.value = ''
  pending.value = true
  try {
    const session = await login(email.value, password.value)
    if (!session.user.platform_role) {
      await logout()
      error.value = 'This account does not have platform administrator access.'
      return
    }
    const next = typeof route.query.next === 'string' ? route.query.next : '/admin'
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
      <span class="flex size-11 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-600 to-violet-600 text-white shadow-md shadow-brand-600/25">
        <ShieldCheck :size="19" />
      </span>
      <p class="text-[10px] font-bold tracking-[0.2em] text-brand-600 uppercase">
        Platform administration
      </p>
    </div>
    <h1 class="mt-4 text-[1.75rem] leading-[1.1] font-semibold tracking-[-0.035em]">
      Admin sign in
    </h1>
    <p class="mt-3 text-[13.5px] font-medium leading-relaxed text-ink-900/65">
      Review marketplace customers, approve logistics providers and manage admin accounts.
    </p>

    <form class="mt-7 space-y-4" @submit.prevent="submit">
      <label class="block">
        <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Admin email</span>
        <input
          v-model="email"
          type="email"
          required
          autocomplete="email"
          class="w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 text-sm transition placeholder:text-ink-900/35 hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5"
          placeholder="admin@bol.example"
        />
      </label>

      <label class="block">
        <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Password</span>
        <PasswordInput v-model="password" required autocomplete="current-password" placeholder="Your password" />
      </label>

      <p v-if="error" role="alert" class="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
        <CircleAlert :size="14" class="mt-0.5 shrink-0" />
        {{ error }}
      </p>

      <button
        type="submit"
        :disabled="pending"
        class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-brand-600 to-violet-600 px-4 py-3 text-sm font-semibold text-white shadow-[0_12px_26px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-60"
      >
        <Loader2 v-if="pending" :size="15" class="animate-spin" />
        {{ pending ? 'Signing in…' : 'Sign in to admin' }}
        <ArrowRight v-if="!pending" :size="15" />
      </button>
    </form>

    <p class="mt-6 text-center text-[12.5px] font-medium text-ink-900/65">
      Customer or company account?
      <NuxtLink to="/login" class="font-semibold text-brand-600 hover:text-brand-700">
        Use the workspace sign in
      </NuxtLink>
    </p>

    <p class="mt-4 flex items-start gap-2 rounded-xl border border-violet-200/60 bg-gradient-to-br from-violet-50/80 to-cyan-50/60 px-3 py-2.5 text-[11px] leading-relaxed text-ink-900/50 dark:border-white/10 dark:from-violet-400/10 dark:to-cyan-400/5">
      <UserPlus :size="14" class="mt-0.5 shrink-0 text-violet-500 dark:text-violet-300" />
      Admin accounts are created by the super administrator inside the admin dashboard.
    </p>
  </div>
</template>
