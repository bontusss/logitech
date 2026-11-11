<script setup lang="ts">
/**
 * The partial-dispatch centrepiece: one bag leaving with most of its parcels
 * while the rest stay on the counter, each with its own recorded reason.
 */
const parcels = [
  { id: 'PKG-2026-00124', item: 'Electronics · 2 items', state: 'In bag' },
  { id: 'PKG-2026-00126', item: 'Documents · 1 item', state: 'In bag' },
  { id: 'PKG-2026-00131', item: 'Spare parts · 3 items', state: 'In bag' },
  { id: 'PKG-2026-00119', item: 'Textiles · 1 item', state: 'Held' },
  { id: 'PKG-2026-00133', item: 'Cosmetics · 4 items', state: 'Held' },
  { id: 'PKG-2026-00135', item: 'Books · 2 items', state: 'In bag' },
] as const

const CHIP: Record<string, string> = {
  'In bag': 'bg-[#2674da]/10 text-[#1b5cab]',
  Held: 'bg-[#b7791f]/12 text-[#96631a]',
}
</script>

<template>
  <div class="grid gap-3 lg:grid-cols-[minmax(0,1.25fr)_minmax(0,0.95fr)]">
    <!-- Manifest -->
    <div class="rounded-2xl border border-[#0b1424]/10 bg-white p-4 shadow-[0_24px_60px_-40px_rgba(11,20,36,0.45)] sm:p-5">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div>
          <p class="text-[11px] font-medium tracking-[0.12em] text-[#0b1424]/40 uppercase">Dispatch manifest</p>
          <p class="mt-1 text-[15px] font-semibold tracking-tight text-[#0b1424]">BAG-2026-00018</p>
        </div>
        <span class="inline-flex items-center gap-1.5 rounded-[5px] bg-[#b7791f]/12 px-2 py-1 text-[10px] font-semibold tracking-[0.08em] text-[#96631a] uppercase">
          Partially dispatched
        </span>
      </div>

      <!-- Route -->
      <div class="mt-5 flex items-center gap-3">
        <div class="min-w-0">
          <p class="text-[9px] font-medium tracking-[0.12em] text-[#0b1424]/35 uppercase">From</p>
          <p class="truncate text-[12px] font-medium text-[#0b1424]">Lagos — Ikeja</p>
        </div>
        <div class="flex flex-1 items-center gap-1.5">
          <span class="size-1.5 shrink-0 rounded-full bg-[#0b1424]" />
          <span class="h-px flex-1 bg-[linear-gradient(to_right,rgba(11,20,36,0.35),rgba(11,20,36,0.12))]" />
          <span class="text-[9px] whitespace-nowrap text-[#0b1424]/35">6h 20m</span>
          <span class="h-px flex-1 bg-[linear-gradient(to_right,rgba(11,20,36,0.12),rgba(11,20,36,0.35))]" />
          <span class="size-1.5 shrink-0 rounded-full border border-[#0b1424]/35 bg-white" />
        </div>
        <div class="min-w-0 text-right">
          <p class="text-[9px] font-medium tracking-[0.12em] text-[#0b1424]/35 uppercase">To</p>
          <p class="truncate text-[12px] font-medium text-[#0b1424]">Port Harcourt</p>
        </div>
      </div>

      <ul class="mt-5 divide-y divide-[#0b1424]/[0.07] border-t border-[#0b1424]/[0.07]">
        <li v-for="parcel in parcels" :key="parcel.id" class="flex items-center justify-between gap-3 py-2.5">
          <div class="min-w-0">
            <p class="truncate text-[11.5px] font-medium text-[#0b1424] tabular-nums">{{ parcel.id }}</p>
            <p class="truncate text-[10.5px] text-[#0b1424]/40">{{ parcel.item }}</p>
          </div>
          <span
            class="shrink-0 rounded-[3px] px-1.5 py-0.5 text-[9px] font-semibold tracking-[0.05em] uppercase"
            :class="CHIP[parcel.state]"
          >
            {{ parcel.state }}
          </span>
        </li>
      </ul>

      <div class="mt-4 rounded-lg bg-[#f7f8fa] px-3 py-2.5">
        <div class="flex items-center justify-between text-[10.5px] text-[#0b1424]/55">
          <span>6 of 8 parcels dispatched</span>
          <span class="tabular-nums">75%</span>
        </div>
        <div class="mt-1.5 h-1 overflow-hidden rounded-full bg-[#0b1424]/10">
          <div class="h-full w-3/4 rounded-full bg-[#0b1424]" />
        </div>
        <p class="mt-2 text-[10px] leading-[1.5] text-[#0b1424]/40">
          Two parcels were held at Lagos — Ikeja: one awaiting payment, one still with the customer.
        </p>
      </div>
    </div>

    <!-- Explanation -->
    <div class="flex flex-col gap-3">
      <div class="rounded-2xl border border-[#0b1424]/10 bg-white p-4 sm:p-5">
        <p class="text-[11px] font-medium tracking-[0.12em] text-[#0b1424]/40 uppercase">On the record</p>
        <ul class="mt-3 space-y-3">
          <li v-for="line in [
            'A bag can leave without every parcel it was built for.',
            'Each held parcel keeps its own status and reason.',
            'The manifest shows exactly what left, when and with whom.',
          ]" :key="line" class="flex gap-2.5">
            <span class="mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-full bg-[#0b1424]/[0.06] text-[#0b1424]">
              <UiIcon name="check" :size="9" />
            </span>
            <span class="text-[12.5px] leading-[1.55] text-[#0b1424]/65">{{ line }}</span>
          </li>
        </ul>
      </div>

      <div class="rounded-2xl border border-[#0b1424]/10 bg-[#0b1424] p-4 text-white sm:p-5">
        <p class="text-[11px] font-medium tracking-[0.12em] text-white/40 uppercase">Custody, next leg</p>
        <div class="mt-3 space-y-3">
          <div class="flex items-start gap-2.5">
            <span class="mt-1 size-1.5 shrink-0 rounded-full bg-white/40" />
            <div>
              <p class="text-[11.5px] font-medium">Departed Lagos — Ikeja</p>
              <p class="text-[10px] text-white/45">11 Nov · 14:20 · TRF-2026-00042</p>
            </div>
          </div>
          <div class="flex items-start gap-2.5">
            <span class="mt-1 size-1.5 shrink-0 rounded-full bg-[#5b9dff]" />
            <div>
              <p class="text-[11.5px] font-medium">Arrives Port Harcourt</p>
              <p class="text-[10px] text-white/45">11 Nov · 20:40 · expected</p>
            </div>
          </div>
          <div class="flex items-start gap-2.5">
            <span class="mt-1 size-1.5 shrink-0 rounded-full bg-white/20" />
            <div>
              <p class="text-[11.5px] font-medium text-white/55">Customer hand-over</p>
              <p class="text-[10px] text-white/35">Signature captured on collection</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
