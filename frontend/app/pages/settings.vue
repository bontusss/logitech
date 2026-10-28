<script setup lang="ts">
import { ApiError } from '~/utils/api'

useHead({ title: 'Settings' })
definePageMeta({ layout: 'default' })

const { user, token, logout } = useAuth()
const {
  isOwner,
  isManager,
  isSystemAdmin,
  canViewReports,
  canManageStaff,
  canManageBranches,
  canManageFleet,
} = usePermissions()
const { name: branchName, code: branchCode, load: loadBranches } = useBranches()

const currentPin = ref('')
const newPin = ref('')
const confirmPin = ref('')
const changingPin = ref(false)
const pinError = ref('')
const pinSuccess = ref('')

const signingOut = ref(false)

const roleLabel = computed(() => {
  if (!user.value) return ''
  if (user.value.role === 'system_admin') return 'System admin'
  const job = user.value.job_role ? ` · ${humanise(user.value.job_role)}` : ''
  return `${humanise(user.value.role)}${job}`
})

const branch = computed(() => {
  const name = branchName(user.value?.branch_id ?? '')
  const code = branchCode(user.value?.branch_id ?? '')
  return name ? `${name}${code ? ` (${code})` : ''}` : ''
})

const links = computed(() => {
  const items: Array<{ to: string; label: string; description: string; icon: string }> = []
  if (canViewReports.value) {
    items.push({ to: '/reports', label: 'Reports', description: 'Shipments, revenue and collections.', icon: 'chart' })
  }
  if (canManageStaff.value) {
    items.push({ to: '/staff', label: 'Staff', description: 'Accounts, PIN resets and access.', icon: 'users' })
  }
  if (canManageBranches.value) {
    items.push({ to: '/branches', label: 'Branches', description: 'Where parcels are handled.', icon: 'building' })
  }
  if (canManageFleet.value) {
    items.push({ to: '/vehicles', label: 'Vehicles', description: 'The fleet that moves parcels.', icon: 'truck' })
    items.push({ to: '/drivers', label: 'Drivers', description: 'Licences and availability.', icon: 'user' })
  }
  return items
})

async function changePin() {
  pinError.value = ''
  pinSuccess.value = ''
  if (newPin.value.length < 4 || newPin.value.length > 10) {
    pinError.value = 'The new PIN must be 4 to 10 characters.'
    return
  }
  if (newPin.value !== confirmPin.value) {
    pinError.value = 'The new PINs do not match.'
    return
  }
  if (newPin.value === currentPin.value) {
    pinError.value = 'The new PIN must differ from the current one.'
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
  } catch (err) {
    pinError.value = err instanceof ApiError ? err.message : 'Could not update your PIN.'
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

onMounted(loadBranches)
</script>

<template>
  <div v-if="user" class="space-y-5">
    <header>
      <h1 class="text-h5 font-medium tracking-tight text-ink">Settings</h1>
      <p class="mt-1 text-label text-ink-64">Your account and the parts of the business you run.</p>
    </header>

    <section class="rounded-lg bg-surface p-4 shadow-card">
      <div class="flex items-center gap-3">
        <span class="flex size-12 shrink-0 items-center justify-center rounded-full bg-canvas text-p1 font-medium text-ink">
          {{ user.full_name.split(' ').slice(0, 2).map((part) => part[0]).join('').toUpperCase() }}
        </span>
        <div class="min-w-0">
          <p class="truncate text-p1 font-medium text-ink">{{ user.full_name }}</p>
          <p class="truncate text-label text-ink-64">{{ roleLabel }}</p>
        </div>
      </div>

      <dl class="mt-4 divide-y divide-line">
        <div class="flex items-start justify-between gap-3 py-2.5">
          <dt class="text-label text-ink-64">Staff ID</dt>
          <dd class="text-right text-label font-medium text-ink">{{ user.staff_id }}</dd>
        </div>
        <div v-if="branch" class="flex items-start justify-between gap-3 py-2.5">
          <dt class="text-label text-ink-64">Branch</dt>
          <dd class="text-right text-label font-medium text-ink">{{ branch }}</dd>
        </div>
        <div class="flex items-start justify-between gap-3 py-2.5">
          <dt class="text-label text-ink-64">Phone</dt>
          <dd class="text-right text-label font-medium text-ink">{{ user.phone || '—' }}</dd>
        </div>
        <div class="flex items-start justify-between gap-3 py-2.5">
          <dt class="text-label text-ink-64">Email</dt>
          <dd class="max-w-[60%] break-all text-right text-label font-medium text-ink">{{ user.email || '—' }}</dd>
        </div>
        <div class="flex items-start justify-between gap-3 py-2.5">
          <dt class="text-label text-ink-64">Status</dt>
          <dd class="flex items-center gap-2">
            <UiPill :tone="user.status === 'active' ? 'info' : 'neutral'">{{ humanise(user.status) }}</UiPill>
            <UiPill v-if="user.locked" tone="danger">Locked</UiPill>
          </dd>
        </div>
      </dl>
    </section>

    <section v-if="!isSystemAdmin" class="rounded-lg bg-surface p-4 shadow-card">
      <h2 class="text-h5 font-medium text-ink">Change PIN</h2>
      <p class="mt-0.5 text-label text-ink-64">Use 4 to 10 characters. Your current PIN is required.</p>

      <form class="mt-4 max-w-[420px] space-y-4" novalidate @submit.prevent="changePin">
        <UiField
          id="current-pin"
          v-model="currentPin"
          label="Current PIN"
          type="password"
          autocomplete="current-password"
          :maxlength="10"
        />
        <UiField
          id="new-pin"
          v-model="newPin"
          label="New PIN"
          type="password"
          autocomplete="new-password"
          :maxlength="10"
        />
        <UiField
          id="confirm-pin"
          v-model="confirmPin"
          label="Confirm new PIN"
          type="password"
          autocomplete="new-password"
          :maxlength="10"
        />
        <p v-if="pinError" class="text-label text-danger" role="alert">{{ pinError }}</p>
        <p v-if="pinSuccess" class="text-label text-blue" role="status">{{ pinSuccess }}</p>
        <UiButton type="submit" :loading="changingPin">Update PIN</UiButton>
      </form>
    </section>

    <section v-if="links.length" class="rounded-lg bg-surface p-4 shadow-card">
      <h2 class="text-h5 font-medium text-ink">Manage</h2>
      <ul class="mt-2 divide-y divide-line">
        <li v-for="item in links" :key="item.to">
          <NuxtLink :to="item.to" class="flex items-center gap-3 py-3 text-ink transition hover:text-blue">
            <span class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-canvas text-blue">
              <UiIcon :name="item.icon" :size="17" />
            </span>
            <span class="min-w-0 flex-1">
              <span class="block text-p1 font-medium">{{ item.label }}</span>
              <span class="block text-meta text-ink-64">{{ item.description }}</span>
            </span>
            <UiIcon name="chevronRight" :size="16" class="text-ink-40" />
          </NuxtLink>
        </li>
      </ul>
    </section>

    <section class="rounded-lg bg-surface p-4 shadow-card">
      <h2 class="text-h5 font-medium text-ink">Session</h2>
      <p class="mt-0.5 text-label text-ink-64">
        Signing out on this device ends the session{{ isOwner || isManager ? ', but your branch keeps working' : '' }}.
      </p>
      <UiButton variant="tertiary" class="mt-3" :loading="signingOut" @click="signOut">
        <UiIcon name="logout" :size="17" />
        Sign out
      </UiButton>
    </section>
  </div>
</template>
