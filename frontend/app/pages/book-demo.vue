<script setup lang="ts">
import { ArrowLeft, ArrowRight, CheckCircle2, Mail } from 'lucide-vue-next'

useSeoMeta({
  title: 'Book a demo',
  description:
    'See how BOL helps logistics businesses manage branches, staff, shipments, manifests, pricing and marketplace bookings.',
})

const sent = ref(false)

const form = reactive({
  name: '',
  company: '',
  email: '',
  phone: '',
  branches: '1–2 branches',
  message: '',
})

const branchOptions = ['1–2 branches', '3–5 branches', '6–10 branches', 'More than 10 branches']

const inputClass =
  'w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-sm text-ink-900 transition placeholder:text-ink-900/30 hover:border-slate-400 focus:border-brand-500'

function submit() {
  sent.value = true
}
</script>

<template>
  <div class="mx-auto max-w-[1100px] px-5 py-14 sm:px-8 sm:py-20">
    <NuxtLink
      to="/"
      class="inline-flex items-center gap-1.5 text-[12.5px] font-medium text-ink-900/50 transition hover:text-ink-900"
    >
      <ArrowLeft :size="14" />
      Back to home
    </NuxtLink>

    <div class="mt-8 grid gap-10 lg:grid-cols-[minmax(0,1fr)_minmax(0,0.9fr)]">
      <div>
        <p class="text-[11px] font-semibold tracking-[0.2em] text-brand-600 uppercase">
          Book a demo
        </p>
        <h1 class="mt-3 text-[clamp(1.9rem,4.4vw,2.9rem)] leading-[1.05] font-semibold tracking-[-0.04em]">
          See the platform with your own routes in mind.
        </h1>
        <p class="mt-4 max-w-md text-[15px] leading-[1.65] text-ink-900/55">
          Tell us a little about your operation. We will walk through branch management, shipment
          operations, rate publishing and marketplace bookings.
        </p>

        <ul class="mt-8 space-y-3">
          <li
            v-for="item in [
              'A walkthrough tailored to your branch structure',
              'How rate publishing and marketplace booking work',
              'Practical next steps for onboarding your team',
            ]"
            :key="item"
            class="flex items-start gap-2.5 text-[13.5px] text-ink-900/65"
          >
            <CheckCircle2 :size="16" class="mt-0.5 shrink-0 text-brand-600" />
            {{ item }}
          </li>
        </ul>

        <div class="mt-8 rounded-xl border border-ink-900/10 bg-white p-4">
          <p class="text-[12px] font-semibold">Prefer email?</p>
          <a
            href="mailto:hello@basedonlogistics.com?subject=BOL%20demo%20request"
            class="mt-2 inline-flex items-center gap-2 text-[13px] font-medium text-brand-600 hover:text-brand-700"
          >
            <Mail :size="14" />
            hello@basedonlogistics.com
          </a>
        </div>
      </div>

      <div class="rounded-2xl border border-ink-900/10 bg-white p-5 shadow-[0_24px_60px_-45px_rgba(10,18,32,0.4)] sm:p-6">
        <div v-if="sent" role="status" class="py-8 text-center">
          <div class="mx-auto flex size-11 items-center justify-center rounded-full bg-brand-50 text-brand-600">
            <CheckCircle2 :size="22" />
          </div>
          <h2 class="mt-4 text-[17px] font-semibold">Request captured</h2>
          <p class="mx-auto mt-2 max-w-xs text-[13px] leading-relaxed text-ink-900/55">
            This demo form is currently a front-end prototype. Form delivery will be connected
            before launch — in the meantime, email
            <a href="mailto:hello@basedonlogistics.com" class="font-medium text-brand-600">
              hello@basedonlogistics.com
            </a>
            and the team will follow up.
          </p>
          <button
            type="button"
            class="mt-5 text-[12.5px] font-medium text-ink-900/50 underline underline-offset-4 hover:text-ink-900"
            @click="sent = false"
          >
            Edit request
          </button>
        </div>

        <form v-else class="space-y-4" @submit.prevent="submit">
          <h2 class="text-[15px] font-semibold">Request a demo</h2>

          <div class="grid gap-3 sm:grid-cols-2">
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Full name</span>
              <input v-model="form.name" type="text" required autocomplete="name" :class="inputClass" placeholder="Ada Obi" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Company</span>
              <input v-model="form.company" type="text" required autocomplete="organization" :class="inputClass" placeholder="Company name" />
            </label>
          </div>

          <label class="block">
            <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Work email</span>
            <input v-model="form.email" type="email" required autocomplete="email" :class="inputClass" placeholder="you@company.com" />
          </label>

          <div class="grid gap-3 sm:grid-cols-2">
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Phone (optional)</span>
              <input v-model="form.phone" type="tel" autocomplete="tel" :class="inputClass" placeholder="+234" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Branches</span>
              <select v-model="form.branches" :class="inputClass">
                <option v-for="option in branchOptions" :key="option" :value="option">{{ option }}</option>
              </select>
            </label>
          </div>

          <label class="block">
            <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">
              What would you like to cover?
            </span>
            <textarea
              v-model="form.message"
              rows="4"
              :class="inputClass"
              placeholder="Routes, branches, shipments, current tools…"
            />
          </label>

          <button
            type="submit"
            class="flex w-full items-center justify-center gap-2 rounded-lg bg-brand-500 px-4 py-3 text-sm font-semibold text-white shadow-[0_10px_28px_-14px_rgba(43,75,255,0.85)] transition hover:bg-brand-600"
          >
            Request Demo
            <ArrowRight :size="15" />
          </button>

          <p class="text-[11px] leading-relaxed text-ink-900/40">
            By submitting, you agree to be contacted about your request. This form does not send
            data to a backend yet.
          </p>
        </form>
      </div>
    </div>
  </div>
</template>
