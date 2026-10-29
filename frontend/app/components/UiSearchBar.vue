<script setup lang="ts">
/** Search field with a leading icon and a clear button. */
const props = withDefaults(
  defineProps<{ modelValue?: string; placeholder?: string; label?: string }>(),
  {
    modelValue: '',
    placeholder: 'Search waybill number, name or phone',
    label: 'Search',
  },
)

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const input = ref<HTMLInputElement | null>(null)

function onInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}

function clear() {
  emit('update:modelValue', '')
  input.value?.focus()
}
</script>

<template>
  <div class="relative">
    <span class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-ink-40">
      <UiIcon name="search" :size="17" />
    </span>
    <input
      ref="input"
      type="search"
      :value="props.modelValue"
      :placeholder="props.placeholder"
      :aria-label="props.label"
      class="h-10 w-full rounded-lg border border-line bg-surface pr-9 pl-9 text-p1 text-ink transition focus:border-blue focus:ring-2 focus:ring-blue/20 focus:outline-none [&::-webkit-search-cancel-button]:appearance-none"
      @input="onInput"
    >
    <button
      v-if="props.modelValue"
      type="button"
      class="absolute top-1/2 right-2 flex size-6 -translate-y-1/2 items-center justify-center rounded-md text-ink-40 transition hover:bg-canvas hover:text-ink"
      aria-label="Clear search"
      @click="clear"
    >
      <UiIcon name="close" :size="14" />
    </button>
  </div>
</template>
