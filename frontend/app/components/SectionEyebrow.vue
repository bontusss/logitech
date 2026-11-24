<script setup lang="ts">
withDefaults(
  defineProps<{
    label?: string
    tone?: 'dark' | 'light' | 'surface'
    dot?: boolean
  }>(),
  { tone: 'dark', dot: true },
)
</script>

<template>
  <p
    class="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[10px] font-bold tracking-[0.18em] uppercase"
    :class="
      tone === 'light'
        ? 'border-white/15 bg-white/5 text-accent-300'
        : tone === 'surface'
          ? 'border-ink-900/10 bg-surface/85 text-ink-900 shadow-sm backdrop-blur-sm'
          : 'border-brand-200/70 bg-brand-50/70 text-brand-700 dark:border-white/10 dark:bg-white/5'
    "
  >
    <span v-if="dot" class="size-1.5 shrink-0 rounded-full bg-accent-500" />
    <slot>{{ label }}</slot>
  </p>
</template>
