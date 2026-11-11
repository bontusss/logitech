<script setup lang="ts">
const year = new Date().getFullYear()

const columns = [
  {
    title: 'Platform',
    links: [
      { label: 'How it works', to: '#how-it-works' },
      { label: 'Partial dispatch', to: '#partial-dispatch' },
      { label: 'Modules', to: '#platform' },
      { label: 'White-label', to: '#multi-tenant' },
    ],
  },
  {
    title: 'Solutions',
    links: [
      { label: 'Courier operators', to: '#solutions' },
      { label: 'Haulage & trucking', to: '#solutions' },
      { label: 'Freight forwarders', to: '#solutions' },
      { label: 'Branch networks', to: '#solutions' },
    ],
  },
  {
    title: 'Get started',
    links: [
      { label: 'Pricing', to: '#pricing' },
      { label: 'Questions', to: '#faq' },
      { label: 'Sign in', to: '/sign-in' },
    ],
  },
]
</script>

<template>
  <footer class="border-t border-white/10 bg-[#070f1d] text-white">
    <div class="mx-auto max-w-[1200px] px-5 py-14 sm:px-8">
      <div class="grid gap-10 lg:grid-cols-[minmax(0,1.4fr)_minmax(0,2fr)]">
        <div>
          <LandingBrandLogo :size="36" tone="light" />
          <p class="mt-5 max-w-[320px] text-[14.5px] leading-[1.6] text-white/45">
            The operating system for courier, haulage and consolidation businesses — one record for every
            waybill, bag, route and naira.
          </p>
          <div class="mt-6">
            <LandingActionLink to="/sign-in" variant="light">
              Get started free
              <UiIcon name="arrowRight" :size="15" />
            </LandingActionLink>
          </div>
        </div>

        <div class="grid gap-8 sm:grid-cols-3">
          <div v-for="column in columns" :key="column.title">
            <p class="text-[11px] font-semibold tracking-[0.2em] text-white/35 uppercase">
              {{ column.title }}
            </p>
            <ul class="mt-4 space-y-2.5">
              <li v-for="link in column.links" :key="link.label">
                <a
                  :href="link.to"
                  class="text-[14px] text-white/55 transition hover:text-white focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white/60"
                >
                  {{ link.label }}
                </a>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <div class="mt-12 flex flex-col gap-3 border-t border-white/10 pt-6 sm:flex-row sm:items-center sm:justify-between">
        <p class="text-[13px] text-white/35">
          © {{ year }} based on logistics (BOL). All rights reserved.
        </p>
        <p class="text-[13px] text-white/35">Every parcel. Every movement. Accounted for.</p>
      </div>
    </div>
  </footer>
</template>
