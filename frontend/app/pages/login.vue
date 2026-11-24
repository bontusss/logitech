<script setup lang="ts">
import { LogIn } from 'lucide-vue-next'

definePageMeta({ layout: 'auth' })
useSeoMeta({ title: 'Sign in', description: 'Sign in to your BOL logistics workspace.' })

const route = useRoute()
const { login } = useAuth()

const email = ref('')
const password = ref('')
const error = ref('')
const pending = ref(false)

const nextTarget = computed(() =>
  typeof route.query.next === 'string' ? route.query.next : '/dashboard',
)

const googleError = computed(() => {
  const code = route.query.google_error
  if (typeof code !== 'string') return ''
  const messages: Record<string, string> = {
    admin_blocked: 'Administrator accounts must sign in with email and password.',
    no_account: 'No BOL account uses that Google email. Create an account first.',
    email_exists: 'An account with that email already exists. Sign in with your existing method.',
    suspended: 'This account is suspended. Contact support if you need help.',
    cancelled: 'Google sign-in was cancelled. Try again or use your password.',
    not_configured: 'Google sign-in is not configured yet.',
    state: 'That Google sign-in attempt expired. Please try again.',
    google_failed: 'Could not sign in with Google. Please try again.',
  }
  return messages[code] ?? 'Could not sign in with Google. Please try again.'
})

async function submit() {
  error.value = ''
  pending.value = true
  try {
    const session = await login(email.value, password.value)
    if (session.user.platform_role) {
      await navigateTo('/admin')
      return
    }
    await navigateTo(nextTarget.value)
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
      <IconTile tone="accent-solid" size="lg" rounded="2xl">
        <LogIn :size="19" />
      </IconTile>
      <p class="text-[10px] font-bold tracking-[0.2em] text-accent-700 uppercase dark:text-accent-400">
        Welcome back
      </p>
    </div>
    <h1 class="mt-4 text-[1.75rem] leading-[1.1] font-semibold tracking-[-0.035em]">
      Sign in to your workspace
    </h1>
    <p class="mt-2.5 text-[13.5px] font-medium leading-relaxed text-ink-900/65">
      Manage your company, branches and onboarding.
    </p>

    <p
      v-if="googleError"
      role="alert"
      class="mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700 dark:border-red-400/25 dark:bg-red-400/10 dark:text-red-200"
    >
      {{ googleError }}
    </p>

    <div class="mt-5">
      <GoogleAuthButton intent="login" :next="nextTarget" />
      <div class="my-4 flex items-center gap-3">
        <span class="h-px flex-1 bg-ink-900/10" />
        <span class="text-[10px] font-medium tracking-wide text-ink-900/35 uppercase">or with email</span>
        <span class="h-px flex-1 bg-ink-900/10" />
      </div>
    </div>

    <form class="space-y-4" @submit.prevent="submit">
      <FormTextField
        v-model="email"
        label="Email address"
        type="email"
        autocomplete="email"
        placeholder="you@company.com"
        required
      />

      <label class="block">
        <span class="mb-1.5 flex items-center justify-between text-[11.5px] font-semibold text-ink-900/75">
          Password
          <NuxtLink to="/forgot-password" class="font-normal text-accent-700 hover:text-accent-600 dark:text-accent-400">
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

      <p v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700 dark:border-red-400/25 dark:bg-red-400/10 dark:text-red-200">
        {{ error }}
      </p>

      <SubmitButton :pending="pending" label="Sign in" pending-label="Signing in…" />
    </form>

    <p class="mt-6 text-center text-[12.5px] font-medium text-ink-900/65">
      New to BOL?
      <NuxtLink to="/register" class="font-semibold text-accent-700 hover:text-accent-600 dark:text-accent-400">
        Create an account
      </NuxtLink>
    </p>
  </div>
</template>
