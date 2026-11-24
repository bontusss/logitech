<script setup lang="ts">
import { CircleAlert, Loader2, ShieldCheck, Sparkles } from 'lucide-vue-next'
import type { PermissionInfo, Role, RoleSavePayload } from '~/types/auth'

definePageMeta({ middleware: 'auth', layout: 'app' })
useSeoMeta({ title: 'Roles & permissions' })

const { api } = useApi()
const { load, can, notify, confirm } = useWorkspace()

const roles = ref<Role[]>([])
const permissions = ref<PermissionInfo[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  await load()
  await fetchRoles()
})

async function fetchRoles() {
  loading.value = true
  error.value = ''
  try {
    const data = await api<{ roles: Role[]; permissions: PermissionInfo[] }>('/companies/me/roles')
    roles.value = data.roles
    permissions.value = data.permissions
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not load roles and permissions.')
  } finally {
    loading.value = false
  }
}

async function saveRole(payload: RoleSavePayload) {
  error.value = ''
  try {
    if (payload.id) {
      await api(`/companies/me/roles/${payload.id}`, { method: 'PATCH', body: payload })
      notify('Role updated.')
    } else {
      await api('/companies/me/roles', { method: 'POST', body: payload })
      notify('Role created.')
    }
    await fetchRoles()
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not save the role.')
  }
}

async function removeRole(role: Role) {
  const approved = await confirm({
    title: 'Delete this role?',
    message: `The ${role.name} role will be permanently removed. Roles still assigned to members cannot be deleted.`,
    confirmLabel: 'Delete role',
    danger: true,
  })
  if (!approved) return
  try {
    await api(`/companies/me/roles/${role.id}`, { method: 'DELETE' })
    notify('Role deleted.')
    await fetchRoles()
  } catch (err) {
    error.value = apiErrorMessage(err, 'Could not delete the role.')
  }
}
</script>

<template>
  <div class="dash-page">
    <header class="dash-hero relative overflow-hidden rounded-2xl border p-5 sm:p-6">
      <div class="pointer-events-none absolute -top-16 right-10 size-44 rounded-full bg-accent-400/15 blur-3xl" />
      <div class="pointer-events-none absolute -right-8 -bottom-24 size-52 rounded-full bg-brand-400/15 blur-3xl" />
      <div class="relative flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="dash-eyebrow inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-semibold tracking-wide">
            <Sparkles :size="12" /> ACCESS CONTROL
          </p>
          <h1 class="mt-2 text-[clamp(1.45rem,2.5vw,1.8rem)] leading-tight font-semibold tracking-[-0.04em]">
            Roles &amp; permissions
          </h1>
          <p class="mt-1.5 max-w-xl text-[12px] leading-relaxed text-ink-900/65">
            Control exactly what each staff role can see and change. Administrators cannot grant
            permissions beyond their own authority, and ownership can never be delegated here.
          </p>
        </div>
        <div class="flex items-center gap-2">
          <span class="hidden items-center gap-1.5 rounded-full border border-white/70 bg-white/70 px-3 py-1.5 text-[11px] font-semibold text-ink-900/60 sm:inline-flex dark:border-white/10 dark:bg-white/5">
            <ShieldCheck :size="13" /> {{ roles.length }} roles
          </span>
          <LookupRefreshButton :loading="loading" @refresh="fetchRoles" />
        </div>
      </div>
    </header>

    <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <div v-if="loading && !roles.length" class="dash-panel mt-4 flex items-center gap-2 rounded-2xl border p-6 text-[12.5px] text-ink-900/45">
      <Loader2 :size="15" class="animate-spin" /> Loading roles…
    </div>

    <RolePermissionEditor
      v-else
      class="mt-4"
      :roles="roles"
      :permissions="permissions"
      :can-manage="can('staff.roles.manage')"
      @save="saveRole"
      @remove="removeRole"
    />
  </div>
</template>
