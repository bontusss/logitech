<script setup lang="ts">
withDefaults(
  defineProps<{
    eyebrow?: string
    title: string
    lede?: string
    align?: 'left' | 'center'
    tone?: 'dark' | 'light'
  }>(),
  { eyebrow: '', lede: '', align: 'left', tone: 'dark' },
)
</script>

<template>
  <div :class="align === 'center' ? 'mx-auto max-w-3xl text-center' : 'max-w-2xl'">
    <p
      v-if="eyebrow"
      class="text-[11px] font-semibold tracking-[0.22em] uppercase"
      :class="tone === 'light' ? 'text-[#7fb2ff]' : 'text-[#2674da]'"
    >
      {{ eyebrow }}
    </p>
    <h2
      class="mt-3 text-[clamp(1.65rem,3.4vw,2.6rem)] leading-[1.1] font-semibold tracking-[-0.03em] text-balance"
      :class="tone === 'light' ? 'text-white' : 'text-[#0b1424]'"
    >
      <slot name="title">{{ title }}</slot>
    </h2>
    <p
      v-if="lede"
      class="mt-4 text-[16.5px] leading-[1.65] text-pretty"
      :class="tone === 'light' ? 'text-white/60' : 'text-[#0b1424]/60'"
    >
      {{ lede }}
    </p>
  </div>
</template>
