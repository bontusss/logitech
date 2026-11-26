<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    to?: string
    variant?: 'primary' | 'secondary' | 'ghost' | 'light' | 'outline-light'
    size?: 'sm' | 'md' | 'lg'
    block?: boolean
  }>(),
  { to: '#', variant: 'primary', size: 'md', block: false },
)

const base =
  'inline-flex items-center justify-center gap-2 rounded-lg font-semibold transition duration-200 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent-500'

const variants: Record<string, string> = {
  primary:
    'bg-accent-500 text-white shadow-[0_10px_24px_-14px_rgba(255,122,0,0.75)] hover:-translate-y-0.5 hover:bg-accent-400 hover:shadow-[0_14px_28px_-14px_rgba(255,122,0,0.85)]',
  secondary:
    'border border-ink-900/15 bg-surface text-ink-900 hover:border-ink-900/30 hover:bg-brand-50 dark:bg-white/5 dark:hover:bg-white/10',
  ghost: 'text-ink-900/70 hover:bg-ink-900/5 hover:text-ink-900',
  light: 'bg-accent-500 text-white hover:bg-accent-400',
  'outline-light': 'border border-white/25 text-white hover:bg-white/10',
}

const sizes: Record<string, string> = {
  sm: 'px-3 py-1.5 text-[13px]',
  md: 'px-4 py-2.5 text-sm',
  lg: 'px-5 py-3 text-[15px]',
}

const classes = computed(() => [
  base,
  variants[props.variant],
  sizes[props.size],
  props.block ? 'w-full' : '',
])
</script>

<template>
  <NuxtLink :to="to" :class="classes">
    <slot />
  </NuxtLink>
</template>
