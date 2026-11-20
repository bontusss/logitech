<script setup lang="ts">
import { ArrowRight, Linkedin, Instagram, Twitter } from 'lucide-vue-next'

const year = new Date().getFullYear()
const { demoUrl, whatsappUrl } = useSiteConfig()

const columns = [
  {
    title: 'For shippers',
    links: [
      { label: 'Compare delivery options', to: '/#rate-search' },
      { label: 'How it works', to: '/#how-it-works' },
      { label: 'Questions & answers', to: '/#faq' },
    ],
  },
  {
    title: 'For logistics businesses',
    links: [
      { label: 'Explore the platform', to: '/for-logistics-businesses' },
      { label: 'Branches & staff', to: '/for-logistics-businesses#business' },
      { label: 'Rates & routes', to: '/for-logistics-businesses#features' },
      { label: 'Book a demo', to: demoUrl.value },
    ],
  },
  {
    title: 'Company',
    links: [
      { label: 'Logistics businesses', to: '/for-logistics-businesses' },
      { label: 'Sign in', to: '/login' },
      { label: 'Contact', to: 'mailto:hello@basedonlogistics.com' },
    ],
  },
  {
    title: 'Help',
    links: [
      { label: 'FAQ', to: '#faq' },
      { label: 'Support', to: 'mailto:hello@basedonlogistics.com' },
      { label: 'Terms of Service', to: '#' },
      { label: 'Privacy Policy', to: '#' },
    ],
  },
]

const socials = [
  { label: 'X', to: '#', icon: Twitter },
  { label: 'LinkedIn', to: '#', icon: Linkedin },
  { label: 'Instagram', to: '#', icon: Instagram },
]
</script>

<template>
  <footer id="contact" class="relative border-t border-white/10 bg-ink-950 text-white">
    <div aria-hidden="true" class="h-px w-full bg-gradient-to-r from-transparent via-brand-500/50 to-transparent" />
    <div aria-hidden="true" class="pointer-events-none absolute -top-24 right-10 size-72 rounded-full bg-brand-500/10 blur-3xl" />
    <div class="relative mx-auto max-w-[1200px] px-5 py-14 sm:px-8">
      <div class="grid gap-12 lg:grid-cols-[minmax(0,1.15fr)_minmax(0,2fr)]">
        <div>
          <AppLogo :size="36" tone="light" />
          <p class="mt-5 max-w-[330px] text-[14.5px] leading-[1.65] text-white/45">
            A clearer way to find intercity delivery options across Nigeria.
          </p>

          <div class="mt-6 space-y-1.5 text-[13.5px] text-white/45">
            <p>
              <a
                href="mailto:hello@basedonlogistics.com"
                class="transition hover:text-white"
              >
                hello@basedonlogistics.com
              </a>
            </p>
            <p>Lagos, Nigeria</p>
            <p v-if="whatsappUrl">
              <a :href="whatsappUrl" target="_blank" rel="noopener noreferrer" class="transition hover:text-white">
                Chat with us on WhatsApp
              </a>
            </p>
          </div>

          <div class="mt-6 flex gap-2">
            <a
              v-for="social in socials"
              :key="social.label"
              :href="social.to"
              :aria-label="`${social.label} (placeholder)`"
              class="flex size-9 items-center justify-center rounded-xl border border-white/12 bg-white/5 text-white/50 transition hover:-translate-y-0.5 hover:border-brand-400/40 hover:bg-brand-500/15 hover:text-white"
            >
              <component :is="social.icon" :size="15" />
            </a>
          </div>
        </div>

        <div class="grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
          <div v-for="column in columns" :key="column.title">
            <p class="flex items-center gap-2 text-[10.5px] font-semibold tracking-[0.2em] text-white/40 uppercase">
              <span class="h-px w-3 rounded-full bg-brand-400/70" />
              {{ column.title }}
            </p>
            <ul class="mt-4 space-y-2.5">
              <li v-for="link in column.links" :key="link.label">
                <NuxtLink
                  :to="link.to"
                  class="text-[13px] text-white/55 transition hover:text-white"
                >{{ link.label }}</NuxtLink>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <div
        class="mt-12 flex flex-col gap-3 border-t border-white/10 pt-6 sm:flex-row sm:items-center sm:justify-between"
      >
        <p class="text-[12.5px] text-white/35">
          © {{ year }} based on logistics (BOL). All rights reserved.
        </p>
        <div class="flex items-center gap-4 text-[12.5px] text-white/35">
          <a href="#" class="transition hover:text-white">Terms</a>
          <a href="#" class="transition hover:text-white">Privacy</a>
          <AppButton to="#top" variant="ghost" size="sm" class="rounded-full border border-white/10 text-white/55 hover:bg-white/10 hover:text-white">
            Back to top
            <ArrowRight :size="13" class="-rotate-90" />
          </AppButton>
        </div>
      </div>
    </div>
  </footer>
</template>
