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
    <p
      v-if="eyebrow"
      class="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[10px] font-semibold tracking-[0.18em] uppercase"
      :class="tone === 'light' ? 'border-white/15 bg-white/5 text-brand-200' : 'border-brand-200/70 bg-brand-50/70 text-brand-700 dark:border-brand-400/20 dark:bg-brand-400/10 dark:text-indigo-200'"
    >
      <span
        class="size-1.5 rounded-full bg-gradient-to-r"
        :class="tone === 'light' ? 'from-brand-300 to-cyan-300' : 'from-brand-500 to-violet-500'"
      />
      {{ eyebrow }}
    </p>
    <h2
      class="mt-3 text-[clamp(1.7rem,3.6vw,2.6rem)] leading-[1.1] font-semibold tracking-[-0.035em] text-balance"
      :class="tone === 'light' ? 'text-white' : 'text-ink-900'"
    >
      {{ title }}
    </h2>
    <p
      v-if="lede"
      class="mt-4 text-[16px] leading-[1.65] text-pretty"
      :class="tone === 'light' ? 'text-white/55' : 'text-ink-900/55'"
    >
      {{ lede }}
    </p>
  </div>
</template>
