<script setup lang="ts">
/** Small pill tag: danger (red fill), info (pale blue fill, blue text), or neutral (outline only). */
withDefaults(defineProps<{ tone?: 'danger' | 'info' | 'neutral' }>(), { tone: 'neutral' })
</script>

<template>
  <span
    :class="[
      'inline-flex h-5 items-center rounded-tag px-2 text-meta font-medium tracking-[0.06em] uppercase',
      tone === 'danger' ? 'bg-danger/10 text-danger shadow-inset' : tone === 'info' ? 'bg-chip/45 text-blue' : 'text-ink-40 shadow-inset',
    ]"
  >
    <slot />
  </span>
</template>
