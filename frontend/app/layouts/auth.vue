<script setup lang="ts">
import { ShieldCheck, Truck, Users } from 'lucide-vue-next'

const highlights = [
  { icon: Truck, text: 'Built for Nigerian intercity routes', tile: 'bg-sky-100 text-sky-800 ring-1 ring-sky-200/80 dark:bg-sky-400/15 dark:text-sky-200 dark:ring-sky-400/20' },
  { icon: Users, text: 'Separate company workspaces and permissions', tile: 'bg-violet-100 text-violet-800 ring-1 ring-violet-200/80 dark:bg-violet-400/15 dark:text-violet-200 dark:ring-violet-400/20' },
  { icon: ShieldCheck, text: 'Marketplace bookings and operations together', tile: 'bg-emerald-100 text-emerald-800 ring-1 ring-emerald-200/80 dark:bg-emerald-400/15 dark:text-emerald-200 dark:ring-emerald-400/20' },
]
</script>

<template>
  <div class="auth-page min-h-dvh bg-canvas text-ink-900 lg:grid lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
    <aside class="auth-aside relative hidden overflow-hidden p-10 lg:flex lg:flex-col lg:justify-between">
      <div aria-hidden="true" class="pointer-events-none absolute -top-24 right-[-4rem] size-72 rounded-full bg-white/40 blur-3xl dark:bg-white/5" />
      <div aria-hidden="true" class="pointer-events-none absolute -bottom-28 -left-16 size-80 rounded-full bg-cyan-200/40 blur-3xl dark:bg-cyan-400/10" />

      <div class="relative">
        <NuxtLink to="/" aria-label="based on logistics — home">
          <AppLogo :size="38" />
        </NuxtLink>
      </div>

      <div class="relative max-w-md">
        <p class="inline-flex items-center gap-1.5 rounded-full border border-brand-200/80 bg-white/85 px-2.5 py-1 text-[10px] font-bold tracking-[0.18em] text-brand-700 uppercase shadow-sm dark:border-white/10 dark:bg-white/5 dark:text-indigo-200">
          Logistics, connected
        </p>
        <h1 class="mt-4 text-[clamp(1.9rem,3vw,2.5rem)] leading-[1.08] font-semibold tracking-[-0.04em]">
          Run your operations.
          <span class="block bg-gradient-to-r from-brand-600 via-violet-600 to-cyan-600 bg-clip-text text-transparent dark:from-indigo-300 dark:via-violet-300 dark:to-cyan-300">
            Reach more customers.
          </span>
        </h1>
        <p class="mt-4 text-[14px] font-medium leading-[1.7] text-ink-900/80">
          One workspace for branches, staff, shipments, pricing and marketplace bookings — and one
          marketplace for the customers who need them moved.
        </p>
        <ul class="mt-7 space-y-2.5">
          <li
            v-for="item in highlights"
            :key="item.text"
            class="flex items-center gap-2.5 rounded-xl border border-white/80 bg-white/80 px-3 py-2.5 text-[12px] font-semibold text-ink-900/85 shadow-sm shadow-ink-900/5 backdrop-blur-sm dark:border-white/10 dark:bg-white/5 dark:text-slate-200"
          >
            <span class="flex size-8 shrink-0 items-center justify-center rounded-lg" :class="item.tile">
              <component :is="item.icon" :size="16" :stroke-width="2.1" />
            </span>
            {{ item.text }}
          </li>
        </ul>
      </div>

      <div class="relative flex items-center justify-between text-[11px] font-medium text-ink-900/60">
        <span>© {{ new Date().getFullYear() }} based on logistics</span>
        <NuxtLink to="/" class="rounded-full px-2.5 py-1 transition hover:bg-white/70 hover:text-ink-900 dark:hover:bg-white/10 dark:hover:text-white">
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
            <div class="h-1.5 w-full bg-gradient-to-r from-brand-500 via-violet-500 to-cyan-400" />
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
