<script setup lang="ts">
import { Building2, CircleAlert, CircleCheck, Loader2, LogIn, MailCheck, ShieldCheck } from 'lucide-vue-next'
import type { PublicInvitation, SessionResponse } from '~/types/auth'

definePageMeta({ layout: 'auth' })
useSeoMeta({ title: 'Accept invitation' })

const route = useRoute()
const { api } = useApi()
const { user, fetchMe, logout } = useAuth()
const { selectCompany } = useCompany()

const token = computed(() => String(route.params.token))
const preview = ref<PublicInvitation | null>(null)
const loading = ref(true)
const error = ref('')
const accepting = ref(false)
const fullName = ref('')
const password = ref('')

const invitationPath = computed(() => `/invitations/${token.value}`)
const signInLink = computed(() => `/login?next=${encodeURIComponent(invitationPath.value)}`)

const isInvitee = computed(
  () => Boolean(preview.value && user.value &&
    user.value.email.toLowerCase() === preview.value.invitation.email.toLowerCase()),
)

onMounted(async () => {
  await fetchMe()
  await loadInvitation()
})

async function loadInvitation() {
  loading.value = true
  error.value = ''
  try {
    preview.value = await api<PublicInvitation>(`/invitations/${token.value}`)
  } catch (err) {
    error.value = apiErrorMessage(err, 'This invitation is invalid, expired or has already been used.')
    preview.value = null
  } finally {
    loading.value = false
  }
}

async function accept() {
  if (!preview.value) return
  error.value = ''
  if (!preview.value.invitation.account_exists && (!fullName.value.trim() || password.value.length < 8)) {
    error.value = 'Enter your full name and a password of at least 8 characters.'
    return
  }
  accepting.value = true
  try {
    const session = await api<SessionResponse>(`/invitations/${token.value}/accept`, {
      method: 'POST',
      body: preview.value.invitation.account_exists
        ? {}
        : { full_name: fullName.value.trim(), password: password.value },
    })
    await fetchMe()
    selectCompany(session.companies[0]?.company_id ?? null)
    await navigateTo('/dashboard/company')
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not accept the invitation.')
    await loadInvitation()
  } finally {
    accepting.value = false
  }
}

async function useAnotherAccount() {
  await logout()
  await navigateTo(signInLink.value)
}
</script>

<template>
  <div>
    <div v-if="loading" class="flex items-center gap-2 rounded-xl border border-ink-900/10 bg-canvas p-6 text-[13px] text-ink-900/50 dark:border-white/10 dark:bg-white/5">
      <Loader2 :size="16" class="animate-spin" /> Loading invitation…
    </div>

    <div v-else-if="!preview" class="rounded-2xl border border-ink-900/10 bg-canvas p-6 text-center dark:border-white/10 dark:bg-white/5">
      <div class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-amber-100 text-amber-700 dark:bg-amber-400/15 dark:text-amber-200">
        <CircleAlert :size="20" />
      </div>
      <h1 class="mt-3 text-[18px] font-semibold tracking-[-0.02em]">Invitation unavailable</h1>
      <p class="mt-2 text-[12.5px] leading-relaxed text-ink-900/65">{{ error }}</p>
      <div class="mt-5 flex justify-center gap-2">
        <AppButton to="/login" size="sm">Sign in</AppButton>
        <AppButton to="/" size="sm" variant="secondary">Back to site</AppButton>
      </div>
    </div>

    <template v-else>
      <div class="text-center">
        <span class="mx-auto flex size-12 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-600 to-violet-600 text-[15px] font-semibold text-white shadow-lg shadow-brand-600/25">
          {{ preview.company.name.slice(0, 1).toUpperCase() }}
        </span>
        <p class="mt-3 inline-flex items-center gap-1.5 text-[10px] font-bold tracking-[0.18em] text-brand-600 uppercase">
          <span class="size-1.5 rounded-full bg-gradient-to-r from-brand-500 to-violet-500" />
          Staff invitation
        </p>
        <h1 class="mt-1.5 text-[22px] font-semibold tracking-[-0.03em]">{{ preview.company.name }}</h1>
        <p class="mt-1 text-[12.5px] font-medium text-ink-900/60">has invited you to join their workspace</p>
      </div>

      <div class="mt-5 rounded-2xl border border-ink-900/8 bg-gradient-to-br from-violet-50/50 to-cyan-50/40 p-4 dark:border-white/10 dark:from-violet-400/8 dark:to-cyan-400/5">
        <dl class="space-y-2.5 text-[12px]">
          <div class="flex items-start justify-between gap-3">
            <dt class="flex items-center gap-1.5 text-ink-900/60"><MailCheck :size="13" /> Invited email</dt>
            <dd class="text-right font-medium">{{ preview.invitation.email }}</dd>
          </div>
          <div class="flex items-start justify-between gap-3">
            <dt class="flex items-center gap-1.5 text-ink-900/60"><ShieldCheck :size="13" /> Role</dt>
            <dd class="text-right font-medium">{{ preview.invitation.role_name }}</dd>
          </div>
          <div v-if="preview.invitation.branches.length" class="flex items-start justify-between gap-3">
            <dt class="flex items-center gap-1.5 text-ink-900/60"><Building2 :size="13" /> Branches</dt>
            <dd class="text-right font-medium">
              {{ preview.invitation.branches.map((branch) => branch.name).join(', ') }}
            </dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-t border-ink-900/8 pt-2.5">
            <dt class="text-ink-900/60">Expires</dt>
            <dd class="text-right font-medium">{{ new Date(preview.invitation.expires_at).toLocaleDateString() }}</dd>
          </div>
        </dl>
      </div>

      <p v-if="error" role="alert" class="mt-3 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
        <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ error }}
      </p>

      <!-- Existing account, signed in as the invited email -->
      <div v-if="preview.invitation.account_exists && isInvitee" class="mt-4">
        <button
          type="button"
          :disabled="accepting"
          class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-brand-600 to-violet-600 px-4 py-3 text-[13px] font-semibold text-white shadow-[0_12px_26px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-60"
          @click="accept"
        >
          <Loader2 v-if="accepting" :size="15" class="animate-spin" />
          <CircleCheck v-else :size="15" />
          Accept invitation
        </button>
        <p class="mt-2 text-center text-[10.5px] text-ink-900/55">
          Signed in as {{ user?.email }}
        </p>
      </div>

      <!-- Existing account that needs to sign in -->
      <div v-else-if="preview.invitation.account_exists" class="mt-4">
        <p class="mb-3 rounded-xl border border-brand-200/60 bg-gradient-to-br from-brand-50 to-violet-50/60 px-3 py-2.5 text-[11.5px] leading-relaxed text-brand-800 dark:border-brand-400/20 dark:from-brand-400/10 dark:to-violet-400/5 dark:text-indigo-200">
          This email already has a BOL account. Sign in with that account to accept the invitation.
        </p>
        <AppButton :to="signInLink" block size="md">
          <LogIn :size="15" /> Sign in to accept
        </AppButton>
        <button
          v-if="user"
          type="button"
          class="mt-2 w-full text-center text-[10.5px] text-ink-900/60 underline underline-offset-2 hover:text-ink-900"
          @click="useAnotherAccount"
        >
          Signed in as {{ user.email }} · use another account
        </button>
      </div>

      <!-- New account -->
      <form v-else class="mt-4" @submit.prevent="accept">
        <p class="text-[12px] leading-relaxed text-ink-900/65">
          Create your account with the invited email. You will join the company as
          <strong class="font-semibold text-ink-900">{{ preview.invitation.role_name }}</strong>.
        </p>
        <label class="mt-3 block">
          <span class="mb-1.5 block text-[11px] font-semibold text-ink-900/70">Full name</span>
          <input v-model="fullName" type="text" required minlength="2" class="w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 text-sm transition hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5" />
        </label>
        <label class="mt-3 block">
          <span class="mb-1.5 block text-[11px] font-semibold text-ink-900/70">Password</span>
          <PasswordInput
            v-model="password"
            required
            :minlength="8"
            input-class="w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 text-sm transition hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5"
            placeholder="At least 8 characters"
          />
        </label>
        <button
          type="submit"
          :disabled="accepting"
          class="mt-4 flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-brand-600 to-violet-600 px-4 py-3 text-[13px] font-semibold text-white shadow-[0_12px_26px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-60"
        >
          <Loader2 v-if="accepting" :size="15" class="animate-spin" />
          <CircleCheck v-else :size="15" />
          Accept invitation
        </button>
      </form>

      <p class="mt-4 text-center text-[10px] leading-relaxed text-ink-900/55">
        Invitations expire after seven days and can only be used once. If this was not expected, you
        can safely ignore it.
      </p>
    </template>
  </div>
</template>
