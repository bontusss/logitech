<script setup lang="ts">
import {
  AlertCircle,
  Check,
  Copy,
  ExternalLink,
  Loader2,
  Pencil,
  Plus,
  Search,
  Share2,
  UserRound,
  X,
} from 'lucide-vue-next'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Drivers' })

interface Driver {
  id: string
  full_name: string
  phone: string
  vehicle_plate: string
  email: string
  status: 'active' | 'suspended' | 'inactive'
  portal_url: string
}

const { api } = useApi()
const { load, can, notify, confirm, loading: workspaceLoading } = useWorkspace()
const route = useRoute()

const drivers = ref<Driver[]>([])
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const formOpen = ref(false)
const editingDriverId = ref('')
const newPortalURL = ref('')
const emailSent = ref(false)
const copiedPortalId = ref('')
const searchQuery = ref('')
const canEdit = computed(() => can('staff.update'))
const canManageStatus = computed(() => can('staff.deactivate'))

const form = reactive({
  full_name: '',
  phone: '',
  vehicle_plate: '',
  email: '',
})

onMounted(async () => {
  await load()
  if (route.query.create === '1' && can('staff.invite')) {
    openCreateForm()
  }
  await fetchDrivers()
  window.addEventListener('keydown', onModalKeydown)
})

onBeforeUnmount(() => window.removeEventListener('keydown', onModalKeydown))

function closeCreateForm() {
  if (!saving.value) {
    formOpen.value = false
    editingDriverId.value = ''
    error.value = ''
  }
}

function onModalKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') closeCreateForm()
}

function openCreateForm() {
  editingDriverId.value = ''
  Object.assign(form, { full_name: '', phone: '', vehicle_plate: '', email: '' })
  error.value = ''
  formOpen.value = true
}

function openEditForm(driver: Driver) {
  editingDriverId.value = driver.id
  Object.assign(form, {
    full_name: driver.full_name,
    phone: driver.phone,
    vehicle_plate: driver.vehicle_plate,
    email: driver.email,
  })
  error.value = ''
  formOpen.value = true
}

const counts = computed(() => ({
  total: drivers.value.length,
  available: drivers.value.filter((driver) => driver.status === 'active').length,
  assigned: 0,
  attention: drivers.value.filter((driver) => driver.status !== 'active').length,
}))

const filteredDrivers = computed(() => {
  const query = searchQuery.value.trim().toLocaleLowerCase()
  if (!query) return drivers.value
  return drivers.value.filter((driver) =>
    [driver.full_name, driver.email, driver.phone, driver.vehicle_plate, driver.status]
      .some((value) => value.toLocaleLowerCase().includes(query)),
  )
})

async function fetchDrivers() {
  loading.value = true
  error.value = ''
  try {
    drivers.value = (await api<{ drivers: Driver[] }>('/companies/me/drivers')).drivers
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not load drivers.')
  } finally {
    loading.value = false
  }
}

async function saveDriver() {
  error.value = ''
  saving.value = true
  try {
    const body = {
      full_name: form.full_name.trim(),
      phone: form.phone.trim(),
      vehicle_plate: form.vehicle_plate.trim(),
      email: form.email.trim(),
    }
    if (editingDriverId.value) {
      await api(`/companies/me/drivers/${editingDriverId.value}`, { method: 'PATCH', body })
      notify('Driver details updated.')
    } else {
      const result = await api<{ driver: Driver; portal_url: string; email_sent: boolean }>(
        '/companies/me/drivers',
        { method: 'POST', body },
      )
      newPortalURL.value = result.portal_url
      emailSent.value = result.email_sent
      notify(result.email_sent ? 'Driver created and welcome email sent.' : 'Driver created; welcome email could not be sent.')
    }
    formOpen.value = false
    editingDriverId.value = ''
    await fetchDrivers()
  } catch (err) {
    error.value = apiErrorMessage(err, editingDriverId.value ? 'Could not update this driver.' : 'Could not create this driver.')
  } finally {
    saving.value = false
  }
}

async function copyPortalLink(url: string, driverId: string) {
  if (!url) {
    error.value = 'This driver does not have a portal link yet.'
    return
  }
  try {
    await navigator.clipboard.writeText(url)
    copiedPortalId.value = driverId
    window.setTimeout(() => {
      if (copiedPortalId.value === driverId) copiedPortalId.value = ''
    }, 1800)
  } catch {
    error.value = 'Could not copy the portal link. Select and copy the link directly.'
  }
}

async function sharePortalLink(driver: Driver) {
  if (!driver.portal_url) return
  const shareData = {
    title: `${driver.full_name}'s driver portal`,
    text: `Open your delivery portal for ${driver.full_name}.`,
    url: driver.portal_url,
  }
  if (navigator.share) {
    try {
      await navigator.share(shareData)
      return
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') return
    }
  }
  await copyPortalLink(driver.portal_url, driver.id)
  notify('Portal link copied — paste it into a message to share.')
}

async function setDriverStatus(driver: Driver, nextStatus: Driver['status']) {
  if (nextStatus === driver.status) return
  if (nextStatus !== 'active') {
    const approved = await confirm({
      title: nextStatus === 'inactive' ? 'Deactivate driver?' : 'Suspend driver portal?',
      message: nextStatus === 'inactive'
        ? `${driver.full_name} will be removed from active assignment and cannot open their portal. You can reactivate them later.`
        : `${driver.full_name} will temporarily lose portal access until you reactivate them.`,
      confirmLabel: nextStatus === 'inactive' ? 'Deactivate driver' : 'Suspend portal',
      danger: true,
    })
    if (!approved) return
  }
  try {
    await api(`/companies/me/drivers/${driver.id}/status`, {
      method: 'PATCH',
      body: { status: nextStatus },
    })
    notify(nextStatus === 'active' ? 'Driver reactivated.' : nextStatus === 'inactive' ? 'Driver deactivated.' : 'Driver portal suspended.')
    await fetchDrivers()
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not update the driver.')
  }
}
</script>

<template>
  <div class="driver-workspace dash-page">
    <div class="driver-page-hero flex flex-wrap items-center justify-between gap-4 rounded-2xl border p-5 sm:p-6">
      <div>
        <p class="inline-flex items-center gap-1.5 rounded-full bg-white/70 px-2.5 py-1 text-[10px] font-semibold tracking-wide text-brand-700 dark:bg-slate-950/30 dark:text-indigo-200">YOUR PEOPLE, ON THE MOVE</p>
        <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">Drivers</h1>
        <p class="mt-1.5 max-w-2xl text-[12px] leading-relaxed text-ink-900/60">
          Keep your driver roster organized and ready for the next delivery.
        </p>
      </div>
      <button
        v-if="can('staff.invite')"
        type="button"
        class="inline-flex min-h-10 items-center gap-1.5 rounded-xl bg-brand-600 px-4 py-2.5 text-[12px] font-semibold text-white shadow-sm shadow-brand-600/20 transition hover:-translate-y-0.5 hover:bg-brand-700"
        @click="openCreateForm"
      >
        <Plus :size="14" />
        Add driver
      </button>
    </div>

    <p v-if="!workspaceLoading && !can('staff.invite')" class="mt-4 flex items-start gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3.5 py-2.5 text-[11.5px] text-amber-900">
      <AlertCircle :size="14" class="mt-0.5 shrink-0" />
      Your account can view drivers but doesn’t have permission to create or manage them. Ask a company administrator to update your role.
    </p>

    <p v-if="error && !formOpen" role="alert" class="mt-5 rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      {{ error }}
    </p>

    <section class="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4" aria-label="Driver summary">
      <article class="driver-stat driver-stat--violet rounded-xl border p-4">
        <p class="text-[10px] font-semibold tracking-wide text-ink-900/50 uppercase">Total drivers</p>
        <p class="mt-2 text-[22px] font-semibold tracking-tight">{{ counts.total }}</p>
      </article>
      <article class="driver-stat driver-stat--mint rounded-xl border p-4">
        <p class="text-[10px] font-semibold tracking-wide text-ink-900/50 uppercase">Available</p>
        <p class="mt-2 text-[22px] font-semibold tracking-tight">{{ counts.available }}</p>
        <p class="mt-1 text-[10px] text-ink-900/40">Active driver portals</p>
      </article>
      <article class="driver-stat driver-stat--blue rounded-xl border p-4">
        <p class="text-[10px] font-semibold tracking-wide text-ink-900/50 uppercase">Assigned</p>
        <p class="mt-2 text-[22px] font-semibold tracking-tight">{{ counts.assigned }}</p>
        <p class="mt-1 text-[10px] text-ink-900/40">Assignments aren’t available yet</p>
      </article>
      <article class="driver-stat driver-stat--peach rounded-xl border p-4">
        <p class="text-[10px] font-semibold tracking-wide text-ink-900/50 uppercase">Needs attention</p>
        <p class="mt-2 text-[22px] font-semibold tracking-tight">{{ counts.attention }}</p>
        <p class="mt-1 text-[10px] text-ink-900/40">Suspended or deactivated portals</p>
      </article>
    </section>

    <section v-if="newPortalURL" class="driver-portal-notice mt-4 rounded-xl border p-4">
      <h2 class="text-[13px] font-semibold text-emerald-900">Driver portal created</h2>
      <p class="mt-1 text-[11.5px] text-emerald-900/70">
        {{ emailSent ? 'A welcome email with this private link was sent.' : 'Email delivery failed. Share this private link with the driver securely.' }}
        Anyone with the link can open this driver portal; keep it private.
      </p>
      <div class="mt-3 flex flex-wrap items-center gap-2">
        <code class="min-w-0 max-w-full break-all rounded-lg border border-emerald-200 bg-white px-3 py-2 text-[11px] text-ink-900/70">
          {{ newPortalURL }}
        </code>
        <button
          type="button"
          class="inline-flex shrink-0 items-center gap-1.5 rounded-lg border border-emerald-300 bg-white px-3 py-2 text-[11px] font-semibold text-emerald-800"
          @click="copyPortalLink(newPortalURL, 'new')"
        >
          <Check v-if="copiedPortalId === 'new'" :size="13" />
          <Copy v-else :size="13" />
          {{ copiedPortalId === 'new' ? 'Copied' : 'Copy link' }}
        </button>
      </div>
      <button type="button" class="mt-2 text-[10.5px] text-emerald-900/60 underline" @click="newPortalURL = ''">
        Dismiss this link
      </button>
    </section>

    <div
      v-if="formOpen"
      class="fixed inset-0 z-[80] flex items-center justify-center bg-ink-950/55 px-4 py-6 backdrop-blur-md"
      role="dialog"
      aria-modal="true"
      :aria-labelledby="editingDriverId ? 'edit-driver-title' : 'create-driver-title'"
      @click.self="closeCreateForm"
    >
      <form
        class="driver-modal w-full max-w-xl overflow-hidden rounded-2xl border shadow-2xl"
        @submit.prevent="saveDriver"
      >
        <div class="driver-modal-heading flex items-start justify-between gap-4 p-5 sm:p-6">
          <div class="flex items-start gap-3">
            <span class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-violet-100 text-violet-700 dark:bg-violet-400/15 dark:text-violet-200"><UserRound :size="18" /></span>
            <div>
              <p class="text-[9.5px] font-semibold tracking-[0.16em] text-brand-700 uppercase dark:text-indigo-200">{{ editingDriverId ? 'DRIVER DETAILS' : 'NEW TEAM MEMBER' }}</p>
              <h2 :id="editingDriverId ? 'edit-driver-title' : 'create-driver-title'" class="mt-1 text-[15px] font-semibold">{{ editingDriverId ? 'Edit driver' : 'Add a driver' }}</h2>
              <p class="mt-1 text-[11px] text-ink-900/50">{{ editingDriverId ? 'Update their contact and vehicle details. Their portal link stays the same.' : 'Their private portal is ready as soon as you add them.' }}</p>
            </div>
          </div>
          <button
            type="button"
            class="rounded-lg p-1.5 text-ink-900/45 transition hover:bg-ink-900/5 hover:text-ink-900"
            aria-label="Close dialog"
            :disabled="saving"
            @click="closeCreateForm"
          >
            <X :size="18" />
          </button>
        </div>
        <p v-if="error" role="alert" class="mt-4 rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
          {{ error }}
        </p>
        <div class="grid gap-3 p-5 sm:grid-cols-2 sm:p-6">
          <label class="block">
            <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Full name</span>
            <input v-model="form.full_name" required minlength="2" maxlength="120" autocomplete="name" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" />
          </label>
          <label class="block">
            <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Phone number</span>
            <input v-model="form.phone" required minlength="7" maxlength="30" type="tel" autocomplete="tel" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" />
          </label>
          <label class="block">
            <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Vehicle plate number</span>
            <input v-model="form.vehicle_plate" required minlength="2" maxlength="20" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm uppercase focus:border-brand-500" />
          </label>
          <label class="block">
            <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Email address</span>
            <input v-model="form.email" required type="email" maxlength="254" autocomplete="email" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" />
          </label>
        </div>
        <p class="mx-5 rounded-lg bg-violet-50 px-3 py-2.5 text-[10.5px] leading-relaxed text-violet-900/75 dark:mx-6 dark:bg-violet-400/10 dark:text-violet-100">
          The driver won’t need to sign up. We’ll create a private portal link and email it to them.
        </p>
        <div class="flex justify-end gap-2 p-5 sm:px-6 sm:pb-6">
          <button type="button" :disabled="saving" class="rounded-lg border border-ink-900/10 px-4 py-2.5 text-[11px] font-semibold text-ink-900/60 hover:bg-canvas disabled:opacity-50" @click="closeCreateForm">
            Not now
          </button>
          <button
            type="submit"
            :disabled="saving"
            class="inline-flex items-center gap-1.5 rounded-lg bg-brand-600 px-4 py-2.5 text-[11px] font-semibold text-white shadow-sm transition hover:bg-brand-700 disabled:opacity-50"
          >
            <Loader2 v-if="saving" :size="14" class="animate-spin" />
            <UserRound v-else :size="14" />
            {{ editingDriverId ? 'Save changes' : 'Add driver' }}
          </button>
        </div>
      </form>
    </div>

    <section class="driver-list-panel mt-4 overflow-hidden rounded-2xl border">
      <div class="flex flex-col gap-3 border-b border-ink-900/8 p-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-[13px] font-semibold">Drivers</h2>
          <p class="mt-0.5 text-[10.5px] text-ink-900/40">{{ counts.total }} total</p>
        </div>
        <label class="relative block w-full sm:max-w-xs">
          <Search :size="15" class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-ink-900/35" />
          <input
            v-model="searchQuery"
            type="search"
            placeholder="Search name, email, phone or plate"
            aria-label="Search drivers"
            class="w-full rounded-lg border border-ink-900/10 py-2.5 pr-3 pl-9 text-[11.5px] outline-none focus:border-brand-500"
          />
        </label>
      </div>
      <div v-if="loading" class="flex items-center gap-2 p-6 text-[12.5px] text-ink-900/50">
        <Loader2 :size="15" class="animate-spin" /> Loading drivers…
      </div>
      <div v-else-if="drivers.length === 0" class="p-8 text-center">
        <UserRound :size="21" class="mx-auto text-ink-900/25" />
        <p class="mt-3 text-[13px] font-semibold">No drivers added yet</p>
        <p class="mt-1 text-[11.5px] text-ink-900/45">Create your first driver to prepare for delivery assignments.</p>
      </div>
      <div v-else-if="filteredDrivers.length === 0" class="p-8 text-center text-[11.5px] text-ink-900/45">
        No drivers match “{{ searchQuery }}”.
      </div>
      <div v-else class="driver-card-grid grid gap-3 p-4 sm:grid-cols-2 xl:grid-cols-3">
        <article v-for="driver in filteredDrivers" :key="driver.id" class="driver-card rounded-xl border p-4">
          <div class="flex items-start justify-between gap-3">
            <div class="flex min-w-0 items-center gap-2.5">
              <span class="flex size-9 shrink-0 items-center justify-center rounded-full bg-brand-50 text-brand-700">
                <UserRound :size="16" />
              </span>
              <div class="min-w-0">
                <h2 class="truncate text-[12.5px] font-semibold">{{ driver.full_name }}</h2>
                <p class="truncate text-[10.5px] text-ink-900/40">{{ driver.email }}</p>
              </div>
            </div>
            <span
              class="rounded-full border px-2 py-0.5 text-[9.5px] font-medium capitalize"
              :class="driver.status === 'active'
                ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                : driver.status === 'suspended'
                  ? 'border-amber-200 bg-amber-50 text-amber-700'
                  : 'border-slate-200 bg-slate-50 text-slate-500'"
            >
              {{ driver.status === 'inactive' ? 'deactivated' : driver.status }}
            </span>
          </div>
          <dl class="mt-4 space-y-2 border-t border-ink-900/8 pt-3 text-[11px]">
            <div class="flex justify-between gap-3"><dt class="text-ink-900/40">Phone</dt><dd class="font-medium">{{ driver.phone }}</dd></div>
            <div class="flex justify-between gap-3"><dt class="text-ink-900/40">Vehicle</dt><dd class="font-medium uppercase">{{ driver.vehicle_plate }}</dd></div>
          </dl>

          <div class="mt-3 rounded-lg border border-ink-900/8 bg-canvas/70 p-2.5">
            <div class="flex items-center justify-between gap-2">
              <p class="text-[9px] font-bold tracking-[0.12em] text-ink-900/45 uppercase">Driver portal link</p>
              <a
                v-if="driver.portal_url && driver.status === 'active'"
                :href="driver.portal_url"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center gap-0.5 text-[9.5px] font-semibold text-brand-700 hover:text-brand-800"
              >
                Open <ExternalLink :size="10" />
              </a>
            </div>
            <code class="mt-1.5 block truncate rounded-md bg-white px-2 py-1.5 text-[9.5px] text-ink-900/55 dark:bg-white/5" :title="driver.portal_url">
              {{ driver.portal_url || 'No portal link available' }}
            </code>
            <div class="mt-2 flex flex-wrap items-center gap-1.5">
              <button
                type="button"
                class="inline-flex items-center gap-1 rounded-md border border-ink-900/10 bg-white px-2 py-1 text-[9.5px] font-semibold text-ink-900/65 hover:border-brand-200 hover:text-brand-700 disabled:opacity-50 dark:bg-white/5"
                :disabled="!driver.portal_url"
                @click="copyPortalLink(driver.portal_url, driver.id)"
              >
                <Check v-if="copiedPortalId === driver.id" :size="11" />
                <Copy v-else :size="11" />
                {{ copiedPortalId === driver.id ? 'Copied' : 'Copy' }}
              </button>
              <button
                type="button"
                class="inline-flex items-center gap-1 rounded-md border border-ink-900/10 bg-white px-2 py-1 text-[9.5px] font-semibold text-ink-900/65 hover:border-brand-200 hover:text-brand-700 disabled:opacity-50 dark:bg-white/5"
                :disabled="!driver.portal_url"
                @click="sharePortalLink(driver)"
              >
                <Share2 :size="11" /> Share
              </button>
            </div>
            <p v-if="driver.status !== 'active'" class="mt-1.5 text-[9px] text-amber-700">
              {{ driver.status === 'suspended' ? 'Suspended' : 'Deactivated' }} links stay private and inactive until you reactivate this driver.
            </p>
          </div>

          <div class="mt-3 flex flex-wrap items-center gap-1.5 border-t border-ink-900/8 pt-3">
            <button
              v-if="canEdit"
              type="button"
              class="inline-flex items-center gap-1 rounded-md border border-ink-900/10 px-2 py-1.5 text-[10px] font-semibold text-ink-900/65 hover:bg-canvas"
              @click="openEditForm(driver)"
            >
              <Pencil :size="11" /> Edit info
            </button>
            <button
              v-if="canManageStatus && driver.status === 'active'"
              type="button"
              class="rounded-md border border-amber-200 px-2 py-1.5 text-[10px] font-semibold text-amber-700 hover:bg-amber-50"
              @click="setDriverStatus(driver, 'suspended')"
            >
              Suspend
            </button>
            <button
              v-if="canManageStatus && driver.status !== 'inactive'"
              type="button"
              class="rounded-md border border-rose-200 px-2 py-1.5 text-[10px] font-semibold text-rose-700 hover:bg-rose-50"
              @click="setDriverStatus(driver, 'inactive')"
            >
              Deactivate
            </button>
            <button
              v-if="canManageStatus && driver.status !== 'active'"
              type="button"
              class="rounded-md border border-emerald-200 px-2 py-1.5 text-[10px] font-semibold text-emerald-700 hover:bg-emerald-50"
              @click="setDriverStatus(driver, 'active')"
            >
              Reactivate
            </button>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>
