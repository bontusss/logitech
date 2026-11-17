<script setup lang="ts">
import type { OperatingHours } from '~/types/auth'

const props = withDefaults(
  defineProps<{
    modelValue?: OperatingHours
    disabled?: boolean
  }>(),
  { modelValue: () => ({}), disabled: false },
)

const emit = defineEmits<{ 'update:modelValue': [value: OperatingHours] }>()

const days = [
  { key: 'mon', label: 'Monday' },
  { key: 'tue', label: 'Tuesday' },
  { key: 'wed', label: 'Wednesday' },
  { key: 'thu', label: 'Thursday' },
  { key: 'fri', label: 'Friday' },
  { key: 'sat', label: 'Saturday' },
  { key: 'sun', label: 'Sunday' },
]

const hours = reactive<OperatingHours>({})

function sync(value: OperatingHours) {
  for (const key of Object.keys(hours)) delete hours[key]
  for (const [index, day] of days.entries()) {
    hours[day.key] = {
      open: value?.[day.key]?.open || '08:00',
      close: value?.[day.key]?.close || '18:00',
      closed: value?.[day.key]?.closed ?? index > 4,
    }
  }
}

watch(
  () => props.modelValue,
  (value) => sync(value || {}),
  { immediate: true, deep: true },
)

function update() {
  emit('update:modelValue', JSON.parse(JSON.stringify(hours)) as OperatingHours)
}
</script>

<template>
  <div class="divide-y divide-ink-900/8 rounded-lg border border-ink-900/10">
    <div
      v-for="day in days"
      :key="day.key"
      class="flex flex-wrap items-center gap-2 px-3 py-2.5 sm:flex-nowrap"
    >
      <span class="w-24 text-[11.5px] font-medium text-ink-900/65">{{ day.label }}</span>
      <label class="flex items-center gap-1.5 text-[11px] text-ink-900/45">
        <input
          v-model="hours[day.key]!.closed"
          type="checkbox"
          class="size-3.5 rounded border-slate-300"
          :disabled="disabled"
          @change="update"
        />
        Closed
      </label>
      <div class="ml-auto flex items-center gap-1.5">
        <input
          v-model="hours[day.key]!.open"
          type="time"
          class="rounded-md border border-slate-300 px-2 py-1.5 text-[11.5px] disabled:bg-slate-50 disabled:text-ink-900/35"
          :disabled="disabled || hours[day.key]!.closed"
          @input="update"
        />
        <span class="text-[10px] text-ink-900/30">to</span>
        <input
          v-model="hours[day.key]!.close"
          type="time"
          class="rounded-md border border-slate-300 px-2 py-1.5 text-[11.5px] disabled:bg-slate-50 disabled:text-ink-900/35"
          :disabled="disabled || hours[day.key]!.closed"
          @input="update"
        />
      </div>
    </div>
  </div>
</template>
