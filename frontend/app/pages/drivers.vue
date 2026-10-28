<script setup lang="ts">
import type { Driver } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Drivers' })
definePageMeta({ layout: 'default', roles: ['owner', 'system_admin'] })

const { token } = useAuth()
const { branches, options: branchOptions, name: branchName, load: loadBranches } = useBranches()

const drivers = ref<Driver[]>([])
const loading = ref(true)
const error = ref('')
const notice = ref('')
const branchFilter = ref('')
const search = ref('')

const formOpen = ref(false)
const editingId = ref('')
const form = reactive({
  branch: '',
  fullName: '',
  phone: '',
  licenseNumber: '',
  licenseExpiry: '',
  status: 'available',
})
const saving = ref(false)
const formError = ref('')

const confirmTarget = ref<Driver | null>(null)
const busy = ref(false)

const statusOptions = [
  { value: 'available', label: 'Available' },
  { value: 'on_trip', label: 'On trip' },
  { value: 'inactive', label: 'Inactive' },
]

const filtered = computed(() => {
  const term = search.value.trim().toLowerCase()
  return drivers.value.filter((driver) => {
    if (branchFilter.value && driver.branch_id !== branchFilter.value) return false
    if (!term) return true
    return [driver.full_name, driver.phone, driver.license_number]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(term))
  })
})

function statusTone(status: string) {
  if (status === 'available') return 'info' as const
  if (status === 'on_trip') return 'neutral' as const
  return 'neutral' as const
}

function expiryTone(driver: Driver): string {
  if (!driver.license_expiry) return 'text-ink-64'
  const expiry = new Date(`${driver.license_expiry.slice(0, 10)}T00:00:00`)
  if (Number.isNaN(expiry.getTime())) return 'text-ink-64'
  return expiry.getTime() < Date.now() ? 'text-danger' : 'text-ink-64'
}

async function load() {
  if (!token.value) return
  loading.value = true
  error.value = ''
  try {
    drivers.value = await apiRequest<Driver[]>('/api/drivers', { token: token.value })
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not load drivers.'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = ''
  form.branch = branchFilter.value || branches.value[0]?.id || ''
  form.fullName = ''
  form.phone = ''
  form.licenseNumber = ''
  form.licenseExpiry = ''
  form.status = 'available'
  formError.value = ''
  formOpen.value = true
}

function openEdit(driver: Driver) {
  editingId.value = driver.id
  form.branch = driver.branch_id
  form.fullName = driver.full_name
  form.phone = driver.phone
  form.licenseNumber = driver.license_number ?? ''
  form.licenseExpiry = driver.license_expiry ? driver.license_expiry.slice(0, 10) : ''
  form.status = driver.status
  formError.value = ''
  formOpen.value = true
}

async function save() {
  if (!token.value) return
  formError.value = ''
  if (!form.branch) {
    formError.value = 'Choose the branch this driver belongs to.'
    return
  }
  if (!form.fullName.trim()) {
    formError.value = 'Enter the driver name.'
    return
  }
  if (!editingId.value && !form.phone.trim()) {
    formError.value = 'A phone number is required.'
    return
  }

  saving.value = true
  try {
    if (editingId.value) {
      await apiRequest(`/api/drivers/${editingId.value}`, {
        method: 'PUT',
        token: token.value,
        body: {
          branch_id: form.branch,
          full_name: form.fullName.trim(),
          phone: form.phone.trim(),
          license_number: form.licenseNumber.trim(),
          license_expiry: form.licenseExpiry || undefined,
          status: form.status,
        },
      })
      notice.value = 'Driver updated.'
    } else {
      await apiRequest('/api/drivers', {
        method: 'POST',
        token: token.value,
        body: {
          branch_id: form.branch,
          full_name: form.fullName.trim(),
          phone: form.phone.trim(),
          license_number: form.licenseNumber.trim() || undefined,
          license_expiry: form.licenseExpiry || undefined,
          status: form.status,
        },
      })
      notice.value = 'Driver added.'
    }
    formOpen.value = false
    await load()
  } catch (err) {
    formError.value = err instanceof ApiError ? err.message : 'Could not save the driver.'
  } finally {
    saving.value = false
  }
}

async function deactivate() {
  if (!token.value || !confirmTarget.value) return
  busy.value = true
  try {
    await apiRequest(`/api/drivers/${confirmTarget.value.id}`, { method: 'DELETE', token: token.value })
    notice.value = `${confirmTarget.value.full_name} is now inactive.`
    confirmTarget.value = null
    await load()
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not deactivate the driver.'
  } finally {
    busy.value = false
  }
}

onMounted(async () => {
  await loadBranches()
  await load()
})
</script>

<template>
  <div class="space-y-4">
    <header class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-h5 font-medium tracking-tight text-ink">Drivers</h1>
        <p class="mt-1 text-label text-ink-64">Who drives between branches, and their licences.</p>
      </div>
      <UiButton @click="openCreate">
        <UiIcon name="plus" :size="17" />
        Add driver
      </UiButton>
    </header>

    <div v-if="notice" class="rounded-lg border border-blue/25 bg-blue/5 px-4 py-3 text-label text-ink" role="status">{{ notice }}</div>
    <div v-if="error" class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-danger/25 bg-danger/5 px-4 py-3">
      <p class="text-label text-danger">{{ error }}</p>
      <UiButton variant="tertiary" :loading="loading" @click="load">Try again</UiButton>
    </div>

    <div class="flex flex-wrap gap-2">
      <div class="min-w-[200px] flex-1">
        <UiSearchBar v-model="search" placeholder="Name, phone or licence" />
      </div>
      <UiSelect
        id="driver-branch"
        v-model="branchFilter"
        label="Branch filter"
        hide-label
        :options="[{ value: '', label: 'All branches' }, ...branchOptions]"
        class="w-[210px]"
      />
    </div>

    <div v-if="loading" class="grid gap-3 sm:grid-cols-2" aria-hidden="true">
      <div v-for="index in 4" :key="index" class="h-[124px] animate-pulse rounded-lg bg-surface shadow-card" />
    </div>

    <template v-else>
      <div v-if="filtered.length" class="grid gap-3 sm:grid-cols-2">
        <article v-for="driver in filtered" :key="driver.id" class="rounded-lg bg-surface p-4 shadow-card">
          <div class="flex items-start justify-between gap-3">
            <div class="flex min-w-0 items-center gap-3">
              <span class="flex size-9 shrink-0 items-center justify-center rounded-full bg-canvas text-label font-medium text-ink">
                {{ driver.full_name.split(' ').slice(0, 2).map((part) => part[0]).join('').toUpperCase() }}
              </span>
              <div class="min-w-0">
                <h2 class="truncate text-p1 font-medium text-ink">{{ driver.full_name }}</h2>
                <p class="truncate text-meta text-ink-40">{{ branchName(driver.branch_id) || '—' }}</p>
              </div>
            </div>
            <UiPill :tone="statusTone(driver.status)">{{ humanise(driver.status) }}</UiPill>
          </div>

          <dl class="mt-3 space-y-1 text-label">
            <div class="flex gap-2">
              <dt class="w-20 shrink-0 text-ink-40">Phone</dt>
              <dd class="min-w-0 text-ink">{{ driver.phone }}</dd>
            </div>
            <div class="flex gap-2">
              <dt class="w-20 shrink-0 text-ink-40">Licence</dt>
              <dd class="min-w-0 text-ink">{{ driver.license_number || '—' }}</dd>
            </div>
            <div class="flex gap-2">
              <dt class="w-20 shrink-0 text-ink-40">Expires</dt>
              <dd class="min-w-0" :class="expiryTone(driver)">{{ shortDate(driver.license_expiry) }}</dd>
            </div>
          </dl>

          <div class="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 border-t border-line pt-3">
            <UiButton variant="link" @click="openEdit(driver)">Edit</UiButton>
            <UiButton v-if="driver.status !== 'inactive'" variant="link" class="text-danger" @click="confirmTarget = driver">
              Deactivate
            </UiButton>
          </div>
        </article>
      </div>

      <UiEmptyState v-else icon="user" title="No drivers" description="Add the drivers who move parcels between branches.">
        <UiButton @click="openCreate">Add driver</UiButton>
      </UiEmptyState>
    </template>

    <!-- Create / edit -->
    <UiModal
      :open="formOpen"
      :title="editingId ? 'Edit driver' : 'Add driver'"
      size="md"
      :dismissible="!saving"
      @close="formOpen = false"
    >
      <div v-if="formError" class="mb-3 rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
        {{ formError }}
      </div>
      <div class="space-y-4">
        <UiSelect id="driver-branch-form" v-model="form.branch" label="Branch" :options="branchOptions" />
        <div class="grid gap-4 sm:grid-cols-2">
          <UiField id="driver-name" v-model="form.fullName" label="Full name" autocomplete="off" />
          <UiField id="driver-phone" v-model="form.phone" label="Phone" type="tel" autocomplete="off" />
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <UiField id="driver-licence" v-model="form.licenseNumber" label="Licence number (optional)" autocomplete="off" />
          <UiField id="driver-expiry" v-model="form.licenseExpiry" label="Licence expiry (optional)" type="date" />
        </div>
        <UiSelect v-if="editingId" id="driver-status-form" v-model="form.status" label="Status" :options="statusOptions" />
      </div>
      <template #footer>
        <div class="flex flex-col gap-2 sm:flex-row-reverse">
          <UiButton block :loading="saving" @click="save">{{ editingId ? 'Save changes' : 'Add driver' }}</UiButton>
          <UiButton variant="tertiary" block :disabled="saving" @click="formOpen = false">Cancel</UiButton>
        </div>
      </template>
    </UiModal>

    <!-- Deactivate -->
    <UiModal
      :open="Boolean(confirmTarget)"
      title="Deactivate driver"
      description="Their history stays; they just cannot be assigned new trips."
      :dismissible="!busy"
      @close="confirmTarget = null"
    >
      <p class="text-label text-ink-64">
        Deactivate <span class="font-medium text-ink">{{ confirmTarget?.full_name }}</span>?
      </p>
      <template #footer>
        <div class="flex flex-col gap-2 sm:flex-row-reverse">
          <UiButton variant="danger" block :loading="busy" @click="deactivate">Deactivate</UiButton>
          <UiButton variant="tertiary" block :disabled="busy" @click="confirmTarget = null">Keep active</UiButton>
        </div>
      </template>
    </UiModal>
  </div>
</template>
