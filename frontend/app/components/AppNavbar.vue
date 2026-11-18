<script setup lang="ts">
import { ArrowRight, Menu, X } from 'lucide-vue-next'

const open = ref(false)
const route = useRoute()

const links = [
  { label: 'Send a package', to: '/#rate-search' },
  { label: 'How it works', to: '/#how-it-works' },
  { label: 'For logistics businesses', to: '/for-logistics-businesses' },
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
    <div class="mx-auto flex h-[68px] max-w-[1200px] items-center justify-between gap-4 px-5 sm:px-8">
      <NuxtLink to="/" class="rounded-lg" aria-label="based on logistics — home">
        <AppLogo :size="34" />
      </NuxtLink>

      <nav class="hidden items-center gap-1 lg:flex" aria-label="Primary">
        <NuxtLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="rounded-lg px-3 py-2 text-[13px] font-medium text-ink-900/60 transition hover:bg-ink-900/4 hover:text-ink-900"
        >
          {{ link.label }}
        </NuxtLink>
      </nav>

      <div class="hidden items-center gap-2 lg:flex">
        <ThemeToggle />
        <AppButton to="/login" variant="ghost">Sign in</AppButton>
        <AppButton to="/#rate-search">
          Compare rates
          <ArrowRight :size="15" />
        </AppButton>
      </div>

      <div class="flex items-center gap-2 lg:hidden">
        <ThemeToggle />
        <button
          type="button"
          class="flex size-10 items-center justify-center rounded-xl border border-ink-900/10 text-ink-900 transition hover:bg-ink-900/4"
          :aria-expanded="open"
          aria-controls="mobile-menu"
          :aria-label="open ? 'Close menu' : 'Open menu'"
          @click="open = !open"
        >
          <X v-if="open" :size="20" />
          <Menu v-else :size="20" />
        </button>
      </div>
    </div>

    <div v-show="open" id="mobile-menu" class="border-t border-ink-900/8 bg-canvas lg:hidden">
      <nav class="mx-auto max-w-[1200px] px-5 py-3 sm:px-8" aria-label="Mobile">
        <ul class="divide-y divide-ink-900/6">
          <li v-for="link in links" :key="link.to">
            <NuxtLink
              :to="link.to"
              class="flex items-center justify-between py-3 text-[15px] font-medium text-ink-900/80"
              @click="open = false"
            >
              {{ link.label }}
              <ArrowRight :size="15" class="text-ink-900/30" />
            </NuxtLink>
          </li>
        </ul>
        <div class="flex flex-col gap-2 pt-4 pb-2">
          <AppButton to="/#rate-search" size="lg" block>Compare rates</AppButton>
          <AppButton to="/login" variant="secondary" size="lg" block>Sign in</AppButton>
        </div>
      </nav>
    </div>
  </header>
</template>
