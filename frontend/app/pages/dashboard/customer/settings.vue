<script setup lang="ts">
useHead({ title: 'Settings · BOL' })
</script>

<template>
  <CustomerComingSoon
    eyebrow="Account"
    title="Settings"
    description="Manage your personal details and how providers and BOL reach you."
    :features="[
      'Update your name and phone number',
      'Verify your email address',
      'Choose notification channels',
      'Manage account security',
    ]"
  />
</template>
