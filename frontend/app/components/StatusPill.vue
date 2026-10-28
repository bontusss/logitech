<script setup lang="ts">
/** Blue = done/good, red = needs attention, outline = in progress/neutral. */
const props = withDefaults(
  defineProps<{
    status: string
    kind?: 'shipment' | 'payment' | 'exception' | 'record'
  }>(),
  { kind: 'shipment' },
)

type Tone = 'danger' | 'info' | 'neutral'

const MAPS: Record<string, Record<string, { label: string; tone: Tone }>> = {
  shipment: {
    pending: { label: 'Pending', tone: 'neutral' },
    in_transit: { label: 'In transit', tone: 'info' },
    received: { label: 'Received', tone: 'info' },
    exception: { label: 'Exception', tone: 'danger' },
    cancelled: { label: 'Cancelled', tone: 'neutral' },
  },
  payment: {
    pending: { label: 'Unpaid', tone: 'neutral' },
    partial: { label: 'Part paid', tone: 'danger' },
    paid: { label: 'Paid', tone: 'info' },
  },
  exception: {
    open: { label: 'Open', tone: 'danger' },
    investigating: { label: 'Investigating', tone: 'info' },
    resolved: { label: 'Resolved', tone: 'info' },
    closed: { label: 'Closed', tone: 'neutral' },
  },
}

const entry = computed(() => {
  const key = props.kind === 'record' ? 'shipment' : props.kind
  return MAPS[key]?.[props.status] ?? { label: humaniseStatus(props.status), tone: 'neutral' as Tone }
})

function humaniseStatus(value: string) {
  const spaced = value.replaceAll('_', ' ')
  return spaced.charAt(0).toUpperCase() + spaced.slice(1)
}
</script>

<template>
  <UiPill :tone="entry.tone">{{ entry.label }}</UiPill>
</template>
