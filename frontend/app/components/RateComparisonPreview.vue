<script setup lang="ts">
import { Clock3, MapPin, ShieldCheck } from 'lucide-vue-next'

const providers = [
  { name: 'Provider A', service: 'Express', price: '₦3,500', eta: '1–2 days', pickup: 'Pickup available' },
  { name: 'Provider B', service: 'Standard', price: '₦2,800', eta: '2–3 days', pickup: 'Drop-off' },
  { name: 'Provider C', service: 'Economy', price: '₦2,400', eta: '3–4 days', pickup: 'Drop-off' },
]
</script>

<template>
  <div class="w-full max-w-[500px] rounded-[1.4rem] border border-ink-900/10 bg-white p-5 shadow-[0_35px_80px_-55px_rgba(10,18,32,0.5)] sm:p-6">
    <div class="flex items-start justify-between gap-3">
      <div>
        <p class="text-[10px] font-bold tracking-[0.17em] text-ink-900/40 uppercase">The choice is yours</p>
        <h2 class="mt-1.5 text-[18px] font-semibold tracking-[-0.03em]">Compare at a glance</h2>
      </div>
      <span class="rounded-full bg-amber-50 px-2.5 py-1 text-[9px] font-bold tracking-[0.12em] text-amber-800 uppercase">Example</span>
    </div>

    <div class="mt-4 flex items-center justify-between rounded-xl bg-slate-50 px-3.5 py-3">
      <div>
        <p class="text-[9px] font-semibold tracking-[0.12em] text-ink-900/40 uppercase">Route</p>
        <p class="mt-0.5 text-[13px] font-semibold">Aba <span class="px-1 text-brand-500">→</span> Lagos</p>
      </div>
      <div class="text-right">
        <p class="text-[9px] font-semibold tracking-[0.12em] text-ink-900/40 uppercase">Package</p>
        <p class="mt-0.5 text-[12px] font-medium text-ink-900/70">2 kg · Parcel</p>
      </div>
    </div>

    <ul class="mt-3 space-y-2">
      <li
        v-for="(provider, index) in providers"
        :key="provider.name"
        class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 rounded-xl border p-3.5"
        :class="index === 0 ? 'border-brand-500/30 bg-brand-50/35' : 'border-slate-200'"
      >
        <div class="flex min-w-0 items-center gap-3">
          <div class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-white text-[12px] font-bold text-brand-700 ring-1 ring-ink-900/8">
            {{ provider.name.slice(-1) }}
          </div>
          <div class="min-w-0">
            <p class="truncate text-[13px] font-semibold">{{ provider.name }} <span class="ml-1 rounded bg-white px-1.5 py-0.5 text-[9px] font-medium text-ink-900/50">{{ provider.service }}</span></p>
            <div class="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-[10.5px] text-ink-900/50">
              <span class="inline-flex items-center gap-1"><Clock3 :size="11" /> {{ provider.eta }}</span>
              <span class="inline-flex items-center gap-1"><MapPin :size="11" /> {{ provider.pickup }}</span>
            </div>
          </div>
        </div>
        <p class="whitespace-nowrap text-[14px] font-bold tracking-[-0.03em]">{{ provider.price }}</p>
      </li>
    </ul>

    <p class="mt-4 flex items-start gap-2 border-t border-slate-100 pt-3 text-[10.5px] leading-relaxed text-ink-900/45">
      <ShieldCheck :size="13" class="mt-0.5 shrink-0" />
      Illustrative comparison only. These are sample names and prices, not live provider quotes or bookable offers.
    </p>
  </div>
</template>
