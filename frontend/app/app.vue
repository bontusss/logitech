<script setup lang="ts">
useHead({
  titleTemplate: (title) => (title ? `${title} · Logitech` : 'Logitech'),
})
</script>

<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
