<script setup lang="ts">
/** Labelled textarea styled like UiField, for notes and comments. */
const props = withDefaults(
  defineProps<{
    id: string
    label: string
    modelValue: string
    hint?: string
    error?: string
    rows?: number
    maxlength?: number
    disabled?: boolean
  }>(),
  { hint: '', error: '', rows: 3, maxlength: undefined, disabled: false },
)

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

function onInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLTextAreaElement).value)
}
</script>

<template>
  <div class="space-y-2">
    <label :for="props.id" :class="['block text-label', props.error ? 'text-danger' : 'text-ink-64']">
      {{ props.label }}
    </label>
    <textarea
      :id="props.id"
      :value="props.modelValue"
      :rows="props.rows"
      :maxlength="props.maxlength"
      :disabled="props.disabled"
      :aria-invalid="props.error ? 'true' : undefined"
      :aria-describedby="props.error || props.hint ? `${props.id}-note` : undefined"
      :class="[
        'w-full resize-y rounded-lg border bg-surface px-3 py-2 text-p1 text-ink transition disabled:bg-muted disabled:text-ink-40',
        'focus:outline-none focus:ring-2',
        props.error ? 'border-danger focus:border-danger focus:ring-danger/20' : 'border-line focus:border-blue focus:ring-blue/20',
      ]"
      @input="onInput"
    />
    <p v-if="props.error" :id="`${props.id}-note`" class="text-label text-danger">{{ props.error }}</p>
    <p v-else-if="props.hint" :id="`${props.id}-note`" class="text-label text-ink-64">{{ props.hint }}</p>
  </div>
</template>
