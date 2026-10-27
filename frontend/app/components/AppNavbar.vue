<script setup lang="ts">
const props = withDefaults(defineProps<{ userName: string; signingOut?: boolean }>(), { signingOut: false })
const firstName = computed(() => props.userName.trim().split(/\s+/)[0] ?? '')

defineEmits<{ logout: [] }>()
</script>

<template>
  <header class="sticky top-0 z-30 flex min-h-14 items-center justify-between gap-3 border-b border-line bg-surface/95 px-4 py-3 backdrop-blur">
    <NuxtLink to="/" class="shrink-0 text-h5 font-semibold text-ink" aria-label="Galopin home">
      Galopin
    </NuxtLink>
    <div class="flex min-w-0 items-center gap-3">
      <span class="max-w-[9rem] truncate text-label text-ink-64">{{ firstName }}</span>
      <button
        type="button"
        class="flex size-10 shrink-0 items-center justify-center rounded-lg text-ink-64 transition hover:bg-canvas hover:text-ink disabled:cursor-wait disabled:opacity-60"
        aria-label="Log out"
        title="Log out"
        :disabled="signingOut"
        @click="$emit('logout')"
      >
        <svg v-if="!signingOut" class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M10 17l5-5-5-5"/>
          <path d="M15 12H3"/>
          <path d="M12 3h6a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-6"/>
        </svg>
        <span v-else class="size-4 animate-spin rounded-full border-2 border-ink-40 border-t-blue" aria-hidden="true" />
      </button>
    </div>
  </header>
</template>
