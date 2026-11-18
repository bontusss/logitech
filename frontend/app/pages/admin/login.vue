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
    <p class="flex items-center gap-1.5 text-[11px] font-semibold tracking-[0.2em] text-brand-600 uppercase">
      <ShieldCheck :size="13" /> Platform administration
    </p>
    <h1 class="mt-2 text-[1.8rem] leading-[1.1] font-semibold tracking-[-0.035em]">
      Admin sign in
    </h1>
    <p class="mt-3 text-[13.5px] leading-relaxed text-ink-900/50">
      Review marketplace customers, approve logistics providers and manage admin accounts.
    </p>

    <form class="mt-7 space-y-4" @submit.prevent="submit">
      <label class="block">
        <span class="mb-1.5 block text-[11.5px] font-medium text-ink-900/60">Admin email</span>
        <input
          v-model="email"
          type="email"
          required
          autocomplete="email"
          class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-sm transition hover:border-slate-400 focus:border-brand-500"
          placeholder="admin@bol.example"
        />
      </label>

      <label class="block">
        <span class="mb-1.5 block text-[11.5px] font-medium text-ink-900/60">Password</span>
        <PasswordInput v-model="password" required autocomplete="current-password" placeholder="Your password" />
      </label>

      <p v-if="error" role="alert" class="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
        <CircleAlert :size="14" class="mt-0.5 shrink-0" />
        {{ error }}
      </p>

      <button
        type="submit"
        :disabled="pending"
        class="flex w-full items-center justify-center gap-2 rounded-lg bg-brand-500 px-4 py-3 text-sm font-semibold text-white transition hover:bg-brand-600 disabled:opacity-60"
      >
        <Loader2 v-if="pending" :size="15" class="animate-spin" />
        {{ pending ? 'Signing in…' : 'Sign in to admin' }}
        <ArrowRight v-if="!pending" :size="15" />
      </button>
    </form>

    <p class="mt-6 text-center text-[12.5px] text-ink-900/50">
      Customer or company account?
      <NuxtLink to="/login" class="font-semibold text-brand-600 hover:text-brand-700">
        Use the workspace sign in
      </NuxtLink>
    </p>

    <p class="mt-4 flex items-start gap-2 rounded-lg border border-ink-900/8 bg-canvas px-3 py-2.5 text-[11px] leading-relaxed text-ink-900/45">
      <UserPlus :size="14" class="mt-0.5 shrink-0 text-ink-900/35" />
      Admin accounts are created by the super administrator inside the admin dashboard.
    </p>
  </div>
</template>
