<script setup lang="ts">
/**
 * Panel shell: an off-white card holding the screen title, with a white inner
 * card for the content. Used for dialogs and app panels.
 */
defineProps<{ title?: string }>()
</script>

<template>
  <section class="w-full max-w-[360px] rounded-lg bg-canvas p-3 shadow-card">
    <h1 v-if="title" class="px-1 pb-3 text-label font-normal text-ink">{{ title }}</h1>
    <div class="rounded-lg bg-surface p-4 shadow-card">
      <slot />
    </div>
  </section>
</template>
