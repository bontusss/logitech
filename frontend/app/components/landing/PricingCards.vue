<script setup lang="ts">
import { PLANS, PRICING_NOTE, formatPrice } from '~/utils/pricing'
</script>

<template>
  <div>
    <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <div
        v-for="plan in PLANS"
        :key="plan.id"
        class="relative flex flex-col rounded-2xl border p-5"
        :class="
          plan.featured
            ? 'border-[#0b1424] bg-[#0b1424] text-white shadow-[0_28px_60px_-35px_rgba(11,20,36,0.7)]'
            : 'border-[#0b1424]/10 bg-white'
        "
      >
        <div class="flex items-center justify-between gap-2">
          <p
            class="text-[13px] font-semibold tracking-[0.02em]"
            :class="plan.featured ? 'text-white' : 'text-[#0b1424]'"
          >
            {{ plan.name }}
          </p>
          <span
            v-if="plan.featured"
            class="rounded-[3px] bg-white/12 px-1.5 py-0.5 text-[9px] font-semibold tracking-[0.1em] text-white/80 uppercase"
          >
            Most popular
          </span>
        </div>

        <p
          class="mt-1 text-[11px] tracking-[0.02em]"
          :class="plan.featured ? 'text-white/45' : 'text-[#0b1424]/40'"
        >
          {{ plan.tagline }}
        </p>

        <p class="mt-5 flex items-baseline gap-1.5">
          <span
            class="text-[30px] leading-none font-semibold tracking-[-0.03em] tabular-nums"
            :class="plan.featured ? 'text-white' : 'text-[#0b1424]'"
          >
            {{ formatPrice(plan) }}
          </span>
          <span
            v-if="plan.price > 0"
            class="text-[12px]"
            :class="plan.featured ? 'text-white/40' : 'text-[#0b1424]/40'"
          >
            / month
          </span>
        </p>

        <p
          class="mt-3 min-h-[40px] text-[12.5px] leading-[1.55]"
          :class="plan.featured ? 'text-white/55' : 'text-[#0b1424]/55'"
        >
          {{ plan.description }}
        </p>

        <ul
          class="mt-5 space-y-2.5 border-t pt-5"
          :class="plan.featured ? 'border-white/12' : 'border-[#0b1424]/8'"
        >
          <li v-for="feature in plan.features" :key="feature" class="flex gap-2">
            <UiIcon
              name="check"
              :size="13"
              class="mt-0.5 shrink-0"
              :class="plan.featured ? 'text-[#7fb2ff]' : 'text-[#2674da]'"
            />
            <span
              class="text-[12.5px] leading-[1.5]"
              :class="plan.featured ? 'text-white/70' : 'text-[#0b1424]/65'"
            >
              {{ feature }}
            </span>
          </li>
        </ul>

        <div class="mt-6 pt-1">
          <LandingActionLink
            :to="plan.cta.to"
            :variant="plan.featured ? 'light' : 'secondary'"
            class="w-full"
          >
            {{ plan.cta.label }}
          </LandingActionLink>
        </div>
      </div>
    </div>

    <p class="mt-6 max-w-3xl text-[12.5px] leading-[1.6] text-[#0b1424]/45">
      {{ PRICING_NOTE }}
    </p>
  </div>
</template>
