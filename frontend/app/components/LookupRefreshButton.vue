<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'

withDefaults(defineProps<{ loading?: boolean }>(), { loading: false })
defineEmits<{ refresh: [] }>()
</script>

<template>
  <button
    type="button"
    :disabled="loading"
    class="inline-flex items-center gap-1.5 rounded-lg border border-slate-300 bg-white px-3 py-2 text-[11.5px] font-medium transition hover:bg-slate-50 disabled:opacity-50"
    @click="$emit('refresh')"
  >
    <RefreshCw :size="13" :class="loading ? 'animate-spin' : ''" />
    Refresh
  </button>
</template>
