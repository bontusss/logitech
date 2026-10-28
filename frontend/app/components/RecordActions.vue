<script setup lang="ts">
import type { Shipment } from '~/types/api'
import { ApiError } from '~/utils/api'

/**
 * The actions a record allows, as buttons plus the dialog each one opens.
 * Visibility is decided by the caller; the page re-fetches after an update.
 */
const props = defineProps<{
  record: Shipment
  canReceive: boolean
  canDispatch: boolean
  canEdit: boolean
  canCancel: boolean
}>()

const emit = defineEmits<{ updated: [] }>()

const { token } = useAuth()
const { options: branchOptions, load: loadBranches } = useBranches()

type Action = 'receive' | 'dispatch' | 'payment' | 'edit' | 'exception' | 'cancel'
const action = ref<Action | null>(null)
const saving = ref(false)
const error = ref('')

const paymentMethods = [
  { value: 'cash', label: 'Cash' },
  { value: 'bank_transfer', label: 'Bank transfer' },
  { value: 'pos', label: 'POS' },
  { value: 'other', label: 'Other' },
]

const exceptionTypes = [
  { value: 'missing_goods', label: 'Missing goods' },
  { value: 'damaged_goods', label: 'Damaged goods' },
  { value: 'quantity_mismatch', label: 'Quantity mismatch' },
  { value: 'wrong_goods', label: 'Wrong goods' },
  { value: 'wrong_destination', label: 'Wrong destination' },
  { value: 'delivery_delay', label: 'Delivery delay' },
  { value: 'document_mismatch', label: 'Document mismatch' },
  { value: 'payment_discrepancy', label: 'Payment discrepancy' },
  { value: 'other', label: 'Other' },
]

const severityOptions = [
  { value: 'low', label: 'Low' },
  { value: 'medium', label: 'Medium' },
  { value: 'high', label: 'High' },
  { value: 'critical', label: 'Critical' },
]

const conditions = [
  { value: 'good', label: 'Good' },
  { value: 'damaged', label: 'Damaged' },
  { value: 'missing', label: 'Missing' },
]

const received = computed(() => props.record.status === 'received' || Boolean(props.record.received_at))
const awaitingDelivery = computed(() => props.record.status !== 'cancelled' && !received.value)
const hasBalance = computed(() => props.record.balance > 0 && props.record.status !== 'cancelled')

// Receive
const receivedQuantity = ref('')
const condition = ref('good')
const collectBalance = ref(true)
const collectAmount = ref('')
const collectMethod = ref('cash')
const collectReference = ref('')

// Dispatch / receive notes
const notes = ref('')

// Payment
const paymentAmount = ref('')
const paymentMethod = ref('cash')
const paymentReference = ref('')

// Edit
const edit = reactive({
  description: '',
  quantity: '',
  receiverName: '',
  receiverPhone: '',
  destination: '',
  amount: '',
  notes: '',
})

// Exception
const exception = reactive({
  type: 'damaged_goods',
  severity: 'medium',
  title: '',
  description: '',
  comment: '',
})

const titles: Record<Action, string> = {
  receive: 'Receive this record',
  dispatch: 'Dispatch this record',
  payment: 'Record a payment',
  edit: 'Edit this record',
  exception: 'Raise an exception',
  cancel: 'Cancel this record',
}

function open(next: Action) {
  action.value = next
  error.value = ''
  notes.value = ''
  const record = props.record

  if (next === 'receive') {
    receivedQuantity.value = String(record.received_quantity ?? record.quantity)
    condition.value = record.condition || 'good'
    collectBalance.value = record.balance > 0
    collectAmount.value = String(record.balance)
    collectMethod.value = 'cash'
    collectReference.value = ''
  }
  if (next === 'payment') {
    paymentAmount.value = String(record.balance)
    paymentMethod.value = 'cash'
    paymentReference.value = ''
  }
  if (next === 'edit') {
    edit.description = record.description
    edit.quantity = String(record.quantity)
    edit.receiverName = record.receiver.full_name
    edit.receiverPhone = record.receiver.phone ?? ''
    edit.destination = record.destination_branch_id
    edit.amount = String(record.amount)
    edit.notes = record.notes ?? ''
  }
  if (next === 'exception') {
    exception.title = ''
    exception.description = ''
    exception.comment = ''
    exception.type = record.condition === 'damaged' ? 'damaged_goods' : 'quantity_mismatch'
    exception.severity = 'medium'
  }
}

function close() {
  if (saving.value) return
  action.value = null
  error.value = ''
}

async function run<T>(request: () => Promise<T>, closeAfter = true): Promise<T | null> {
  if (!token.value) return null
  saving.value = true
  error.value = ''
  try {
    const result = await request()
    if (closeAfter) close()
    emit('updated')
    return result
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'That did not work. Please try again.'
    return null
  } finally {
    saving.value = false
  }
}

function receive() {
  const quantity = Number(receivedQuantity.value)
  if (!Number.isInteger(quantity) || quantity < 0) {
    error.value = 'Received quantity must be zero or more.'
    return
  }
  const amount = Number(collectAmount.value || 0)
  if (collectBalance.value && (Number.isNaN(amount) || amount <= 0)) {
    error.value = 'Enter the balance collected, or untick the box.'
    return
  }

  void run(async () => {
    await apiRequest<Shipment>(`/api/records/${props.record.id}/receive`, {
      method: 'POST',
      token: token.value,
      body: {
        received_quantity: quantity,
        condition: condition.value,
        notes: notes.value.trim() || undefined,
        payment:
          collectBalance.value && amount > 0
            ? { amount, method: collectMethod.value, reference: collectReference.value.trim() || undefined }
            : undefined,
      },
    })
  })
}

function dispatch() {
  void run(() =>
    apiRequest<Shipment>(`/api/records/${props.record.id}/dispatch`, {
      method: 'POST',
      token: token.value,
      body: { notes: notes.value.trim() || undefined },
    }),
  )
}

function recordPayment() {
  const amount = Number(paymentAmount.value)
  if (Number.isNaN(amount) || amount <= 0) {
    error.value = 'Enter an amount greater than zero.'
    return
  }
  void run(() =>
    apiRequest<Shipment>(`/api/records/${props.record.id}/payments`, {
      method: 'POST',
      token: token.value,
      body: {
        amount,
        method: paymentMethod.value,
        reference: paymentReference.value.trim() || undefined,
      },
    }),
  )
}

function saveEdit() {
  const quantity = Number(edit.quantity)
  const amount = Number(edit.amount)
  if (!edit.description.trim()) {
    error.value = 'Description is required.'
    return
  }
  if (!Number.isInteger(quantity) || quantity < 1) {
    error.value = 'Quantity must be at least 1.'
    return
  }
  if (!edit.receiverName.trim()) {
    error.value = "The receiver's name is required."
    return
  }
  if (Number.isNaN(amount) || amount < 0) {
    error.value = 'Enter a valid amount.'
    return
  }
  void run(() =>
    apiRequest<Shipment>(`/api/records/${props.record.id}`, {
      method: 'PUT',
      token: token.value,
      body: {
        description: edit.description.trim(),
        quantity,
        receiver: { full_name: edit.receiverName.trim(), phone: edit.receiverPhone.trim() },
        destination_branch_id: edit.destination,
        amount,
        notes: edit.notes.trim(),
      },
    }),
  )
}

function raiseException() {
  if (!exception.title.trim()) {
    error.value = 'Give the exception a short title.'
    return
  }
  void run(() =>
    apiRequest(`/api/records/${props.record.id}/exceptions`, {
      method: 'POST',
      token: token.value,
      body: {
        type: exception.type,
        severity: exception.severity,
        title: exception.title.trim(),
        description: exception.description.trim() || undefined,
        comment: exception.comment.trim() || undefined,
      },
    }),
  )
}

function cancelRecord() {
  void run(() =>
    apiRequest<Shipment>(`/api/records/${props.record.id}/cancel`, {
      method: 'POST',
      token: token.value,
    }),
  )
}

watch(action, (next) => {
  if (next === 'edit') void loadBranches()
})

const destinationOptions = computed(() => [
  { value: '', label: 'Choose a branch' },
  ...branchOptions.value.filter((option) => option.value !== props.record.origin_branch_id),
])

defineExpose({ open })
</script>

<template>
  <div class="flex flex-wrap items-center gap-2">
    <UiButton v-if="canReceive && awaitingDelivery" @click="open('receive')">
      <UiIcon name="receive" :size="17" />
      Receive
    </UiButton>
    <UiButton v-else-if="canDispatch && record.status === 'pending'" @click="open('dispatch')">
      <UiIcon name="send" :size="17" />
      Dispatch
    </UiButton>

    <UiButton v-if="hasBalance" variant="tertiary" @click="open('payment')">
      <UiIcon name="wallet" :size="17" />
      Record payment
    </UiButton>
    <UiButton v-if="canEdit && record.status === 'pending'" variant="tertiary" @click="open('edit')">
      <UiIcon name="edit" :size="17" />
      Edit
    </UiButton>
    <UiButton v-if="record.status !== 'cancelled'" variant="tertiary" @click="open('exception')">
      <UiIcon name="alert" :size="17" />
      Raise exception
    </UiButton>
    <UiButton v-if="canCancel" variant="link" class="text-danger" @click="open('cancel')">Cancel record</UiButton>
  </div>

  <!-- Receive -->
  <UiModal
    :open="action === 'receive'"
    :title="titles.receive"
    :description="`${record.tracking_number} · ${record.description}`"
    size="md"
    :dismissible="!saving"
    @close="close"
  >
    <form class="space-y-4" novalidate @submit.prevent="receive">
      <div v-if="error" class="rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
        {{ error }}
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <UiField id="received-quantity" v-model="receivedQuantity" label="Quantity received" inputmode="numeric" autocomplete="off" />
        <UiSelect id="received-condition" v-model="condition" label="Condition" :options="conditions" />
      </div>

      <div class="rounded-lg border border-line p-3">
        <p class="text-label font-medium text-ink">Balance due {{ money(record.balance, record.currency) }}</p>
        <label v-if="record.balance > 0" class="mt-2 flex items-center gap-2 text-label text-ink-64">
          <input v-model="collectBalance" type="checkbox" class="size-4 accent-[#2674da]">
          Collect the balance now
        </label>
        <div v-if="collectBalance && record.balance > 0" class="mt-3 grid gap-3 sm:grid-cols-3">
          <UiField id="collect-amount" v-model="collectAmount" label="Amount (₦)" inputmode="decimal" autocomplete="off" />
          <UiSelect id="collect-method" v-model="collectMethod" label="Method" :options="paymentMethods" />
          <UiField id="collect-reference" v-model="collectReference" label="Reference" autocomplete="off" />
        </div>
      </div>

      <UiField id="receive-notes" v-model="notes" label="Notes (optional)" autocomplete="off" />
    </form>

    <template #footer>
      <div class="flex flex-col gap-2 sm:flex-row-reverse">
        <UiButton block :loading="saving" @click="receive">Confirm receipt</UiButton>
        <UiButton variant="tertiary" block :disabled="saving" @click="close">Close</UiButton>
      </div>
    </template>
  </UiModal>

  <!-- Dispatch -->
  <UiModal
    :open="action === 'dispatch'"
    :title="titles.dispatch"
    :description="`${record.tracking_number} · ${record.description}`"
    :dismissible="!saving"
    @close="close"
  >
    <div v-if="error" class="mb-3 rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
      {{ error }}
    </div>
    <UiField id="dispatch-notes" v-model="notes" label="Dispatch notes (optional)" autocomplete="off" />
    <template #footer>
      <div class="flex flex-col gap-2 sm:flex-row-reverse">
        <UiButton block :loading="saving" @click="dispatch">Mark as dispatched</UiButton>
        <UiButton variant="tertiary" block :disabled="saving" @click="close">Close</UiButton>
      </div>
    </template>
  </UiModal>

  <!-- Payment -->
  <UiModal
    :open="action === 'payment'"
    :title="titles.payment"
    :description="`${record.tracking_number} · balance ${money(record.balance, record.currency)}`"
    :dismissible="!saving"
    @close="close"
  >
    <div v-if="error" class="mb-3 rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
      {{ error }}
    </div>
    <form class="space-y-4" novalidate @submit.prevent="recordPayment">
      <UiField id="payment-amount" v-model="paymentAmount" label="Amount (₦)" inputmode="decimal" autocomplete="off" />
      <UiSelect id="payment-method" v-model="paymentMethod" label="Method" :options="paymentMethods" />
      <UiField id="payment-reference" v-model="paymentReference" label="Reference (optional)" autocomplete="off" />
    </form>
    <template #footer>
      <div class="flex flex-col gap-2 sm:flex-row-reverse">
        <UiButton block :loading="saving" @click="recordPayment">Save payment</UiButton>
        <UiButton variant="tertiary" block :disabled="saving" @click="close">Close</UiButton>
      </div>
    </template>
  </UiModal>

  <!-- Edit -->
  <UiModal
    :open="action === 'edit'"
    :title="titles.edit"
    description="Only a record that has not been dispatched can be edited."
    size="md"
    :dismissible="!saving"
    @close="close"
  >
    <div v-if="error" class="mb-3 rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
      {{ error }}
    </div>
    <form class="space-y-4" novalidate @submit.prevent="saveEdit">
      <UiField id="edit-description" v-model="edit.description" label="Description" autocomplete="off" />
      <div class="grid gap-4 sm:grid-cols-2">
        <UiField id="edit-quantity" v-model="edit.quantity" label="Quantity" inputmode="numeric" autocomplete="off" />
        <UiField id="edit-amount" v-model="edit.amount" label="Amount charged (₦)" inputmode="decimal" autocomplete="off" />
        <UiField id="edit-receiver" v-model="edit.receiverName" label="Receiver name" autocomplete="off" />
        <UiField id="edit-receiver-phone" v-model="edit.receiverPhone" label="Receiver phone" type="tel" autocomplete="off" />
      </div>
      <UiSelect id="edit-destination" v-model="edit.destination" label="Destination branch" :options="destinationOptions" />
      <UiField id="edit-notes" v-model="edit.notes" label="Notes" autocomplete="off" />
    </form>
    <template #footer>
      <div class="flex flex-col gap-2 sm:flex-row-reverse">
        <UiButton block :loading="saving" @click="saveEdit">Save changes</UiButton>
        <UiButton variant="tertiary" block :disabled="saving" @click="close">Close</UiButton>
      </div>
    </template>
  </UiModal>

  <!-- Exception -->
  <UiModal
    :open="action === 'exception'"
    :title="titles.exception"
    :description="`${record.tracking_number} · ${record.description}`"
    size="md"
    :dismissible="!saving"
    @close="close"
  >
    <div v-if="error" class="mb-3 rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
      {{ error }}
    </div>
    <form class="space-y-4" novalidate @submit.prevent="raiseException">
      <div class="grid gap-4 sm:grid-cols-2">
        <UiSelect id="exception-type" v-model="exception.type" label="What happened" :options="exceptionTypes" />
        <UiSelect id="exception-severity" v-model="exception.severity" label="Severity" :options="severityOptions" />
      </div>
      <UiField id="exception-title" v-model="exception.title" label="Short title" autocomplete="off" />
      <UiField id="exception-description" v-model="exception.description" label="Description (optional)" autocomplete="off" />
      <UiField id="exception-comment" v-model="exception.comment" label="First comment (optional)" autocomplete="off" />
      <p class="text-meta text-ink-40">The record is flagged until this exception is resolved.</p>
    </form>
    <template #footer>
      <div class="flex flex-col gap-2 sm:flex-row-reverse">
        <UiButton block :loading="saving" @click="raiseException">Raise exception</UiButton>
        <UiButton variant="tertiary" block :disabled="saving" @click="close">Close</UiButton>
      </div>
    </template>
  </UiModal>

  <!-- Cancel -->
  <UiModal
    :open="action === 'cancel'"
    :title="titles.cancel"
    description="The record stops moving. Payments already taken stay on the account."
    :dismissible="!saving"
    @close="close"
  >
    <div v-if="error" class="mb-3 rounded-lg border border-danger/25 bg-danger/5 px-3 py-2 text-label text-danger" role="alert">
      {{ error }}
    </div>
    <p class="text-label text-ink-64">
      Cancel <span class="font-medium text-ink">{{ record.tracking_number }}</span> for
      {{ record.sender.full_name }} → {{ record.receiver.full_name }}?
    </p>
    <template #footer>
      <div class="flex flex-col gap-2 sm:flex-row-reverse">
        <UiButton variant="danger" block :loading="saving" @click="cancelRecord">Cancel record</UiButton>
        <UiButton variant="tertiary" block :disabled="saving" @click="close">Keep it</UiButton>
      </div>
    </template>
  </UiModal>
</template>
