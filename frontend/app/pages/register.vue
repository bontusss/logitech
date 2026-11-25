<script setup lang="ts">
import { Building2, CheckCircle2, ShoppingBag, UserPlus } from 'lucide-vue-next'
import { isValidNigerianPhone } from '~/utils/nigeria'

definePageMeta({ layout: 'auth' })
useSeoMeta({
  title: 'Create an account',
  description: 'Create a BOL marketplace account or register your logistics company.',
})

const route = useRoute()
const { registerCustomer, registerCompany, setupCompany: setupGoogleCompany, fetchMe, companies } = useAuth()

const mode = ref<'customer' | 'company'>('customer')
const error = ref('')
const pending = ref(false)
const googleSetup = ref(false)
const googleAccountEmail = ref('')

const customer = reactive({ full_name: '', email: '', phone: '', password: '' })
const company = reactive({
  company_name: '',
  business_email: '',
  phone: '',
  contact_name: '',
  country: 'NG',
  state: '',
  city: '',
  address: '',
  password: '',
})

onMounted(async () => {
  if (route.query.google !== 'provider') return

  googleSetup.value = true
  mode.value = 'company'

  const current = await fetchMe()
  if (!current) {
    googleSetup.value = false
    error.value = 'Your Google session expired. Sign in with Google again.'
    return
  }
  googleAccountEmail.value = current.email

  if (current.account_type !== 'company') {
    googleSetup.value = false
    error.value = 'Only provider accounts can finish a company setup.'
    return
  }
  if (companies.value.length > 0) {
    await navigateTo('/dashboard')
    return
  }

  company.business_email = current.email
  company.contact_name = current.full_name
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
  if (!isValidNigerianPhone(company.phone)) {
    error.value = 'Enter a valid Nigerian phone number.'
    return
  }
  pending.value = true
  try {
    if (googleSetup.value) {
      await setupGoogleCompany({
        company_name: company.company_name,
        business_email: company.business_email,
        phone: company.phone,
        contact_name: company.contact_name,
        country: company.country,
        state: company.state,
        city: company.city,
        address: company.address,
      })
    } else {
      await registerCompany({ ...company })
    }
    await navigateTo('/onboarding')
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not register your company.')
  } finally {
    pending.value = false
  }
}

const labelClass = 'mb-1.5 block text-[11.5px] font-semibold text-ink-900/75'
</script>

<template>
  <div>
    <div class="flex items-center gap-3">
      <IconTile tone="accent-solid" size="lg" rounded="2xl">
        <UserPlus :size="19" />
      </IconTile>
      <p class="text-[10px] font-bold tracking-[0.2em] text-accent-700 uppercase dark:text-accent-400">
        Get started
      </p>
    </div>
    <h1 class="mt-4 text-[1.75rem] leading-[1.1] font-semibold tracking-[-0.035em]">
      {{ googleSetup ? 'Finish your company setup' : 'Create your BOL account' }}
    </h1>
    <p class="mt-2.5 text-[13.5px] font-medium leading-relaxed text-ink-900/65">
      {{ googleSetup
        ? 'Your Google account is connected. Add your company details to open your workspace.'
        : 'Ship with logistics providers, or bring your own company onto the platform.' }}
    </p>

    <div v-if="!googleSetup" class="mt-6 grid grid-cols-2 gap-1.5 rounded-2xl border border-ink-900/10 bg-canvas p-1.5">
      <button
        type="button"
        class="flex items-center justify-center gap-2 rounded-xl px-3 py-2.5 text-[12px] font-semibold transition"
        :class="mode === 'customer' ? 'bg-accent-500 text-ink-950 shadow-sm' : 'text-ink-900/55 hover:bg-surface hover:text-ink-900'"
        @click="mode = 'customer'"
      >
        <ShoppingBag :size="14" />
        Marketplace customer
      </button>
      <button
        type="button"
        class="flex items-center justify-center gap-2 rounded-xl px-3 py-2.5 text-[12px] font-semibold transition"
        :class="mode === 'company' ? 'bg-accent-500 text-ink-950 shadow-sm' : 'text-ink-900/55 hover:bg-surface hover:text-ink-900'"
        @click="mode = 'company'"
      >
        <Building2 :size="14" />
        Logistics company
      </button>
    </div>

    <div v-if="!googleSetup" class="mt-6">
      <GoogleAuthButton
        :intent="mode === 'customer' ? 'customer' : 'provider'"
        label="Sign up with Google"
      />
      <div class="my-4 flex items-center gap-3">
        <span class="h-px flex-1 bg-ink-900/10" />
        <span class="text-[10px] font-medium tracking-wide text-ink-900/35 uppercase">or with email</span>
        <span class="h-px flex-1 bg-ink-900/10" />
      </div>
    </div>

    <p v-if="error" role="alert" class="mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700 dark:border-red-400/25 dark:bg-red-400/10 dark:text-red-200">
      {{ error }}
    </p>

    <div v-if="googleSetup" class="mt-4 flex items-start gap-2.5 rounded-xl border border-brand-100 bg-brand-50/60 p-3 text-brand-700 dark:border-brand-400/20 dark:bg-brand-400/10 dark:text-brand-200">
      <CheckCircle2 :size="15" class="mt-0.5 shrink-0" />
      <p class="text-[11px] leading-relaxed">
        Google connected as <strong class="font-semibold">{{ googleAccountEmail }}</strong>.
        You can sign in with Google from now on.
      </p>
    </div>

    <form v-if="mode === 'customer'" class="mt-5 space-y-3.5" @submit.prevent="submitCustomer">
      <FormTextField
        v-model="customer.full_name"
        label="Full name"
        autocomplete="name"
        placeholder="Ada Obi"
        required
      />
      <div class="grid gap-3 sm:grid-cols-2">
        <FormTextField
          v-model="customer.email"
          label="Email"
          type="email"
          autocomplete="email"
          placeholder="you@example.com"
          required
        />
        <label class="block">
          <span :class="labelClass">Phone (optional)</span>
          <NigeriaPhoneInput v-model="customer.phone" />
        </label>
      </div>
      <label class="block">
        <span :class="labelClass">Password</span>
        <PasswordInput
          v-model="customer.password"
          required
          :minlength="8"
          autocomplete="new-password"
          placeholder="At least 8 characters"
        />
      </label>

      <SubmitButton
        :pending="pending"
        label="Create customer account"
        pending-label="Creating account…"
      />
    </form>

    <form v-else class="mt-5 space-y-3.5" @submit.prevent="submitCompany">
      <div class="grid gap-3 sm:grid-cols-2">
        <FormTextField
          v-model="company.company_name"
          label="Company name"
          placeholder="Aba Freight Ltd"
          required
        />
      </div>

      <div class="grid gap-3 sm:grid-cols-2">
        <FormTextField
          v-model="company.business_email"
          label="Business email"
          type="email"
          autocomplete="email"
          placeholder="ops@company.com"
          required
        />
        <label class="block">
          <span :class="labelClass">Phone</span>
          <NigeriaPhoneInput v-model="company.phone" required />
        </label>
      </div>

      <FormTextField
        v-model="company.contact_name"
        label="Contact person"
        autocomplete="name"
        placeholder="Owner's full name"
        required
      />

      <NigeriaLocationFields v-model:state="company.state" v-model:city="company.city" required />

      <FormTextField
        v-model="company.address"
        label="Business address"
        placeholder="12 Port Harcourt Road"
        required
      />

      <label v-if="!googleSetup" class="block">
        <span :class="labelClass">Password</span>
        <PasswordInput
          v-model="company.password"
          required
          :minlength="8"
          autocomplete="new-password"
          placeholder="At least 8 characters"
        />
      </label>

      <SubmitButton
        :pending="pending"
        :label="googleSetup ? 'Finish company setup' : 'Register company'"
        :pending-label="googleSetup ? 'Saving company…' : 'Registering company…'"
      />
      <p class="text-[11px] leading-relaxed text-ink-900/55">
        You become the company owner. Your company starts as pending until its profile is reviewed.
      </p>
    </form>

    <p v-if="!googleSetup" class="mt-6 text-center text-[12.5px] font-medium text-ink-900/65">
      Already have an account?
      <NuxtLink to="/login" class="font-semibold text-accent-700 hover:text-accent-600 dark:text-accent-400">
        Sign in
      </NuxtLink>
    </p>
  </div>
</template>
