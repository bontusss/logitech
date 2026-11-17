<script setup lang="ts">
import { ArrowLeftRight, ArrowRight, MapPin, PackageCheck, Search } from 'lucide-vue-next'

const cities = ['Aba', 'Abuja', 'Enugu', 'Lagos', 'Onitsha', 'Port Harcourt', 'Umuahia']
const categories = ['Document', 'Parcel', 'Electronics', 'Fragile goods', 'Bulk cargo']
const dimensions = [
  { key: 'length', label: 'Length' },
  { key: 'width', label: 'Width' },
  { key: 'height', label: 'Height' },
] as const

const form = reactive({
  origin: 'Aba',
  destination: 'Lagos',
  weight: 2,
  length: 20,
  width: 15,
  height: 10,
  category: 'Parcel',
})

const searched = ref(false)

const fieldClass =
  'w-full rounded-xl border border-slate-200 bg-white px-3.5 py-3 text-[13px] font-medium text-ink-900 transition placeholder:text-ink-900/35 hover:border-slate-300 focus:border-brand-500'
const labelClass = 'mb-1.5 block text-[11px] font-semibold text-ink-900/55'

function swapCities() {
  const origin = form.origin
  form.origin = form.destination
  form.destination = origin
}

function submitSearch() {
  searched.value = true
}
</script>

<template>
  <div
    id="rate-search"
    class="w-full max-w-[540px] rounded-[1.4rem] border border-white/80 bg-white p-5 shadow-[0_35px_90px_-45px_rgba(27,40,90,0.42)] sm:p-7"
  >
    <div class="flex items-start justify-between gap-4">
      <div>
        <p class="text-[10px] font-bold tracking-[0.18em] text-brand-600 uppercase">Start here</p>
        <h2 class="mt-1.5 text-[19px] font-semibold tracking-[-0.03em]">Where are you sending?</h2>
        <p class="mt-1 text-[12.5px] text-ink-900/50">Enter your shipment to check delivery options.</p>
      </div>
      <div class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-brand-50 text-brand-600">
        <PackageCheck :size="19" />
      </div>
    </div>

    <form class="mt-5 space-y-3.5" @submit.prevent="submitSearch">
      <div class="grid grid-cols-[1fr_auto_1fr] items-end gap-2">
        <label class="min-w-0">
          <span :class="labelClass">Sending from</span>
          <input
            v-model="form.origin"
            :class="fieldClass"
            list="bol-cities"
            name="origin"
            autocomplete="address-level2"
            placeholder="Pickup city"
            required
          />
        </label>
        <button
          type="button"
          class="mb-1 flex size-9 shrink-0 items-center justify-center rounded-xl border border-slate-200 text-ink-900/50 transition hover:border-brand-200 hover:text-brand-600"
          aria-label="Swap origin and destination"
          @click="swapCities"
        >
          <ArrowLeftRight :size="15" />
        </button>
        <label class="min-w-0">
          <span :class="labelClass">Sending to</span>
          <input
            v-model="form.destination"
            :class="fieldClass"
            list="bol-cities"
            name="destination"
            autocomplete="address-level2"
            placeholder="Destination city"
            required
          />
        </label>
        <datalist id="bol-cities">
          <option v-for="city in cities" :key="city" :value="city" />
        </datalist>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <label>
          <span :class="labelClass">Package weight</span>
          <span class="relative block">
            <input
              v-model.number="form.weight"
              :class="fieldClass"
              type="number"
              name="weight"
              min="0.1"
              max="1000"
              step="0.1"
              inputmode="decimal"
              required
            />
            <span class="pointer-events-none absolute top-1/2 right-3.5 -translate-y-1/2 text-[11px] text-ink-900/40">kg</span>
          </span>
        </label>
        <label>
          <span :class="labelClass">Package type</span>
          <select v-model="form.category" :class="fieldClass" name="category">
            <option v-for="category in categories" :key="category" :value="category">{{ category }}</option>
          </select>
        </label>
      </div>

      <details class="group">
        <summary class="cursor-pointer list-none text-[11.5px] font-semibold text-ink-900/55 marker:hidden [&::-webkit-details-marker]:hidden">
          Add package dimensions <span class="font-normal text-ink-900/35">(optional)</span>
        </summary>
        <div class="mt-3 grid grid-cols-3 gap-2">
          <label v-for="dimension in dimensions" :key="dimension.key">
            <span :class="labelClass">{{ dimension.label }} (cm)</span>
            <input
              v-model.number="form[dimension.key]"
              :class="fieldClass"
              type="number"
              min="1"
              step="1"
              :name="dimension.key"
            />
          </label>
        </div>
      </details>

      <button
        type="submit"
        class="flex w-full items-center justify-center gap-2 rounded-xl bg-brand-500 px-4 py-3.5 text-[14px] font-semibold text-white shadow-[0_14px_28px_-16px_rgba(43,75,255,0.9)] transition hover:bg-brand-600"
      >
        <Search :size="16" />
        Compare delivery options
        <ArrowRight :size="15" class="ml-auto" />
      </button>
    </form>

    <p class="mt-3 text-center text-[10.5px] leading-relaxed text-ink-900/40">
      No commitment. Review providers and service options before you choose.
    </p>

    <div
      v-if="searched"
      role="status"
      aria-live="polite"
      class="mt-4 rounded-xl border border-amber-200 bg-amber-50 p-3.5 text-[12px] leading-relaxed text-amber-900"
    >
      <p class="font-semibold">{{ form.origin }} → {{ form.destination }} is ready to search.</p>
      <p class="mt-1 text-amber-900/75">
        Live provider rates and booking are not connected yet. We won't show you made-up prices.
        The comparison below is an example of how options will appear.
      </p>
    </div>
  </div>
</template>
