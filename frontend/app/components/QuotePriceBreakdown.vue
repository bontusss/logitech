<script setup lang="ts">
import type { MarketplaceQuote } from '~/types/marketplace'

const props = defineProps<{ quote: MarketplaceQuote }>()

const currency = computed(() => props.quote.currency || 'NGN')
const rows = computed(() => props.quote.breakdown ?? [])
</script>

<template>
  <div class="rounded-2xl border border-ink-900/8 bg-white/60 p-4 dark:border-white/10 dark:bg-white/5">
    <p class="text-[10px] font-bold tracking-[0.16em] text-ink-900/45 uppercase">Price breakdown</p>
    <dl class="mt-2.5 space-y-2">
      <div v-for="row in rows" :key="row.name" class="flex items-start justify-between gap-3">
        <div>
          <dt class="text-[12px] font-medium text-ink-900/75">{{ row.name }}</dt>
          <p v-if="row.description" class="text-[10.5px] text-ink-900/45">{{ row.description }}</p>
        </div>
        <dd class="text-[12px] font-semibold tabular-nums">{{ formatMinor(row.amount_minor, currency) }}</dd>
      </div>
      <div v-if="quote.additional_charges_minor > 0" class="flex items-center justify-between gap-3">
        <dt class="text-[12px] font-medium text-ink-900/75">Additional charges</dt>
        <dd class="text-[12px] font-semibold tabular-nums">
          {{ formatMinor(quote.additional_charges_minor, currency) }}
        </dd>
      </div>
      <div v-if="quote.discount_minor > 0" class="flex items-center justify-between gap-3">
        <dt class="text-[12px] font-medium text-emerald-700">Discount</dt>
        <dd class="text-[12px] font-semibold tabular-nums text-emerald-700">
          -{{ formatMinor(quote.discount_minor, currency) }}
        </dd>
      </div>
    </dl>
    <div class="mt-3 flex items-center justify-between border-t border-ink-900/8 pt-3 dark:border-white/10">
      <span class="text-[12px] font-semibold">Total</span>
      <span class="text-[18px] font-bold tracking-[-0.03em] tabular-nums">
        {{ formatMinor(quote.total_price_minor, currency) }}
      </span>
    </div>
    <p class="mt-1.5 text-[10.5px] leading-relaxed text-ink-900/45">
      Chargeable weight {{ quote.chargeable_weight_kg }} kg over {{ quote.distance_km }} km.
    </p>
  </div>
</template>
