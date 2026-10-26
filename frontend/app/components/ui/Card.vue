<script setup lang="ts">
/** Floating surface: white card, faint outer ring, soft layered shadow. */
withDefaults(defineProps<{ halo?: boolean; padded?: boolean }>(), {
  halo: false,
  padded: true,
})
</script>

<template>
  <div
    :class="[
      'rounded-lg bg-surface shadow-card',
      padded && 'p-4',
      halo ? 'shadow-halo' : '',
    ]"
  >
    <slot />
  </div>
</template>
