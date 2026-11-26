<script setup lang="ts">
useHead({ title: 'Help & Support · BOL' })
</script>

<template>
  <CustomerComingSoon
    eyebrow="Support"
    title="Help & support"
    description="Get help with a delivery, report a problem, or find the right provider contact."
    :features="[
      'Report delayed, missing or damaged parcels',
      'Contact the provider from a shipment',
      'Read answers to common delivery questions',
      'Reach the BOL support team',
    ]"
  />
</template>
