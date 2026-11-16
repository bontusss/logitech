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
</script>

<template>
  <div class="grid gap-4 sm:grid-cols-2">
    <div
      v-for="group in groups"
      :key="group.title"
      class="rounded-xl border border-ink-900/10 bg-white p-5 transition hover:border-ink-900/20 hover:shadow-[0_18px_40px_-30px_rgba(10,18,32,0.4)]"
    >
      <div class="flex items-center gap-2.5">
        <div
          class="flex size-8 items-center justify-center rounded-lg border border-brand-100 bg-brand-50 text-brand-600"
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
          <Check :size="13" class="mt-0.5 shrink-0 text-brand-600" />
          {{ item }}
        </li>
      </ul>
    </div>
  </div>
</template>
