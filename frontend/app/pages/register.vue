<script setup lang="ts">
import { ArrowRight, Building2, ShoppingBag } from 'lucide-vue-next'
import { isValidNigerianPhone } from '~/utils/nigeria'

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
  business_type: 'Intercity transport',
  business_email: '',
  phone: '',
  contact_name: '',
  country: 'NG',
  state: '',
  city: '',
  address: '',
  password: '',
})

const businessTypes = [
  'Intercity transport',
  'Courier and parcels',
  'Freight forwarding',
  'Haulage and trucking',
  'Marketplace aggregator',
  'Other logistics service',
]

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
  if (!isValidNigerianPhone(company.phone)) {
    error.value = 'Enter a valid Nigerian phone number.'
    return
  }
  pending.value = true
  try {
    await registerCompany({ ...company })
    await navigateTo('/onboarding')
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not register your company.')
  } finally {
    pending.value = false
  }
}

const inputClass =
  'w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-sm transition hover:border-slate-400 focus:border-brand-500'
const labelClass = 'mb-1.5 block text-[11px] font-medium text-ink-900/55'
</script>

<template>
  <div>
    <p class="text-[11px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Get started</p>
    <h1 class="mt-2 text-[1.8rem] leading-[1.1] font-semibold tracking-[-0.035em]">
      Create your BOL account
    </h1>
    <p class="mt-3 text-[13.5px] leading-relaxed text-ink-900/50">
      Ship with logistics providers, or bring your own company onto the platform.
    </p>

    <div class="mt-6 grid grid-cols-2 gap-2 rounded-xl border border-ink-900/10 bg-white p-1">
      <button
        type="button"
        class="flex items-center justify-center gap-2 rounded-lg px-3 py-2.5 text-[12px] font-semibold transition"
        :class="mode === 'customer' ? 'bg-brand-500 text-white' : 'text-ink-900/55 hover:text-ink-900'"
        @click="mode = 'customer'"
      >
        <ShoppingBag :size="14" />
        Marketplace customer
      </button>
      <button
        type="button"
        class="flex items-center justify-center gap-2 rounded-lg px-3 py-2.5 text-[12px] font-semibold transition"
        :class="mode === 'company' ? 'bg-brand-500 text-white' : 'text-ink-900/55 hover:text-ink-900'"
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
        <input v-model="customer.password" type="password" required minlength="8" autocomplete="new-password" :class="inputClass" placeholder="At least 8 characters" />
      </label>

      <button
        type="submit"
        :disabled="pending"
        class="flex w-full items-center justify-center gap-2 rounded-lg bg-brand-500 px-4 py-3 text-sm font-semibold text-white transition hover:bg-brand-600 disabled:opacity-60"
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
        <label class="block">
          <span :class="labelClass">Business type</span>
          <select v-model="company.business_type" :class="inputClass">
            <option v-for="type in businessTypes" :key="type" :value="type">{{ type }}</option>
          </select>
        </label>
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
        <input v-model="company.password" type="password" required minlength="8" autocomplete="new-password" :class="inputClass" placeholder="At least 8 characters" />
      </label>

      <button
        type="submit"
        :disabled="pending"
        class="flex w-full items-center justify-center gap-2 rounded-lg bg-brand-500 px-4 py-3 text-sm font-semibold text-white transition hover:bg-brand-600 disabled:opacity-60"
      >
        {{ pending ? 'Registering company…' : 'Register company' }}
        <ArrowRight v-if="!pending" :size="15" />
      </button>
      <p class="text-[11px] leading-relaxed text-ink-900/40">
        You become the company owner. Your company starts as pending until its profile is reviewed.
      </p>
    </form>

    <p class="mt-6 text-center text-[12.5px] text-ink-900/50">
      Already have an account?
      <NuxtLink to="/login" class="font-semibold text-brand-600 hover:text-brand-700">
        Sign in
      </NuxtLink>
    </p>
  </div>
</template>
