<script setup lang="ts">
import { AlertCircle, Loader2, MapPin, Truck, UserRound } from 'lucide-vue-next'

definePageMeta({ layout: false })
useSeoMeta({
  title: 'Driver portal',
  robots: 'noindex, nofollow, noarchive',
})
useHead({ meta: [{ name: 'referrer', content: 'no-referrer' }] })

const route = useRoute()
const { api } = useApi()

interface DriverPortal {
  driver: {
    full_name: string
    vehicle_plate: string
    company_name: string
  }
  active_assignment: null
  location_tracking: string
}

const portal = ref<DriverPortal | null>(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    portal.value = await api<DriverPortal>(`/public/riders/${encodeURIComponent(String(route.params.token))}`)
  } catch (err) {
    error.value = apiErrorMessage(err, 'This driver portal link is invalid or inactive.')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <main class="flex min-h-dvh items-center justify-center bg-canvas px-4 py-8 text-ink-900">
    <section class="w-full max-w-lg rounded-2xl border border-ink-900/10 bg-white p-6 shadow-sm sm:p-8">
      <AppLogo :size="34" />

      <div v-if="loading" class="mt-8 flex items-center gap-2 text-[13px] text-ink-900/50">
        <Loader2 :size="16" class="animate-spin" /> Opening your driver portal…
      </div>
      <div v-else-if="error" role="alert" class="mt-8 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 p-4 text-[12.5px] text-red-700">
        <AlertCircle :size="16" class="mt-0.5 shrink-0" /> {{ error }}
      </div>
      <template v-else-if="portal">
        <p class="mt-8 text-[10px] font-semibold tracking-[0.18em] text-brand-600 uppercase">Driver portal</p>
        <h1 class="mt-2 text-[24px] font-semibold tracking-[-0.035em]">Hello, {{ portal.driver.full_name }}</h1>
        <p class="mt-1 text-[13px] text-ink-900/50">{{ portal.driver.company_name }}</p>

        <div class="mt-6 grid gap-3 sm:grid-cols-2">
          <div class="rounded-xl border border-ink-900/10 bg-canvas p-4">
            <UserRound :size="16" class="text-brand-600" />
            <p class="mt-2 text-[10px] text-ink-900/40">Driver</p>
            <p class="mt-0.5 text-[12px] font-semibold">{{ portal.driver.full_name }}</p>
          </div>
          <div class="rounded-xl border border-ink-900/10 bg-canvas p-4">
            <Truck :size="16" class="text-brand-600" />
            <p class="mt-2 text-[10px] text-ink-900/40">Vehicle plate</p>
            <p class="mt-0.5 text-[12px] font-semibold uppercase">{{ portal.driver.vehicle_plate }}</p>
          </div>
        </div>

        <div class="mt-5 rounded-xl border border-brand-100 bg-brand-50/60 p-4">
          <div class="flex items-center gap-2 text-brand-800">
            <MapPin :size="16" />
            <h2 class="text-[12.5px] font-semibold">Location privacy</h2>
          </div>
          <p class="mt-2 text-[11.5px] leading-relaxed text-ink-900/60">
            {{ portal.location_tracking }}
          </p>
          <p class="mt-2 text-[10.5px] leading-relaxed text-ink-900/45">
            There are no active delivery assignments yet. Your browser will not request or share location until a delivery is assigned to you.
          </p>
        </div>

        <p class="mt-5 text-[10px] leading-relaxed text-ink-900/35">
          This private link identifies your portal. Don’t forward it. Contact {{ portal.driver.company_name }} if you think the link has been shared by mistake.
        </p>
      </template>
    </section>
  </main>
</template>
