<script setup lang="ts">
/** Signed-out screens: the panel is centred and the theme toggle floats above it. */
const { isDark, label, toggle } = useTheme()
</script>

<template>
  <div class="relative flex min-h-dvh flex-col items-center justify-center px-4 py-16">
    <button
      type="button"
      class="absolute top-3 right-3 flex size-9 items-center justify-center rounded-lg text-ink-64 transition hover:bg-surface hover:text-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
      :aria-label="label"
      :title="label"
      @click="toggle"
    >
      <UiIcon :name="isDark ? 'sun' : 'moon'" :size="17" />
    </button>

    <slot />
  </div>
</template>
