<script setup lang="ts">
import { ApiError } from '~/utils/api'
import type { BranchOptions } from '~/types/staff'
import UiSelect from '~/components/ui/Select.vue'

definePageMeta({ layout: 'default' })

useHead({ title: 'Drivers' })

const { user, token } = useAuth()
const route = useRoute()

// ---- data sources ----
const drivers = ref<Array<{
  id: string
  branch_id: string
  full_name: string
  phone: string
  license_number?: string
  license_expiry?: string
  status: string
}>>([])
const branches = ref<Array<{ id: string; name: string }>>([])
const loading = ref(false)

// ---- add / edit modal state ----
const editing = ref<typeof drivers.value[0] | null>(null)
const adding = ref(false)
const saving = ref(false)

const addFullName = ref('')
const addPhone = ref('')
const addLicenseNumber = ref('')
const addLicenseExpiry = ref('')
const addBranch = ref('')

// ---- menu state ----
const menuFor = ref<string | null>(null)

// ---- confirm modal ----
const confirmModal = ref<{ mode: 'delete'; driverId: string; driverName: string } | null>(null)

// ---- driver type options ----
const driverStatusOptions = [
  { value: '', label: 'Select status' },
  { value: 'available', label: 'Available' },
  { value: 'on_trip', label: 'On trip' },
  { value: 'inactive', label: 'Inactive' },
]

const branchOptions = computed<BranchOptions[]>(() => [
  { value: '', label: 'Select branch' },
  ...branches.value.map((b) => ({ value: b.id, label: b.name })),
])

// ---- refresh loop ----
onMounted(refresh)
watch(user, () => { if (user.value) refresh() })
watch(token, () => { if (token.value) refresh() })

async function refresh() {
  if (!token.value) return
  loading.value = true
  try {
    const [d, b] = await Promise.all([fetchDrivers(), fetchBranchOptions()])
    drivers.value = d
    branches.value = b
  } finally {
    loading.value = false
  }
}

async function fetchDrivers(): Promise<typeof drivers.value> {
  const rows = await apiRequest<typeof drivers.value>('/api/drivers', { token: token.value })
  return rows.map(normalizeDriver)
}

async function fetchBranchOptions(): Promise<Array<{ id: string; name: string }>> {
  const rows = await apiRequest<Array<{ id: string; name: string; code: string }>>('/api/branches', { token: token.value })
  return rows.map((b) => ({ id: b.id, name: `${b.name} (${b.code})` }))
}

function normalizeDriver(row: typeof drivers.value[0]): typeof drivers.value[0] {
  return {
    ...row,
    license_number: row.license_number ?? undefined,
    license_expiry: row.license_expiry ?? undefined,
  }
}

// ---- add driver ----
function openAdd() {
  editing.value = null
  adding.value = true
  addFullName.value = ''
  addPhone.value = ''
  addLicenseNumber.value = ''
  addLicenseExpiry.value = ''
  addBranch.value = branches.value[0]?.id ?? ''
}

async function saveAdd() {
  if (!addFullName.value.trim() || !addPhone.value.trim() || !addBranch.value) return
  saving.value = true
  try {
    await apiRequest('/api/drivers', {
      method: 'POST',
      token: token.value,
      body: {
        full_name: addFullName.value.trim(),
        phone: addPhone.value.trim(),
        license_number: addLicenseNumber.value.trim(),
        license_expiry: addLicenseExpiry.value.trim(),
        branch_id: addBranch.value,
      },
    })
    adding.value = false
    await refresh()
  } catch (err) {
    if (err instanceof ApiError) {
      console.error(err.message)
    }
  } finally {
    saving.value = false
  }
}

// ---- edit driver ----
function openEdit(row: typeof drivers.value[0]) {
  editing.value = row
  addFullName.value = row.full_name
  addPhone.value = row.phone
  addLicenseNumber.value = row.license_number ?? ''
  addLicenseExpiry.value = row.license_expiry ?? ''
  addBranch.value = row.branch_id
  adding.value = false
}

async function saveEdit() {
  if (!editing.value) return
  saving.value = true
  try {
    await apiRequest(`/api/drivers/${editing.value.id}`, {
      method: 'PUT',
      token: token.value,
      body: {
        full_name: addFullName.value.trim(),
        phone: addPhone.value.trim(),
        license_number: addLicenseNumber.value.trim(),
        license_expiry: addLicenseExpiry.value.trim(),
        branch_id: addBranch.value,
        status: editing.value.status,
      },
    })
    editing.value = null
    await refresh()
  } catch (err) {
    if (err instanceof ApiError) {
      console.error(err.message)
    }
  } finally {
    saving.value = false
  }
}

function closeEditor() {
  editing.value = null
  adding.value = false
}

// ---- row menu ----
function openMenu(id: string) {
  menuFor.value = menuFor.value === id ? null : id
}

// ---- delete action ----
async function confirmDelete(row: typeof drivers.value[0]) {
  menuFor.value = null
  confirmModal.value = {
    mode: 'delete',
    driverId: row.id,
    driverName: row.full_name,
  }
}

async function executeDelete() {
  if (!confirmModal.value) return
  saving.value = true
  try {
    await apiRequest(`/api/drivers/${confirmModal.value.driverId}`, {
      method: 'DELETE',
      token: token.value,
    })
    confirmModal.value = null
    await refresh()
  } catch (err) {
    if (err instanceof ApiError) {
      console.error(err.message)
    }
  } finally {
    saving.value = false
  }
}

function closeConfirm() {
  confirmModal.value = null
}

// ---- status badge ----
function statusBadge(status: string): { label: string; tone: 'danger' | 'info' | 'neutral' } {
  return {
    available: { label: 'Available', tone: 'info' as const },
    on_trip: { label: 'On trip', tone: 'neutral' as const },
    inactive: { label: 'Inactive', tone: 'neutral' as const },
  }[status] ?? { label: status, tone: 'neutral' as const }
}

// ---- escape handler ----
function onEscape() {
  if (menuFor.value) menuFor.value = null
  else if (editing.value || adding.value) closeEditor()
  else if (confirmModal.value) closeConfirm()
}
</script>

<template>
  <div class="min-h-dvh" @keydown.esc="onEscape">
    <!-- top tabs -->
    <div class="border-b border-line bg-canvas">
      <div class="mx-auto flex max-w-[720px] border-b border-line px-4">
        <nav class="flex gap-1">
          <NuxtLink
            to="/staff"
            class="-(m-1) flex h-11 flex-1 items-center justify-center rounded-lg px-3 text-button font-medium transition duration-150 hover:bg-canvas/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
            :class="$route.path === '/staff' ? 'bg-surface shadow-raised text-ink' : 'text-ink-64'"
          >
            Staff
          </NuxtLink>
          <NuxtLink
            to="/drivers"
            class="-(m-1) flex h-11 flex-1 items-center justify-center rounded-lg px-3 text-button font-medium transition duration-150 last:rounded-tr-lg hover:bg-canvas/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
            :class="$route.path === '/drivers' ? 'bg-surface shadow-raised text-ink' : 'text-ink-64'"
          >
            Drivers
          </NuxtLink>
          <NuxtLink
            to="/vehicles"
            class="-(m-1) flex h-11 flex-1 items-center justify-center rounded-lg px-3 text-button font-medium transition duration-150 hover:bg-canvas/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
            :class="$route.path === '/vehicles' ? 'bg-surface shadow-raised text-ink' : 'text-ink-64'"
          >
            Vehicles
          </NuxtLink>
          <NuxtLink
            to="/branches"
            class="-(m-1) flex h-11 flex-1 items-center justify-center rounded-lg px-3 text-button font-medium transition duration-150 hover:bg-canvas/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
            :class="$route.path === '/branches' ? 'bg-surface shadow-raised text-ink' : 'text-ink-64'"
          >
            Branches
          </NuxtLink>
        </nav>
      </div>
    </div>

    <!-- main content -->
    <main class="mx-auto max-w-[720px] px-4 pb-14 pt-6">
      <div class="space-y-4">
        <!-- header -->
        <div class="flex items-start justify-between gap-4">
          <div>
            <h1 class="text-h3 text-ink">Drivers</h1>
            <p class="mt-1 text-label text-ink-64">
              Manage drivers and their license information.
            </p>
          </div>
          <UiButton
            type="button"
            :disabled="!token"
            @click="openAdd"
          >
            <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" aria-hidden="true">
              <path d="M12 5v14M5 12h14"/>
            </svg>
            Add driver
            <UiKbd>Enter</UiKbd>
          </UiButton>
        </div>

        <!-- drivers list -->
        <div class="divide-y divide-line">
          <div
            v-for="row in drivers"
            :key="row.id"
            class="group relative flex items-start gap-3 py-3 first:pt-0 last:pb-0"
          >
            <div class="min-w-0 flex-1">
              <div class="flex items-start justify-between gap-2">
                <button
                  type="button"
                  class="flex min-w-0 flex-1 items-center gap-2 text-left"
                  @click="openMenu(row.id)"
                >
                  <span class="truncate font-medium text-p1 text-ink">{{ row.full_name }}</span>
                  <UiPill
                    v-if="row.status"
                    :tone="statusBadge(row.status).tone"
                    class="capitalize shrink-0"
                  >
                    {{ statusBadge(row.status).label }}
                  </UiPill>
                </button>
                <button
                  type="button"
                  class="ml-auto flex size-8 items-center justify-center rounded text-ink-40 transition duration-150 group-hover:bg-canvas group-hover:text-ink menu-trigger"
                  aria-label="Open menu"
                  @click.stop="openMenu(row.id)"
                >
                  <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <circle cx="12" cy="5" r="1.5"/>
                    <circle cx="12" cy="12" r="1.5"/>
                    <circle cx="12" cy="19" r="1.5"/>
                  </svg>
                </button>
              </div>
              <p class="mt-1 truncate text-label text-ink-64">
                {{ row.phone }}
                <span v-if="row.license_number" class="text-ink-64"> — {{ row.license_number }}</span>
              </p>
              <p v-if="row.license_expiry" class="mt-0.5 truncate text-label text-ink-64">
                License expires: {{ row.license_expiry }}
              </p>
            </div>

            <!-- floating menu -->
            <div
              v-if="menuFor === row.id"
              class="absolute right-0 left-full -translate-x-1/2 top-full mt-1 z-50 w-48 rounded-lg bg-surface shadow-card border border-line py-1 fade-in"
              tabindex="-1"
            >
              <button
                type="button"
                class="flex w-full items-center gap-2 px-3 py-2 text-p1 text-ink transition duration-150 hover:bg-canvas"
                @click="openEdit(row); menuFor = null"
              >
                <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"/>
                </svg>
                Edit
              </button>
              <div class="my-1 border-t border-line"/>
              <button
                type="button"
                class="flex w-full items-center gap-2 px-3 py-2 text-p1 text-danger transition duration-150 hover:bg-red/5"
                @click="confirmDelete(row)"
              >
                <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="m3 6 4 14h10L19 6"/>
                  <path d="M8 6V4a1 1 0 0 1 1-1h6a1 1 0 0 1 1 1v2"/>
                </svg>
                Delete driver
              </button>
            </div>
          </div>

          <div v-if="drivers.length === 0 && !loading" class="py-8 text-center text-label text-ink-64">
            No drivers found.
          </div>
        </div>
      </div>
    </main>

    <!-- add / edit modal -->
    <Teleport to="body">
      <div
        v-if="adding || editing"
        class="fixed inset-0 z-50 flex items-center justify-center bg-ink/30 px-4 py-6 transition-opacity"
        @click.self="closeEditor"
      >
        <UiCard class="mx-auto max-w-sm">
          <div class="flex items-center justify-between pb-3 bg-canvas">
            <h2 class="text-h5 text-ink">{{ editing ? `Edit ${editing.full_name}` : 'Add driver' }}</h2>
            <div class="flex items-center gap-2">
              <span class="inline-flex items-center rounded bg-ink/5 px-1.5 text-meta font-normal text-ink-64 shadow-inset">Esc</span>
              <button
                type="button"
                class="flex size-8 items-center justify-center rounded-full text-ink-64 transition hover:bg-red/10 hover:text-danger focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
                aria-label="Close"
                @click="closeEditor"
              >
                <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true">
                  <path d="M18 6 6 18M6 6l12 12"/>
                </svg>
              </button>
            </div>
          </div>

          <form class="space-y-4" @submit.prevent="editing ? saveEdit() : saveAdd()">
            <UiField
              id="driver-name"
              v-model="addFullName"
              label="Full name"
              type="text"
              :error="''"
              autofocus
            />
            <UiField
              id="driver-phone"
              v-model="addPhone"
              label="Phone number"
              type="tel"
              :error="''"
            />
            <UiField
              id="driver-license"
              v-model="addLicenseNumber"
              label="License number"
              type="text"
              :error="''"
              placeholder="e.g. DL-12345"
            />
            <UiField
              id="driver-license-expiry"
              v-model="addLicenseExpiry"
              label="License expiry"
              type="text"
              :error="''"
              placeholder="YYYY-MM-DD"
            />
            <div>
              <UiSelect
                id="driver-branch"
                v-model="addBranch"
                label="Branch"
                :options="branchOptions"
                hint="branches.length ? '' : 'No branches available'"
              />
            </div>

            <div class="mt-5 flex items-center gap-2">
              <UiButton
                type="button"
                :loading="saving"
                @click="editing ? saveEdit() : saveAdd()"
              >
                {{ editing ? 'Save' : 'Add driver' }}
                <UiKbd>Enter</UiKbd>
              </UiButton>
              <UiButton
                v-if="editing"
                variant="link"
                type="button"
                @click="closeEditor"
              >
                Cancel
              </UiButton>
            </div>
          </form>
        </UiCard>
      </div>
    </Teleport>

    <!-- confirm delete modal -->
    <Teleport to="body">
      <div
        v-if="confirmModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-ink/30 px-4 py-6 transition-opacity"
        @click.self="closeConfirm"
      >
        <UiCard class="mx-auto max-w-sm">
          <div class="flex items-center justify-between pb-3 bg-canvas">
            <h2 class="text-h5 text-ink">Delete {{ confirmModal.driverName }}?</h2>
            <div class="flex items-center gap-2">
              <span class="inline-flex items-center rounded bg-ink/5 px-1.5 text-meta font-normal text-ink-64 shadow-inset">Esc</span>
              <button
                type="button"
                class="flex size-8 items-center justify-center rounded-full text-ink-64 transition hover:bg-red/10 hover:text-danger focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
                aria-label="Close"
                @click="closeConfirm"
              >
                <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true">
                  <path d="M18 6 6 18M6 6l12 12"/>
                </svg>
              </button>
            </div>
          </div>

          <p class="text-p1 text-ink">
            This driver will be marked inactive. Their data stays in the records.
          </p>

          <div class="mt-5 flex items-center gap-2">
            <UiButton
              type="button"
              :loading="saving"
              variant="danger"
              @click="executeDelete"
            >
              Delete driver
              <UiKbd>Enter</UiKbd>
            </UiButton>
          </div>
        </UiCard>
      </div>
    </Teleport>
  </div>
</template>
