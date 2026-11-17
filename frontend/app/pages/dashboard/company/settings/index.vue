<script setup lang="ts">
import { CircleAlert, Loader2, Save } from 'lucide-vue-next'
import type { CompanySettings } from '~/types/auth'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Company settings' })

const { api } = useApi()
const { detail, loading, error, load, can, notify } = useWorkspace()

const form = reactive({
  timezone: 'Africa/Lagos',
  currency: 'NGN',
  weight_unit: 'kg' as 'kg' | 'lb',
  auto_accept_bookings: false,
  notify_email: true,
})
const saving = ref(false)
const saveError = ref('')

const timezones = [
  'Africa/Lagos',
  'Africa/Accra',
  'Africa/Nairobi',
  'Africa/Johannesburg',
  'Europe/London',
  'Europe/Paris',
  'America/New_York',
  'UTC',
]

onMounted(async () => {
  await load()
  if (detail.value) Object.assign(form, detail.value.settings)
})

watch(
  () => detail.value?.settings,
  (settings) => {
    if (settings) Object.assign(form, settings)
  },
)

async function save() {
  saveError.value = ''
  saving.value = true
  try {
    const data = await api<{ settings: CompanySettings }>('/companies/me/settings', {
      method: 'PATCH',
      body: {
        timezone: form.timezone,
        currency: form.currency.toUpperCase(),
        weight_unit: form.weight_unit,
        auto_accept_bookings: form.auto_accept_bookings,
        notify_email: form.notify_email,
      },
    })
    if (detail.value) detail.value.settings = data.settings
    notify('Company settings updated.')
  } catch (err) {
    saveError.value = apiErrorMessage(err, 'Could not save company settings.')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Company management</p>
        <h1 class="mt-2 text-[clamp(1.6rem,3vw,2.1rem)] leading-tight font-semibold tracking-[-0.035em]">Settings</h1>
        <p class="mt-2 max-w-2xl text-[13.5px] leading-relaxed text-ink-900/50">
          Workspace defaults used by rates, bookings and operations. These settings apply across all
          branches.
        </p>
      </div>
      <LookupRefreshButton :loading="loading" @refresh="load(true)" />
    </div>

    <div v-if="loading && !detail" class="mt-8 flex items-center gap-2 rounded-xl border border-ink-900/10 bg-white p-6 text-[13px] text-ink-900/50">
      <Loader2 :size="16" class="animate-spin" /> Loading settings…
    </div>

    <p v-else-if="error" role="alert" class="mt-8 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 p-4 text-[12.5px] text-red-700">
      <CircleAlert :size="16" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <form v-else-if="detail" class="mt-5 max-w-2xl rounded-xl border border-ink-900/10 bg-white p-5 sm:p-6" @submit.prevent="save">
      <h2 class="text-[15px] font-semibold">Operational defaults</h2>
      <p class="mt-1 text-[11.5px] text-ink-900/45">
        Changing the currency affects how new rates are displayed; existing records keep their values.
      </p>

      <div class="mt-5 grid gap-3.5 sm:grid-cols-2">
        <label class="block">
          <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Timezone</span>
          <select v-model="form.timezone" :disabled="!can('company.settings.manage')" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500 disabled:bg-slate-50">
            <option v-for="timezone in timezones" :key="timezone" :value="timezone">{{ timezone }}</option>
          </select>
        </label>
        <label class="block">
          <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Currency</span>
          <input v-model="form.currency" maxlength="3" :disabled="!can('company.settings.manage')" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm uppercase focus:border-brand-500 disabled:bg-slate-50" />
        </label>
        <label class="block">
          <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Weight unit</span>
          <select v-model="form.weight_unit" :disabled="!can('company.settings.manage')" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500 disabled:bg-slate-50">
            <option value="kg">Kilograms (kg)</option>
            <option value="lb">Pounds (lb)</option>
          </select>
        </label>
      </div>

      <div class="mt-5 space-y-2 border-t border-ink-900/8 pt-4">
        <label class="flex cursor-pointer items-start gap-2.5 rounded-lg px-1 py-1.5">
          <input v-model="form.auto_accept_bookings" type="checkbox" :disabled="!can('company.settings.manage')" class="mt-0.5 size-4 rounded border-slate-300" />
          <span>
            <span class="block text-[12px] font-medium">Automatically accept bookings</span>
            <span class="block text-[10.5px] text-ink-900/40">Bookings that meet published terms are accepted without manual review.</span>
          </span>
        </label>
        <label class="flex cursor-pointer items-start gap-2.5 rounded-lg px-1 py-1.5">
          <input v-model="form.notify_email" type="checkbox" :disabled="!can('company.settings.manage')" class="mt-0.5 size-4 rounded border-slate-300" />
          <span>
            <span class="block text-[12px] font-medium">Email notifications</span>
            <span class="block text-[10.5px] text-ink-900/40">Send operational and booking notifications to the company email.</span>
          </span>
        </label>
      </div>

      <p v-if="saveError" role="alert" class="mt-4 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
        <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ saveError }}
      </p>

      <div v-if="can('company.settings.manage')" class="mt-5 flex justify-end">
        <button type="submit" :disabled="saving" class="inline-flex items-center gap-1.5 rounded-lg bg-brand-500 px-4 py-2.5 text-[12.5px] font-semibold text-white transition hover:bg-brand-600 disabled:opacity-60">
          <Loader2 v-if="saving" :size="14" class="animate-spin" />
          <Save v-else :size="14" />
          Save settings
        </button>
      </div>
    </form>
  </div>
</template>
