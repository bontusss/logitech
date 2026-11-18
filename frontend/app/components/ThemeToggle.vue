<script setup lang="ts">
import { Moon, Sun } from 'lucide-vue-next'

const { isDark, toggleTheme } = useTheme()
</script>

<template>
  <button
    type="button"
    class="inline-flex size-10 shrink-0 items-center justify-center rounded-lg border border-ink-900/20 text-ink-900 transition hover:bg-ink-900/5"
    :aria-label="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
    :title="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
    @click="toggleTheme"
  >
    <Sun v-if="isDark" :size="17" aria-hidden="true" />
    <Moon v-else :size="17" aria-hidden="true" />
  </button>
</template>
