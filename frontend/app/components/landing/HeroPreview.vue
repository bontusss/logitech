<script setup lang="ts">
/**
 * Illustrative product preview — a coherent dashboard, not a screenshot.
 * Everything here is demo data so the page stays honest: no invented customers,
 * no invented performance numbers, just what the screen looks like in use.
 */
const metrics = [
  { label: 'Waybills', value: '1,284', hint: '62 recorded today' },
  { label: 'In transit', value: '312', hint: 'Across 4 branches' },
  { label: 'Collected', value: '₦4.82m', hint: '92.4% of billed' },
  { label: 'Open exceptions', value: '6', hint: '2 under review', alert: true },
]

const movements = [
  {
    id: 'PKG-2026-00124',
    route: 'Lagos — Ikeja → Abuja — Wuse',
    status: 'In transit',
    tone: 'info',
    value: '₦12,500',
  },
  {
    id: 'BAG-2026-00018',
    route: 'Lagos — Ikeja → Port Harcourt',
    status: 'Sealed',
    tone: 'neutral',
    value: '42 parcels',
  },
  {
    id: 'TRF-2026-00042',
    route: 'Kano → Lagos — Apapa',
    status: 'Received',
    tone: 'ok',
    value: '₦86,000',
  },
  {
    id: 'PKG-2026-00119',
    route: 'Abuja — Wuse → Enugu',
    status: 'Exception',
    tone: 'danger',
    value: '₦7,200',
  },
  {
    id: 'PKG-2026-00117',
    route: 'Port Harcourt → Lagos — Ikeja',
    status: 'Delivered',
    tone: 'ok',
    value: '₦9,800',
  },
] as const

const custody = [
  { at: '09:12', title: 'Recorded at Lagos — Ikeja', note: 'Counter 2 · T. Adeyemi' },
  { at: '13:40', title: 'Bagged into BAG-2026-00018', note: 'Sealed with 42 parcels' },
  { at: '18:05', title: 'Departed Lagos hub', note: 'TRF-2026-00042 · 14 tonnes' },
  { at: '07:22', title: 'Arrived Abuja — Wuse', note: 'Waiting for hand-over' },
]

const TONES: Record<string, string> = {
  info: 'bg-[#2674da]/10 text-[#1b5cab]',
  ok: 'bg-[#0f9d8e]/10 text-[#0a7c6f]',
  danger: 'bg-[#dc3e42]/10 text-[#bd3236]',
  neutral: 'bg-[#0b1424]/[0.06] text-[#0b1424]/55',
}
</script>

<template>
  <div
    class="overflow-hidden rounded-[16px] border border-[#0b1424]/10 bg-white shadow-[0_40px_90px_-50px_rgba(11,20,36,0.55),0_2px_6px_rgba(11,20,36,0.05)]"
    role="img"
    aria-label="Preview of the based on logistics dashboard: metrics for waybills, parcels in transit, collections and open exceptions, a live movements table and the custody trail of parcel PKG-2026-00124."
  >
    <!-- Window chrome -->
    <div class="flex items-center gap-3 border-b border-[#0b1424]/8 bg-[#fafaf8] px-4 py-2.5">
      <div class="flex gap-1.5" aria-hidden="true">
        <span class="size-2 rounded-full bg-[#0b1424]/12" />
        <span class="size-2 rounded-full bg-[#0b1424]/12" />
        <span class="size-2 rounded-full bg-[#0b1424]/12" />
      </div>
      <div
        class="mx-auto flex h-6 w-full max-w-[300px] items-center justify-center gap-1.5 rounded-md border border-[#0b1424]/8 bg-white text-[10px] text-[#0b1424]/40"
      >
        <UiIcon name="lock" :size="9" />
        app.basedonlogistics.com/dashboard
      </div>
      <div class="w-[46px]" aria-hidden="true" />
    </div>

    <div class="flex">
      <!-- Sidebar -->
      <aside class="hidden w-[178px] shrink-0 flex-col border-r border-[#0b1424]/8 bg-[#fcfcfb] px-3 py-4 sm:flex">
        <div class="flex items-center gap-2 px-1">
          <span class="flex size-6 items-center justify-center rounded-md bg-[#0b1424] text-[10px] font-semibold text-white">
            B
          </span>
          <span class="text-[11px] font-semibold tracking-tight text-[#0b1424]">based on logistics</span>
        </div>

        <p class="mt-5 px-1 text-[9px] font-semibold tracking-[0.16em] text-[#0b1424]/30 uppercase">
          Operations
        </p>
        <ul class="mt-2 space-y-0.5">
          <li
            v-for="item in [
              { label: 'Dashboard', icon: 'home', active: true },
              { label: 'Records', icon: 'package' },
              { label: 'Containers', icon: 'truck' },
              { label: 'Branches', icon: 'building' },
              { label: 'Reports', icon: 'chart' },
              { label: 'Exceptions', icon: 'alert' },
            ]"
            :key="item.label"
          >
            <span
              class="flex items-center gap-2 rounded-md px-2 py-1.5 text-[11px]"
              :class="item.active ? 'bg-white font-medium text-[#0b1424] shadow-[0_0_0_1px_rgba(11,20,36,0.07)]' : 'text-[#0b1424]/50'"
            >
              <UiIcon :name="item.icon" :size="13" />
              {{ item.label }}
            </span>
          </li>
        </ul>

        <div class="mt-auto border-t border-[#0b1424]/8 pt-3">
          <div class="flex items-center gap-2">
            <span class="flex size-6 items-center justify-center rounded-full bg-[#0b1424]/8 text-[9px] font-medium text-[#0b1424]/70">
              AO
            </span>
            <div class="min-w-0">
              <p class="truncate text-[10px] font-medium text-[#0b1424]">Adaeze O.</p>
              <p class="truncate text-[9px] text-[#0b1424]/40">Owner · all branches</p>
            </div>
          </div>
        </div>
      </aside>

      <!-- Workspace -->
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2 border-b border-[#0b1424]/8 px-4 py-2.5">
          <div
            class="flex h-7 min-w-0 flex-1 items-center gap-1.5 rounded-md border border-[#0b1424]/8 bg-[#fafaf8] px-2 text-[10px] text-[#0b1424]/35"
          >
            <UiIcon name="search" :size="11" />
            <span class="truncate">Search waybills, bags, phone numbers…</span>
          </div>
          <span class="hidden items-center gap-1 rounded-md border border-[#0b1424]/8 px-2 py-1 text-[10px] text-[#0b1424]/50 md:inline-flex">
            <UiIcon name="building" :size="10" />
            All branches
          </span>
          <span class="flex size-7 items-center justify-center rounded-full bg-[#0b1424] text-[9px] font-medium text-white">
            AO
          </span>
        </div>

        <div class="space-y-3 bg-white p-4">
          <div class="flex items-end justify-between gap-3">
            <div>
              <p class="text-[13px] font-semibold tracking-tight text-[#0b1424]">Dashboard</p>
              <p class="mt-0.5 text-[10px] text-[#0b1424]/40">Wed 11 Nov · last 30 days</p>
            </div>
            <span class="inline-flex items-center gap-1 rounded-md bg-[#2674da] px-2 py-1 text-[10px] font-medium text-white">
              <UiIcon name="plus" :size="10" />
              Send parcel
            </span>
          </div>

          <div class="grid grid-cols-2 gap-2 md:grid-cols-4">
            <div
              v-for="metric in metrics"
              :key="metric.label"
              class="rounded-lg border border-[#0b1424]/8 bg-white p-2.5"
            >
              <p class="truncate text-[9.5px] text-[#0b1424]/45">{{ metric.label }}</p>
              <p
                class="mt-1 text-[15px] font-semibold tracking-tight tabular-nums"
                :class="metric.alert ? 'text-[#c23236]' : 'text-[#0b1424]'"
              >
                {{ metric.value }}
              </p>
              <p class="mt-0.5 truncate text-[9px] text-[#0b1424]/35">{{ metric.hint }}</p>
            </div>
          </div>

          <div class="grid gap-2 lg:grid-cols-[minmax(0,1.6fr)_minmax(0,1fr)]">
            <!-- Movements -->
            <div class="rounded-lg border border-[#0b1424]/8">
              <div class="flex items-center justify-between gap-2 border-b border-[#0b1424]/8 px-3 py-2">
                <p class="text-[11px] font-medium text-[#0b1424]">Live movements</p>
                <span class="text-[9.5px] text-[#2674da]">View all</span>
              </div>
              <table class="w-full border-collapse text-left">
                <thead>
                  <tr class="text-[9px] tracking-[0.06em] text-[#0b1424]/35 uppercase">
                    <th class="px-3 py-1.5 font-medium">Parcel</th>
                    <th class="hidden px-2 py-1.5 font-medium sm:table-cell">Route</th>
                    <th class="px-2 py-1.5 font-medium">Status</th>
                    <th class="hidden px-3 py-1.5 text-right font-medium md:table-cell">Value</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="movement in movements"
                    :key="movement.id"
                    class="border-t border-[#0b1424]/[0.06]"
                  >
                    <td class="px-3 py-2 text-[10px] font-medium whitespace-nowrap text-[#0b1424] tabular-nums">
                      {{ movement.id }}
                    </td>
                    <td class="hidden max-w-[170px] truncate px-2 py-2 text-[9.5px] text-[#0b1424]/50 sm:table-cell">
                      {{ movement.route }}
                    </td>
                    <td class="px-2 py-2">
                      <span
                        class="inline-flex rounded-[3px] px-1.5 py-0.5 text-[8.5px] font-medium tracking-[0.04em] uppercase"
                        :class="TONES[movement.tone]"
                      >
                        {{ movement.status }}
                      </span>
                    </td>
                    <td class="hidden px-3 py-2 text-right text-[9.5px] whitespace-nowrap text-[#0b1424]/55 tabular-nums md:table-cell">
                      {{ movement.value }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Side column -->
            <div class="space-y-2">
              <div class="rounded-lg border border-[#0b1424]/8">
                <div class="flex items-center justify-between gap-2 border-b border-[#0b1424]/8 px-3 py-2">
                  <p class="text-[11px] font-medium text-[#0b1424]">Custody trail</p>
                  <span class="text-[9px] text-[#0b1424]/35 tabular-nums">PKG-00124</span>
                </div>
                <ol class="space-y-2 px-3 py-2.5">
                  <li v-for="(event, index) in custody" :key="event.at" class="relative flex gap-2">
                    <div class="flex flex-col items-center">
                      <span
                        class="mt-1 size-1.5 shrink-0 rounded-full"
                        :class="index === custody.length - 1 ? 'bg-[#2674da]' : 'bg-[#0b1424]/20'"
                      />
                      <span
                        v-if="index < custody.length - 1"
                        class="mt-0.5 w-px flex-1 bg-[#0b1424]/10"
                      />
                    </div>
                    <div class="min-w-0 pb-1">
                      <p class="truncate text-[9.5px] font-medium text-[#0b1424]">{{ event.title }}</p>
                      <p class="truncate text-[8.5px] text-[#0b1424]/40">{{ event.at }} · {{ event.note }}</p>
                    </div>
                  </li>
                </ol>
              </div>

              <div class="rounded-lg border border-[#0b1424]/8 bg-[#fafaf8] p-2.5">
                <div class="flex items-center justify-between gap-2">
                  <p class="text-[10px] font-medium text-[#0b1424]">BAG-2026-00018</p>
                  <span class="inline-flex items-center gap-1 text-[8.5px] font-medium text-[#0a7c6f]">
                    <span class="size-1.5 rounded-full bg-[#0f9d8e]" />
                    Sealed
                  </span>
                </div>
                <p class="mt-1 text-[9px] text-[#0b1424]/45">Lagos — Ikeja → Port Harcourt</p>
                <div class="mt-2 h-1 overflow-hidden rounded-full bg-[#0b1424]/8">
                  <div class="h-full w-[68%] rounded-full bg-[#0b1424]" />
                </div>
                <div class="mt-1.5 flex justify-between text-[8.5px] text-[#0b1424]/40 tabular-nums">
                  <span>28 of 42 scanned</span>
                  <span>Arrives 18:40</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
