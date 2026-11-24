<script setup lang="ts">
import { NIGERIAN_LOCATIONS, NIGERIAN_STATES } from '~/utils/nigeria'

const props = withDefaults(
  defineProps<{
    state: string
    city: string
    label: string
    tone?: 'brand' | 'accent'
  }>(),
  { tone: 'brand' },
)

const emit = defineEmits<{
  'update:state': [value: string]
  'update:city': [value: string]
}>()

const cities = computed(() => NIGERIAN_LOCATIONS[props.state] ?? [])

watch(
  () => props.state,
  () => {
    if (props.city && !cities.value.includes(props.city)) emit('update:city', '')
  },
)

const selectClass =
  'w-full rounded-xl border border-ink-900/12 bg-canvas px-3 py-2.5 text-[13px] font-medium text-ink-900 transition hover:border-ink-900/25 focus:border-accent-500 focus:bg-surface focus:outline-none disabled:opacity-50'
</script>

<template>
  <div class="rounded-2xl border border-ink-900/8 bg-surface/70 p-3.5">
    <p class="mb-2.5 flex items-center gap-1.5 text-[10px] font-bold tracking-[0.14em] uppercase">
      <span
        class="size-2 rounded-full"
        :class="tone === 'accent' ? 'bg-accent-500' : 'bg-brand-600'"
      />
      <span class="text-ink-900/60">{{ label }}</span>
    </p>
    <div class="grid grid-cols-2 gap-2">
      <label class="block">
        <span class="mb-1 block text-[10.5px] font-semibold text-ink-900/55">State</span>
        <select
          :value="state"
          :class="selectClass"
          aria-label="State"
          @change="emit('update:state', ($event.target as HTMLSelectElement).value)"
        >
          <option value="" disabled>Select state</option>
          <option v-for="item in NIGERIAN_STATES" :key="item" :value="item">{{ item }}</option>
        </select>
      </label>
      <label class="block">
        <span class="mb-1 block text-[10.5px] font-semibold text-ink-900/55">City</span>
        <select
          :value="city"
          :class="selectClass"
          :disabled="!state"
          aria-label="City"
          @change="emit('update:city', ($event.target as HTMLSelectElement).value)"
        >
          <option value="" disabled>{{ state ? 'Select city' : 'Pick a state' }}</option>
          <option v-for="item in cities" :key="item" :value="item">{{ item }}</option>
        </select>
      </label>
    </div>
  </div>
</template>
