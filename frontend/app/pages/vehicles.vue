<script setup lang="ts">
import type { Branch, Vehicle } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Vehicles' })
definePageMeta({ layout: 'default', roles: ['owner', 'system_admin'] })

const { token } = useAuth()
const { branches, options: branchOptions, name: branchName, load: loadBranches } = useBranches()

const vehicles = ref<Vehicle[]>([])
const loading = ref(true)
const error = ref('')
const notice = ref('')
const branchFilter = ref('')
const statusFilter = ref('')

const formOpen = ref(false)
const editingId = ref('')
const form = reactive({
  branch: '',
  registration: '',
  type: 'truck',
  makeModel: '',
  capacity: '',
  status: 'available',
})
const saving = ref(false)
const formError = ref('')

const deleteTarget = ref<Vehicle | null>(null)
const deleting = ref(false)

const statusOptions = [
  { value: 'available', label: 'Available' },
  { value: 'on_trip', label: 'On trip' },
  { value: 'maintenance', label: 'Maintenance' },
  { value: 'inactive', label: 'Inactive' },
]

const filtered = computed(() =>
  vehicles.value.filter((vehicle) => {
    if (branchFilter.value && vehicle.branch_id !== branchFilter.value) return false
    if (statusFilter.value && vehicle.status !== statusFilter.value) return false
    return true
  }),
)

function statusTone(status: string) {
  if (status === 'available') return 'info' as const
  if (status === 'on_trip') return 'neutral' as const
  if (status === 'maintenance') return 'danger' as const
  return 'neutral' as const
}

async function load() {
  if (!token.value) return
  loading.value = true
  error.value = ''
  try {
    vehicles.value = await apiRequest<Vehicle[]>('/api/fleet', { token: token.value })
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not load vehicles.'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = ''
  form.branch = branchFilter.value || branches.value[0]?.id || ''
  form.registration = ''
  form.type = 'truck'
  form.makeModel = ''
  form.capacity = ''
  form.status = 'available'
  formError.value = ''
  formOpen.value = true
}

function openEdit(vehicle: Vehicle) {
  editingId.value = vehicle.id
  form.branch = vehicle.branch_id
  form.registration = vehicle.registration_number
  form.type = vehicle.vehicle_type
  form.makeModel = vehicle.make_model ?? ''
  form.capacity = vehicle.capacity_kg ?? ''
  form.status = vehicle.status
  formError.value = ''
  formOpen.value = true
}

async function save() {
  if (!token.value) return
  formError.value = ''
  if (!form.branch) {
    formError.value = 'Choose the branch this vehicle belongs to.'
    return
  }
  if (!editingId.value && !form.registration.trim()) {
    formError.value = 'Enter the registration number.'
    return
  }
  if (!form.type.trim()) {
    formError.value = 'Enter the vehicle type.'
    return
  }

  saving.value = true
  try {
    if (editingId.value) {
      await apiRequest(`/api/fleet/${editingId.value}`, {
        method: 'PUT',
        token: token.value,
        body: {
          branch_id: form.branch,
          vehicle_type: form.type.trim(),
          make_model: form.makeModel.trim(),
          capacity_kg: form.capacity.trim(),
          status: form.status,
        },
      })
      notice.value = 'Vehicle updated.'
    } else {
      await apiRequest('/api/fleet', {
        method: 'POST',
        token: token.value,
        body: {
          branch_id: form.branch,
          registration_number: form.registration.trim().toUpperCase(),
          vehicle_type: form.type.trim(),
          make_model: form.makeModel.trim() || undefined,
          capacity_kg: form.capacity.trim() || undefined,
        },
      })
      notice.value = 'Vehicle added.'
    }
    formOpen.value = false
    await load()
  } catch (err) {
    formError.value = err instanceof ApiError ? err.message : 'Could not save the vehicle.'
  } finally {
    saving.value = false
  }
}

async function remove() {
  if (!token.value || !deleteTarget.value) return
  deleting.value = true
  try {
    await apiRequest(`/api/fleet/${deleteTarget.value.id}`, { method: 'DELETE', token: token.value })
    notice.value = `${deleteTarget.value.registration_number} was removed.`
    deleteTarget.value = null
    await load()
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not delete the vehicle.'
  } finally {
    deleting.value = false
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
        <h1 class="text-h5 font-medium tracking-tight text-ink">Vehicles</h1>
        <p class="mt-1 text-label text-ink-64">The trucks and vans that move parcels between branches.</p>
      </div>
      <UiButton @click="openCreate">
        <UiIcon name="plus" :size="17" />
        Add vehicle
      </UiButton>
    </header>

    <div v-if="notice" class="rounded-lg border border-blue/25 bg-blue/5 px-4 py-3 text-label text-ink" role="status">{{ notice }}</div>
    <div v-if="error" class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-danger/25 bg-danger/5 px-4 py-3">
      <p class="text-label text-danger">{{ error }}</p>
      <UiButton variant="tertiary" :loading="loading" @click="load">Try again</UiButton>
    </div>

    <div class="flex flex-wrap gap-2">
      <UiSelect
        id="vehicle-branch"
        v-model="branchFilter"
        label="Branch filter"
        hide-label
        :options="[{ value: '', label: 'All branches' }, ...branchOptions]"
        class="w-[210px]"
      />
      <UiSelect
        id="vehicle-status"
        v-model="statusFilter"
        label="Status filter"
        hide-label
        :options="[{ value: '', label: 'Any status' }, ...statusOptions]"
        class="w-[170px]"
      />
    </div>

    <div v-if="loading" class="grid gap-3 sm:grid-cols-2" aria-hidden="true">
      <div v-for="index in 4" :key="index" class="h-[124px] animate-pulse rounded-lg bg-surface shadow-card" />
    </div>

    <template v-else>
      <div v-if="filtered.length" class="grid gap-3 sm:grid-cols-2">
        <article v-for="vehicle in filtered" :key="vehicle.id" class="rounded-lg bg-surface p-4 shadow-card">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="text-meta tracking-wide text-ink-40">{{ humanise(vehicle.vehicle_type) }}</p>
              <h2 class="mt-0.5 truncate text-p1 font-medium text-ink">{{ vehicle.registration_number }}</h2>
            </div>
            <UiPill :tone="statusTone(vehicle.status)">{{ humanise(vehicle.status) }}</UiPill>
          </div>

          <dl class="mt-3 space-y-1 text-label">
            <div class="flex gap-2">
              <dt class="w-20 shrink-0 text-ink-40">Model</dt>
              <dd class="min-w-0 text-ink">{{ vehicle.make_model || '—' }}</dd>
            </div>
            <div class="flex gap-2">
              <dt class="w-20 shrink-0 text-ink-40">Capacity</dt>
              <dd class="min-w-0 text-ink">{{ vehicle.capacity_kg ? `${vehicle.capacity_kg} kg` : '—' }}</dd>
            </div>
            <div class="flex gap-2">
              <dt class="w-20 shrink-0 text-ink-40">Branch</dt>
              <dd class="min-w-0 text-ink">{{ branchName(vehicle.branch_id) || '—' }}</dd>
            </div>
          </dl>

          <div class="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 border-t border-line pt-3">
            <UiButton variant="link" @click="openEdit(vehicle)">Edit</UiButton>
            <UiButton variant="link" class="text-danger" @click="deleteTarget = vehicle">Delete</UiButton>
          </div>
        </article>
      </div>

      <UiEmptyState v-else icon="truck" title="No vehicles" description="Add the vehicles this business uses to move parcels.">
        <UiButton @click="openCreate">Add vehicle</UiButton>
      </UiEmptyState>
    </template>

    <!-- Create / edit -->
    <UiModal
      :open="formOpen"
      :title="editingId ? 'Edit vehicle' : 'Add vehicle'"
      :description="editingId ? 'The registration number cannot change.' : 'Registration numbers are unique across the business.'"
      size="md"
      :dismissible="!saving"
      @close="formOpen = false"
    >
      <div v-if="formError" class="mb-3 rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
        {{ formError }}
      </div>
      <div class="space-y-4">
        <UiSelect id="vehicle-branch-form" v-model="form.branch" label="Branch" :options="branchOptions" />
        <div class="grid gap-4 sm:grid-cols-2">
          <UiField
            id="vehicle-registration"
            v-model="form.registration"
            label="Registration number"
            :disabled="Boolean(editingId)"
            autocomplete="off"
          />
          <UiField id="vehicle-type" v-model="form.type" label="Type" hint="Truck, van, pickup…" autocomplete="off" />
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <UiField id="vehicle-model" v-model="form.makeModel" label="Make and model (optional)" autocomplete="off" />
          <UiField id="vehicle-capacity" v-model="form.capacity" label="Capacity in kg (optional)" inputmode="numeric" autocomplete="off" />
        </div>
        <UiSelect v-if="editingId" id="vehicle-status-form" v-model="form.status" label="Status" :options="statusOptions" />
      </div>
      <template #footer>
        <div class="flex flex-col gap-2 sm:flex-row-reverse">
          <UiButton block :loading="saving" @click="save">{{ editingId ? 'Save changes' : 'Add vehicle' }}</UiButton>
          <UiButton variant="tertiary" block :disabled="saving" @click="formOpen = false">Cancel</UiButton>
        </div>
      </template>
    </UiModal>

    <!-- Delete -->
    <UiModal
      :open="Boolean(deleteTarget)"
      title="Delete vehicle"
      description="This removes the vehicle from the fleet."
      :dismissible="!deleting"
      @close="deleteTarget = null"
    >
      <p class="text-label text-ink-64">
        Delete <span class="font-medium text-ink">{{ deleteTarget?.registration_number }}</span>?
      </p>
      <template #footer>
        <div class="flex flex-col gap-2 sm:flex-row-reverse">
          <UiButton variant="danger" block :loading="deleting" @click="remove">Delete vehicle</UiButton>
          <UiButton variant="tertiary" block :disabled="deleting" @click="deleteTarget = null">Keep it</UiButton>
        </div>
      </template>
    </UiModal>
  </div>
</template>
