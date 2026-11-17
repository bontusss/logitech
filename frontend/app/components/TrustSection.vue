<script setup lang="ts">
import { BadgeCheck, CircleDollarSign, Clock3 } from 'lucide-vue-next'

const benefits = [
  {
    icon: CircleDollarSign,
    title: 'Know the price first',
    body: 'Compare provider pricing before you decide where your package goes.',
  },
  {
    icon: Clock3,
    title: 'Choose your pace',
    body: 'Balance delivery speed against cost and pick the service that fits.',
  },
  {
    icon: BadgeCheck,
    title: 'Book with confidence',
    body: 'See who is handling your delivery and what to expect at each step.',
  },
]
</script>

<template>
  <section class="border-y border-ink-900/8 bg-white">
    <div class="mx-auto grid max-w-[1200px] gap-px px-5 py-5 sm:grid-cols-3 sm:px-8 sm:py-0">
      <article
        v-for="benefit in benefits"
        :key="benefit.title"
        class="flex items-start gap-3 py-5 sm:px-5 sm:py-7"
      >
        <div class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-brand-50 text-brand-600">
          <component :is="benefit.icon" :size="17" />
        </div>
        <div>
          <h2 class="text-[13px] font-semibold">{{ benefit.title }}</h2>
          <p class="mt-1 text-[12px] leading-relaxed text-ink-900/50">{{ benefit.body }}</p>
        </div>
      </article>
    </div>
  </section>
</template>
