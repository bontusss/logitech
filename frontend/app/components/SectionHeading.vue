<script setup lang="ts">
withDefaults(
  defineProps<{
    eyebrow?: string
    title: string
    lede?: string
    align?: 'left' | 'center'
    tone?: 'dark' | 'light'
  }>(),
  { align: 'left', tone: 'dark' },
)
</script>

<template>
  <div :class="align === 'center' ? 'mx-auto max-w-2xl text-center' : 'max-w-2xl'">
    <SectionEyebrow v-if="eyebrow" :label="eyebrow" :tone="tone" :class="align === 'center' ? 'justify-center' : ''" />
    <h2
      class="mt-3 text-[clamp(1.7rem,3.6vw,2.6rem)] leading-[1.1] font-semibold tracking-[-0.035em] text-balance"
      :class="tone === 'light' ? 'text-white' : 'text-ink-900'"
    >
      {{ title }}
    </h2>
    <p
      v-if="lede"
      class="mt-4 text-[16px] leading-[1.65] text-pretty"
      :class="tone === 'light' ? 'text-white/60' : 'text-ink-900/60'"
    >
      {{ lede }}
    </p>
  </div>
</template>
