<script setup lang="ts">
import { Building2, Loader2, MapPin, Plus, Power } from 'lucide-vue-next'
import type { Branch, StaffMember } from '~/types/auth'

const props = withDefaults(
  defineProps<{
    branches: Branch[]
    members?: StaffMember[]
    loading?: boolean
    canCreate?: boolean
    canUpdate?: boolean
    canDeactivate?: boolean
  }>(),
  { members: () => [], loading: false, canCreate: false, canUpdate: false, canDeactivate: false },
)

defineEmits<{
  create: []
  toggle: [branch: Branch]
}>()

const managerNames = computed(() => {
  const map = new Map<string, string>()
  for (const member of props.members) map.set(member.id, member.full_name)
  return map
})

function managerName(branch: Branch) {
  if (!branch.manager_member_id) return 'No manager assigned'
  return managerNames.value.get(branch.manager_member_id) ?? 'Assigned staff member'
}
</script>

<template>
  <div class="dash-panel overflow-hidden rounded-2xl border">
    <div class="flex flex-wrap items-center justify-between gap-3 border-b border-ink-900/8 bg-gradient-to-r from-sky-50/60 via-white to-violet-50/50 px-4 py-3.5 sm:px-5 dark:from-sky-400/5 dark:via-transparent dark:to-violet-400/5">
      <div>
        <h2 class="text-[14px] font-semibold">Branches</h2>
        <p class="mt-0.5 text-[10.5px] text-ink-900/40">
          {{ branches.length }} {{ branches.length === 1 ? 'branch' : 'branches' }}
        </p>
      </div>
      <button
        v-if="canCreate"
        type="button"
        class="inline-flex items-center gap-1.5 rounded-lg bg-brand-600 px-3 py-2 text-[12px] font-semibold text-white shadow-sm shadow-brand-600/20 transition hover:bg-brand-700"
        @click="$emit('create')"
      >
        <Plus :size="14" /> Add branch
      </button>
    </div>

    <div v-if="loading" class="flex items-center gap-2 px-5 py-10 text-[12.5px] text-ink-900/45">
      <Loader2 :size="15" class="animate-spin" /> Loading branches…
    </div>

    <div v-else-if="branches.length === 0" class="px-5 py-12 text-center">
      <span class="mx-auto flex size-11 items-center justify-center rounded-2xl bg-gradient-to-br from-sky-100 to-violet-50 text-sky-700 dark:from-sky-400/15 dark:to-violet-400/10 dark:text-sky-200">
        <Building2 :size="19" />
      </span>
      <p class="mt-3 text-[13px] font-semibold">No branches yet</p>
      <p class="mx-auto mt-1 max-w-sm text-[11.5px] leading-relaxed text-ink-900/45">
        Add the first location so staff can be assigned and shipments can be routed through it.
      </p>
      <button
        v-if="canCreate"
        type="button"
        class="mt-4 inline-flex items-center gap-1.5 rounded-lg border border-slate-300 px-3 py-2 text-[12px] font-medium transition hover:bg-slate-50"
        @click="$emit('create')"
      >
        <Plus :size="13" /> Add your first branch
      </button>
    </div>

    <ul v-else class="divide-y divide-ink-900/8">
      <li v-for="branch in branches" :key="branch.id" class="flex flex-wrap items-start gap-3 px-4 py-4 transition hover:bg-sky-50/30 sm:px-5">
        <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-sky-100 to-cyan-50 text-sky-700 dark:from-sky-400/15 dark:to-cyan-400/10 dark:text-sky-200">
          <MapPin :size="16" />
        </span>
        <div class="min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-2">
            <NuxtLink
              :to="`/dashboard/company/branches/${branch.id}`"
              class="text-[13px] font-semibold hover:text-brand-600"
            >
              {{ branch.name }}
            </NuxtLink>
            <span class="rounded border border-ink-900/10 bg-canvas px-1.5 py-0.5 font-mono text-[9px] tracking-wide text-ink-900/45 uppercase dark:border-white/10 dark:bg-white/5">
              {{ branch.code }}
            </span>
            <span
              class="rounded-full border px-2 py-0.5 text-[9.5px] font-medium capitalize"
              :class="branch.status === 'active'
                ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                : 'border-slate-200 bg-slate-50 text-slate-500'"
            >
              {{ branch.status }}
            </span>
          </div>
          <p class="mt-1 text-[11px] text-ink-900/45">
            {{ [branch.address, branch.city, branch.state].filter(Boolean).join(', ') }}
          </p>
          <p class="mt-0.5 text-[10.5px] text-ink-900/35">
            Manager: {{ managerName(branch) }}
          </p>
        </div>
        <div class="flex shrink-0 items-center gap-1.5">
          <NuxtLink
            :to="`/dashboard/company/branches/${branch.id}`"
            class="rounded-lg border border-slate-300 px-2.5 py-1.5 text-[11px] font-medium transition hover:bg-slate-50"
          >
            {{ canUpdate ? 'Manage' : 'View' }}
          </NuxtLink>
          <button
            v-if="canDeactivate"
            type="button"
            class="inline-flex items-center gap-1 rounded-lg border border-slate-300 px-2.5 py-1.5 text-[11px] font-medium transition hover:bg-slate-50"
            @click="$emit('toggle', branch)"
          >
            <Power :size="12" />
            {{ branch.status === 'active' ? 'Deactivate' : 'Activate' }}
          </button>
        </div>
      </li>
    </ul>
  </div>
</template>
