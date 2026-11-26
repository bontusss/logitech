<script setup lang="ts">
useHead({ title: 'Favourite Providers · BOL' })
</script>

<template>
  <CustomerComingSoon
    eyebrow="Personal"
    title="Favourite providers"
    description="Keep trusted logistics providers close so your next shipment takes less time to arrange."
    :features="[
      'Save providers from search results',
      'See provider ratings and route coverage',
      'Open a provider booking page in one tap',
      'Remove providers you no longer use',
    ]"
  />
</template>
