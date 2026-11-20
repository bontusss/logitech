<script setup lang="ts">
import { ArrowLeft, CircleAlert, Loader2 } from 'lucide-vue-next'
import type { Branch, BranchInput, BranchStaff, StaffMember } from '~/types/auth'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Branch details' })

const route = useRoute()
const { api } = useApi()
const { load, can, notify, confirm } = useWorkspace()

const branchID = computed(() => String(route.params.id))
const branch = ref<Branch | null>(null)
const staff = ref<BranchStaff[]>([])
const members = ref<StaffMember[]>([])
const loading = ref(true)
const error = ref('')
const editing = ref(false)
const saving = ref(false)

onMounted(async () => {
  await load()
  await fetchBranch()
})

async function fetchBranch() {
  loading.value = true
  error.value = ''
  try {
    const data = await api<{ branch: Branch; staff: BranchStaff[] }>(`/companies/me/branches/${branchID.value}`)
    branch.value = data.branch
    staff.value = data.staff
    if (can('staff.view')) {
      const data = await api<{ members: StaffMember[] }>('/companies/me/members')
      members.value = data.members
    }
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not load this branch. It may not exist or you may not have access.')
    branch.value = null
  } finally {
    loading.value = false
  }
}

async function saveBranch(payload: BranchInput) {
  saving.value = true
  error.value = ''
  try {
    await api(`/companies/me/branches/${branchID.value}`, { method: 'PATCH', body: payload })
    notify('Branch updated.')
    editing.value = false
    await fetchBranch()
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not update the branch.')
  } finally {
    saving.value = false
  }
}

async function toggleStatus() {
  if (!branch.value) return
  const deactivating = branch.value.status === 'active'
  const approved = await confirm({
    title: deactivating ? 'Deactivate this branch?' : 'Activate this branch?',
    message: deactivating
      ? 'The branch will remain in records but stop accepting new staff assignments.'
      : 'The branch will become available for staff assignments again.',
    confirmLabel: deactivating ? 'Deactivate' : 'Activate',
    danger: deactivating,
  })
  if (!approved) return
  saving.value = true
  try {
    await api(`/companies/me/branches/${branchID.value}/${deactivating ? 'deactivate' : 'activate'}`, {
      method: 'POST',
      body: {},
    })
    notify(deactivating ? 'Branch deactivated.' : 'Branch activated.')
    await fetchBranch()
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not change the branch status.')
  } finally {
    saving.value = false
  }
}

async function assignStaff(memberIDs: string[]) {
  saving.value = true
  try {
    await api(`/companies/me/branches/${branchID.value}/staff`, {
      method: 'PUT',
      body: { member_ids: memberIDs },
    })
    notify('Branch staff assignments saved.')
    await fetchBranch()
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not save staff assignments.')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="dash-page">
    <NuxtLink
      to="/dashboard/company/branches"
      class="inline-flex items-center gap-1.5 rounded-full border border-ink-900/10 bg-white/80 px-3 py-1.5 text-[11.5px] font-medium text-ink-900/50 shadow-sm transition hover:-translate-x-0.5 hover:text-ink-900 dark:bg-white/5"
    >
      <ArrowLeft :size="13" /> All branches
    </NuxtLink>

    <div v-if="loading" class="dash-panel mt-4 flex items-center gap-2 rounded-2xl border p-6 text-[12.5px] text-ink-900/45">
      <Loader2 :size="15" class="animate-spin" /> Loading branch…
    </div>

    <p v-else-if="error && !branch" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 p-4 text-[12.5px] text-red-700">
      <CircleAlert :size="15" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <template v-else-if="branch">
      <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
        <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ error }}
      </p>

      <div class="mt-4">
        <BranchForm
          v-if="editing"
          :branch="branch"
          :members="members"
          :saving="saving"
          @submit="saveBranch"
          @cancel="editing = false"
        />
        <BranchDetails
          v-else
          :branch="branch"
          :staff="staff"
          :all-members="members"
          :can-update="can('branches.update')"
          :can-deactivate="can('branches.deactivate')"
          :can-assign="can('staff.update')"
          :saving="saving"
          @edit="editing = true"
          @toggle="toggleStatus"
          @assign="assignStaff"
        />
      </div>
    </template>
  </div>
</template>
