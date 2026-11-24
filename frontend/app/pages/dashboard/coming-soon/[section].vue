<script setup lang="ts">
import {
  ArrowLeft,
  ChartNoAxesColumnIncreasing,
  CircleCheck,
  ClipboardList,
  CreditCard,
  Settings,
  Sparkles,
  Truck,
} from 'lucide-vue-next'

definePageMeta({ middleware: 'auth', layout: 'app' })

const route = useRoute()

const sections: Record<
  string,
  { title: string; description: string; icon: typeof Truck; iconClass: string }
> = {
  deliveries: {
    title: 'Deliveries',
    description: 'Manage delivery assignments, driver progress and proof of delivery from one place.',
    icon: Truck,
    iconClass: 'bg-brand-100 text-brand-700 dark:bg-brand-400/15 dark:text-brand-200',
  },
  requests: {
    title: 'Booking requests',
    description: 'Review and respond to new booking requests from your public booking page.',
    icon: ClipboardList,
    iconClass: 'bg-accent-100 text-accent-700 dark:bg-accent-400/15 dark:text-accent-200',
  },
  analytics: {
    title: 'Analytics',
    description: 'Understand booking activity, delivery performance and trends across your business.',
    icon: ChartNoAxesColumnIncreasing,
    iconClass: 'bg-success-100 text-success-700 dark:bg-success-500/15 dark:text-success-200',
  },
  finance: {
    title: 'Finance',
    description: 'Track collections, outstanding balances and cash currently held by drivers.',
    icon: CreditCard,
    iconClass: 'bg-brand-100 text-brand-700 dark:bg-brand-400/15 dark:text-brand-200',
  },
  rates: {
    title: 'Rates',
    description: 'Create and publish the delivery rates customers can use to plan their shipments.',
    icon: ClipboardList,
    iconClass: 'bg-brand-100 text-brand-700 dark:bg-brand-400/15 dark:text-brand-200',
  },
  billing: {
    title: 'Billing',
    description: 'Manage your plan, invoices and payment methods.',
    icon: Settings,
    iconClass: 'bg-accent-100 text-accent-700 dark:bg-accent-400/15 dark:text-accent-200',
  },
}

const section = computed(() => sections[String(route.params.section)] ?? null)
useSeoMeta({ title: () => section.value ? `${section.value.title} · BOL` : 'Dashboard · BOL' })
</script>

<template>
  <div class="dash-page mx-auto max-w-3xl">
    <div class="dash-hero relative overflow-hidden rounded-2xl border p-6 sm:p-8">
      <div class="pointer-events-none absolute -top-20 right-6 size-48 rounded-full bg-accent-400/15 blur-3xl" />
      <div class="pointer-events-none absolute -right-10 -bottom-24 size-56 rounded-full bg-brand-400/15 blur-3xl" />
      <div class="relative">
        <span class="inline-flex items-center gap-1.5 rounded-full border border-white/70 bg-white/75 px-2.5 py-1 text-[9.5px] font-semibold tracking-wide text-brand-700 uppercase dark:border-white/10 dark:bg-slate-950/30 dark:text-brand-200">
          <CircleCheck :size="12" /> Coming soon
        </span>
        <template v-if="section">
          <div class="mt-4 flex items-start gap-3.5">
            <span class="flex size-12 shrink-0 items-center justify-center rounded-2xl shadow-sm" :class="section.iconClass">
              <component :is="section.icon" :size="22" />
            </span>
            <div>
              <h1 class="text-[clamp(1.5rem,3vw,1.9rem)] leading-tight font-semibold tracking-[-0.04em]">
                {{ section.title }}
              </h1>
              <p class="mt-1.5 max-w-xl text-[12.5px] leading-relaxed text-ink-900/60">
                {{ section.description }}
              </p>
            </div>
          </div>
          <p class="mt-6 flex items-start gap-2 rounded-xl border border-white/60 bg-white/60 p-4 text-[11.5px] leading-relaxed text-ink-900/50 dark:border-white/10 dark:bg-white/5">
            <Sparkles :size="14" class="mt-0.5 shrink-0 text-brand-500" />
            This workspace is being prepared. Your current company profile, drivers and booking page
            remain available from the sidebar.
          </p>
        </template>
        <template v-else>
          <h1 class="mt-4 text-[clamp(1.5rem,3vw,1.9rem)] font-semibold tracking-[-0.04em]">Section not found</h1>
          <p class="mt-2 text-[12.5px] leading-relaxed text-ink-900/60">That dashboard section doesn’t exist.</p>
        </template>
        <NuxtLink
          to="/dashboard"
          class="mt-5 inline-flex items-center gap-1.5 rounded-full bg-accent-500 px-3.5 py-2 text-[11.5px] font-semibold text-ink-950 shadow-sm shadow-accent-500/25 transition hover:-translate-x-0.5 hover:bg-accent-400"
        >
          <ArrowLeft :size="14" /> Back to overview
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
