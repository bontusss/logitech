<script setup lang="ts">
import { CircleAlert, Loader2, Send, X } from 'lucide-vue-next'
import type { Branch, Role } from '~/types/auth'

const props = withDefaults(
  defineProps<{
    open: boolean
    roles: Role[]
    branches?: Branch[]
    saving?: boolean
  }>(),
  { branches: () => [], saving: false },
)

const emit = defineEmits<{
  close: []
  invite: [payload: { email: string; role_key: string; branch_ids: string[] }]
}>()

const form = reactive({
  email: '',
  role_key: 'operations_staff',
  branch_ids: [] as string[],
})
const error = ref('')

watch(
  () => props.open,
  (open) => {
    if (open) {
      form.email = ''
      form.role_key = props.roles.find((role) => role.key === 'operations_staff')?.key
        ?? props.roles.find((role) => !role.key.startsWith('custom_'))?.key
        ?? ''
      form.branch_ids = []
      error.value = ''
    }
  },
)

function toggleBranch(id: string) {
  form.branch_ids = form.branch_ids.includes(id)
    ? form.branch_ids.filter((branchID) => branchID !== id)
    : [...form.branch_ids, id]
}

function submit() {
  error.value = ''
  if (!form.email.trim() || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim())) {
    error.value = 'Enter a valid email address.'
    return
  }
  if (!form.role_key) {
    error.value = 'Choose a role for this invitation.'
    return
  }
  emit('invite', {
    email: form.email.trim().toLowerCase(),
    role_key: form.role_key,
    branch_ids: form.branch_ids,
  })
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-[80] flex items-center justify-center bg-ink-950/45 px-4 py-6 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      aria-label="Invite staff"
      @click.self="$emit('close')"
    >
      <form class="max-h-full w-full max-w-lg overflow-y-auto rounded-2xl border border-ink-900/10 bg-white p-5 shadow-2xl" @submit.prevent="submit">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-[16px] font-semibold">Invite a staff member</h2>
            <p class="mt-1 text-[11.5px] leading-relaxed text-ink-900/45">
              They receive a secure, expiring link. The invitation must be accepted with the same email address.
            </p>
          </div>
          <button type="button" class="text-ink-900/35 hover:text-ink-900" aria-label="Close invitation dialog" @click="$emit('close')">
            <X :size="16" />
          </button>
        </div>

        <div class="mt-4 space-y-3.5">
          <label class="block">
            <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Email address</span>
            <input v-model="form.email" type="email" required class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500" placeholder="staff@company.com" />
          </label>

          <label class="block">
            <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Role</span>
            <select v-model="form.role_key" required class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-brand-500">
              <option v-for="role in roles.filter((item) => item.key !== 'owner')" :key="role.id" :value="role.key">
                {{ role.name }}{{ role.is_system ? '' : ' (custom)' }}
              </option>
            </select>
            <span class="mt-1 block text-[10px] text-ink-900/35">
              You can only grant permissions you hold yourself.
            </span>
          </label>

          <div>
            <span class="mb-1.5 block text-[11px] font-medium text-ink-900/55">Branch access</span>
            <div v-if="branches.length" class="grid gap-1.5 sm:grid-cols-2">
              <label
                v-for="branch in branches"
                :key="branch.id"
                class="flex cursor-pointer items-center gap-2 rounded-lg border px-2.5 py-2 text-[11.5px] transition"
                :class="form.branch_ids.includes(branch.id) ? 'border-brand-200 bg-brand-50' : 'border-ink-900/8 hover:bg-canvas'"
              >
                <input
                  type="checkbox"
                  class="size-3.5 rounded border-slate-300"
                  :checked="form.branch_ids.includes(branch.id)"
                  @change="toggleBranch(branch.id)"
                />
                <span class="min-w-0 flex-1 truncate">{{ branch.name }}</span>
                <span class="font-mono text-[9px] text-ink-900/35">{{ branch.code }}</span>
              </label>
            </div>
            <p v-else class="rounded-lg border border-dashed border-ink-900/12 px-3 py-2 text-[11px] text-ink-900/40">
              No branches yet. The member can be assigned later.
            </p>
          </div>
        </div>

        <p v-if="error" role="alert" class="mt-4 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-[12px] text-red-700">
          <CircleAlert :size="14" class="mt-0.5 shrink-0" /> {{ error }}
        </p>

        <div class="mt-5 flex justify-end gap-2">
          <button type="button" class="rounded-lg border border-slate-300 px-3.5 py-2 text-[12.5px] font-medium transition hover:bg-slate-50" @click="$emit('close')">
            Cancel
          </button>
          <button type="submit" :disabled="saving" class="inline-flex items-center gap-1.5 rounded-lg bg-brand-500 px-4 py-2.5 text-[12.5px] font-semibold text-white transition hover:bg-brand-600 disabled:opacity-60">
            <Loader2 v-if="saving" :size="14" class="animate-spin" />
            <Send v-else :size="14" />
            Send invitation
          </button>
        </div>
      </form>
    </div>
  </Teleport>
</template>
