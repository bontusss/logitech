<script setup lang="ts">
import type { DailyReportRow } from '~/types/api'

const props = withDefaults(
  defineProps<{ rows: DailyReportRow[]; metric?: 'collected' | 'billed'; days?: number }>(),
  { metric: 'collected', days: 14 },
)

const slice = computed(() => props.rows.slice(-props.days))
const max = computed(() => Math.max(1, ...slice.value.map((row) => row[props.metric])))
const total = computed(() => slice.value.reduce((sum, row) => sum + row[props.metric], 0))
const peak = computed(() => slice.value.reduce<DailyReportRow | null>((best, row) => (!best || row[props.metric] > best[props.metric] ? row : best), null))

function barHeight(value: number): string {
  if (value <= 0) return '2px'
  return `${Math.max(4, Math.round((value / max.value) * 100))}%`
}
</script>

<template>
  <div>
    <div class="flex items-baseline justify-between gap-3">
      <p class="text-h5 font-medium text-ink">{{ money(total) }}</p>
      <p v-if="peak" class="text-meta text-ink-40">Best day {{ shortDate(peak.date) }} · {{ money(peak[props.metric]) }}</p>
    </div>

    <div class="mt-4 flex h-28 items-end gap-[3px]" role="img" :aria-label="`${metric} over the last ${slice.length} days`">
      <div
        v-for="row in slice"
        :key="row.date"
        class="group flex h-full flex-1 flex-col justify-end"
        :title="`${shortDate(row.date)} · ${money(row[props.metric])}`"
      >
        <div
          class="w-full rounded-t-[3px] bg-blue/20 transition group-hover:bg-blue/50"
          :style="{ height: barHeight(row[props.metric]) }"
        />
      </div>
    </div>

    <div class="mt-2 flex justify-between text-meta text-ink-40">
      <span>{{ shortDate(slice[0]?.date) }}</span>
      <span>{{ shortDate(slice[slice.length - 1]?.date) }}</span>
    </div>
  </div>
</template>
