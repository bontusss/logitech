<script setup lang="ts">
import type { Branch } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Branches' })
definePageMeta({ layout: 'default', roles: ['owner', 'system_admin'] })

const { token } = useAuth()
const { branches, load: loadBranches, loading: branchesLoading } = useBranches()

const loading = computed(() => branchesLoading.value && !branches.value.length)
const error = ref('')
const notice = ref('')

const formOpen = ref(false)
const editingId = ref('')
const form = reactive({ name: '', code: '', address: '', state: '', phone: '' })
const saving = ref(false)
const formError = ref('')

const deleteTarget = ref<Branch | null>(null)
const deleting = ref(false)
const deleteError = ref('')
const busyId = ref('')

async function load() {
  await loadBranches(true)
}

function openCreate() {
  editingId.value = ''
  form.name = ''
  form.code = ''
  form.address = ''
  form.state = ''
  form.phone = ''
  formError.value = ''
  formOpen.value = true
}

function openEdit(branch: Branch) {
  editingId.value = branch.id
  form.name = branch.name
  form.code = branch.code
  form.address = branch.address ?? ''
  form.state = branch.state ?? ''
  form.phone = branch.phone ?? ''
  formError.value = ''
  formOpen.value = true
}

async function save() {
  if (!token.value) return
  formError.value = ''
  if (!form.name.trim() || !form.code.trim()) {
    formError.value = 'Name and code are required.'
    return
  }
  saving.value = true
  try {
    if (editingId.value) {
      await apiRequest(`/api/branches/${editingId.value}`, {
        method: 'PUT',
        token: token.value,
        body: {
          name: form.name.trim(),
          address: form.address.trim(),
          state: form.state.trim(),
          phone: form.phone.trim(),
        },
      })
      notice.value = 'Branch updated.'
    } else {
      await apiRequest('/api/branches', {
        method: 'POST',
        token: token.value,
        body: {
          name: form.name.trim(),
          code: form.code.trim().toUpperCase(),
          address: form.address.trim() || undefined,
          state: form.state.trim() || undefined,
          phone: form.phone.trim() || undefined,
        },
      })
      notice.value = 'Branch created.'
    }
    formOpen.value = false
    await load()
  } catch (err) {
    formError.value = err instanceof ApiError ? err.message : 'Could not save the branch.'
  } finally {
    saving.value = false
  }
}

async function toggleStatus(branch: Branch) {
  if (!token.value) return
  busyId.value = branch.id
  notice.value = ''
  try {
    await apiRequest(`/api/branches/${branch.id}`, {
      method: 'PUT',
      token: token.value,
      body: { status: branch.status === 'active' ? 'inactive' : 'active' },
    })
    notice.value = `${branch.name} is now ${branch.status === 'active' ? 'inactive' : 'active'}.`
    await load()
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not update the branch.'
  } finally {
    busyId.value = ''
  }
}

async function remove() {
  if (!token.value || !deleteTarget.value) return
  deleting.value = true
  deleteError.value = ''
  try {
    await apiRequest(`/api/branches/${deleteTarget.value.id}`, { method: 'DELETE', token: token.value })
    notice.value = `${deleteTarget.value.name} was removed.`
    deleteTarget.value = null
    await load()
  } catch (err) {
    deleteError.value = err instanceof ApiError ? err.message : 'Could not delete the branch.'
  } finally {
    deleting.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <header class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-h5 font-medium tracking-tight text-ink">Branches</h1>
        <p class="mt-1 text-label text-ink-64">Where parcels are dropped off and collected.</p>
      </div>
      <UiButton @click="openCreate">
        <UiIcon name="plus" :size="17" />
        New branch
      </UiButton>
    </header>

    <div v-if="notice" class="rounded-lg border border-blue/25 bg-blue/5 px-4 py-3 text-label text-ink" role="status">{{ notice }}</div>
    <div v-if="error" class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-danger/25 bg-danger/5 px-4 py-3">
      <p class="text-label text-danger">{{ error }}</p>
      <UiButton variant="tertiary" :loading="loading" @click="load">Try again</UiButton>
    </div>

    <div v-if="loading" class="grid gap-3 sm:grid-cols-2" aria-hidden="true">
      <div v-for="index in 4" :key="index" class="h-[132px] animate-pulse rounded-lg bg-surface shadow-card" />
    </div>

    <template v-else>
      <div v-if="branches.length" class="grid gap-3 sm:grid-cols-2">
        <article v-for="branch in branches" :key="branch.id" class="rounded-lg bg-surface p-4 shadow-card">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="text-meta tracking-wide text-ink-40">{{ branch.code }}</p>
              <h2 class="mt-0.5 truncate text-p1 font-medium text-ink">{{ branch.name }}</h2>
            </div>
            <UiPill :tone="branch.status === 'active' ? 'info' : 'neutral'">
              {{ branch.status === 'active' ? 'Active' : 'Inactive' }}
            </UiPill>
          </div>

          <dl class="mt-3 space-y-1 text-label">
            <div class="flex gap-2">
              <dt class="w-16 shrink-0 text-ink-40">Address</dt>
              <dd class="min-w-0 text-ink">{{ branch.address || '—' }}</dd>
            </div>
            <div class="flex gap-2">
              <dt class="w-16 shrink-0 text-ink-40">State</dt>
              <dd class="min-w-0 text-ink">{{ branch.state || '—' }}</dd>
            </div>
            <div class="flex gap-2">
              <dt class="w-16 shrink-0 text-ink-40">Phone</dt>
              <dd class="min-w-0 text-ink">{{ branch.phone || '—' }}</dd>
            </div>
          </dl>

          <div class="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 border-t border-line pt-3">
            <UiButton variant="link" @click="openEdit(branch)">Edit</UiButton>
            <UiButton variant="link" :loading="busyId === branch.id" @click="toggleStatus(branch)">
              {{ branch.status === 'active' ? 'Deactivate' : 'Activate' }}
            </UiButton>
            <UiButton variant="link" class="text-danger" @click="deleteTarget = branch">Delete</UiButton>
          </div>
        </article>
      </div>

      <UiEmptyState v-else icon="building" title="No branches yet" description="Create the first branch to start recording waybills.">
        <UiButton @click="openCreate">New branch</UiButton>
      </UiEmptyState>
    </template>

    <!-- Create / edit -->
    <UiModal
      :open="formOpen"
      :title="editingId ? 'Edit branch' : 'New branch'"
      :description="editingId ? 'The code is permanent; the rest can change.' : 'The code prefixes waybill numbers, e.g. LAG-01.'"
      size="md"
      :dismissible="!saving"
      @close="formOpen = false"
    >
      <div v-if="formError" class="mb-3 rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
        {{ formError }}
      </div>
      <div class="space-y-4">
        <div class="grid gap-4 sm:grid-cols-2">
          <UiField id="branch-name" v-model="form.name" label="Name" autocomplete="off" />
          <UiField
            id="branch-code"
            v-model="form.code"
            label="Code"
            :disabled="Boolean(editingId)"
            autocomplete="off"
          />
        </div>
        <UiField id="branch-address" v-model="form.address" label="Address (optional)" autocomplete="off" />
        <div class="grid gap-4 sm:grid-cols-2">
          <UiField id="branch-state" v-model="form.state" label="State (optional)" autocomplete="off" />
          <UiField id="branch-phone" v-model="form.phone" label="Phone (optional)" type="tel" autocomplete="off" />
        </div>
      </div>
      <template #footer>
        <div class="flex flex-col gap-2 sm:flex-row-reverse">
          <UiButton block :loading="saving" @click="save">{{ editingId ? 'Save changes' : 'Create branch' }}</UiButton>
          <UiButton variant="tertiary" block :disabled="saving" @click="formOpen = false">Cancel</UiButton>
        </div>
      </template>
    </UiModal>

    <!-- Delete -->
    <UiModal
      :open="Boolean(deleteTarget)"
      title="Delete branch"
      description="Only an empty branch can be deleted; deactivate one that has history instead."
      :dismissible="!deleting"
      @close="deleteTarget = null"
    >
      <div v-if="deleteError" class="mb-3 rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
        {{ deleteError }}
      </div>
      <p class="text-label text-ink-64">
        Delete <span class="font-medium text-ink">{{ deleteTarget?.name }}</span> ({{ deleteTarget?.code }})?
      </p>
      <template #footer>
        <div class="flex flex-col gap-2 sm:flex-row-reverse">
          <UiButton variant="danger" block :loading="deleting" @click="remove">Delete branch</UiButton>
          <UiButton variant="tertiary" block :disabled="deleting" @click="deleteTarget = null">Keep it</UiButton>
        </div>
      </template>
    </UiModal>
  </div>
</template>
