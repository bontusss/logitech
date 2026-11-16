<script setup lang="ts">
import { ArrowLeftRight, ChevronDown, Info, MapPin, Search } from 'lucide-vue-next'

const cities = ['Aba', 'Abuja', 'Enugu', 'Lagos', 'Onitsha', 'Port Harcourt', 'Umuahia']
const categories = ['Documents', 'Parcels', 'Electronics', 'Furniture', 'Bulk cargo']

const form = reactive({
  pickup: 'Aba',
  destination: 'Lagos',
  weight: 5,
  category: 'Parcels',
})

const submitted = ref(false)

const fieldClass =
  'w-full appearance-none rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-sm font-medium text-ink-900 transition hover:border-slate-400 focus:border-brand-500'
const labelClass =
  'mb-1.5 flex items-center gap-1.5 text-[10.5px] font-semibold tracking-[0.12em] text-ink-900/50 uppercase'

function swapCities() {
  const pickup = form.pickup
  form.pickup = form.destination
  form.destination = pickup
}

function submitSearch() {
  submitted.value = true
}
</script>

<template>
  <div
    id="rate-search"
    class="rounded-2xl border border-ink-900/10 bg-white p-5 shadow-[0_24px_60px_-40px_rgba(10,18,32,0.45)] sm:p-6"
  >
    <div class="flex items-start justify-between gap-3">
      <div>
        <h2 class="text-[15px] font-semibold tracking-[-0.01em]">Compare shipping rates</h2>
        <p class="mt-1 text-[12.5px] text-ink-900/50">Intercity routes across Nigeria</p>
      </div>
      <span
        class="rounded-md border border-brand-200 bg-brand-50 px-2 py-1 text-[10px] font-semibold tracking-[0.12em] text-brand-700 uppercase"
      >
        Preview
      </span>
    </div>

    <form class="mt-5 space-y-4" @submit.prevent="submitSearch">
      <div class="grid gap-2 sm:grid-cols-[1fr_auto_1fr] sm:items-end">
        <label class="block">
          <span :class="labelClass"><MapPin :size="12" /> Pickup city</span>
          <div class="relative">
            <select v-model="form.pickup" :class="fieldClass">
              <option v-for="city in cities" :key="city" :value="city">{{ city }}</option>
            </select>
            <ChevronDown
              :size="15"
              class="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-ink-900/40"
            />
          </div>
        </label>

        <button
          type="button"
          class="mx-auto flex size-9 items-center justify-center rounded-lg border border-slate-200 text-ink-900/50 transition hover:border-slate-300 hover:text-ink-900"
          aria-label="Swap pickup and destination cities"
          @click="swapCities"
        >
          <ArrowLeftRight :size="15" />
        </button>

        <label class="block">
          <span :class="labelClass"><MapPin :size="12" /> Destination</span>
          <div class="relative">
            <select v-model="form.destination" :class="fieldClass">
              <option v-for="city in cities" :key="city" :value="city">{{ city }}</option>
            </select>
            <ChevronDown
              :size="15"
              class="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-ink-900/40"
            />
          </div>
        </label>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <label class="block">
          <span :class="labelClass">Weight (kg)</span>
          <input
            v-model.number="form.weight"
            type="number"
            min="0.5"
            step="0.5"
            inputmode="decimal"
            class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-sm font-medium text-ink-900 transition hover:border-slate-400 focus:border-brand-500"
          />
        </label>
        <label class="block">
          <span :class="labelClass">Category</span>
          <div class="relative">
            <select v-model="form.category" :class="fieldClass">
              <option v-for="category in categories" :key="category" :value="category">
                {{ category }}
              </option>
            </select>
            <ChevronDown
              :size="15"
              class="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-ink-900/40"
            />
          </div>
        </label>
      </div>

      <button
        type="submit"
        class="flex w-full items-center justify-center gap-2 rounded-lg bg-brand-500 px-4 py-3 text-sm font-semibold text-white shadow-[0_10px_28px_-14px_rgba(43,75,255,0.85)] transition hover:bg-brand-600"
      >
        <Search :size="16" />
        Compare Rates
      </button>
    </form>

    <p class="mt-3 flex items-start gap-1.5 text-[11.5px] leading-relaxed text-ink-900/45">
      <Info :size="13" class="mt-0.5 shrink-0" />
      Illustrative UI with sample routes. Live provider rates will arrive with the marketplace API.
    </p>

    <div
      v-if="submitted"
      role="status"
      class="mt-3 rounded-lg border border-brand-200 bg-brand-50 px-3 py-2.5 text-[12px] leading-relaxed text-brand-700"
    >
      This is a UI preview and does not request live quotes yet. Once the marketplace API is
      connected, matching providers for {{ form.pickup }} → {{ form.destination }} will appear here.
    </div>
  </div>
</template>
