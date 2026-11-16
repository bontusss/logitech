<script setup lang="ts">
withDefaults(
  defineProps<{
    size?: number
    tone?: 'dark' | 'light'
    wordmark?: boolean
  }>(),
  { size: 36, tone: 'dark', wordmark: true },
)

// The mark appears more than once per page, so gradient ids must be unique.
const uid = useId()
</script>

<template>
  <span class="inline-flex items-center gap-2.5 select-none">
    <svg
      :width="size"
      :height="size"
      viewBox="0 0 40 40"
      fill="none"
      role="img"
      aria-label="based on logistics"
      class="shrink-0"
    >
      <defs>
        <linearGradient
          :id="`bol-tile-${uid}`"
          x1="4"
          y1="0"
          x2="36"
          y2="40"
          gradientUnits="userSpaceOnUse"
        >
          <stop stop-color="#22346b" />
          <stop offset="1" stop-color="#0a1220" />
        </linearGradient>
      </defs>

      <rect width="40" height="40" rx="10" :fill="`url(#bol-tile-${uid})`" />
      <rect
        x="0.6"
        y="0.6"
        width="38.8"
        height="38.8"
        rx="9.4"
        stroke="#ffffff"
        stroke-opacity="0.14"
      />

      <g stroke="#ffffff" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12.2 11v18" />
        <path d="M12.2 11h6.4a3.6 3.6 0 0 1 0 7.2h-6.4" />
        <path d="M12.2 18.2h7.1a4.5 4.5 0 0 1 0 9h-7.1" />
      </g>

      <circle cx="29.4" cy="11.6" r="2.6" fill="#7d9bff" />
      <path d="M18.6 33.4h8.2" stroke="#7d9bff" stroke-width="2" stroke-linecap="round" />
    </svg>

    <span v-if="wordmark" class="flex flex-col leading-none">
      <span
        class="text-[15px] font-semibold tracking-[-0.02em]"
        :class="tone === 'light' ? 'text-white' : 'text-ink-900'"
      >
        based on logistics
      </span>
      <span
        class="mt-1.5 text-[9px] font-semibold tracking-[0.32em] uppercase"
        :class="tone === 'light' ? 'text-white/40' : 'text-ink-900/40'"
      >
        BOL
      </span>
    </span>
  </span>
</template>
