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
      <div class="absolute inset-0 grid-paper opacity-40 [mask-image:linear-gradient(to_bottom,black,transparent_85%)] dark:opacity-25" />
      <div class="absolute -top-56 right-[-8rem] h-[620px] w-[620px] rounded-full bg-brand-500/15 blur-[120px]" />
      <div class="absolute top-24 left-[-12rem] h-[420px] w-[420px] rounded-full bg-violet-400/15 blur-[110px]" />
      <div class="absolute bottom-[-18rem] left-[-10rem] h-[500px] w-[500px] rounded-full bg-[#ffc76b]/20 blur-[100px]" />
      <div class="absolute right-[-6rem] bottom-[-14rem] h-[420px] w-[420px] rounded-full bg-cyan-300/20 blur-[110px]" />
    </div>

    <div class="relative mx-auto max-w-[1200px] px-5 pt-14 pb-16 sm:px-8 sm:pt-20 lg:pt-24 lg:pb-24">
      <div class="mx-auto max-w-[780px] text-center">
        <p class="hero-country-badge reveal inline-flex items-center gap-2 rounded-full border border-brand-200/70 bg-white/85 px-3 py-1.5 text-[11px] font-bold tracking-[0.12em] text-brand-700 uppercase shadow-sm backdrop-blur-sm dark:border-white/10 dark:bg-white/5 dark:text-indigo-200">
          <span class="flex size-4 items-center justify-center rounded-full bg-gradient-to-br from-brand-500 to-violet-500 text-white">
            <Sparkles :size="9" />
          </span>
          Nigeria, one delivery at a time
        </p>

        <h1 class="reveal reveal-1 mt-6 text-[clamp(2.8rem,6.2vw,5.2rem)] leading-[0.94] font-semibold tracking-[-0.06em] text-balance">
          Send it right.
          <span class="block bg-gradient-to-r from-brand-600 via-violet-600 to-cyan-600 bg-clip-text pb-1 text-transparent dark:from-indigo-300 dark:via-violet-300 dark:to-cyan-300">
            Know your options.
          </span>
        </h1>

        <p class="reveal reveal-2 mx-auto mt-6 max-w-[620px] text-[16px] leading-[1.7] text-ink-900/65 sm:text-[17px]">
          Compare intercity delivery options by price and speed, then choose the service that works for you. No more calling around to figure out where to start.
        </p>

        <div class="reveal reveal-3 mt-6 flex flex-wrap items-center justify-center gap-x-3 gap-y-2 text-[12px] font-semibold text-ink-900/75">
          <span class="inline-flex items-center gap-1.5 rounded-full border border-white/90 bg-white/85 px-3 py-1.5 shadow-sm dark:border-white/10 dark:bg-white/5">
            <ShieldCheck :size="15" class="text-emerald-600 dark:text-emerald-300" /> Compare with clarity
          </span>
          <span class="inline-flex items-center gap-1.5 rounded-full border border-white/90 bg-white/85 px-3 py-1.5 shadow-sm dark:border-white/10 dark:bg-white/5">
            <Truck :size="15" class="text-sky-600 dark:text-sky-300" /> Built for intercity shipping
          </span>
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
            class="mt-4 inline-flex items-center gap-2 rounded-full border border-white/90 bg-white/85 px-3.5 py-2 text-[12.5px] font-semibold text-ink-900/75 shadow-sm transition hover:-translate-y-0.5 hover:border-brand-200 hover:text-brand-700 dark:border-white/10 dark:bg-white/5"
          >
            How shipping with BOL works
            <ArrowDown :size="14" />
          </NuxtLink>
        </div>
      </div>
    </div>
  </section>
</template>
