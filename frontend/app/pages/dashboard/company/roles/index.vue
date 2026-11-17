<script setup lang="ts">
import { CircleAlert, Loader2 } from 'lucide-vue-next'
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
  <div>
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-[10.5px] font-semibold tracking-[0.2em] text-brand-600 uppercase">Company management</p>
        <h1 class="mt-2 text-[clamp(1.6rem,3vw,2.1rem)] leading-tight font-semibold tracking-[-0.035em]">
          Roles & permissions
        </h1>
        <p class="mt-2 max-w-2xl text-[13.5px] leading-relaxed text-ink-900/50">
          Control exactly what each staff role can see and change. Administrators cannot grant
          permissions beyond their own authority, and ownership can never be delegated here.
        </p>
      </div>
      <LookupRefreshButton :loading="loading" @refresh="fetchRoles" />
    </div>

    <p v-if="error" role="alert" class="mt-5 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-[12px] text-red-700">
      <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ error }}
    </p>

    <div v-if="loading && !roles.length" class="mt-5 flex items-center gap-2 rounded-xl border border-ink-900/10 bg-white p-6 text-[12.5px] text-ink-900/45">
      <Loader2 :size="15" class="animate-spin" /> Loading roles…
    </div>

    <RolePermissionEditor
      v-else
      class="mt-5"
      :roles="roles"
      :permissions="permissions"
      :can-manage="can('staff.roles.manage')"
      @save="saveRole"
      @remove="removeRole"
    />
  </div>
</template>
