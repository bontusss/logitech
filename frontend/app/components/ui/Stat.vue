<script setup lang="ts">
/** A single metric: label, value, optional hint and trend tone. Can be clickable with `to`. */
withDefaults(
  defineProps<{
    label: string
    value: string
    hint?: string
    icon?: string
    tone?: 'default' | 'info' | 'danger'
    to?: string
  }>(),
  { hint: '', icon: '', tone: 'default', to: '' },
)
</script>

<template>
  <NuxtLink
    v-if="to"
    :to="to"
    class="block rounded-lg bg-surface p-4 shadow-card transition hover:shadow-halo focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
  >
    <div class="flex items-start justify-between gap-3">
      <p class="text-label text-ink-64">{{ label }}</p>
      <span
        v-if="icon"
        class="flex size-8 shrink-0 items-center justify-center rounded-lg"
        :class="tone === 'danger' ? 'bg-danger/10 text-danger' : tone === 'info' ? 'bg-chip/45 text-blue' : 'bg-canvas text-ink-64'"
      >
        <UiIcon :name="icon" :size="16" />
      </span>
    </div>
    <p
      class="mt-3 text-h5 font-medium tracking-tight"
      :class="tone === 'danger' ? 'text-danger' : 'text-ink'"
    >
      {{ value }}
    </p>
    <p v-if="hint" class="mt-1 text-meta text-ink-40">{{ hint }}</p>
  </NuxtLink>
  <div v-else class="rounded-lg bg-surface p-4 shadow-card">
    <div class="flex items-start justify-between gap-3">
      <p class="text-label text-ink-64">{{ label }}</p>
      <span
        v-if="icon"
        class="flex size-8 shrink-0 items-center justify-center rounded-lg"
        :class="tone === 'danger' ? 'bg-danger/10 text-danger' : tone === 'info' ? 'bg-chip/45 text-blue' : 'bg-canvas text-ink-64'"
      >
        <UiIcon :name="icon" :size="16" />
      </span>
    </div>
    <p
      class="mt-3 text-h5 font-medium tracking-tight"
      :class="tone === 'danger' ? 'text-danger' : 'text-ink'"
    >
      {{ value }}
    </p>
    <p v-if="hint" class="mt-1 text-meta text-ink-40">{{ hint }}</p>
  </div>
</template>
