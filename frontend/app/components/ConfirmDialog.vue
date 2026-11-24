<script setup lang="ts">
const { confirmState, resolveConfirm } = useWorkspace()
</script>

<template>
  <Teleport to="body">
    <div
      v-if="confirmState?.open"
      class="fixed inset-0 z-[90] flex items-center justify-center bg-ink-950/45 px-4 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      :aria-label="confirmState.title"
      @click.self="resolveConfirm(false)"
    >
      <div class="w-full max-w-md rounded-2xl border border-ink-900/10 bg-surface p-5 shadow-2xl">
        <h2 class="text-[16px] font-semibold tracking-[-0.02em]">{{ confirmState.title }}</h2>
        <p class="mt-2 text-[12.5px] leading-relaxed text-ink-900/55">{{ confirmState.message }}</p>
        <div class="mt-5 flex justify-end gap-2">
          <button
            type="button"
            class="rounded-lg border border-ink-900/15 bg-surface px-3.5 py-2 text-[12.5px] font-medium text-ink-900 transition hover:bg-brand-50"
            @click="resolveConfirm(false)"
          >
            {{ confirmState.cancelLabel }}
          </button>
          <button
            type="button"
            class="rounded-lg px-3.5 py-2 text-[12.5px] font-semibold transition"
            :class="confirmState.danger ? 'bg-red-600 text-white hover:bg-red-700' : 'bg-accent-500 text-ink-950 hover:bg-accent-400'"
            @click="resolveConfirm(true)"
          >
            {{ confirmState.confirmLabel }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
