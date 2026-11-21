<script setup lang="ts">
import { SlidersHorizontal, X } from 'lucide-vue-next'
import type { MarketplaceFilters } from '~/types/marketplace'

const props = defineProps<{ modelValue: MarketplaceFilters }>()
const emit = defineEmits<{ 'update:modelValue': [value: MarketplaceFilters] }>()

const open = ref(false)

function patch(partial: Partial<MarketplaceFilters>) {
  emit('update:modelValue', { ...props.modelValue, ...partial })
}

const activeCount = computed(() => {
  const value = props.modelValue
  return [
    value.service !== '',
    value.confirmation !== '',
    value.maxPriceNaira !== '',
    value.pickup,
    value.doorDelivery,
  ].filter(Boolean).length
})

function clear() {
  emit('update:modelValue', {
    service: '',
    confirmation: '',
    maxPriceNaira: '',
    pickup: false,
    doorDelivery: false,
  })
}

const fieldClass =
  'w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-[12px] font-medium text-ink-900 dark:border-white/10 dark:bg-slate-900'
</script>

<template>
  <div class="site-panel rounded-2xl border">
    <button
      type="button"
      class="flex w-full items-center justify-between gap-2 px-3.5 py-3 text-left lg:cursor-default"
      @click="open = !open"
    >
      <span class="inline-flex items-center gap-1.5 text-[12px] font-semibold">
        <SlidersHorizontal :size="14" class="text-brand-600" />
        Filters
        <span
          v-if="activeCount"
          class="rounded-full bg-brand-100 px-1.5 py-0.5 text-[9.5px] font-bold text-brand-700 dark:bg-brand-400/15 dark:text-indigo-200"
        >
          {{ activeCount }}
        </span>
      </span>
      <span class="text-[10.5px] font-semibold text-ink-900/40 lg:hidden">{{ open ? 'Hide' : 'Show' }}</span>
    </button>
    <div class="space-y-3 border-t border-ink-900/8 p-3.5 dark:border-white/10" :class="open ? 'block' : 'hidden lg:block'">
      <label class="block">
        <span class="mb-1 block text-[10.5px] font-semibold text-ink-900/55">Service</span>
        <select :value="modelValue.service" :class="fieldClass" @change="patch({ service: ($event.target as HTMLSelectElement).value })">
          <option value="">All services</option>
          <option value="standard">Standard</option>
          <option value="express">Express</option>
          <option value="same_day">Same-day</option>
        </select>
      </label>
      <label class="block">
        <span class="mb-1 block text-[10.5px] font-semibold text-ink-900/55">Price type</span>
        <select :value="modelValue.confirmation" :class="fieldClass" @change="patch({ confirmation: ($event.target as HTMLSelectElement).value })">
          <option value="">Any price type</option>
          <option value="confirmed">Confirmed only</option>
          <option value="estimated">Estimated only</option>
        </select>
      </label>
      <label class="block">
        <span class="mb-1 block text-[10.5px] font-semibold text-ink-900/55">Maximum price (₦)</span>
        <input
          :value="modelValue.maxPriceNaira"
          type="number"
          min="0"
          inputmode="numeric"
          placeholder="No limit"
          :class="fieldClass"
          @input="patch({ maxPriceNaira: ($event.target as HTMLInputElement).value })"
        />
      </label>
      <label class="flex items-center gap-2 text-[11.5px] font-medium text-ink-900/65">
        <input
          type="checkbox"
          class="size-3.5 rounded border-slate-300"
          :checked="modelValue.pickup"
          @change="patch({ pickup: ($event.target as HTMLInputElement).checked })"
        />
        Offers pickup
      </label>
      <label class="flex items-center gap-2 text-[11.5px] font-medium text-ink-900/65">
        <input
          type="checkbox"
          class="size-3.5 rounded border-slate-300"
          :checked="modelValue.doorDelivery"
          @change="patch({ doorDelivery: ($event.target as HTMLInputElement).checked })"
        />
        Door delivery
      </label>
      <button
        v-if="activeCount"
        type="button"
        class="inline-flex items-center gap-1 text-[11px] font-semibold text-ink-900/50 transition hover:text-red-600"
        @click="clear"
      >
        <X :size="12" /> Clear filters
      </button>
    </div>
  </div>
</template>
