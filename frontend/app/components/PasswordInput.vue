<script setup lang="ts">
import { Eye, EyeOff } from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    modelValue: string
    required?: boolean
    minlength?: number
    autocomplete?: string
    placeholder?: string
    inputClass?: string
  }>(),
  {
    inputClass:
      'w-full rounded-xl border border-ink-900/15 bg-canvas px-3.5 py-2.5 text-sm text-ink-900 transition placeholder:text-ink-900/35 hover:border-ink-900/30 focus:border-accent-500 focus:bg-surface focus:outline-none',
  },
)

defineEmits<{ 'update:modelValue': [value: string] }>()

const visible = ref(false)
</script>

<template>
  <div class="relative">
    <input
      :value="props.modelValue"
      :type="visible ? 'text' : 'password'"
      :required="props.required"
      :minlength="props.minlength"
      :autocomplete="props.autocomplete"
      :placeholder="props.placeholder"
      :class="props.inputClass"
      style="padding-right: 2.75rem"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <button
      type="button"
      class="absolute inset-y-0 right-0 flex w-11 items-center justify-center text-ink-900/40 transition hover:text-ink-900"
      :aria-label="visible ? 'Hide password' : 'Show password'"
      :aria-pressed="visible"
      @click="visible = !visible"
    >
      <EyeOff v-if="visible" :size="16" />
      <Eye v-else :size="16" />
    </button>
  </div>
</template>
