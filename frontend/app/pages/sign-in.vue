<script setup lang="ts">
import { ArrowLeft, ArrowRight, Lock } from 'lucide-vue-next'

useSeoMeta({
  title: 'Sign in',
  description: 'Sign in to your BOL logistics workspace.',
})

const { demoUrl } = useSiteConfig()
</script>

<template>
  <div class="mx-auto flex max-w-[520px] flex-col items-center px-5 py-20 text-center sm:py-28">
    <div class="flex size-11 items-center justify-center rounded-xl border border-brand-100 bg-brand-50 text-brand-600">
      <Lock :size="20" />
    </div>
    <h1 class="mt-5 text-[clamp(1.8rem,4vw,2.4rem)] leading-[1.1] font-semibold tracking-[-0.04em]">
      Sign in is on its way.
    </h1>
    <p class="mt-4 max-w-sm text-[14.5px] leading-[1.65] text-ink-900/55">
      The customer and business workspaces are being built. Landing page visitors can still explore
      the marketplace preview or book a walkthrough.
    </p>

    <div class="mt-7 flex flex-col gap-2.5 sm:flex-row">
      <AppButton :to="demoUrl">
        Book a Demo
        <ArrowRight :size="15" />
      </AppButton>
      <AppButton to="/" variant="secondary">
        <ArrowLeft :size="15" />
        Back to home
      </AppButton>
    </div>
  </div>
</template>
