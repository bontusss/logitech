<script setup lang="ts">
import { ArrowRight, MessageCircle, Package, Store } from 'lucide-vue-next'

const { demoUrl, whatsappHref } = useSiteConfig()
</script>

<template>
  <section class="bg-white px-5 pb-20 sm:px-8 sm:pb-24">
    <div
      class="relative mx-auto max-w-[1200px] overflow-hidden rounded-2xl bg-ink-950 px-6 py-14 text-white sm:px-10 sm:py-16"
    >
      <div aria-hidden="true" class="absolute inset-0 grid-paper-light opacity-60" />
      <div
        aria-hidden="true"
        class="absolute -top-24 right-0 h-[320px] w-[420px] rounded-full bg-brand-500/20 blur-[100px]"
      />

      <div class="relative">
        <div class="mx-auto max-w-2xl text-center">
          <h2
            class="text-[clamp(1.8rem,4vw,2.7rem)] leading-[1.05] font-semibold tracking-[-0.04em] text-balance"
          >
            Ready to move logistics forward?
          </h2>
          <p class="mx-auto mt-4 max-w-xl text-[15px] leading-[1.65] text-white/50 text-pretty">
            Whether you are shipping a package or building a logistics business, there is a better
            way to move.
          </p>
        </div>

        <div class="mx-auto mt-10 grid max-w-3xl gap-3 sm:grid-cols-2">
          <div class="flex flex-col rounded-xl border border-white/10 bg-white/[0.04] p-5">
            <Package :size="18" class="text-brand-200" />
            <p class="mt-4 text-[10px] font-semibold tracking-[0.18em] text-white/35 uppercase">
              For customers
            </p>
            <h3 class="mt-2 text-[15.5px] leading-snug font-semibold">
              Find a logistics provider and compare shipping rates.
            </h3>
            <div class="mt-5 pt-1">
              <AppButton to="#rate-search" variant="light" size="md">
                Compare Shipping Rates
                <ArrowRight :size="15" />
              </AppButton>
            </div>
          </div>

          <div class="flex flex-col rounded-xl border border-white/10 bg-white/[0.04] p-5">
            <Store :size="18" class="text-brand-200" />
            <p class="mt-4 text-[10px] font-semibold tracking-[0.18em] text-white/35 uppercase">
              For logistics businesses
            </p>
            <h3 class="mt-2 text-[15.5px] leading-snug font-semibold">
              Digitise your operations and reach more customers.
            </h3>
            <div class="mt-4 flex flex-wrap gap-2">
              <AppButton :to="demoUrl" variant="light" size="md">
                Book a Demo
                <ArrowRight :size="14" />
              </AppButton>
              <AppButton :to="whatsappHref" variant="outline-light" size="md">
                <MessageCircle :size="14" />
                Chat on WhatsApp
              </AppButton>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
