<script setup lang="ts">
import { BadgeCheck, CircleDollarSign, Clock3 } from 'lucide-vue-next'

const benefits = [
  {
    icon: CircleDollarSign,
    title: 'Know the price first',
    body: 'Compare provider pricing before you decide where your package goes.',
    tone: 'accent' as const,
  },
  {
    icon: Clock3,
    title: 'Choose your pace',
    body: 'Balance delivery speed against cost and pick the service that fits.',
    tone: 'brand' as const,
  },
  {
    icon: BadgeCheck,
    title: 'Book with confidence',
    body: 'See who is handling your delivery and what to expect at each step.',
    tone: 'success' as const,
  },
]
</script>

<template>
  <section class="border-y border-ink-900/8 bg-canvas">
    <div class="mx-auto grid max-w-[1200px] gap-3 px-5 py-6 sm:grid-cols-3 sm:px-8 sm:py-7">
      <article
        v-for="benefit in benefits"
        :key="benefit.title"
        class="flex items-start gap-3 rounded-2xl border border-ink-900/8 bg-surface p-4 shadow-sm sm:p-5"
      >
        <IconTile :tone="benefit.tone" rounded="xl" size="sm">
          <component :is="benefit.icon" :size="16" />
        </IconTile>
        <div>
          <h2 class="text-[13px] font-semibold">{{ benefit.title }}</h2>
          <p class="mt-1 text-[12px] leading-relaxed text-ink-900/65">{{ benefit.body }}</p>
        </div>
      </article>
    </div>
  </section>
</template>
