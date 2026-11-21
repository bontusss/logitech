<script setup lang="ts">
import { NIGERIAN_LOCATIONS, NIGERIAN_STATES } from '~/utils/nigeria'

const props = withDefaults(
  defineProps<{
    state: string
    city: string
    label: string
    tone?: 'violet' | 'sky'
  }>(),
  { tone: 'violet' },
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
  'w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-[13px] font-medium text-ink-900 transition hover:border-slate-300 focus:border-brand-500 disabled:opacity-50 dark:border-white/10 dark:bg-slate-900'
</script>

<template>
  <div class="rounded-2xl border border-ink-900/8 bg-white/70 p-3.5 dark:border-white/10 dark:bg-white/5">
    <p class="mb-2.5 flex items-center gap-1.5 text-[10px] font-bold tracking-[0.14em] uppercase">
      <span
        class="size-2 rounded-full"
        :class="tone === 'violet' ? 'bg-violet-500' : 'bg-sky-500'"
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
