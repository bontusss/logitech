<script setup lang="ts">
import { ArrowRight, MapPin, ShieldCheck, Sparkles, Truck } from 'lucide-vue-next'
</script>

<template>
  <section id="top" class="relative overflow-hidden">
    <div aria-hidden="true" class="pointer-events-none absolute inset-0">
      <div
        class="absolute inset-0 grid-paper [mask-image:linear-gradient(to_bottom,black,transparent_88%)]"
      />
      <div
        class="absolute -top-40 left-1/2 h-[540px] w-[820px] -translate-x-1/2 rounded-full bg-brand-500/10 blur-[120px]"
      />
    </div>

    <div class="relative mx-auto max-w-[1200px] px-5 pt-16 pb-16 sm:px-8 sm:pt-24 sm:pb-20">
      <div class="mx-auto max-w-3xl text-center">
        <p
          class="reveal inline-flex items-center gap-2 rounded-full border border-ink-900/10 bg-white px-3 py-1.5 text-[11.5px] font-medium text-ink-900/65"
        >
          <Sparkles :size="13" class="text-brand-600" />
          A marketplace and operations platform for African logistics
        </p>

        <h1
          class="reveal reveal-1 mt-6 text-[clamp(2.4rem,6.2vw,4.35rem)] leading-[0.98] font-semibold tracking-[-0.045em] text-balance"
        >
          One platform. Every shipment.
          <span
            class="block bg-gradient-to-r from-brand-700 via-brand-500 to-[#7088ff] bg-clip-text text-transparent"
          >
            More possibilities.
          </span>
        </h1>

        <p
          class="reveal reveal-2 mx-auto mt-6 max-w-2xl text-[16.5px] leading-[1.65] text-pretty text-ink-900/60"
        >
          Discover logistics companies, compare shipping rates and move packages across cities.
          Or grow your logistics business with tools that bring branches, staff, shipments and
          customers into one place.
        </p>

        <div class="reveal reveal-3 mt-8 flex flex-col items-center justify-center gap-2.5 sm:flex-row">
          <AppButton to="#rate-search" size="lg">
            Compare Shipping Rates
            <ArrowRight :size="16" />
          </AppButton>
          <AppButton to="#business" variant="secondary" size="lg">
            For Logistics Businesses
            <ArrowRight :size="16" />
          </AppButton>
        </div>

        <div
          class="reveal reveal-3 mt-6 flex flex-wrap items-center justify-center gap-x-5 gap-y-2 text-[12px] text-ink-900/50"
        >
          <span class="inline-flex items-center gap-1.5"><Truck :size="13" /> Marketplace + operations</span>
          <span class="inline-flex items-center gap-1.5"><ShieldCheck :size="13" /> Verified provider profiles</span>
          <span class="inline-flex items-center gap-1.5"><MapPin :size="13" /> Built for Nigerian intercity routes</span>
        </div>
      </div>

      <div
        class="reveal reveal-2 mx-auto mt-12 grid max-w-[1050px] gap-4 lg:grid-cols-[minmax(0,420px)_minmax(0,1fr)] lg:items-start"
      >
        <ShippingRateSearch />
        <RateComparisonPreview />
      </div>
    </div>
  </section>
</template>
