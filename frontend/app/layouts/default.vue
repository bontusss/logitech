<script setup lang="ts">
const route = useRoute()
const { user, logout } = useAuth()
const { sections, mobileItems } = useNav()
const { name: branchName, load: loadBranches } = useBranches()
const signingOut = ref(false)

const initials = computed(() => {
  const parts = (user.value?.full_name ?? '').trim().split(/\s+/)
  return `${parts[0]?.[0] ?? ''}${parts[1]?.[0] ?? ''}`.toUpperCase()
})

const roleLabel = computed(() => {
  if (!user.value) return ''
  const role = user.value.role === 'system_admin' ? 'System admin' : user.value.role
  const job = user.value.job_role ? ` · ${user.value.job_role}` : ''
  return `${role}${job}`
})

const branch = computed(() => branchName(user.value?.branch_id ?? ''))

function isActive(to: string): boolean {
  if (to === '/') return route.path === '/'
  return route.path === to || route.path.startsWith(`${to}/`)
}

async function signOut() {
  signingOut.value = true
  try {
    await logout()
    await navigateTo('/sign-in')
  } finally {
    signingOut.value = false
  }
}

onMounted(() => {
  if (user.value?.branch_id) void loadBranches()
})
watch(user, (next) => {
  if (next?.branch_id) void loadBranches()
})
</script>

<template>
  <div class="min-h-dvh bg-canvas">
    <!-- Desktop sidebar -->
    <aside class="fixed inset-y-0 left-0 z-30 hidden w-[248px] flex-col border-r border-line bg-surface px-4 py-5 md:flex">
      <NuxtLink to="/" class="flex items-center gap-2 px-1 text-h5 font-semibold text-ink">
        <span class="flex size-8 items-center justify-center rounded-lg bg-ink text-white">
          <UiIcon name="package" :size="17" />
        </span>
        Galopin
      </NuxtLink>

      <nav class="mt-7 flex-1 space-y-6 overflow-y-auto" aria-label="Main">
        <div v-for="section in sections" :key="section.title">
          <p class="px-2 text-meta font-medium tracking-[0.08em] text-ink-40 uppercase">{{ section.title }}</p>
          <ul class="mt-1.5 space-y-0.5">
            <li v-for="item in section.items" :key="item.to">
              <NuxtLink
                :to="item.to"
                class="flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-label transition"
                :class="isActive(item.to) ? 'bg-canvas font-medium text-ink shadow-inset' : 'text-ink-64 hover:bg-canvas/70 hover:text-ink'"
              >
                <UiIcon :name="item.icon" :size="17" />
                {{ item.label }}
              </NuxtLink>
            </li>
          </ul>
        </div>
      </nav>

      <div class="mt-4 border-t border-line pt-4">
        <div class="flex items-center gap-2.5">
          <span class="flex size-9 shrink-0 items-center justify-center rounded-full bg-canvas text-label font-medium text-ink">
            {{ initials }}
          </span>
          <div class="min-w-0 flex-1">
            <p class="truncate text-label font-medium text-ink">{{ user?.full_name }}</p>
            <p class="truncate text-meta text-ink-64">{{ roleLabel }}</p>
            <p v-if="branch" class="truncate text-meta text-ink-40">{{ branch }}</p>
          </div>
          <button
            type="button"
            class="flex size-8 shrink-0 items-center justify-center rounded-lg text-ink-64 transition hover:bg-canvas hover:text-ink"
            aria-label="Sign out"
            title="Sign out"
            :disabled="signingOut"
            @click="signOut"
          >
            <UiIcon name="logout" :size="16" />
          </button>
        </div>
      </div>
    </aside>

    <!-- Mobile top bar -->
    <header class="sticky top-0 z-30 flex h-14 items-center justify-between gap-3 border-b border-line bg-surface/95 px-4 backdrop-blur md:hidden">
      <NuxtLink to="/" class="flex items-center gap-2 text-p1 font-semibold text-ink">
        <span class="flex size-7 items-center justify-center rounded-lg bg-ink text-white">
          <UiIcon name="package" :size="15" />
        </span>
        Galopin
      </NuxtLink>
      <div class="flex min-w-0 items-center gap-2">
        <div class="min-w-0 text-right">
          <p class="truncate text-label font-medium text-ink">{{ user?.full_name }}</p>
          <p class="truncate text-meta text-ink-40">{{ branch || roleLabel }}</p>
        </div>
        <button
          type="button"
          class="flex size-9 shrink-0 items-center justify-center rounded-lg text-ink-64 transition hover:bg-canvas hover:text-ink"
          aria-label="Sign out"
          :disabled="signingOut"
          @click="signOut"
        >
          <UiIcon name="logout" :size="17" />
        </button>
      </div>
    </header>

    <main class="mx-auto w-full max-w-[1180px] px-4 pt-5 pb-24 md:pl-[272px] md:pr-8 md:pt-8 md:pb-10">
      <slot />
    </main>

    <!-- Mobile bottom navigation -->
    <nav class="fixed inset-x-0 bottom-0 z-30 border-t border-line bg-surface/95 pb-[env(safe-area-inset-bottom)] backdrop-blur md:hidden" aria-label="Main">
      <div class="mx-auto flex max-w-[480px] items-stretch justify-around">
        <NuxtLink
          v-for="item in mobileItems"
          :key="item.to"
          :to="item.to"
          class="flex flex-1 flex-col items-center gap-1 py-2.5 text-[11px] transition"
          :class="isActive(item.to) ? 'text-blue' : 'text-ink-64'"
        >
          <UiIcon :name="item.icon" :size="20" />
          {{ item.label }}
        </NuxtLink>
      </div>
    </nav>
  </div>
</template>
