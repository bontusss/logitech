<script setup lang="ts">
const open = ref(false)

const links = [
  { label: 'Platform', to: '#platform' },
  { label: 'How it works', to: '#how-it-works' },
  { label: 'Dispatch', to: '#partial-dispatch' },
  { label: 'Solutions', to: '#solutions' },
  { label: 'Pricing', to: '#pricing' },
  { label: 'FAQ', to: '#faq' },
]

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') open.value = false
}

const route = useRoute()
watch(() => route.fullPath, () => (open.value = false))
onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <header class="sticky top-0 z-50 border-b border-[#0b1424]/8 bg-white/85 backdrop-blur-xl">
    <div class="mx-auto flex h-16 max-w-[1200px] items-center justify-between gap-4 px-5 sm:px-8">
      <a href="#top" class="rounded-lg focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[#2674da]">
        <LandingBrandLogo :size="34" />
      </a>

      <nav class="hidden items-center gap-0.5 lg:flex" aria-label="Primary">
        <a
          v-for="link in links"
          :key="link.to"
          :href="link.to"
          class="rounded-lg px-3 py-2 text-[14px] font-medium text-[#0b1424]/60 transition hover:bg-[#0b1424]/[0.04] hover:text-[#0b1424] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#2674da]"
        >
          {{ link.label }}
        </a>
      </nav>

      <div class="hidden items-center gap-1.5 lg:flex">
        <LandingActionLink to="/sign-in" variant="quiet">Sign in</LandingActionLink>
        <LandingActionLink to="/sign-in">
          Get started free
          <UiIcon name="arrowRight" :size="15" />
        </LandingActionLink>
      </div>

      <button
        type="button"
        class="flex size-10 items-center justify-center rounded-lg border border-[#0b1424]/10 text-[#0b1424] transition hover:bg-[#0b1424]/[0.04] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#2674da] lg:hidden"
        :aria-expanded="open"
        aria-controls="landing-mobile-menu"
        :aria-label="open ? 'Close menu' : 'Open menu'"
        @click="open = !open"
      >
        <UiIcon :name="open ? 'close' : 'menu'" :size="20" />
      </button>
    </div>

    <div v-show="open" id="landing-mobile-menu" class="border-t border-[#0b1424]/8 bg-white lg:hidden">
      <nav class="mx-auto max-w-[1200px] px-5 py-3 sm:px-8" aria-label="Mobile">
        <ul class="divide-y divide-[#0b1424]/6">
          <li v-for="link in links" :key="link.to">
            <a
              :href="link.to"
              class="flex items-center justify-between py-3 text-[15px] font-medium text-[#0b1424]/80 transition hover:text-[#0b1424]"
              @click="open = false"
            >
              {{ link.label }}
              <UiIcon name="chevronRight" :size="16" class="text-[#0b1424]/30" />
            </a>
          </li>
        </ul>
        <div class="flex flex-col gap-2 pt-4 pb-2">
          <LandingActionLink to="/sign-in" size="lg">Get started free</LandingActionLink>
          <LandingActionLink to="/sign-in" variant="secondary" size="lg">Sign in</LandingActionLink>
        </div>
      </nav>
    </div>
  </header>
</template>
