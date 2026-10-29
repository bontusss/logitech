<script setup lang="ts">
import type { CreateUserResponse, Staff } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Staff' })
definePageMeta({ layout: 'default', roles: ['owner', 'manager', 'system_admin'] })

const { token, user } = useAuth()
const { isOwner, isSystemAdmin, canViewAllBranches, branchId } = usePermissions()
const { options: branchOptions, name: branchName, code: branchCode, load: loadBranches } = useBranches()

const staff = ref<Staff[]>([])
const loading = ref(true)
const error = ref('')
const search = ref('')
const branchFilter = ref('')
const statusFilter = ref<'all' | 'active' | 'inactive'>('all')

const canCreateManager = computed(() => isOwner.value || isSystemAdmin.value)

const jobRoles = [
  { value: 'operations', label: 'Operations' },
  { value: 'dispatch', label: 'Dispatch' },
  { value: 'receiving', label: 'Receiving' },
]

// Add dialog
const addOpen = ref(false)
const addRole = ref<'staff' | 'manager'>('staff')
const addForm = reactive({ fullName: '', phone: '', email: '', branch: '', jobRole: 'operations' })
const adding = ref(false)
const addError = ref('')

// Edit dialog
const editTarget = ref<Staff | null>(null)
const editForm = reactive({ fullName: '', phone: '', email: '', jobRole: '' })
const editing = ref(false)
const editError = ref('')

// Revoke dialog
const revokeTarget = ref<Staff | null>(null)
const revoking = ref(false)
const revokeError = ref('')

// One-time password handover
const otp = ref<{ name: string; staffId: string; password: string } | null>(null)
const copied = ref(false)

const busyId = ref('')
const notice = ref('')

const filtered = computed(() => {
  const term = search.value.trim().toLowerCase()
  return staff.value.filter((member) => {
    if (statusFilter.value !== 'all' && member.status !== statusFilter.value) return false
    if (!term) return true
    return [member.full_name, member.staff_id, member.phone, member.email, member.job_role]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(term))
  })
})

const activeCount = computed(() => staff.value.filter((member) => member.status === 'active').length)
const managerCount = computed(() => staff.value.filter((member) => member.role === 'manager').length)

function roleLabel(member: Staff): string {
  if (member.role === 'manager') return 'Branch manager'
  if (member.role === 'owner') return 'Owner'
  if (member.role === 'system_admin') return 'System admin'
  return humanise(member.job_role || 'staff')
}

function branchLabel(member: Staff): string {
  const name = branchName(member.branch_id)
  const code = branchCode(member.branch_id)
  return name ? `${name}${code ? ` (${code})` : ''}` : member.branch_id
}

async function load() {
  if (!token.value) return
  loading.value = true
  error.value = ''
  const params = new URLSearchParams()
  if (canViewAllBranches.value && branchFilter.value) params.set('branch_id', branchFilter.value)
  try {
    staff.value = await apiRequest<Staff[]>(`/api/staff${params.size ? `?${params.toString()}` : ''}`, {
      token: token.value,
    })
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not load staff.'
  } finally {
    loading.value = false
  }
}

function openAdd(role: 'staff' | 'manager') {
  addRole.value = role
  addError.value = ''
  addForm.fullName = ''
  addForm.phone = ''
  addForm.email = ''
  addForm.jobRole = 'operations'
  addForm.branch = branchId.value || branchFilter.value || branchOptions.value[0]?.value || ''
  addOpen.value = true
}

async function addPerson() {
  if (!token.value) return
  addError.value = ''
  if (!addForm.fullName.trim()) {
    addError.value = 'Enter the full name.'
    return
  }
  if (!addForm.branch) {
    addError.value = 'Choose the branch this person works in.'
    return
  }

  adding.value = true
  try {
    const created = await apiRequest<CreateUserResponse>(
      addRole.value === 'manager' ? '/api/managers' : '/api/staff',
      {
        method: 'POST',
        token: token.value,
        body: {
          branch_id: addForm.branch,
          full_name: addForm.fullName.trim(),
          phone: addForm.phone.trim() || undefined,
          email: addForm.email.trim() || undefined,
          ...(addRole.value === 'staff' ? { job_role: addForm.jobRole } : {}),
        },
      },
    )
    addOpen.value = false
    otp.value = { name: created.full_name, staffId: created.staff_id, password: created.one_time_password }
    await load()
  } catch (err) {
    addError.value = err instanceof ApiError ? err.message : 'Could not create the account.'
  } finally {
    adding.value = false
  }
}

function openEdit(member: Staff) {
  editTarget.value = member
  editForm.fullName = member.full_name
  editForm.phone = member.phone ?? ''
  editForm.email = member.email ?? ''
  editForm.jobRole = member.job_role ?? 'operations'
  editError.value = ''
}

async function saveEdit() {
  if (!token.value || !editTarget.value) return
  editError.value = ''
  if (!editForm.fullName.trim()) {
    editError.value = 'Enter the full name.'
    return
  }

  editing.value = true
  try {
    await apiRequest<Staff>(`/api/staff/${editTarget.value.id}`, {
      method: 'PUT',
      token: token.value,
      body: {
        full_name: editForm.fullName.trim(),
        phone: editForm.phone.trim(),
        email: editForm.email.trim(),
        ...(editTarget.value.role === 'staff' ? { job_role: editForm.jobRole } : {}),
      },
    })
    editTarget.value = null
    notice.value = 'Staff details updated.'
    await load()
  } catch (err) {
    editError.value = err instanceof ApiError ? err.message : 'Could not save the changes.'
  } finally {
    editing.value = false
  }
}

async function resetPin(member: Staff) {
  if (!token.value) return
  busyId.value = member.id
  notice.value = ''
  error.value = ''
  try {
    const result = await apiRequest<{ staff_id: string; one_time_password: string }>(
      `/api/staff/${member.id}/reset-pin`,
      { method: 'POST', token: token.value },
    )
    otp.value = { name: member.full_name, staffId: result.staff_id, password: result.one_time_password }
    await load()
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not reset the PIN.'
  } finally {
    busyId.value = ''
  }
}

async function unlock(member: Staff) {
  if (!token.value) return
  busyId.value = member.id
  notice.value = ''
  try {
    await apiRequest(`/api/staff/${member.id}/unlock`, { method: 'POST', token: token.value })
    notice.value = `${member.full_name}'s account is unlocked.`
    await load()
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not unlock the account.'
  } finally {
    busyId.value = ''
  }
}

async function reactivate(member: Staff) {
  if (!token.value) return
  busyId.value = member.id
  notice.value = ''
  try {
    await apiRequest(`/api/staff/${member.id}`, {
      method: 'PUT',
      token: token.value,
      body: { status: 'active' },
    })
    notice.value = `${member.full_name} can sign in again.`
    await load()
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not reactivate the account.'
  } finally {
    busyId.value = ''
  }
}

async function revoke() {
  if (!token.value || !revokeTarget.value) return
  revoking.value = true
  revokeError.value = ''
  try {
    await apiRequest(`/api/staff/${revokeTarget.value.id}`, { method: 'DELETE', token: token.value })
    notice.value = `${revokeTarget.value.full_name} no longer has access.`
    revokeTarget.value = null
    await load()
  } catch (err) {
    revokeError.value = err instanceof ApiError ? err.message : 'Could not revoke access.'
  } finally {
    revoking.value = false
  }
}

async function copyOtp() {
  if (!otp.value) return
  try {
    await navigator.clipboard.writeText(otp.value.password)
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  } catch {
    copied.value = false
  }
}

watch(branchFilter, load)
watch(otp, () => (copied.value = false))

onMounted(async () => {
  if (!canViewAllBranches.value && branchId.value) branchFilter.value = branchId.value
  await loadBranches()
  await load()
})
</script>

<template>
  <div class="space-y-4">
    <header class="flex flex-wrap items-end justify-between gap-3">
      <div class="min-w-0">
        <h1 class="text-h5 font-medium tracking-tight text-ink">Staff</h1>
        <p class="mt-1 text-label text-ink-64">
          {{ isOwner ? 'Everyone across every branch.' : `The team at ${branchName(branchId) || 'your branch'}.` }}
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <UiButton v-if="canCreateManager" variant="tertiary" @click="openAdd('manager')">
          <UiIcon name="user" :size="17" />
          Add manager
        </UiButton>
        <UiButton @click="openAdd('staff')">
          <UiIcon name="plus" :size="17" />
          Add staff
        </UiButton>
      </div>
    </header>

    <div v-if="notice" class="rounded-lg border border-blue/25 bg-blue/5 px-4 py-3 text-label text-ink" role="status">
      {{ notice }}
    </div>
    <div v-if="error" class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-danger/25 bg-danger/5 px-4 py-3">
      <p class="text-label text-danger">{{ error }}</p>
      <UiButton variant="tertiary" :loading="loading" @click="load">Try again</UiButton>
    </div>

    <div class="flex flex-wrap gap-2">
      <div class="min-w-[220px] flex-1">
        <UiSearchBar v-model="search" placeholder="Name, staff ID or phone" />
      </div>
      <UiSelect
        v-if="canViewAllBranches"
        id="staff-branch"
        v-model="branchFilter"
        label="Branch filter"
        hide-label
        :options="[{ value: '', label: 'All branches' }, ...branchOptions]"
        class="w-[210px]"
      />
      <UiSelect
        id="staff-status"
        v-model="statusFilter"
        label="Status filter"
        hide-label
        :options="[
          { value: 'all', label: 'All statuses' },
          { value: 'active', label: 'Active' },
          { value: 'inactive', label: 'Inactive' },
        ]"
        class="w-[150px]"
      />
    </div>

    <div v-if="loading" class="space-y-2" aria-hidden="true">
      <div v-for="index in 3" :key="index" class="h-[92px] animate-pulse rounded-lg bg-surface shadow-card" />
    </div>

    <template v-else>
      <section v-if="staff.length" class="rounded-lg bg-surface shadow-card">
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-line px-4 py-3">
          <p class="text-label text-ink-64">{{ filtered.length }} of {{ staff.length }} people</p>
          <p class="text-meta text-ink-40">
            {{ activeCount }} active<template v-if="managerCount"> · {{ managerCount }} manager{{ managerCount === 1 ? '' : 's' }}</template>
          </p>
        </div>

        <ul class="divide-y divide-line">
          <li v-for="member in filtered" :key="member.id" class="p-4">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div class="flex min-w-0 items-start gap-3">
                <span class="flex size-10 shrink-0 items-center justify-center rounded-full bg-canvas text-label font-medium text-ink">
                  {{ member.full_name.split(' ').slice(0, 2).map((part) => part[0]).join('').toUpperCase() }}
                </span>
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-1.5">
                    <p class="text-p1 font-medium text-ink">{{ member.full_name }}</p>
                    <UiPill v-if="member.role === 'manager'" tone="info">Manager</UiPill>
                    <UiPill v-else-if="member.locked" tone="danger">Locked</UiPill>
                    <UiPill v-else-if="member.status !== 'active'" tone="neutral">Inactive</UiPill>
                    <UiPill v-else-if="!member.first_login_done" tone="info">First login pending</UiPill>
                  </div>
                  <p class="mt-0.5 text-meta text-ink-40">
                    {{ member.staff_id }} · {{ roleLabel(member) }}
                    <template v-if="canViewAllBranches"> · {{ branchLabel(member) }}</template>
                  </p>
                  <p class="mt-0.5 text-meta text-ink-64">
                    {{ member.phone || 'No phone' }}<template v-if="member.email"> · {{ member.email }}</template>
                  </p>
                </div>
              </div>

              <div class="flex flex-wrap items-center gap-x-3 gap-y-1">
                <UiButton variant="link" @click="openEdit(member)">Edit</UiButton>
                <UiButton variant="link" :loading="busyId === member.id" @click="resetPin(member)">Reset PIN</UiButton>
                <UiButton v-if="member.locked" variant="link" :loading="busyId === member.id" @click="unlock(member)">
                  Unlock
                </UiButton>
                <UiButton
                  v-if="member.status !== 'active' && canViewAllBranches"
                  variant="link"
                  :loading="busyId === member.id"
                  @click="reactivate(member)"
                >
                  Reactivate
                </UiButton>
                <UiButton
                  v-if="member.status === 'active' && member.id !== user?.id"
                  variant="link"
                  class="text-danger"
                  @click="revokeTarget = member"
                >
                  Revoke
                </UiButton>
              </div>
            </div>
          </li>
        </ul>

        <div v-if="!filtered.length" class="px-4 py-10 text-center">
          <p class="text-p1 font-medium text-ink">No one matches those filters</p>
          <p class="mt-1 text-label text-ink-64">Try a different search or branch.</p>
        </div>
      </section>

      <UiEmptyState
        v-else
        icon="users"
        title="No staff yet"
        description="Add the people who record and receive waybills."
      >
        <UiButton @click="openAdd('staff')">Add staff</UiButton>
      </UiEmptyState>
    </template>

    <!-- Add -->
    <UiModal
      :open="addOpen"
      :title="addRole === 'manager' ? 'Add branch manager' : 'Add staff'"
      :description="
        addRole === 'manager'
          ? 'A manager runs one branch: sending, receiving and its staff.'
          : 'They will get a staff ID and a one-time password to set their PIN.'
      "
      size="md"
      :dismissible="!adding"
      @close="addOpen = false"
    >
      <div v-if="addError" class="mb-3 rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
        {{ addError }}
      </div>
      <div class="space-y-4">
        <UiField id="add-name" v-model="addForm.fullName" label="Full name" autocomplete="off" />
        <div class="grid gap-4 sm:grid-cols-2">
          <UiField id="add-phone" v-model="addForm.phone" label="Phone (optional)" type="tel" autocomplete="off" />
          <UiField id="add-email" v-model="addForm.email" label="Email (optional)" type="email" autocomplete="off" />
        </div>
        <UiSelect
          id="add-branch"
          v-model="addForm.branch"
          label="Branch"
          :options="branchOptions"
          :disabled="!canViewAllBranches"
          :hint="canViewAllBranches ? undefined : 'Managers create staff in their own branch.'"
        />
        <UiSelect
          v-if="addRole === 'staff'"
          id="add-job-role"
          v-model="addForm.jobRole"
          label="Job role"
          :options="jobRoles"
          hint="Operations and dispatch send; receiving confirms deliveries."
        />
      </div>
      <template #footer>
        <div class="flex flex-col gap-2 sm:flex-row-reverse">
          <UiButton block :loading="adding" @click="addPerson">Create account</UiButton>
          <UiButton variant="tertiary" block :disabled="adding" @click="addOpen = false">Cancel</UiButton>
        </div>
      </template>
    </UiModal>

    <!-- Edit -->
    <UiModal
      :open="Boolean(editTarget)"
      title="Edit staff"
      :description="editTarget ? `${editTarget.staff_id} · ${roleLabel(editTarget)}` : ''"
      size="md"
      :dismissible="!editing"
      @close="editTarget = null"
    >
      <div v-if="editError" class="mb-3 rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
        {{ editError }}
      </div>
      <div class="space-y-4">
        <UiField id="edit-name" v-model="editForm.fullName" label="Full name" autocomplete="off" />
        <div class="grid gap-4 sm:grid-cols-2">
          <UiField id="edit-phone" v-model="editForm.phone" label="Phone" type="tel" autocomplete="off" />
          <UiField id="edit-email" v-model="editForm.email" label="Email" type="email" autocomplete="off" />
        </div>
        <UiSelect
          v-if="editTarget?.role === 'staff'"
          id="edit-job-role"
          v-model="editForm.jobRole"
          label="Job role"
          :options="jobRoles"
        />
      </div>
      <template #footer>
        <div class="flex flex-col gap-2 sm:flex-row-reverse">
          <UiButton block :loading="editing" @click="saveEdit">Save changes</UiButton>
          <UiButton variant="tertiary" block :disabled="editing" @click="editTarget = null">Cancel</UiButton>
        </div>
      </template>
    </UiModal>

    <!-- Revoke -->
    <UiModal
      :open="Boolean(revokeTarget)"
      title="Revoke access"
      description="The account stays on record but can no longer sign in."
      :dismissible="!revoking"
      @close="revokeTarget = null"
    >
      <div v-if="revokeError" class="mb-3 rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
        {{ revokeError }}
      </div>
      <p class="text-label text-ink-64">
        Revoke <span class="font-medium text-ink">{{ revokeTarget?.full_name }}</span> ({{ revokeTarget?.staff_id }})?
      </p>
      <template #footer>
        <div class="flex flex-col gap-2 sm:flex-row-reverse">
          <UiButton variant="danger" block :loading="revoking" @click="revoke">Revoke access</UiButton>
          <UiButton variant="tertiary" block :disabled="revoking" @click="revokeTarget = null">Keep access</UiButton>
        </div>
      </template>
    </UiModal>

    <!-- One-time password -->
    <UiModal
      :open="Boolean(otp)"
      title="One-time password"
      description="Hand this over privately. It is shown once and replaces the old PIN."
      :dismissible="false"
      @close="otp = null"
    >
      <div v-if="otp" class="space-y-4">
        <div class="rounded-lg bg-canvas p-4">
          <p class="text-label text-ink-64">{{ otp.name }}</p>
          <p class="mt-1 text-h5 font-medium tracking-tight text-ink">{{ otp.staffId }}</p>
          <p class="mt-3 font-mono text-h3 leading-none tracking-[0.08em] text-ink">{{ otp.password }}</p>
        </div>
        <p class="text-label text-ink-64">
          They sign in with their staff ID and this password, then choose a new 4–10 character PIN.
        </p>
        <UiButton variant="tertiary" block @click="copyOtp">{{ copied ? 'Copied' : 'Copy password' }}</UiButton>
      </div>
      <template #footer>
        <UiButton block @click="otp = null">Done</UiButton>
      </template>
    </UiModal>
  </div>
</template>
