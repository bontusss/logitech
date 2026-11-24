<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    intent: 'login' | 'customer' | 'provider'
    next?: string
    label?: string
  }>(),
  { label: 'Continue with Google' },
)

const { baseURL } = useApi()
const redirecting = ref(false)

function start() {
  const apiRoot = new URL(baseURL, window.location.origin).toString().replace(/\/$/, '')
  const params = new URLSearchParams({ intent: props.intent })
  if (props.next) params.set('next', props.next)

  redirecting.value = true
  window.location.assign(`${apiRoot}/auth/google/start?${params.toString()}`)
}
</script>

<template>
  <button
    type="button"
    :disabled="redirecting"
    class="inline-flex min-h-11 w-full items-center justify-center gap-2.5 rounded-xl border border-ink-900/15 bg-surface px-4 py-2.5 text-[12.5px] font-semibold text-ink-900 shadow-sm transition hover:border-ink-900/25 hover:bg-canvas disabled:cursor-wait disabled:opacity-60"
    @click="start"
  >
    <svg viewBox="0 0 24 24" class="size-[17px] shrink-0" aria-hidden="true">
      <path
        fill="#4285F4"
        d="M23.49 12.27c0-.79-.07-1.54-.2-2.27H12v4.51h6.47a5.54 5.54 0 0 1-2.4 3.63v3h3.88c2.27-2.09 3.54-5.17 3.54-8.87Z"
      />
      <path
        fill="#34A853"
        d="M12 24c3.24 0 5.96-1.07 7.95-2.91l-3.88-3c-1.07.72-2.45 1.15-4.07 1.15-3.13 0-5.78-2.11-6.73-4.95H1.26v3.09A12 12 0 0 0 12 24Z"
      />
      <path
        fill="#FBBC05"
        d="M5.27 14.29A7.2 7.2 0 0 1 4.89 12c0-.8.14-1.57.38-2.29V6.62H1.26A12 12 0 0 0 0 12c0 1.94.46 3.77 1.26 5.38l4.01-3.09Z"
      />
      <path
        fill="#EA4335"
        d="M12 4.77c1.76 0 3.34.61 4.58 1.8l3.44-3.44A11.6 11.6 0 0 0 12 0 12 12 0 0 0 1.26 6.62l4.01 3.09C6.22 6.88 8.87 4.77 12 4.77Z"
      />
    </svg>
    {{ redirecting ? 'Opening Google…' : label }}
  </button>
</template>
