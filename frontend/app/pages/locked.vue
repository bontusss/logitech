<script setup lang="ts">
definePageMeta({ layout: 'auth' })

useHead({ title: 'Account locked' })

const route = useRoute()

const staffId = computed(() => (typeof route.query.staff_id === 'string' ? route.query.staff_id : ''))
const lockedAt = computed(() => (typeof route.query.at === 'string' ? route.query.at : ''))
</script>

<template>
  <div class="w-full max-w-[360px]">
    <UiPanel title="Account locked" @keydown.esc="navigateTo('/sign-in')">
      <p class="text-p1 text-ink">
        Ask a manager or the owner to reset your password. They give you a one-time password to
        sign in with.
      </p>

      <div class="mt-4">
        <UiButton variant="link" @click="navigateTo('/sign-in')">
          Back to sign in
          <UiKbd>Esc</UiKbd>
        </UiButton>
      </div>
    </UiPanel>

    <p v-if="staffId && lockedAt" class="mt-6 text-center text-label text-danger">
      3 wrong passwords for {{ staffId }}. Locked at {{ lockedAt }}.
    </p>
  </div>
</template>
