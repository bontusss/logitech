<script setup lang="ts">
import type { Shipment } from '~/types/api'
import { ApiError } from '~/utils/api'

useHead({ title: 'Send parcel' })
definePageMeta({ layout: 'default' })

const { token, user } = useAuth()
const { isOwner, branchId } = usePermissions()
const { options: branchOptions, name: branchName, load: loadBranches } = useBranches()

interface DraftItem {
  description: string
  quantity: string
  amount: string
  amountPaid: string
  paymentMethod: string
  notes: string
}

const paymentMethods = [
  { value: 'cash', label: 'Cash' },
  { value: 'bank_transfer', label: 'Bank transfer' },
  { value: 'pos', label: 'POS' },
  { value: 'other', label: 'Other' },
]

function blankItem(): DraftItem {
  return { description: '', quantity: '1', amount: '', amountPaid: '', paymentMethod: 'cash', notes: '' }
}

const sender = reactive({ fullName: '', phone: '' })
const receiver = reactive({ fullName: '', phone: '' })
const origin = ref(branchId.value)
const destination = ref('')
const destinationNotes = ref('')
const items = ref<DraftItem[]>([blankItem()])

const submitting = ref(false)
const error = ref('')

const ownBranchName = computed(() => branchName(branchId.value))

const charged = computed(() => items.value.reduce((sum, item) => sum + (Number(item.amount) || 0), 0))
const paid = computed(() =>
  items.value.reduce((sum, item) => sum + Math.min(Number(item.amountPaid) || 0, Number(item.amount) || 0), 0),
)
const balance = computed(() => Math.max(0, charged.value - paid.value))

const destinationOptions = computed(() => [
  { value: '', label: 'Choose a branch' },
  ...branchOptions.value.filter((option) => option.value !== origin.value),
])

function addItem() {
  items.value.push(blankItem())
}

function removeItem(index: number) {
  if (items.value.length === 1) return
  items.value.splice(index, 1)
}

function payInFull(item: DraftItem) {
  item.amountPaid = item.amount
}

function validate(): string {
  if (!origin.value) return 'Choose the branch sending this parcel.'
  if (!destination.value) return 'Choose the branch the parcel is going to.'
  if (destination.value === origin.value) return 'The destination must be different from the origin branch.'
  if (!sender.fullName.trim()) return "Enter the sender's name."
  if (!receiver.fullName.trim()) return "Enter the receiver's name."

  for (const [index, item] of items.value.entries()) {
    const label = `Item ${index + 1}`
    if (!item.description.trim()) return `${label}: describe what is inside.`
    const quantity = Number(item.quantity)
    if (!Number.isInteger(quantity) || quantity < 1) return `${label}: quantity must be at least 1.`
    const amount = Number(item.amount)
    if (!item.amount.trim() || Number.isNaN(amount) || amount < 0) return `${label}: enter the amount charged.`
    const amountPaid = Number(item.amountPaid || 0)
    if (Number.isNaN(amountPaid) || amountPaid < 0) return `${label}: the amount paid is not a valid number.`
    if (amountPaid > amount) return `${label}: the amount paid cannot exceed the amount charged.`
  }
  return ''
}

async function submit() {
  if (submitting.value || !token.value) return
  error.value = validate()
  if (error.value) {
    window.scrollTo({ top: 0, behavior: 'smooth' })
    return
  }

  submitting.value = true
  try {
    const created = await apiRequest<Shipment[]>('/api/records', {
      method: 'POST',
      token: token.value,
      body: {
        origin_branch_id: isOwner.value ? origin.value : user.value?.branch_id,
        destination_branch_id: destination.value,
        items: items.value.map((item) => ({
          description: item.description.trim(),
          quantity: Number(item.quantity),
          sender: { full_name: sender.fullName.trim(), phone: sender.phone.trim() },
          receiver: { full_name: receiver.fullName.trim(), phone: receiver.phone.trim() },
          amount: Number(item.amount),
          amount_paid: Number(item.amountPaid || 0),
          payment_method: Number(item.amountPaid || 0) > 0 ? item.paymentMethod : undefined,
          notes: [item.notes.trim(), destinationNotes.value.trim()].filter(Boolean).join(' · ') || undefined,
        })),
      },
    })
    await navigateTo(created.length === 1 ? `/records/${created[0]!.id}` : '/records')
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Could not save the record. Please try again.'
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await loadBranches()
  if (!origin.value && branchOptions.value[0]) origin.value = branchOptions.value[0].value
})
</script>

<template>
  <div class="space-y-5">
    <div class="space-y-3">
      <NuxtLink to="/records" class="inline-flex items-center gap-1.5 text-label text-ink-64 transition hover:text-ink">
        <UiIcon name="chevronLeft" :size="16" />
        Records
      </NuxtLink>
      <header>
        <h1 class="text-h5 font-medium tracking-tight text-ink">Send parcel</h1>
        <p class="mt-1 text-label text-ink-64">
          What the customer handed in, where it is going and the money taken at the counter.
        </p>
      </header>
    </div>

    <div v-if="error" class="rounded-lg border border-danger/25 bg-danger/5 px-4 py-3 text-label text-danger" role="alert">
      {{ error }}
    </div>

    <form class="space-y-4" novalidate @submit.prevent="submit">
      <!-- Route -->
      <section class="rounded-lg bg-surface p-4 shadow-card">
        <h2 class="text-h5 font-medium text-ink">Route</h2>
        <div class="mt-3 grid gap-4 sm:grid-cols-2">
          <UiSelect
            v-if="isOwner"
            id="origin"
            v-model="origin"
            label="Sending branch"
            :options="branchOptions"
          />
          <div v-else>
            <p class="text-label text-ink-64">Sending branch</p>
            <p class="mt-2 flex h-10 items-center text-p1 font-medium text-ink">
              {{ ownBranchName || 'Your branch' }}
            </p>
          </div>
          <UiSelect
            id="destination"
            v-model="destination"
            label="Destination branch"
            :options="destinationOptions"
          />
        </div>
      </section>

      <!-- Customer -->
      <section class="rounded-lg bg-surface p-4 shadow-card">
        <h2 class="text-h5 font-medium text-ink">Customer</h2>
        <p class="mt-0.5 text-label text-ink-64">Phone numbers let the same customer be recognised next time.</p>
        <div class="mt-3 grid gap-4 sm:grid-cols-2">
          <UiField id="sender-name" v-model="sender.fullName" label="Sender name" autocomplete="off" />
          <UiField id="sender-phone" v-model="sender.phone" label="Sender phone" type="tel" inputmode="tel" autocomplete="off" />
          <UiField id="receiver-name" v-model="receiver.fullName" label="Receiver name" autocomplete="off" />
          <UiField id="receiver-phone" v-model="receiver.phone" label="Receiver phone" type="tel" inputmode="tel" autocomplete="off" />
        </div>
      </section>

      <!-- Items -->
      <section class="space-y-3">
        <div class="flex items-end justify-between gap-3">
          <div>
            <h2 class="text-h5 font-medium text-ink">Items</h2>
            <p class="mt-0.5 text-label text-ink-64">One card per item. The balance is collected at the destination.</p>
          </div>
          <UiButton variant="tertiary" type="button" @click="addItem">
            <UiIcon name="plus" :size="16" />
            Add item
          </UiButton>
        </div>

        <div v-for="(item, index) in items" :key="index" class="rounded-lg bg-surface p-4 shadow-card">
          <div class="flex items-center justify-between gap-3">
            <p class="text-meta font-medium tracking-[0.08em] text-ink-40 uppercase">Item {{ index + 1 }}</p>
            <UiButton
              v-if="items.length > 1"
              variant="link"
              type="button"
              class="text-danger"
              @click="removeItem(index)"
            >
              Remove
            </UiButton>
          </div>

          <div class="mt-2 grid gap-4 sm:grid-cols-2">
            <UiField id="item-description" v-model="item.description" label="Description" autocomplete="off" />
            <UiField id="item-quantity" v-model="item.quantity" label="Quantity" type="text" inputmode="numeric" autocomplete="off" />
          </div>

          <div class="mt-4 grid gap-4 sm:grid-cols-3">
            <UiField id="item-amount" v-model="item.amount" label="Amount charged (₦)" inputmode="decimal" autocomplete="off" />
            <UiField id="item-paid" v-model="item.amountPaid" label="Paid now (₦)" inputmode="decimal" autocomplete="off" />
            <UiSelect
              id="item-method"
              v-model="item.paymentMethod"
              label="Payment method"
              :options="paymentMethods"
              :disabled="!Number(item.amountPaid)"
            />
          </div>

          <div class="mt-3 flex flex-wrap items-center justify-between gap-2">
            <UiButton
              v-if="Number(item.amount) > 0"
              variant="link"
              type="button"
              @click="payInFull(item)"
            >
              Paid in full
            </UiButton>
            <p class="text-label text-ink-64">
              Balance on this item
              <span class="font-medium" :class="Number(item.amount) - Number(item.amountPaid || 0) > 0 ? 'text-danger' : 'text-ink'">
                {{ money(Math.max(0, (Number(item.amount) || 0) - (Number(item.amountPaid) || 0))) }}
              </span>
            </p>
          </div>

          <div class="mt-3">
            <UiField id="item-notes" v-model="item.notes" label="Notes (optional)" autocomplete="off" />
          </div>
        </div>
      </section>

      <!-- Handover notes -->
      <section class="rounded-lg bg-surface p-4 shadow-card">
        <UiField
          id="destination-notes"
          v-model="destinationNotes"
          label="Handover notes (optional)"
          hint="Anything the receiving branch should know."
          autocomplete="off"
        />
      </section>

      <!-- Summary -->
      <section class="rounded-lg bg-surface p-4 shadow-card">
        <h2 class="text-h5 font-medium text-ink">Summary</h2>
        <dl class="mt-3 divide-y divide-line">
          <div class="flex items-center justify-between gap-4 py-2.5">
            <dt class="text-label text-ink-64">Charged</dt>
            <dd class="text-p1 font-medium text-ink">{{ money(charged) }}</dd>
          </div>
          <div class="flex items-center justify-between gap-4 py-2.5">
            <dt class="text-label text-ink-64">Paid now</dt>
            <dd class="text-p1 font-medium text-ink">{{ money(paid) }}</dd>
          </div>
          <div class="flex items-center justify-between gap-4 py-2.5">
            <dt class="text-label text-ink-64">Balance due on delivery</dt>
            <dd class="text-p1 font-medium" :class="balance > 0 ? 'text-danger' : 'text-ink'">{{ money(balance) }}</dd>
          </div>
        </dl>
        <div class="mt-4 flex flex-col gap-2 sm:flex-row-reverse">
          <UiButton type="submit" block :loading="submitting">Save record</UiButton>
          <UiButton variant="tertiary" block type="button" @click="navigateTo('/records')">Cancel</UiButton>
        </div>
      </section>
    </form>
  </div>
</template>
