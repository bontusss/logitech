<script setup lang="ts">
const active = ref<'customers' | 'businesses'>('customers')

const tabs = [
  { id: 'customers' as const, label: 'For customers', hint: 'Compare and ship' },
  { id: 'businesses' as const, label: 'For logistics businesses', hint: 'Run and grow' },
]

const steps = {
  customers: [
    {
      title: 'Enter shipment details',
      body: 'Tell us the pickup city, destination, package weight and category.',
    },
    {
      title: 'Compare logistics companies',
      body: 'Review providers and compare price, delivery time and service coverage.',
    },
    {
      title: 'Select your preferred service',
      body: 'Choose the option that fits your budget and how quickly it needs to arrive.',
    },
    {
      title: 'Book and track your shipment',
      body: 'Confirm the booking and follow the shipment from pickup through to delivery.',
    },
  ],
  businesses: [
    {
      title: 'Create your company account',
      body: 'Set up your logistics business on the platform with your own profile and identity.',
    },
    {
      title: 'Set up branches and staff',
      body: 'Add terminals, assign roles and give your team access to the parts they need.',
    },
    {
      title: 'Publish your shipping rates',
      body: 'Define routes, service areas, weight-based pricing and delivery timelines.',
    },
    {
      title: 'Receive bookings and manage shipments',
      body: 'Handle incoming bookings, manifests, transfers and delivery confirmations.',
    },
    {
      title: 'Grow your customer base',
      body: 'Be discoverable in the marketplace and reach customers beyond your existing network.',
    },
  ],
}
</script>

<template>
  <section id="how-it-works" class="border-t border-ink-900/10 bg-canvas">
    <div class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-24">
      <SectionHeading
        align="center"
        eyebrow="How it works"
        title="Two sides of the same network."
        lede="Whether you are sending a package or running the operation, the process is designed to be straightforward."
      />

      <div class="mx-auto mt-8 flex max-w-md rounded-xl border border-ink-900/10 bg-white p-1" role="tablist" aria-label="Choose your path">
        <button
          v-for="tab in tabs"
          :id="`tab-${tab.id}`"
          :key="tab.id"
          type="button"
          role="tab"
          :aria-selected="active === tab.id"
          :aria-controls="`panel-${tab.id}`"
          class="flex-1 rounded-lg px-3 py-2.5 text-center transition"
          :class="active === tab.id ? 'bg-ink-900 text-white shadow-sm' : 'text-ink-900/55 hover:text-ink-900'"
          @click="active = tab.id"
        >
          <span class="block text-[12.5px] font-semibold">{{ tab.label }}</span>
          <span
            class="mt-0.5 block text-[10px]"
            :class="active === tab.id ? 'text-white/50' : 'text-ink-900/35'"
          >
            {{ tab.hint }}
          </span>
        </button>
      </div>

      <div class="mt-10">
        <ol
          :id="`panel-${active}`"
          role="tabpanel"
          :aria-labelledby="`tab-${active}`"
          class="grid gap-4 sm:grid-cols-2"
          :class="active === 'businesses' ? 'lg:grid-cols-5' : 'lg:grid-cols-4'"
        >
          <li
            v-for="(step, index) in steps[active]"
            :key="step.title"
            class="relative rounded-xl border border-ink-900/10 bg-white p-5"
          >
            <span
              class="flex size-7 items-center justify-center rounded-md border border-brand-100 bg-brand-50 text-[11.5px] font-semibold text-brand-700"
            >
              {{ index + 1 }}
            </span>
            <h3 class="mt-3.5 text-[13.5px] leading-snug font-semibold tracking-[-0.01em]">
              {{ step.title }}
            </h3>
            <p class="mt-1.5 text-[12.5px] leading-[1.6] text-ink-900/55">{{ step.body }}</p>
          </li>
        </ol>
      </div>
    </div>
  </section>
</template>
