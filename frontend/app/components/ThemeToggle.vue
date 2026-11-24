<script setup lang="ts">
import { Moon, Sun } from 'lucide-vue-next'

const { isDark, toggleTheme } = useTheme()
</script>

<template>
  <button
    type="button"
    class="inline-flex size-10 shrink-0 items-center justify-center rounded-lg border border-ink-900/15 text-ink-900/70 transition hover:border-accent-500/40 hover:bg-accent-50 hover:text-ink-900"
    :aria-label="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
    :title="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
    @click="toggleTheme"
  >
    <Sun v-if="isDark" :size="17" aria-hidden="true" />
    <Moon v-else :size="17" aria-hidden="true" />
  </button>
</template>
