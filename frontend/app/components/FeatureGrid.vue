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

const tones: Record<string, 'brand' | 'accent' | 'success'> = {
  building: 'brand',
  truck: 'brand',
  tag: 'accent',
  chart: 'success',
}
</script>

<template>
  <div class="grid gap-4 sm:grid-cols-2">
    <div
      v-for="group in groups"
      :key="group.title"
      class="site-card rounded-2xl border p-5"
    >
      <div class="flex items-center gap-2.5">
        <IconTile :tone="tones[group.icon]" size="sm" rounded="lg">
          <component :is="icons[group.icon]" :size="16" />
        </IconTile>
        <h3 class="text-[14px] font-semibold tracking-[-0.01em]">{{ group.title }}</h3>
      </div>

      <ul class="mt-4 space-y-2">
        <li
          v-for="item in group.items"
          :key="item"
          class="flex items-start gap-2 text-[13px] leading-snug text-ink-900/60"
        >
          <Check :size="13" class="mt-0.5 shrink-0 text-success-600" />
          {{ item }}
        </li>
      </ul>
    </div>
  </div>
</template>
