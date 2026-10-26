<script setup lang="ts">
import { ApiError } from '~/utils/api'
import type { BranchDetails, BranchOptions } from '~/types/staff'

definePageMeta({ layout: 'default' })

useHead({ title: 'Branches' })

const { token } = useAuth()

// ---- data sources ----
const branches = ref<BranchDetails[]>([])
const loading = ref(false)

// ---- add / edit modal state ----
const editing = ref<BranchDetails | null>(null)
const adding = ref(false)
const saving = ref(false)

const addName = ref('')
const addCode = ref('')
const addAddress = ref('')
const addState = ref('')
const addPhone = ref('')

// ---- menu state ----
const menuFor = ref<string | null>(null)

// ---- confirm modal ----
const confirmModal = ref<{ mode: 'delete'; branchId: string; branchName: string } | null>(null)

// ---- refresh loop ----
onMounted(refresh)
watch(token, () => { if (token.value) refresh() })

async function refresh() {
  if (!token.value) return
  loading.value = true
  try {
    branches.value = await fetchBranches()
  } finally {
    loading.value = false
  }
}

async function fetchBranches(): Promise<BranchDetails[]> {
  const rows = await apiRequest<Array<{ id: string; name: string; code: string; address?: string; state?: string; phone?: string; status: string }>>(
    '/api/branches',
    { token: token.value },
  )
  return rows.map(normalizeBranch)
}

function normalizeBranch(row: { id: string; name: string; code: string; address?: string; state?: string; phone?: string; status: string }): BranchDetails {
  return {
    id: row.id,
    name: row.name,
    code: row.code,
    address: row.address ?? undefined,
    state: row.state ?? undefined,
    phone: row.phone ?? undefined,
    status: row.status,
  }
}

// ---- add branch ----
function openAdd() {
  editing.value = null
  adding.value = true
  addName.value = ''
  addCode.value = ''
  addAddress.value = ''
  addState.value = ''
  addPhone.value = ''
}

async function saveAdd() {
  if (!addName.value.trim() || !addCode.value.trim()) return
  saving.value = true
  try {
    await apiRequest('/api/branches', {
      method: 'POST',
      token: token.value,
      body: {
        name: addName.value.trim(),
        code: addCode.value.trim().toUpperCase(),
        address: addAddress.value.trim(),
        state: addState.value.trim(),
        phone: addPhone.value.trim(),
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

// ---- edit branch ----
function openEdit(row: BranchDetails) {
  editing.value = row
  addName.value = row.name
  addCode.value = row.code
  addAddress.value = row.address ?? ''
  addState.value = row.state ?? ''
  addPhone.value = row.phone ?? ''
  adding.value = false
}

async function saveEdit() {
  if (!editing.value) return
  saving.value = true
  try {
    await apiRequest(`/api/branches/${editing.value.id}`, {
      method: 'PUT',
      token: token.value,
      body: {
        name: addName.value.trim(),
        code: addCode.value.trim().toUpperCase(),
        address: addAddress.value.trim(),
        state: addState.value.trim(),
        phone: addPhone.value.trim(),
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
async function confirmDelete(row: BranchDetails) {
  menuFor.value = null
  confirmModal.value = {
    mode: 'delete',
    branchId: row.id,
    branchName: row.name,
  }
}

async function executeDelete() {
  if (!confirmModal.value) return
  saving.value = true
  try {
    await apiRequest(`/api/branches/${confirmModal.value.branchId}`, {
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
            class="-(m-1) flex h-11 flex-1 items-center justify-center rounded-lg px-3 text-button font-medium transition duration-150 hover:bg-canvas/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
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
            class="-(m-1) flex h-11 flex-1 items-center justify-center rounded-lg px-3 text-button font-medium transition duration-150 last:rounded-tr-lg hover:bg-canvas/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
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
            <h1 class="text-h3 text-ink">Branches</h1>
            <p class="mt-1 text-label text-ink-64">
              Manage warehouse and office locations.
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
            Add branch
            <UiKbd>Enter</UiKbd>
          </UiButton>
        </div>

        <!-- branches list -->
        <div class="divide-y divide-line">
          <div
            v-for="row in branches"
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
                  <span class="truncate font-medium text-p1 text-ink">{{ row.name }}</span>
                  <UiPill v-if="row.code" class="shrink-0">
                    {{ row.code }}
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
                {{ row.address ? `${row.address}${row.state ? `, ${row.state}` : ''}` : row.state ?? 'No address' }}
              </p>
              <p v-if="row.phone" class="mt-0.5 truncate text-label text-ink-64">
                {{ row.phone }}
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
                Delete branch
              </button>
            </div>
          </div>

          <div v-if="branches.length === 0 && !loading" class="py-8 text-center text-label text-ink-64">
            No branches found.
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
            <h2 class="text-h5 text-ink">{{ editing ? `Edit ${editing.name}` : 'Add branch' }}</h2>
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
              id="branch-name"
              v-model="addName"
              label="Branch name"
              type="text"
              :error="''"
              autofocus
            />
            <UiField
              id="branch-code"
              v-model="addCode"
              label="Branch code"
              type="text"
              :error="''"
              hint="e.g. LAG-01"
            />
            <UiField
              id="branch-address"
              v-model="addAddress"
              label="Address"
              type="text"
              :error="''"
            />
            <UiField
              id="branch-state"
              v-model="addState"
              label="State"
              type="text"
              :error="''"
            />
            <UiField
              id="branch-phone"
              v-model="addPhone"
              label="Phone"
              type="tel"
              :error="''"
            />

            <div class="mt-5 flex items-center gap-2">
              <UiButton
                type="button"
                :loading="saving"
                @click="editing ? saveEdit() : saveAdd()"
              >
                {{ editing ? 'Save' : 'Add branch' }}
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
            <h2 class="text-h5 text-ink">Delete {{ confirmModal.branchName }}?</h2>
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
            This branch will be marked inactive. Its data stays in the records.
          </p>

          <div class="mt-5 flex items-center gap-2">
            <UiButton
              type="button"
              :loading="saving"
              variant="danger"
              @click="executeDelete"
            >
              Delete branch
              <UiKbd>Enter</UiKbd>
            </UiButton>
          </div>
        </UiCard>
      </div>
    </Teleport>
  </div>
</template>
