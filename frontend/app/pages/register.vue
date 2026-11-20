<script setup lang="ts">
import { ArrowRight, Building2, ShoppingBag, UserPlus } from 'lucide-vue-next'
import { isValidNigerianPhone } from '~/utils/nigeria'
import {
  providerBusinessTypes,
  serializeProviderBusinessTypes,
  type ProviderBusinessType,
} from '~/utils/providerBusinessTypes'

definePageMeta({ layout: 'auth' })
useSeoMeta({
  title: 'Create an account',
  description: 'Create a BOL marketplace account or register your logistics company.',
})

const { registerCustomer, registerCompany } = useAuth()

const mode = ref<'customer' | 'company'>('customer')
const error = ref('')
const pending = ref(false)

const customer = reactive({ full_name: '', email: '', phone: '', password: '' })
const company = reactive({
  company_name: '',
  business_types: ['intercity'] as ProviderBusinessType[],
  business_email: '',
  phone: '',
  contact_name: '',
  country: 'NG',
  state: '',
  city: '',
  address: '',
  password: '',
})

async function submitCustomer() {
  error.value = ''
  if (customer.phone && !isValidNigerianPhone(customer.phone)) {
    error.value = 'Enter a valid Nigerian phone number.'
    return
  }
  pending.value = true
  try {
    await registerCustomer({ ...customer })
    await navigateTo('/dashboard')
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not create your account.')
  } finally {
    pending.value = false
  }
}

async function submitCompany() {
  error.value = ''
  if (!company.business_types.length) {
    error.value = 'Select at least one business type.'
    return
  }
  if (!isValidNigerianPhone(company.phone)) {
    error.value = 'Enter a valid Nigerian phone number.'
    return
  }
  pending.value = true
  try {
    await registerCompany({
      ...company,
      business_type: serializeProviderBusinessTypes(company.business_types),
      business_types: company.business_types,
    })
    await navigateTo('/onboarding')
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not register your company.')
  } finally {
    pending.value = false
  }
}

const inputClass =
  'w-full rounded-xl border border-slate-200 bg-canvas px-3.5 py-2.5 text-sm transition placeholder:text-ink-900/35 hover:border-slate-300 focus:border-brand-500 dark:border-white/10 dark:bg-white/5'
const labelClass = 'mb-1.5 block text-[11px] font-semibold text-ink-900/70'
</script>

<template>
  <div>
    <div class="flex items-center gap-3">
      <span class="flex size-11 items-center justify-center rounded-2xl bg-gradient-to-br from-sky-100 to-violet-50 text-sky-700 shadow-sm dark:from-sky-400/15 dark:to-violet-400/10 dark:text-sky-200">
        <UserPlus :size="19" />
      </span>
      <p class="text-[10px] font-bold tracking-[0.2em] text-brand-600 uppercase">Get started</p>
    </div>
    <h1 class="mt-4 text-[1.75rem] leading-[1.1] font-semibold tracking-[-0.035em]">
      Create your BOL account
    </h1>
    <p class="mt-2.5 text-[13.5px] font-medium leading-relaxed text-ink-900/65">
      Ship with logistics providers, or bring your own company onto the platform.
    </p>

    <div class="mt-6 grid grid-cols-2 gap-1.5 rounded-2xl border border-ink-900/8 bg-canvas p-1.5 dark:border-white/10 dark:bg-white/5">
      <button
        type="button"
        class="flex items-center justify-center gap-2 rounded-xl px-3 py-2.5 text-[12px] font-semibold transition"
        :class="mode === 'customer' ? 'bg-gradient-to-r from-brand-600 to-violet-600 text-white shadow-sm' : 'text-ink-900/55 hover:bg-white/70 hover:text-ink-900 dark:hover:bg-white/10'"
        @click="mode = 'customer'"
      >
        <ShoppingBag :size="14" />
        Marketplace customer
      </button>
      <button
        type="button"
        class="flex items-center justify-center gap-2 rounded-xl px-3 py-2.5 text-[12px] font-semibold transition"
        :class="mode === 'company' ? 'bg-gradient-to-r from-brand-600 to-violet-600 text-white shadow-sm' : 'text-ink-900/55 hover:bg-white/70 hover:text-ink-900 dark:hover:bg-white/10'"
        @click="mode = 'company'"
      >
        <Building2 :size="14" />
        Logistics company
      </button>
    </div>

    <p v-if="error" role="alert" class="mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
      {{ error }}
    </p>

    <form v-if="mode === 'customer'" class="mt-5 space-y-3.5" @submit.prevent="submitCustomer">
      <label class="block">
        <span :class="labelClass">Full name</span>
        <input v-model="customer.full_name" type="text" required autocomplete="name" :class="inputClass" placeholder="Ada Obi" />
      </label>
      <div class="grid gap-3 sm:grid-cols-2">
        <label class="block">
          <span :class="labelClass">Email</span>
          <input v-model="customer.email" type="email" required autocomplete="email" :class="inputClass" placeholder="you@example.com" />
        </label>
        <label class="block">
          <span :class="labelClass">Phone (optional)</span>
          <NigeriaPhoneInput v-model="customer.phone" />
        </label>
      </div>
      <label class="block">
        <span :class="labelClass">Password</span>
        <PasswordInput v-model="customer.password" required :minlength="8" autocomplete="new-password" :input-class="inputClass" placeholder="At least 8 characters" />
      </label>

      <button
        type="submit"
        :disabled="pending"
        class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-brand-600 to-violet-600 px-4 py-3 text-sm font-semibold text-white shadow-[0_12px_26px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-60"
      >
        {{ pending ? 'Creating account…' : 'Create customer account' }}
        <ArrowRight v-if="!pending" :size="15" />
      </button>
    </form>

    <form v-else class="mt-5 space-y-3.5" @submit.prevent="submitCompany">
      <div class="grid gap-3 sm:grid-cols-2">
        <label class="block">
          <span :class="labelClass">Company name</span>
          <input v-model="company.company_name" type="text" required :class="inputClass" placeholder="Aba Freight Ltd" />
        </label>
        <fieldset class="block">
          <legend :class="labelClass">Business type</legend>
          <div class="flex flex-wrap gap-2">
            <label
              v-for="type in providerBusinessTypes"
              :key="type.value"
              class="flex cursor-pointer items-center gap-2 rounded-full border px-3 py-2 text-[11.5px] transition"
              :class="company.business_types.includes(type.value) ? 'border-brand-400 bg-brand-50 text-brand-700 dark:border-brand-400/40 dark:bg-brand-400/10 dark:text-indigo-200' : 'border-slate-200 bg-canvas text-ink-900/60 hover:border-slate-300 dark:border-white/10 dark:bg-white/5'"
            >
              <input v-model="company.business_types" type="checkbox" :value="type.value" class="size-3.5 rounded border-slate-300" />
              {{ type.label }}
            </label>
          </div>
        </fieldset>
      </div>

      <div class="grid gap-3 sm:grid-cols-2">
        <label class="block">
          <span :class="labelClass">Business email</span>
          <input v-model="company.business_email" type="email" required autocomplete="email" :class="inputClass" placeholder="ops@company.com" />
        </label>
        <label class="block">
          <span :class="labelClass">Phone</span>
          <NigeriaPhoneInput v-model="company.phone" required />
        </label>
      </div>

      <label class="block">
        <span :class="labelClass">Contact person</span>
        <input v-model="company.contact_name" type="text" required autocomplete="name" :class="inputClass" placeholder="Owner's full name" />
      </label>

      <NigeriaLocationFields v-model:state="company.state" v-model:city="company.city" required />

      <label class="block">
        <span :class="labelClass">Business address</span>
        <input v-model="company.address" type="text" required :class="inputClass" placeholder="12 Port Harcourt Road" />
      </label>

      <label class="block">
        <span :class="labelClass">Password</span>
        <PasswordInput v-model="company.password" required :minlength="8" autocomplete="new-password" :input-class="inputClass" placeholder="At least 8 characters" />
      </label>

      <button
        type="submit"
        :disabled="pending"
        class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-brand-600 to-violet-600 px-4 py-3 text-sm font-semibold text-white shadow-[0_12px_26px_-16px_rgba(43,75,255,0.9)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:opacity-60"
      >
        {{ pending ? 'Registering company…' : 'Register company' }}
        <ArrowRight v-if="!pending" :size="15" />
      </button>
      <p class="text-[11px] leading-relaxed text-ink-900/55">
        You become the company owner. Your company starts as pending until its profile is reviewed.
      </p>
    </form>

    <p class="mt-6 text-center text-[12.5px] font-medium text-ink-900/65">
      Already have an account?
      <NuxtLink to="/login" class="font-semibold text-brand-600 hover:text-brand-700">
        Sign in
      </NuxtLink>
    </p>
  </div>
</template>
