<script setup lang="ts">
import { normalizeNigerianLocalPhone } from '~/utils/nigeria'

const props = withDefaults(
  defineProps<{
    modelValue: string
    required?: boolean
    disabled?: boolean
    autocomplete?: string
    placeholder?: string
  }>(),
  {
    required: false,
    disabled: false,
    autocomplete: 'tel',
    placeholder: '801 234 5678',
  },
)

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const localNumber = computed(() => normalizeNigerianLocalPhone(props.modelValue))

function updatePhone(event: Event) {
  const local = normalizeNigerianLocalPhone((event.target as HTMLInputElement).value)
  emit('update:modelValue', local ? `+234${local}` : '')
}
</script>

<template>
  <div class="flex w-full overflow-hidden rounded-xl border border-ink-900/15 bg-surface transition hover:border-ink-900/30 focus-within:border-accent-500">
    <span class="flex items-center border-r border-ink-900/15 bg-canvas px-3 text-sm font-medium text-ink-900/70">
      +234
    </span>
    <input
      :value="localNumber"
      type="tel"
      inputmode="numeric"
      :autocomplete="autocomplete"
      :required="required"
      :disabled="disabled"
      maxlength="10"
      pattern="[789][0-9]{9}"
      :placeholder="placeholder"
      aria-label="Nigerian phone number without country code"
      class="min-w-0 flex-1 bg-transparent px-3 py-2.5 text-sm text-ink-900 outline-none placeholder:text-ink-900/35 disabled:bg-canvas disabled:text-ink-900/45"
      @input="updatePhone"
    />
  </div>
</template>
