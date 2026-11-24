<script setup lang="ts">
import { CircleAlert, ImagePlus, Loader2, Palette, Save } from 'lucide-vue-next'
import type { CompanyBranding } from '~/types/auth'

const props = withDefaults(
  defineProps<{
    branding: CompanyBranding
    subdomain?: string
    canManage?: boolean
  }>(),
  { subdomain: '', canManage: false },
)

const emit = defineEmits<{
  updated: [branding: CompanyBranding]
  'subdomain-updated': [subdomain: string]
}>()

const { api } = useApi()
const { notify } = useWorkspace()

const form = reactive({
  public_name: '',
  description: '',
  pwa_short_name: '',
  logo_url: '',
  favicon_url: '',
  cover_image_url: '',
  primary_color: '#2b4bff',
  secondary_color: '#0a1220',
})
const saving = ref(false)
const uploading = ref('')
const error = ref('')

watch(
  () => props.branding,
  (branding) => {
    Object.assign(form, {
      public_name: branding.public_name ?? '',
      description: branding.description ?? '',
      pwa_short_name: branding.pwa_short_name ?? '',
      logo_url: branding.logo_url ?? '',
      favicon_url: branding.favicon_url ?? '',
      cover_image_url: branding.cover_image_url ?? '',
      primary_color: branding.primary_color,
      secondary_color: branding.secondary_color,
    })
  },
  { immediate: true, deep: true },
)

async function save() {
  error.value = ''
  if (!/^#[0-9a-fA-F]{6}$/.test(form.primary_color) || !/^#[0-9a-fA-F]{6}$/.test(form.secondary_color)) {
    error.value = 'Colours must be six-digit hex values such as #2b4bff.'
    return
  }
  saving.value = true
  try {
    const data = await api<{ branding: CompanyBranding }>('/companies/me/branding', {
      method: 'PATCH',
      body: {
        public_name: form.public_name.trim(),
        description: form.description.trim(),
        pwa_short_name: form.pwa_short_name.trim(),
        logo_url: form.logo_url.trim(),
        favicon_url: form.favicon_url.trim(),
        cover_image_url: form.cover_image_url.trim(),
        primary_color: form.primary_color,
        secondary_color: form.secondary_color,
      },
    })
    notify('Branding updated.')
    emit('updated', data.branding)
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not save the branding.')
  } finally {
    saving.value = false
  }
}

async function upload(kind: 'logo' | 'cover' | 'favicon', event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  error.value = ''
  uploading.value = kind
  try {
    const body = new FormData()
    body.append('kind', kind)
    body.append('file', file)
    const data = await api<{ branding: CompanyBranding }>('/companies/me/branding/upload', {
      method: 'POST',
      body,
    })
    notify(`${kind === 'cover' ? 'Cover image' : kind === 'favicon' ? 'Favicon' : 'Logo'} uploaded.`)
    emit('updated', data.branding)
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not upload that image.')
  } finally {
    uploading.value = ''
    input.value = ''
  }
}
</script>

<template>
  <div class="space-y-5">
    <form class="dash-panel rounded-2xl border p-5 sm:p-6" @submit.prevent="save">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div class="flex items-start gap-3">
          <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-red-100 text-red-700 dark:bg-red-400/15 dark:text-red-200">
            <Palette :size="16" />
          </span>
          <div>
            <h2 class="text-[15px] font-semibold">Brand identity</h2>
            <p class="mt-1 text-[11.5px] text-ink-900/45">
              Used on your public company page, workspace accents and installable app.
            </p>
          </div>
        </div>
        <span v-if="!canManage" class="rounded-md border border-ink-900/10 bg-canvas px-2 py-1 text-[10px] text-ink-900/45">Read only</span>
      </div>

      <div class="mt-5 grid gap-5 lg:grid-cols-[minmax(0,1fr)_260px]">
        <div class="space-y-3.5">
          <label class="block">
            <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Public display name</span>
            <input v-model="form.public_name" :disabled="!canManage" maxlength="160" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500 disabled:bg-slate-50" placeholder="Aba Freight Express" />
          </label>
          <label class="block">
            <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Public description</span>
            <textarea v-model="form.description" :disabled="!canManage" rows="3" maxlength="2000" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500 disabled:bg-slate-50" />
          </label>
          <label class="block">
            <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">PWA short name</span>
            <input v-model="form.pwa_short_name" :disabled="!canManage" maxlength="30" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500 disabled:bg-slate-50" placeholder="AbaFreight" />
            <span class="mt-1 block text-[10px] text-ink-900/35">Up to 30 characters, shown under the installed app icon.</span>
          </label>

          <div class="grid gap-3 sm:grid-cols-2">
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Primary colour</span>
              <div class="flex items-center gap-2">
                <input v-model="form.primary_color" type="color" :disabled="!canManage" class="size-9 cursor-pointer rounded border border-slate-300 bg-white p-1 disabled:opacity-50" />
                <input v-model="form.primary_color" :disabled="!canManage" class="w-full rounded-lg border border-slate-300 px-2.5 py-2 font-mono text-[11.5px] uppercase focus:border-brand-500 disabled:bg-slate-50" />
              </div>
            </label>
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Secondary colour</span>
              <div class="flex items-center gap-2">
                <input v-model="form.secondary_color" type="color" :disabled="!canManage" class="size-9 cursor-pointer rounded border border-slate-300 bg-white p-1 disabled:opacity-50" />
                <input v-model="form.secondary_color" :disabled="!canManage" class="w-full rounded-lg border border-slate-300 px-2.5 py-2 font-mono text-[11.5px] uppercase focus:border-brand-500 disabled:bg-slate-50" />
              </div>
            </label>
          </div>
        </div>

        <div class="space-y-3">
          <div class="overflow-hidden rounded-xl border border-ink-900/10">
            <div class="relative h-24" :style="{ backgroundColor: form.secondary_color }">
              <img v-if="form.cover_image_url" :src="form.cover_image_url" alt="" class="h-full w-full object-cover opacity-80" />
              <div class="absolute inset-0 flex items-end p-3">
                <div class="flex items-center gap-2">
                  <span class="flex size-9 items-center justify-center overflow-hidden rounded-lg text-[12px] font-semibold text-white" :style="{ backgroundColor: form.primary_color }">
                    <img v-if="form.logo_url" :src="form.logo_url" alt="" class="h-full w-full object-cover" />
                    <template v-else>{{ form.public_name.slice(0, 1).toUpperCase() || 'B' }}</template>
                  </span>
                  <div>
                    <p class="text-[12px] font-semibold text-white">{{ form.public_name || 'Your company' }}</p>
                    <p class="text-[9px] text-white/60">Company preview</p>
                  </div>
                </div>
              </div>
            </div>
            <div class="bg-white p-3">
              <div class="h-1.5 w-16 rounded-full" :style="{ backgroundColor: form.primary_color }" />
              <p class="mt-2 line-clamp-2 text-[10px] leading-relaxed text-ink-900/45">{{ form.description || 'Public description appears here.' }}</p>
            </div>
          </div>

          <label v-if="canManage" class="flex cursor-pointer items-center justify-center gap-1.5 rounded-lg border border-dashed border-ink-900/20 px-3 py-2 text-[11px] font-medium text-ink-900/55 transition hover:bg-canvas">
            <ImagePlus :size="13" />
            {{ uploading === 'logo' ? 'Uploading…' : 'Upload logo' }}
            <input type="file" accept="image/png,image/jpeg,image/webp" class="hidden" :disabled="Boolean(uploading)" @change="upload('logo', $event)" />
          </label>
          <label v-if="canManage" class="flex cursor-pointer items-center justify-center gap-1.5 rounded-lg border border-dashed border-ink-900/20 px-3 py-2 text-[11px] font-medium text-ink-900/55 transition hover:bg-canvas">
            <ImagePlus :size="13" />
            {{ uploading === 'cover' ? 'Uploading…' : 'Upload cover image' }}
            <input type="file" accept="image/png,image/jpeg,image/webp" class="hidden" :disabled="Boolean(uploading)" @change="upload('cover', $event)" />
          </label>
          <label v-if="canManage" class="flex cursor-pointer items-center justify-center gap-1.5 rounded-lg border border-dashed border-ink-900/20 px-3 py-2 text-[11px] font-medium text-ink-900/55 transition hover:bg-canvas">
            <ImagePlus :size="13" />
            {{ uploading === 'favicon' ? 'Uploading…' : 'Upload favicon' }}
            <input type="file" accept="image/png,image/jpeg,image/webp" class="hidden" :disabled="Boolean(uploading)" @change="upload('favicon', $event)" />
          </label>
        </div>
      </div>

      <div class="mt-5 grid gap-3 border-t border-ink-900/8 pt-4 sm:grid-cols-3">
        <label class="block">
          <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Logo URL</span>
          <input v-model="form.logo_url" :disabled="!canManage" type="url" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-[11.5px] focus:border-brand-500 disabled:bg-slate-50" />
        </label>
        <label class="block">
          <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Cover image URL</span>
          <input v-model="form.cover_image_url" :disabled="!canManage" type="url" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-[11.5px] focus:border-brand-500 disabled:bg-slate-50" />
        </label>
        <label class="block">
          <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Favicon URL</span>
          <input v-model="form.favicon_url" :disabled="!canManage" type="url" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-[11.5px] focus:border-brand-500 disabled:bg-slate-50" />
        </label>
      </div>

      <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
        <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ error }}
      </p>

      <div v-if="canManage" class="mt-5 flex justify-end">
        <button type="submit" :disabled="saving" class="inline-flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-brand-600 to-brand-600 px-4 py-2.5 text-[12.5px] font-semibold text-white shadow-sm shadow-brand-600/25 transition hover:-translate-y-0.5 disabled:opacity-60">
          <Loader2 v-if="saving" :size="14" class="animate-spin" />
          <Save v-else :size="14" />
          Save branding
        </button>
      </div>
    </form>

    <SubdomainSettings
      :subdomain="subdomain"
      :can-manage="canManage"
      @updated="emit('subdomain-updated', $event)"
    />
  </div>
</template>
