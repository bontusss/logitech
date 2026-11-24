<script setup lang="ts">
import { RefreshCw, WifiOff } from 'lucide-vue-next'

useSeoMeta({ title: 'You are offline', robots: 'noindex' })

const retrying = ref(false)

function retry() {
  retrying.value = true
  window.location.reload()
}
</script>

<template>
  <section class="mx-auto flex min-h-[62vh] max-w-lg flex-col items-center justify-center px-5 py-16 text-center">
    <IconTile tone="accent" size="lg" rounded="2xl">
      <WifiOff :size="20" />
    </IconTile>
    <h1 class="mt-5 text-[clamp(1.6rem,4vw,2.1rem)] leading-tight font-semibold tracking-[-0.035em]">
      You’re offline
    </h1>
    <p class="mt-3 text-[14px] leading-[1.7] text-ink-900/60">
      BOL needs a connection to compare live rates, open bookings and show your dashboard.
      Reconnect and try again — anything you already had open is safe.
    </p>
    <button
      type="button"
      :disabled="retrying"
      class="mt-6 inline-flex items-center gap-2 rounded-xl bg-accent-500 px-4 py-2.5 text-sm font-semibold text-ink-950 shadow-[0_12px_24px_-16px_rgba(255,122,0,0.75)] transition hover:bg-accent-400 disabled:opacity-60"
      @click="retry"
    >
      <RefreshCw :size="15" :class="retrying ? 'animate-spin' : ''" />
      {{ retrying ? 'Reconnecting…' : 'Try again' }}
    </button>
    <p class="mt-4 text-[11.5px] text-ink-900/40">
      Public pages you visited before are still available through the installed app.
    </p>
  </section>
</template>
