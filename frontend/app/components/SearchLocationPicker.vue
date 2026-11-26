<script setup lang="ts">
import type { BookingLocation } from '~/components/BookingLocationField.vue'

export interface SearchGeoZone {
  id: string
  state: string
  lga: string
  name: string
  parent_id?: string | null
}

export interface SearchLocationValue {
  state: string
  lgaKey: string
  zoneId: string
  area: string
  address: string
  latitude: number | null
  longitude: number | null
}

interface LgaOption {
  key: string
  label: string
  zoneId: string
  areas: SearchGeoZone[]
}

const props = withDefaults(
  defineProps<{
    label: string
    zones: SearchGeoZone[]
    modelValue: SearchLocationValue
    loading?: boolean
    showPin?: boolean
  }>(),
  { loading: false, showPin: true },
)

const emit = defineEmits<{ 'update:modelValue': [value: SearchLocationValue] }>()

const fieldClass =
  'w-full rounded-xl border border-ink-900/12 bg-canvas px-3 py-2.5 text-[12.5px] font-medium text-ink-900 transition hover:border-ink-900/25 focus:border-accent-500 disabled:cursor-not-allowed disabled:opacity-60'
const labelClass = 'mb-1.5 block text-[11px] font-semibold text-ink-900/55'

function lgaKeyFor(zone: Pick<SearchGeoZone, 'state' | 'lga'>) {
  return `${zone.state}|${zone.lga}`
}

const states = computed(() =>
  [...new Set(props.zones.map((zone) => zone.state))].sort((a, b) => a.localeCompare(b)),
)

const lgas = computed<LgaOption[]>(() => {
  const options = new Map<string, LgaOption>()
  for (const zone of props.zones) {
    if (zone.state !== props.modelValue.state) continue
    const key = lgaKeyFor(zone)
    let option = options.get(key)
    if (!option) {
      option = { key, label: zone.lga, zoneId: '', areas: [] }
      options.set(key, option)
    }
    if (zone.parent_id) {
      option.areas.push(zone)
    } else {
      option.zoneId = zone.id
    }
  }
  return [...options.values()]
    .map((option) => ({
      ...option,
      areas: option.areas.sort((a, b) => a.name.localeCompare(b.name)),
    }))
    .sort((a, b) => a.label.localeCompare(b.label))
})

const activeLga = computed<LgaOption | null>(() => {
  const direct = lgas.value.find((option) => option.key === props.modelValue.lgaKey)
  if (direct) return direct
  const zone = props.zones.find((item) => item.id === props.modelValue.zoneId)
  if (!zone) return null
  return lgas.value.find((option) => option.key === lgaKeyFor(zone)) ?? null
})

const areas = computed(() => activeLga.value?.areas ?? [])
const selectedAreaId = computed(() => {
  if (!props.modelValue.area) return ''
  return areas.value.find((area) => area.id === props.modelValue.zoneId)?.id ?? ''
})

const place = computed<BookingLocation | null>(() => {
  if (
    props.modelValue.latitude === null ||
    props.modelValue.longitude === null ||
    !props.modelValue.address
  ) {
    return null
  }
  return {
    address: props.modelValue.address,
    latitude: props.modelValue.latitude,
    longitude: props.modelValue.longitude,
  }
})

function patch(value: Partial<SearchLocationValue>) {
  emit('update:modelValue', { ...props.modelValue, ...value })
}

function selectState(event: Event) {
  patch({
    state: (event.target as HTMLSelectElement).value,
    lgaKey: '',
    zoneId: '',
    area: '',
  })
}

function selectLga(event: Event) {
  const option = lgas.value.find((item) => item.key === (event.target as HTMLSelectElement).value)
  if (!option) return
  patch({ lgaKey: option.key, zoneId: option.zoneId, area: '' })
}

function selectArea(event: Event) {
  const id = (event.target as HTMLSelectElement).value
  if (!id) {
    patch({ zoneId: activeLga.value?.zoneId ?? '', area: '' })
    return
  }
  const area = areas.value.find((item) => item.id === id)
  if (!area) return
  patch({ zoneId: area.id, area: area.name })
}

function setPlace(value: BookingLocation | null) {
  patch({
    address: value?.address ?? '',
    latitude: value?.latitude ?? null,
    longitude: value?.longitude ?? null,
  })
}
</script>

<template>
  <div class="rounded-2xl border border-ink-900/10 bg-canvas/50 p-3" style="--booking-primary: #ff7a00; --booking-primary-soft: rgba(255, 122, 0, 0.12)">
    <p class="mb-2.5 text-[11.5px] font-semibold text-ink-900">{{ label }}</p>

    <div class="grid grid-cols-2 gap-2">
      <label>
        <span :class="labelClass">State</span>
        <select
          :value="modelValue.state"
          :class="fieldClass"
          :disabled="loading || !states.length"
          :aria-label="`${label} state`"
          @change="selectState"
        >
          <option value="" disabled>{{ loading ? 'Loading…' : 'Choose a state' }}</option>
          <option v-for="state in states" :key="state" :value="state">{{ state }}</option>
        </select>
      </label>
      <label>
        <span :class="labelClass">Zone / LGA</span>
        <select
          :value="modelValue.lgaKey"
          :class="fieldClass"
          :disabled="!modelValue.state || !lgas.length"
          :aria-label="`${label} zone`"
          @change="selectLga"
        >
          <option value="" disabled>{{ lgas.length ? 'Choose an LGA' : 'Choose a state first' }}</option>
          <option v-for="lga in lgas" :key="lga.key" :value="lga.key">{{ lga.label }}</option>
        </select>
      </label>
    </div>

    <label class="mt-2 block">
      <span :class="labelClass">Area / neighbourhood</span>
      <select
        :value="selectedAreaId"
        :class="fieldClass"
        :disabled="!modelValue.lgaKey || loading || (!areas.length && !activeLga?.zoneId)"
        :aria-label="`${label} area`"
        @change="selectArea"
      >
        <option value="" :disabled="!activeLga?.zoneId">
          {{ activeLga?.zoneId ? 'Whole LGA' : 'Choose a neighbourhood' }}
        </option>
        <option v-for="area in areas" :key="area.id" :value="area.id">{{ area.name }}</option>
      </select>
      <span v-if="!areas.length && activeLga?.zoneId" class="mt-1 block text-[10px] leading-snug text-ink-900/40">
        No neighbourhoods have been added under this LGA yet. Choose the whole LGA.
      </span>
    </label>

    <div v-if="showPin" class="mt-2.5">
      <BookingLocationField
        :label="`${label} — exact map pin`"
        :model-value="place"
        @update:model-value="setPlace"
      />
    </div>
  </div>
</template>
