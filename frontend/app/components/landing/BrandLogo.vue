<script setup lang="ts">
/**
 * Interim brand mark for based on logistics (BOL).
 *
 * The mark is a drawn "B" set in a navy tile with a single accent node — the
 * parcel sitting on the route. Replace this file (and the ids in the template)
 * when the real logo arrives; every surface on the landing page reads from here.
 */
withDefaults(
  defineProps<{
    size?: number
    wordmark?: boolean
    /** 'dark' for light backgrounds, 'light' for the navy footer and CTA. */
    tone?: 'dark' | 'light'
  }>(),
  { size: 36, wordmark: true, tone: 'dark' },
)

// Gradient ids must be unique: the same logo renders in the header and footer.
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
        <linearGradient :id="`bol-tile-${uid}`" x1="4" y1="0" x2="36" y2="40" gradientUnits="userSpaceOnUse">
          <stop stop-color="#1c2f52" />
          <stop offset="1" stop-color="#07101f" />
        </linearGradient>
      </defs>

      <rect width="40" height="40" rx="11" :fill="`url(#bol-tile-${uid})`" />
      <rect x="0.6" y="0.6" width="38.8" height="38.8" rx="10.4" stroke="#ffffff" stroke-opacity="0.14" stroke-width="1.2" />

      <g stroke="#f7f9ff" stroke-width="2.7" stroke-linecap="round" stroke-linejoin="round">
        <path d="M13.6 11.2v17.6" />
        <path d="M13.6 11.2h6.2a3.9 3.9 0 0 1 0 7.8h-6.2" />
        <path d="M13.6 19h7a4.9 4.9 0 0 1 0 9.8h-7" />
      </g>

      <circle cx="30" cy="12.4" r="2.6" fill="#5b9dff" />
    </svg>

    <span v-if="wordmark" class="flex flex-col leading-none">
      <span
        class="text-[15px] font-semibold tracking-[-0.015em]"
        :class="tone === 'light' ? 'text-white' : 'text-[#0b1424]'"
      >
        based on logistics
      </span>
      <span
        class="mt-1.5 text-[9px] font-semibold tracking-[0.3em] uppercase"
        :class="tone === 'light' ? 'text-white/45' : 'text-[#0b1424]/40'"
      >
        BOL
      </span>
    </span>
  </span>
</template>
