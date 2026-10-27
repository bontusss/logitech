<script setup lang="ts">
const props = withDefaults(defineProps<{ modelValue?: string; placeholder?: string }>(), {
  modelValue: '',
  placeholder: 'Waybill, phone, or manifest number',
})

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

function onInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}
</script>

<template>
  <div class="relative mb-8">
    <span class="absolute left-3 top-1/2 -translate-y-1/2 text-ink-64">
      <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <circle cx="11" cy="11" r="7" />
        <path d="m20 20-3.5-3.5" />
      </svg>
    </span>
    <input
      type="search"
      :value="props.modelValue"
      :placeholder="props.placeholder"
      aria-label="Search records by waybill number, phone number, or manifest number"
      class="w-full border-b border-ink bg-transparent py-3 pl-10 pr-4 text-p1 text-ink placeholder:text-ink-40 outline-none"
      @input="onInput"
    >
  </div>
</template>
