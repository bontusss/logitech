<script setup lang="ts">
import { ApiError } from '~/utils/api'

definePageMeta({ layout: 'auth' })

useHead({ title: 'Sign in' })

const route = useRoute()
const { login } = useAuth()
const pendingFirstLogin = useFirstLogin()

// "pin" is a normal sign-in; "otp" is the one-time password a manager or the
// owner hands over after a reset.
const mode = ref<'pin' | 'otp'>('pin')
const staffId = ref('')
const secret = ref('')
const error = ref('')
const failing = ref<'staff' | 'secret' | null>(null)
const attempts = ref(0)
const submitting = ref(false)

const secretLabel = computed(() => (mode.value === 'pin' ? 'Password' : 'One-time password'))
const secretHint = computed(() =>
  mode.value === 'otp' ? 'The password a manager or the owner gave you.' : '',
)

function usePassword() {
  mode.value = 'pin'
  secret.value = ''
  error.value = ''
  failing.value = null
}

function useOneTimePassword() {
  mode.value = 'otp'
  secret.value = ''
  error.value = ''
  failing.value = null
}

function fail(message: string, field: 'staff' | 'secret') {
  error.value = message
  failing.value = field
}

async function submit() {
  if (submitting.value) return

  error.value = ''
  failing.value = null

  if (!staffId.value.trim()) {
    fail('Enter your phone number or staff ID.', 'staff')
    return
  }
  if (!secret.value) {
    fail(mode.value === 'pin' ? 'Enter your password.' : 'Enter your one-time password.', 'secret')
    return
  }

  submitting.value = true
  try {
    await login({ staffId: staffId.value.trim(), pin: secret.value })
    await navigateTo(typeof route.query.redirect === 'string' ? route.query.redirect : '/')
  } catch (err) {
    if (err instanceof ApiError && err.needsNewPin) {
      pendingFirstLogin.value = {
        staffId: staffId.value.trim(),
        oneTimePassword: secret.value,
      }
      await navigateTo('/first-sign-in')
      return
    }

    if (err instanceof ApiError && err.isLocked) {
      await navigateTo({
        path: '/locked',
        query: { staff_id: staffId.value.trim(), at: clockTime() },
      })
      return
    }

    if (err instanceof ApiError && err.status === 401) {
      // The API locks the account on the third failure; keep the count honest
      // instead of guessing how many attempts are left.
      attempts.value += 1
      fail(
        attempts.value >= 2 ? 'Those details are wrong. 1 try left.' : 'Those details are wrong.',
        'secret',
      )
      return
    }

    if (err instanceof ApiError && err.status === 403) {
      fail(err.message, 'staff')
      return
    }

    if (err instanceof ApiError && err.status === 400) {
      fail(err.message, 'secret')
      return
    }

    fail("Can't reach the server. Check your connection and try again.", 'secret')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <UiPanel title="Sign in">
    <form novalidate class="space-y-4" @submit.prevent="submit">
      <UiField
        id="staff-id"
        v-model="staffId"
        label="Phone number or staff ID"
        autocomplete="username"
        :error="failing === 'staff' ? error : ''"
        autofocus
      />

      <UiField
        id="secret"
        v-model="secret"
        :label="secretLabel"
        type="password"
        :autocomplete="mode === 'pin' ? 'current-password' : 'one-time-code'"
        :hint="secretHint"
        :error="failing === 'secret' ? error : ''"
      />

      <UiButton type="submit" :loading="submitting">
        Sign in
        <UiKbd>Enter</UiKbd>
      </UiButton>
    </form>

    <div class="mt-4 border-t border-line pt-3 space-y-1">
      <UiButton v-if="mode === 'pin'" variant="link" @click="useOneTimePassword">
        Signing in with a one-time password?
      </UiButton>
      <UiButton v-else variant="link" @click="usePassword">Use my password instead</UiButton>
      <div>
        <UiButton variant="link" @click="navigateTo('/sign-in/admin')">
          System admin sign in
        </UiButton>
      </div>
    </div>
  </UiPanel>
</template>
