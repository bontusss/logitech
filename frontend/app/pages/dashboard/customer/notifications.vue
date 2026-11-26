<script setup lang="ts">
useHead({ title: 'Notifications · BOL' })
</script>

<template>
  <CustomerComingSoon
    eyebrow="Support"
    title="Notifications"
    description="Follow pickup, transit and delivery updates without checking every shipment."
    :features="[
      'Pickup and delivery updates',
      'Provider request and cancellation alerts',
      'Issue report updates',
      'Email and WhatsApp notification preferences',
    ]"
  />
</template>
