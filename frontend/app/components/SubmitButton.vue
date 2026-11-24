<script setup lang="ts">
import { ArrowRight, Loader2 } from 'lucide-vue-next'

withDefaults(
  defineProps<{
    label: string
    pendingLabel?: string
    pending?: boolean
    icon?: boolean
  }>(),
  { pending: false, icon: true },
)
</script>

<template>
  <button
    type="submit"
    :disabled="pending"
    class="flex w-full items-center justify-center gap-2 rounded-xl bg-accent-500 px-4 py-3 text-sm font-semibold text-ink-950 shadow-[0_14px_26px_-16px_rgba(255,122,0,0.75)] transition hover:-translate-y-0.5 hover:bg-accent-400 disabled:translate-y-0 disabled:opacity-60"
  >
    <Loader2 v-if="pending" :size="15" class="animate-spin" />
    {{ pending ? (pendingLabel ?? label) : label }}
    <ArrowRight v-if="icon && !pending" :size="15" />
  </button>
</template>
