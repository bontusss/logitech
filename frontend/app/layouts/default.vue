<script setup lang="ts">
const { user, logout } = useAuth()
const signingOut = ref(false)

async function signOut() {
  signingOut.value = true
  try {
    await logout()
    await navigateTo('/sign-in')
  } finally {
    signingOut.value = false
  }
}
</script>

<template>
  <div class="min-h-dvh">
    <main class="w-full">
      <slot />
    </main>
  </div>
</template>
