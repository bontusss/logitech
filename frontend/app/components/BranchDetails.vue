<script setup lang="ts">
import { Clock3, Mail, MapPin, Pencil, Phone, Power, UserRound } from 'lucide-vue-next'
import type { Branch, StaffMember } from '~/types/auth'

const props = withDefaults(
  defineProps<{
    branch: Branch
    staff: Array<{
      id: string
      full_name: string
      email: string
      title: string | null
      status: string
      is_manager: boolean
      roles: string[]
    }>
    allMembers?: StaffMember[]
    canUpdate?: boolean
    canDeactivate?: boolean
    canAssign?: boolean
    saving?: boolean
  }>(),
  { allMembers: () => [], canUpdate: false, canDeactivate: false, canAssign: false, saving: false },
)

const emit = defineEmits<{
  edit: []
  toggle: []
  assign: [memberIDs: string[]]
}>()

const selected = ref<string[]>([])

watch(
  () => props.staff,
  (staff) => {
    selected.value = staff.map((member) => member.id)
  },
  { immediate: true, deep: true },
)

const managerName = computed(() => {
  const manager = props.staff.find((member) => member.id === props.branch.manager_member_id)
  if (manager) return manager.full_name
  const member = props.allMembers.find((item) => item.id === props.branch.manager_member_id)
  return member?.full_name ?? 'No manager assigned'
})

const dayLabels: Record<string, string> = {
  mon: 'Mon', tue: 'Tue', wed: 'Wed', thu: 'Thu', fri: 'Fri', sat: 'Sat', sun: 'Sun',
}

const hoursSummary = computed(() => {
  const entries = Object.entries(props.branch.operating_hours ?? {})
  if (!entries.length) return 'Not set'
  const open = entries.filter(([, day]) => !day.closed)
  if (!open.length) return 'Always closed'
  return `${open.length} open ${open.length === 1 ? 'day' : 'days'}`
})

function toggleMember(memberID: string) {
  selected.value = selected.value.includes(memberID)
    ? selected.value.filter((id) => id !== memberID)
    : [...selected.value, memberID]
}
</script>

<template>
  <div class="space-y-4">
    <div class="dash-panel rounded-2xl border p-5">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div class="flex items-start gap-3">
          <span class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-brand-100 to-brand-50 text-brand-700 dark:from-brand-400/15 dark:to-brand-400/10 dark:text-brand-200">
            <MapPin :size="18" />
          </span>
          <div>
          <div class="flex flex-wrap items-center gap-2">
            <h1 class="text-[19px] font-semibold tracking-[-0.03em]">{{ branch.name }}</h1>
            <span class="rounded border border-ink-900/10 bg-canvas px-1.5 py-0.5 font-mono text-[9.5px] uppercase text-ink-900/45">
              {{ branch.code }}
            </span>
            <span
              class="rounded-full border px-2 py-0.5 text-[9.5px] font-medium capitalize"
              :class="branch.status === 'active'
                ? 'border-success-200 bg-success-50 text-success-700'
                : 'border-slate-200 bg-slate-50 text-slate-500'"
            >
              {{ branch.status }}
            </span>
          </div>
          <p class="mt-1.5 text-[11.5px] text-ink-900/45">
            {{ branch.address }}, {{ branch.city }}, {{ branch.state }}, {{ branch.country }}
          </p>
          </div>
        </div>
        <div class="flex gap-2">
          <button
            v-if="canUpdate"
            type="button"
            class="inline-flex items-center gap-1.5 rounded-lg border border-slate-300 px-3 py-2 text-[12px] font-medium transition hover:bg-slate-50"
            @click="$emit('edit')"
          >
            <Pencil :size="13" /> Edit
          </button>
          <button
            v-if="canDeactivate"
            type="button"
            :disabled="saving"
            class="inline-flex items-center gap-1.5 rounded-lg border border-slate-300 px-3 py-2 text-[12px] font-medium transition hover:bg-slate-50 disabled:opacity-50"
            @click="$emit('toggle')"
          >
            <Power :size="13" />
            {{ branch.status === 'active' ? 'Deactivate' : 'Activate' }}
          </button>
        </div>
      </div>

      <dl class="mt-5 grid gap-3 border-t border-ink-900/8 pt-4 text-[11.5px] sm:grid-cols-2 lg:grid-cols-3">
        <div class="flex items-start gap-2 rounded-xl bg-brand-50/50 p-2.5 dark:bg-brand-400/5">
          <MapPin :size="14" class="mt-0.5 shrink-0 text-brand-600 dark:text-brand-300" />
          <div><dt class="text-ink-900/40">Address</dt><dd class="mt-0.5 font-medium">{{ branch.address }}</dd></div>
        </div>
        <div class="flex items-start gap-2 rounded-xl bg-brand-50/50 p-2.5 dark:bg-brand-400/5">
          <UserRound :size="14" class="mt-0.5 shrink-0 text-brand-600 dark:text-brand-300" />
          <div><dt class="text-ink-900/40">Manager</dt><dd class="mt-0.5 font-medium">{{ managerName }}</dd></div>
        </div>
        <div class="flex items-start gap-2 rounded-xl bg-accent-50/50 p-2.5 dark:bg-accent-400/5">
          <Clock3 :size="14" class="mt-0.5 shrink-0 text-accent-600 dark:text-accent-300" />
          <div><dt class="text-ink-900/40">Operating hours</dt><dd class="mt-0.5 font-medium">{{ hoursSummary }}</dd></div>
        </div>
        <div class="flex items-start gap-2 rounded-xl bg-success-50/50 p-2.5 dark:bg-success-500/5">
          <Phone :size="14" class="mt-0.5 shrink-0 text-success-600 dark:text-success-200" />
          <div><dt class="text-ink-900/40">Phone</dt><dd class="mt-0.5 font-medium">{{ branch.phone || 'Not set' }}</dd></div>
        </div>
        <div class="flex items-start gap-2 rounded-xl bg-red-50/50 p-2.5 dark:bg-red-400/5">
          <Mail :size="14" class="mt-0.5 shrink-0 text-red-600 dark:text-red-300" />
          <div><dt class="text-ink-900/40">Email</dt><dd class="mt-0.5 font-medium">{{ branch.email || 'Not set' }}</dd></div>
        </div>
      </dl>

      <div v-if="Object.keys(branch.operating_hours ?? {}).length" class="mt-4 flex flex-wrap gap-1.5 border-t border-ink-900/8 pt-3">
        <span
          v-for="(day, key) in branch.operating_hours"
          :key="key"
          class="rounded-md border px-2 py-1 text-[10px]"
          :class="day.closed ? 'border-ink-900/8 bg-canvas text-ink-900/35' : 'border-ink-900/10 bg-white text-ink-900/60'"
        >
          {{ dayLabels[key] ?? key }}: {{ day.closed ? 'Closed' : `${day.open}–${day.close}` }}
        </span>
      </div>
    </div>

    <div class="dash-panel rounded-2xl border p-5">
      <div class="flex items-center justify-between gap-3">
        <div>
          <h2 class="text-[14px] font-semibold">Branch staff</h2>
          <p class="mt-0.5 text-[10.5px] text-ink-900/40">
            Only assigned staff can view and operate this branch.
          </p>
        </div>
        <span class="text-[10.5px] text-ink-900/40">{{ staff.length }} assigned</span>
      </div>

      <div v-if="canAssign" class="mt-4">
        <div v-if="allMembers.length === 0" class="text-[11.5px] text-ink-900/40">No active staff members to assign.</div>
        <div v-else class="grid gap-1.5 sm:grid-cols-2">
          <label
            v-for="member in allMembers.filter((item) => item.status === 'active')"
            :key="member.id"
            class="flex cursor-pointer items-center gap-2 rounded-lg border px-3 py-2 text-[11.5px] transition"
            :class="selected.includes(member.id) ? 'border-brand-200 bg-brand-50' : 'border-ink-900/8 hover:bg-canvas'"
          >
            <input
              type="checkbox"
              class="size-3.5 rounded border-slate-300"
              :checked="selected.includes(member.id)"
              @change="toggleMember(member.id)"
            />
            <span class="min-w-0 flex-1">
              <span class="block truncate font-medium">{{ member.full_name }}</span>
              <span class="block truncate text-[10px] text-ink-900/40">{{ member.email }}</span>
            </span>
            <span v-if="member.id === branch.manager_member_id" class="text-[9px] font-semibold text-brand-600 uppercase">Manager</span>
          </label>
        </div>
        <div class="mt-3 flex justify-end">
          <button
            type="button"
            :disabled="saving"
            class="rounded-lg bg-accent-500 px-3.5 py-2 text-[12px] font-semibold text-white shadow-sm shadow-accent-500/20 transition hover:bg-accent-400 disabled:opacity-60"
            @click="$emit('assign', selected)"
          >
            Save staff assignments
          </button>
        </div>
      </div>

      <ul v-else-if="staff.length" class="mt-4 divide-y divide-ink-900/8">
        <li v-for="member in staff" :key="member.id" class="flex items-center gap-2.5 py-2.5">
          <span class="flex size-7 items-center justify-center rounded-full bg-gradient-to-br from-brand-100 to-brand-100 text-[10px] font-semibold text-brand-700 dark:from-brand-400/20 dark:to-brand-400/20 dark:text-brand-100">
            {{ member.full_name.slice(0, 2).toUpperCase() }}
          </span>
          <div class="min-w-0 flex-1">
            <p class="truncate text-[12px] font-medium">{{ member.full_name }}</p>
            <p class="truncate text-[10px] text-ink-900/40">{{ member.email }}</p>
          </div>
          <span v-if="member.is_manager" class="rounded border border-brand-200 bg-brand-50 px-1.5 py-0.5 text-[9px] font-semibold text-brand-700">Manager</span>
        </li>
      </ul>
      <p v-else class="mt-4 text-[11.5px] text-ink-900/40">No staff assigned to this branch yet.</p>
    </div>
  </div>
</template>
