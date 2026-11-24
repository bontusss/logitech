<script setup lang="ts">
import {
  ArrowUpRight,
  BarChart3,
  Bell,
  ChevronDown,
  ClipboardList,
  LayoutDashboard,
  Package,
  Route as RouteIcon,
  Search,
  Settings,
  Truck,
  Users,
  Wallet,
} from 'lucide-vue-next'

const navigation = [LayoutDashboard, Package, RouteIcon, ClipboardList, Users, BarChart3]

const stats = [
  { label: 'Shipments today', value: '128', note: '+12% vs yesterday', icon: Truck },
  { label: 'In transit', value: '46', note: '18 active routes', icon: RouteIcon },
  { label: 'Open manifests', value: '8', note: '3 arriving today', icon: ClipboardList },
  { label: 'Collections', value: '₦2.4m', note: 'This week', icon: Wallet },
]

const shipments = [
  { id: 'WB-2041', route: 'Aba → Lagos', customer: 'Ogadinma Textiles', status: 'In transit', tone: 'brand' },
  { id: 'WB-2038', route: 'Onitsha → Abuja', customer: 'M. Okafor', status: 'Delivered', tone: 'success' },
  { id: 'WB-2035', route: 'Port Harcourt → Lagos', customer: 'Delta Supply Co.', status: 'Pending pickup', tone: 'accent' },
]

const toneClasses: Record<string, string> = {
  brand: 'border-brand-200 bg-brand-50 text-ink-900 dark:border-white/10 dark:bg-white/10 dark:text-brand-200',
  success: 'border-success-200 bg-success-50 text-success-700 dark:border-success-500/25 dark:bg-success-500/10',
  accent: 'border-accent-200 bg-accent-50 text-accent-800 dark:border-accent-500/25 dark:bg-accent-500/10 dark:text-accent-300',
}

const bars = [38, 52, 44, 68, 58, 76, 62, 84, 71, 92, 80, 96]
</script>

<template>
  <div class="relative">
    <div
      aria-hidden="true"
      class="absolute -inset-x-6 -top-6 bottom-0 rounded-3xl bg-gradient-to-b from-brand-100/50 to-transparent dark:from-brand-400/10"
    />

    <div
      class="relative overflow-hidden rounded-2xl border border-ink-900/10 bg-surface shadow-[0_40px_90px_-55px_rgba(5,13,31,0.55)]"
    >
      <!-- Browser chrome -->
      <div class="flex items-center gap-2 border-b border-ink-900/8 bg-canvas px-3 py-2.5">
        <span class="size-2 rounded-full bg-ink-900/15" />
        <span class="size-2 rounded-full bg-ink-900/15" />
        <span class="size-2 rounded-full bg-ink-900/15" />
        <div
          class="mx-auto flex items-center gap-1.5 rounded-md border border-ink-900/10 bg-surface px-2.5 py-1 text-[9.5px] text-ink-900/40"
        >
          <span class="size-1.5 rounded-full bg-success-500" />
          app.basedonlogistics.com
        </div>
        <span
          class="rounded border border-ink-900/10 bg-surface px-1.5 py-0.5 text-[9px] font-medium tracking-[0.08em] text-ink-900/40 uppercase"
        >
          Sample
        </span>
      </div>

      <div class="flex">
        <!-- Icon rail -->
        <aside
          class="hidden w-[56px] shrink-0 flex-col items-center gap-1 border-r border-white/10 bg-ink-950 py-3 sm:flex"
        >
          <AppLogo :size="26" :wordmark="false" class="mb-2" />
          <button
            v-for="(icon, index) in navigation"
            :key="index"
            type="button"
            tabindex="-1"
            class="flex size-8 items-center justify-center rounded-md"
            :class="index === 0 ? 'bg-accent-500 text-ink-950 shadow-sm shadow-accent-500/40' : 'text-white/35'"
            :aria-label="`Dashboard navigation item ${index + 1} (illustrative)`"
          >
            <component :is="icon" :size="15" />
          </button>
          <button
            type="button"
            tabindex="-1"
            class="mt-auto flex size-8 items-center justify-center rounded-md text-white/35"
            aria-label="Settings (illustrative)"
          >
            <Settings :size="15" />
          </button>
        </aside>

        <div class="min-w-0 flex-1">
          <!-- Top bar -->
          <div class="flex items-center gap-3 border-b border-ink-900/8 px-4 py-2.5">
            <div class="min-w-0">
              <p class="text-[12px] font-semibold tracking-[-0.01em]">Operations overview</p>
              <p class="text-[9.5px] text-ink-900/40">Aba headquarters · All branches</p>
            </div>
            <div
              class="ml-auto hidden items-center gap-1.5 rounded-md border border-ink-900/10 px-2 py-1 text-[9.5px] text-ink-900/40 md:flex"
            >
              <Search :size="11" /> Search
            </div>
            <button
              type="button"
              tabindex="-1"
              class="text-ink-900/35"
              aria-label="Notifications (illustrative)"
            >
              <Bell :size="14" />
            </button>
            <div class="flex items-center gap-1.5">
              <span
                class="flex size-6 items-center justify-center rounded-full bg-ink-950 text-[8px] font-semibold text-white"
              >
                AO
              </span>
              <ChevronDown :size="11" class="hidden text-ink-900/30 sm:block" />
            </div>
          </div>

          <div class="p-4 sm:p-5">
            <!-- Stat cards -->
            <div class="grid grid-cols-2 gap-2 lg:grid-cols-4">
              <div
                v-for="stat in stats"
                :key="stat.label"
                class="rounded-lg border border-ink-900/8 bg-canvas p-2.5"
              >
                <div class="flex items-center justify-between">
                  <p class="text-[9.5px] font-medium text-ink-900/45">{{ stat.label }}</p>
                  <component :is="stat.icon" :size="12" class="text-ink-900/30" />
                </div>
                <p class="mt-1.5 text-[16px] font-semibold tracking-[-0.03em]">{{ stat.value }}</p>
                <p class="mt-0.5 text-[9px] text-ink-900/35">{{ stat.note }}</p>
              </div>
            </div>

            <div class="mt-3 grid gap-3 lg:grid-cols-[minmax(0,1.65fr)_minmax(0,1fr)]">
              <!-- Shipment table -->
              <div class="overflow-hidden rounded-lg border border-ink-900/8">
                <div class="flex items-center justify-between border-b border-ink-900/8 px-3 py-2">
                  <p class="text-[10.5px] font-semibold">Recent shipments</p>
                  <span class="inline-flex items-center gap-0.5 text-[9px] font-medium text-accent-700 dark:text-accent-400">
                    View all <ArrowUpRight :size="10" />
                  </span>
                </div>
                <div class="overflow-x-auto">
                  <table class="w-full min-w-[380px] text-left">
                    <thead>
                      <tr class="border-b border-ink-900/6 text-[8.5px] tracking-[0.06em] text-ink-900/35 uppercase">
                        <th class="px-3 py-1.5 font-medium">Waybill</th>
                        <th class="px-2 py-1.5 font-medium">Route</th>
                        <th class="px-2 py-1.5 font-medium">Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="shipment in shipments"
                        :key="shipment.id"
                        class="border-b border-ink-900/5 last:border-0"
                      >
                        <td class="px-3 py-2 align-top">
                          <p class="text-[10px] font-semibold">{{ shipment.id }}</p>
                          <p class="mt-0.5 text-[8.5px] text-ink-900/35">{{ shipment.customer }}</p>
                        </td>
                        <td class="px-2 py-2 align-top text-[9.5px] whitespace-nowrap text-ink-900/60">
                          {{ shipment.route }}
                        </td>
                        <td class="px-2 py-2 align-top">
                          <span
                            class="inline-block rounded border px-1.5 py-0.5 text-[8.5px] font-medium whitespace-nowrap"
                            :class="toneClasses[shipment.tone]"
                          >
                            {{ shipment.status }}
                          </span>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <!-- Volume chart -->
              <div class="rounded-lg border border-ink-900/8 p-3">
                <div class="flex items-start justify-between">
                  <div>
                    <p class="text-[10.5px] font-semibold">Weekly volume</p>
                    <p class="mt-0.5 text-[8.5px] text-ink-900/35">Shipments created</p>
                  </div>
                  <span class="rounded border border-success-200 bg-success-50 px-1.5 py-0.5 text-[8px] font-medium text-success-700 dark:border-success-500/25 dark:bg-success-500/10">
                    +14%
                  </span>
                </div>
                <div class="mt-3 flex h-[74px] items-end gap-1">
                  <div
                    v-for="(bar, index) in bars"
                    :key="index"
                    class="flex-1 rounded-sm"
                    :class="index === bars.length - 1 ? 'bg-accent-500' : 'bg-brand-100 dark:bg-white/10'"
                    :style="{ height: `${bar}%` }"
                  />
                </div>
                <div class="mt-2 flex justify-between text-[8px] text-ink-900/30">
                  <span>Mon</span><span>Wed</span><span>Fri</span><span>Sun</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <p class="mt-2.5 text-center text-[10.5px] text-ink-900/35">
      Illustrative dashboard with sample data.
    </p>
  </div>
</template>
