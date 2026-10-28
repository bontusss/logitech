<script setup lang="ts">
withDefaults(defineProps<{ title: string; description?: string; icon?: string }>(), {
  description: '',
  icon: 'package',
})
</script>

<template>
  <div class="flex flex-col items-center rounded-lg border border-dashed border-line px-6 py-10 text-center">
    <span class="flex size-11 items-center justify-center rounded-full bg-canvas text-ink-40">
      <UiIcon :name="icon" :size="20" />
    </span>
    <p class="mt-3 text-p1 font-medium text-ink">{{ title }}</p>
    <p v-if="description" class="mt-1 max-w-[34ch] text-label text-ink-64">{{ description }}</p>
    <div v-if="$slots.default" class="mt-4">
      <slot />
    </div>
  </div>
</template>
