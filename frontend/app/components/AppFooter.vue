<script setup lang="ts">
import { ArrowRight, Linkedin, Instagram, Twitter } from 'lucide-vue-next'

const year = new Date().getFullYear()
const { demoUrl } = useSiteConfig()

const columns = [
  {
    title: 'Marketplace',
    links: [
      { label: 'Compare shipping rates', to: '#rate-search' },
      { label: 'Browse providers', to: '#marketplace' },
      { label: 'Book a shipment', to: '#rate-search' },
      { label: 'Track a shipment', to: '#marketplace' },
    ],
  },
  {
    title: 'Business platform',
    links: [
      { label: 'Company profile', to: '#business' },
      { label: 'Branches and staff', to: '#business' },
      { label: 'Shipments and manifests', to: '#features' },
      { label: 'Pricing and rates', to: '#business' },
      { label: 'Analytics', to: '#features' },
    ],
  },
  {
    title: 'Company',
    links: [
      { label: 'How it works', to: '#how-it-works' },
      { label: 'Pricing', to: '#pricing' },
      { label: 'Book a demo', to: demoUrl },
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
  <footer id="contact" class="border-t border-white/10 bg-ink-950 text-white">
    <div class="mx-auto max-w-[1200px] px-5 py-14 sm:px-8">
      <div class="grid gap-12 lg:grid-cols-[minmax(0,1.15fr)_minmax(0,2fr)]">
        <div>
          <AppLogo :size="36" tone="light" />
          <p class="mt-5 max-w-[330px] text-[14.5px] leading-[1.65] text-white/45">
            A logistics marketplace and operations platform for African intercity transport —
            connecting the businesses that move goods with the customers who need them moved.
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
            <p class="text-white/30">WhatsApp contact is being configured.</p>
          </div>

          <div class="mt-6 flex gap-2">
            <a
              v-for="social in socials"
              :key="social.label"
              :href="social.to"
              :aria-label="`${social.label} (placeholder)`"
              class="flex size-9 items-center justify-center rounded-lg border border-white/12 text-white/50 transition hover:border-white/25 hover:text-white"
            >
              <component :is="social.icon" :size="15" />
            </a>
          </div>
        </div>

        <div class="grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
          <div v-for="column in columns" :key="column.title">
            <p class="text-[10.5px] font-semibold tracking-[0.2em] text-white/35 uppercase">
              {{ column.title }}
            </p>
            <ul class="mt-4 space-y-2.5">
              <li v-for="link in column.links" :key="link.label">
                <a
                  :href="link.to"
                  class="text-[13.5px] text-white/55 transition hover:text-white"
                >
                  {{ link.label }}
                </a>
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
          <AppButton to="#top" variant="ghost" size="sm" class="text-white/50 hover:bg-white/10 hover:text-white">
            Back to top
            <ArrowRight :size="13" class="-rotate-90" />
          </AppButton>
        </div>
      </div>
    </div>
  </footer>
</template>
