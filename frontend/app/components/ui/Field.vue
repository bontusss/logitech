<script setup lang="ts">
/** Labelled text field: 40px high, 8px corners, optional hint or error. */
const props = withDefaults(
  defineProps<{
    id: string
    label: string
    modelValue: string
    type?: 'text' | 'password' | 'tel' | 'email'
    name?: string
    hint?: string
    error?: string
    autocomplete?: string
    inputmode?: 'text' | 'numeric' | 'decimal' | 'tel' | 'email'
    maxlength?: number
    autofocus?: boolean
    disabled?: boolean
  }>(),
  { type: 'text', name: undefined, hint: '', error: '', autocomplete: 'off', maxlength: undefined, autofocus: false, disabled: false },
)

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

function onInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}
</script>

<template>
  <div class="space-y-2">
    <label
      :for="props.id"
      :class="['block text-label', props.error ? 'text-danger' : 'text-ink-64']"
    >
      {{ props.label }}
    </label>

    <input
      :id="props.id"
      :name="props.name"
      :type="props.type"
      :value="props.modelValue"
      :autocomplete="props.autocomplete"
      :inputmode="props.inputmode"
      :maxlength="props.maxlength"
      :autofocus="props.autofocus"
      :disabled="props.disabled"
      :aria-invalid="props.error ? 'true' : undefined"
      :aria-describedby="props.error || props.hint ? `${props.id}-note` : undefined"
      :class="[
        'h-10 w-full rounded-lg border bg-surface px-3 text-p1 font-normal text-ink transition duration-150 disabled:bg-muted disabled:text-ink-40',
        'focus:outline-none focus:ring-2',
        props.error
          ? 'border-danger focus:border-danger focus:ring-danger/20'
          : 'border-line focus:border-blue focus:ring-blue/20',
      ]"
      @input="onInput"
    />

    <p v-if="props.error" :id="`${props.id}-note`" class="text-label text-danger">
      {{ props.error }}
    </p>
    <p v-else-if="props.hint" :id="`${props.id}-note`" class="text-label text-ink-64">
      {{ props.hint }}
    </p>
  </div>
</template>
