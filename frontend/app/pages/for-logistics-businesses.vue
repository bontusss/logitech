<script setup lang="ts">
import { ArrowRight, BadgeCheck, Building2, ClipboardList, MessageCircle, Tags, Truck } from 'lucide-vue-next'

const { demoUrl, whatsappHref, whatsappUrl } = useSiteConfig()

useSeoMeta({
  title: 'Logistics business operations platform',
  description:
    'Manage branches, teams, shipments and pricing in one place with BOL. Book a demo or talk with our team.',
})

const capabilities = [
  {
    icon: Building2,
    title: 'One view across branches',
    body: 'Keep branch information, terminals and staff access organised as your network grows.',
  },
  {
    icon: ClipboardList,
    title: 'Less scattered operations',
    body: 'Manage shipments, manifests and interbranch transfers in a shared workspace.',
  },
  {
    icon: Tags,
    title: 'Pricing you control',
    body: 'Set up service routes and pricing rules that reflect how your company actually operates.',
  },
  {
    icon: Truck,
    title: 'More ways to get discovered',
    body: 'Build your company profile and prepare to receive bookings from customers looking for a provider.',
  },
]
</script>

<template>
  <div>
    <section class="relative overflow-hidden bg-[#f3f5ff]">
      <div aria-hidden="true" class="pointer-events-none absolute inset-0 grid-paper opacity-50 [mask-image:linear-gradient(to_bottom,black,transparent_90%)]" />
      <div aria-hidden="true" class="absolute -top-56 right-[-8rem] h-[560px] w-[560px] rounded-full bg-brand-500/15 blur-[120px]" />
      <div class="relative mx-auto max-w-[1200px] px-5 py-16 sm:px-8 sm:py-24">
        <p class="inline-flex items-center gap-2 rounded-full border border-brand-500/15 bg-white/80 px-3 py-1.5 text-[10px] font-bold tracking-[0.16em] text-brand-700 uppercase">
          <Building2 :size="13" />
          For logistics businesses
        </p>
        <div class="mt-6 grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(260px,0.65fr)] lg:items-end">
          <div class="max-w-[740px]">
            <h1 class="text-[clamp(2.6rem,6vw,4.7rem)] leading-[0.96] font-semibold tracking-[-0.06em] text-balance">
              Run smoother.
              <span class="block text-brand-600">Reach further.</span>
            </h1>
            <p class="mt-5 max-w-[650px] text-[16px] leading-[1.7] text-ink-900/60">
              Bring branches, people, shipments and pricing into one clear workspace—and make it easier for customers to find your services.
            </p>
          </div>
          <div class="flex flex-col gap-2.5 sm:flex-row lg:flex-col lg:items-stretch">
            <AppButton :to="demoUrl" size="lg">
              Book a demo
              <ArrowRight :size="16" />
            </AppButton>
            <AppButton v-if="whatsappUrl" :to="whatsappHref" variant="secondary" size="lg">
              <MessageCircle :size="16" />
              Chat on WhatsApp
            </AppButton>
            <p v-if="!whatsappUrl" class="text-center text-[10.5px] text-ink-900/40">
              WhatsApp contact will be available soon. Email
              <a class="font-semibold text-brand-600" href="mailto:hello@basedonlogistics.com">hello@basedonlogistics.com</a>.
            </p>
          </div>
        </div>
        <div class="mt-10 flex flex-wrap gap-2">
          <span
            v-for="item in ['Branches & terminals', 'Staff & permissions', 'Shipments & manifests', 'Rates & routes']"
            :key="item"
            class="rounded-full border border-ink-900/10 bg-white/70 px-3 py-1.5 text-[11px] font-medium text-ink-900/60"
          >
            {{ item }}
          </span>
        </div>
      </div>
    </section>

    <BusinessPlatformSection />

    <section class="border-t border-ink-900/10 bg-canvas">
      <div class="mx-auto max-w-[1200px] px-5 py-16 sm:px-8 sm:py-20">
        <div class="mx-auto max-w-[680px] text-center">
          <p class="text-[10px] font-bold tracking-[0.18em] text-brand-600 uppercase">Built around your operation</p>
          <h2 class="mt-3 text-[clamp(1.8rem,4vw,2.6rem)] leading-[1.05] font-semibold tracking-[-0.045em]">
            Your routes are different. Your tools should understand that.
          </h2>
          <p class="mt-4 text-[14px] leading-[1.7] text-ink-900/55">
            BOL is being built for the real variety of intercity logistics: different branches, service types and pricing models, without forcing every operator into the same setup.
          </p>
        </div>
        <div class="mt-9 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <article v-for="item in capabilities" :key="item.title" class="rounded-2xl border border-ink-900/10 bg-white p-5">
            <div class="flex size-9 items-center justify-center rounded-xl bg-brand-50 text-brand-600">
              <component :is="item.icon" :size="17" />
            </div>
            <h3 class="mt-4 text-[13px] font-semibold">{{ item.title }}</h3>
            <p class="mt-2 text-[12px] leading-[1.65] text-ink-900/55">{{ item.body }}</p>
          </article>
        </div>
      </div>
    </section>

    <section class="bg-white px-5 py-14 sm:px-8">
      <div class="mx-auto flex max-w-[960px] flex-col gap-5 rounded-2xl border border-ink-900/10 bg-canvas p-6 sm:flex-row sm:items-center sm:justify-between sm:p-8">
        <div class="flex items-start gap-3">
          <div class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-brand-50 text-brand-600">
            <BadgeCheck :size="18" />
          </div>
          <div>
            <h2 class="text-[16px] font-semibold tracking-[-0.02em]">See how it could work for your company.</h2>
            <p class="mt-1 text-[12.5px] text-ink-900/55">Talk through your branches, routes and current processes with our team.</p>
          </div>
        </div>
        <AppButton :to="demoUrl" size="lg" class="shrink-0">
          Book a demo
          <ArrowRight :size="15" />
        </AppButton>
      </div>
    </section>
  </div>
</template>
