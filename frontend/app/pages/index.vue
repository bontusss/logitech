<script setup lang="ts">
useHead({ title: 'Home' })

interface Profile {
  label: string
  value: string
}

const { user } = useAuth()

const profile = computed<Profile[]>(() => {
  if (!user.value) return []
  return [
    { label: 'Staff ID', value: user.value.staff_id },
    { label: 'Role', value: user.value.job_role ? `${user.value.role} · ${user.value.job_role}` : user.value.role },
    { label: 'Branch', value: user.value.branch_id || 'Not assigned' },
    { label: 'Phone', value: user.value.phone || 'Not provided' },
    { label: 'Email', value: user.value.email || 'Not provided' },
    { label: 'PIN', value: user.value.first_login_done ? 'Set' : 'Not set' },
  ]
})
</script>

<template>
  <div v-if="user" class="space-y-6">
    <div>
      <p class="text-label text-ink-64">Signed in as</p>
      <h1 class="mt-1 text-h5 text-ink">{{ user.full_name }}</h1>
      <div class="mt-3 flex flex-wrap items-center gap-2">
        <UiPill tone="info">{{ user.role }}</UiPill>
        <UiPill>{{ user.status }}</UiPill>
        <UiPill v-if="user.job_role">{{ user.job_role }}</UiPill>
      </div>
    </div>

    <UiCard>
      <dl class="divide-y divide-line">
        <div v-for="row in profile" :key="row.label" class="flex items-start justify-between gap-6 py-3 first:pt-0 last:pb-0">
          <dt class="text-label text-ink-64">{{ row.label }}</dt>
          <dd class="text-right text-p1 text-ink">{{ row.value }}</dd>
        </div>
      </dl>
    </UiCard>

    <p class="text-label text-ink-64">
      Use the menu to go to the staff, drivers, vehicles or branches screens.
    </p>
  </div>
</template>
