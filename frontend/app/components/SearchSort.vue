<script setup lang="ts">
const props = defineProps<{ modelValue: string; count: number }>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const options = [
  { value: 'recommended', label: 'Recommended' },
  { value: 'cheapest', label: 'Cheapest' },
  { value: 'fastest', label: 'Fastest' },
] as const
</script>

<template>
  <div class="flex flex-wrap items-center justify-between gap-2">
    <p class="text-[12px] font-medium text-ink-900/55">
      <span class="font-bold text-ink-900">{{ count }}</span>
      {{ count === 1 ? 'option' : 'options' }} found
    </p>
    <div class="flex items-center gap-1.5">
      <span class="text-[10.5px] font-semibold text-ink-900/40">Sort</span>
      <div class="flex rounded-full border border-ink-900/10 bg-white/70 p-0.5 dark:border-white/10 dark:bg-white/5">
        <button
          v-for="option in options"
          :key="option.value"
          type="button"
          class="rounded-full px-2.5 py-1 text-[10.5px] font-semibold transition"
          :class="
            modelValue === option.value
              ? 'bg-ink-900 text-white dark:bg-white dark:text-ink-900'
              : 'text-ink-900/55 hover:text-ink-900'
          "
          @click="emit('update:modelValue', option.value)"
        >
          {{ option.label }}
        </button>
      </div>
    </div>
  </div>
</template>
