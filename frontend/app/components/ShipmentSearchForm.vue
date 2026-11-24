<script setup lang="ts">
import { ArrowRight, Loader2, MapPin, Search } from 'lucide-vue-next'
import type { MarketplaceSearchPayload } from '~/types/marketplace'

const props = withDefaults(
  defineProps<{
    loading?: boolean
    originState?: string
    originCity?: string
    destinationState?: string
    destinationCity?: string
    weightKg?: number | string
    compact?: boolean
  }>(),
  { loading: false, compact: false, weightKg: 2 },
)

const emit = defineEmits<{ submit: [payload: MarketplaceSearchPayload] }>()

const originState = ref(props.originState ?? '')
const originCity = ref(props.originCity ?? '')
const destinationState = ref(props.destinationState ?? '')
const destinationCity = ref(props.destinationCity ?? '')
const weightKg = ref<string | number>(props.weightKg)
const lengthCm = ref<string | number>('')
const widthCm = ref<string | number>('')
const heightCm = ref<string | number>('')
const category = ref('parcel')
const speed = ref<'standard' | 'express' | 'same_day' | ''>('')
const pickupRequired = ref(false)
const deliveryRequired = ref(false)
const error = ref('')

const speeds = [
  { value: '', label: 'Any speed' },
  { value: 'standard', label: 'Standard' },
  { value: 'express', label: 'Express' },
  { value: 'same_day', label: 'Same-day' },
] as const

function submit() {
  error.value = ''
  const weight = Number(weightKg.value)
  if (!originState.value || !originCity.value || !destinationState.value || !destinationCity.value) {
    error.value = 'Choose the origin and destination states and cities.'
    return
  }
  if (!Number.isFinite(weight) || weight <= 0) {
    error.value = 'Enter a package weight greater than zero.'
    return
  }
  const dims = [lengthCm.value, widthCm.value, heightCm.value].map((value) => Number(value))
  const hasSomeDimensions = dims.some((value) => value > 0)
  if (hasSomeDimensions && !dims.every((value) => value > 0)) {
    error.value = 'Enter length, width and height together.'
    return
  }
  emit('submit', {
    origin_state: originState.value,
    origin_city: originCity.value,
    destination_state: destinationState.value,
    destination_city: destinationCity.value,
    weight_kg: weight,
    length_cm: hasSomeDimensions ? dims[0] : undefined,
    width_cm: hasSomeDimensions ? dims[1] : undefined,
    height_cm: hasSomeDimensions ? dims[2] : undefined,
    category: category.value,
    speed: speed.value,
    pickup_required: pickupRequired.value,
    delivery_required: deliveryRequired.value,
  })
}
</script>

<template>
  <form class="site-panel rounded-3xl border p-5 sm:p-6" @submit.prevent="submit">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <p class="inline-flex items-center gap-1.5 text-[10px] font-bold tracking-[0.18em] text-brand-600 uppercase">
          <Search :size="12" /> Shipment search
        </p>
        <h2 class="mt-1.5 text-[19px] font-semibold tracking-[-0.03em]">
          Where is it going?
        </h2>
      </div>
      <span class="flex size-10 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-100 to-brand-50 text-brand-700 dark:from-brand-400/15 dark:to-brand-400/10 dark:text-brand-200">
        <MapPin :size="18" />
      </span>
    </div>

    <div class="mt-4 grid gap-2.5 sm:grid-cols-2">
      <LocationSelector
        v-model:state="originState"
        v-model:city="originCity"
        label="Sending from"
        tone="brand"
      />
      <LocationSelector
        v-model:state="destinationState"
        v-model:city="destinationCity"
        label="Sending to"
        tone="accent"
      />
    </div>

    <div class="mt-3">
      <PackageDetailsForm
        v-model:weightKg="weightKg"
        v-model:lengthCm="lengthCm"
        v-model:widthCm="widthCm"
        v-model:heightCm="heightCm"
        v-model:category="category"
      />
    </div>

    <div class="mt-4">
      <p class="mb-1.5 text-[10.5px] font-semibold text-ink-900/55">Preferred speed</p>
      <div class="flex flex-wrap gap-1.5">
        <button
          v-for="option in speeds"
          :key="option.value"
          type="button"
          class="rounded-full border px-3 py-1.5 text-[11.5px] font-semibold transition"
          :class="
            speed === option.value
              ? 'border-brand-400 bg-brand-50 text-brand-700 dark:border-brand-400/40 dark:bg-brand-400/10 dark:text-brand-200'
              : 'border-slate-200 bg-white/70 text-ink-900/60 hover:border-slate-300 dark:border-white/10 dark:bg-white/5'
          "
          @click="speed = option.value"
        >
          {{ option.label }}
        </button>
      </div>
    </div>

    <div class="mt-4 grid gap-2 sm:grid-cols-2">
      <label class="flex items-center gap-2 rounded-xl border border-ink-900/8 bg-white/60 px-3 py-2.5 text-[11.5px] font-medium text-ink-900/65 dark:border-white/10 dark:bg-white/5">
        <input v-model="pickupRequired" type="checkbox" class="size-3.5 rounded border-slate-300" />
        Pickup from my address
      </label>
      <label class="flex items-center gap-2 rounded-xl border border-ink-900/8 bg-white/60 px-3 py-2.5 text-[11.5px] font-medium text-ink-900/65 dark:border-white/10 dark:bg-white/5">
        <input v-model="deliveryRequired" type="checkbox" class="size-3.5 rounded border-slate-300" />
        Deliver to the door
      </label>
    </div>

    <p v-if="error" role="alert" class="mt-3 rounded-xl border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] font-medium text-red-700">
      {{ error }}
    </p>

    <button
      type="submit"
      :disabled="loading"
      class="mt-4 flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-brand-600 to-brand-600 px-4 py-3.5 text-[14px] font-semibold text-white shadow-[0_14px_28px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-60"
    >
      <Loader2 v-if="loading" :size="16" class="animate-spin" />
      <template v-else>
        Compare delivery options
        <ArrowRight :size="15" />
      </template>
    </button>
    <p class="mt-2 text-center text-[10.5px] text-ink-900/45">
      No account needed. Prices come directly from each provider's published rates.
    </p>
  </form>
</template>
