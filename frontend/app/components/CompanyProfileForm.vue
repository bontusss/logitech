<script setup lang="ts">
import { Building2, CircleAlert, Loader2, Save } from 'lucide-vue-next'
import type { Company, OperatingHours } from '~/types/auth'
import { isValidNigerianPhone } from '~/utils/nigeria'
import {
  parseProviderBusinessTypes,
  providerBusinessTypes,
  serializeProviderBusinessTypes,
  type ProviderBusinessType,
} from '~/utils/providerBusinessTypes'

const props = defineProps<{
  company: Company
  canEdit: boolean
}>()

const emit = defineEmits<{ updated: [company: Company] }>()

const { api } = useApi()
const { notify } = useWorkspace()

const form = reactive({
  name: '',
  business_email: '',
  phone: '',
  business_types: [] as ProviderBusinessType[],
  country: 'NG',
  state: '',
  city: '',
  address: '',
  description: '',
})
const hours = ref<OperatingHours>({})
const saving = ref(false)
const error = ref('')

watch(
  () => props.company,
  (company) => {
    Object.assign(form, {
      name: company.name,
      business_email: company.business_email,
      phone: company.phone,
      business_types: parseProviderBusinessTypes(company.business_type),
      country: company.country,
      state: company.state ?? '',
      city: company.city ?? '',
      address: company.address ?? '',
      description: company.description ?? '',
    })
    hours.value = company.operating_hours ?? {}
  },
  { immediate: true, deep: true },
)

async function save() {
  error.value = ''
  if (!form.name.trim() || !form.business_email.trim() || !form.phone.trim()) {
    error.value = 'Name, business email, phone and business type are required.'
    return
  }
  if (!form.business_types.length) {
    error.value = 'Select at least one business type.'
    return
  }
  if (!isValidNigerianPhone(form.phone)) {
    error.value = 'Enter a valid Nigerian phone number.'
    return
  }
  saving.value = true
  try {
    const data = await api<{ company: Company }>('/companies/me', {
      method: 'PATCH',
      body: {
        name: form.name.trim(),
        business_email: form.business_email.trim(),
        phone: form.phone.trim(),
        business_type: serializeProviderBusinessTypes(form.business_types),
        business_types: form.business_types,
        country: form.country.trim().toUpperCase(),
        state: form.state.trim(),
        city: form.city.trim(),
        address: form.address.trim(),
        description: form.description.trim(),
        operating_hours: hours.value,
      },
    })
    notify('Company profile updated.')
    emit('updated', data.company)
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not save the company profile.')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <form class="dash-panel rounded-2xl border p-5 sm:p-6" @submit.prevent="save">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="flex items-start gap-3">
        <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-violet-100 text-violet-700 dark:bg-violet-400/15 dark:text-violet-200">
          <Building2 :size="16" />
        </span>
        <div>
          <h2 class="text-[15px] font-semibold">Company profile</h2>
          <p class="mt-1 text-[11.5px] text-ink-900/45">
            This information appears on your company page and marketplace profile.
          </p>
        </div>
      </div>
      <span
        v-if="!canEdit"
        class="rounded-md border border-ink-900/10 bg-canvas px-2 py-1 text-[10px] font-medium text-ink-900/45"
      >
        Read only
      </span>
    </div>

    <div class="mt-5 grid gap-3.5 sm:grid-cols-2">
      <label class="block sm:col-span-2">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Company name</span>
        <input
          v-model="form.name"
          type="text"
          required
          :disabled="!canEdit"
          class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500 disabled:bg-slate-50 disabled:text-ink-900/45"
        />
      </label>
      <label class="block">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Business email</span>
        <input
          v-model="form.business_email"
          type="email"
          required
          :disabled="!canEdit"
          class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500 disabled:bg-slate-50 disabled:text-ink-900/45"
        />
      </label>
      <label class="block">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Phone</span>
        <NigeriaPhoneInput v-model="form.phone" required :disabled="!canEdit" />
      </label>
      <fieldset class="block">
        <legend class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Business type</legend>
        <div class="flex flex-wrap gap-2">
          <label
            v-for="type in providerBusinessTypes"
            :key="type.value"
            class="flex cursor-pointer items-center gap-2 rounded-lg border px-3 py-2 text-[12px]"
            :class="form.business_types.includes(type.value) ? 'border-brand-500 bg-brand-50 text-brand-700' : 'border-slate-300 bg-white text-ink-900/60'"
          >
            <input v-model="form.business_types" type="checkbox" :value="type.value" :disabled="!canEdit" class="size-3.5 rounded border-slate-300" />
            {{ type.label }}
          </label>
        </div>
      </fieldset>
      <label class="block">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Country code</span>
        <input
          v-model="form.country"
          type="text"
          maxlength="2"
          :disabled="!canEdit"
          class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm uppercase focus:border-brand-500 disabled:bg-slate-50 disabled:text-ink-900/45"
        />
      </label>
      <NigeriaLocationFields
        class="sm:col-span-2"
        v-model:state="form.state"
        v-model:city="form.city"
        :disabled="!canEdit"
      />
      <label class="block sm:col-span-2">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Address</span>
        <input
          v-model="form.address"
          type="text"
          :disabled="!canEdit"
          class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500 disabled:bg-slate-50 disabled:text-ink-900/45"
        />
      </label>
      <label class="block sm:col-span-2">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Description</span>
        <textarea
          v-model="form.description"
          rows="3"
          maxlength="2000"
          :disabled="!canEdit"
          class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500 disabled:bg-slate-50 disabled:text-ink-900/45"
          placeholder="What your company moves, where it operates and what customers should know."
        />
      </label>
    </div>

    <div class="mt-5">
      <div class="flex items-center justify-between gap-2">
        <span class="text-[11px] font-medium text-ink-900/55">Operating hours</span>
        <span class="text-[10px] text-ink-900/35">Company-wide default</span>
      </div>
      <div class="mt-2">
        <OperatingHoursEditor v-model="hours" :disabled="!canEdit" />
      </div>
    </div>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />
      {{ error }}
    </p>

    <div v-if="canEdit" class="mt-5 flex justify-end">
      <button
        type="submit"
        :disabled="saving"
        class="inline-flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-brand-600 to-violet-600 px-4 py-2.5 text-[12.5px] font-semibold text-white shadow-sm shadow-brand-600/25 transition hover:-translate-y-0.5 disabled:opacity-60"
      >
        <Loader2 v-if="saving" :size="14" class="animate-spin" />
        <Save v-else :size="14" />
        Save profile
      </button>
    </div>
  </form>
</template>
