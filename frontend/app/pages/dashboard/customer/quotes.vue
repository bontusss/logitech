<script setup lang="ts">
useHead({ title: 'Saved Quotes · BOL' })
</script>

<template>
  <CustomerComingSoon
    eyebrow="Shipments"
    title="Saved quotes"
    description="Keep provider quotes, compare confirmed and estimated prices, and finish an incomplete booking."
    :features="[
      'View saved quotes and quote history',
      'Compare confirmed and estimated prices',
      'Resume incomplete bookings',
      'Open the provider booking page directly',
    ]"
  />
</template>
