<script setup lang="ts">
import { CheckCircle2, CircleAlert, Globe2, Loader2, Save } from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    subdomain?: string
    canManage?: boolean
  }>(),
  { subdomain: '', canManage: false },
)

const emit = defineEmits<{ updated: [subdomain: string] }>()

const { api } = useApi()
const { notify } = useWorkspace()
const config = useRuntimeConfig()

const domain = computed(() => (config.public.appDomain as string) || 'localhost')
const value = ref(props.subdomain)
const checking = ref(false)
const saving = ref(false)
const availability = ref<{ available: boolean | null; reason: string }>({ available: null, reason: '' })

watch(
  () => props.subdomain,
  (subdomain) => {
    value.value = subdomain
    availability.value = { available: null, reason: '' }
  },
)

const previewURL = computed(() => {
  const slug = value.value.trim().toLowerCase()
  return slug ? `${slug}.${domain.value}` : `your-company.${domain.value}`
})

function validateLocally(slug: string) {
  if (!slug) return 'Enter a subdomain.'
  if (!/^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$/.test(slug)) {
    return 'Use 3–30 lowercase letters, numbers or hyphens, without a leading or trailing hyphen.'
  }
  if (slug.includes('--')) return 'Consecutive hyphens are not allowed.'
  return ''
}

async function check() {
  const slug = value.value.trim().toLowerCase()
  value.value = slug
  const localReason = validateLocally(slug)
  if (localReason) {
    availability.value = { available: false, reason: localReason }
    return false
  }
  checking.value = true
  try {
    const data = await api<{ available: boolean; reason?: string }>('/companies/me/subdomain/availability', {
      query: { subdomain: slug },
    })
    availability.value = { available: data.available, reason: data.reason ?? '' }
    return data.available
  } catch (err) {
    availability.value = { available: false, reason: apiErrorMessage(err, 'Could not check that subdomain.') }
    return false
  } finally {
    checking.value = false
  }
}

async function save() {
  if (!(await check())) return
  saving.value = true
  try {
    const data = await api<{ subdomain: string }>('/companies/me/subdomain', {
      method: 'PUT',
      body: { subdomain: value.value.trim().toLowerCase() },
    })
    notify('Company subdomain updated.')
    emit('updated', data.subdomain)
    availability.value = { available: true, reason: 'This subdomain is active for your company.' }
  } catch (err) {
    availability.value = { available: false, reason: apiErrorMessage(err, 'Could not save the subdomain.') }
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="dash-panel rounded-2xl border p-5">
    <div class="flex items-start gap-2.5">
      <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-brand-100 text-brand-700 dark:bg-brand-400/15 dark:text-brand-200">
        <Globe2 :size="16" />
      </span>
      <div>
        <h2 class="text-[14px] font-semibold">Company subdomain</h2>
        <p class="mt-1 text-[11.5px] leading-relaxed text-ink-900/45">
          Your public company page and installable app are served from this address. Subdomains are
          unique across the platform and reserved words are blocked.
        </p>
      </div>
    </div>

    <div class="mt-4 flex flex-wrap items-end gap-2">
      <label class="block min-w-[220px] flex-1">
        <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Subdomain</span>
        <div class="flex rounded-lg border border-slate-300 focus-within:border-brand-500">
          <input
            v-model="value"
            type="text"
            maxlength="30"
            :disabled="!canManage || saving"
            class="min-w-0 flex-1 rounded-l-lg bg-transparent px-3 py-2.5 text-sm lowercase outline-none disabled:bg-slate-50"
            placeholder="aba-freight"
            @blur="check"
          />
          <span class="flex items-center border-l border-slate-200 bg-canvas px-2.5 text-[11.5px] text-ink-900/40">
            .{{ domain }}
          </span>
        </div>
      </label>
      <button
        v-if="canManage"
        type="button"
        :disabled="checking || saving"
        class="inline-flex items-center gap-1.5 rounded-lg bg-accent-500 px-3.5 py-2.5 text-[12px] font-semibold text-ink-950 shadow-sm shadow-accent-500/20 transition hover:bg-accent-400 disabled:opacity-60"
        @click="save"
      >
        <Loader2 v-if="saving || checking" :size="14" class="animate-spin" />
        <Save v-else :size="14" />
        Save subdomain
      </button>
    </div>

    <p class="mt-2 text-[10.5px] text-ink-900/40">
      Public address: <span class="font-medium text-ink-900/60">{{ previewURL }}</span>
    </p>

    <p
      v-if="availability.available === true"
      class="mt-2 flex items-start gap-1.5 text-[11px] text-success-700"
    >
      <CheckCircle2 :size="13" class="mt-0.5 shrink-0" />
      {{ availability.reason || 'This subdomain is available.' }}
    </p>
    <p
      v-else-if="availability.available === false"
      class="mt-2 flex items-start gap-1.5 text-[11px] text-red-700"
    >
      <CircleAlert :size="13" class="mt-0.5 shrink-0" />
      {{ availability.reason }}
    </p>

    <div class="mt-3 rounded-lg bg-canvas px-3 py-2 text-[10.5px] leading-relaxed text-ink-900/40">
      DNS is configured once by the platform. Company subdomains do not require separate wildcard
      certificates from your team.
    </div>
  </div>
</template>
