<script setup lang="ts">
import { Building2, CircleAlert, Loader2, Save, X } from 'lucide-vue-next'
import type { Branch, BranchInput, OperatingHours, StaffMember } from '~/types/auth'
import { isValidNigerianPhone } from '~/utils/nigeria'

const props = withDefaults(
  defineProps<{
    branch?: Branch | null
    members?: StaffMember[]
    saving?: boolean
  }>(),
  { branch: null, members: () => [], saving: false },
)

const emit = defineEmits<{
  submit: [payload: BranchInput]
  cancel: []
}>()

const form = reactive({
  name: '',
  code: '',
  address: '',
  city: '',
  state: '',
  country: 'NG',
  phone: '',
  email: '',
  manager_member_id: '',
})
const hours = ref<OperatingHours>({})
const error = ref('')

watch(
  () => props.branch,
  (branch) => {
    Object.assign(form, {
      name: branch?.name ?? '',
      code: branch?.code ?? '',
      address: branch?.address ?? '',
      city: branch?.city ?? '',
      state: branch?.state ?? '',
      country: branch?.country ?? 'NG',
      phone: branch?.phone ?? '',
      email: branch?.email ?? '',
      manager_member_id: branch?.manager_member_id ?? '',
    })
    hours.value = branch?.operating_hours ?? {}
    error.value = ''
  },
  { immediate: true, deep: true },
)

function submit() {
  error.value = ''
  if (!form.name.trim() || !form.code.trim() || !form.address.trim() || !form.city.trim() || !form.state.trim()) {
    error.value = 'Name, code, address, city and state are required.'
    return
  }
  if (!/^[a-zA-Z0-9-]{2,24}$/.test(form.code.trim())) {
    error.value = 'Branch code must be 2–24 letters, numbers or hyphens.'
    return
  }
  if (form.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
    error.value = 'Enter a valid branch email address.'
    return
  }
  if (form.phone && !isValidNigerianPhone(form.phone)) {
    error.value = 'Enter a valid Nigerian phone number.'
    return
  }
  emit('submit', {
    name: form.name.trim(),
    code: form.code.trim().toUpperCase(),
    address: form.address.trim(),
    city: form.city.trim(),
    state: form.state.trim(),
    country: form.country.trim().toUpperCase() || 'NG',
    phone: form.phone.trim(),
    email: form.email.trim(),
    manager_member_id: form.manager_member_id || null,
    operating_hours: hours.value,
  })
}
</script>

<template>
  <form class="dash-panel rounded-2xl border p-5" @submit.prevent="submit">
    <div class="flex items-start justify-between gap-3">
      <div class="flex items-start gap-3">
        <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-brand-100 text-brand-700 dark:bg-brand-400/15 dark:text-brand-200">
          <Building2 :size="16" />
        </span>
        <div>
          <h2 class="text-[15px] font-semibold">{{ branch ? 'Edit branch' : 'New branch' }}</h2>
          <p class="mt-1 text-[11.5px] text-ink-900/45">
            Codes are unique within your company and used on manifests.
          </p>
        </div>
      </div>
      <button type="button" class="rounded-lg p-1 text-ink-900/35 hover:bg-ink-900/5 hover:text-ink-900" aria-label="Close branch form" @click="$emit('cancel')">
        <X :size="16" />
      </button>
    </div>

    <div class="mt-5 grid gap-3.5 sm:grid-cols-2">
      <label class="block">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Branch name</span>
        <input v-model="form.name" required class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" placeholder="Aba headquarters" />
      </label>
      <label class="block">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Branch code</span>
        <input v-model="form.code" required maxlength="24" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-mono text-sm uppercase focus:border-brand-500" placeholder="ABA" />
      </label>
      <label class="block sm:col-span-2">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Address</span>
        <input v-model="form.address" required class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" />
      </label>
      <NigeriaLocationFields
        class="sm:col-span-2"
        v-model:state="form.state"
        v-model:city="form.city"
        required
      />
      <label class="block">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Country code</span>
        <input v-model="form.country" maxlength="2" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm uppercase focus:border-brand-500" />
      </label>
      <label class="block">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Phone (optional)</span>
        <NigeriaPhoneInput v-model="form.phone" />
      </label>
      <label class="block">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Branch email</span>
        <input v-model="form.email" type="email" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" />
      </label>
      <label class="block">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Branch manager</span>
        <select v-model="form.manager_member_id" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500">
          <option value="">No manager</option>
          <option v-for="member in members.filter((item) => item.status === 'active')" :key="member.id" :value="member.id">
            {{ member.full_name }} · {{ member.email }}
          </option>
        </select>
      </label>
    </div>

    <div class="mt-5">
      <span class="text-[11px] font-medium text-ink-900/55">Branch operating hours</span>
      <div class="mt-2"><OperatingHoursEditor v-model="hours" /></div>
    </div>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <div class="mt-5 flex justify-end gap-2">
      <button type="button" class="rounded-lg border border-slate-300 px-3.5 py-2 text-[12.5px] font-medium transition hover:bg-slate-50" @click="$emit('cancel')">
        Cancel
      </button>
      <button type="submit" :disabled="saving" class="inline-flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-brand-600 to-brand-600 px-4 py-2.5 text-[12.5px] font-semibold text-white shadow-sm shadow-brand-600/25 transition hover:-translate-y-0.5 disabled:opacity-60">
        <Loader2 v-if="saving" :size="14" class="animate-spin" />
        <Save v-else :size="14" />
        {{ branch ? 'Save branch' : 'Create branch' }}
      </button>
    </div>
  </form>
</template>
