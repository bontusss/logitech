<script setup lang="ts">
useHead({
  titleTemplate: (title) => (title ? `${title} · BOL` : 'based on logistics (BOL)'),
})
</script>

<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
