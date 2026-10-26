<script setup lang="ts">
import { ApiError } from '~/utils/api'

definePageMeta({ layout: 'auth', middleware: 'first-login' })

useHead({ title: 'Welcome' })

const { login } = useAuth()
const pending = useFirstLogin()

const step = ref<'password' | 'confirm'>('password')
const password = ref('')
const confirmation = ref('')
const error = ref('')
const failing = ref<'password' | 'confirmation' | null>(null)
const submitting = ref(false)

/** 4–10 characters. */
function validPin(value: string) {
  return value.length >= 4 && value.length <= 10
}

function fail(message: string, field: 'password' | 'confirmation') {
  error.value = message
  failing.value = field
}

function next() {
  error.value = ''
  failing.value = null

  if (!validPin(password.value)) {
    fail('Use 4–10 characters.', 'password')
    return
  }

  step.value = 'confirm'
  confirmation.value = ''
}

function back() {
  step.value = 'password'
  confirmation.value = ''
  error.value = ''
  failing.value = null
}

async function save() {
  if (submitting.value || !pending.value) return

  error.value = ''
  failing.value = null

  if (confirmation.value !== password.value) {
    fail("The two passwords don't match. Type it again.", 'confirmation')
    return
  }

  if (!pending.value) return
  const staffId = pending.value.staffId
  const oneTimePassword = pending.value.oneTimePassword

  submitting.value = true
  try {
    await login({
      staffId,
      pin: oneTimePassword,
      newPin: password.value,
    })
    pending.value = null
    await navigateTo('/')
  } catch (err) {
    if (err instanceof ApiError && err.isLocked && pending.value) {
      await navigateTo({
        path: '/locked',
        query: { staff_id: staffId, at: clockTime() },
      })
      return
    }

    if (err instanceof ApiError && err.status === 401) {
      fail("That one-time password didn't work. Ask for a new one.", 'password')
      step.value = 'password'
      return
    }

    if (err instanceof ApiError && (err.status === 400 || err.status === 403)) {
      fail(err.message, 'password')
      step.value = 'password'
      return
    }

    fail("Can't reach the server. Check your connection and try again.", 'confirmation')
  } finally {
    submitting.value = false
  }
}

function onEscape() {
  if (step.value === 'confirm') back()
}

/** Leave the welcome flow, e.g. after a rejected one-time password. */
function abandon() {
  pending.value = null
  navigateTo('/sign-in')
}
</script>

<template>
  <UiPanel v-if="pending" title="Welcome" @keydown.esc="onEscape">
    <form v-if="step === 'password'" novalidate class="space-y-4" @submit.prevent="next">
      <UiField
        id="password"
        v-model="password"
        label="New password"
        type="password"
        autocomplete="new-password"
        hint="4–10 characters. Do not share it."
        :error="failing === 'password' ? error : ''"
        autofocus
      />

      <div class="flex items-center gap-2">
        <UiButton type="submit">
          Next
          <UiKbd>Enter</UiKbd>
        </UiButton>
        <UiButton variant="link" @click="abandon">Back to sign in</UiButton>
      </div>
    </form>

    <form v-else novalidate class="space-y-4" @submit.prevent="save">
      <UiField
        id="confirmation"
        v-model="confirmation"
        label="Type the password again"
        type="password"
        autocomplete="new-password"
        :error="failing === 'confirmation' ? error : ''"
        autofocus
      />

      <div class="flex items-center gap-2">
        <UiButton type="submit" :loading="submitting">
          Save password
          <UiKbd>Enter</UiKbd>
        </UiButton>
        <UiButton variant="tertiary" @click="back">
          Back
          <UiKbd>Esc</UiKbd>
        </UiButton>
      </div>
    </form>
  </UiPanel>
</template>
