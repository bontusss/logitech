<script setup lang="ts">
import { CalendarClock } from 'lucide-vue-next'
import type { MarketplaceQuote } from '~/types/marketplace'

withDefaults(defineProps<{ quote: MarketplaceQuote; detailed?: boolean }>(), { detailed: false })
</script>

<template>
  <div class="inline-flex items-start gap-2">
    <CalendarClock :size="14" class="mt-0.5 shrink-0 text-brand-600" />
    <div>
      <p class="text-[12px] font-semibold">
        {{ quote.estimated_delivery || `About ${quote.estimated_duration_days} day(s)` }}
      </p>
      <p v-if="detailed && quote.departure_schedule" class="text-[10.5px] text-ink-900/45">
        {{ quote.departure_schedule }}
      </p>
    </div>
  </div>
</template>
