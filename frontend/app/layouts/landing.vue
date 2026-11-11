<script setup lang="ts">
/**
 * Public marketing shell: no app chrome, no session, light-only surfaces.
 * The landing page deliberately sets its own colours instead of the app tokens
 * so a visitor's saved dark theme never leaks into the marketing surfaces.
 */
</script>

<template>
  <div class="landing min-h-dvh bg-white text-[#0b1424]">
    <a
      href="#main"
      class="sr-only focus:not-sr-only focus:fixed focus:top-3 focus:left-3 focus:z-[60] focus:rounded-lg focus:bg-[#0b1424] focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-white"
    >
      Skip to content
    </a>

    <LandingSiteHeader />

    <main id="main">
      <slot />
    </main>

    <LandingSiteFooter />
  </div>
</template>

<style>
.landing {
  /* Section anchors clear the sticky header. */
  scroll-behavior: smooth;
}

.landing [id] {
  scroll-margin-top: 5.5rem;
}

@keyframes bol-rise {
  from {
    opacity: 0;
    transform: translateY(14px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

@keyframes bol-fade {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.landing-reveal {
  animation: bol-rise 0.7s cubic-bezier(0.16, 0.84, 0.44, 1) both;
}

.landing-fade {
  animation: bol-fade 0.9s ease both;
}

@media (prefers-reduced-motion: reduce) {
  .landing {
    scroll-behavior: auto;
  }

  .landing-reveal,
  .landing-fade {
    animation: none;
  }
}

/* A faint grid that reads as engineering paper rather than decoration. */
.landing-grid {
  background-image:
    linear-gradient(to right, rgba(11, 20, 36, 0.045) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(11, 20, 36, 0.045) 1px, transparent 1px);
  background-size: 64px 64px;
}
</style>
