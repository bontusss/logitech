<script setup lang="ts">
import { CircleAlert, ExternalLink, Loader2 } from 'lucide-vue-next'
import type { CompanyBranding } from '~/types/auth'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Branding' })

const { detail, loading, error, load, can, notify } = useWorkspace()
const config = useRuntimeConfig()

onMounted(() => load())

const domain = computed(() => (config.public.appDomain as string) || 'localhost')
const publicURL = computed(() =>
  detail.value?.subdomain ? `http://${detail.value.subdomain}.${domain.value}` : '',
)

function updateBranding(branding: CompanyBranding) {
  if (detail.value) detail.value.branding = branding
}

function updateSubdomain(subdomain: string) {
  if (detail.value) detail.value.subdomain = subdomain
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Company management</p>
        <h1 class="mt-2 text-[clamp(1.6rem,3vw,2.1rem)] leading-tight font-semibold tracking-[-0.035em]">Branding</h1>
        <p class="mt-2 max-w-2xl text-[13.5px] leading-relaxed text-ink-900/50">
          Your logo, colours and public identity carry through to the company page, workspace and the
          installable PWA served from your subdomain.
        </p>
      </div>
      <LookupRefreshButton :loading="loading" @refresh="load(true)" />
    </div>

    <div v-if="loading && !detail" class="mt-8 flex items-center gap-2 rounded-xl border border-ink-900/10 bg-white p-6 text-[13px] text-ink-900/50">
      <Loader2 :size="16" class="animate-spin" /> Loading branding…
    </div>

    <p v-else-if="error" role="alert" class="mt-8 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 p-4 text-[12.5px] text-red-700">
      <CircleAlert :size="16" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <template v-else-if="detail">
      <div v-if="publicURL" class="mt-6 flex flex-wrap items-center gap-2 rounded-xl border border-ink-900/10 bg-white px-4 py-3">
        <p class="text-[11.5px] text-ink-900/45">Public company address</p>
        <span class="font-mono text-[11px] font-medium text-ink-900/70">{{ publicURL }}</span>
        <a :href="publicURL" target="_blank" rel="noopener" class="ml-auto inline-flex items-center gap-1 text-[11px] font-semibold text-brand-600 hover:text-brand-700">
          Open <ExternalLink :size="12" />
        </a>
      </div>

      <BrandingSettings
        class="mt-5"
        :branding="detail.branding"
        :subdomain="detail.subdomain"
        :can-manage="can('branding.manage')"
        @updated="updateBranding"
        @subdomain-updated="updateSubdomain"
      />

      <div class="mt-5 rounded-xl border border-ink-900/10 bg-canvas p-4 text-[11px] leading-relaxed text-ink-900/45">
        <strong class="font-semibold text-ink-900/60">Installable app:</strong>
        when staff visit the company subdomain, the app name, short name, icon and theme colour are
        generated from this branding. One frontend deployment serves every company; no separate build
        is required.
      </div>
    </template>
  </div>
</template>
