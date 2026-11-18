<script setup lang="ts">
import {
  ArrowLeft,
  CalendarCheck,
  CircleAlert,
  ExternalLink,
  Loader2,
  Mail,
  MapPin,
  MessageCircle,
  Phone,
} from 'lucide-vue-next'

definePageMeta({ layout: false })

interface PublicBookingPage {
  booking_page: {
    slug: string
    headline: string | null
    description: string | null
  }
  company: {
    id: string
    name: string
    slug: string
    public_name: string
    business_type: string
    business_email: string
    phone: string
    city: string | null
    state: string | null
    country: string
  }
  branding: {
    logo_url: string | null
    primary_color: string
    secondary_color: string
    cover_image_url: string | null
  }
  subdomain: string
}

const route = useRoute()
const config = useRuntimeConfig()
const origin = useRequestURL().origin
const slug = computed(() => String(route.params.slug ?? ''))

const { data, error, status } = await useAsyncData(
  () => `booking-page:${slug.value}`,
  () =>
    $fetch<PublicBookingPage>(`/public/book/${encodeURIComponent(slug.value)}`, {
      baseURL: (config.public.apiBase as string) || '/api/v1',
    }),
  { watch: [slug] },
)

const company = computed(() => data.value?.company ?? null)
const branding = computed(() => data.value?.branding ?? null)
const primary = computed(() => branding.value?.primary_color || '#2b4bff')
const secondary = computed(() => branding.value?.secondary_color || '#0a1220')

const displayName = computed(() => company.value?.public_name || company.value?.name || '')
const headline = computed(
  () => data.value?.booking_page.headline || `Book with ${displayName.value}`,
)
const description = computed(
  () =>
    data.value?.booking_page.description ||
    `Get in touch with ${displayName.value} to plan your next delivery.`,
)
const location = computed(() =>
  [company.value?.city, company.value?.state, company.value?.country].filter(Boolean).join(', '),
)
const initials = computed(() =>
  displayName.value
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((part: string) => part[0]?.toUpperCase())
    .join(''),
)

const phoneDigits = computed(() => (company.value?.phone || '').replace(/\D/g, ''))
const whatsappHref = computed(() => {
  const digits = phoneDigits.value
  if (!digits) return ''
  return `https://wa.me/${digits.startsWith('0') ? `234${digits.slice(1)}` : digits}`
})
const telHref = computed(() => `tel:${(company.value?.phone || '').replace(/[^+\d]/g, '')}`)
const websiteHref = computed(() =>
  company.value && data.value?.subdomain
    ? `http://${data.value.subdomain}.${config.public.appDomain}`
    : '',
)
const pageURL = computed(() => `${origin}/book/${data.value?.booking_page.slug || slug.value}`)

useSeoMeta({
  title: () =>
    data.value ? `${headline.value} · ${displayName.value}` : 'Booking page',
  description: () => description.value,
  ogTitle: () => headline.value,
  ogDescription: () => description.value,
  ogType: 'website',
  ogUrl: () => pageURL.value,
  twitterCard: 'summary',
})
</script>

<template>
  <div class="min-h-dvh bg-canvas text-ink-900">
    <header class="relative h-44 sm:h-56" :style="{ backgroundColor: secondary }">
      <img
        v-if="branding?.cover_image_url"
        :src="branding.cover_image_url"
        alt=""
        class="h-full w-full object-cover opacity-70"
      />
      <div
        v-else
        class="h-full w-full opacity-90"
        :style="{ background: `linear-gradient(135deg, ${secondary}, ${primary})` }"
      />
    </header>

    <main class="mx-auto -mt-20 max-w-[880px] px-5 pb-16 sm:px-8">
      <div
        v-if="status === 'pending'"
        class="flex items-center gap-2 rounded-2xl border border-ink-900/10 bg-white p-10 text-[13px] text-ink-900/50 shadow-[0_24px_60px_-45px_rgba(10,18,32,0.4)]"
      >
        <Loader2 :size="16" class="animate-spin" /> Loading booking page…
      </div>

      <div
        v-else-if="error || !data || !company"
        class="rounded-2xl border border-ink-900/10 bg-white p-8 text-center shadow-[0_24px_60px_-45px_rgba(10,18,32,0.4)] sm:p-12"
      >
        <div class="mx-auto flex size-11 items-center justify-center rounded-full bg-red-50 text-red-500">
          <CircleAlert :size="22" />
        </div>
        <h1 class="mt-4 text-[19px] font-semibold">This booking page is not available</h1>
        <p class="mx-auto mt-2 max-w-sm text-[13px] leading-relaxed text-ink-900/50">
          The link may be unpublished or incorrect. Check it with the business that shared it.
        </p>
        <NuxtLink
          to="/"
          class="mt-5 inline-flex items-center gap-1.5 text-[12.5px] font-semibold text-brand-600 hover:text-brand-700"
        >
          <ArrowLeft :size="14" /> Back to BOL
        </NuxtLink>
      </div>

      <template v-else>
        <section
          class="rounded-2xl border border-ink-900/10 bg-white p-5 shadow-[0_24px_60px_-45px_rgba(10,18,32,0.4)] sm:p-8"
        >
          <div class="flex flex-wrap items-start gap-4">
            <div
              class="flex size-16 shrink-0 items-center justify-center overflow-hidden rounded-2xl text-[19px] font-bold text-white ring-4 ring-white"
              :style="{ backgroundColor: primary }"
            >
              <img
                v-if="branding?.logo_url"
                :src="branding.logo_url"
                :alt="displayName"
                class="h-full w-full object-cover"
              />
              <span v-else>{{ initials }}</span>
            </div>
            <div class="min-w-0 flex-1">
              <h1 class="text-[clamp(1.5rem,3.4vw,2rem)] leading-tight font-semibold tracking-[-0.03em]">
                {{ displayName }}
              </h1>
              <p class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-[12px] text-ink-900/50">
                <span v-if="company.business_type">{{ company.business_type }}</span>
                <span v-if="location" class="inline-flex items-center gap-1">
                  <MapPin :size="12" /> {{ location }}
                </span>
              </p>
            </div>
            <span
              class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10.5px] font-semibold"
              :style="{ backgroundColor: `${primary}14`, color: primary }"
            >
              <CalendarCheck :size="12" /> Booking page
            </span>
          </div>

          <div class="mt-7 border-t border-ink-900/10 pt-6">
            <h2 class="text-[clamp(1.25rem,2.6vw,1.6rem)] leading-snug font-semibold tracking-[-0.03em]">
              {{ headline }}
            </h2>
            <p class="mt-3 max-w-2xl text-[13.5px] leading-[1.7] whitespace-pre-line text-ink-900/60">
              {{ description }}
            </p>
          </div>

          <div class="mt-6 flex flex-wrap gap-2.5">
            <a
              v-if="whatsappHref"
              :href="whatsappHref"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-2 rounded-lg px-4 py-2.5 text-[12.5px] font-semibold text-white transition hover:opacity-90"
              :style="{ backgroundColor: primary }"
            >
              <MessageCircle :size="15" /> Chat on WhatsApp
            </a>
            <a
              v-if="company.business_email"
              :href="`mailto:${company.business_email}`"
              class="inline-flex items-center gap-2 rounded-lg border border-ink-900/15 bg-white px-4 py-2.5 text-[12.5px] font-semibold transition hover:border-ink-900/30"
            >
              <Mail :size="15" /> {{ company.business_email }}
            </a>
            <a
              v-if="company.phone"
              :href="telHref"
              class="inline-flex items-center gap-2 rounded-lg border border-ink-900/15 bg-white px-4 py-2.5 text-[12.5px] font-semibold transition hover:border-ink-900/30"
            >
              <Phone :size="15" /> {{ company.phone }}
            </a>
            <a
              v-if="websiteHref"
              :href="websiteHref"
              target="_blank"
              rel="noopener"
              class="inline-flex items-center gap-2 rounded-lg border border-ink-900/15 bg-white px-4 py-2.5 text-[12.5px] font-semibold transition hover:border-ink-900/30"
            >
              <ExternalLink :size="15" /> Visit website
            </a>
          </div>

          <p
            v-if="!whatsappHref && !company.business_email && !company.phone"
            class="mt-5 rounded-lg border border-ink-900/10 bg-canvas px-3.5 py-3 text-[12px] text-ink-900/50"
          >
            This business has not added contact details yet.
          </p>
        </section>

        <footer class="mt-8 flex flex-wrap items-center justify-between gap-3 px-1">
          <NuxtLink
            to="/"
            class="inline-flex items-center gap-1.5 text-[12px] font-medium text-ink-900/45 transition hover:text-ink-900"
          >
            <ArrowLeft :size="13" /> Back to BOL
          </NuxtLink>
          <p class="text-[11.5px] text-ink-900/35">
            Powered by
            <NuxtLink to="/" class="font-semibold text-ink-900/55 hover:text-ink-900">
              based on logistics
            </NuxtLink>
          </p>
        </footer>
      </template>
    </main>
  </div>
</template>
