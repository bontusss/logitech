<script setup lang="ts">
import { ArrowRight, ClipboardList, Search, Truck } from 'lucide-vue-next'

const steps = [
  {
    icon: ClipboardList,
    title: 'Tell us about your package',
    body: 'Add your route, package size and weight so providers can return a relevant option.',
    tone: 'brand' as const,
  },
  {
    icon: Search,
    title: 'Compare your options',
    body: 'Review the price, estimated delivery time and hand-off type side by side.',
    tone: 'accent' as const,
  },
  {
    icon: Truck,
    title: 'Choose how it gets there',
    body: 'Select a provider and service, confirm the details and keep track of your shipment.',
    tone: 'success' as const,
  },
]
</script>

<template>
  <section id="how-it-works" class="border-t border-ink-900/10 bg-canvas">
    <div class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-24">
      <SectionHeading
        eyebrow="Simple by design"
        title="From “how much?” to “on its way.”"
        lede="Everything a shipper needs to make a better-informed delivery choice, in one place."
        align="center"
      />

      <ol class="mt-10 grid gap-3 md:grid-cols-3">
        <li
          v-for="(step, index) in steps"
          :key="step.title"
          class="site-card relative rounded-2xl border p-6 sm:p-7"
        >
          <div class="flex items-center justify-between">
            <IconTile :tone="step.tone">
              <component :is="step.icon" :size="18" />
            </IconTile>
            <span class="rounded-full border border-ink-900/10 bg-canvas px-2.5 py-1 text-[11px] font-semibold tracking-[0.12em] text-ink-900/50">
              0{{ index + 1 }}
            </span>
          </div>
          <h3 class="mt-6 text-[16px] font-semibold tracking-[-0.02em]">{{ step.title }}</h3>
          <p class="mt-2 text-[13px] leading-[1.7] text-ink-900/65">{{ step.body }}</p>
          <span
            v-if="index < steps.length - 1"
            class="absolute top-1/2 -right-3 z-10 hidden size-6 -translate-y-1/2 items-center justify-center rounded-full border border-ink-900/10 bg-surface text-accent-600 shadow-sm md:flex dark:text-accent-400"
          >
            <ArrowRight :size="13" />
          </span>
        </li>
      </ol>
    </div>
  </section>
</template>
