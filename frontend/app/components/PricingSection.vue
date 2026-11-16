<script setup lang="ts">
import { ArrowRight, Check } from 'lucide-vue-next'

const { demoUrl } = useSiteConfig()

const tiers = [
  {
    name: 'Starter',
    audience: 'For independent operators getting onto the platform.',
    features: [
      'Basic company profile',
      'Limited branches',
      'Basic shipment management',
      'Basic rate publishing',
      'Essential reporting',
    ],
    cta: 'Contact Sales',
    to: 'mailto:hello@basedonlogistics.com?subject=BOL%20Starter%20plan',
    featured: false,
  },
  {
    name: 'Growth',
    audience: 'For growing multi-branch logistics businesses.',
    features: [
      'Multiple branches',
      'Staff management',
      'Advanced shipment management',
      'Rate management',
      'Marketplace bookings',
      'Analytics and reporting',
    ],
    cta: 'Book a Demo',
    to: demoUrl,
    featured: true,
  },
  {
    name: 'Enterprise',
    audience: 'For established networks and franchise operations.',
    features: [
      'Advanced branch management',
      'Custom roles and permissions',
      'Advanced analytics',
      'API integrations',
      'Priority support',
      'Custom workflows',
    ],
    cta: 'Contact Sales',
    to: 'mailto:hello@basedonlogistics.com?subject=BOL%20Enterprise%20plan',
    featured: false,
  },
]
</script>

<template>
  <section id="pricing" class="border-t border-ink-900/10 bg-white">
    <div class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-24">
      <SectionHeading
        align="center"
        eyebrow="Pricing"
        title="Plans that scale with your operation."
        lede="Start with the essentials and add branches, staff and marketplace capabilities as your business grows."
      />

      <div class="mx-auto mt-10 grid max-w-[1050px] gap-4 lg:grid-cols-3">
        <div
          v-for="tier in tiers"
          :key="tier.name"
          class="flex flex-col rounded-xl border p-6"
          :class="
            tier.featured
              ? 'border-ink-900 bg-ink-950 text-white shadow-[0_30px_70px_-45px_rgba(10,18,32,0.8)]'
              : 'border-ink-900/10 bg-white'
          "
        >
          <div class="flex items-center justify-between">
            <h3 class="text-[17px] font-semibold tracking-[-0.02em]">{{ tier.name }}</h3>
            <span
              v-if="tier.featured"
              class="rounded border border-brand-500/30 bg-brand-500/15 px-2 py-0.5 text-[9.5px] font-semibold tracking-[0.1em] text-brand-200 uppercase"
            >
              Most popular
            </span>
          </div>

          <p class="mt-2 text-[12.5px] leading-relaxed" :class="tier.featured ? 'text-white/50' : 'text-ink-900/50'">
            {{ tier.audience }}
          </p>

          <p
            class="mt-5 text-[13px] font-medium"
            :class="tier.featured ? 'text-white/70' : 'text-ink-900/65'"
          >
            Pricing on request
          </p>

          <div class="mt-5 h-px w-full" :class="tier.featured ? 'bg-white/10' : 'bg-ink-900/8'" />

          <ul class="mt-5 flex-1 space-y-2.5">
            <li
              v-for="feature in tier.features"
              :key="feature"
              class="flex items-start gap-2 text-[13px] leading-snug"
              :class="tier.featured ? 'text-white/65' : 'text-ink-900/65'"
            >
              <Check
                :size="13"
                class="mt-0.5 shrink-0"
                :class="tier.featured ? 'text-brand-200' : 'text-brand-600'"
              />
              {{ feature }}
            </li>
          </ul>

          <div class="mt-6">
            <AppButton
              :to="tier.to"
              :variant="tier.featured ? 'light' : 'secondary'"
              block
            >
              {{ tier.cta }}
              <ArrowRight :size="14" />
            </AppButton>
          </div>
        </div>
      </div>

      <p class="mx-auto mt-6 max-w-xl text-center text-[11.5px] leading-relaxed text-ink-900/45">
        Marketplace booking fees or other transaction-related charges may be separate from SaaS
        subscriptions. Final plan details and pricing are confirmed with our team.
      </p>
    </div>
  </section>
</template>
