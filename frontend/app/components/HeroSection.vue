<script setup lang="ts">
import { ArrowDown, ShieldCheck, Truck } from 'lucide-vue-next'
</script>

<template>
  <section id="top" class="relative overflow-hidden bg-[#f3f5ff]">
    <div aria-hidden="true" class="pointer-events-none absolute inset-0">
      <div class="absolute inset-0 grid-paper opacity-60 [mask-image:linear-gradient(to_bottom,black,transparent_90%)]" />
      <div class="absolute -top-56 right-[-8rem] h-[620px] w-[620px] rounded-full bg-brand-500/15 blur-[120px]" />
      <div class="absolute bottom-[-18rem] left-[-10rem] h-[500px] w-[500px] rounded-full bg-[#ffc76b]/20 blur-[100px]" />
    </div>

    <div class="relative mx-auto max-w-[1200px] px-5 pt-14 pb-16 sm:px-8 sm:pt-20 lg:pt-24 lg:pb-24">
      <div class="mx-auto max-w-[780px] text-center">
        <p class="reveal inline-flex items-center gap-2 rounded-full border border-brand-500/15 bg-white/80 px-3 py-1.5 text-[11px] font-semibold tracking-[0.12em] text-brand-700 uppercase">
          <span class="size-1.5 rounded-full bg-brand-500" />
          Nigeria, one delivery at a time
        </p>

        <h1 class="reveal reveal-1 mt-6 text-[clamp(2.8rem,6.2vw,5.2rem)] leading-[0.94] font-semibold tracking-[-0.06em] text-balance">
          Send it right.
          <span class="block text-brand-600">Know your options.</span>
        </h1>

        <p class="reveal reveal-2 mx-auto mt-6 max-w-[620px] text-[16px] leading-[1.7] text-ink-900/65 sm:text-[17px]">
          Compare intercity delivery options by price and speed, then choose the service that works for you. No more calling around to figure out where to start.
        </p>

        <div class="reveal reveal-3 mt-6 flex flex-wrap items-center justify-center gap-x-5 gap-y-2 text-[12px] font-medium text-ink-900/60">
          <span class="inline-flex items-center gap-1.5"><ShieldCheck :size="15" class="text-brand-600" /> Compare with clarity</span>
          <span class="inline-flex items-center gap-1.5"><Truck :size="15" class="text-brand-600" /> Built for intercity shipping</span>
        </div>

      </div>

      <div class="reveal reveal-2 mx-auto mt-10 grid max-w-[1080px] gap-4 lg:grid-cols-2 lg:items-start">
        <ShippingRateSearch />
        <div>
          <RateComparisonPreview />
          <a href="#how-it-works" class="mt-4 inline-flex items-center gap-2 text-[12.5px] font-semibold text-ink-900/55 transition hover:text-brand-700">
            How shipping with BOL works
            <ArrowDown :size="14" />
          </a>
        </div>
      </div>
    </div>
  </section>
</template>
