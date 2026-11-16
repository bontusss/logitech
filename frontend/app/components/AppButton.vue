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
  'inline-flex items-center justify-center gap-2 rounded-lg font-medium transition duration-200 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-500'

const variants: Record<string, string> = {
  primary:
    'bg-brand-500 text-white shadow-[0_10px_28px_-14px_rgba(43,75,255,0.8)] hover:bg-brand-600',
  secondary:
    'border border-slate-300 bg-white text-ink-900 hover:border-slate-400 hover:bg-slate-50',
  ghost: 'text-ink-900/70 hover:bg-ink-900/5 hover:text-ink-900',
  light: 'bg-white text-ink-900 hover:bg-slate-100',
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
