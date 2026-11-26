<script setup lang="ts">
withDefaults(
  defineProps<{
    size?: number
    tone?: 'dark' | 'light'
    wordmark?: boolean
  }>(),
  { size: 36, tone: 'dark', wordmark: true },
)
</script>

<template>
  <span
    class="inline-flex shrink-0 items-center select-none"
    role="img"
    aria-label="based on logistics (BOL)"
  >
    <img
      v-if="!wordmark"
      src="/brand/mark.png"
      :width="size"
      :height="size"
      alt=""
      class="shrink-0 rounded-full"
    />
    <img
      v-else-if="tone === 'light'"
      src="/brand/logo-white.png"
      alt="BOL"
      class="w-auto shrink-0"
      :style="{ height: `${size}px` }"
    />
    <template v-else>
      <img
        src="/brand/logo-navy.png"
        alt="BOL"
        class="brand-logo-light w-auto shrink-0"
        :style="{ height: `${size}px` }"
      />
      <img
        src="/brand/logo-white.png"
        alt="BOL"
        class="brand-logo-dark w-auto shrink-0"
        :style="{ height: `${size}px` }"
      />
    </template>
  </span>
</template>
