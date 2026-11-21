<script setup lang="ts">
import { BadgeCheck, CircleHelp } from 'lucide-vue-next'

withDefaults(
  defineProps<{
    type: 'confirmed' | 'estimated'
    note?: string
    size?: 'sm' | 'md'
  }>(),
  { note: '', size: 'sm' },
)
</script>

<template>
  <span
    class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 font-semibold"
    :class="[
      type === 'confirmed'
        ? 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-400/25 dark:bg-emerald-400/10 dark:text-emerald-200'
        : 'border-amber-200 bg-amber-50 text-amber-800 dark:border-amber-400/25 dark:bg-amber-400/10 dark:text-amber-200',
      size === 'md' ? 'px-2.5 py-1 text-[11px]' : 'text-[10px]',
    ]"
    :title="note || (type === 'confirmed' ? 'Confirmed by the provider' : 'Estimated by the provider, subject to final confirmation')"
  >
    <BadgeCheck v-if="type === 'confirmed'" :size="size === 'md' ? 13 : 11" />
    <CircleHelp v-else :size="size === 'md' ? 13 : 11" />
    {{ type === 'confirmed' ? 'Confirmed price' : 'Estimated price' }}
  </span>
</template>
