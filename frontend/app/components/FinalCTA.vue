<script setup lang="ts">
import { ArrowRight, PackageSearch } from 'lucide-vue-next'
</script>

<template>
  <section class="bg-canvas px-5 pb-20 sm:px-8 sm:pb-24">
    <div class="relative mx-auto max-w-[1200px] overflow-hidden rounded-[1.6rem] bg-gradient-to-br from-ink-950 via-[#151129] to-[#0a1b26] px-6 py-12 text-white ring-1 ring-white/10 sm:px-10 sm:py-16">
      <div aria-hidden="true" class="absolute inset-0 grid-paper-light opacity-50" />
      <div aria-hidden="true" class="absolute -top-32 right-[-4rem] h-[380px] w-[380px] rounded-full bg-brand-500/30 blur-[100px]" />
      <div aria-hidden="true" class="absolute -bottom-32 left-[-4rem] h-[340px] w-[340px] rounded-full bg-violet-500/25 blur-[100px]" />
      <div aria-hidden="true" class="absolute right-24 bottom-[-10rem] h-[280px] w-[280px] rounded-full bg-cyan-400/20 blur-[90px]" />
      <div class="relative mx-auto max-w-[680px] text-center">
        <div class="mx-auto flex size-12 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-500 to-violet-500 text-white shadow-lg shadow-brand-600/30">
          <PackageSearch :size="21" />
        </div>
        <h2 class="mt-5 text-[clamp(1.8rem,4vw,2.7rem)] leading-[1.05] font-semibold tracking-[-0.045em] text-balance">
          Your next delivery deserves a better way.
        </h2>
        <p class="mx-auto mt-4 max-w-[540px] text-[14px] leading-[1.7] text-white/55">
          BOL is building a simpler way to compare intercity shipping options. Be among the first to know when live rates and booking are ready.
        </p>
        <div class="mt-7 flex flex-col justify-center gap-2.5 sm:flex-row">
          <AppButton to="/register" variant="light" size="lg">
            Get ready to ship
            <ArrowRight :size="15" />
          </AppButton>
          <AppButton to="/for-logistics-businesses" variant="outline-light" size="lg">
            For logistics businesses
          </AppButton>
        </div>
        <p class="mt-3 text-[10.5px] text-white/35">
          Marketplace booking is not live yet. Creating an account does not place an order.
        </p>
      </div>
    </div>
  </section>
</template>
