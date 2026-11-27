<script setup lang="ts">
type StatTone = 'brand' | 'accent' | 'success' | 'neutral'

withDefaults(
  defineProps<{
    label: string
    value: string | number
    detail?: string
    tone?: StatTone
  }>(),
  { tone: 'brand', detail: '' },
)
</script>

<template>
  <article class="dash-stat-card rounded-xl border p-3.5 sm:p-4" :class="`dash-stat-card--${tone}`">
    <p class="text-[10px] font-medium text-ink-900/45">{{ label }}</p>
    <p class="mt-1.5 text-[16px] font-semibold tracking-[-0.035em] tabular-nums sm:mt-2 sm:text-[18px]">
      {{ value }}
    </p>
    <p v-if="detail" class="mt-1 text-[9.5px] leading-snug text-ink-900/35 sm:text-[10px]">{{ detail }}</p>
  </article>
</template>
