<script setup lang="ts">
import { CircleAlert, KeyRound, Pencil, Plus, ShieldCheck, Trash2, X } from 'lucide-vue-next'
import type { PermissionInfo, Role, RoleSavePayload } from '~/types/auth'

const props = withDefaults(
  defineProps<{
    roles: Role[]
    permissions: PermissionInfo[]
    canManage?: boolean
  }>(),
  { canManage: false },
)

const emit = defineEmits<{
  save: [payload: RoleSavePayload]
  remove: [role: Role]
}>()

const groupLabels: Record<string, string> = {
  company: 'Company',
  branches: 'Branches',
  staff: 'Staff',
  shipments: 'Shipments',
  rates: 'Rates',
  reports: 'Reports',
  branding: 'Branding',
  subdomain: 'Subdomain',
}

const groups = computed(() => {
  const grouped = new Map<string, PermissionInfo[]>()
  for (const permission of props.permissions) {
    const key = permission.key.split('.')[0] || 'other'
    grouped.set(key, [...(grouped.get(key) ?? []), permission])
  }
  return [...grouped.entries()].map(([key, items]) => ({
    key,
    label: groupLabels[key] ?? key,
    items,
  }))
})

const creating = ref(false)
const editing = ref<Role | null>(null)
const form = reactive({
  name: '',
  description: '',
  permissions: [] as string[],
})
const error = ref('')

function reset() {
  creating.value = false
  editing.value = null
  form.name = ''
  form.description = ''
  form.permissions = []
  error.value = ''
}

function startCreate() {
  reset()
  creating.value = true
}

function startEdit(role: Role) {
  reset()
  editing.value = role
  form.name = role.name
  form.description = role.description ?? ''
  form.permissions = [...role.permissions]
}

function togglePermission(key: string) {
  form.permissions = form.permissions.includes(key)
    ? form.permissions.filter((permission) => permission !== key)
    : [...form.permissions, key]
}

function submit() {
  error.value = ''
  if (!form.name.trim()) {
    error.value = 'Give this role a name.'
    return
  }
  if (form.permissions.length === 0) {
    error.value = 'Select at least one permission.'
    return
  }
  emit('save', {
    id: editing.value?.id,
    name: form.name.trim(),
    description: form.description.trim(),
    permissions: form.permissions,
  })
  reset()
}

function permissionLabel(key: string) {
  return key.replaceAll('.', ' · ').replaceAll('_', ' ')
}
</script>

<template>
  <div class="space-y-5">
    <div class="flex flex-wrap items-start justify-between gap-3 rounded-xl border border-ink-900/10 bg-white p-5">
      <div>
        <h2 class="text-[15px] font-semibold">Roles and permissions</h2>
        <p class="mt-1 max-w-2xl text-[11.5px] leading-relaxed text-ink-900/45">
          System roles are fixed. Custom roles let you grant a precise set of permissions for a
          company function. You can only grant permissions you hold yourself.
        </p>
      </div>
      <button
        v-if="canManage"
        type="button"
        class="inline-flex items-center gap-1.5 rounded-lg bg-brand-500 px-3.5 py-2 text-[12px] font-semibold text-white transition hover:bg-brand-600"
        @click="startCreate"
      >
        <Plus :size="14" /> New role
      </button>
    </div>

    <form
      v-if="creating || editing"
      class="rounded-xl border border-brand-200 bg-brand-50/30 p-5"
      @submit.prevent="submit"
    >
      <div class="flex items-start justify-between gap-3">
        <div>
          <h3 class="text-[14px] font-semibold">{{ editing ? 'Edit role' : 'New custom role' }}</h3>
          <p class="mt-1 text-[11px] text-ink-900/45">
            {{ editing?.key ? `Key: ${editing.key}` : 'A unique role key is generated automatically.' }}
          </p>
        </div>
        <button type="button" class="text-ink-900/35 hover:text-ink-900" aria-label="Close role editor" @click="reset">
          <X :size="16" />
        </button>
      </div>

      <div class="mt-4 grid gap-3 sm:grid-cols-2">
        <label class="block">
          <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Role name</span>
          <input v-model="form.name" required class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-sm focus:border-brand-500" placeholder="Dispatcher" />
        </label>
        <label class="block">
          <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Description</span>
          <input v-model="form.description" class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-sm focus:border-brand-500" placeholder="Coordinates branch dispatch" />
        </label>
      </div>

      <div class="mt-4 grid gap-4 sm:grid-cols-2">
        <fieldset v-for="group in groups" :key="group.key">
          <legend class="text-[10px] font-semibold tracking-[0.12em] text-ink-900/40 uppercase">
            {{ group.label }}
          </legend>
          <div class="mt-2 space-y-1">
            <label
              v-for="permission in group.items"
              :key="permission.key"
              class="flex cursor-pointer items-start gap-2 rounded-md px-1 py-1 text-[11px] hover:bg-white/60"
            >
              <input
                type="checkbox"
                class="mt-0.5 size-3.5 rounded border-slate-300"
                :checked="form.permissions.includes(permission.key)"
                @change="togglePermission(permission.key)"
              />
              <span>
                <span class="block font-medium capitalize">{{ permissionLabel(permission.key) }}</span>
                <span v-if="permission.description" class="block text-[10px] text-ink-900/40">{{ permission.description }}</span>
              </span>
            </label>
          </div>
        </fieldset>
      </div>

      <p v-if="error" role="alert" class="mt-3 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[11.5px] text-red-700">
        <CircleAlert :size="13" class="mt-0.5 shrink-0" /> {{ error }}
      </p>

      <div class="mt-4 flex justify-end gap-2">
        <button type="button" class="rounded-lg border border-slate-300 bg-white px-3.5 py-2 text-[12px] font-medium transition hover:bg-slate-50" @click="reset">
          Cancel
        </button>
        <button type="submit" class="rounded-lg bg-brand-500 px-4 py-2 text-[12px] font-semibold text-white transition hover:bg-brand-600">
          {{ editing ? 'Save role' : 'Create role' }}
        </button>
      </div>
    </form>

    <div class="grid gap-3 lg:grid-cols-2">
      <article v-for="role in roles" :key="role.id" class="rounded-xl border border-ink-900/10 bg-white p-4">
        <div class="flex items-start justify-between gap-2">
          <div class="flex items-start gap-2.5">
            <span class="flex size-8 shrink-0 items-center justify-center rounded-lg" :class="role.key === 'owner' ? 'bg-amber-50 text-amber-600' : 'bg-ink-900/5 text-ink-900/45'">
              <ShieldCheck :size="15" />
            </span>
            <div>
              <div class="flex flex-wrap items-center gap-1.5">
                <h3 class="text-[13px] font-semibold">{{ role.name }}</h3>
                <span v-if="role.is_system" class="rounded border border-ink-900/10 bg-canvas px-1.5 py-0.5 text-[8.5px] font-medium tracking-wide text-ink-900/40 uppercase">System</span>
                <span v-else class="rounded border border-brand-200 bg-brand-50 px-1.5 py-0.5 text-[8.5px] font-medium tracking-wide text-brand-700 uppercase">Custom</span>
              </div>
              <p class="mt-1 font-mono text-[9.5px] text-ink-900/35">{{ role.key }}</p>
              <p v-if="role.description" class="mt-1 text-[10.5px] leading-relaxed text-ink-900/45">{{ role.description }}</p>
            </div>
          </div>
          <div v-if="canManage && !role.is_system" class="flex shrink-0 gap-1">
            <button type="button" class="rounded-md border border-slate-300 p-1.5 transition hover:bg-slate-50" aria-label="Edit role" @click="startEdit(role)">
              <Pencil :size="12" />
            </button>
            <button type="button" class="rounded-md border border-red-200 p-1.5 text-red-600 transition hover:bg-red-50" aria-label="Delete role" @click="$emit('remove', role)">
              <Trash2 :size="12" />
            </button>
          </div>
        </div>

        <div class="mt-3 border-t border-ink-900/8 pt-2.5">
          <p class="text-[9.5px] font-medium tracking-[0.1em] text-ink-900/35 uppercase">
            {{ role.permissions.length }} permissions
          </p>
          <div class="mt-1.5 flex flex-wrap gap-1">
            <span v-for="permission in role.permissions.slice(0, 6)" :key="permission" class="rounded bg-ink-900/5 px-1.5 py-0.5 text-[9px] text-ink-900/55">
              {{ permission }}
            </span>
            <span v-if="role.permissions.length > 6" class="px-1 py-0.5 text-[9px] text-ink-900/35">
              +{{ role.permissions.length - 6 }}
            </span>
            <span v-if="role.permissions.length === 0" class="text-[10px] text-ink-900/35">No permissions</span>
          </div>
        </div>
      </article>
    </div>

    <div class="flex items-start gap-2 rounded-xl border border-ink-900/10 bg-canvas p-4 text-[11px] leading-relaxed text-ink-900/45">
      <KeyRound :size="14" class="mt-0.5 shrink-0" />
      Custom roles can be selected when inviting staff or changing a member's role. Deleting a role
      requires moving its members first.
    </div>
  </div>
</template>
