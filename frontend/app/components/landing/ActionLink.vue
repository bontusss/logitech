<script setup lang="ts">
/** Landing-only action: navy primary, quiet secondary, or white on navy. */
const props = withDefaults(
  defineProps<{
    to: string
    variant?: 'primary' | 'secondary' | 'light' | 'quiet'
    size?: 'md' | 'lg'
  }>(),
  { variant: 'primary', size: 'md' },
)

const isAnchor = computed(() => /^(https?:|mailto:|tel:|#)/.test(props.to))

const VARIANTS = {
  primary:
    'bg-[#0b1424] text-white shadow-[0_1px_2px_rgba(11,20,36,0.28),inset_0_1px_0_rgba(255,255,255,0.12)] hover:bg-[#17243c] active:bg-[#0b1424]',
  secondary:
    'border border-[#0b1424]/12 bg-white text-[#0b1424] hover:border-[#0b1424]/25 hover:bg-[#0b1424]/[0.035]',
  light: 'bg-white text-[#0b1424] shadow-[0_1px_2px_rgba(0,0,0,0.2)] hover:bg-[#f2f4f8]',
  quiet: 'text-[#0b1424]/70 hover:text-[#0b1424]',
} as const

const SIZES = {
  md: 'h-10 px-4 text-[14px]',
  lg: 'h-12 px-5 text-[15px]',
} as const
</script>

<template>
  <a
    v-if="isAnchor"
    :href="to"
    :class="[
      'inline-flex shrink-0 items-center justify-center gap-2 rounded-[10px] font-medium whitespace-nowrap transition duration-200 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#2674da]',
      VARIANTS[props.variant],
      SIZES[props.size],
    ]"
  >
    <slot />
  </a>
  <NuxtLink
    v-else
    :to="to"
    :class="[
      'inline-flex shrink-0 items-center justify-center gap-2 rounded-[10px] font-medium whitespace-nowrap transition duration-200 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#2674da]',
      VARIANTS[props.variant],
      SIZES[props.size],
    ]"
  >
    <slot />
  </NuxtLink>
</template>
