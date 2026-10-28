<script setup lang="ts">
import type { Shipment } from '~/types/api'

/** One record: tracking number, route, parties, money and state. */
withDefaults(defineProps<{ record: Shipment; showMoney?: boolean }>(), { showMoney: true })
</script>

<template>
  <NuxtLink
    :to="`/records/${record.id}`"
    class="group block rounded-lg bg-surface p-4 shadow-card transition hover:shadow-halo focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
  >
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <p class="text-meta tracking-wide text-ink-40">{{ record.tracking_number }}</p>
        <p class="mt-0.5 truncate text-p1 font-medium text-ink">{{ record.description }}</p>
      </div>
      <div class="flex shrink-0 items-center gap-1.5">
        <StatusPill :status="record.status" />
      </div>
    </div>

    <p class="mt-1.5 flex flex-wrap items-center gap-x-1.5 text-meta text-ink-64">
      <span>{{ record.origin_branch }}</span>
      <UiIcon name="arrowRight" :size="12" class="text-ink-40" />
      <span>{{ record.destination_branch }}</span>
      <span class="text-ink-40">·</span>
      <span>Qty {{ record.quantity }}</span>
    </p>
    <p class="mt-0.5 truncate text-meta text-ink-40">
      {{ record.sender.full_name }} → {{ record.receiver.full_name }}
    </p>

    <div v-if="showMoney" class="mt-3 flex items-end justify-between gap-3 border-t border-line pt-2.5">
      <div>
        <p class="text-meta text-ink-40">Charged</p>
        <p class="text-label font-medium text-ink">{{ money(record.amount, record.currency) }}</p>
      </div>
      <div>
        <p class="text-meta text-ink-40">Paid</p>
        <p class="text-label font-medium text-ink-64">{{ money(record.amount_paid, record.currency) }}</p>
      </div>
      <div class="text-right">
        <p class="text-meta text-ink-40">Balance</p>
        <p class="text-label font-medium" :class="record.balance > 0 ? 'text-danger' : 'text-ink-64'">
          {{ money(record.balance, record.currency) }}
        </p>
      </div>
    </div>
  </NuxtLink>
</template>
