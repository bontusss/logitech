<script setup lang="ts">
useHead({ title: 'Find Providers · BOL' })
</script>

<template>
  <CustomerComingSoon
    eyebrow="Shipments"
    title="Find providers"
    description="Search and compare logistics providers by route, price and delivery method."
    :features="[
      'Search by pickup and destination zone',
      'Compare live rates and delivery estimates',
      'Filter by service, price and delivery method',
      'View provider profiles and reviews',
      'Save favourite providers for later',
    ]"
  />
</template>
