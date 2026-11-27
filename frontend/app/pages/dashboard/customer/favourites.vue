<script setup lang="ts">
import {
  ArrowRight,
  CircleAlert,
  Loader2,
  Star,
  Store,
  Trash2,
} from 'lucide-vue-next'
import type { CustomerFavouriteProvider } from '~/types/customer'

useHead({ title: 'Favourite Providers · BOL' })

const { api } = useApi()
const { notify, confirm } = useWorkspace()

const providers = ref<CustomerFavouriteProvider[]>([])
const loading = ref(true)
const error = ref('')
const removingId = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const result = await api<{ providers: CustomerFavouriteProvider[] }>('/customer/favourites')
    providers.value = result.providers
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not load your favourite providers.')
  } finally {
    loading.value = false
  }
}

async function remove(provider: CustomerFavouriteProvider) {
  const confirmed = await confirm({
    title: `Remove ${provider.name}?`,
    message: 'You can save this provider again from the find providers page.',
    confirmLabel: 'Remove provider',
    cancelLabel: 'Keep it',
    danger: true,
  })
  if (!confirmed) return
  removingId.value = provider.id
  error.value = ''
  try {
    await api(`/customer/favourites/${provider.id}`, { method: 'DELETE' })
    providers.value = providers.value.filter((item) => item.id !== provider.id)
    notify('Provider removed from favourites.')
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not remove this provider.')
  } finally {
    removingId.value = ''
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString('en-NG', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

onMounted(load)
</script>

<template>
  <div class="dash-page">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Personal</p>
        <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Favourite providers</h1>
        <p class="mt-1.5 max-w-2xl text-[12.5px] leading-relaxed text-ink-900/50">
          Keep the providers you trust close so your next shipment is quicker to arrange.
        </p>
      </div>
      <AppButton to="/dashboard/customer/providers" variant="secondary" size="md">
        Find providers <ArrowRight :size="14" />
      </AppButton>
    </div>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" />{{ error }}
    </p>

    <div v-if="loading" class="mt-4 flex items-center justify-center gap-2 rounded-xl border border-ink-900/10 bg-surface p-12 text-[12px] text-ink-900/45">
      <Loader2 :size="15" class="animate-spin" /> Loading your providers…
    </div>

    <div v-else-if="!providers.length" class="mt-4 rounded-xl border border-ink-900/10 bg-surface p-10 text-center">
      <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-brand-50 text-brand-700 dark:bg-brand-400/10 dark:text-brand-200">
        <Star :size="20" />
      </span>
      <h2 class="mt-3 text-[13.5px] font-semibold">No favourite providers yet</h2>
      <p class="mx-auto mt-1 max-w-md text-[11.5px] leading-relaxed text-ink-900/45">
        Save a provider while browsing the marketplace and they will appear here for one-tap booking.
      </p>
      <AppButton to="/dashboard/customer/providers" size="sm" class="mt-4">Browse providers</AppButton>
    </div>

    <ul v-else class="mt-4 grid gap-2.5 sm:grid-cols-2">
      <li v-for="provider in providers" :key="provider.id" class="flex flex-col rounded-xl border border-ink-900/10 bg-surface p-4">
        <div class="flex items-start gap-3">
          <span class="flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-ink-900/8 bg-gradient-to-br from-slate-50 to-white text-[13px] font-bold text-brand-700 dark:border-white/10 dark:from-white/10 dark:to-transparent">
            <img v-if="provider.logo_url" :src="provider.logo_url" :alt="`${provider.name} logo`" class="size-full object-cover" />
            <span v-else>{{ provider.name.charAt(0).toUpperCase() }}</span>
          </span>
          <div class="min-w-0 flex-1">
            <p class="truncate text-[12.5px] font-semibold">{{ provider.name }}</p>
            <div class="mt-1 flex flex-wrap items-center gap-x-2.5 gap-y-1 text-[10.5px] text-ink-900/45">
              <span v-if="provider.rating > 0" class="inline-flex items-center gap-1 font-semibold text-amber-700">
                <Star :size="10" class="fill-amber-500 text-amber-500" /> {{ provider.rating.toFixed(1) }}
              </span>
              <span class="inline-flex items-center gap-1"><Store :size="10" /> {{ provider.route_count }} routes</span>
              <span>Saved {{ formatDate(provider.saved_at) }}</span>
            </div>
          </div>
          <button
            type="button"
            :disabled="removingId === provider.id"
            class="rounded-lg p-1.5 text-ink-900/35 transition hover:bg-rose-50 hover:text-rose-700 disabled:opacity-50"
            :aria-label="`Remove ${provider.name} from favourites`"
            @click="remove(provider)"
          >
            <Loader2 v-if="removingId === provider.id" :size="13" class="animate-spin" />
            <Trash2 v-else :size="13" />
          </button>
        </div>

        <p v-if="provider.description" class="mt-2.5 line-clamp-2 text-[11px] leading-relaxed text-ink-900/55">
          {{ provider.description }}
        </p>

        <div class="mt-3 flex items-center justify-between gap-2 border-t border-ink-900/8 pt-3">
          <NuxtLink
            v-if="provider.slug"
            :to="`/book/${provider.slug}`"
            class="inline-flex items-center gap-1 text-[11px] font-semibold text-brand-700 hover:text-brand-600"
          >
            Open booking page <ArrowRight :size="12" />
          </NuxtLink>
          <span v-else class="text-[10.5px] text-ink-900/35">Booking page not published</span>
        </div>
      </li>
    </ul>
  </div>
</template>
