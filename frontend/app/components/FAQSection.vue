<script setup lang="ts">
import { ChevronDown } from 'lucide-vue-next'

const { demoUrl } = useSiteConfig()

const faqs = [
  {
    q: 'How does the shipping marketplace work?',
    a: 'You enter the pickup city, destination, package weight and category, and the platform shows logistics providers whose published routes and rates match. You can compare options and book directly. The marketplace is being rolled out progressively, so the rate comparison shown today is an illustrative preview rather than a live quotation.',
  },
  {
    q: 'Can I compare different logistics companies?',
    a: 'Yes. The comparison view is designed to show price, estimated delivery time, service type and pickup or drop-off availability side by side, so you can choose based on what matters for that shipment.',
  },
  {
    q: 'How can a logistics company join the platform?',
    a: 'A logistics business creates a company account, sets up its profile, branches and staff, then publishes its routes and rates. You can book a demo first if you would like a walkthrough before setting up.',
  },
  {
    q: 'Can I manage multiple branches?',
    a: 'Yes. The business platform is built around multi-branch operations, including branch-level users, interbranch transfers and visibility across locations. The number of branches available depends on your plan.',
  },
  {
    q: 'Can I publish my own shipping rates?',
    a: 'Yes. Providers can configure routes and service areas, use weight-based or fixed pricing, and define delivery timelines. Those rates are what customers see when comparing options.',
  },
  {
    q: 'Can customers book shipments directly?',
    a: 'Yes, where a provider has enabled marketplace bookings. Bookings arrive in the business workspace for confirmation and are then managed like any other shipment.',
  },
  {
    q: 'How does shipment tracking work?',
    a: 'Shipments move through recorded statuses from acceptance and dispatch through transit, arrival and delivery. Customers and branch staff see the latest status and delivery confirmation in one place.',
  },
  {
    q: 'Can I use my company\'s branding?',
    a: 'Your company profile carries your business name, logo and branch details, so customers know who they are booking with. Deeper branding and custom workflows are scoped with Enterprise.',
  },
  {
    q: 'Is the platform available across Nigeria?',
    a: 'We are starting with intercity routes between major commercial cities such as Lagos, Aba, Onitsha, Enugu, Abuja, Port Harcourt and Umuahia. Coverage expands as more providers publish routes; not every route is available today.',
  },
  {
    q: 'How much does it cost?',
    a: 'Plans are quoted based on the size of your operation, so pricing is available on request. Marketplace booking fees or transaction-related charges may be separate from the SaaS subscription. Contact sales or book a demo for a specific quote.',
  },
]
</script>

<template>
  <section id="faq" class="border-t border-ink-900/10 bg-canvas">
    <div class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-24">
      <div class="grid gap-10 lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)] lg:gap-16">
        <div>
          <SectionHeading
            eyebrow="FAQ"
            title="Questions, answered."
            lede="What customers and logistics businesses usually ask before getting started."
          />
          <div class="mt-6 rounded-xl border border-ink-900/10 bg-white p-4">
            <p class="text-[12px] font-semibold">Still deciding?</p>
            <p class="mt-1 text-[12px] leading-relaxed text-ink-900/50">
              Talk to the team about your routes, branches or shipping needs.
            </p>
            <NuxtLink
              :to="demoUrl"
              class="mt-3 inline-flex items-center gap-1.5 text-[12.5px] font-semibold text-brand-600 hover:text-brand-700"
            >
              Book a demo
              <span aria-hidden="true">→</span>
            </NuxtLink>
          </div>
        </div>

        <div class="divide-y divide-ink-900/10 border-y border-ink-900/10">
          <details v-for="(faq, index) in faqs" :key="faq.q" :name="'faq'" class="group py-1" :open="index === 0">
            <summary
              class="flex cursor-pointer list-none items-start justify-between gap-4 py-4 text-[14px] leading-snug font-medium text-ink-900 marker:hidden [&::-webkit-details-marker]:hidden"
            >
              {{ faq.q }}
              <ChevronDown
                :size="16"
                class="mt-0.5 shrink-0 text-ink-900/40 transition duration-200 group-open:rotate-180"
              />
            </summary>
            <p class="pr-8 pb-5 text-[13.5px] leading-[1.7] text-ink-900/55">{{ faq.a }}</p>
          </details>
        </div>
      </div>
    </div>
  </section>
</template>
