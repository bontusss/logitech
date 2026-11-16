<script setup lang="ts">
import { ArrowUpRight, Check, Clock, MapPin, ShieldCheck } from 'lucide-vue-next'

const providers = [
  {
    id: 'A',
    service: 'Express',
    price: '₦3,500',
    eta: '2 days',
    handoff: 'Door pickup',
    tag: 'Best balance',
    featured: true,
  },
  { id: 'B', service: 'Priority', price: '₦4,000', eta: '1 day', handoff: 'Drop-off', tag: '', featured: false },
  { id: 'C', service: 'Economy', price: '₦2,800', eta: '3 days', handoff: 'Drop-off', tag: '', featured: false },
]
</script>

<template>
  <div
    class="rounded-2xl border border-ink-900/10 bg-white p-4 shadow-[0_30px_70px_-45px_rgba(10,18,32,0.5)] sm:p-5"
  >
    <div class="flex items-start justify-between gap-3">
      <div>
        <p class="text-[10px] font-semibold tracking-[0.18em] text-ink-900/40 uppercase">
          Rate comparison
        </p>
        <p class="mt-1.5 flex items-center gap-1.5 text-[17px] font-semibold tracking-[-0.02em]">
          Aba <span class="text-ink-900/25">→</span> Lagos
        </p>
      </div>
      <span
        class="rounded-md border border-amber-200 bg-amber-50 px-2 py-1 text-[10px] font-semibold tracking-[0.1em] text-amber-700 uppercase"
      >
        Sample data
      </span>
    </div>

    <div class="mt-3 flex flex-wrap gap-1.5 text-[11px] font-medium text-ink-900/55">
      <span class="rounded border border-slate-200 bg-slate-50 px-1.5 py-0.5">5 kg</span>
      <span class="rounded border border-slate-200 bg-slate-50 px-1.5 py-0.5">Parcels</span>
      <span class="rounded border border-slate-200 bg-slate-50 px-1.5 py-0.5">Pickup today</span>
    </div>

    <ul class="mt-4 space-y-2">
      <li
        v-for="provider in providers"
        :key="provider.id"
        class="rounded-xl border p-3 transition"
        :class="
          provider.featured
            ? 'border-brand-500/40 bg-brand-50/50 ring-1 ring-brand-500/10'
            : 'border-slate-200 hover:border-slate-300'
        "
      >
        <div class="flex items-center gap-3">
          <div
            class="flex size-9 shrink-0 items-center justify-center rounded-lg border border-slate-200 bg-white text-[13px] font-semibold text-ink-900/60"
          >
            {{ provider.id }}
          </div>

          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-1.5">
              <p class="text-[13.5px] font-semibold">Provider {{ provider.id }}</p>
              <span class="rounded bg-slate-100 px-1.5 py-0.5 text-[10px] font-medium text-ink-900/60">
                {{ provider.service }}
              </span>
              <span
                v-if="provider.tag"
                class="rounded bg-brand-100 px-1.5 py-0.5 text-[10px] font-semibold text-brand-700"
              >
                {{ provider.tag }}
              </span>
            </div>
            <div class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-[11.5px] text-ink-900/50">
              <span class="inline-flex items-center gap-1"><Clock :size="11" /> {{ provider.eta }}</span>
              <span class="inline-flex items-center gap-1"><MapPin :size="11" /> {{ provider.handoff }}</span>
            </div>
          </div>

          <div class="text-right">
            <p class="text-[15px] font-semibold tracking-[-0.02em]">{{ provider.price }}</p>
            <p class="text-[10px] text-ink-900/40">est. total</p>
          </div>
        </div>

        <div
          v-if="provider.featured"
          class="mt-3 flex items-center justify-between gap-2 border-t border-brand-500/15 pt-2.5"
        >
          <span class="inline-flex items-center gap-1 text-[11px] font-medium text-brand-700">
            <Check :size="12" /> Ships within 24 hrs of pickup
          </span>
          <span class="inline-flex items-center gap-1 text-[11.5px] font-semibold text-brand-700">
            View details <ArrowUpRight :size="12" />
          </span>
        </div>
      </li>
    </ul>

    <div class="mt-4 flex items-start gap-1.5 border-t border-slate-100 pt-3 text-[11px] leading-relaxed text-ink-900/45">
      <ShieldCheck :size="13" class="mt-0.5 shrink-0" />
      Illustrative rates, not live quotations. Provider coverage varies by route.
    </div>
  </div>
</template>
