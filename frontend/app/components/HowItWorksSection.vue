<script setup lang="ts">
import { ArrowRight, ClipboardList, Search, Truck } from 'lucide-vue-next'

const steps = [
  {
    icon: ClipboardList,
    title: 'Tell us about your package',
    body: 'Add your route, package size and weight so providers can return a relevant option.',
  },
  {
    icon: Search,
    title: 'Compare your options',
    body: 'Review the price, estimated delivery time and hand-off type side by side.',
  },
  {
    icon: Truck,
    title: 'Choose how it gets there',
    body: 'Select a provider and service, confirm the details and keep track of your shipment.',
  },
]
</script>

<template>
  <section id="how-it-works" class="border-t border-ink-900/10 bg-white">
    <div class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-24">
      <div class="mx-auto max-w-[680px] text-center">
        <p class="text-[10px] font-bold tracking-[0.18em] text-brand-600 uppercase">Simple by design</p>
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
          class="relative rounded-2xl border border-ink-900/10 bg-canvas p-6 sm:p-7"
        >
          <div class="flex items-center justify-between">
            <div class="flex size-10 items-center justify-center rounded-xl bg-brand-50 text-brand-600">
              <component :is="step.icon" :size="18" />
            </div>
            <span class="text-[11px] font-semibold tracking-[0.12em] text-ink-900/30">0{{ index + 1 }}</span>
          </div>
          <h3 class="mt-6 text-[16px] font-semibold tracking-[-0.02em]">{{ step.title }}</h3>
          <p class="mt-2 text-[13px] leading-[1.7] text-ink-900/55">{{ step.body }}</p>
          <ArrowRight
            v-if="index < steps.length - 1"
            :size="16"
            class="absolute top-1/2 -right-2.5 z-10 hidden -translate-y-1/2 text-brand-500 md:block"
          />
        </li>
      </ol>
    </div>
  </section>
</template>
