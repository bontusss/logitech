<script setup lang="ts">
import { Download, Share, X } from 'lucide-vue-next'

const { bannerVisible: visible, install, dismissBanner: dismiss, initialize } = usePwaInstall()

const showHelp = ref(false)

onMounted(initialize)

async function installApp() {
  const result = await install()
  if (result === 'unavailable') showHelp.value = true
}
</script>

<template>
  <section
    v-if="visible"
    class="relative overflow-hidden rounded-2xl border border-ink-900/10 bg-gradient-to-br from-[#123560] via-[#0b1f44] to-[#07162f] p-4 text-white shadow-sm lg:hidden"
    aria-labelledby="install-app-title"
  >
    <div aria-hidden="true" class="pointer-events-none absolute -top-16 right-[-2rem] size-40 rounded-full bg-accent-500/25 blur-2xl" />

    <div class="relative flex items-start gap-3">
      <span class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-accent-500 text-white">
        <Download :size="18" />
      </span>
      <div class="min-w-0 flex-1">
        <h2 id="install-app-title" class="text-[13px] font-semibold tracking-[-0.01em]">
          Install the BOL app
        </h2>
        <p class="mt-1 text-[11.5px] leading-relaxed text-white/65">
          Keep your bookings and deliveries one tap away on your home screen.
        </p>
      </div>
      <button
        type="button"
        class="shrink-0 rounded-md p-1 text-white/45 transition hover:bg-white/10 hover:text-white"
        aria-label="Dismiss install banner"
        @click="dismiss"
      >
        <X :size="15" />
      </button>
    </div>

    <div class="relative mt-3 flex gap-2">
      <button
        type="button"
        class="inline-flex min-h-9 flex-1 items-center justify-center gap-1.5 rounded-lg bg-accent-500 px-3 py-2 text-[11.5px] font-semibold text-white transition hover:bg-accent-400"
        @click="installApp"
      >
        <Download :size="14" />
        Install app
      </button>
      <button
        type="button"
        class="inline-flex min-h-9 items-center justify-center rounded-lg border border-white/15 px-3 py-2 text-[11.5px] font-semibold text-white/70 transition hover:bg-white/10 hover:text-white"
        @click="dismiss"
      >
        Not now
      </button>
    </div>

    <div
      v-if="showHelp"
      class="relative mt-3 rounded-xl border border-white/10 bg-white/5 p-3 text-[11px] leading-relaxed text-white/70"
    >
      Open your browser menu and choose <strong class="font-semibold text-white">Add to Home screen</strong>.
      On iPhone, tap <Share :size="11" class="inline align-[-1px]" /> then
      <strong class="font-semibold text-white">Add to Home Screen</strong>.
    </div>
  </section>
</template>
