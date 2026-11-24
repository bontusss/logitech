<script setup lang="ts">
const model = defineModel<string>({ default: '' })

withDefaults(
  defineProps<{
    label: string
    type?: string
    placeholder?: string
    autocomplete?: string
    required?: boolean
    minlength?: number
  }>(),
  { type: 'text' },
)
</script>

<template>
  <label class="block">
    <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">{{ label }}</span>
    <input
      v-model="model"
      :type="type"
      :placeholder="placeholder"
      :autocomplete="autocomplete"
      :required="required"
      :minlength="minlength"
      class="w-full rounded-xl border border-ink-900/15 bg-canvas px-3.5 py-2.5 text-sm text-ink-900 transition placeholder:text-ink-900/35 hover:border-ink-900/30 focus:border-accent-500 focus:bg-surface focus:outline-none"
    />
  </label>
</template>
