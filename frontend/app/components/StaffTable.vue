<script setup lang="ts">
import { KeyRound, Pencil, Trash2, UserRound } from 'lucide-vue-next'
import type { Branch, StaffMember } from '~/types/auth'

const props = withDefaults(
  defineProps<{
    members: StaffMember[]
    branches?: Branch[]
    canUpdate?: boolean
    canDeactivate?: boolean
    canRoles?: boolean
  }>(),
  { branches: () => [], canUpdate: false, canDeactivate: false, canRoles: false },
)

defineEmits<{
  edit: [member: StaffMember]
  role: [member: StaffMember]
  remove: [member: StaffMember]
}>()

const branchNames = computed(() => {
  const map = new Map<string, string>()
  for (const branch of props.branches) map.set(branch.id, branch.name)
  return map
})

function branchLabel(member: StaffMember) {
  if (!member.branch_ids.length) return 'All branches'
  const names = member.branch_ids.map((id) => branchNames.value.get(id)).filter(Boolean)
  return names.length ? names.join(', ') : `${member.branch_ids.length} assigned`
}

function isOwner(member: StaffMember) {
  return member.roles.includes('owner')
}
</script>

<template>
  <div class="dash-panel overflow-hidden rounded-2xl border">
    <div v-if="members.length === 0" class="px-5 py-12 text-center">
      <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-100 to-brand-50 text-brand-700 dark:from-brand-400/15 dark:to-brand-400/10 dark:text-brand-200">
        <UserRound :size="19" />
      </span>
      <p class="mt-3 text-[13px] font-semibold">No staff match this view</p>
      <p class="mt-1 text-[11.5px] text-ink-900/45">Invite staff or clear the filters to see your team.</p>
    </div>

    <div v-else class="overflow-x-auto">
      <table class="w-full min-w-[760px] text-left">
        <thead>
          <tr class="border-b border-ink-900/8 bg-gradient-to-r from-brand-50/60 via-white to-brand-50/40 text-[9px] tracking-[0.08em] text-ink-900/40 uppercase dark:from-brand-400/5 dark:via-transparent dark:to-brand-400/5">
            <th class="px-4 py-2.5 font-semibold">Staff member</th>
            <th class="px-3 py-2.5 font-semibold">Role</th>
            <th class="px-3 py-2.5 font-semibold">Branches</th>
            <th class="px-3 py-2.5 font-semibold">Status</th>
            <th class="px-4 py-2.5 text-right font-semibold">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="member in members" :key="member.id" class="border-b border-ink-900/6 transition last:border-0 hover:bg-brand-50/25">
            <td class="px-4 py-3">
              <div class="flex items-center gap-2.5">
                <span class="flex size-7 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-brand-100 to-brand-100 text-[10px] font-semibold text-brand-700 dark:from-brand-400/20 dark:to-brand-400/20 dark:text-brand-100">
                  {{ member.full_name.slice(0, 2).toUpperCase() }}
                </span>
                <div class="min-w-0">
                  <p class="truncate text-[12.5px] font-medium">{{ member.full_name }}</p>
                  <p class="truncate text-[10.5px] text-ink-900/40">{{ member.email }}</p>
                </div>
              </div>
            </td>
            <td class="px-3 py-3">
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="role in member.roles"
                  :key="role"
                  class="rounded border px-1.5 py-0.5 text-[9.5px] font-medium capitalize"
                  :class="role === 'owner' ? 'border-accent-200 bg-accent-50 text-accent-700' : 'border-brand-200/70 bg-brand-50/70 text-brand-700 dark:border-brand-400/20 dark:bg-brand-400/10 dark:text-brand-200'"
                >
                  {{ role.replaceAll('_', ' ') }}
                </span>
                <span v-if="!member.roles.length" class="text-[10px] text-ink-900/35">No role</span>
              </div>
              <p v-if="member.title" class="mt-1 text-[10px] text-ink-900/35">{{ member.title }}</p>
            </td>
            <td class="max-w-[190px] px-3 py-3 text-[11px] text-ink-900/55">{{ branchLabel(member) }}</td>
            <td class="px-3 py-3">
              <span
                class="rounded-full border px-2 py-0.5 text-[9.5px] font-medium capitalize"
                :class="member.status === 'active'
                  ? 'border-success-200 bg-success-50 text-success-700'
                  : 'border-red-200 bg-red-50 text-red-700'"
              >
                {{ member.status }}
              </span>
            </td>
            <td class="px-4 py-3">
              <div class="flex justify-end gap-1">
                <button
                  v-if="canUpdate && !isOwner(member)"
                  type="button"
                  class="rounded-md border border-slate-300 px-2 py-1 text-[10.5px] font-medium transition hover:bg-slate-50"
                  @click="$emit('edit', member)"
                >
                  <Pencil :size="11" class="inline" /> Edit
                </button>
                <button
                  v-if="canRoles && !isOwner(member)"
                  type="button"
                  class="rounded-md border border-slate-300 px-2 py-1 text-[10.5px] font-medium transition hover:bg-slate-50"
                  @click="$emit('role', member)"
                >
                  <KeyRound :size="11" class="inline" /> Role
                </button>
                <button
                  v-if="canDeactivate && !isOwner(member)"
                  type="button"
                  class="rounded-md border border-red-200 px-2 py-1 text-[10.5px] font-medium text-red-600 transition hover:bg-red-50"
                  @click="$emit('remove', member)"
                >
                  <Trash2 :size="11" class="inline" /> Remove
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
