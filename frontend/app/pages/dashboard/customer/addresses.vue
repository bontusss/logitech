<script setup lang="ts">
useHead({ title: 'Address Book · BOL' })
</script>

<template>
  <CustomerComingSoon
    eyebrow="Personal"
    title="Address book"
    description="Save the pickup and delivery places you use most, then select them during booking."
    :features="[
      'Save home, office and warehouse addresses',
      'Store contact details and landmarks',
      'Set a default pickup and delivery address',
      'Select a saved address while booking',
    ]"
  />
</template>
