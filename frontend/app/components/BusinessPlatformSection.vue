<script setup lang="ts">
import { ArrowRight, MessageCircle } from 'lucide-vue-next'

const { demoUrl, whatsappHref, whatsappUrl } = useSiteConfig()

const groups = [
  {
    icon: 'building' as const,
    title: 'Business management',
    items: [
      'Branded company profile',
      'Company branches and terminals',
      'Staff accounts and role permissions',
      'Customer management',
    ],
  },
  {
    icon: 'truck' as const,
    title: 'Shipment operations',
    items: [
      'Shipment and manifest management',
      'Incoming and outgoing shipments',
      'Interbranch transfers',
      'Driver and vehicle management',
      'Tracking and delivery confirmation',
      'Exception management',
    ],
  },
  {
    icon: 'tag' as const,
    title: 'Pricing and marketplace visibility',
    items: [
      'Publish shipping rates',
      'Configure routes and service areas',
      'Weight-based and fixed pricing',
      'Define delivery timelines',
      'Receive customer bookings',
      'Manage booking confirmations',
    ],
  },
  {
    icon: 'chart' as const,
    title: 'Business intelligence',
    items: [
      'Shipment analytics',
      'Revenue reports',
      'Branch performance',
      'Operational insights',
    ],
  },
]
</script>

<template>
  <section id="business" class="overflow-hidden border-t border-ink-900/10 bg-white">
    <div class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-24">
      <div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
        <SectionHeading
          eyebrow="For logistics businesses"
          title="Run your logistics business. Reach more customers."
          lede="Everything you need to manage your logistics operations, digitise your workflows and grow your customer base, all from one platform."
        />

        <div class="flex flex-col gap-2.5 sm:flex-row lg:pb-1">
          <AppButton :to="demoUrl" size="lg">
            Book a Demo
            <ArrowRight :size="16" />
          </AppButton>
          <AppButton v-if="whatsappUrl" :to="whatsappHref" variant="secondary" size="lg">
            <MessageCircle :size="16" />
            Chat on WhatsApp
          </AppButton>
        </div>
      </div>

      <div class="mt-12">
        <DashboardMockup />
      </div>

      <div id="features" class="mt-16 scroll-mt-28">
        <div class="flex items-center gap-3">
          <h3 class="text-[11.5px] font-semibold tracking-[0.18em] text-ink-900/45 uppercase">
            Platform features
          </h3>
          <span class="h-px flex-1 bg-ink-900/10" />
        </div>
        <FeatureGrid :groups="groups" class="mt-6" />
      </div>
    </div>
  </section>
</template>
