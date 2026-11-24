<script setup lang="ts">
import { ArrowDown, ShieldCheck, Sparkles, Truck } from 'lucide-vue-next'
import type {
  MarketplaceSearchPayload,
  MarketplaceSearchResponse,
} from '~/types/marketplace'
import { apiErrorMessage } from '~/composables/useApi'

const { api } = useApi()
const searchLoading = ref(false)
const searchError = ref('')
const searchResponse = ref<MarketplaceSearchResponse | null>(null)
const searchPayload = ref<MarketplaceSearchPayload | null>(null)
let requestId = 0

async function runSearch(payload: MarketplaceSearchPayload) {
  const currentRequest = ++requestId
  searchPayload.value = payload
  searchLoading.value = true
  searchError.value = ''
  searchResponse.value = null

  try {
    const response = await api<MarketplaceSearchResponse>('/marketplace/search', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
    if (currentRequest === requestId) searchResponse.value = response
  } catch (error: unknown) {
    if (currentRequest === requestId) {
      searchError.value = apiErrorMessage(
        error,
        'The search could not be completed. Please try again.',
      )
    }
  } finally {
    if (currentRequest === requestId) searchLoading.value = false
  }
}
</script>

<template>
  <section id="top" class="marketing-hero relative overflow-hidden">
    <div aria-hidden="true" class="pointer-events-none absolute inset-0">
      <div class="absolute inset-0 grid-paper opacity-50 [mask-image:linear-gradient(to_bottom,black,transparent_85%)] dark:opacity-30" />
      <div class="absolute -top-52 right-[-8rem] h-[620px] w-[620px] rounded-full bg-accent-500/15 blur-[120px]" />
      <div class="absolute top-24 left-[-12rem] h-[460px] w-[460px] rounded-full bg-brand-600/10 blur-[110px] dark:bg-brand-400/10" />
      <div class="absolute bottom-[-18rem] left-[-10rem] h-[500px] w-[500px] rounded-full bg-accent-500/10 blur-[110px]" />
    </div>

    <div class="relative mx-auto max-w-[1200px] px-5 pt-14 pb-16 sm:px-8 sm:pt-20 lg:pt-24 lg:pb-24">
      <div class="mx-auto max-w-[780px] text-center">
        <SectionEyebrow
          tone="surface"
          class="hero-country-badge reveal"
        >
          <span class="flex size-4 items-center justify-center rounded-full bg-accent-500 text-ink-950">
            <Sparkles :size="9" />
          </span>
          Nigeria, one delivery at a time
        </SectionEyebrow>

        <h1 class="reveal reveal-1 mt-6 text-[clamp(2.8rem,6.2vw,5.2rem)] leading-[0.94] font-semibold tracking-[-0.06em] text-balance">
          Send it right.
          <span class="block text-accent-600 dark:text-accent-400">
            Know your options.
          </span>
        </h1>

        <p class="reveal reveal-2 mx-auto mt-6 max-w-[620px] text-[16px] leading-[1.7] text-ink-900/65 sm:text-[17px]">
          Compare intercity delivery options by price and speed, then choose the service that works for you. No more calling around to figure out where to start.
        </p>

        <div class="reveal reveal-3 mt-6 flex flex-wrap items-center justify-center gap-x-3 gap-y-2 text-[12px] font-semibold text-ink-900/75">
          <SurfaceChip class="px-3 py-1.5">
            <ShieldCheck :size="15" class="text-success-600" /> Compare with clarity
          </SurfaceChip>
          <SurfaceChip class="px-3 py-1.5">
            <Truck :size="15" class="text-ink-900/70" /> Built for intercity shipping
          </SurfaceChip>
        </div>
      </div>

      <div class="reveal reveal-2 mx-auto mt-10 grid max-w-[1080px] gap-4 lg:grid-cols-2 lg:items-start">
        <ShippingRateSearch :loading="searchLoading" @search="runSearch" />
        <div>
          <LiveRateComparison
            :loading="searchLoading"
            :error="searchError"
            :response="searchResponse"
            :search="searchPayload"
          />
          <NuxtLink
            to="/#how-it-works"
            class="mt-4 inline-flex items-center gap-2 rounded-full border border-ink-900/10 bg-surface/85 px-3.5 py-2 text-[12.5px] font-semibold text-ink-900/70 shadow-sm backdrop-blur-sm transition hover:-translate-y-0.5 hover:border-accent-500/40 hover:text-ink-900"
          >
            How shipping with BOL works
            <ArrowDown :size="14" class="text-accent-600 dark:text-accent-400" />
          </NuxtLink>
        </div>
      </div>
    </div>
  </section>
</template>
