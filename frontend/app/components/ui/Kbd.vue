<script setup lang="ts">
/** Keyboard keycap hint shown next to a button label, e.g. Sign in [Enter]. */
withDefaults(defineProps<{ label?: string }>(), { label: '' })
</script>

<template>
  <kbd
    class="inline-flex h-5 min-w-5 items-center justify-center rounded-tag bg-ink/5 px-1.5 font-sans text-meta leading-none font-normal text-ink-64 shadow-inset"
  >
    <slot>{{ label }}</slot>
  </kbd>
</template>
