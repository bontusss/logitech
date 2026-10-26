<script setup lang="ts">
import { ApiError, apiRequest } from '~/utils/api'
import type { AuthResponse } from '~/composables/useAuth'

definePageMeta({ layout: 'auth' })

useHead({ title: 'System Admin Sign in' })

const route = useRoute()
const { token, user } = useAuth()

const username = ref('')
const password = ref('')
const error = ref('')
const failing = ref<'username' | 'password' | null>(null)
const submitting = ref(false)

function fail(message: string, field: 'username' | 'password') {
  error.value = message
  failing.value = field
}

async function submit() {
  if (submitting.value) return

  error.value = ''
  failing.value = null

  if (!username.value.trim()) {
    fail('Enter your admin username or email.', 'username')
    return
  }
  if (!password.value) {
    fail('Enter your admin password.', 'password')
    return
  }

  submitting.value = true
  try {
    const response = await apiRequest<AuthResponse>('/api/auth/admin/login', {
      method: 'POST',
      body: {
        username: username.value.trim(),
        password: password.value,
      },
    })

    token.value = response.token
    user.value = response.user
    await navigateTo(typeof route.query.redirect === 'string' ? route.query.redirect : '/')
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      fail('Those credentials are wrong.', 'password')
      return
    }
    if (err instanceof ApiError && err.status === 400) {
      fail(err.message, 'username')
      return
    }
    fail("Can't reach the server. Check your connection and try again.", 'password')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <UiPanel title="System Admin Sign in">
    <form novalidate class="space-y-4" @submit.prevent="submit">
      <UiField
        id="admin-username"
        v-model="username"
        label="Username or email"
        autocomplete="username"
        :error="failing === 'username' ? error : ''"
        autofocus
      />

      <UiField
        id="admin-password"
        v-model="password"
        label="Password"
        type="password"
        autocomplete="current-password"
        :error="failing === 'password' ? error : ''"
      />

      <UiButton type="submit" :loading="submitting">
        Sign in
        <UiKbd>Enter</UiKbd>
      </UiButton>
    </form>

    <div class="mt-4 border-t border-line pt-3">
      <UiButton variant="link" @click="navigateTo('/sign-in')">
        Staff sign in
      </UiButton>
    </div>
  </UiPanel>
</template>
