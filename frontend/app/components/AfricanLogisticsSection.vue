<script setup lang="ts">
const problems = [
  'Fragmented logistics providers',
  'Difficulty discovering reliable intercity transport',
  'Inconsistent and inaccessible shipping rates',
  'Manual shipment and manifest management',
  'Poor visibility across branches',
  'Limited digital infrastructure for independent operators',
]

const cities = [
  { name: 'Lagos', x: 14, y: 72 },
  { name: 'Abuja', x: 52, y: 24 },
  { name: 'Port Harcourt', x: 60, y: 84 },
  { name: 'Aba', x: 68, y: 70 },
  { name: 'Umuahia', x: 76, y: 62 },
  { name: 'Onitsha', x: 62, y: 50 },
  { name: 'Enugu', x: 71, y: 44 },
]

const routes = [
  [14, 72, 52, 24],
  [14, 72, 60, 84],
  [60, 84, 68, 70],
  [68, 70, 76, 62],
  [68, 70, 62, 50],
  [62, 50, 71, 44],
  [71, 44, 52, 24],
  [52, 24, 62, 50],
]
</script>

<template>
  <section class="border-t border-ink-900/10 bg-ink-950 text-white">
    <div class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-24">
      <div class="grid gap-12 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] lg:items-center">
        <div>
          <SectionHeading
            tone="light"
            eyebrow="Built for African logistics"
            title="Designed for the realities of African logistics."
            lede="Intercity transport is often fragmented across independent operators, paper manifests and informal rate sharing. BOL connects those pieces without pretending the infrastructure is uniform."
          />

          <ul class="mt-8 grid gap-x-6 gap-y-3 sm:grid-cols-2">
            <li
              v-for="problem in problems"
              :key="problem"
              class="flex items-start gap-2.5 text-[13px] leading-snug text-white/55"
            >
              <span class="mt-1.5 size-1.5 shrink-0 rounded-full bg-brand-500" />
              {{ problem }}
            </li>
          </ul>
        </div>

        <div>
          <div
            class="relative h-[360px] overflow-hidden rounded-2xl border border-white/10 bg-white/[0.03] sm:h-[400px]"
          >
            <div aria-hidden="true" class="absolute inset-0 grid-paper-light opacity-70" />
            <div
              aria-hidden="true"
              class="absolute top-1/2 left-1/2 h-[280px] w-[280px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-brand-500/15 blur-[80px]"
            />

            <div class="absolute top-4 left-4 z-10">
              <p class="text-[10px] font-semibold tracking-[0.18em] text-white/40 uppercase">
                Example intercity network
              </p>
              <p class="mt-1 text-[12px] text-white/60">Southern and central corridors</p>
            </div>

            <svg
              class="absolute inset-0 h-full w-full"
              viewBox="0 0 100 100"
              preserveAspectRatio="none"
              aria-hidden="true"
            >
              <line
                v-for="(route, index) in routes"
                :key="index"
                :x1="route[0]"
                :y1="route[1]"
                :x2="route[2]"
                :y2="route[3]"
                stroke="#2b4bff"
                stroke-width="0.55"
                stroke-opacity="0.45"
              />
            </svg>

            <div
              v-for="city in cities"
              :key="city.name"
              class="absolute z-10 -translate-x-1/2 -translate-y-1/2"
              :style="{ left: `${city.x}%`, top: `${city.y}%` }"
            >
              <span class="flex items-center gap-1.5">
                <span class="relative flex size-2">
                  <span
                    class="absolute inline-flex size-full animate-ping rounded-full bg-brand-500 opacity-40"
                  />
                  <span class="relative inline-flex size-2 rounded-full bg-brand-500" />
                </span>
                <span
                  class="rounded border border-white/12 bg-ink-950/90 px-1.5 py-0.5 text-[9.5px] font-medium whitespace-nowrap text-white/75"
                >
                  {{ city.name }}
                </span>
              </span>
            </div>
          </div>

          <p class="mt-3 text-[11px] leading-relaxed text-white/35">
            Illustrative network only. Routes shown are examples, not a guarantee of current
            coverage. Marketplace availability grows as providers publish routes.
          </p>
        </div>
      </div>
    </div>
  </section>
</template>
