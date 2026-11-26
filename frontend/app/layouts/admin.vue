<script setup lang="ts">
import { LogOut, ShieldCheck } from 'lucide-vue-next'

const { user, logout } = useAuth()

async function handleLogout() {
  await logout()
  await navigateTo('/admin/login')
}

const initials = computed(() => {
  const name = user.value?.full_name ?? ''
  return (
    name
      .split(' ')
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase())
      .join('') || 'AD'
  )
})
</script>

<template>
  <div class="admin-shell min-h-dvh text-ink-900">
    <header class="app-topbar sticky top-0 z-30 border-b backdrop-blur-xl">
      <div class="mx-auto flex h-16 max-w-[1180px] items-center gap-3 px-4 sm:px-6">
        <NuxtLink to="/admin" class="rounded-lg" aria-label="BOL platform administration">
          <AppLogo :size="32" />
        </NuxtLink>
        <span class="hidden items-center gap-1.5 rounded-full bg-accent-500 px-2.5 py-1 text-[10px] font-semibold tracking-[0.06em] text-white uppercase shadow-sm shadow-accent-500/25 sm:inline-flex">
          <ShieldCheck :size="12" />
          Platform admin
        </span>

        <div class="ml-auto flex items-center gap-2.5">
          <div class="hidden text-right sm:block">
            <p class="max-w-[180px] truncate text-[11.5px] font-semibold">{{ user?.full_name }}</p>
            <p class="text-[10px] text-ink-900/45 capitalize">
              {{ user?.platform_role?.replace('_', ' ') }}
            </p>
          </div>
          <span class="flex size-8 items-center justify-center rounded-xl bg-accent-500 text-[10px] font-semibold text-white shadow-sm shadow-accent-500/25">
            {{ initials }}
          </span>
          <ThemeToggle />
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-xl border border-ink-900/10 bg-surface px-2.5 py-2 text-[11.5px] font-medium shadow-sm transition hover:border-accent-500/40 hover:text-ink-900 dark:bg-white/5"
            @click="handleLogout"
          >
            <LogOut :size="14" />
            <span class="hidden sm:inline">Sign out</span>
          </button>
        </div>
      </div>
    </header>

    <main class="relative mx-auto max-w-[1180px] px-4 py-6 sm:px-6 sm:py-8">
      <slot />
    </main>
  </div>
</template>
