<script setup lang="ts">
import {
  ArrowLeftRight,
  Building2,
  ChevronDown,
  ClipboardList,
  LayoutDashboard,
  LogOut,
  MapPin,
  Menu,
  Palette,
  Settings,
  ShieldCheck,
  UserRound,
  Users,
  X,
} from 'lucide-vue-next'

const route = useRoute()
const { user, logout } = useAuth()
const { companies, currentCompany, selectCompany } = useCompany()
const { detail: workspaceDetail, load: loadWorkspace } = useWorkspace()

const mobileOpen = ref(false)
const companyOpen = ref(false)
const profileOpen = ref(false)

const navigation = [
  { label: 'Overview', to: '/dashboard', icon: LayoutDashboard },
  { label: 'Onboarding', to: '/onboarding', icon: ClipboardList },
]

const companyNavigation = [
  { label: 'Company profile', to: '/dashboard/company', icon: Building2, permission: 'company.view' },
  { label: 'Branches', to: '/dashboard/company/branches', icon: MapPin, permission: 'branches.view' },
  { label: 'Staff', to: '/dashboard/company/staff', icon: Users, permission: 'staff.view' },
  { label: 'Roles & permissions', to: '/dashboard/company/roles', icon: ShieldCheck, permission: 'staff.roles.manage' },
  { label: 'Branding', to: '/dashboard/company/branding', icon: Palette, permission: 'branding.manage' },
  { label: 'Settings', to: '/dashboard/company/settings', icon: Settings, permission: 'company.settings.manage' },
]

const visibleCompanyNavigation = computed(() =>
  companyNavigation.filter((item) => {
    if (!workspaceDetail.value) return true
    return (
      workspaceDetail.value.platform_access ||
      workspaceDetail.value.membership.permissions.includes(item.permission)
    )
  }),
)

const upcoming = [
  { label: 'Shipments', icon: ArrowLeftRight },
  { label: 'Rates', icon: ClipboardList },
]

const initials = computed(() => {
  const name = user.value?.full_name ?? ''
  return (
    name
      .split(' ')
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase())
      .join('') || 'BO'
  )
})

function isActive(to: string) {
  if (to === '/dashboard/company') return route.path === to
  return route.path === to || route.path.startsWith(`${to}/`)
}

const pageTitle = computed(() => {
  if (route.path.startsWith('/onboarding')) return 'Company onboarding'
  const companyItem = companyNavigation.find((item) => isActive(item.to))
  if (companyItem) return companyItem.label
  return 'Workspace overview'
})

onMounted(() => {
  if (currentCompany.value && !workspaceDetail.value) loadWorkspace()
})

function chooseCompany(companyId: string) {
  selectCompany(companyId)
  companyOpen.value = false
  if (import.meta.client) window.location.assign('/dashboard')
}

async function handleLogout() {
  profileOpen.value = false
  await logout()
  await navigateTo('/login')
}

watch(
  () => route.fullPath,
  () => {
    mobileOpen.value = false
    profileOpen.value = false
  },
)
</script>

<template>
  <div class="min-h-dvh bg-canvas text-ink-900">
    <!-- Mobile overlay -->
    <div
      v-if="mobileOpen"
      class="fixed inset-0 z-40 bg-ink-950/40 backdrop-blur-sm lg:hidden"
      @click="mobileOpen = false"
    />

    <aside
      class="fixed inset-y-0 left-0 z-50 flex w-[264px] flex-col border-r border-ink-900/10 bg-white transition-transform duration-200 lg:translate-x-0"
      :class="mobileOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <div class="flex h-16 items-center justify-between border-b border-ink-900/8 px-4">
        <NuxtLink to="/dashboard" aria-label="BOL workspace">
          <AppLogo :size="32" />
        </NuxtLink>
        <button
          type="button"
          class="flex size-8 items-center justify-center rounded-lg text-ink-900/40 hover:bg-ink-900/5 lg:hidden"
          aria-label="Close navigation"
          @click="mobileOpen = false"
        >
          <X :size="18" />
        </button>
      </div>

      <div class="relative px-3 pt-3">
        <button
          type="button"
          class="flex w-full items-center gap-2.5 rounded-lg border border-ink-900/10 bg-canvas px-3 py-2.5 text-left transition hover:border-ink-900/20"
          :aria-expanded="companyOpen"
          @click="companyOpen = !companyOpen"
        >
          <span class="flex size-7 items-center justify-center rounded-md bg-ink-900 text-[10px] font-semibold text-white">
            {{ (currentCompany?.company_name ?? 'B').slice(0, 1).toUpperCase() }}
          </span>
          <span class="min-w-0 flex-1">
            <span class="block truncate text-[12.5px] font-semibold">
              {{ currentCompany?.company_name ?? 'No company selected' }}
            </span>
            <span class="block truncate text-[10px] text-ink-900/40">
              {{ currentCompany?.roles?.[0]?.replace('_', ' ') ?? 'Marketplace account' }}
            </span>
          </span>
          <ChevronDown :size="14" class="shrink-0 text-ink-900/35 transition" :class="companyOpen ? 'rotate-180' : ''" />
        </button>

        <div
          v-if="companyOpen"
          class="absolute inset-x-3 top-[calc(100%+4px)] z-20 overflow-hidden rounded-lg border border-ink-900/10 bg-white shadow-lg"
        >
          <button
            v-for="company in companies"
            :key="company.company_id"
            type="button"
            class="flex w-full items-center justify-between gap-2 px-3 py-2.5 text-left text-[12px] transition hover:bg-canvas"
            @click="chooseCompany(company.company_id)"
          >
            <span class="truncate">{{ company.company_name }}</span>
            <span
              v-if="company.company_id === currentCompany?.company_id"
              class="text-[10px] font-semibold text-brand-600"
            >
              Current
            </span>
          </button>
          <p v-if="companies.length === 0" class="px-3 py-2.5 text-[11.5px] text-ink-900/40">
            No companies yet.
          </p>
        </div>
      </div>

      <nav class="flex-1 overflow-y-auto px-3 py-5" aria-label="Workspace">
        <p class="px-2 text-[9.5px] font-semibold tracking-[0.18em] text-ink-900/35 uppercase">
          Workspace
        </p>
        <ul class="mt-2 space-y-0.5">
          <li v-for="item in navigation" :key="item.to">
            <NuxtLink
              :to="item.to"
              class="flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-[13px] font-medium transition"
              :class="
                route.path === item.to
                  ? 'bg-brand-50 text-brand-700'
                  : 'text-ink-900/60 hover:bg-ink-900/[0.03] hover:text-ink-900'
              "
            >
              <component :is="item.icon" :size="15" />
              {{ item.label }}
            </NuxtLink>
          </li>
        </ul>

        <p class="mt-6 px-2 text-[9.5px] font-semibold tracking-[0.18em] text-ink-900/35 uppercase">
          Company
        </p>
        <ul class="mt-2 space-y-0.5">
          <li v-for="item in visibleCompanyNavigation" :key="item.to">
            <NuxtLink
              :to="item.to"
              class="flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-[13px] font-medium transition"
              :class="
                isActive(item.to)
                  ? 'bg-brand-50 text-brand-700'
                  : 'text-ink-900/60 hover:bg-ink-900/[0.03] hover:text-ink-900'
              "
            >
              <component :is="item.icon" :size="15" />
              {{ item.label }}
            </NuxtLink>
          </li>
        </ul>

        <p class="mt-6 px-2 text-[9.5px] font-semibold tracking-[0.18em] text-ink-900/35 uppercase">
          Coming next
        </p>
        <ul class="mt-2 space-y-0.5">
          <li
            v-for="item in upcoming"
            :key="item.label"
            class="flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-[13px] text-ink-900/30"
          >
            <component :is="item.icon" :size="15" />
            {{ item.label }}
            <span class="ml-auto rounded border border-ink-900/10 px-1.5 py-0.5 text-[8.5px] tracking-wide uppercase">
              Soon
            </span>
          </li>
        </ul>
      </nav>

      <div class="border-t border-ink-900/8 p-3">
        <NuxtLink
          to="/"
          class="flex items-center gap-2 rounded-lg px-2.5 py-2 text-[12px] text-ink-900/45 transition hover:bg-ink-900/[0.03] hover:text-ink-900"
        >
          <UserRound :size="14" />
          View public site
        </NuxtLink>
      </div>
    </aside>

    <div class="lg:pl-[264px]">
      <header class="sticky top-0 z-30 border-b border-ink-900/8 bg-canvas/90 backdrop-blur-xl">
        <div class="flex h-16 items-center gap-3 px-4 sm:px-6">
          <button
            type="button"
            class="flex size-9 items-center justify-center rounded-lg border border-ink-900/10 lg:hidden"
            aria-label="Open navigation"
            @click="mobileOpen = true"
          >
            <Menu :size="18" />
          </button>

          <div class="min-w-0">
            <p class="truncate text-[13px] font-semibold tracking-[-0.01em]">{{ pageTitle }}</p>
            <p class="truncate text-[10.5px] text-ink-900/40">
              {{ currentCompany?.company_name ?? 'BOL account' }}
            </p>
          </div>

          <ThemeToggle class="ml-auto" />

          <div class="relative">
            <button
              type="button"
              class="flex items-center gap-2 rounded-lg border border-ink-900/10 bg-white px-2 py-1.5 transition hover:border-ink-900/20"
              :aria-expanded="profileOpen"
              @click="profileOpen = !profileOpen"
            >
              <span class="flex size-7 items-center justify-center rounded-md bg-ink-900 text-[10px] font-semibold text-white">
                {{ initials }}
              </span>
              <span class="hidden text-left sm:block">
                <span class="block max-w-[160px] truncate text-[11.5px] font-semibold">
                  {{ user?.full_name }}
                </span>
                <span class="block max-w-[160px] truncate text-[9.5px] text-ink-900/40">
                  {{ user?.email }}
                </span>
              </span>
              <ChevronDown :size="13" class="text-ink-900/35" />
            </button>

            <div
              v-if="profileOpen"
              class="absolute right-0 top-[calc(100%+6px)] z-30 w-56 overflow-hidden rounded-lg border border-ink-900/10 bg-white shadow-lg"
            >
              <div class="border-b border-ink-900/8 px-3 py-2.5">
                <p class="truncate text-[12px] font-semibold">{{ user?.full_name }}</p>
                <p class="truncate text-[10.5px] text-ink-900/40">{{ user?.email }}</p>
                <p class="mt-1 inline-flex rounded bg-ink-900/5 px-1.5 py-0.5 text-[9px] font-medium capitalize">
                  {{ user?.account_type }} account
                </p>
              </div>
              <button
                type="button"
                class="flex w-full items-center gap-2 px-3 py-2.5 text-left text-[12px] text-ink-900/65 transition hover:bg-canvas"
                @click="handleLogout"
              >
                <LogOut :size="14" />
                Sign out
              </button>
            </div>
          </div>
        </div>
      </header>

      <main class="px-4 py-6 sm:px-6 sm:py-8">
        <div class="mx-auto max-w-[1000px]">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>
