<script setup lang="ts">
import { ArrowRight, Check, MapPin } from 'lucide-vue-next'

const features = [
  'Compare shipping rates across providers',
  'Discover verified logistics providers',
  'Track shipments from pickup to delivery',
  'Manage bookings and delivery details',
  'Access multiple providers from one platform',
]

const routes = ['Aba → Lagos', 'Onitsha → Abuja', 'Port Harcourt → Enugu', 'Lagos → Umuahia']
</script>

<template>
  <section id="marketplace" class="relative">
    <div class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-24">
      <div class="grid gap-12 lg:grid-cols-2 lg:items-center">
        <div>
          <SectionHeading
            eyebrow="Marketplace"
            title="Find the right logistics partner for every shipment."
            lede="Compare shipping options from multiple logistics providers and choose based on price, delivery time and service coverage."
          />

          <ul class="mt-7 grid gap-3 sm:grid-cols-2">
            <li
              v-for="feature in features"
              :key="feature"
              class="flex items-start gap-2.5 text-[13.5px] leading-snug text-ink-900/70"
            >
              <span
                class="mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-full bg-brand-100 text-brand-700"
              >
                <Check :size="10" :stroke-width="3" />
              </span>
              {{ feature }}
            </li>
          </ul>

          <div class="mt-7 flex flex-wrap gap-1.5">
            <span
              v-for="route in routes"
              :key="route"
              class="inline-flex items-center gap-1.5 rounded-md border border-ink-900/10 bg-white px-2.5 py-1.5 text-[11.5px] font-medium text-ink-900/60"
            >
              <MapPin :size="11" class="text-brand-600" />
              {{ route }}
            </span>
          </div>

          <div class="mt-8">
            <AppButton to="#rate-search">
              Explore the Marketplace
              <ArrowRight :size="15" />
            </AppButton>
          </div>
        </div>

        <div>
          <RateComparisonPreview />
          <p class="mt-3 px-1 text-[11.5px] leading-relaxed text-ink-900/40">
            Example comparison for illustration. Provider names, prices and delivery times shown
            here are sample values, not live quotations.
          </p>
        </div>
      </div>
    </div>
  </section>
</template>
