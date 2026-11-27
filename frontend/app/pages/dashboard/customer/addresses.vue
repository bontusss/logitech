<script setup lang="ts">
import {
  CircleAlert,
  Loader2,
  MapPin,
  Pencil,
  Plus,
  Star,
  Trash2,
  X,
} from 'lucide-vue-next'
import type {
  CustomerAddressInput,
  CustomerAddressType,
  CustomerSavedAddress,
} from '~/types/customer'

useHead({ title: 'Address Book · BOL' })

const { api } = useApi()
const { notify, confirm } = useWorkspace()

const addresses = ref<CustomerSavedAddress[]>([])
const loading = ref(true)
const error = ref('')
const formOpen = ref(false)
const saving = ref(false)
const formError = ref('')
const editingId = ref<string | null>(null)
const form = reactive<CustomerAddressInput>(blankAddress())

const addressTypes: { value: CustomerAddressType; label: string }[] = [
  { value: 'residential', label: 'Home / residential' },
  { value: 'business', label: 'Business' },
  { value: 'warehouse', label: 'Warehouse' },
  { value: 'other', label: 'Other' },
]

const defaultPickup = computed(() => addresses.value.find((address) => address.default_pickup))
const defaultDelivery = computed(() => addresses.value.find((address) => address.default_delivery))

function blankAddress(): CustomerAddressInput {
  return {
    label: '',
    address_type: 'residential',
    contact_name: '',
    phone: '',
    state: '',
    city: '',
    area: '',
    street_address: '',
    directions: '',
    latitude: null,
    longitude: null,
    default_pickup: false,
    default_delivery: false,
  }
}

function typeLabel(type: CustomerAddressType) {
  return addressTypes.find((option) => option.value === type)?.label || 'Address'
}

function addressLine(address: CustomerSavedAddress) {
  return [address.street_address, address.area, address.city, address.state]
    .filter(Boolean)
    .join(', ')
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const result = await api<{ addresses: CustomerSavedAddress[] }>('/customer/addresses')
    addresses.value = result.addresses
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not load your saved addresses.')
  } finally {
    loading.value = false
  }
}

function openAdd() {
  editingId.value = null
  Object.assign(form, blankAddress())
  formError.value = ''
  formOpen.value = true
}

function openEdit(address: CustomerSavedAddress) {
  editingId.value = address.id
  Object.assign(form, {
    label: address.label,
    address_type: address.address_type,
    contact_name: address.contact_name,
    phone: address.phone,
    state: address.state,
    city: address.city,
    area: address.area,
    street_address: address.street_address,
    directions: address.directions,
    latitude: address.latitude,
    longitude: address.longitude,
    default_pickup: address.default_pickup,
    default_delivery: address.default_delivery,
  })
  formError.value = ''
  formOpen.value = true
}

function closeForm() {
  formOpen.value = false
  formError.value = ''
}

async function save() {
  if (!form.label.trim() || !form.street_address.trim() || !form.state || !form.city) {
    formError.value = 'Add a label, state, city and street address.'
    return
  }
  saving.value = true
  formError.value = ''
  try {
    const body = {
      ...form,
      label: form.label.trim(),
      street_address: form.street_address.trim(),
      contact_name: form.contact_name.trim(),
      phone: form.phone.trim(),
      area: form.area.trim(),
      directions: form.directions.trim(),
    }
    const result = editingId.value
      ? await api<{ address: CustomerSavedAddress }>(`/customer/addresses/${editingId.value}`, {
          method: 'PATCH',
          body,
        })
      : await api<{ address: CustomerSavedAddress }>('/customer/addresses', {
          method: 'POST',
          body,
        })
    await load()
    notify(editingId.value ? 'Address updated.' : 'Address saved.')
    editingId.value = null
    formOpen.value = false
    return result
  } catch (cause) {
    formError.value = apiErrorMessage(cause, 'Could not save this address.')
  } finally {
    saving.value = false
  }
}

async function remove(address: CustomerSavedAddress) {
  const confirmed = await confirm({
    title: `Delete “${address.label}”?`,
    message: 'This address will be removed from your address book.',
    confirmLabel: 'Delete address',
    cancelLabel: 'Keep it',
    danger: true,
  })
  if (!confirmed) return
  try {
    await api(`/customer/addresses/${address.id}`, { method: 'DELETE' })
    addresses.value = addresses.value.filter((item) => item.id !== address.id)
    notify('Address deleted.')
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not delete this address.')
  }
}

onMounted(load)
</script>

<template>
  <div class="dash-page">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Personal</p>
        <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Address book</h1>
        <p class="mt-1.5 max-w-2xl text-[12.5px] leading-relaxed text-ink-900/50">
          Save the pickup and delivery places you use most so booking takes less typing next time.
        </p>
      </div>
      <button
        type="button"
        class="inline-flex items-center gap-1.5 rounded-lg bg-accent-500 px-3.5 py-2 text-[11.5px] font-semibold text-white transition hover:bg-accent-400"
        @click="openAdd"
      >
        <Plus :size="14" /> Add address
      </button>
    </div>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ error }}
    </p>

    <div class="mt-4 grid gap-2.5 sm:grid-cols-2">
      <div class="flex items-start gap-2.5 rounded-xl border border-ink-900/8 bg-surface px-3.5 py-3">
        <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-emerald-50 text-emerald-700 dark:bg-emerald-400/10 dark:text-emerald-200">
          <Star :size="14" />
        </span>
        <div class="min-w-0">
          <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">Default pickup</p>
          <p class="mt-0.5 truncate text-[11.5px] font-medium">{{ defaultPickup?.label || 'Not set yet' }}</p>
        </div>
      </div>
      <div class="flex items-start gap-2.5 rounded-xl border border-ink-900/8 bg-surface px-3.5 py-3">
        <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
          <MapPin :size="14" />
        </span>
        <div class="min-w-0">
          <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">Default delivery</p>
          <p class="mt-0.5 truncate text-[11.5px] font-medium">{{ defaultDelivery?.label || 'Not set yet' }}</p>
        </div>
      </div>
    </div>

    <div v-if="loading" class="mt-4 flex items-center justify-center gap-2 rounded-xl border border-ink-900/10 bg-surface p-12 text-[12px] text-ink-900/45">
      <Loader2 :size="15" class="animate-spin" /> Loading your addresses…
    </div>

    <div v-else-if="!addresses.length" class="mt-4 rounded-xl border border-ink-900/10 bg-surface p-10 text-center">
      <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
        <MapPin :size="20" />
      </span>
      <h2 class="mt-3 text-[13.5px] font-semibold">No saved addresses yet</h2>
      <p class="mx-auto mt-1 max-w-md text-[11.5px] leading-relaxed text-ink-900/45">
        Add your home, office or a warehouse and set one as your default pickup or delivery.
      </p>
      <button
        type="button"
        class="mt-4 inline-flex items-center gap-1.5 rounded-lg bg-accent-500 px-3.5 py-2 text-[11.5px] font-semibold text-white transition hover:bg-accent-400"
        @click="openAdd"
      >
        <Plus :size="13" /> Add your first address
      </button>
    </div>

    <ul v-else class="mt-4 grid gap-2.5 sm:grid-cols-2">
      <li v-for="address in addresses" :key="address.id" class="rounded-xl border border-ink-900/10 bg-surface p-4">
        <div class="flex items-start justify-between gap-2">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-1.5">
              <p class="text-[12.5px] font-semibold">{{ address.label }}</p>
              <span class="rounded-full border border-ink-900/10 bg-canvas px-2 py-0.5 text-[9px] font-semibold text-ink-900/50">
                {{ typeLabel(address.address_type) }}
              </span>
            </div>
            <p class="mt-1 text-[11px] leading-relaxed text-ink-900/60">{{ addressLine(address) }}</p>
          </div>
          <div class="flex shrink-0 items-center gap-1">
            <button
              type="button"
              class="rounded-lg p-1.5 text-ink-900/40 transition hover:bg-canvas hover:text-ink-900"
              :aria-label="`Edit ${address.label}`"
              @click="openEdit(address)"
            >
              <Pencil :size="13" />
            </button>
            <button
              type="button"
              class="rounded-lg p-1.5 text-ink-900/40 transition hover:bg-rose-50 hover:text-rose-700"
              :aria-label="`Delete ${address.label}`"
              @click="remove(address)"
            >
              <Trash2 :size="13" />
            </button>
          </div>
        </div>

        <div v-if="address.contact_name || address.phone" class="mt-2 text-[10.5px] text-ink-900/45">
          {{ address.contact_name }}<span v-if="address.contact_name && address.phone"> · </span>{{ address.phone }}
        </div>
        <p v-if="address.directions" class="mt-1.5 rounded-lg bg-canvas/70 px-2.5 py-2 text-[10.5px] leading-relaxed text-ink-900/50">
          {{ address.directions }}
        </p>
        <div class="mt-2 flex flex-wrap gap-1.5">
          <span v-if="address.default_pickup" class="rounded-full bg-emerald-50 px-2 py-0.5 text-[9px] font-semibold text-emerald-700 dark:bg-emerald-400/10 dark:text-emerald-200">
            Default pickup
          </span>
          <span v-if="address.default_delivery" class="rounded-full bg-brand-50 px-2 py-0.5 text-[9px] font-semibold text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
            Default delivery
          </span>
        </div>
      </li>
    </ul>

    <Teleport to="body">
      <div
        v-if="formOpen"
        class="fixed inset-0 z-[70] flex items-start justify-center overflow-y-auto bg-ink-950/45 px-3 py-6 backdrop-blur-sm sm:px-5"
        role="dialog"
        aria-modal="true"
        :aria-label="editingId ? 'Edit address' : 'Add address'"
        @click.self="closeForm"
      >
        <form class="my-auto w-full max-w-lg overflow-hidden rounded-2xl border border-ink-900/10 bg-surface shadow-2xl" @submit.prevent="save">
          <div class="flex items-start justify-between gap-3 border-b border-ink-900/8 px-5 py-4">
            <div>
              <p class="text-[14px] font-semibold">{{ editingId ? 'Edit address' : 'Add address' }}</p>
              <p class="mt-0.5 text-[10.5px] text-ink-900/40">Saved addresses are private to your account.</p>
            </div>
            <button type="button" class="text-ink-900/35 hover:text-ink-900" aria-label="Close form" @click="closeForm">
              <X :size="18" />
            </button>
          </div>

          <div class="max-h-[calc(100dvh-11rem)] space-y-3 overflow-y-auto p-5">
            <p v-if="formError" role="alert" class="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[11px] text-red-700">
              <CircleAlert :size="13" class="mt-0.5 shrink-0" />{{ formError }}
            </p>

            <div class="grid gap-3 sm:grid-cols-2">
              <label class="block">
                <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Label</span>
                <input
                  v-model="form.label"
                  type="text"
                  required
                  maxlength="80"
                  placeholder="Home, Office, Warehouse…"
                  class="w-full rounded-xl border border-ink-900/15 bg-canvas px-3.5 py-2.5 text-sm outline-none focus:border-accent-500"
                />
              </label>
              <label class="block">
                <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Address type</span>
                <select v-model="form.address_type" class="w-full rounded-xl border border-ink-900/15 bg-canvas px-3.5 py-2.5 text-sm outline-none focus:border-accent-500">
                  <option v-for="option in addressTypes" :key="option.value" :value="option.value">{{ option.label }}</option>
                </select>
              </label>
            </div>

            <div class="grid gap-3 sm:grid-cols-2">
              <label class="block">
                <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Contact name</span>
                <input
                  v-model="form.contact_name"
                  type="text"
                  maxlength="120"
                  class="w-full rounded-xl border border-ink-900/15 bg-canvas px-3.5 py-2.5 text-sm outline-none focus:border-accent-500"
                />
              </label>
              <label class="block">
                <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Phone number</span>
                <input
                  v-model="form.phone"
                  type="tel"
                  maxlength="30"
                  class="w-full rounded-xl border border-ink-900/15 bg-canvas px-3.5 py-2.5 text-sm outline-none focus:border-accent-500"
                />
              </label>
            </div>

            <NigeriaLocationFields v-model:state="form.state" v-model:city="form.city" required />

            <label class="block">
              <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Street address</span>
              <input
                v-model="form.street_address"
                type="text"
                required
                maxlength="240"
                placeholder="Building, street and number"
                class="w-full rounded-xl border border-ink-900/15 bg-canvas px-3.5 py-2.5 text-sm outline-none focus:border-accent-500"
              />
            </label>

            <label class="block">
              <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Area / landmark</span>
              <input
                v-model="form.area"
                type="text"
                maxlength="120"
                class="w-full rounded-xl border border-ink-900/15 bg-canvas px-3.5 py-2.5 text-sm outline-none focus:border-accent-500"
              />
            </label>

            <label class="block">
              <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">Directions for the rider</span>
              <textarea
                v-model="form.directions"
                rows="2"
                maxlength="500"
                placeholder="Landmarks, gate colour, best time to arrive…"
                class="w-full rounded-xl border border-ink-900/15 bg-canvas px-3.5 py-2.5 text-sm outline-none focus:border-accent-500"
              />
            </label>

            <div class="grid gap-2 sm:grid-cols-2">
              <label class="flex items-center gap-2 rounded-xl border border-ink-900/10 bg-canvas/60 px-3 py-2.5 text-[11.5px] font-medium">
                <input v-model="form.default_pickup" type="checkbox" class="size-3.5 accent-emerald-600" />
                Use as default pickup
              </label>
              <label class="flex items-center gap-2 rounded-xl border border-ink-900/10 bg-canvas/60 px-3 py-2.5 text-[11.5px] font-medium">
                <input v-model="form.default_delivery" type="checkbox" class="size-3.5 accent-brand-600" />
                Use as default delivery
              </label>
            </div>
          </div>

          <div class="flex flex-wrap justify-end gap-2 border-t border-ink-900/8 px-5 py-4">
            <button
              type="button"
              class="rounded-lg border border-ink-900/15 bg-surface px-3 py-2 text-[11.5px] font-semibold transition hover:bg-canvas"
              @click="closeForm"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="saving"
              class="inline-flex items-center gap-1.5 rounded-lg bg-accent-500 px-3.5 py-2 text-[11.5px] font-semibold text-white transition hover:bg-accent-400 disabled:opacity-50"
            >
              <Loader2 v-if="saving" :size="13" class="animate-spin" />
              {{ editingId ? 'Save changes' : 'Save address' }}
            </button>
          </div>
        </form>
      </div>
    </Teleport>
  </div>
</template>
