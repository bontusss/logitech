<script setup lang="ts">
import { ApiError } from '~/utils/api'

useHead({ title: 'Home' })

definePageMeta({ layout: 'default' })

const { user, token, logout } = useAuth()
const signingOut = ref(false)

// --- camera capture -----------------------------------------------------------
// The hidden file input uses `capture="environment"` so phones open the camera
// directly; desktops fall back to the normal file picker.
const cameraInput = ref<HTMLInputElement | null>(null)
const photoState = ref<'idle' | 'uploading' | 'saved' | 'error'>('idle')
const savedPhotoPath = ref('')

function openCamera() {
  cameraInput.value?.click()
}

async function onPhotoSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  photoState.value = 'uploading'
  try {
    const form = new FormData()
    form.append('image', file, file.name)
    const saved = await apiRequest<{ filename: string; path: string; size: number }>('/api/photos', {
      method: 'POST',
      body: form,
      token: token.value,
    })
    savedPhotoPath.value = saved.path
    photoState.value = 'saved'
  } catch {
    photoState.value = 'error'
  } finally {
    input.value = ''
  }
}

// --- manual manifest entry ---------------------------------------------------
// One line of the manifest form. Kept as strings so the inputs stay controlled.
interface ManifestItemDraft {
  id?: string
  serial_number: string
  description: string
  quantity: string
  sender_name: string
  sender_phone: string
  receiver_name: string
  receiver_phone: string
  destination_branch_id: string
  amount: string
  currency: string
}

interface BranchOption {
  id: string
  name: string
  code: string
}

function todayISO(): string {
  const d = new Date()
  const m = (d.getMonth() + 1).toString().padStart(2, '0')
  const day = d.getDate().toString().padStart(2, '0')
  return `${d.getFullYear()}-${m}-${day}`
}

function emptyItem(): ManifestItemDraft {
  return {
    serial_number: '',
    description: '',
    quantity: '1',
    sender_name: '',
    sender_phone: '',
    receiver_name: '',
    receiver_phone: '',
    destination_branch_id: '',
    amount: '',
    currency: 'NGN',
  }
}

const manualOpen = ref(false)
const manualSaving = ref(false)
const manifestLoading = ref(false)
const manifestReadOnly = ref(false)
const manualError = ref('')
const manualSuccess = ref('')
const editingManifestId = ref<string | null>(null)
const branches = ref<BranchOption[]>([])
const branchesLoading = ref(false)
const manifestNumber = ref('')
const manifestDate = ref(todayISO())
const originBranchId = ref('')
const manifestDestination = ref('')
const manifestStatus = ref('draft')
const manifestItems = ref<ManifestItemDraft[]>([emptyItem()])

const manifestStatusOptions = computed(() => [
  { value: 'draft', label: 'Draft' },
  ...(editingManifestId.value ? [{ value: 'review_required', label: 'Review required' }] : []),
  { value: 'confirmed', label: 'Confirmed' },
  { value: 'awaiting_dispatch', label: 'Awaiting dispatch' },
  ...(editingManifestId.value && !['draft', 'review_required', 'confirmed', 'awaiting_dispatch'].includes(manifestStatus.value)
    ? [{ value: manifestStatus.value, label: manifestStatus.value.replaceAll('_', ' ') }]
    : []),
])

const destinationSelectOptions = computed(() => [
  { value: '', label: 'Set per item' },
  ...branches.value.map((b) => ({ value: b.id, label: `${b.name} (${b.code})` })),
])

const manifestDestinationLabel = computed(() => {
  const branch = branches.value.find((b) => b.id === manifestDestination.value)
  return branch ? `${branch.name} (${branch.code})` : ''
})

const branchSelectOptions = computed(() => [
  { value: '', label: 'Select branch' },
  ...branches.value.map((b) => ({ value: b.id, label: `${b.name} (${b.code})` })),
])

async function loadBranches() {
  if (!token.value || branchesLoading.value) return
  branchesLoading.value = true
  try {
    branches.value = await apiRequest<BranchOption[]>('/api/branches', { token: token.value })
  } catch {
    // The form still opens; the branch selects show an empty option list.
  } finally {
    branchesLoading.value = false
  }
}

async function openManual() {
  manualError.value = ''
  manualSuccess.value = ''
  editingManifestId.value = null
  manifestLoading.value = false
  manifestReadOnly.value = false
  manifestNumber.value = ''
  manifestDate.value = todayISO()
  originBranchId.value = user.value?.branch_id ?? ''
  manifestDestination.value = ''
  manifestStatus.value = 'draft'
  manifestItems.value = [emptyItem()]
  manualOpen.value = true
  document.body.style.overflow = 'hidden'
  if (!branches.value.length) await loadBranches()
  if (!originBranchId.value && user.value?.branch_id) {
    originBranchId.value = user.value.branch_id
  }
}

async function openManifest(manifest: ManifestQueueEntry) {
  manualError.value = ''
  manualSuccess.value = ''
  editingManifestId.value = manifest.id
  manifestReadOnly.value = user.value?.role !== 'manager' && manifest.created_by !== user.value?.id
  manualOpen.value = true
  manifestLoading.value = true
  manifestNumber.value = manifest.manifest_number
  manifestDate.value = todayISO()
  originBranchId.value = ''
  manifestDestination.value = ''
  manifestStatus.value = 'draft'
  manifestItems.value = [emptyItem()]
  document.body.style.overflow = 'hidden'

  try {
    await loadBranches()
    const detail = await apiRequest<{
      id: string
      manifest_number: string
      origin_branch_id: string
      manifest_date: string
      status: string
      items: Array<{
        id: string
        serial_number: string
        description: string
        quantity: number
        sender: { full_name: string; phone: string }
        receiver: { full_name: string; phone: string }
        destination_branch_id: string
        amount?: number
        currency: string
      }>
    }>(`/api/manifests/${manifest.id}`, { token: token.value })

    if (editingManifestId.value !== manifest.id) return
    manifestNumber.value = detail.manifest_number
    manifestDate.value = detail.manifest_date
    originBranchId.value = detail.origin_branch_id
    manifestStatus.value = detail.status
    const destinations = new Set(detail.items.map((item) => item.destination_branch_id))
    manifestDestination.value = destinations.size === 1 ? detail.items[0]?.destination_branch_id ?? '' : ''
    manifestItems.value = detail.items.map((item) => ({
      id: item.id,
      serial_number: item.serial_number,
      description: item.description,
      quantity: String(item.quantity),
      sender_name: item.sender.full_name,
      sender_phone: item.sender.phone,
      receiver_name: item.receiver.full_name,
      receiver_phone: item.receiver.phone,
      destination_branch_id: item.destination_branch_id,
      amount: item.amount === undefined ? '' : String(item.amount),
      currency: item.currency,
    }))
  } catch (err) {
    if (editingManifestId.value === manifest.id) {
      manualError.value = err instanceof ApiError ? err.message : 'Couldn’t load the manifest. Try again.'
    }
  } finally {
    if (editingManifestId.value === manifest.id) manifestLoading.value = false
  }
}

function closeManual() {
  manualOpen.value = false
  editingManifestId.value = null
  manifestLoading.value = false
  manifestReadOnly.value = false
  document.body.style.overflow = ''
}

function addManifestItem() {
  manifestItems.value.push(emptyItem())
}

function removeManifestItem(index: number) {
  if (manifestItems.value.length === 1) return
  manifestItems.value.splice(index, 1)
}

function validateManifestForm(): string {
  const number = manifestNumber.value.trim()
  if (editingManifestId.value && !number) return 'Manifest number is required when editing.'
  if (number && !/^\d{5}$/.test(number)) return 'Manifest number must be 5 digits, or left empty to auto-generate one.'
  if (!originBranchId.value) return 'Choose the branch this manifest leaves from.'
  if (!manifestDate.value) return 'Manifest date is required.'
  if (!manifestItems.value.length) return 'Add at least one item.'

  for (const [i, item] of manifestItems.value.entries()) {
    const label = `Item ${i + 1}:`
    if (!item.description.trim()) return `${label} description is required.`
    if (!item.sender_name.trim()) return `${label} sender name is required.`
    if (!item.receiver_name.trim()) return `${label} receiver name is required.`
    if (!manifestDestination.value && !item.destination_branch_id) return `${label} going-to branch is required unless the manifest has a To branch.`
    if (item.amount.trim() && Number.isNaN(Number(item.amount))) return `${label} amount must be a number.`
  }
  return ''
}

async function saveManifest() {
  if (manualSaving.value) return
  manualError.value = validateManifestForm()
  if (manualError.value) return

  manualSaving.value = true
  try {
    const saved = await apiRequest<{ id: string; manifest_number: string; item_count: number }>(
      editingManifestId.value ? `/api/manifests/${editingManifestId.value}` : '/api/manifests', {
      method: editingManifestId.value ? 'PUT' : 'POST',
      token: token.value,
      body: {
        ...(manifestNumber.value.trim() ? { manifest_number: manifestNumber.value.trim() } : {}),
        origin_branch_id: originBranchId.value,
        ...(manifestDestination.value ? { destination_branch_id: manifestDestination.value } : {}),
        manifest_date: manifestDate.value,
        status: manifestStatus.value,
        items: manifestItems.value.map((item) => ({
          ...(item.id ? { id: item.id } : {}),
          serial_number: item.serial_number.trim() || undefined,
          description: item.description.trim(),
          quantity: Math.max(1, Number(item.quantity) || 1),
          sender: {
            full_name: item.sender_name.trim(),
            ...(item.sender_phone.trim() ? { phone: item.sender_phone.trim() } : {}),
          },
          receiver: {
            full_name: item.receiver_name.trim(),
            ...(item.receiver_phone.trim() ? { phone: item.receiver_phone.trim() } : {}),
          },
          destination_branch_id: manifestDestination.value || item.destination_branch_id,
          ...(item.amount.trim() ? { amount: Number(item.amount) } : {}),
          ...(item.currency.trim() ? { currency: item.currency.trim() } : {}),
        })),
      },
    })

    manualSuccess.value = `Manifest ${saved.manifest_number} saved with ${saved.item_count} item${saved.item_count === 1 ? '' : 's'}.`
    void loadAttention()
    // Reset the form for the next entry.
    manifestNumber.value = ''
    manifestDate.value = todayISO()
    originBranchId.value = user.value?.branch_id ?? ''
    manifestDestination.value = ''
    manifestStatus.value = 'draft'
    manifestItems.value = [emptyItem()]
    closeManual()
  } catch (err) {
    manualError.value = err instanceof ApiError ? err.message : 'Couldn’t save the manifest. Try again.'
  } finally {
    manualSaving.value = false
  }
}

function onManualKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && manualOpen.value) closeManual()
}

onMounted(() => document.addEventListener('keydown', onManualKeydown))
onUnmounted(() => {
  document.removeEventListener('keydown', onManualKeydown)
  document.body.style.overflow = ''
})

// --- manifest queues ---------------------------------------------------------
interface ManifestQueueEntry {
  id: string
  created_by: string
  manifest_number: string
  status: string
  manifest_date: string
  item_count: number
  exception_count: number
}

const drafts = ref<ManifestQueueEntry[]>([])
const attention = ref<ManifestQueueEntry[]>([])
const attentionLoading = ref(false)
const attentionError = ref('')

function attentionLabel(manifest: ManifestQueueEntry): string {
  if (manifest.status === 'review_required' && manifest.exception_count > 0) return 'Review + exception'
  if (manifest.status === 'review_required') return 'Review needed'
  return 'Exception'
}

async function loadAttention() {
  if (!token.value) return
  attentionLoading.value = true
  try {
    const [myDrafts, attentionManifests] = await Promise.all([
      apiRequest<ManifestQueueEntry[]>('/api/manifests?created_by=me&statuses=draft', { token: token.value }),
      apiRequest<ManifestQueueEntry[]>('/api/manifests/attention', { token: token.value }),
    ])
    drafts.value = myDrafts
    attention.value = attentionManifests
    attentionError.value = ''
  } catch {
    attentionError.value = 'Couldn’t load manifests. Try again later.'
  } finally {
    attentionLoading.value = false
  }
}

onMounted(loadAttention)
watch(token, () => { if (token.value) loadAttention() })

async function signOut() {
  signingOut.value = true
  try {
    await logout()
  } finally {
    signingOut.value = false
  }
}

// Mobile action sheets: each card opens a small menu with Take a photo and
// enter manually.
const sendSheetOpen = ref(false)
const receiveSheetOpen = ref(false)

function openSendActionSheet() {
  sendSheetOpen.value = true
  window.scrollTo({ top: 0, behavior: 'auto' })
}

function openReceiveActionSheet() {
  receiveSheetOpen.value = true
  window.scrollTo({ top: 0, behavior: 'auto' })
}

const now = new Date()
const formattedDate = computed(() => {
  const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  return `${days[now.getDay()]} ${now.getDate()} ${months[now.getMonth()]}`
})

const formattedTime = computed(() => {
  const h = now.getHours().toString().padStart(2, '0')
  const m = now.getMinutes().toString().padStart(2, '0')
  return `${h}:${m}`
})

const userName = computed(() => {
  if (!user.value) return ''
  const parts = user.value.full_name.split(' ')
  return parts[0] || user.value.full_name
})
</script>

<template>
  <div v-if="user">
    <div class="min-h-dvh flex flex-col bg-canvas md:hidden">
      <AppNavbar :user-name="user.full_name" :signing-out="signingOut" @logout="signOut" />
    <!-- left sidebar (desktop only) -->
    <aside class="hidden">
      <div class="flex flex-col gap-6 p-5">
        <!-- user info -->
        <div>
          <p class="text-label text-ink-64">Signed in as</p>
          <h2 class="mt-1 text-h5 text-ink">{{ userName }}</h2>
          <button
            type="button"
            class="mt-2 inline-flex items-center gap-1 text-label text-ink-64 underline decoration-ink underline-offset-4 transition hover:text-blue"
            @click="signOut"
          >
            Sign out
          </button>
        </div>

        <!-- prompt -->
        <div>
          <h3 class="text-h4 text-ink">What do you want to do today?</h3>
        </div>

        <!-- send a parcel -->
        <div class="rounded-lg bg-canvas p-4">
          <h4 class="text-label font-medium text-ink">Send a parcel</h4>
          <p class="mt-1 text-label text-ink-64">For when a customer is bringing goods to send.</p>
          <UiButton
            type="button"
            variant="tertiary"
            class="mt-3"
          >
            Send
          </UiButton>
        </div>

        <!-- receive a parcel -->
        <div class="rounded-lg bg-canvas p-4">
          <h4 class="text-label font-medium text-ink">Receive a parcel</h4>
          <p class="mt-1 text-label text-ink-64">For when goods have arrived with a driver</p>
          <div class="mt-4">
            <UiButton
              type="button"
              :disabled="photoState === 'uploading'"
              @click="openCamera"
            >
              {{ photoState === 'uploading' ? 'Saving…' : 'Take a photo' }}
              <UiKbd>R</UiKbd>
            </UiButton>
          </div>
          <p v-if="photoState === 'saved'" class="mt-2 text-label text-blue">Photo saved to {{ savedPhotoPath }}</p>
          <p v-else-if="photoState === 'error'" class="mt-2 text-label text-danger">Couldn’t save the photo. Try again.</p>
          <div class="mt-4">
            <button
              type="button"
              class="inline-flex items-center gap-1 text-label text-ink-64 underline decoration-ink underline-offset-4 transition hover:text-blue"
              @click="openManual"
            >
              Or enter manually
              <UiKbd>M</UiKbd>
            </button>
          </div>
          <p v-if="manualSuccess" class="mt-2 text-label text-blue">{{ manualSuccess }}</p>
        </div>
      </div>

      <!-- sidebar footer -->
      <div class="mt-auto flex items-center justify-between gap-4 border-t border-line p-5">
        <span class="text-label text-ink-64">Galopin, Aba</span>
        <span class="flex items-center gap-1 text-label text-ink-64">
          {{ formattedDate }}
          <span class="text-ink-40">•</span>
          {{ formattedTime }}
          <svg class="ml-1 size-4 text-ink-64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="10"/>
            <path d="M12 6v6l4 2"/>
          </svg>
        </span>
      </div>
    </aside>

    <!-- main content -->
    <main class="mx-auto w-full max-w-[480px] flex-1 bg-white/80 p-4 pb-24">
      <div class="max-w-2xl">
        <!-- drafts card -->
        <section class="mb-4">
          <div class="rounded-lg border border-line bg-surface p-4">
            <div class="mb-3 flex items-center justify-between gap-2">
              <h3 class="text-h5 text-ink">Draft manifests</h3>
              <UiPill tone="info">{{ drafts.length }}</UiPill>
            </div>
            <p v-if="attentionLoading" class="text-label text-ink-64">Loading drafts…</p>
            <p v-else-if="attentionError" class="text-label text-danger">{{ attentionError }}</p>
            <p v-else-if="drafts.length === 0" class="text-label text-ink-64">No draft manifests.</p>
            <div v-else class="space-y-2">
              <button
                v-for="manifest in drafts"
                :key="manifest.id"
                type="button"
                class="flex w-full items-center justify-between rounded bg-surface/5 p-2 text-left transition hover:bg-canvas focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
                @click="openManifest(manifest)"
              >
                <span class="flex min-w-0 items-center gap-3">
                  <span class="truncate text-label text-ink-64">{{ manifest.manifest_number }}</span>
                  <span class="text-meta text-ink-64">{{ manifest.item_count }} item{{ manifest.item_count === 1 ? '' : 's' }}</span>
                </span>
                <UiPill tone="info" class="shrink-0">Draft</UiPill>
              </button>
            </div>
          </div>
        </section>

        <!-- needs attention card -->
        <section class="mb-4">
          <div class="rounded-lg border border-line bg-surface p-4">
            <div class="flex items-center justify-between gap-2 mb-3">
              <h3 class="text-h5 text-ink">Needs attention</h3>
              <svg class="size-4 text-ink-40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <circle cx="12" cy="12" r="10"/>
                <path d="M12 6v6l4 2"/>
              </svg>
            </div>
            <p class="mb-3 text-label text-ink-64">Review-required manifests with exceptions.</p>
            <p v-if="attentionLoading" class="text-label text-ink-64">Loading manifests…</p>
            <p v-else-if="attentionError" class="text-label text-danger">{{ attentionError }}</p>
            <p v-else-if="attention.length === 0" class="text-label text-ink-64">No manifests need attention.</p>
            <div v-else class="space-y-2">
              <button
                v-for="manifest in attention"
                :key="manifest.id"
                type="button"
                class="flex w-full items-center justify-between rounded bg-surface/5 p-2 text-left transition hover:bg-canvas focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
                @click="openManifest(manifest)"
              >
                <span class="flex min-w-0 items-center gap-3">
                  <span class="truncate text-label text-ink-64">{{ manifest.manifest_number }}</span>
                  <span class="text-meta text-ink-64">{{ manifest.exception_count }} exception{{ manifest.exception_count === 1 ? '' : 's' }}</span>
                </span>
                <UiPill tone="danger" class="shrink-0">
                  {{ attentionLabel(manifest) }}
                </UiPill>
              </button>
            </div>
          </div>
        </section>

        <!-- how goods get here (desktop only: mobile dashboard stays focused) -->
        <section class="mt-10 hidden md:block">
          <h2 class="text-h4 text-ink">How goods get here</h2>
          <div class="mt-4 flex gap-4">
            <!-- card 01 -->
            <div class="flex w-1/3 flex-col rounded-lg border border-line bg-surface p-4">
              <div class="flex h-20 w-full items-center justify-center rounded-md bg-ink/5">
                <svg class="size-8 text-ink-40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                  <line x1="16" x2="8" y1="13" y2="13"/>
                  <line x1="16" x2="8" y1="17" y2="17"/>
                </svg>
                <svg class="ml-2 size-8 text-ink-40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <rect x="5" y="2" width="14" height="20" rx="2" ry="2"/>
                  <line x1="12" x2="12" y1="18" y2="18"/>
                </svg>
              </div>
              <p class="mt-2 text-meta font-medium text-blue">01</p>
              <h3 class="mt-1 text-p1 font-medium text-ink">Take a photo of the manifest</h3>
              <p class="mt-1 text-label text-ink-64">Use the camera or your phone.</p>
            </div>

            <!-- card 02 -->
            <div class="flex w-1/3 flex-col rounded-lg border border-line bg-surface p-4">
              <div class="flex h-20 w-full items-center justify-center rounded-md bg-ink/5 relative">
                <svg class="size-8 text-ink-40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                  <line x1="16" x2="8" y1="13" y2="13"/>
                  <line x1="16" x2="8" y1="17" y2="17"/>
                </svg>
                <span class="absolute right-2 top-2 flex size-5 items-center justify-center rounded-full bg-danger text-white text-xs font-bold">!</span>
              </div>
              <p class="mt-2 text-meta font-medium text-blue">02</p>
              <h3 class="mt-1 text-p1 font-medium text-ink">Check the manifest</h3>
              <p class="mt-1 text-label text-ink-64">Fix any line the photo got wrong.</p>
            </div>

            <!-- card 03 -->
            <div class="flex w-1/3 flex-col rounded-lg border border-line bg-surface p-4">
              <div class="flex h-20 w-full items-center justify-center rounded-md bg-ink/5 relative">
                <svg class="size-8 text-ink-40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <rect x="3" y="3" width="18" height="18" rx="2"/>
                  <path d="M3 9h18"/>
                  <path d="M9 21V9"/>
                </svg>
                <span class="absolute right-2 top-2 flex size-5 items-center justify-center rounded-full bg-blue text-white text-xs font-bold">✓</span>
                <span class="absolute right-8 top-2 text-meta text-ink-64 text-xs">2 of 3 here</span>
              </div>
              <p class="mt-2 text-meta font-medium text-blue">03</p>
              <h3 class="mt-1 text-p1 font-medium text-ink">Confirm what arrived</h3>
              <p class="mt-1 text-label text-ink-64">Count what arrived. We tell the receivers.</p>
            </div>
          </div>
        </section>

        <div class="mt-8 flex-1"></div>
      </div>
    </main>

    <!-- hidden camera input (mobile + desktop) -->
    <input
      ref="cameraInput"
      type="file"
      accept="image/*"
      capture="environment"
      class="hidden"
      aria-hidden="true"
      tabindex="-1"
      @change="onPhotoSelected"
    />

    <!-- manual manifest entry: full-screen, mobile-first sheet -->
    <Teleport to="body">
      <div
        v-if="manualOpen"
        class="fixed inset-0 z-50 flex flex-col bg-canvas"
        role="dialog"
        aria-modal="true"
        :aria-label="manifestReadOnly ? 'View manifest' : editingManifestId ? 'Edit manifest' : 'Enter manifest manually'"
      >
        <!-- sticky header -->
        <header class="flex items-start gap-3 border-b border-line bg-surface px-4 pt-4 pb-3">
          <button
            type="button"
            class="-ml-2 flex size-10 shrink-0 items-center justify-center rounded-lg text-ink transition duration-150 hover:bg-canvas focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
            aria-label="Close"
            @click="closeManual"
          >
            <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true">
              <path d="M18 6 6 18M6 6l12 12"/>
            </svg>
          </button>
          <div class="min-w-0">
            <h2 class="text-h5 text-ink">{{ manifestReadOnly ? `Manifest ${manifestNumber}` : editingManifestId ? `Edit manifest ${manifestNumber}` : 'New manifest' }}</h2>
            <p class="mt-0.5 text-label text-ink-64">{{ manifestReadOnly ? 'View manifest details.' : editingManifestId ? 'Update the waybill and its parcels.' : 'Fill the waybill and the parcels on it.' }}</p>
          </div>
          <span class="ml-auto mt-1 inline-flex items-center rounded bg-ink/5 px-1.5 text-meta font-normal text-ink-64 shadow-inset">Esc</span>
        </header>

        <!-- scrollable body -->
        <div class="flex-1 overflow-y-auto overscroll-contain px-4 py-4">
          <fieldset :disabled="manifestReadOnly" class="min-w-0 space-y-6">
            <!-- waybill details -->
            <section>
              <h3 class="text-label font-medium text-ink">Waybill details</h3>
              <div class="mt-3 space-y-4">
                <UiField
                  id="manual-number"
                  v-model="manifestNumber"
                  label="Manifest number"
                  type="text"
                  inputmode="numeric"
                  :maxlength="5"
                  placeholder="Auto-generated"
                  hint="Leave empty for a unique 5-digit number, or type one (5 digits)."
                  :error="''"
                />
                <UiField
                  id="manual-date"
                  v-model="manifestDate"
                  label="Manifest date"
                  type="text"
                  placeholder="YYYY-MM-DD"
                  hint="Today’s date is filled in for you."
                  :error="''"
                />
                <UiSelect
                  id="manual-origin"
                  v-model="originBranchId"
                  label="Sending from"
                  :options="branchSelectOptions"
                  :hint="branchesLoading ? 'Loading branches…' : (branches.length ? '' : 'No branches available yet')"
                  :error="''"
                />
                <UiSelect
                  id="manual-destination"
                  v-model="manifestDestination"
                  label="To (receiving branch)"
                  :options="destinationSelectOptions"
                  hint="Set once for the whole manifest, or leave as ‘Set per item’ to choose on each item."
                  :error="''"
                />
                <UiSelect
                  id="manual-status"
                  v-model="manifestStatus"
                  label="Status"
                  :options="manifestStatusOptions"
                  :error="''"
                />
              </div>
            </section>

            <!-- items -->
            <section>
              <div class="flex items-center justify-between gap-3">
                <h3 class="text-label font-medium text-ink">
                  Items <span class="text-ink-64">({{ manifestItems.length }})</span>
                </h3>
              </div>

              <div class="mt-3 space-y-4">
                <div
                  v-for="(item, index) in manifestItems"
                  :key="index"
                  class="rounded-lg border border-line bg-surface p-4"
                >
                  <div class="flex items-center justify-between gap-2">
                    <span class="text-meta font-medium text-blue">ITEM {{ index + 1 }}</span>
                    <button
                      v-if="manifestItems.length > 1 && !manifestReadOnly"
                      type="button"
                      class="flex size-10 items-center justify-center rounded-lg text-ink-64 transition duration-150 hover:bg-red/5 hover:text-danger focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
                      :aria-label="`Remove item ${index + 1}`"
                      @click="removeManifestItem(index)"
                    >
                      <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                        <path d="M3 6h18"/>
                        <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                        <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/>
                      </svg>
                    </button>
                  </div>

                  <div class="mt-3 space-y-4">
                    <UiField
                      :id="`item-${index}-description`"
                      v-model="item.description"
                      label="Description"
                      type="text"
                      placeholder="e.g. Bags of rice"
                      :error="''"
                    />

                    <div class="grid grid-cols-2 gap-3">
                      <UiField
                        :id="`item-${index}-quantity`"
                        v-model="item.quantity"
                        label="Quantity"
                        type="text"
                        inputmode="numeric"
                        placeholder="1"
                        :error="''"
                      />
                      <UiField
                        :id="`item-${index}-serial`"
                        v-model="item.serial_number"
                        label="Waybill No"
                        type="text"
                        placeholder="Optional"
                        :error="''"
                      />
                    </div>

                    <UiField
                      :id="`item-${index}-sender`"
                      v-model="item.sender_name"
                      label="Sender name"
                      type="text"
                      placeholder="Full name"
                      :error="''"
                    />
                    <UiField
                      :id="`item-${index}-sender-phone`"
                      v-model="item.sender_phone"
                      label="Sender phone"
                      type="tel"
                      placeholder="Optional"
                      :error="''"
                    />

                    <UiField
                      :id="`item-${index}-receiver`"
                      v-model="item.receiver_name"
                      label="Receiver name"
                      type="text"
                      placeholder="Full name"
                      :error="''"
                    />
                    <UiField
                      :id="`item-${index}-receiver-phone`"
                      v-model="item.receiver_phone"
                      label="Receiver phone"
                      type="tel"
                      placeholder="Optional"
                      :error="''"
                    />

                    <p
                      v-if="manifestDestination"
                      class="space-y-2"
                    >
                      <span class="block text-label text-ink-64">Going to</span>
                      <span class="block h-10 rounded-lg border border-line bg-canvas px-3 text-p1 leading-10 text-ink truncate">{{ manifestDestinationLabel }}</span>
                    </p>
                    <UiSelect
                      v-else
                      :id="`item-${index}-destination`"
                      v-model="item.destination_branch_id"
                      label="Going to"
                      :options="branchSelectOptions"
                      :hint="branches.length ? '' : 'No branches available yet'"
                      :error="''"
                    />

                    <div class="grid grid-cols-2 gap-3">
                      <UiField
                        :id="`item-${index}-amount`"
                        v-model="item.amount"
                        label="Amount (₦)"
                        type="text"
                        inputmode="decimal"
                        placeholder="Optional"
                        :error="''"
                      />
                      <UiField
                        :id="`item-${index}-currency`"
                        v-model="item.currency"
                        label="Currency"
                        type="text"
                        placeholder="NGN"
                        :error="''"
                      />
                    </div>
                  </div>
                </div>
              </div>

              <UiButton
                v-if="!manifestReadOnly"
                type="button"
                variant="tertiary"
                block
                class="mt-4"
                @click="addManifestItem"
              >
                <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" aria-hidden="true">
                  <path d="M12 5v14M5 12h14"/>
                </svg>
                Add another item
              </UiButton>
            </section>

            <p v-if="manualError" class="text-label text-danger">{{ manualError }}</p>

            <div class="h-2"></div>
          </fieldset>
        </div>

        <!-- sticky footer -->
        <footer class="border-t border-line bg-surface px-4 pt-3 pb-[max(env(safe-area-inset-bottom),0.75rem)]">
          <UiButton
            v-if="!manifestReadOnly"
            type="button"
            block
            :loading="manualSaving || manifestLoading"
            :disabled="branchesLoading || manifestLoading"
            @click="saveManifest"
          >
            {{ manifestLoading ? 'Loading…' : manualSaving ? 'Saving…' : editingManifestId ? 'Save changes' : 'Save manifest' }}
            <UiKbd>Enter</UiKbd>
          </UiButton>
          <div class="mt-3 flex justify-center">
            <UiButton type="button" variant="link" @click="closeManual">
              {{ manifestReadOnly ? 'Close' : 'Cancel' }}
            </UiButton>
          </div>
        </footer>
      </div>
    </Teleport>



    <!-- send action sheet -->
    <Teleport to="body">
      <div
        v-if="sendSheetOpen"
        class="fixed inset-0 z-60 flex items-end justify-center bg-ink/30 px-4 py-6 transition-opacity"
        @click.self="sendSheetOpen = false"
        role="dialog"
        aria-modal="true"
        aria-label="Send a parcel"
      >
        <div
          class="w-full max-w-sm rounded-t-2xl bg-surface shadow-raised border-t-2 border-line"
        >
          <div class="flex items-center justify-between px-4 pt-3 pb-1">
            <h3 class="text-h5 text-ink">Send a parcel</h3>
            <button
              type="button"
              class="flex size-10 items-center justify-center rounded-lg text-ink transition duration-150 hover:bg-canvas focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
              aria-label="Close"
              @click="sendSheetOpen = false"
            >
              <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true">
                <path d="M18 6 6 18M6 6l12 12"/>
              </svg>
            </button>
          </div>

          <div class="px-4 pb-4 flex flex-col gap-2">
            <button
              type="button"
              class="flex flex-col gap-1 rounded-lg bg-canvas px-4 py-4 text-left transition hover:bg-surface"
              @click="openCamera; sendSheetOpen = false"
            >
              <span class="flex items-center gap-2">
                <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"/>
                  <circle cx="12" cy="13" r="4"/>
                </svg>
                <span class="text-p1 text-ink">Take a photo of the parcel</span>
              </span>
              <span class="text-label text-ink-64">
                Use the camera. The photo is saved to <span class="font-medium text-ink">/photos</span>.
              </span>
            </button>

            <button
              type="button"
              class="flex flex-col gap-1 rounded-lg bg-canvas px-4 py-4 text-left transition hover:bg-surface"
              @click="$event => { openManual(); sendSheetOpen = false }"
            >
              <span class="flex items-center gap-2">
                <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                  <line x1="16" x2="8" y1="13" y2="13"/>
                  <line x1="16" x2="8" y1="17" y2="17"/>
                </svg>
                <span class="text-p1 text-ink">Add a manifest manually</span>
              </span>
              <span class="text-label text-ink-64">
                Create a manifest and its items by hand.
              </span>
            </button>
          </div>

          <div class="px-4 pb-4">
            <button
              type="button"
              class="w-full rounded-lg border border-line bg-surface py-3 text-button text-ink-64 transition hover:bg-canvas hover:text-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
              @click="sendSheetOpen = false"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- receive action sheet -->
    <Teleport to="body">
      <div
        v-if="receiveSheetOpen"
        class="fixed inset-0 z-60 flex items-end justify-center bg-ink/30 px-4 py-6 transition-opacity"
        @click.self="receiveSheetOpen = false"
        role="dialog"
        aria-modal="true"
        aria-label="Receive a parcel"
      >
        <div
          class="w-full max-w-sm rounded-t-2xl bg-surface shadow-raised border-t-2 border-line"
        >
          <div class="flex items-center justify-between px-4 pt-3 pb-1">
            <h3 class="text-h5 text-ink">Receive a parcel</h3>
            <button
              type="button"
              class="flex size-10 items-center justify-center rounded-lg text-ink transition duration-150 hover:bg-canvas focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
              aria-label="Close"
              @click="receiveSheetOpen = false"
            >
              <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true">
                <path d="M18 6 6 18M6 6l12 12"/>
              </svg>
            </button>
          </div>

          <div class="px-4 pb-4 flex flex-col gap-2">
            <button
              type="button"
              class="flex flex-col gap-1 rounded-lg bg-canvas px-4 py-4 text-left transition hover:bg-surface"
              @click="openCamera; receiveSheetOpen = false"
            >
              <span class="flex items-center gap-2">
                <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"/>
                  <circle cx="12" cy="13" r="4"/>
                </svg>
                <span class="text-p1 text-ink">Take a photo of the parcel</span>
              </span>
              <span class="text-label text-ink-64">
                Use the camera. The photo is saved to <span class="font-medium text-ink">/photos</span>.
              </span>
            </button>

            <button
              type="button"
              class="flex flex-col gap-1 rounded-lg bg-canvas px-4 py-4 text-left transition hover:bg-surface"
              @click="$event => { openManual(); receiveSheetOpen = false }"
            >
              <span class="flex items-center gap-2">
                <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                  <line x1="16" x2="8" y1="13" y2="13"/>
                  <line x1="16" x2="8" y1="17" y2="17"/>
                </svg>
                <span class="text-p1 text-ink">Add a manifest manually</span>
              </span>
              <span class="text-label text-ink-64">
                Create a manifest and its items by hand.
              </span>
            </button>
          </div>

          <div class="px-4 pb-4">
            <button
              type="button"
              class="w-full rounded-lg border border-line bg-surface py-3 text-button text-ink-64 transition hover:bg-canvas hover:text-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue"
              @click="receiveSheetOpen = false"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- bottom nav (mobile only) -->
    <nav class="md:hidden fixed bottom-0 left-0 right-0 border-t border-line bg-surface z-10">
      <div class="flex items-center justify-around h-16 pb-[max(env(safe-area-inset-bottom),0)]">
        <!-- send -->
        <button
          type="button"
          class="relative flex flex-col items-center gap-1 py-2"
          @click="openSendActionSheet"
          aria-label="Send a parcel"
        >
          <svg class="size-6 text-ink-64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M17 3h5v8h-5V3zm-4 11h8v2h-8v-2zm0-4h8v6h-8v-6zm-4-4h8v2h-8V7zm0-4h8v2h-8V3z"/>
            <polyline points="7 10 12 15 17 10"/>
            <line x1="12" x2="12" y1="15" y2="3"/>
          </svg>
          <span class="text-xs text-ink-64">Send</span>
          <span class="absolute  -left-0.5 -top-0.5 block h-3.5 w-3.5 rounded-full bg-blue shadow-raised ring-1 ring-white" />
        </button>

        <!-- receive -->
        <button
          type="button"
          class="relative flex flex-col items-center gap-1 py-2"
          @click="openReceiveActionSheet"
          aria-label="Receive a parcel"
        >
          <svg class="size-6 text-ink-64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7 10 12 15 17 10"/>
            <line x1="12" x2="12" y1="15" y2="3"/>
          </svg>
          <span class="text-xs text-ink-64">Receive</span>
          <span class="absolute  -left-0.5 -top-0.5 block h-3.5 w-3.5 rounded-full bg-blue shadow-raised ring-1 ring-white" />
        </button>

        <!-- records -->
        <NuxtLink
          to="/records"
          class="flex flex-col items-center gap-1 py-2"
        >
          <svg class="size-6 text-ink-64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M8 3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-3"/>
            <path d="M16 3h5v5"/>
            <path d="m10 14 11-11"/>
            <path d="M7 8h4M7 12h2"/>
          </svg>
          <span class="text-xs text-ink-64">Records</span>
        </NuxtLink>

        <!-- settings -->
        <NuxtLink
          to="/settings"
          class="flex flex-col items-center gap-1 py-2"
        >
          <svg class="size-6 text-ink-64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="3"/>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9c0 .64-.52 1.16-1.17 1.34l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82V3a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09a1.65 1.65 0 0 0 1.51 1z"/>
          </svg>
          <span class="text-xs text-ink-64">Settings</span>
        </NuxtLink>
      </div>
    </nav>
    </div>
    <div class="hidden min-h-dvh items-center justify-center bg-canvas px-6 text-center md:flex">
      <p class="text-label text-ink-64">Galopin is available on mobile screens.</p>
    </div>
  </div>
</template>
