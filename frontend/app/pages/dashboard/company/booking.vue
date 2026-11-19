<script setup lang="ts">
import {
  CalendarCheck,
  Check,
  CheckCircle2,
  CircleAlert,
  Copy,
  Download,
  ExternalLink,
  Loader2,
  Save,
} from 'lucide-vue-next'
import * as QRCode from 'qrcode'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Booking page' })

interface BookingPage {
  slug: string
  headline: string | null
  description: string | null
  is_published: boolean
}

const { api } = useApi()
const { detail, loading, load, can, notify } = useWorkspace()

const page = ref<BookingPage | null>(null)
const pageLoading = ref(true)
const pageError = ref('')
const saving = ref(false)
const copied = ref(false)
const error = ref('')
const checkingSlug = ref(false)
const slugCheckError = ref('')
const slugAvailability = ref<{ slug: string; available: boolean; reason?: string } | null>(null)
const qrCode = ref('')
const qrError = ref('')
const generatingQr = ref(false)
const origin = useRequestURL().origin

const form = reactive({
  slug: '',
  headline: '',
  description: '',
  is_published: true,
})

const canManage = computed(() => can('bookings.manage'))
const isVerified = computed(() => detail.value?.company.status === 'verified')
const publicURL = computed(() => `${origin}/book/${form.slug || 'your-page'}`)
const availabilityForCurrentSlug = computed(() =>
  slugAvailability.value?.slug === form.slug ? slugAvailability.value : null,
)

onMounted(async () => {
  await load()
  if (canManage.value) await loadPage()
})

async function loadPage() {
  pageLoading.value = true
  pageError.value = ''
  try {
    const data = await api<{ booking_page: BookingPage | null }>('/companies/me/booking-page')
    page.value = data.booking_page
    if (data.booking_page) {
      Object.assign(form, {
        slug: data.booking_page.slug,
        headline: data.booking_page.headline ?? '',
        description: data.booking_page.description ?? '',
        is_published: data.booking_page.is_published,
      })
    } else if (detail.value) {
      form.slug = defaultSlug(detail.value.company.name)
      if (form.slug && isVerified.value) await checkSlugAvailability()
    }
  } catch (err) {
    pageError.value = apiErrorMessage(err, 'Could not load your booking page.')
  } finally {
    pageLoading.value = false
  }
}

function normalizeSlug() {
  form.slug = form.slug.toLowerCase().replace(/[^a-z0-9-]/g, '')
  slugAvailability.value = null
  slugCheckError.value = ''
  qrCode.value = ''
}

function defaultSlug(companyName: string) {
  const firstWord = companyName.trim().split(/\s+/)[0] ?? ''
  return firstWord
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, '')
    .replace(/-+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 40)
    .replace(/-+$/g, '')
}

async function checkSlugAvailability() {
  slugCheckError.value = ''
  if (!form.slug) {
    slugAvailability.value = null
    return
  }
  checkingSlug.value = true
  try {
    slugAvailability.value = await api('/companies/me/booking-page/availability', {
      query: { slug: form.slug },
    })
  } catch (err) {
    slugCheckError.value = apiErrorMessage(err, 'Could not check this page name.')
  } finally {
    checkingSlug.value = false
  }
}

async function save() {
  error.value = ''
  if (!/^[a-z0-9]([a-z0-9-]{0,38}[a-z0-9])?$/.test(form.slug) || form.slug.includes('--')) {
    error.value =
      'Use 1–40 lowercase letters, numbers or single hyphens, without a leading or trailing hyphen.'
    return
  }
  saving.value = true
  try {
    const data = await api<{ booking_page: BookingPage }>('/companies/me/booking-page', {
      method: 'PUT',
      body: {
        slug: form.slug,
        headline: form.headline.trim(),
        description: form.description.trim(),
        is_published: form.is_published,
      },
    })
    page.value = data.booking_page
    notify('Booking page saved.')
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not save your booking page.')
  } finally {
    saving.value = false
  }
}

async function copyLink() {
  try {
    await navigator.clipboard.writeText(publicURL.value)
    copied.value = true
    setTimeout(() => (copied.value = false), 1800)
  } catch {
    error.value = 'Could not copy the link. Copy it from the address shown instead.'
  }
}

async function generateQrCode() {
  qrError.value = ''
  generatingQr.value = true
  try {
    qrCode.value = await QRCode.toDataURL(publicURL.value, { width: 256, margin: 2 })
  } catch {
    qrError.value = 'Could not create the QR code. Please try again.'
  } finally {
    generatingQr.value = false
  }
}

function downloadQrCode() {
  if (!qrCode.value) return
  const link = document.createElement('a')
  link.href = qrCode.value
  link.download = `${form.slug}-booking-page.png`
  link.click()
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">
          Company management
        </p>
        <h1 class="mt-2 text-[clamp(1.6rem,3vw,2.1rem)] leading-tight font-semibold tracking-[-0.035em]">
          Booking page
        </h1>
        <p class="mt-2 max-w-2xl text-[13.5px] leading-relaxed text-ink-900/50">
          A public page customers can open to learn about your business and get in touch. Your page
          lives at <span class="font-mono text-[12.5px] text-ink-900/70">{{ publicURL }}</span>.
        </p>
      </div>
      <LookupRefreshButton :loading="pageLoading" @refresh="loadPage" />
    </div>

    <p
      v-if="loading && !detail"
      class="mt-8 flex items-center gap-2 rounded-xl border border-ink-900/10 bg-white p-6 text-[13px] text-ink-900/50"
    >
      <Loader2 :size="16" class="animate-spin" /> Loading workspace…
    </p>

    <p
      v-else-if="!canManage"
      role="alert"
      class="mt-8 flex items-start gap-2 rounded-xl border border-amber-200 bg-amber-50 p-4 text-[12.5px] text-amber-800"
    >
      <CircleAlert :size="16" class="mt-0.5 shrink-0" />
      You do not have permission to manage the booking page.
    </p>

    <template v-else>
      <div
        v-if="detail && !isVerified"
        class="mt-6 flex flex-wrap items-start gap-3 rounded-xl border border-amber-200 bg-amber-50 p-4 text-[12.5px] text-amber-800"
      >
        <CircleAlert :size="16" class="mt-0.5 shrink-0" />
        <p class="flex-1">
          Your booking page becomes available once your company is approved. Keep your company
          profile and verification details up to date in the meantime.
        </p>
        <NuxtLink
          to="/dashboard/company"
          class="font-semibold text-amber-900 underline underline-offset-4"
        >
          Company profile
        </NuxtLink>
      </div>

      <p
        v-if="pageError"
        role="alert"
        class="mt-6 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 p-4 text-[12.5px] text-red-700"
      >
        <CircleAlert :size="16" class="mt-0.5 shrink-0" /> {{ pageError }}
      </p>

      <div class="mt-5 grid gap-5 lg:grid-cols-[minmax(0,1fr)_320px]">
        <form
          class="rounded-xl border border-ink-900/10 bg-white p-5 sm:p-6"
          @submit.prevent="save"
        >
          <h2 class="flex items-center gap-2 text-[15px] font-semibold">
            <CalendarCheck :size="16" /> Page details
          </h2>
          <p class="mt-1 text-[11.5px] text-ink-900/45">
            Choose the link customers will see and the message shown on the page.
          </p>

          <div class="mt-5 space-y-4">
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Page name</span>
              <div
                class="flex items-center rounded-lg border border-slate-300 bg-white focus-within:border-brand-500"
              >
                <span class="border-r border-slate-200 px-3 py-2.5 font-mono text-[12px] text-ink-900/40">
                  /book/
                </span>
                <input
                  v-model="form.slug"
                  :disabled="!isVerified || saving"
                  maxlength="40"
                  class="w-full rounded-r-lg px-3 py-2.5 font-mono text-[13px] outline-none disabled:bg-slate-50"
                  placeholder="aba-freight"
                  @input="normalizeSlug"
                  @blur="checkSlugAvailability"
                />
              </div>
              <span class="mt-1 block text-[10px] text-ink-900/35">
                1–40 lowercase letters, numbers or hyphens. This is the public address of your page.
              </span>
              <span v-if="checkingSlug" class="mt-1 flex items-center gap-1 text-[10.5px] text-ink-900/45">
                <Loader2 :size="12" class="animate-spin" /> Checking availability…
              </span>
              <span
                v-else-if="availabilityForCurrentSlug?.available"
                class="mt-1 flex items-center gap-1 text-[10.5px] font-medium text-emerald-700"
              >
                <CheckCircle2 :size="12" /> This page name is available.
              </span>
              <span
                v-else-if="availabilityForCurrentSlug"
                class="mt-1 flex items-center gap-1 text-[10.5px] text-red-700"
              >
                <CircleAlert :size="12" /> {{ availabilityForCurrentSlug.reason || 'That page name is taken.' }}
              </span>
              <span v-else-if="slugCheckError" role="alert" class="mt-1 block text-[10.5px] text-red-700">
                {{ slugCheckError }}
              </span>
            </label>

            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Headline</span>
              <input
                v-model="form.headline"
                :disabled="!isVerified || saving"
                maxlength="120"
                class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500 disabled:bg-slate-50"
                placeholder="Book your next delivery with us"
              />
            </label>

            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Description</span>
              <textarea
                v-model="form.description"
                :disabled="!isVerified || saving"
                rows="4"
                maxlength="600"
                class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500 disabled:bg-slate-50"
                placeholder="Tell customers what you deliver, where you operate and how to reach you."
              />
              <span class="mt-1 block text-right text-[10px] text-ink-900/35">
                {{ form.description.length }}/600
              </span>
            </label>

            <label class="flex items-start gap-2.5 rounded-lg border border-ink-900/10 bg-canvas px-3 py-3">
              <input
                v-model="form.is_published"
                type="checkbox"
                :disabled="!isVerified || saving"
                class="mt-0.5 size-4 accent-brand-600"
              />
              <span>
                <span class="block text-[12.5px] font-medium">Publish this page</span>
                <span class="block text-[11px] text-ink-900/45">
                  Unpublished pages return “not found” until you turn publishing back on.
                </span>
              </span>
            </label>
          </div>

          <p
            v-if="error"
            role="alert"
            class="mt-4 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 p-3 text-[12px] text-red-700"
          >
            <CircleAlert :size="15" class="mt-0.5 shrink-0" /> {{ error }}
          </p>

          <div class="mt-5 flex items-center justify-end">
            <button
              type="submit"
              :disabled="!isVerified || saving"
              class="inline-flex items-center gap-2 rounded-lg bg-ink-900 px-4 py-2.5 text-[12.5px] font-semibold text-white transition hover:bg-ink-900/90 disabled:cursor-not-allowed disabled:opacity-40"
            >
              <Loader2 v-if="saving" :size="15" class="animate-spin" />
              <Save v-else :size="15" />
              {{ page ? 'Save changes' : 'Create booking page' }}
            </button>
          </div>
        </form>

        <aside class="space-y-4">
          <div class="rounded-xl border border-ink-900/10 bg-white p-5">
            <h2 class="text-[13px] font-semibold">Your page link</h2>
            <p class="mt-2 break-all rounded-lg bg-canvas px-3 py-2.5 font-mono text-[11.5px] text-ink-900/70">
              {{ publicURL }}
            </p>
            <div class="mt-3 flex flex-wrap gap-2">
              <button
                type="button"
                :disabled="!page"
                class="inline-flex items-center gap-1.5 rounded-lg border border-ink-900/10 px-3 py-2 text-[11.5px] font-semibold text-ink-900/70 transition hover:border-ink-900/25 disabled:opacity-40"
                @click="copyLink"
              >
                <Check v-if="copied" :size="13" class="text-emerald-600" />
                <Copy v-else :size="13" />
                {{ copied ? 'Copied' : 'Copy link' }}
              </button>
              <a
                v-if="page"
                :href="publicURL"
                target="_blank"
                rel="noopener"
                class="inline-flex items-center gap-1.5 rounded-lg border border-ink-900/10 px-3 py-2 text-[11.5px] font-semibold text-brand-600 transition hover:border-brand-300"
              >
                Open page <ExternalLink :size="13" />
              </a>
            </div>
            <div v-if="page" class="mt-4 border-t border-ink-900/8 pt-4">
              <button
                type="button"
                :disabled="generatingQr"
                class="inline-flex items-center gap-1.5 rounded-lg border border-ink-900/10 px-3 py-2 text-[11.5px] font-semibold text-ink-900/70 transition hover:border-ink-900/25 disabled:opacity-50"
                @click="generateQrCode"
              >
                <Loader2 v-if="generatingQr" :size="13" class="animate-spin" />
                <CalendarCheck v-else :size="13" />
                {{ qrCode ? 'Refresh QR code' : 'Create shareable QR code' }}
              </button>
              <p v-if="qrError" role="alert" class="mt-2 text-[11px] text-red-700">{{ qrError }}</p>
              <div v-if="qrCode" class="mt-3 flex flex-col items-center rounded-lg border border-ink-900/8 bg-white p-3">
                <img :src="qrCode" alt="QR code for the company booking page" class="size-48" />
                <button
                  type="button"
                  class="mt-2 inline-flex items-center gap-1.5 text-[11.5px] font-semibold text-brand-600 hover:text-brand-700"
                  @click="downloadQrCode"
                >
                  <Download :size="13" /> Download QR code
                </button>
              </div>
            </div>
            <p v-if="page && !page.is_published" class="mt-3 text-[11px] text-amber-700">
              This page is currently unpublished, so visitors will see “not found”.
            </p>
          </div>

          <div class="rounded-xl border border-ink-900/10 bg-canvas p-5 text-[11.5px] leading-relaxed text-ink-900/50">
            <p class="font-semibold text-ink-900/70">What customers see</p>
            <p class="mt-1.5">
              Your public name, logo, colours, business details and contact options, along with the
              headline and description you set here.
            </p>
          </div>
        </aside>
      </div>
    </template>
  </div>
</template>
