<script setup lang="ts">
import { ArrowLeftRight, ClipboardList, Search } from 'lucide-vue-next'

const pillars = [
  {
    icon: Search,
    title: 'Discover logistics providers',
    body: 'Find operators by route and service instead of relying on scattered contacts and referrals.',
  },
  {
    icon: ArrowLeftRight,
    title: 'Compare shipping services',
    body: 'See price, estimated delivery time and service type side by side before you commit.',
  },
  {
    icon: ClipboardList,
    title: 'Manage logistics operations',
    body: 'Run branches, staff, shipments, manifests and pricing from a single business workspace.',
  },
]
</script>

<template>
  <section class="border-y border-ink-900/10 bg-white/60">
    <div class="mx-auto max-w-[1200px] px-5 py-16 sm:px-8 sm:py-20">
      <div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(0,0.85fr)] lg:items-end">
        <SectionHeading
          eyebrow="One connected network"
          title="Connecting Africa's logistics ecosystem."
          lede="Independent couriers, established transport companies, businesses and individuals have historically moved goods in separate lanes. BOL brings them into one connected place."
        />
        <p class="max-w-md text-[14.5px] leading-[1.7] text-ink-900/55 lg:justify-self-end">
          Providers keep their own operations, identity and pricing. Customers get a single view of
          who can move a shipment, how long it takes and what it costs.
        </p>
      </div>

      <div
        class="mt-10 grid gap-px overflow-hidden rounded-xl border border-ink-900/10 bg-ink-900/10 sm:grid-cols-3"
      >
        <div v-for="pillar in pillars" :key="pillar.title" class="bg-white p-6">
          <div
            class="flex size-9 items-center justify-center rounded-lg border border-brand-100 bg-brand-50 text-brand-600"
          >
            <component :is="pillar.icon" :size="17" />
          </div>
          <h3 class="mt-4 text-[15px] font-semibold tracking-[-0.01em]">{{ pillar.title }}</h3>
          <p class="mt-1.5 text-[13.5px] leading-[1.65] text-ink-900/55">{{ pillar.body }}</p>
        </div>
      </div>
    </div>
  </section>
</template>
