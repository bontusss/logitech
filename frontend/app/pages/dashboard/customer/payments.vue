<script setup lang="ts">
useHead({ title: 'Payments & Receipts · BOL' })
</script>

<template>
  <CustomerComingSoon
    eyebrow="Personal"
    title="Payments & receipts"
    description="See what you have paid, what is outstanding, and download receipts for your shipments."
    :features="[
      'View payment status for every booking',
      'Compare confirmed and estimated totals',
      'Download or print booking receipts',
      'See the payment method recorded by the provider',
    ]"
  />
</template>
