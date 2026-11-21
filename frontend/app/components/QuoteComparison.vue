<script setup lang="ts">
import type { MarketplaceQuote } from '~/types/marketplace'

defineProps<{
  quotes: MarketplaceQuote[]
  busyReference?: string
  selectedReference?: string
}>()

defineEmits<{ select: [quote: MarketplaceQuote] }>()
</script>

<template>
  <div class="space-y-2.5">
    <ProviderResultCard
      v-for="quote in quotes"
      :key="quote.quote_reference"
      :quote="quote"
      :busy="busyReference === quote.quote_reference"
      :selected="selectedReference === quote.quote_reference"
      @select="$emit('select', quote)"
    />
  </div>
</template>
