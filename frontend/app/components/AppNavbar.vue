<script setup lang="ts">
import { Menu, X, ArrowRight } from 'lucide-vue-next'

const open = ref(false)
const route = useRoute()
const { demoUrl } = useSiteConfig()

const links = [
  { label: 'Marketplace', to: '#marketplace' },
  { label: 'For Logistics Businesses', to: '#business' },
  { label: 'Features', to: '#features' },
  { label: 'How It Works', to: '#how-it-works' },
  { label: 'Pricing', to: '#pricing' },
]

watch(
  () => route.fullPath,
  () => {
    open.value = false
  },
)

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') open.value = false
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <header class="sticky top-0 z-50 border-b border-ink-900/8 bg-canvas/85 backdrop-blur-xl">
    <div class="mx-auto flex h-16 max-w-[1200px] items-center justify-between gap-4 px-5 sm:px-8">
      <a href="#top" class="rounded-lg" aria-label="based on logistics — back to top">
        <AppLogo :size="34" />
      </a>

      <nav class="hidden items-center gap-0.5 lg:flex" aria-label="Primary">
        <a
          v-for="link in links"
          :key="link.to"
          :href="link.to"
          class="rounded-lg px-3 py-2 text-[13.5px] font-medium text-ink-900/60 transition hover:bg-ink-900/4 hover:text-ink-900"
        >
          {{ link.label }}
        </a>
      </nav>

      <div class="hidden items-center gap-1.5 lg:flex">
        <AppButton to="/sign-in" variant="ghost">Sign in</AppButton>
        <AppButton :to="demoUrl">
          Book a Demo
          <ArrowRight :size="15" />
        </AppButton>
      </div>

      <button
        type="button"
        class="flex size-10 items-center justify-center rounded-lg border border-ink-900/10 text-ink-900 transition hover:bg-ink-900/4 lg:hidden"
        :aria-expanded="open"
        aria-controls="mobile-menu"
        :aria-label="open ? 'Close menu' : 'Open menu'"
        @click="open = !open"
      >
        <X v-if="open" :size="20" />
        <Menu v-else :size="20" />
      </button>
    </div>

    <div v-show="open" id="mobile-menu" class="border-t border-ink-900/8 bg-canvas lg:hidden">
      <nav class="mx-auto max-w-[1200px] px-5 py-3 sm:px-8" aria-label="Mobile">
        <ul class="divide-y divide-ink-900/6">
          <li v-for="link in links" :key="link.to">
            <a
              :href="link.to"
              class="flex items-center justify-between py-3 text-[15px] font-medium text-ink-900/80"
              @click="open = false"
            >
              {{ link.label }}
              <ArrowRight :size="15" class="text-ink-900/30" />
            </a>
          </li>
        </ul>
        <div class="flex flex-col gap-2 pt-4 pb-2">
          <AppButton :to="demoUrl" size="lg" block>Book a Demo</AppButton>
          <AppButton to="/sign-in" variant="secondary" size="lg" block>Sign in</AppButton>
        </div>
      </nav>
    </div>
  </header>
</template>
