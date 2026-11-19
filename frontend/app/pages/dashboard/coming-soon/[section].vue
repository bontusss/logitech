<script setup lang="ts">
import { ArrowLeft, CircleCheck } from 'lucide-vue-next'

definePageMeta({ middleware: 'auth', layout: 'app' })

const route = useRoute()

const sections: Record<string, { title: string; description: string }> = {
  deliveries: {
    title: 'Deliveries',
    description: 'Manage delivery assignments, driver progress and proof of delivery from one place.',
  },
  customers: {
    title: 'Customers',
    description: 'A searchable directory of the people and businesses booking deliveries.',
  },
  requests: {
    title: 'Booking requests',
    description: 'Review and respond to new booking requests from your public booking page.',
  },
  analytics: {
    title: 'Analytics',
    description: 'Understand booking activity, delivery performance and trends across your business.',
  },
  finance: {
    title: 'Finance',
    description: 'Track collections, outstanding balances and cash currently held by drivers.',
  },
  rates: {
    title: 'Rates',
    description: 'Create and publish the delivery rates customers can use to plan their shipments.',
  },
  billing: {
    title: 'Billing',
    description: 'Manage your plan, invoices and payment methods.',
  },
}

const section = computed(() => sections[String(route.params.section)] ?? null)
useSeoMeta({ title: () => section.value ? `${section.value.title} · BOL` : 'Dashboard · BOL' })
</script>

<template>
  <div class="mx-auto max-w-3xl">
    <div class="rounded-2xl border border-ink-900/10 bg-white p-6 sm:p-8">
      <span class="inline-flex items-center gap-1.5 rounded-full border border-brand-100 bg-brand-50 px-2.5 py-1 text-[9.5px] font-semibold tracking-wide text-brand-700 uppercase">
        <CircleCheck :size="12" /> Coming soon
      </span>
      <template v-if="section">
        <h1 class="mt-4 text-[clamp(1.6rem,3vw,2rem)] font-semibold tracking-[-0.035em]">{{ section.title }}</h1>
        <p class="mt-2 max-w-xl text-[13px] leading-relaxed text-ink-900/50">{{ section.description }}</p>
        <p class="mt-5 rounded-xl bg-canvas p-4 text-[11.5px] leading-relaxed text-ink-900/45">
          This workspace is being prepared. Your current company profile, drivers and booking page remain available from the sidebar.
        </p>
      </template>
      <template v-else>
        <h1 class="mt-4 text-[clamp(1.6rem,3vw,2rem)] font-semibold tracking-[-0.035em]">Section not found</h1>
        <p class="mt-2 text-[13px] leading-relaxed text-ink-900/50">That dashboard section doesn’t exist.</p>
      </template>
      <NuxtLink
        to="/dashboard"
        class="mt-5 inline-flex items-center gap-1.5 text-[12px] font-semibold text-brand-700 hover:text-brand-800"
      >
        <ArrowLeft :size="14" /> Back to overview
      </NuxtLink>
    </div>
  </div>
</template>
