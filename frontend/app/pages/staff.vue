<script setup lang="ts">
import { ApiError } from '~/utils/api'
import type { StaffItem, BranchItem, ResetOtpResponse, ConfirmModal, OtpModal } from '~/types/staff'
import UiSelect from '~/components/ui/Select.vue'

definePageMeta({ layout: 'default' })

useHead({ title: 'Staff' })

const { user } = useAuth()

// ---- data sources ----
const staff = ref<StaffItem[]>([])
const branches = ref<BranchItem[]>([])
const searchOpen = ref(false)
const loading = ref(false)

// ---- search / filter query state ----
const q = ref('')
const branchFilter = ref<string | null>(null)
const roleFilter = ref<string | null>(null)
const statusFilter = ref<string | null>(null)

// ---- add / edit modal state ----
const editing = ref<StaffItem | null>(null)
const adding = ref(false)
const saving = ref(false)

const addName = ref('')
const addPhone = ref('')
const addRole = ref('receiving')
const addBranch = ref('')

// ---- menu state ----
const menuFor = ref<string | null>(null)

// ---- confirmation modals ----
const confirmModal = ref<ConfirmModal | null>(null)

// ---- one-time password after a reset ----
const otpModal = ref<OtpModal | null>(null)

// ---- mapping helpers ----
const roleLabel: Record<string, string> = {
  staff: 'Staff',
  manager: 'Manager',
  owner: 'Owner',
}

const statusTone: Record<string, 'danger' | 'info' | 'neutral'> = {
  active: 'info',
  inactive: 'neutral',
}

const branchOptions = computed(() => [
  { value: '', label: 'All branches' },
  ...branches.value.map((b) => ({ value: b.id, label: b.name })),
])

const roleOptions = [
  { value: '', label: 'All roles' },
  { value: 'receiving', label: 'Receiving' },
  { value: 'dispatch', label: 'Dispatch' },
  { value: 'operations', label: 'Operations' },
]

const statusOptions = [
  { value: '', label: 'All' },
  { value: 'active', label: 'Active' },
  { value: 'inactive', label: 'Off' },
]

// ---- refresh loops ----
onMounted(refresh)
watch(user, () => { if (user.value) refresh() })

async function refresh() {
  if (!user.value) return
  loading.value = true
  try {
    await Promise.all([fetchStaff(), fetchBranches()])
  } finally {
    loading.value = false
  }
}

async function fetchStaff() {
  const params = new URLSearchParams()
  if (branchFilter.value) params.set('branch_id', branchFilter.value)
  const rows = await apiRequest<StaffItem[]>('/api/staff?' + params.toString(), { token: user.value?.token })
  staff.value = rows.map(normalizeStaff)
}

async function fetchBranches() {
  const rows = await apiRequest<BranchItem[]>('/api/branches', { token: user.value?.token })
  branches.value = rows.map((b) => ({ id: b.id, name: b.name }))
}

// ---- staff row normalization ----
function normalizeStaff(row: StaffItem) {
  const branchName = branches.value.find((b) => b.id === row.branch_id)?.name ?? 'Unknown branch'
  const manager = row.manager_name ?? null

  const sub = computed(() => {
    const parts = [roleLabel[row.role] ?? row.role, branchName]
    if (row.locked && row.locked_at) {
      parts.push(`${row.failed_login_attempts} wrong passwords at ${row.locked_at}`)
    } else if (row.first_login_done) {
      if (row.last_sign_in) parts.push(`Last signed in ${row.last_sign_in}`)
    } else {
      parts.push('First sign-in waiting')
    }
    if (row.status === 'inactive' && row.inactive_since) {
      parts[parts.length - 1] = `Access turned off ${row.inactive_since}`
    }
    return parts.join(' • ')
  })

  const status: StatusBadge = {
    active: { label: 'Active', tone: 'info' as const },
    inactive: { label: 'Off', tone: 'neutral' as const },
  }[row.status]

  if (row.locked) {
    badge.label = 'Locked'
    badge.tone = 'danger'
  } else if (!row.first_login_done) {
    badge.label = 'First sign-in waiting'
    badge.tone = 'neutral'
  }

  return {
    ...row,
    branch_name: branchName,
    subtext: sub.value,
    badge,
    manager,
  }
}

// ---- search / filter modal ----
function openSearch() {
  searchOpen.value = true
  q.value = ''
  branchFilter.value = ''
  roleFilter.value = ''
  statusFilter.value = ''
}

function closeSearch() {
  searchOpen.value = false
}

async function applyFilters() {
  await fetchStaff()
  closeSearch()
}

function clearFilters() {
  q.value = ''
  branchFilter.value = ''
  roleFilter.value = ''
  statusFilter.value = ''
}

// ---- add staff ----
function openAdd() {
  editing.value = null
  adding.value = true
  addName.value = ''
  addPhone.value = ''
  addRole.value = 'receiving'
  addBranch.value = branches.value[0]?.id ?? ''
}

async function saveAdd() {
  if (!addName.value.trim() || !addBranch.value) return
  saving.value = true
  try {
    await apiRequest('/api/staff', {
      method: 'POST',
      token: user.value?.token,
      body: {
        full_name: addName.value.trim(),
        phone: addPhone.value.trim(),
        job_role: addRole.value,
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

// ---- edit staff ----
function openEdit(row: StaffItem) {
  editing.value = row
  addName.value = row.full_name
  addPhone.value = row.phone ?? ''
  addRole.value = row.job_role ?? 'receiving'
  addBranch.value = row.branch_id
  adding.value = false
}

async function saveEdit() {
  if (!editing.value) return
  saving.value = true
  try {
    await apiRequest(`/api/staff/${editing.value.id}`, {
      method: 'PUT',
      token: user.value?.token,
      body: {
        full_name: addName.value.trim(),
        phone: addPhone.value.trim(),
        job_role: addRole.value,
        branch_id: addBranch.value,
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

// ---- actions from menu ----
async function resetPassword(row: StaffItem) {
  menuFor.value = null
  try {
    const result = await apiRequest<ResetOtpResponse>(`/api/staff/${row.id}/reset-pin`, {
      method: 'POST',
      token: user.value?.token,
    })
    otpModal.value = { staffId: row.staff_id, oneTimePassword: result.one_time_password, name: row.full_name }
  } catch (err) {
    if (err instanceof ApiError) {
      console.error(err.message)
    }
  }
}

async function turnOff(row: StaffItem) {
  menuFor.value = null
  confirmModal.value = {
    mode: 'off',
    staffId: row.id,
    staffName: row.full_name,
    currentStatus: row.status,
  }
}

async function turnOn(row: StaffItem) {
  menuFor.value = null
  confirmModal.value = {
    mode: 'on',
    staffId: row.id,
    staffName: row.full_name,
    currentStatus: row.status,
  }
}

async function removeStaff(row: StaffItem) {
  menuFor.value = null
  confirmModal.value = {
    mode: 'remove',
    staffId: row.id,
    staffName: row.full_name,
    currentStatus: row.status,
  }
}

async function confirmAction() {
  if (!confirmModal.value) return
  saving.value = true
  try {
    if (confirmModal.value.mode === 'off') {
      await apiRequest(`/api/staff/${confirmModal.value.staffId}`, {
        method: 'PUT',
        token: user.value?.token,
        body: { status: 'inactive' },
      })
    } else if (confirmModal.value.mode === 'on') {
      await apiRequest(`/api/staff/${confirmModal.value.staffId}`, {
        method: 'PUT',
        token: user.value?.token,
        body: { status: 'active' },
      })
    } else if (confirmModal.value.mode === 'remove') {
      await apiRequest(`/api/staff/${confirmModal.value.staffId}`, {
        method: 'DELETE',
        token: user.value?.token,
      })
    }
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

function closeOtp() {
  otpModal.value = null
}

// close any modal on escape
function onEscape() {
  if (menuFor.value) menuFor.value = null
  else if (searchOpen.value) closeSearch()
  else if (editing.value || adding.value) closeEditor()
  else if (confirmModal.value) closeConfirm()
  else if (otpModal.value) closeOtp()
}

const displayedStaff = computed(() => {
  let rows = staff.value
  if (q.value.trim()) {
    const ql = q.value.toLowerCase()
    rows = rows.filter(
      (r) =>
        r.full_name.toLowerCase().includes(ql) ||
        r.phone?.toLowerCase().includes(ql),
    )
  }
  if (branchFilter.value) {
    rows = rows.filter((r) => r.branch_id === branchFilter.value)
  }
  if (roleFilter.value) {
    rows = rows.filter((r) => r.role === roleFilter.value)
  }
  if (statusFilter.value) {
    rows = rows.filter((r) => {
      if (statusFilter.value === 'active') return r.status === 'active' && !r.locked
      if (statusFilter.value === 'inactive') return r.status === 'inactive' || r.locked
      return true
    })
  }
  return rows
})
</script>

<template>
  <div class="min-h-dvh" @keydown.esc="onEscape">
    <!-- top tabs -->
    <div class="border-b border-line bg-canvas">
      <div class="mx-auto flex max-w-[720px] border-b border-line px-4">
        <nav class="flex gap-1 after:col-start-2">
          <NuxtLink
            to="/staff"
            class="-(m-1) flex h-11 flex-1 items-center justify-center rounded-tl-lg px-3 text-button font-medium transition duration-150 first:rounded-tl-lg last:rounded-tr-lg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
            :class="[
              user?.role === 'staff'
                ? 'text-ink'
                : 'hover:bg-canvas/60',
              route.path === '/staff'
                ? 'bg-surface shadow-raised text-ink'
                : 'text-ink-64',
            ]"
          >
            Staff
          </NuxtLink>
          <NuxtLink
            to="/drivers"
            class="-(m-1) flex h-11 flex-1 items-center justify-center rounded-lg px-3 text-button font-medium transition duration-150 hover:bg-canvas/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
            :class="route.path === '/drivers' ? 'bg-surface shadow-raised text-ink' : 'text-ink-64'"
          >
            Drivers
          </NuxtLink>
          <NuxtLink
            to="/vehicles"
            class="-(m-1) flex h-11 flex-1 items-center justify-center rounded-lg px-3 text-button font-medium transition duration-150 hover:bg-canvas/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
            :class="route.path === '/vehicles' ? 'bg-surface shadow-raised text-ink' : 'text-ink-64'"
          >
            Vehicles
          </NuxtLink>
          <NuxtLink
            to="/branches"
            class="-(m-1) flex h-11 flex-1 items-center justify-center rounded-tr-lg px-3 text-button font-medium transition duration-150 last:rounded-tr-lg hover:bg-canvas/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
            :class="route.path === '/branches' ? 'bg-surface shadow-raised text-ink' : 'text-ink-64'"
          >
            Branches
          </NuxtLink>
        </nav>
      </div>
    </div>

    <!-- main content -->
    <main class="mx-auto max-w-[720px] px-4 pb-14 pt-6">
      <div class="space-y-4">
        <!-- search bar -->
        <div class="relative">
          <label for="staff-search" class="sr-only">Search by name or phone number</label>
          <input
            id="staff-search"
            v-model="q"
            type="search"
            placeholder="Search by name or phone number"
            class="h-11 w-full rounded-lg border border-line bg-surface pl-10 pr-10 text-p1 text-ink placeholder:text-ink-40 transition duration-150 focus:border-blue focus:outline-none focus:ring-2 focus:ring-blue/20"
            @focus="searchOpen = true"
          />
          <span class="absolute left-3 top-1/2 -translate-y-1/2 text-ink-40">
            <!-- magnifier -->
            <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <circle cx="11" cy="11" r="7"/>
              <path d="m20 20-3.5-3.5"/>
            </svg>
          </span>
          <span class="absolute right-3 top-1/2 -translate-y-1/2 text-ink-40">
            <span class="inline-flex items-center rounded bg-ink/5 px-1 text-meta font-normal text-ink-64 shadow-inset">/</span>
          </span>
        </div>

        <!-- header -->          <div class="flex items-start justify-between gap-4">
          <div>
            <h1 class="text-h3 text-ink">Staff</h1>
            <p class="mt-1 text-label text-ink-64">
              Only a manager or the owner can add staff. There is no sign-up.
            </p>
          </div>
          <button
            type="button"
            class="flex h-10 shrink-0 items-center justify-center gap-1 rounded-lg border border-line bg-surface px-3 text-label text-ink-64 shadow-inset transition duration-150 hover:border-blue hover:text-blue focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
            @click="openSearch"
          >
            <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M4 6h16M7 12h10M10 18h4"/>
            </svg>
            <span class="text-meta font-medium tracking-[0.06em] uppercase">F</span>
          </button>
        </div>

        <!-- staff list -->
        <div class="divide-y divide-line">
          <div
            v-for="row in displayedStaff"
            :key="row.id"
            class="group flex items-start gap-3 py-3 first:pt-0 last:pb-0"
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
                    :tone="row.badge.tone"
                    class="shrink-0 capitalize"
                  >
                    {{ row.badge.label }}
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
              <p class="mt-1 truncate text-label text-ink-64">{{ row.subtext }}</p>
            </div>

            <!-- floating menu -->
            <div
              v-if="menuFor === row.id"
              class="absolute right-0 z-50 mt-2 w-48 rounded-lg bg-surface shadow-card border border-line py-1"
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
              <button
                type="button"
                class="flex w-full items-center gap-2 px-3 py-2 text-p1 text-ink transition duration-150 hover:bg-canvas"
                @click="resetPassword(row)"
              >
                <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <rect width="18" height="11" x="3" y="11" rx="2"/>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
                </svg>
                Reset password
              </button>
              <button
                type="button"
                class="flex w-full items-center gap-2 px-3 py-2 text-p1 text-ink transition duration-150 hover:bg-canvas"
                @click="row.status === 'active' ? turnOff(row) : turnOn(row)"
              >
                <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M15 14c-1.5 0-2.8-1.4-3-2-1-2-1-2-1-2s1-1 1-1a7 7 0 0 1 14 0s1 1 1 1a1 1 0 0 1 0 2c0 0-1 0-1 2-1 2-2.8 2-3 2Z"/>
                  <path d="M7 12a5 5 0 0 1 10 0"/>
                  <path d="m21 3-8 8"/>
                </svg>
                {{ row.status === 'active' ? 'Turn off access' : 'Turn on access' }}
              </button>
              <div class="my-1 border-t border-line"/>
              <button
                type="button"
                class="flex w-full items-center gap-2 px-3 py-2 text-p1 text-danger transition duration-150 hover:bg-red/5"
                @click="removeStaff(row)"
              >
                <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="m3 6 4 14h10L19 6"/>
                  <path d="M8 6V4a1 1 0 0 1 1-1h6a1 1 0 0 1 1 1v2"/>
                </svg>
                Remove staff
              </button>
            </div>
          </div>

          <div v-if="displayedStaff.length === 0 && !loading" class="py-8 text-center text-label text-ink-64">
            No staff found.
          </div>
        </div>

        <!-- add button -->
        <button
          type="button"
          class="flex items-center gap-2 rounded-lg bg-blue px-4 text-white shadow-raised transition duration-150 hover:brightness-105 active:brightness-95 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue disabled:opacity-60"
          :disabled="!user || user.role === 'staff'"
          @click="openAdd"
        >
          <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" aria-hidden="true">
            <path d="M12 5v14M5 12h14"/>
          </svg>
          Add staff
          <UiKbd>Enter</UiKbd>
        </button>
      </div>
    </main>

    <!-- filter modal -->
    <Teleport to="body">
      <div
        v-if="searchOpen"
        class="fixed inset-0 z-50 bg-ink/30 px-4 py-6 transition-opacity"
        @click.self="closeSearch"
      >
        <div class="mx-auto max-w-sm rounded-lg bg-surface shadow-halo p-4">
          <div class="flex items-center justify-between pb-3">
            <h2 class="text-h5 text-ink">Filter staff</h2>
            <div class="flex items-center gap-2">
              <span class="inline-flex items-center rounded bg-ink/5 px-1.5 text-meta font-normal text-ink-64 shadow-inset">Esc</span>
              <button
                type="button"
                class="flex size-8 items-center justify-center rounded-full text-ink-64 transition hover:bg-red/10 hover:text-danger focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
                aria-label="Close"
                @click="closeSearch"
              >
                <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true">
                  <path d="M18 6 6 18M6 6l12 12"/>
                </svg>
              </button>
            </div>
          </div>

          <div class="space-y-3">
            <UiSelect
              id="filter-branch"
              v-model="branchFilter"
              label="Branch"
              :model-value="branchFilter ?? ''"
              :options="branchOptions"
              hint="All branches"
            />
            <UiSelect
              id="filter-role"
              v-model="roleFilter"
              label="Role"
              :model-value="roleFilter ?? ''"
              :options="roleOptions"
              hint="All roles"
            />
            <UiSelect
              id="filter-status"
              v-model="statusFilter"
              label="Status"
              :model-value="statusFilter ?? ''"
              :options="statusOptions"
              hint="All"
            />
          </div>

          <div class="mt-5 flex items-center gap-2">
            <UiButton type="button" @click="applyFilters">
              Show results
              <UiKbd>Enter</UiKbd>
            </UiButton>
            <UiButton variant="tertiary" type="button" @click="clearFilters">
              Clear filters
              <UiKbd>C</UiKbd>
            </UiButton>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- add / edit modal -->
    <Teleport to="body">
      <div
        v-if="adding || editing"
        class="fixed inset-0 z-50 bg-ink/30 px-4 py-6 transition-opacity"
        @click.self="closeEditor"
      >
        <UiCard class="mx-auto max-w-sm">
          <div class="flex items-center justify-between pb-3">
            <h2 class="text-h5 text-ink">{{ editing ? `Edit ${editing.full_name}` : 'Add staff' }}</h2>
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

          <template v-if="editing">
            <p class="mb-4 text-label text-ink-64">
              After saving, you get a one-time password to give them.
            </p>
          </template>

          <form
            v-if="!editing"
            class="space-y-4"
            @submit.prevent="saveAdd"
          >
            <UiField
              id="add-name"
              v-model="addName"
              label="Full name"
              type="text"
              :error="''"
              autofocus
            />
            <UiField
              id="add-phone"
              v-model="addPhone"
              label="Phone number"
              type="tel"
              :error="''"
            />              <div class="grid grid-cols-2 gap-4">
              <div>
                <UiSelect
                  id="add-role"
                  v-model="addRole"
                  label="Role"
                  :model-value="addRole"
                  :options="roleOptions"
                />
                <p class="mt-1 text-label text-ink-64">Each branch has one manager.</p>
              </div>
              <UiSelect
                id="add-branch"
                v-model="addBranch"
                label="Branch"
                :model-value="addBranch"
                :options="branchOptions"
                :hint="branches.length ? '' : 'No branches available'"
              />
            </div>
          </form>

          <form
            v-else
            class="space-y-4"
            @submit.prevent="saveEdit"
          >
            <UiField
              id="edit-name"
              v-model="addName"
              label="Full name"
              type="text"
              :error="''"
            />
            <UiField
              id="edit-phone"
              v-model="addPhone"
              label="Phone number"
              type="tel"
              :error="''"
            />              <div class="grid grid-cols-2 gap-4">
              <div>
                <UiSelect
                  id="edit-role"
                  v-model="addRole"
                  label="Role"
                  :model-value="addRole"
                  :options="roleOptions"
                />
                <p v-if="editing?.manager_name" class="mt-1 text-label text-ink-64">
                  {{ editing.manager_name }} is the manager.
                </p>
              </div>
              <UiSelect
                id="edit-branch"
                v-model="addBranch"
                label="Branch"
                :model-value="addBranch"
                :options="branchOptions"
              />
            </div>
          </form>

          <div class="mt-5 flex items-center gap-2">
            <UiButton
              type="button"
              :loading="saving"
              :variant="editing ? 'primary' : 'primary'"
              @click="editing ? saveEdit() : saveAdd()"
            >
              {{ editing ? 'Save' : 'Add staff' }}
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
        </UiCard>
      </div>
    </Teleport>

    <!-- confirm action modal (turn off / turn on / remove) -->
    <Teleport to="body">
      <div
        v-if="confirmModal"
        class="fixed inset-0 z-50 bg-ink/30 px-4 py-6 transition-opacity"
        @click.self="closeConfirm"
      >
        <UiCard class="mx-auto max-w-sm">
          <div class="flex items-center justify-between pb-3">
            <h2 class="text-h5 text-ink">
              {{ confirmModal.mode === 'off' ? `Turn off ${confirmModal.staffName}'s access?` : '' }}
              {{ confirmModal.mode === 'on' ? `Turn on ${confirmModal.staffName}'s access?` : '' }}
              {{ confirmModal.mode === 'remove' ? `Remove ${confirmModal.staffName}?` : '' }}
            </h2>
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
            {{ confirmModal.mode === 'off'
              ? `${confirmModal.staffName} cannot sign in any more. Everything ${confirmModal.staffName.toLowerCase()} did stays in the records.`
              : confirmModal.mode === 'on'
                ? `${confirmModal.staffName} can sign in again with the same staff ID and password.`
                : `${confirmModal.staffName} will not be able to sign in. Work ${confirmModal.staffName.toLowerCase()} already did stays in the records.`
            }}
          </p>

          <div class="mt-5 flex items-center gap-2">
            <UiButton
              type="button"
              :loading="saving"
              :variant="confirmModal.mode === 'remove' ? 'danger' : 'primary'"
              @click="confirmAction"
            >
              {{ confirmModal.mode === 'remove' ? 'Remove staff' : confirmModal.mode === 'off' ? 'Turn off' : 'Turn on access' }}
              <UiKbd>Enter</UiKbd>
            </UiButton>
          </div>
        </UiCard>
      </div>
    </Teleport>

    <!-- one-time password modal -->
    <Teleport to="body">
      <div
        v-if="otpModal"
        class="fixed inset-0 z-50 bg-ink/30 px-4 py-6 transition-opacity"
        @click.self="closeOtp"
      >
        <UiCard class="mx-auto max-w-sm">
          <div class="flex items-center justify-between pb-3">
            <h2 class="text-h5 text-ink">{{ otpModal.name }} can sign in now</h2>
            <div class="flex items-center gap-2">
              <span class="inline-flex items-center rounded bg-ink/5 px-1.5 text-meta font-normal text-ink-64 shadow-inset">Esc</span>
              <button
                type="button"
                class="flex size-8 items-center justify-center rounded-full text-ink-64 transition hover:bg-red/10 hover:text-danger focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
                aria-label="Close"
                @click="closeOtp"
              >
                <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true">
                  <path d="M18 6 6 18M6 6l12 12"/>
                </svg>
              </button>
            </div>
          </div>

          <p class="text-p1 text-ink">
            Give this password to {{ otpModal.name.toLowerCase() }} in person. It works once. Then {{ otpModal.name.toLowerCase() }} chooses a new password.
          </p>

          <div class="mt-4 rounded-lg bg-canvas p-3 font-mono text-center text-p3 font-medium tracking-wider text-ink shadow-inset">
            {{ otpModal.oneTimePassword }}
          </div>

          <div class="mt-5 flex items-center gap-2">
            <UiButton type="button" @click="closeOtp">
              Done
              <UiKbd>Enter</UiKbd>
            </UiButton>
          </div>
        </UiCard>
      </div>
    </Teleport>
  </div>
</template>
