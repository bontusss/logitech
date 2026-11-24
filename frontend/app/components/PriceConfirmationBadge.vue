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
        ? 'border-success-200 bg-success-50 text-success-700 dark:border-success-500/25 dark:bg-success-500/10 dark:text-success-200'
        : 'border-accent-200 bg-accent-50 text-accent-800 dark:border-accent-400/25 dark:bg-accent-400/10 dark:text-accent-200',
      size === 'md' ? 'px-2.5 py-1 text-[11px]' : 'text-[10px]',
    ]"
    :title="note || (type === 'confirmed' ? 'Confirmed by the provider' : 'Estimated by the provider, subject to final confirmation')"
  >
    <BadgeCheck v-if="type === 'confirmed'" :size="size === 'md' ? 13 : 11" />
    <CircleHelp v-else :size="size === 'md' ? 13 : 11" />
    {{ type === 'confirmed' ? 'Confirmed price' : 'Estimated price' }}
  </span>
</template>
