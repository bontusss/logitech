<script setup lang="ts">
import { NIGERIAN_LOCATIONS, NIGERIAN_STATES } from '~/utils/nigeria'

const props = withDefaults(
  defineProps<{
    state: string
    city: string
    required?: boolean
    disabled?: boolean
  }>(),
  { required: false, disabled: false },
)

const emit = defineEmits<{
  'update:state': [value: string]
  'update:city': [value: string]
}>()

const cityOptions = computed(() => {
  const cities = NIGERIAN_LOCATIONS[props.state] ?? []
  return props.city && !cities.includes(props.city) ? [props.city, ...cities] : cities
})

function updateState(event: Event) {
  const state = (event.target as HTMLSelectElement).value
  emit('update:state', state)
  emit('update:city', '')
}
</script>

<template>
  <div class="grid gap-3 sm:grid-cols-2">
    <label class="block">
      <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">State</span>
      <select
        :value="state"
        :required="required"
        :disabled="disabled"
        class="w-full rounded-xl border border-ink-900/15 bg-canvas px-3.5 py-2.5 text-sm text-ink-900 transition hover:border-ink-900/30 focus:border-accent-500 focus:bg-surface focus:outline-none disabled:bg-canvas/60 disabled:text-ink-900/45"
        @change="updateState"
      >
        <option value="">Select a state</option>
        <option v-for="item in NIGERIAN_STATES" :key="item" :value="item">{{ item }}</option>
      </select>
    </label>
    <label class="block">
      <span class="mb-1.5 block text-[11.5px] font-semibold text-ink-900/75">City / LGA</span>
      <select
        :value="city"
        :required="required"
        :disabled="disabled || !state"
        class="w-full rounded-xl border border-ink-900/15 bg-canvas px-3.5 py-2.5 text-sm text-ink-900 transition hover:border-ink-900/30 focus:border-accent-500 focus:bg-surface focus:outline-none disabled:bg-canvas/60 disabled:text-ink-900/45"
        @change="emit('update:city', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">Select a city or LGA</option>
        <option v-for="item in cityOptions" :key="item" :value="item">{{ item }}</option>
      </select>
    </label>
  </div>
</template>
