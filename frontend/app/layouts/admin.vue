<script setup lang="ts">
import { LogOut, ShieldCheck } from 'lucide-vue-next'

const { user, logout } = useAuth()

async function handleLogout() {
  await logout()
  await navigateTo('/admin/login')
}
</script>

<template>
  <div class="min-h-dvh bg-canvas text-ink-900">
    <header class="sticky top-0 z-30 border-b border-ink-900/8 bg-canvas/90 backdrop-blur-xl">
      <div class="mx-auto flex h-16 max-w-[1180px] items-center gap-3 px-4 sm:px-6">
        <NuxtLink to="/admin" aria-label="BOL platform administration">
          <AppLogo :size="32" />
        </NuxtLink>
        <span class="hidden items-center gap-1.5 rounded-md border border-ink-900/10 bg-white px-2 py-1 text-[10px] font-semibold tracking-[0.08em] uppercase sm:inline-flex">
          <ShieldCheck :size="12" class="text-brand-600" />
          Platform admin
        </span>

        <div class="ml-auto flex items-center gap-3">
          <div class="hidden text-right sm:block">
            <p class="max-w-[180px] truncate text-[11.5px] font-semibold">{{ user?.full_name }}</p>
            <p class="text-[10px] text-ink-900/40 capitalize">
              {{ user?.platform_role?.replace('_', ' ') }}
            </p>
          </div>
          <ThemeToggle />
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-lg border border-ink-900/10 bg-white px-2.5 py-2 text-[11.5px] font-medium transition hover:border-ink-900/20"
            @click="handleLogout"
          >
            <LogOut :size="14" />
            <span class="hidden sm:inline">Sign out</span>
          </button>
        </div>
      </div>
    </header>

    <main class="mx-auto max-w-[1180px] px-4 py-6 sm:px-6 sm:py-8">
      <slot />
    </main>
  </div>
</template>
