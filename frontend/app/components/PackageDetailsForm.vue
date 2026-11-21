<script setup lang="ts">
const props = defineProps<{
  weightKg: number | string
  lengthCm: number | string
  widthCm: number | string
  heightCm: number | string
  category: string
}>()

const emit = defineEmits<{
  'update:weightKg': [value: string]
  'update:lengthCm': [value: string]
  'update:widthCm': [value: string]
  'update:heightCm': [value: string]
  'update:category': [value: string]
}>()

const dimensionsOpen = ref(false)

const inputClass =
  'w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-[13px] font-medium text-ink-900 transition placeholder:text-ink-900/35 hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-slate-900'

const categories = ['parcel', 'document', 'electronics', 'fragile', 'bulk', 'other']
</script>

<template>
  <div class="space-y-3">
    <div class="grid gap-2 sm:grid-cols-2">
      <label class="block">
        <span class="mb-1 block text-[10.5px] font-semibold text-ink-900/55">Package weight (kg)</span>
        <input
          :value="weightKg"
          type="number"
          min="0.1"
          max="500"
          step="0.1"
          inputmode="decimal"
          :class="inputClass"
          placeholder="e.g. 2"
          @input="emit('update:weightKg', ($event.target as HTMLInputElement).value)"
        />
      </label>
      <label class="block">
        <span class="mb-1 block text-[10.5px] font-semibold text-ink-900/55">Package type</span>
        <select
          :value="category"
          :class="inputClass"
          @change="emit('update:category', ($event.target as HTMLSelectElement).value)"
        >
          <option v-for="item in categories" :key="item" :value="item" class="capitalize">
            {{ item }}
          </option>
        </select>
      </label>
    </div>

    <button
      type="button"
      class="inline-flex items-center gap-1.5 rounded-full border border-ink-900/10 bg-white/70 px-2.5 py-1 text-[11px] font-semibold text-ink-900/60 transition hover:border-brand-200 hover:text-brand-700 dark:border-white/10 dark:bg-white/5"
      @click="dimensionsOpen = !dimensionsOpen"
    >
      {{ dimensionsOpen ? 'Hide' : 'Add' }} package dimensions (optional)
    </button>

    <div v-if="dimensionsOpen" class="grid grid-cols-3 gap-2">
      <label class="block">
        <span class="mb-1 block text-[10.5px] font-semibold text-ink-900/55">Length (cm)</span>
        <input
          :value="lengthCm"
          type="number"
          min="1"
          max="300"
          inputmode="decimal"
          :class="inputClass"
          placeholder="40"
          @input="emit('update:lengthCm', ($event.target as HTMLInputElement).value)"
        />
      </label>
      <label class="block">
        <span class="mb-1 block text-[10.5px] font-semibold text-ink-900/55">Width (cm)</span>
        <input
          :value="widthCm"
          type="number"
          min="1"
          max="300"
          inputmode="decimal"
          :class="inputClass"
          placeholder="30"
          @input="emit('update:widthCm', ($event.target as HTMLInputElement).value)"
        />
      </label>
      <label class="block">
        <span class="mb-1 block text-[10.5px] font-semibold text-ink-900/55">Height (cm)</span>
        <input
          :value="heightCm"
          type="number"
          min="1"
          max="300"
          inputmode="decimal"
          :class="inputClass"
          placeholder="20"
          @input="emit('update:heightCm', ($event.target as HTMLInputElement).value)"
        />
      </label>
    </div>
  </div>
</template>
