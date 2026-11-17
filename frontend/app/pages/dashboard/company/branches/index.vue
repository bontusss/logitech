<script setup lang="ts">
import { CircleAlert, Loader2, Plus, Search, SlidersHorizontal } from 'lucide-vue-next'
import type { Branch, BranchInput, StaffMember } from '~/types/auth'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Branches' })

const { api } = useApi()
const { detail, load, can, notify, confirm } = useWorkspace()

const branches = ref<Branch[]>([])
const members = ref<StaffMember[]>([])
const loading = ref(true)
const error = ref('')
const search = ref('')
const status = ref<'all' | 'active' | 'inactive'>('all')
const formOpen = ref(false)
const saving = ref(false)

onMounted(async () => {
  await load()
  await fetchData()
})

async function fetchData() {
  loading.value = true
  error.value = ''
  try {
    const data = await api<{ branches: Branch[] }>('/companies/me/branches')
    branches.value = data.branches
    if (can('staff.view')) {
      const staff = await api<{ members: StaffMember[] }>('/companies/me/members')
      members.value = staff.members
    }
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not load branches.')
  } finally {
    loading.value = false
  }
}

const filtered = computed(() => {
  const term = search.value.trim().toLowerCase()
  return branches.value.filter((branch) => {
    const matchesStatus = status.value === 'all' || branch.status === status.value
    const matchesSearch = !term ||
      branch.name.toLowerCase().includes(term) ||
      branch.code.toLowerCase().includes(term) ||
      branch.city.toLowerCase().includes(term) ||
      branch.state.toLowerCase().includes(term)
    return matchesStatus && matchesSearch
  })
})

async function createBranch(payload: BranchInput) {
  saving.value = true
  error.value = ''
  try {
    await api('/companies/me/branches', { method: 'POST', body: payload })
    notify('Branch created.')
    formOpen.value = false
    await fetchData()
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not create the branch.')
  } finally {
    saving.value = false
  }
}

async function toggleBranch(branch: Branch) {
  const deactivating = branch.status === 'active'
  const approved = await confirm({
    title: deactivating ? 'Deactivate branch?' : 'Activate branch?',
    message: deactivating
      ? `${branch.name} will stop accepting new assignments. Existing records are kept and the branch can be reactivated.`
      : `${branch.name} will become available for assignments again.`,
    confirmLabel: deactivating ? 'Deactivate' : 'Activate',
    danger: deactivating,
  })
  if (!approved) return
  try {
    await api(`/companies/me/branches/${branch.id}/${deactivating ? 'deactivate' : 'activate'}`, {
      method: 'POST',
      body: {},
    })
    notify(deactivating ? 'Branch deactivated.' : 'Branch activated.')
    await fetchData()
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not change the branch status.')
  }
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Company management</p>
        <h1 class="mt-2 text-[clamp(1.6rem,3vw,2.1rem)] leading-tight font-semibold tracking-[-0.035em]">Branches</h1>
        <p class="mt-2 max-w-2xl text-[13.5px] leading-relaxed text-ink-900/50">
          Manage locations, assign managers and control which branches each staff member can access.
          Branches are deactivated, never deleted, so their records remain intact.
        </p>
      </div>
      <div class="flex items-center gap-2">
        <LookupRefreshButton :loading="loading" @refresh="fetchData" />
        <button
          v-if="can('branches.create')"
          type="button"
          class="inline-flex items-center gap-1.5 rounded-lg bg-brand-500 px-3.5 py-2.5 text-[12.5px] font-semibold text-white transition hover:bg-brand-600"
          @click="formOpen = true"
        >
          <Plus :size="14" /> Add branch
        </button>
      </div>
    </div>

    <p v-if="error" role="alert" class="mt-5 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <BranchForm
      v-if="formOpen"
      class="mt-5"
      :members="members"
      :saving="saving"
      @submit="createBranch"
      @cancel="formOpen = false"
    />

    <div class="mt-5 flex flex-wrap items-center gap-2">
      <label class="relative min-w-[200px] flex-1">
        <Search :size="14" class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-ink-900/30" />
        <input
          v-model="search"
          type="search"
          class="w-full rounded-lg border border-slate-300 bg-white py-2.5 pr-3 pl-8 text-[12.5px] focus:border-brand-500"
          placeholder="Search by name, code, city or state"
        />
      </label>
      <label class="flex items-center gap-1.5 rounded-lg border border-slate-300 bg-white px-2.5 py-2 text-[11.5px]">
        <SlidersHorizontal :size="13" class="text-ink-900/35" />
        <select v-model="status" class="bg-transparent text-[11.5px] outline-none">
          <option value="all">All statuses</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>
      </label>
      <span class="text-[10.5px] text-ink-900/35">{{ filtered.length }} shown</span>
    </div>

    <div v-if="loading && !branches.length" class="mt-5 flex items-center gap-2 rounded-xl border border-ink-900/10 bg-white p-6 text-[12.5px] text-ink-900/45">
      <Loader2 :size="15" class="animate-spin" /> Loading branches…
    </div>

    <BranchList
      v-else
      class="mt-5"
      :branches="filtered"
      :members="members"
      :loading="loading"
      :can-create="can('branches.create')"
      :can-update="can('branches.update')"
      :can-deactivate="can('branches.deactivate')"
      @create="formOpen = true"
      @toggle="toggleBranch"
    />

    <div v-if="detail && !can('branches.create')" class="mt-4 rounded-xl border border-ink-900/10 bg-canvas p-4 text-[11px] leading-relaxed text-ink-900/45">
      Your role can view and operate assigned branches. Creating, deactivating or reassigning branches
      requires an administrator permission.
    </div>
  </div>
</template>
