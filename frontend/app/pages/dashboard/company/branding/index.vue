<script setup lang="ts">
import { CircleAlert, ExternalLink, Loader2, Palette, Sparkles } from 'lucide-vue-next'
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
  <div class="dash-page">
    <header class="dash-hero relative overflow-hidden rounded-2xl border p-5 sm:p-6">
      <div class="pointer-events-none absolute -top-16 right-10 size-44 rounded-full bg-accent-400/15 blur-3xl" />
      <div class="pointer-events-none absolute -right-8 -bottom-24 size-52 rounded-full bg-brand-400/15 blur-3xl" />
      <div class="relative flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="dash-eyebrow inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-semibold tracking-wide">
            <Sparkles :size="12" /> LOOK &amp; FEEL
          </p>
          <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Branding</h1>
          <p class="mt-1.5 max-w-xl text-[12px] leading-relaxed text-ink-900/65">
            Your logo, colours and public identity carry through to the company page, workspace and
            the installable PWA served from your subdomain.
          </p>
        </div>
        <LookupRefreshButton :loading="loading" @refresh="load(true)" />
      </div>
    </header>

    <div v-if="loading && !detail" class="dash-panel mt-4 flex items-center gap-2 rounded-2xl border p-6 text-[13px] text-ink-900/50">
      <Loader2 :size="16" class="animate-spin" /> Loading branding…
    </div>

    <p v-else-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 p-4 text-[12.5px] text-red-700">
      <CircleAlert :size="16" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <template v-else-if="detail">
      <div
        v-if="publicURL"
        class="mt-4 flex flex-wrap items-center gap-2 rounded-xl border border-ink-900/10 bg-gradient-to-r from-brand-50/80 via-white to-brand-50/70 px-4 py-3 dark:border-white/10 dark:from-brand-400/8 dark:via-transparent dark:to-brand-400/8"
      >
        <span class="flex size-7 items-center justify-center rounded-lg bg-white text-brand-700 shadow-sm dark:bg-white/10 dark:text-brand-200">
          <Palette :size="13" />
        </span>
        <p class="text-[11.5px] text-ink-900/45">Public company address</p>
        <span class="font-mono text-[11px] font-medium text-ink-900/70">{{ publicURL }}</span>
        <a :href="publicURL" target="_blank" rel="noopener" class="ml-auto inline-flex items-center gap-1 text-[11px] font-semibold text-brand-600 hover:text-brand-700">
          Open <ExternalLink :size="12" />
        </a>
      </div>

      <BrandingSettings
        class="mt-4"
        :branding="detail.branding"
        :subdomain="detail.subdomain"
        :can-manage="can('branding.manage')"
        @updated="updateBranding"
        @subdomain-updated="updateSubdomain"
      />

      <div class="mt-4 flex items-start gap-2 rounded-xl border border-ink-900/10 bg-canvas p-4 text-[11px] leading-relaxed text-ink-900/45">
        <Sparkles :size="14" class="mt-0.5 shrink-0 text-brand-500" />
        <span>
          <strong class="font-semibold text-ink-900/60">Installable app:</strong>
          when staff visit the company subdomain, the app name, short name, icon and theme colour are
          generated from this branding. One frontend deployment serves every company; no separate build
          is required.
        </span>
      </div>
    </template>
  </div>
</template>
