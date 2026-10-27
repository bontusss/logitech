<script setup lang="ts">
import { ApiError } from '~/utils/api'

const { user, token, logout } = useAuth()

useHead({ title: 'Settings' })
definePageMeta({ layout: 'default' })

const signingOut = ref(false)
const currentPin = ref('')
const newPin = ref('')
const confirmPin = ref('')
const changingPin = ref(false)
const pinError = ref('')
const pinSuccess = ref('')
const branchName = ref('')

const managementLinks = [
  { title: 'Staff', description: 'Create and manage staff accounts.', to: '/staff', icon: 'staff' },
  { title: 'Vehicles', description: 'Manage fleet vehicles.', to: '/vehicles', icon: 'vehicles' },
  { title: 'Drivers', description: 'Manage drivers and licenses.', to: '/drivers', icon: 'drivers' },
]

const roleLabel = computed(() => {
  if (!user.value) return ''
  return user.value.job_role
    ? `${user.value.role} · ${user.value.job_role}`
    : user.value.role
})

onMounted(async () => {
  if (!user.value?.branch_id || !token.value) return
  try {
    const branches = await apiRequest<Array<{ id: string; name: string; code: string }>>('/api/branches', { token: token.value })
    const branch = branches.find((item) => item.id === user.value?.branch_id)
    branchName.value = branch ? `${branch.name} (${branch.code})` : ''
  } catch {
    branchName.value = ''
  }
})

async function changePin() {
  pinError.value = ''
  pinSuccess.value = ''
  if (newPin.value.length < 4 || newPin.value.length > 10) {
    pinError.value = 'New PIN must be 4 to 10 characters.'
    return
  }
  if (newPin.value !== confirmPin.value) {
    pinError.value = 'New PINs do not match.'
    return
  }
  if (newPin.value === currentPin.value) {
    pinError.value = 'New PIN must differ from the current PIN.'
    return
  }

  changingPin.value = true
  try {
    await apiRequest<{ message: string }>('/api/auth/change-pin', {
      method: 'POST',
      token: token.value,
      body: { current_pin: currentPin.value, new_pin: newPin.value },
    })
    currentPin.value = ''
    newPin.value = ''
    confirmPin.value = ''
    pinSuccess.value = 'Your PIN has been updated.'
  } catch (error) {
    pinError.value = error instanceof ApiError ? error.message : 'Could not update your PIN. Try again.'
  } finally {
    changingPin.value = false
  }
}

async function signOut() {
  signingOut.value = true
  try {
    await logout()
    await navigateTo('/sign-in')
  } finally {
    signingOut.value = false
  }
}
</script>

<template>
  <div v-if="user">
    <div class="min-h-dvh bg-canvas md:hidden">
      <AppNavbar :user-name="user.full_name" :signing-out="signingOut" @logout="signOut" />

      <main class="mx-auto w-full max-w-[480px] space-y-6 px-4 pb-10 pt-5">
        <NuxtLink to="/" class="inline-flex items-center gap-2 text-label text-ink-64 transition hover:text-blue">
          <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="m15 18-6-6 6-6"/>
          </svg>
          Dashboard
        </NuxtLink>

        <header>
          <h1 class="text-h3 text-ink">Settings</h1>
        </header>

        <section class="border-b border-line pb-5">
          <h2 class="text-h5 text-ink">Profile</h2>
          <dl class="mt-3 divide-y divide-line">
            <div class="flex justify-between gap-4 py-3">
              <dt class="text-label text-ink-64">Name</dt>
              <dd class="text-right text-label font-medium text-ink">{{ user.full_name }}</dd>
            </div>
            <div class="flex justify-between gap-4 py-3">
              <dt class="text-label text-ink-64">Staff ID</dt>
              <dd class="text-right text-label font-medium text-ink">{{ user.staff_id }}</dd>
            </div>
            <div class="flex justify-between gap-4 py-3">
              <dt class="text-label text-ink-64">Role</dt>
              <dd class="text-right text-label font-medium capitalize text-ink">{{ roleLabel }}</dd>
            </div>
            <div v-if="branchName" class="flex justify-between gap-4 py-3">
              <dt class="text-label text-ink-64">Branch</dt>
              <dd class="text-right text-label font-medium text-ink">{{ branchName }}</dd>
            </div>
            <div v-if="user.email" class="flex justify-between gap-4 py-3">
              <dt class="text-label text-ink-64">Email</dt>
              <dd class="max-w-[65%] break-all text-right text-label font-medium text-ink">{{ user.email }}</dd>
            </div>
            <div v-if="user.phone" class="flex justify-between gap-4 py-3">
              <dt class="text-label text-ink-64">Phone</dt>
              <dd class="text-right text-label font-medium text-ink">{{ user.phone }}</dd>
            </div>
          </dl>
        </section>

        <section v-if="user.role !== 'system_admin'" class="border-b border-line pb-6">
          <div>
            <h2 class="text-h5 text-ink">Change PIN</h2>
            <p class="mt-1 text-label text-ink-64">Use 4 to 10 characters. Your current PIN is required.</p>
          </div>
          <form class="mt-4 space-y-4" @submit.prevent="changePin">
            <UiField
              id="current-pin"
              v-model="currentPin"
              label="Current PIN"
              type="password"
              autocomplete="current-password"
              :maxlength="10"
              :error="''"
            />
            <UiField
              id="new-pin"
              v-model="newPin"
              label="New PIN"
              type="password"
              autocomplete="new-password"
              :maxlength="10"
              :error="''"
            />
            <UiField
              id="confirm-pin"
              v-model="confirmPin"
              label="Confirm new PIN"
              type="password"
              autocomplete="new-password"
              :maxlength="10"
              :error="''"
            />
            <p v-if="pinError" class="text-label text-danger" role="alert">{{ pinError }}</p>
            <p v-if="pinSuccess" class="text-label text-blue" role="status">{{ pinSuccess }}</p>
            <UiButton type="submit" block :loading="changingPin">
              Update PIN
            </UiButton>
          </form>
        </section>

        <section v-if="user.role === 'manager'">
          <h2 class="text-h5 text-ink">Management</h2>
          <nav class="mt-2 divide-y divide-line" aria-label="Management">
            <NuxtLink
              v-for="item in managementLinks"
              :key="item.to"
              :to="item.to"
              class="flex items-center gap-3 py-4 text-ink transition hover:text-blue"
            >
              <span class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-surface text-blue" aria-hidden="true">
                <svg v-if="item.icon === 'staff'" class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M16 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                  <circle cx="10" cy="7" r="4"/>
                  <path d="M20 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/>
                </svg>
                <svg v-else-if="item.icon === 'vehicles'" class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M3 7h11v10H3zM14 10h4l3 3v4h-7z"/>
                  <circle cx="7.5" cy="18" r="1.5"/>
                  <circle cx="18" cy="18" r="1.5"/>
                </svg>
                <svg v-else class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="7" r="4"/>
                  <path d="M5.5 21a6.5 6.5 0 0 1 13 0M19 8h3v7h-3"/>
                </svg>
              </span>
              <span class="min-w-0 flex-1">
                <span class="block text-p1 font-medium">{{ item.title }}</span>
                <span class="mt-0.5 block text-label text-ink-64">{{ item.description }}</span>
              </span>
              <svg class="size-4 shrink-0 text-ink-40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="m9 18 6-6-6-6"/>
              </svg>
            </NuxtLink>
          </nav>
        </section>
      </main>
    </div>

    <div class="hidden min-h-dvh items-center justify-center bg-canvas px-6 text-center md:flex">
      <p class="text-label text-ink-64">Galopin is available on mobile screens.</p>
    </div>
  </div>
</template>
