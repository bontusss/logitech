<script setup lang="ts">
import { ShieldCheck, Truck, Users } from 'lucide-vue-next'

const highlights = [
  { icon: Truck, text: 'Built for Nigerian intercity routes' },
  { icon: Users, text: 'Separate company workspaces and permissions' },
  { icon: ShieldCheck, text: 'Marketplace bookings and operations together' },
]
</script>

<template>
  <div class="auth-page min-h-dvh bg-canvas text-ink-900 lg:grid lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
    <aside class="auth-aside relative hidden overflow-hidden p-10 lg:flex lg:flex-col lg:justify-between">
      <div aria-hidden="true" class="pointer-events-none absolute -top-24 right-[-4rem] size-80 rounded-full bg-accent-500/20 blur-3xl" />
      <div aria-hidden="true" class="pointer-events-none absolute -bottom-28 -left-16 size-80 rounded-full bg-[#2b5c96]/25 blur-3xl" />

      <div class="relative">
        <NuxtLink to="/" aria-label="based on logistics — home">
          <AppLogo :size="38" tone="light" />
        </NuxtLink>
      </div>

      <div class="relative max-w-md">
        <SectionEyebrow tone="light" label="Logistics, connected" />
        <h1 class="mt-4 text-[clamp(1.9rem,3vw,2.5rem)] leading-[1.08] font-semibold tracking-[-0.04em] text-white">
          Run your operations.
          <span class="block text-accent-300">Reach more customers.</span>
        </h1>
        <p class="mt-4 text-[14px] font-medium leading-[1.7] text-white/65">
          One workspace for branches, staff, shipments, pricing and marketplace bookings — and one
          marketplace for the customers who need them moved.
        </p>
        <ul class="mt-7 space-y-2.5">
          <li
            v-for="item in highlights"
            :key="item.text"
            class="flex items-center gap-2.5 rounded-xl border border-white/10 bg-white/5 px-3 py-2.5 text-[12px] font-semibold text-white/80 backdrop-blur-sm"
          >
            <IconTile tone="on-dark" size="sm" rounded="lg">
              <component :is="item.icon" :size="16" :stroke-width="2.1" />
            </IconTile>
            {{ item.text }}
          </li>
        </ul>
      </div>

      <div class="relative flex items-center justify-between text-[11px] font-medium text-white/50">
        <span>© {{ new Date().getFullYear() }} based on logistics</span>
        <NuxtLink to="/" class="rounded-full px-2.5 py-1 transition hover:bg-white/10 hover:text-white">
          Back to site
        </NuxtLink>
      </div>
    </aside>

    <div class="flex min-h-dvh flex-col">
      <header class="flex items-center justify-between px-5 py-5 lg:hidden">
        <NuxtLink to="/" aria-label="based on logistics — home">
          <AppLogo :size="32" />
        </NuxtLink>
        <div class="flex items-center gap-2">
          <ThemeToggle />
          <NuxtLink to="/" class="rounded-full px-2.5 py-1 text-[12px] font-medium text-ink-900/70 hover:text-ink-900">
            Back to site
          </NuxtLink>
        </div>
      </header>

      <main class="auth-main flex flex-1 items-center justify-center px-5 py-8 sm:px-8">
        <div class="w-full max-w-[460px]">
          <div class="mb-3 hidden justify-end lg:flex">
            <ThemeToggle />
          </div>
          <div class="auth-panel overflow-hidden rounded-3xl border">
            <div class="h-1.5 w-full bg-accent-500" />
            <div class="p-6 sm:p-8">
              <slot />
            </div>
          </div>
        </div>
      </main>

      <footer class="px-5 pb-6 text-center text-[11px] text-ink-900/35 sm:px-8">
        Secure workspace access · based on logistics (BOL)
      </footer>
    </div>
  </div>
</template>
