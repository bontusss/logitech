<script setup lang="ts">
import { ArrowRight, ClipboardList, Search, Truck } from 'lucide-vue-next'

const steps = [
  {
    icon: ClipboardList,
    title: 'Tell us about your package',
    body: 'Add your route, package size and weight so providers can return a relevant option.',
    card: 'from-violet-50 via-white to-white dark:from-violet-400/10 dark:via-transparent dark:to-transparent',
    tile: 'bg-violet-100 text-violet-700 dark:bg-violet-400/15 dark:text-violet-200',
    chip: 'bg-violet-100 text-violet-800 dark:bg-violet-400/15 dark:text-violet-200',
  },
  {
    icon: Search,
    title: 'Compare your options',
    body: 'Review the price, estimated delivery time and hand-off type side by side.',
    card: 'from-sky-50 via-white to-white dark:from-sky-400/10 dark:via-transparent dark:to-transparent',
    tile: 'bg-sky-100 text-sky-700 dark:bg-sky-400/15 dark:text-sky-200',
    chip: 'bg-sky-100 text-sky-800 dark:bg-sky-400/15 dark:text-sky-200',
  },
  {
    icon: Truck,
    title: 'Choose how it gets there',
    body: 'Select a provider and service, confirm the details and keep track of your shipment.',
    card: 'from-emerald-50 via-white to-white dark:from-emerald-400/10 dark:via-transparent dark:to-transparent',
    tile: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-400/15 dark:text-emerald-200',
    chip: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-400/15 dark:text-emerald-200',
  },
]
</script>

<template>
  <section id="how-it-works" class="border-t border-ink-900/10 bg-canvas">
    <div class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-24">
      <div class="mx-auto max-w-[680px] text-center">
        <p class="inline-flex items-center gap-1.5 rounded-full border border-brand-200/70 bg-brand-50/70 px-2.5 py-1 text-[10px] font-bold tracking-[0.18em] text-brand-700 uppercase">
          <span class="size-1.5 rounded-full bg-gradient-to-r from-brand-500 to-violet-500" />
          Simple by design
        </p>
        <h2 class="mt-3 text-[clamp(1.8rem,4vw,2.7rem)] leading-[1.05] font-semibold tracking-[-0.045em]">
          From “how much?” to “on its way.”
        </h2>
        <p class="mt-4 text-[14px] leading-[1.7] text-ink-900/55">
          Everything a shipper needs to make a better-informed delivery choice, in one place.
        </p>
      </div>

      <ol class="mt-10 grid gap-3 md:grid-cols-3">
        <li
          v-for="(step, index) in steps"
          :key="step.title"
          class="site-card relative rounded-2xl border bg-gradient-to-br p-6 sm:p-7"
          :class="step.card"
        >
          <div class="flex items-center justify-between">
            <div class="flex size-10 items-center justify-center rounded-xl shadow-sm" :class="step.tile">
              <component :is="step.icon" :size="18" />
            </div>
            <span class="rounded-full px-2.5 py-1 text-[11px] font-semibold tracking-[0.12em]" :class="step.chip">0{{ index + 1 }}</span>
          </div>
          <h3 class="mt-6 text-[16px] font-semibold tracking-[-0.02em]">{{ step.title }}</h3>
          <p class="mt-2 text-[13px] leading-[1.7] text-ink-900/70">{{ step.body }}</p>
          <span
            v-if="index < steps.length - 1"
            class="absolute top-1/2 -right-3 z-10 hidden size-6 -translate-y-1/2 items-center justify-center rounded-full border border-ink-900/8 bg-white text-brand-600 shadow-sm md:flex dark:border-white/10 dark:bg-slate-800 dark:text-indigo-300"
          >
            <ArrowRight :size="13" />
          </span>
        </li>
      </ol>
    </div>
  </section>
</template>
