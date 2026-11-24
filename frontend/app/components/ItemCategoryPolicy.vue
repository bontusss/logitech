<script setup lang="ts">
import { Check, Loader2, PackageCheck, Pencil, ShieldAlert, X } from 'lucide-vue-next'
import { ITEM_CATEGORIES, itemCategoryLabel } from '~/utils/itemCategories'

const { api } = useApi()
const { can, notify } = useWorkspace()

const accepted = ref<string[]>([])
const prohibited = ref<string[]>([])
const draftAccepted = ref<string[]>([])
const draftProhibited = ref<string[]>([])
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const editorOpen = ref(false)
const canEdit = computed(() => can('rates.update'))

const allAccepted = computed(() =>
  ITEM_CATEGORIES.every((category) => accepted.value.includes(category.value))
  && prohibited.value.length === 0,
)
const acceptedOptions = computed(() =>
  ITEM_CATEGORIES.filter((category) => accepted.value.includes(category.value)),
)
const prohibitedOptions = computed(() =>
  ITEM_CATEGORIES.filter((category) =>
    prohibited.value.includes(category.value) || !accepted.value.includes(category.value),
  ),
)

onMounted(fetchPolicy)

async function fetchPolicy() {
  loading.value = true
  error.value = ''
  try {
    const response = await api<{ accepted: string[]; prohibited: string[] }>('/companies/me/item-categories')
    accepted.value = response.accepted?.length ? response.accepted : ITEM_CATEGORIES.map((category) => category.value)
    prohibited.value = response.prohibited ?? []
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not load your item categories.')
  } finally {
    loading.value = false
  }
}

function openEditor() {
  draftAccepted.value = [...accepted.value]
  draftProhibited.value = prohibitedOptions.value.map((category) => category.value)
  error.value = ''
  editorOpen.value = true
}

function closeEditor() {
  if (!saving.value) editorOpen.value = false
}

function accept(category: string) {
  draftProhibited.value = draftProhibited.value.filter((value) => value !== category)
  if (!draftAccepted.value.includes(category)) draftAccepted.value.push(category)
}

function prohibit(category: string) {
  draftAccepted.value = draftAccepted.value.filter((value) => value !== category)
  if (!draftProhibited.value.includes(category)) draftProhibited.value.push(category)
}

function acceptAll() {
  draftAccepted.value = ITEM_CATEGORIES.map((category) => category.value)
  draftProhibited.value = []
}

async function savePolicy() {
  if (!draftAccepted.value.length) {
    error.value = 'Keep at least one acceptable category, or mark the others as prohibited.'
    return
  }
  saving.value = true
  error.value = ''
  try {
    const response = await api<{ accepted: string[]; prohibited: string[] }>('/companies/me/item-categories', {
      method: 'PUT',
      body: { accepted: draftAccepted.value, prohibited: draftProhibited.value },
    })
    accepted.value = response.accepted ?? []
    prohibited.value = response.prohibited ?? []
    editorOpen.value = false
    notify('Item categories updated.')
  } catch (cause) {
    error.value = apiErrorMessage(cause, 'Could not save your item categories.')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <section class="dash-panel mt-4 overflow-hidden rounded-2xl border">
    <div class="flex flex-col gap-3 border-b border-ink-900/8 p-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex items-start gap-3">
        <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-200">
          <PackageCheck :size="17" />
        </span>
        <div>
          <h2 class="text-[13px] font-semibold">Items you can move</h2>
          <p class="mt-0.5 max-w-2xl text-[10.5px] leading-relaxed text-ink-900/45">
            Customers only see these categories on your booking page. Prohibited items are blocked before a booking is sent.
          </p>
        </div>
      </div>
      <button
        v-if="canEdit"
        type="button"
        class="inline-flex shrink-0 items-center gap-1.5 rounded-lg border border-ink-900/10 px-3 py-2 text-[10.5px] font-semibold text-ink-900/65 hover:bg-canvas"
        @click="openEditor"
      >
        <Pencil :size="12" /> Edit categories
      </button>
    </div>

    <div v-if="loading" class="flex items-center gap-2 p-5 text-[12px] text-ink-900/45">
      <Loader2 :size="14" class="animate-spin" /> Loading item categories…
    </div>
    <p v-else-if="error" role="alert" class="m-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[11.5px] text-red-700">
      {{ error }}
    </p>
    <div v-else class="grid gap-4 p-4 sm:grid-cols-2">
      <div class="rounded-xl border border-success-200/70 bg-success-50/45 p-3 dark:border-success-500/15 dark:bg-success-500/5">
        <p class="text-[9.5px] font-bold tracking-[0.12em] text-success-700 uppercase dark:text-success-200">Acceptable</p>
        <p v-if="allAccepted" class="mt-1.5 text-[11px] font-medium text-success-700/70 dark:text-success-100/70">
          All standard categories accepted.
        </p>
        <div v-else class="mt-2 flex flex-wrap gap-1.5">
          <span
            v-for="category in acceptedOptions"
            :key="category.value"
            class="rounded-full border border-success-200 bg-white px-2 py-1 text-[10px] font-semibold text-success-700 dark:border-success-500/20 dark:bg-white/5 dark:text-success-200"
          >
            {{ category.emoji }} {{ itemCategoryLabel(category.value) }}
          </span>
        </div>
      </div>
      <div class="rounded-xl border border-red-200/70 bg-red-50/45 p-3 dark:border-red-400/15 dark:bg-red-400/5">
        <p class="text-[9.5px] font-bold tracking-[0.12em] text-red-800 uppercase dark:text-red-200">Prohibited</p>
        <div v-if="prohibitedOptions.length" class="mt-2 flex flex-wrap gap-1.5">
          <span
            v-for="category in prohibitedOptions"
            :key="category.value"
            class="rounded-full border border-red-200 bg-white px-2 py-1 text-[10px] font-semibold text-red-800 dark:border-red-400/20 dark:bg-white/5 dark:text-red-200"
          >
            {{ itemCategoryLabel(category.value) }}
          </span>
        </div>
        <p v-else class="mt-1.5 text-[11px] text-red-900/55 dark:text-red-100/55">No categories prohibited.</p>
      </div>
    </div>
  </section>

  <div
    v-if="editorOpen"
    class="fixed inset-0 z-[80] flex items-center justify-center bg-ink-950/55 px-3 py-4 backdrop-blur-md"
    role="dialog"
    aria-modal="true"
    aria-labelledby="item-category-policy-title"
    @click.self="closeEditor"
  >
    <form class="max-h-full w-full max-w-xl overflow-y-auto rounded-2xl border bg-white shadow-2xl dark:bg-ink-950" @submit.prevent="savePolicy">
      <header class="sticky top-0 z-10 flex items-start justify-between gap-3 border-b border-ink-900/8 bg-white px-5 py-4 dark:bg-ink-950">
        <div>
          <p class="text-[9.5px] font-semibold tracking-[0.16em] text-brand-700 uppercase dark:text-brand-200">ITEM POLICY</p>
          <h2 id="item-category-policy-title" class="mt-1 text-[15px] font-semibold">What can you move?</h2>
          <p class="mt-1 text-[11px] text-ink-900/50">Mark each category as acceptable or prohibited. Customers will only see acceptable categories.</p>
        </div>
        <button type="button" class="rounded-lg p-1.5 text-ink-900/45 hover:bg-ink-900/5" aria-label="Close" :disabled="saving" @click="closeEditor">
          <X :size="18" />
        </button>
      </header>

      <p v-if="error" role="alert" class="mx-5 mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[11.5px] text-red-700">
        {{ error }}
      </p>

      <div class="flex items-center justify-between gap-3 px-5 pt-4">
        <p class="text-[10.5px] text-ink-900/45">Choose one state for every category.</p>
        <button type="button" class="text-[10.5px] font-semibold text-brand-700 hover:text-brand-600" @click="acceptAll">Accept all</button>
      </div>

      <div class="space-y-2 p-5">
        <div
          v-for="category in ITEM_CATEGORIES"
          :key="category.value"
          class="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-ink-900/8 bg-canvas/60 p-3 dark:bg-white/5"
        >
          <div class="flex min-w-0 items-center gap-2.5">
            <span class="text-base">{{ category.emoji }}</span>
            <div>
              <p class="text-[12px] font-semibold">{{ category.label }}</p>
              <p class="text-[9.5px] text-ink-900/40">{{ category.hint }}</p>
            </div>
          </div>
          <div class="flex shrink-0 items-center gap-1 rounded-lg border border-ink-900/8 bg-white p-0.5 dark:bg-ink-950">
            <button
              type="button"
              class="inline-flex items-center gap-1 rounded-md px-2 py-1.5 text-[10px] font-semibold transition"
              :class="draftAccepted.includes(category.value)
                ? 'bg-success-600 text-white'
                : 'text-ink-900/50 hover:text-success-700'"
              @click="accept(category.value)"
            >
              <Check :size="11" /> Accept
            </button>
            <button
              type="button"
              class="inline-flex items-center gap-1 rounded-md px-2 py-1.5 text-[10px] font-semibold transition"
              :class="draftProhibited.includes(category.value)
                ? 'bg-red-600 text-white'
                : 'text-ink-900/50 hover:text-red-700'"
              @click="prohibit(category.value)"
            >
              <ShieldAlert :size="11" /> Prohibit
            </button>
          </div>
        </div>
      </div>

      <footer class="sticky bottom-0 flex items-center justify-between gap-3 border-t border-ink-900/8 bg-white px-5 py-3.5 dark:bg-ink-950">
        <p class="hidden text-[10px] text-ink-900/40 sm:block">Prohibited always wins over acceptable.</p>
        <div class="ml-auto flex gap-2">
          <button type="button" class="rounded-lg border border-ink-900/10 px-3 py-2 text-[11px] font-semibold" :disabled="saving" @click="closeEditor">Cancel</button>
          <button type="submit" :disabled="saving" class="inline-flex items-center gap-1.5 rounded-lg bg-accent-500 px-4 py-2 text-[11px] font-semibold text-ink-950 hover:bg-accent-400 disabled:opacity-60">
            <Loader2 v-if="saving" :size="13" class="animate-spin" />
            <Check v-else :size="13" />
            {{ saving ? 'Saving…' : 'Save categories' }}
          </button>
        </div>
      </footer>
    </form>
  </div>
</template>
