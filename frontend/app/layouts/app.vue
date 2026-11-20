<script setup lang="ts">
import {
  Building2,
  CalendarCheck,
  ChartNoAxesColumnIncreasing,
  ChevronDown,
  ClipboardList,
  CreditCard,
  LayoutDashboard,
  LogOut,
  Menu,
  Palette,
  Send,
  Settings,
  Truck,
  UserRound,
  Users,
  X,
} from 'lucide-vue-next'
import type { Component } from 'vue'

const route = useRoute()
const { user, logout } = useAuth()
const { companies, currentCompany, selectCompany } = useCompany()
const { detail: workspaceDetail, load: loadWorkspace } = useWorkspace()

const mobileOpen = ref(false)
const companyOpen = ref(false)
const profileOpen = ref(false)

const isCustomer = computed(() => user.value?.account_type === 'customer')
const isVerified = computed(
  () => (workspaceDetail.value?.company.status ?? currentCompany.value?.status) === 'verified',
)

type NavigationTone = 'violet' | 'sky' | 'mint' | 'amber' | 'rose' | 'cyan'

interface NavigationItem {
  label: string
  to: string
  icon: Component
  tone: NavigationTone
  comingSoon?: boolean
  permission?: string
}

interface NavigationGroup {
  label: string
  items: NavigationItem[]
}

const navigationGroups = computed<NavigationGroup[]>(() => {
  if (isCustomer.value) {
    return [{
      label: 'Workspace',
      items: [
        { label: 'Overview', to: '/dashboard', icon: LayoutDashboard, tone: 'violet' },
        { label: 'Send a package', to: '/#rate-search', icon: Send, tone: 'sky' },
      ],
    }]
  }
  if (!isVerified.value) {
    return [{
      label: 'Workspace',
      items: [
        { label: 'Overview', to: '/dashboard', icon: LayoutDashboard, tone: 'violet' },
        { label: 'Onboarding', to: '/onboarding', icon: ClipboardList, tone: 'amber' },
      ],
    }]
  }

  const groups: NavigationGroup[] = [
    {
      label: 'Workspace',
      items: [{ label: 'Overview', to: '/dashboard', icon: LayoutDashboard, tone: 'violet' }],
    },
    {
      label: 'Operations',
      items: [
        { label: 'Deliveries', to: '/dashboard/deliveries', icon: Truck, tone: 'sky', permission: 'shipments.view' },
        { label: 'Customers', to: '/dashboard/coming-soon/customers', icon: Users, tone: 'rose', comingSoon: true, permission: 'staff.view' },
        { label: 'Requests', to: '/dashboard/requests', icon: ClipboardList, tone: 'amber', permission: 'bookings.manage' },
        { label: 'Analytics', to: '/dashboard/coming-soon/analytics', icon: ChartNoAxesColumnIncreasing, tone: 'mint', comingSoon: true, permission: 'reports.view' },
        { label: 'Finance', to: '/dashboard/coming-soon/finance', icon: CreditCard, tone: 'cyan', comingSoon: true, permission: 'reports.view' },
        { label: 'Drivers', to: '/dashboard/company/drivers', icon: UserRound, tone: 'violet', permission: 'staff.view' },
      ],
    },
    {
      label: 'Setup',
      items: [
        { label: 'Booking page', to: '/dashboard/company/booking', icon: CalendarCheck, tone: 'mint', permission: 'bookings.manage' },
        { label: 'Rates & Coverage', to: '/dashboard/company/rates', icon: ClipboardList, tone: 'sky', permission: 'rates.view' },
      ],
    },
    {
      label: 'Settings',
      items: [
        { label: 'Profile', to: '/dashboard/company', icon: Building2, tone: 'violet', permission: 'company.view' },
        { label: 'Branding', to: '/dashboard/company/branding', icon: Palette, tone: 'rose', permission: 'branding.manage' },
        { label: 'Billing', to: '/dashboard/coming-soon/billing', icon: Settings, tone: 'cyan', comingSoon: true, permission: 'company.settings.manage' },
      ],
    },
  ]
  return groups.map((group) => ({
    ...group,
    items: group.items.filter((item) =>
      !item.permission ||
      !workspaceDetail.value ||
      workspaceDetail.value.platform_access ||
      workspaceDetail.value.membership.permissions.includes(item.permission),
    ),
  }))
})

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
  if (to === '/dashboard') return route.path === to
  if (to === '/dashboard/company') return route.path === to
  return route.path === to || route.path.startsWith(`${to}/`)
}

const pageTitle = computed(() => {
  if (isCustomer.value) return 'Marketplace dashboard'
  if (route.path.startsWith('/onboarding')) return 'Company onboarding'
  const activeItem = navigationGroups.value.flatMap((group) => group.items).find((item) => isActive(item.to))
  if (activeItem) return activeItem.label
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
      class="app-sidebar fixed inset-y-0 left-0 z-50 flex w-[264px] flex-col border-r transition-transform duration-200 lg:translate-x-0"
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

      <div v-if="!isCustomer" class="relative px-3 pt-3">
        <button
          type="button"
          class="sidebar-company flex w-full items-center gap-2.5 rounded-xl border px-3 py-2.5 text-left transition hover:shadow-sm"
          :aria-expanded="companyOpen"
          @click="companyOpen = !companyOpen"
        >
          <span class="flex size-8 items-center justify-center rounded-lg bg-gradient-to-br from-brand-500 to-violet-500 text-[11px] font-semibold text-white shadow-sm shadow-brand-600/30">
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
        <section v-for="group in navigationGroups" :key="group.label">
          <p
            class="flex items-center gap-2 px-2 text-[10px] font-semibold tracking-[0.16em] text-ink-900/40 uppercase"
            :class="group.label !== 'Workspace' ? 'mt-6' : ''"
          >
            <span class="h-px w-3 rounded-full bg-brand-400/60" />
            {{ group.label }}
          </p>
          <ul class="mt-2 space-y-0.5">
            <li v-for="item in group.items" :key="item.to">
              <NuxtLink
                :to="item.to"
                class="sidebar-link flex items-center gap-2.5 rounded-xl px-2 py-1.5 text-[13px] font-medium"
                :class="isActive(item.to) ? 'sidebar-link--active' : 'text-ink-900/65 hover:bg-white/80 hover:text-ink-900 dark:hover:bg-white/5'"
              >
                <span class="sidebar-icon" :class="`sidebar-icon--${item.tone}`">
                  <component :is="item.icon" :size="14" />
                </span>
                <span class="min-w-0 flex-1 truncate">{{ item.label }}</span>
                <span
                  v-if="item.comingSoon"
                  class="ml-auto shrink-0 rounded-full border border-amber-200/80 bg-amber-50 px-1.5 py-0.5 text-[8.5px] font-semibold tracking-wide text-amber-700 uppercase"
                >
                  Soon
                </span>
              </NuxtLink>
            </li>
          </ul>
        </section>
      </nav>

      <div class="border-t border-ink-900/8 p-3">
        <NuxtLink
          to="/"
          class="sidebar-footer-link flex items-center gap-2 rounded-xl border border-ink-900/8 bg-white/70 px-2.5 py-2 text-[12px] text-ink-900/55 hover:text-ink-900 dark:bg-white/5"
        >
          <span class="flex size-6 items-center justify-center rounded-lg bg-brand-50 text-brand-700 dark:bg-white/10">
            <UserRound :size="12" />
          </span>
          View public site
        </NuxtLink>
      </div>
    </aside>

    <div class="lg:pl-[264px]">
      <header class="app-topbar sticky top-0 z-30 border-b backdrop-blur-xl">
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
              {{ isCustomer ? 'Customer account' : (currentCompany?.company_name ?? 'BOL account') }}
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
              <span class="flex size-7 items-center justify-center rounded-lg bg-gradient-to-br from-brand-500 to-fuchsia-500 text-[10px] font-semibold text-white shadow-sm shadow-brand-600/25">
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

      <main class="app-main px-4 py-6 sm:px-6 sm:py-8">
        <div class="mx-auto max-w-[1000px]">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>
