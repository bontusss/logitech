<script setup lang="ts">
import { Clock3, MapPin, ShieldCheck } from 'lucide-vue-next'

const providers = [
  {
    name: 'Provider A',
    service: 'Express',
    price: '₦3,500',
    eta: '1–2 days',
    pickup: 'Pickup available',
    card: 'from-violet-50 to-white dark:from-violet-400/10 dark:to-transparent',
    badge: 'bg-violet-100 text-violet-800 dark:bg-violet-400/15 dark:text-violet-200',
  },
  {
    name: 'Provider B',
    service: 'Standard',
    price: '₦2,800',
    eta: '2–3 days',
    pickup: 'Drop-off',
    card: 'from-sky-50 to-white dark:from-sky-400/10 dark:to-transparent',
    badge: 'bg-sky-100 text-sky-800 dark:bg-sky-400/15 dark:text-sky-200',
  },
  {
    name: 'Provider C',
    service: 'Economy',
    price: '₦2,400',
    eta: '3–4 days',
    pickup: 'Drop-off',
    card: 'from-emerald-50 to-white dark:from-emerald-400/10 dark:to-transparent',
    badge: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-400/15 dark:text-emerald-200',
  },
]
</script>

<template>
  <div class="site-panel w-full max-w-[500px] rounded-[1.4rem] border p-5 sm:p-6">
    <div class="flex items-start justify-between gap-3">
      <div>
        <p class="text-[10px] font-bold tracking-[0.17em] text-ink-900/60 uppercase">The choice is yours</p>
        <h2 class="mt-1.5 text-[18px] font-semibold tracking-[-0.03em]">Compare at a glance</h2>
      </div>
      <span class="rounded-full border border-amber-200/80 bg-amber-50 px-2.5 py-1 text-[9px] font-bold tracking-[0.12em] text-amber-800 uppercase dark:border-amber-400/20 dark:bg-amber-400/10 dark:text-amber-200">Example</span>
    </div>

    <div class="mt-4 flex items-center justify-between rounded-xl border border-ink-900/8 bg-canvas px-3.5 py-3 dark:border-white/10 dark:bg-white/5">
      <div>
        <p class="text-[9px] font-semibold tracking-[0.12em] text-ink-900/60 uppercase">Route</p>
        <p class="mt-0.5 text-[13px] font-semibold">Aba <span class="bg-gradient-to-r from-brand-500 to-violet-500 bg-clip-text px-1 text-transparent">→</span> Lagos</p>
      </div>
      <div class="text-right">
        <p class="text-[9px] font-semibold tracking-[0.12em] text-ink-900/60 uppercase">Package</p>
        <p class="mt-0.5 text-[12px] font-medium text-ink-900/70">2 kg · Parcel</p>
      </div>
    </div>

    <ul class="mt-3 space-y-2">
      <li
        v-for="(provider, index) in providers"
        :key="provider.name"
        class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 rounded-xl border bg-gradient-to-br p-3.5"
        :class="[provider.card, index === 0 ? 'border-brand-500/30' : 'border-ink-900/8 dark:border-white/10']"
      >
        <div class="flex min-w-0 items-center gap-3">
          <div class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-white text-[12px] font-bold text-brand-700 shadow-sm ring-1 ring-ink-900/8 dark:bg-white/10 dark:text-indigo-200 dark:ring-white/10">
            {{ provider.name.slice(-1) }}
          </div>
          <div class="min-w-0">
            <p class="truncate text-[13px] font-semibold">
              {{ provider.name }}
              <span class="ml-1 rounded-full px-1.5 py-0.5 text-[10px] font-semibold" :class="provider.badge">{{ provider.service }}</span>
            </p>
            <div class="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-[10.5px] font-medium text-ink-900/65">
              <span class="inline-flex items-center gap-1"><Clock3 :size="11" /> {{ provider.eta }}</span>
              <span class="inline-flex items-center gap-1"><MapPin :size="11" /> {{ provider.pickup }}</span>
            </div>
          </div>
        </div>
        <p class="whitespace-nowrap rounded-lg bg-white/80 px-2 py-1 text-[14px] font-bold tracking-[-0.03em] shadow-sm dark:bg-white/10">{{ provider.price }}</p>
      </li>
    </ul>

    <p class="mt-4 flex items-start gap-2 border-t border-ink-900/8 pt-3 text-[10.5px] leading-relaxed text-ink-900/60">
      <ShieldCheck :size="13" class="mt-0.5 shrink-0 text-emerald-700 dark:text-emerald-300" />
      Illustrative comparison only. These are sample names and prices, not live provider quotes or bookable offers.
    </p>
  </div>
</template>
