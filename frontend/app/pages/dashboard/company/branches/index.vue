<script setup lang="ts">
import {
  Building2,
  CheckCircle2,
  CircleAlert,
  Loader2,
  Plus,
  Power,
  Search,
  SlidersHorizontal,
  Sparkles,
} from 'lucide-vue-next'
import type { Branch, BranchInput, StaffMember } from '~/types/auth'
import { parseProviderBusinessTypes } from '~/utils/providerBusinessTypes'

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

const lastMileOnly = computed(() => {
  const types = parseProviderBusinessTypes(detail.value?.company.business_type ?? '')
  return types.length === 1 && types[0] === 'lastmile'
})

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
  <div class="dash-page">
    <header class="dash-hero relative overflow-hidden rounded-2xl border p-5 sm:p-6">
      <div class="pointer-events-none absolute -top-16 right-10 size-44 rounded-full bg-accent-400/15 blur-3xl" />
      <div class="pointer-events-none absolute -right-8 -bottom-24 size-52 rounded-full bg-brand-400/15 blur-3xl" />
      <div class="relative flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="dash-eyebrow inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-semibold tracking-wide">
            <Sparkles :size="12" /> YOUR FOOTPRINT
          </p>
          <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Branches</h1>
          <p class="mt-1.5 max-w-xl text-[12px] leading-relaxed text-ink-900/65">
            Manage locations, assign managers and control which branches each staff member can
            access. Branches are deactivated, never deleted, so records stay intact.
          </p>
        </div>
        <div class="flex items-center gap-2">
          <LookupRefreshButton :loading="loading" @refresh="fetchData" />
          <button
            v-if="can('branches.create') && !lastMileOnly"
            type="button"
            class="inline-flex min-h-10 items-center gap-1.5 rounded-xl bg-accent-500 px-4 py-2.5 text-[12px] font-semibold text-ink-950 shadow-sm shadow-accent-500/20 transition hover:-translate-y-0.5 hover:bg-accent-400"
            @click="formOpen = true"
          >
            <Plus :size="14" /> Add branch
          </button>
        </div>
      </div>
    </header>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <section class="mt-4 grid gap-3 sm:grid-cols-3" aria-label="Branch summary">
      <article class="dash-stat dash-stat--brand rounded-xl border p-4">
        <div class="flex items-center justify-between gap-2">
          <p class="text-[10px] font-semibold tracking-wide text-ink-900/50 uppercase">Total branches</p>
          <span class="flex size-6 items-center justify-center rounded-lg bg-white/80 text-brand-700 dark:bg-white/10 dark:text-brand-200"><Building2 :size="12" /></span>
        </div>
        <p class="mt-2 text-[22px] leading-none font-semibold tracking-tight">{{ branches.length }}</p>
      </article>
      <article class="dash-stat dash-stat--success rounded-xl border p-4">
        <div class="flex items-center justify-between gap-2">
          <p class="text-[10px] font-semibold tracking-wide text-ink-900/50 uppercase">Active</p>
          <span class="flex size-6 items-center justify-center rounded-lg bg-white/80 text-success-700 dark:bg-white/10 dark:text-success-200"><CheckCircle2 :size="12" /></span>
        </div>
        <p class="mt-2 text-[22px] leading-none font-semibold tracking-tight">
          {{ branches.filter((branch) => branch.status === 'active').length }}
        </p>
        <p class="mt-1 text-[10px] text-ink-900/45">Accepting assignments</p>
      </article>
      <article class="dash-stat dash-stat--accent rounded-xl border p-4">
        <div class="flex items-center justify-between gap-2">
          <p class="text-[10px] font-semibold tracking-wide text-ink-900/50 uppercase">Inactive</p>
          <span class="flex size-6 items-center justify-center rounded-lg bg-white/80 text-accent-700 dark:bg-white/10 dark:text-accent-200"><Power :size="12" /></span>
        </div>
        <p class="mt-2 text-[22px] leading-none font-semibold tracking-tight">
          {{ branches.filter((branch) => branch.status === 'inactive').length }}
        </p>
        <p class="mt-1 text-[10px] text-ink-900/45">Paused but kept on record</p>
      </article>
    </section>

    <BranchForm
      v-if="formOpen && !lastMileOnly"
      class="mt-4"
      :members="members"
      :saving="saving"
      @submit="createBranch"
      @cancel="formOpen = false"
    />

    <div class="mt-4 flex flex-wrap items-center gap-2">
      <label class="relative min-w-[200px] flex-1">
        <Search :size="14" class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-ink-900/30" />
        <input
          v-model="search"
          type="search"
          class="w-full rounded-xl border border-ink-900/10 bg-white py-2.5 pr-3 pl-8 text-[12.5px] shadow-sm outline-none focus:border-brand-500 dark:bg-white/5"
          placeholder="Search by name, code, city or state"
        />
      </label>
      <label class="flex items-center gap-1.5 rounded-xl border border-ink-900/10 bg-white px-2.5 py-2 text-[11.5px] shadow-sm dark:bg-white/5">
        <SlidersHorizontal :size="13" class="text-ink-900/35" />
        <select v-model="status" class="bg-transparent text-[11.5px] outline-none">
          <option value="all">All statuses</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>
      </label>
      <span class="text-[10.5px] text-ink-900/40">{{ filtered.length }} shown</span>
    </div>

    <div v-if="loading && !branches.length" class="dash-panel mt-4 flex items-center gap-2 rounded-2xl border p-6 text-[12.5px] text-ink-900/45">
      <Loader2 :size="15" class="animate-spin" /> Loading branches…
    </div>

    <BranchList
      v-else
      class="mt-4"
      :branches="filtered"
      :members="members"
      :loading="loading"
      :can-create="can('branches.create') && !lastMileOnly"
      :can-update="can('branches.update')"
      :can-deactivate="can('branches.deactivate')"
      @create="formOpen = true"
      @toggle="toggleBranch"
    />

    <div v-if="detail && !can('branches.create')" class="mt-4 rounded-xl border border-accent-200/70 bg-accent-50/60 p-4 text-[11px] leading-relaxed text-accent-800/70 dark:border-accent-400/20 dark:bg-accent-400/5 dark:text-accent-200">
      Your role can view and operate assigned branches. Creating, deactivating or reassigning branches
      requires an administrator permission.
    </div>
  </div>
</template>
