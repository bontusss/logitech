<script setup lang="ts">
import { ChevronDown } from 'lucide-vue-next'

const faqs = [
  {
    q: 'Are the rates on this page live?',
    a: 'Not yet. The comparison is an illustrative preview, not a live quote. We are building the public provider-rate and booking flow and will only display real provider prices once that connection is ready.',
  },
  {
    q: 'Can I book a delivery on BOL today?',
    a: 'Direct marketplace booking is not available yet. We are working toward a flow where you can compare real provider services and book without calling around.',
  },
  {
    q: 'What kind of shipments will BOL support?',
    a: 'BOL is designed for intercity parcel deliveries, including documents, everyday parcels, electronics and larger shipments. What you can send will depend on each provider’s service rules.',
  },
  {
    q: 'Will I be able to compare delivery times as well as prices?',
    a: 'Yes. The marketplace is designed to show key details such as delivery estimate, service type and pickup or drop-off options alongside the rate.',
  },
  {
    q: 'I run a logistics company. Where can I learn more?',
    a: 'Visit the logistics business page to see how BOL supports branches, staff, pricing and shipment operations, then book a demo or contact the team on WhatsApp.',
  },
]
</script>

<template>
  <section id="faq" class="border-t border-ink-900/10 bg-canvas">
    <div class="mx-auto max-w-[920px] px-5 py-20 sm:px-8 sm:py-24">
      <div class="mb-9">
        <SectionHeading eyebrow="Good to know" title="A few quick answers." align="center" />
      </div>

      <div class="site-panel overflow-hidden rounded-3xl border px-5 sm:px-7">
        <details
          v-for="(faq, index) in faqs"
          :key="faq.q"
          name="shipper-faq"
          class="group border-b border-ink-900/8 py-1 last:border-0"
          :open="index === 0"
        >
          <summary class="flex cursor-pointer list-none items-start justify-between gap-4 py-4 text-[14px] font-semibold text-ink-900 transition marker:hidden hover:text-accent-700 [&::-webkit-details-marker]:hidden dark:hover:text-accent-400">
            {{ faq.q }}
            <ChevronDown :size="16" class="mt-0.5 shrink-0 text-ink-900/40 transition group-open:rotate-180" />
          </summary>
          <p class="max-w-[740px] pr-8 pb-5 text-[13px] leading-[1.75] text-ink-900/65">
            {{ faq.a }}
            <NuxtLink
              v-if="index === 4"
              to="/for-logistics-businesses"
              class="ml-1 font-semibold text-accent-700 hover:text-accent-600 dark:text-accent-400"
            >
              Explore BOL for businesses →
            </NuxtLink>
          </p>
        </details>
      </div>
    </div>
  </section>
</template>
