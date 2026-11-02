<script setup lang="ts">
const { isDark, themeColor, init } = useTheme()

// A manifest's colours are baked in, so light and dark ship as separate files
// and the installed app launches with the splash that matches the theme.
const manifestHref = computed(() =>
  isDark.value ? '/manifest.dark.webmanifest' : '/manifest.webmanifest',
)

useHead({
  titleTemplate: (title) => (title ? `${title} · Logitech` : 'Logitech'),
  htmlAttrs: {
    class: computed(() => (isDark.value ? 'dark' : '')),
  },
  link: [{ rel: 'manifest', href: manifestHref }],
  meta: [
    {
      name: 'theme-color',
      content: computed(() => themeColor.value),
    },
  ],
})

onMounted(init)
</script>

<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
