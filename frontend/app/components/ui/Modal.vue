<script setup lang="ts">
/** Bottom sheet on phones, centred dialog on wider screens. */
const props = withDefaults(
  defineProps<{
    open: boolean
    title?: string
    description?: string
    size?: 'sm' | 'md' | 'lg'
    dismissible?: boolean
  }>(),
  { title: '', description: '', size: 'sm', dismissible: true },
)

const emit = defineEmits<{ close: [] }>()

const width = computed(() => ({ sm: 'sm:max-w-[380px]', md: 'sm:max-w-[560px]', lg: 'sm:max-w-[760px]' })[props.size])

function close() {
  if (props.dismissible) emit('close')
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && props.open) close()
}

let previousOverflow = ''
watch(
  () => props.open,
  (open) => {
    if (import.meta.client) {
      if (open) {
        previousOverflow = document.body.style.overflow
        document.body.style.overflow = 'hidden'
      } else {
        document.body.style.overflow = previousOverflow
      }
    }
  },
)

onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown)
  if (import.meta.client) document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-end justify-center bg-ink/35 sm:items-center sm:px-4 sm:py-8"
      role="dialog"
      aria-modal="true"
      :aria-label="title || 'Dialog'"
      @click.self="close"
    >
      <div
        class="flex max-h-[92dvh] w-full flex-col overflow-hidden rounded-t-2xl bg-surface shadow-halo sm:rounded-lg"
        :class="width"
      >
        <header class="flex items-start gap-3 border-b border-line px-4 py-3 sm:px-5">
          <div class="min-w-0 flex-1">
            <h2 v-if="title" class="text-h5 font-medium text-ink">{{ title }}</h2>
            <p v-if="description" class="mt-0.5 text-label text-ink-64">{{ description }}</p>
          </div>
          <button
            v-if="dismissible"
            type="button"
            class="-mr-1 flex size-9 shrink-0 items-center justify-center rounded-lg text-ink-64 transition hover:bg-canvas hover:text-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
            aria-label="Close"
            @click="close"
          >
            <UiIcon name="close" :size="18" />
          </button>
        </header>

        <div class="min-h-0 flex-1 overflow-y-auto px-4 py-4 sm:px-5">
          <slot />
        </div>

        <footer v-if="$slots.footer" class="border-t border-line bg-surface px-4 py-3 sm:px-5">
          <slot name="footer" />
        </footer>
      </div>
    </div>
  </Teleport>
</template>
