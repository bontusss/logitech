<script setup lang="ts">
const config = useRuntimeConfig()
const apiBase = config.public.apiBase

const raw = ref<string | null>(null)
const ok = ref<boolean | null>(null)
const checking = ref(false)

async function checkHealth() {
  checking.value = true
  raw.value = null
  try {
    const res = await fetch(`${apiBase}/health`)
    raw.value = await res.text()
    ok.value = res.ok
  } catch (e) {
    raw.value = String(e)
    ok.value = false
  } finally {
    checking.value = false
  }
}
</script>

<template>
  <main class="page">
    <h1>Go + Nuxt</h1>
    <p>Go server (API + embedded frontend) on :8080 — Nuxt dev server on :3000.</p>
    <button :disabled="checking" @click="checkHealth">
      {{ checking ? 'Checking…' : 'Check /health' }}
    </button>
    <p v-if="ok === true" class="ok">healthy ✔</p>
    <p v-else-if="ok === false" class="err">unhealthy ✘</p>
    <pre v-if="raw">{{ raw }}</pre>
  </main>
</template>

<style scoped>
.page {
  font-family: system-ui, sans-serif;
  max-width: 560px;
  margin: 4rem auto;
}
.ok { color: #0a7a0a; }
.err { color: #b00020; }
pre { background: #f4f4f5; padding: 0.75rem; border-radius: 8px; }
button { cursor: pointer; padding: 0.5rem 1rem; }
</style>
