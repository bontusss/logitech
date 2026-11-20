<script setup lang="ts">
import { AlertCircle, CheckCircle2, Clock3, Loader2, MapPin, Package } from 'lucide-vue-next'

definePageMeta({ layout: false })
useSeoMeta({ title: 'Track your delivery', robots: 'noindex, nofollow, noarchive' })

const route = useRoute()
const { api } = useApi()
const loading = ref(true)
const error = ref('')
const tracking = ref<{ tracking_code:string; status:string; pickup_address:string; dropoff_address:string; updated_at:string; timeline:Array<{status:string;note:string|null;created_at:string}> }|null>(null)

onMounted(async()=>{
  try{tracking.value=await api(`/public/tracking/${encodeURIComponent(String(route.params.code))}`)}
  catch(err){error.value=apiErrorMessage(err,'This tracking code was not found.')}
  finally{loading.value=false}
})

function statusLabel(status:string){return status.replaceAll('_',' ')}
</script>

<template>
  <main class="min-h-dvh bg-canvas px-4 py-8 text-ink-900">
    <section class="mx-auto w-full max-w-xl rounded-2xl border border-ink-900/10 bg-white p-5 shadow-sm sm:p-7">
      <AppLogo :size="34" />
      <div v-if="loading" class="mt-8 flex items-center gap-2 text-[12.5px] text-ink-900/50"><Loader2 :size="16" class="animate-spin"/>Loading tracking details…</div>
      <div v-else-if="error" role="alert" class="mt-8 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 p-4 text-[12.5px] text-red-700"><AlertCircle :size="16" class="mt-0.5 shrink-0"/>{{ error }}</div>
      <template v-else-if="tracking">
        <p class="mt-7 text-[10px] font-semibold tracking-[0.18em] text-brand-600 uppercase">Live delivery tracking</p>
        <h1 class="mt-2 text-[23px] font-semibold tracking-[-0.035em]">{{ tracking.tracking_code }}</h1>
        <div class="mt-5 flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-3.5 py-3 text-emerald-800"><CheckCircle2 :size="16"/><div><p class="text-[9.5px] uppercase text-emerald-700/70">Current status</p><p class="mt-0.5 text-[12.5px] font-semibold capitalize">{{ statusLabel(tracking.status) }}</p></div></div>
        <div class="mt-4 grid gap-3 sm:grid-cols-2"><div class="rounded-xl border border-ink-900/10 p-3"><MapPin :size="14" class="text-brand-600"/><p class="mt-2 text-[9.5px] text-ink-900/40">Pickup</p><p class="mt-1 text-[11.5px] font-medium">{{ tracking.pickup_address }}</p></div><div class="rounded-xl border border-ink-900/10 p-3"><Package :size="14" class="text-rose-600"/><p class="mt-2 text-[9.5px] text-ink-900/40">Destination</p><p class="mt-1 text-[11.5px] font-medium">{{ tracking.dropoff_address }}</p></div></div>
        <section class="mt-5 border-t border-ink-900/8 pt-5"><h2 class="flex items-center gap-2 text-[12.5px] font-semibold"><Clock3 :size="15"/>Journey</h2><div v-if="tracking.timeline.length" class="mt-4 space-y-3"><div v-for="event in tracking.timeline" :key="event.created_at" class="flex gap-3"><span class="mt-1 size-2 shrink-0 rounded-full bg-brand-500 ring-4 ring-brand-50"/><div><p class="text-[11.5px] font-semibold capitalize">{{ statusLabel(event.status) }}</p><p class="mt-0.5 text-[10.5px] text-ink-900/45">{{ event.note || 'Status updated' }} · {{ new Date(event.created_at).toLocaleString() }}</p></div></div></div><p v-else class="mt-3 text-[11.5px] text-ink-900/45">No journey updates yet.</p></section>
      </template>
    </section>
  </main>
</template>
