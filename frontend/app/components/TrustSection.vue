<script setup lang="ts">
import { BadgeCheck, CircleDollarSign, Clock3 } from 'lucide-vue-next'

const benefits = [
  {
    icon: CircleDollarSign,
    title: 'Know the price first',
    body: 'Compare provider pricing before you decide where your package goes.',
    card: 'from-violet-50 to-white dark:from-violet-400/10 dark:to-transparent',
    tile: 'bg-violet-100 text-violet-700 dark:bg-violet-400/15 dark:text-violet-200',
  },
  {
    icon: Clock3,
    title: 'Choose your pace',
    body: 'Balance delivery speed against cost and pick the service that fits.',
    card: 'from-amber-50 to-white dark:from-amber-400/10 dark:to-transparent',
    tile: 'bg-amber-100 text-amber-700 dark:bg-amber-400/15 dark:text-amber-200',
  },
  {
    icon: BadgeCheck,
    title: 'Book with confidence',
    body: 'See who is handling your delivery and what to expect at each step.',
    card: 'from-emerald-50 to-white dark:from-emerald-400/10 dark:to-transparent',
    tile: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-400/15 dark:text-emerald-200',
  },
]
</script>

<template>
  <section class="border-y border-ink-900/8 bg-canvas">
    <div class="mx-auto grid max-w-[1200px] gap-3 px-5 py-6 sm:grid-cols-3 sm:px-8 sm:py-7">
      <article
        v-for="benefit in benefits"
        :key="benefit.title"
        class="flex items-start gap-3 rounded-2xl border border-ink-900/8 bg-gradient-to-br p-4 sm:p-5"
        :class="benefit.card"
      >
        <div class="flex size-9 shrink-0 items-center justify-center rounded-xl shadow-sm" :class="benefit.tile">
          <component :is="benefit.icon" :size="17" />
        </div>
        <div>
          <h2 class="text-[13px] font-semibold">{{ benefit.title }}</h2>
          <p class="mt-1 text-[12px] leading-relaxed text-ink-900/70">{{ benefit.body }}</p>
        </div>
      </article>
    </div>
  </section>
</template>
