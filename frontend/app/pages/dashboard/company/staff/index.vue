<script setup lang="ts">
import { CircleAlert, Loader2, Search, Send, SlidersHorizontal, UserPlus, X } from 'lucide-vue-next'
import type { Branch, CompanyInvitation, Role, StaffMember } from '~/types/auth'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Staff' })

const { api } = useApi()
const { load, can, notify, confirm } = useWorkspace()

const members = ref<StaffMember[]>([])
const branches = ref<Branch[]>([])
const invitations = ref<CompanyInvitation[]>([])
const roles = ref<Role[]>([])
const loading = ref(true)
const error = ref('')
const search = ref('')
const status = ref<'all' | 'active' | 'suspended'>('all')
const branchFilter = ref('all')

const inviteOpen = ref(false)
const inviteSaving = ref(false)
const devInviteToken = ref('')

const editing = ref<StaffMember | null>(null)
const editingSaving = ref(false)
const editError = ref('')
const editForm = reactive({
  title: '',
  status: 'active',
  branch_ids: [] as string[],
  role_key: '',
})

onMounted(async () => {
  await load()
  await fetchData()
})

async function fetchData() {
  loading.value = true
  error.value = ''
  try {
    members.value = (await api<{ members: StaffMember[] }>('/companies/me/members')).members
    branches.value = (await api<{ branches: Branch[] }>('/companies/me/branches')).branches
    invitations.value = (await api<{ invitations: CompanyInvitation[] }>('/companies/me/invitations')).invitations
    if (can('staff.roles.manage')) {
      roles.value = (await api<{ roles: Role[] }>('/companies/me/roles')).roles
    }
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not load the staff directory.')
  } finally {
    loading.value = false
  }
}

const filtered = computed(() => {
  const term = search.value.trim().toLowerCase()
  return members.value.filter((member) => {
    const matchesStatus = status.value === 'all' || member.status === status.value
    const matchesSearch = !term ||
      member.full_name.toLowerCase().includes(term) ||
      member.email.toLowerCase().includes(term) ||
      (member.title ?? '').toLowerCase().includes(term)
    const matchesBranch = branchFilter.value === 'all'
      || (branchFilter.value === 'none' && member.branch_ids.length === 0)
      || member.branch_ids.includes(branchFilter.value)
    return matchesStatus && matchesSearch && matchesBranch
  })
})

const roleNames = computed(() => {
  const map = new Map<string, string>()
  for (const role of roles.value) map.set(role.key, role.name)
  return map
})

function roleLabel(key: string) {
  return roleNames.value.get(key) ?? key.replaceAll('_', ' ')
}

function branchNames(ids: string[]) {
  return ids.map((id) => branches.value.find((branch) => branch.id === id)?.name).filter(Boolean).join(', ')
}

async function invite(payload: { email: string; role_key: string; branch_ids: string[] }) {
  inviteSaving.value = true
  error.value = ''
  try {
    const response = await api<{ dev_invite_token?: string }>('/companies/me/invitations', {
      method: 'POST',
      body: payload,
    })
    devInviteToken.value = response.dev_invite_token ?? ''
    inviteOpen.value = false
    notify('Invitation sent.')
    await fetchData()
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not send the invitation.')
  } finally {
    inviteSaving.value = false
  }
}

function openEdit(member: StaffMember) {
  editing.value = member
  editForm.title = member.title ?? ''
  editForm.status = member.status === 'suspended' ? 'suspended' : 'active'
  editForm.branch_ids = [...member.branch_ids]
  editForm.role_key = member.roles.find((role) => role !== 'owner') ?? ''
  editError.value = ''
}

function toggleEditBranch(id: string) {
  editForm.branch_ids = editForm.branch_ids.includes(id)
    ? editForm.branch_ids.filter((branchID) => branchID !== id)
    : [...editForm.branch_ids, id]
}

async function saveMember() {
  if (!editing.value) return
  editingSaving.value = true
  editError.value = ''
  try {
    await api(`/companies/me/members/${editing.value.id}`, {
      method: 'PATCH',
      body: { title: editForm.title.trim(), status: editForm.status },
    })
    if (editForm.status === 'active') {
      await api(`/companies/me/members/${editing.value.id}/branches`, {
        method: 'PUT',
        body: { branch_ids: editForm.branch_ids },
      })
    }
    const currentRole = editing.value.roles.find((role) => role !== 'owner') ?? ''
    if (can('staff.roles.manage') && editForm.role_key && editForm.role_key !== currentRole) {
      await api(`/companies/me/members/${editing.value.id}/role`, {
        method: 'PATCH',
        body: { role_key: editForm.role_key },
      })
    }
    notify('Staff member updated.')
    editing.value = null
    await fetchData()
  } catch (err) {
    editError.value = apiErrorMessage(err, 'Could not update this staff member.')
  } finally {
    editingSaving.value = false
  }
}

async function removeMember(member: StaffMember) {
  const approved = await confirm({
    title: 'Remove staff member?',
    message: `${member.full_name} will lose access to this company immediately. Their branch assignments are removed.`,
    confirmLabel: 'Remove',
    danger: true,
  })
  if (!approved) return
  try {
    await api(`/companies/me/members/${member.id}`, { method: 'DELETE' })
    notify('Staff member removed.')
    await fetchData()
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not remove this staff member.')
  }
}

async function resend(invitation: CompanyInvitation) {
  try {
    const response = await api<{ dev_invite_token?: string }>(
      `/companies/me/invitations/${invitation.id}/resend`,
      { method: 'POST', body: {} },
    )
    devInviteToken.value = response.dev_invite_token ?? devInviteToken.value
    notify('Invitation resent with a fresh link.')
    await fetchData()
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not resend the invitation.')
  }
}

async function revoke(invitation: CompanyInvitation) {
  const approved = await confirm({
    title: 'Revoke invitation?',
    message: `The invitation for ${invitation.email} will stop working immediately.`,
    confirmLabel: 'Revoke',
    danger: true,
  })
  if (!approved) return
  try {
    await api(`/companies/me/invitations/${invitation.id}`, { method: 'DELETE' })
    notify('Invitation revoked.')
    await fetchData()
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not revoke the invitation.')
  }
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Company management</p>
        <h1 class="mt-2 text-[clamp(1.6rem,3vw,2.1rem)] leading-tight font-semibold tracking-[-0.035em]">Staff</h1>
        <p class="mt-2 max-w-2xl text-[13.5px] leading-relaxed text-ink-900/50">
          Invite team members, assign them to branches and keep roles precise. Suspended members lose
          access immediately; invitations expire after seven days.
        </p>
      </div>
      <div class="flex items-center gap-2">
        <LookupRefreshButton :loading="loading" @refresh="fetchData" />
        <button
          v-if="can('staff.invite')"
          type="button"
          class="inline-flex items-center gap-1.5 rounded-lg bg-brand-500 px-3.5 py-2.5 text-[12.5px] font-semibold text-white transition hover:bg-brand-600"
          @click="inviteOpen = true"
        >
          <UserPlus :size="14" /> Invite staff
        </button>
      </div>
    </div>

    <p v-if="error" role="alert" class="mt-5 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <div class="mt-5 flex flex-wrap items-center gap-2">
      <label class="relative min-w-[200px] flex-1">
        <Search :size="14" class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-ink-900/30" />
        <input v-model="search" type="search" class="w-full rounded-lg border border-slate-300 bg-white py-2.5 pr-3 pl-8 text-[12.5px] focus:border-brand-500" placeholder="Search name, email or title" />
      </label>
      <label class="flex items-center gap-1.5 rounded-lg border border-slate-300 bg-white px-2.5 py-2 text-[11.5px]">
        <SlidersHorizontal :size="13" class="text-ink-900/35" />
        <select v-model="status" class="bg-transparent text-[11.5px] outline-none">
          <option value="all">All statuses</option>
          <option value="active">Active</option>
          <option value="suspended">Suspended</option>
        </select>
      </label>
      <label class="rounded-lg border border-slate-300 bg-white px-2.5 py-2">
        <select v-model="branchFilter" class="bg-transparent text-[11.5px] outline-none">
          <option value="all">All branches</option>
          <option value="none">No branch assigned</option>
          <option v-for="branch in branches" :key="branch.id" :value="branch.id">{{ branch.name }}</option>
        </select>
      </label>
      <span class="text-[10.5px] text-ink-900/35">{{ filtered.length }} shown</span>
    </div>

    <div v-if="loading && !members.length" class="mt-5 flex items-center gap-2 rounded-xl border border-ink-900/10 bg-white p-6 text-[12.5px] text-ink-900/45">
      <Loader2 :size="15" class="animate-spin" /> Loading staff…
    </div>

    <StaffTable
      v-else
      class="mt-5"
      :members="filtered"
      :branches="branches"
      :can-update="can('staff.update')"
      :can-deactivate="can('staff.deactivate')"
      :can-roles="can('staff.roles.manage')"
      @edit="openEdit"
      @role="openEdit"
      @remove="removeMember"
    />

    <div v-if="devInviteToken" class="mt-4 rounded-xl border border-amber-200 bg-amber-50 p-4">
      <p class="text-[11.5px] font-semibold text-amber-900">Development invitation link</p>
      <p class="mt-1 text-[10.5px] leading-relaxed text-amber-800/80">
        Email delivery is logged in development. Share this one-time link with the invited staff member.
      </p>
      <NuxtLink :to="`/invitations/${devInviteToken}`" class="mt-2 inline-block break-all text-[11px] font-semibold text-amber-900 underline underline-offset-2">
        /invitations/{{ devInviteToken }}
      </NuxtLink>
    </div>

    <div v-if="can('staff.invite')" class="mt-6 rounded-xl border border-ink-900/10 bg-white">
      <div class="flex items-center justify-between gap-3 border-b border-ink-900/8 px-4 py-3.5">
        <div>
          <h2 class="text-[13.5px] font-semibold">Pending invitations</h2>
          <p class="mt-0.5 text-[10.5px] text-ink-900/40">Unaccepted links expire automatically.</p>
        </div>
        <span class="text-[10.5px] text-ink-900/40">{{ invitations.length }}</span>
      </div>

      <div v-if="invitations.length === 0" class="px-4 py-8 text-center text-[11.5px] text-ink-900/40">
        No pending invitations.
      </div>
      <ul v-else class="divide-y divide-ink-900/8">
        <li v-for="invitation in invitations" :key="invitation.id" class="flex flex-wrap items-center gap-2 px-4 py-3">
          <div class="min-w-0 flex-1">
            <p class="truncate text-[12px] font-medium">{{ invitation.email }}</p>
            <p class="mt-0.5 text-[10px] text-ink-900/40">
              {{ roleLabel(invitation.role_key) }}
              <span v-if="invitation.branch_ids.length"> · {{ branchNames(invitation.branch_ids) }}</span>
              · expires {{ new Date(invitation.expires_at).toLocaleDateString() }}
            </p>
          </div>
          <button type="button" class="rounded-md border border-slate-300 px-2 py-1 text-[10.5px] font-medium transition hover:bg-slate-50" @click="resend(invitation)">
            Resend
          </button>
          <button type="button" class="rounded-md border border-red-200 px-2 py-1 text-[10.5px] font-medium text-red-600 transition hover:bg-red-50" @click="revoke(invitation)">
            Revoke
          </button>
        </li>
      </ul>
    </div>

    <StaffInvitationDialog
      :open="inviteOpen"
      :roles="roles"
      :branches="branches"
      :saving="inviteSaving"
      @close="inviteOpen = false"
      @invite="invite"
    />

    <Teleport to="body">
      <div
        v-if="editing"
        class="fixed inset-0 z-[80] flex items-center justify-center bg-ink-950/45 px-4 py-6 backdrop-blur-sm"
        role="dialog"
        aria-modal="true"
        aria-label="Edit staff member"
        @click.self="editing = null"
      >
        <form class="max-h-full w-full max-w-lg overflow-y-auto rounded-2xl border border-ink-900/10 bg-white p-5 shadow-2xl" @submit.prevent="saveMember">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h2 class="text-[16px] font-semibold">Edit staff member</h2>
              <p class="mt-1 text-[11.5px] text-ink-900/45">{{ editing.full_name }} · {{ editing.email }}</p>
            </div>
            <button type="button" class="text-ink-900/35 hover:text-ink-900" aria-label="Close" @click="editing = null">
              <X :size="16" />
            </button>
          </div>

          <div class="mt-4 space-y-3.5">
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Title</span>
              <input v-model="editForm.title" maxlength="80" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" placeholder="Branch supervisor" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Status</span>
              <select v-model="editForm.status" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500">
                <option value="active">Active</option>
                <option value="suspended">Suspended</option>
              </select>
              <span class="mt-1 block text-[10px] text-ink-900/35">Suspending ends their sessions and removes branch assignments.</span>
            </label>
            <label v-if="can('staff.roles.manage')" class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Role</span>
              <select v-model="editForm.role_key" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500">
                <option v-for="role in roles.filter((item) => item.key !== 'owner')" :key="role.id" :value="role.key">
                  {{ role.name }}{{ role.is_system ? '' : ' (custom)' }}
                </option>
              </select>
            </label>

            <div v-if="editForm.status === 'active'">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Branch assignments</span>
              <div v-if="branches.length" class="grid gap-1.5 sm:grid-cols-2">
                <label
                  v-for="branch in branches"
                  :key="branch.id"
                  class="flex cursor-pointer items-center gap-2 rounded-lg border px-2.5 py-2 text-[11.5px] transition"
                  :class="editForm.branch_ids.includes(branch.id) ? 'border-brand-200 bg-brand-50' : 'border-ink-900/8 hover:bg-canvas'"
                >
                  <input type="checkbox" class="size-3.5 rounded border-slate-300" :checked="editForm.branch_ids.includes(branch.id)" @change="toggleEditBranch(branch.id)" />
                  <span class="min-w-0 flex-1 truncate">{{ branch.name }}</span>
                </label>
              </div>
              <p v-else class="text-[11px] text-ink-900/40">No branches have been created yet.</p>
            </div>
          </div>

          <p v-if="editError" role="alert" class="mt-4 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
            <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ editError }}
          </p>

          <div class="mt-5 flex justify-end gap-2">
            <button type="button" class="rounded-lg border border-slate-300 px-3.5 py-2 text-[12.5px] font-medium transition hover:bg-slate-50" @click="editing = null">Cancel</button>
            <button type="submit" :disabled="editingSaving" class="inline-flex items-center gap-1.5 rounded-lg bg-brand-500 px-4 py-2.5 text-[12.5px] font-semibold text-white transition hover:bg-brand-600 disabled:opacity-60">
              <Loader2 v-if="editingSaving" :size="14" class="animate-spin" />
              <Send v-else :size="14" />
              Save changes
            </button>
          </div>
        </form>
      </div>
    </Teleport>
  </div>
</template>
