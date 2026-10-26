<script setup lang="ts">
/**
 * Raised action button.
 * primary   — the one main action on a screen (blue, raised)
 * secondary — dark CTA for supporting actions (grey gradient, raised)
 * tertiary  — light surface button with a soft shadow
 * link      — plain text action
 */
type Variant = 'primary' | 'secondary' | 'tertiary' | 'link' | 'danger'

defineOptions({ inheritAttrs: false })

const props = withDefaults(
  defineProps<{
    variant?: Variant
    type?: 'button' | 'submit' | 'reset'
    loading?: boolean
    disabled?: boolean
    block?: boolean
  }>(),
  { variant: 'primary', type: 'button', loading: false, disabled: false, block: false },
)

const VARIANT_CLASSES: Record<Variant, string> = {
  primary:
    'bg-blue text-white shadow-raised hover:brightness-105 active:brightness-95 disabled:bg-muted disabled:text-ink-40 disabled:shadow-none',
  secondary:
    'bg-linear-to-b from-grad-start to-grad-end text-white shadow-raised hover:brightness-110 active:brightness-95 disabled:bg-muted disabled:bg-none disabled:text-ink-40 disabled:shadow-none',
  tertiary:
    'bg-surface text-ink shadow-light hover:bg-canvas disabled:bg-muted disabled:text-ink-40 disabled:shadow-none',
  link: 'h-auto px-0 text-ink underline decoration-ink-40 underline-offset-4 hover:decoration-ink disabled:bg-transparent disabled:text-ink-40',
  danger:
    'bg-danger text-white shadow-raised hover:brightness-105 active:brightness-95 disabled:bg-muted disabled:text-ink-40 disabled:shadow-none',
}
</script>

<template>
  <button
    v-bind="$attrs"
    :type="props.type"
    :disabled="props.disabled || props.loading"
    :class="[
      'inline-flex h-10 shrink-0 items-center justify-center gap-2 rounded-lg px-3 text-button whitespace-nowrap transition duration-150 select-none',
      'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue disabled:cursor-not-allowed',
      VARIANT_CLASSES[props.variant],
      props.block && 'w-full',
    ]"
  >
    <span
      v-if="props.loading"
      class="size-3 animate-spin rounded-full border-2 border-current border-t-transparent"
      aria-hidden="true"
    />
    <slot />
  </button>
</template>
