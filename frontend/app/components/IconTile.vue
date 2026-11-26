<script setup lang="ts">
withDefaults(
  defineProps<{
    tone?: 'brand' | 'accent' | 'accent-solid' | 'success' | 'neutral' | 'on-dark'
    size?: 'sm' | 'md' | 'lg'
    rounded?: 'lg' | 'xl' | '2xl'
  }>(),
  { tone: 'brand', size: 'md', rounded: 'xl' },
)

const tones: Record<string, string> = {
  brand: 'bg-brand-50 text-ink-900 dark:bg-white/5',
  accent: 'bg-accent-50 text-accent-700 dark:bg-accent-500/10 dark:text-accent-300',
  'accent-solid': 'bg-accent-500 text-white shadow-accent-500/25',
  success: 'bg-success-50 text-success-700 dark:bg-success-500/10',
  neutral: 'bg-canvas text-ink-900/55 dark:bg-white/5',
  'on-dark': 'bg-white/10 text-accent-300',
}

const sizes: Record<string, string> = {
  sm: 'size-8',
  md: 'size-10',
  lg: 'size-11',
}

const radii: Record<string, string> = {
  lg: 'rounded-lg',
  xl: 'rounded-xl',
  '2xl': 'rounded-2xl',
}
</script>

<template>
  <span
    class="flex shrink-0 items-center justify-center shadow-sm"
    :class="[tones[tone], sizes[size], radii[rounded]]"
  >
    <slot />
  </span>
</template>
