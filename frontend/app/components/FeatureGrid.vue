<script setup lang="ts">
import { BarChart3, Building2, Check, Tags, Truck } from 'lucide-vue-next'

defineProps<{
  groups: Array<{
    icon: 'building' | 'truck' | 'tag' | 'chart'
    title: string
    items: string[]
  }>
}>()

const icons = {
  building: Building2,
  truck: Truck,
  tag: Tags,
  chart: BarChart3,
}

const tones: Record<string, { card: string; tile: string }> = {
  building: {
    card: 'from-violet-50 to-white dark:from-violet-400/10 dark:to-transparent',
    tile: 'bg-violet-100 text-violet-700 dark:bg-violet-400/15 dark:text-violet-200',
  },
  truck: {
    card: 'from-sky-50 to-white dark:from-sky-400/10 dark:to-transparent',
    tile: 'bg-sky-100 text-sky-700 dark:bg-sky-400/15 dark:text-sky-200',
  },
  tag: {
    card: 'from-amber-50 to-white dark:from-amber-400/10 dark:to-transparent',
    tile: 'bg-amber-100 text-amber-700 dark:bg-amber-400/15 dark:text-amber-200',
  },
  chart: {
    card: 'from-emerald-50 to-white dark:from-emerald-400/10 dark:to-transparent',
    tile: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-400/15 dark:text-emerald-200',
  },
}
</script>

<template>
  <div class="grid gap-4 sm:grid-cols-2">
    <div
      v-for="group in groups"
      :key="group.title"
      class="site-card rounded-2xl border bg-gradient-to-br p-5"
      :class="tones[group.icon].card"
    >
      <div class="flex items-center gap-2.5">
        <div
          class="flex size-8 items-center justify-center rounded-lg shadow-sm"
          :class="tones[group.icon].tile"
        >
          <component :is="icons[group.icon]" :size="16" />
        </div>
        <h3 class="text-[14px] font-semibold tracking-[-0.01em]">{{ group.title }}</h3>
      </div>

      <ul class="mt-4 space-y-2">
        <li
          v-for="item in group.items"
          :key="item"
          class="flex items-start gap-2 text-[13px] leading-snug text-ink-900/60"
        >
          <Check :size="13" class="mt-0.5 shrink-0 text-emerald-600 dark:text-emerald-300" />
          {{ item }}
        </li>
      </ul>
    </div>
  </div>
</template>
