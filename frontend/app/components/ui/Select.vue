<script setup lang="ts">
/** Labelled dropdown select styled like UiField. */
const props = withDefaults(
  defineProps<{
    id: string
    label: string
    modelValue: string
    options: { value: string; label: string }[]
    hint?: string
    error?: string
    disabled?: boolean
    hideLabel?: boolean
  }>(),
  { hint: '', error: '', disabled: false, hideLabel: false },
)

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

function onInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLSelectElement).value)
}
</script>

<template>
  <div class="space-y-2">
    <label
      :for="props.id"
      :class="['block text-label', props.error ? 'text-danger' : 'text-ink-64', props.hideLabel && 'sr-only']"
    >
      {{ props.label }}
    </label>

    <div class="relative">
      <select
        :id="props.id"
        :value="props.modelValue"
        :disabled="props.disabled"
        :aria-invalid="props.error ? 'true' : undefined"
        :aria-describedby="props.error || props.hint ? `${props.id}-note` : undefined"
        class="h-10 w-full appearance-none rounded-lg border bg-surface px-3 pr-8 text-p1 font-normal text-ink transition duration-150 disabled:bg-muted disabled:text-ink-40 cursor-pointer"
        :class="[
          props.error
            ? 'border-danger focus:border-danger focus:ring-danger/20'
            : 'border-line focus:border-blue focus:ring-blue/20',
        ]"
        @input="onInput"
      >
        <option
          v-for="option in props.options"
          :key="option.value"
          :value="option.value"
        >
          {{ option.label }}
        </option>
      </select>

      <span class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-ink-40">
        <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="m6 9 6 6 6-6"/>
        </svg>
      </span>
    </div>

    <p v-if="props.error" :id="`${props.id}-note`" class="text-label text-danger">
      {{ props.error }}
    </p>
    <p v-else-if="props.hint" :id="`${props.id}-note`" class="text-label text-ink-64">
      {{ props.hint }}
    </p>
  </div>
</template>
