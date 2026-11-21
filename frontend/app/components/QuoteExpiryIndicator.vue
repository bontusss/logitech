<script setup lang="ts">
import { Clock3 } from 'lucide-vue-next'

const props = defineProps<{ expiresAt: string }>()

const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | undefined

onMounted(() => {
  timer = setInterval(() => {
    now.value = Date.now()
  }, 1000)
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})

const remainingMs = computed(() => new Date(props.expiresAt).getTime() - now.value)
const expired = computed(() => remainingMs.value <= 0)
const label = computed(() => {
  if (expired.value) return 'Quote expired'
  const totalSeconds = Math.max(0, Math.floor(remainingMs.value / 1000))
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  return `Expires in ${minutes}:${String(seconds).padStart(2, '0')}`
})
</script>

<template>
  <span
    class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[10px] font-semibold"
    :class="
      expired
        ? 'border-red-200 bg-red-50 text-red-700 dark:border-red-400/25 dark:bg-red-400/10 dark:text-red-200'
        : 'border-slate-200 bg-white/70 text-ink-900/55 dark:border-white/10 dark:bg-white/5'
    "
  >
    <Clock3 :size="11" />
    {{ label }}
  </span>
</template>
