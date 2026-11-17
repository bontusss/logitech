<script setup lang="ts">
import { CheckCircle2, CircleAlert, Info, X } from 'lucide-vue-next'

const { toasts, dismissToast } = useWorkspace()

const tone = {
  success: 'border-emerald-200 bg-emerald-50 text-emerald-800',
  error: 'border-red-200 bg-red-50 text-red-800',
  info: 'border-brand-200 bg-brand-50 text-brand-800',
}

const icon = {
  success: CheckCircle2,
  error: CircleAlert,
  info: Info,
}
</script>

<template>
  <div class="pointer-events-none fixed inset-x-4 top-4 z-[100] flex flex-col items-end gap-2 sm:inset-x-auto sm:right-5">
    <TransitionGroup
      enter-active-class="transition duration-200"
      enter-from-class="translate-y-1 opacity-0"
      leave-active-class="transition duration-150"
      leave-to-class="opacity-0"
    >
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="pointer-events-auto flex max-w-sm items-start gap-2 rounded-xl border px-3.5 py-3 text-[12.5px] shadow-lg"
        :class="tone[toast.type]"
        role="status"
      >
        <component :is="icon[toast.type]" :size="15" class="mt-0.5 shrink-0" />
        <p class="flex-1 leading-relaxed">{{ toast.message }}</p>
        <button
          type="button"
          class="shrink-0 opacity-50 transition hover:opacity-100"
          aria-label="Dismiss notification"
          @click="dismissToast(toast.id)"
        >
          <X :size="13" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>
