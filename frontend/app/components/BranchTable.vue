<script setup lang="ts">
import type { BranchReportRow } from '~/types/api'

defineProps<{ rows: BranchReportRow[] }>()
</script>

<template>
  <div class="-mx-4 overflow-x-auto px-4 sm:mx-0 sm:px-0">
    <table class="w-full min-w-[620px] border-separate border-spacing-0 text-left">
      <thead>
        <tr class="text-meta text-ink-40">
          <th class="border-b border-line pb-2 font-medium">Branch</th>
          <th class="border-b border-line pb-2 text-right font-medium">Sent</th>
          <th class="border-b border-line pb-2 text-right font-medium">Received</th>
          <th class="border-b border-line pb-2 text-right font-medium">Billed</th>
          <th class="border-b border-line pb-2 text-right font-medium">Collected</th>
          <th class="border-b border-line pb-2 text-right font-medium">Outstanding</th>
          <th class="border-b border-line pb-2 text-right font-medium">Open</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in rows" :key="row.branch_id" class="text-label text-ink">
          <td class="border-b border-line py-2.5 pr-3">
            <span class="block font-medium">{{ row.branch_name }}</span>
            <span class="text-meta text-ink-40">{{ row.branch_code }}</span>
          </td>
          <td class="border-b border-line py-2.5 text-right tabular-nums">{{ row.shipments_sent }}</td>
          <td class="border-b border-line py-2.5 text-right tabular-nums">{{ row.shipments_received }}</td>
          <td class="border-b border-line py-2.5 text-right tabular-nums">{{ money(row.billed) }}</td>
          <td class="border-b border-line py-2.5 text-right tabular-nums">{{ money(row.collected) }}</td>
          <td
            class="border-b border-line py-2.5 text-right font-medium tabular-nums"
            :class="row.outstanding > 0 ? 'text-danger' : 'text-ink-64'"
          >
            {{ money(row.outstanding) }}
          </td>
          <td class="border-b border-line py-2.5 text-right tabular-nums" :class="row.open_exceptions > 0 ? 'text-danger' : 'text-ink-64'">
            {{ row.open_exceptions }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
