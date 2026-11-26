<script setup lang="ts">
import { ArrowLeft, ArrowRight, Clock3, Sparkles } from 'lucide-vue-next'

withDefaults(
  defineProps<{
    eyebrow?: string
    title: string
    description: string
    features?: string[]
  }>(),
  { eyebrow: 'Customer dashboard', features: () => [] },
)
</script>

<template>
  <div>
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">{{ eyebrow }}</p>
        <h1 class="mt-2 text-[clamp(1.5rem,2.6vw,1.9rem)] font-semibold tracking-[-0.035em]">{{ title }}</h1>
        <p class="mt-1.5 max-w-2xl text-[12.5px] leading-relaxed text-ink-900/50">{{ description }}</p>
      </div>
      <span class="inline-flex items-center gap-1.5 rounded-full border border-accent-300/70 bg-accent-50 px-3 py-1.5 text-[10px] font-semibold tracking-wide text-accent-700 uppercase dark:border-accent-400/20 dark:bg-accent-400/10 dark:text-accent-300">
        <Clock3 :size="12" /> Coming soon
      </span>
    </div>

    <section class="mt-5 overflow-hidden rounded-2xl border border-ink-900/10 bg-surface">
      <div class="relative border-b border-ink-900/8 bg-gradient-to-br from-brand-50/80 via-surface to-accent-50/60 px-5 py-10 text-center dark:from-brand-400/5 dark:via-surface dark:to-accent-400/5">
        <span class="mx-auto flex size-12 items-center justify-center rounded-2xl bg-accent-500 text-white shadow-lg shadow-accent-500/25">
          <Sparkles :size="21" />
        </span>
        <h2 class="mt-4 text-[16px] font-semibold tracking-[-0.02em]">We’re building this next</h2>
        <p class="mx-auto mt-1.5 max-w-md text-[11.5px] leading-relaxed text-ink-900/50">
          This part of your customer dashboard is not live yet. Your shipments and tracking are ready to use in the meantime.
        </p>
      </div>

      <div v-if="features.length" class="p-5">
        <p class="text-[9.5px] font-semibold tracking-[0.14em] text-ink-900/40 uppercase">What’s planned</p>
        <ul class="mt-3 grid gap-2 sm:grid-cols-2">
          <li v-for="feature in features" :key="feature" class="flex items-start gap-2 rounded-xl border border-ink-900/8 bg-canvas/50 px-3 py-2.5">
            <span class="mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-full bg-accent-500/15 text-accent-700">
              <Sparkles :size="9" />
            </span>
            <span class="text-[11px] leading-relaxed text-ink-900/60">{{ feature }}</span>
          </li>
        </ul>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-3 border-t border-ink-900/8 px-5 py-4">
        <NuxtLink to="/dashboard" class="inline-flex items-center gap-1.5 text-[11.5px] font-semibold text-ink-900/55 transition hover:text-ink-900">
          <ArrowLeft :size="13" /> Back to overview
        </NuxtLink>
        <NuxtLink
          to="/dashboard/customer/shipments"
          class="inline-flex items-center gap-1.5 rounded-lg bg-accent-500 px-3.5 py-2 text-[11.5px] font-semibold text-white transition hover:bg-accent-400"
        >
          Go to my shipments <ArrowRight :size="13" />
        </NuxtLink>
      </div>
    </section>
  </div>
</template>
