import type { CustomerBooking, CustomerRecentAddress, DeliveryStatus } from '~/types/customer'

export interface BookingStateMeta {
  label: string
  tone: string
  dot: string
  active: boolean
}

const terminalStatuses: DeliveryStatus[] = ['delivered', 'cancelled']

function bookingState(booking: CustomerBooking): BookingStateMeta {
  if (booking.request_status === 'pending') {
    return {
      label: 'Awaiting provider',
      tone: 'border-amber-200 bg-amber-50 text-amber-700',
      dot: 'bg-amber-500',
      active: true,
    }
  }
  if (booking.request_status === 'declined' || booking.status === 'cancelled') {
    return {
      label: booking.request_status === 'declined' ? 'Declined' : 'Cancelled',
      tone: 'border-rose-200 bg-rose-50 text-rose-700',
      dot: 'bg-rose-500',
      active: false,
    }
  }
  switch (booking.status) {
    case 'requested':
    case 'created':
      return {
        label: 'Confirmed',
        tone: 'border-slate-200 bg-slate-50 text-slate-600',
        dot: 'bg-slate-400',
        active: true,
      }
    case 'assigned':
      return {
        label: 'Rider assigned',
        tone: 'border-sky-200 bg-sky-50 text-sky-700',
        dot: 'bg-sky-500',
        active: true,
      }
    case 'picked_up':
      return {
        label: 'Picked up',
        tone: 'border-sky-200 bg-sky-50 text-sky-700',
        dot: 'bg-sky-500',
        active: true,
      }
    case 'in_transit':
      return {
        label: 'In transit',
        tone: 'border-brand-200 bg-brand-50 text-brand-700',
        dot: 'bg-brand-500',
        active: true,
      }
    case 'delivered':
      return {
        label: 'Delivered',
        tone: 'border-emerald-200 bg-emerald-50 text-emerald-700',
        dot: 'bg-emerald-500',
        active: false,
      }
    default:
      return {
        label: 'Cancelled',
        tone: 'border-rose-200 bg-rose-50 text-rose-700',
        dot: 'bg-rose-500',
        active: false,
      }
  }
}

export function formatBookingMoney(minor: number, currency = 'NGN') {
  return new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  }).format(minor / 100)
}

export function bookingStatusLabel(status: DeliveryStatus) {
  return status.replaceAll('_', ' ')
}

export function useCustomerDashboard() {
  const { api } = useApi()
  const bookings = ref<CustomerBooking[]>([])
  const loading = ref(false)
  const error = ref('')
  const loaded = ref(false)

  async function load(force = false) {
    if (loading.value || (loaded.value && !force)) return
    loading.value = true
    error.value = ''
    try {
      const result = await api<{ bookings: CustomerBooking[] }>('/customer/bookings')
      bookings.value = result.bookings
      loaded.value = true
    } catch (cause) {
      error.value = apiErrorMessage(cause, 'Could not load your bookings.')
    } finally {
      loading.value = false
    }
  }

  const activeBookings = computed(() =>
    bookings.value.filter((booking) => bookingState(booking).active),
  )

  const inTransitCount = computed(
    () =>
      bookings.value.filter(
        (booking) =>
          booking.request_status === 'accepted' &&
          ['assigned', 'picked_up', 'in_transit'].includes(booking.status),
      ).length,
  )

  const deliveredBookings = computed(() =>
    bookings.value.filter((booking) => booking.status === 'delivered'),
  )

  const providers = computed(() => {
    const seen = new Set<string>()
    return bookings.value.filter((booking) => {
      if (seen.has(booking.provider_id)) return false
      seen.add(booking.provider_id)
      return true
    })
  })

  const pendingBookings = computed(() =>
    bookings.value.filter(
      (booking) => booking.request_status === 'pending' || booking.status === 'requested',
    ),
  )

  const cancelledBookings = computed(() =>
    bookings.value.filter(
      (booking) => booking.status === 'cancelled' || booking.request_status === 'declined',
    ),
  )

  const recentBookings = computed(() =>
    [...bookings.value]
      .sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime())
      .slice(0, 5),
  )

  const addresses = computed<CustomerRecentAddress[]>(() => {
    const seen = new Set<string>()
    const result: CustomerRecentAddress[] = []
    for (const booking of bookings.value) {
      const entries: { address: string; role: CustomerRecentAddress['role'] }[] = [
        { address: booking.pickup_address, role: 'pickup' },
        { address: booking.dropoff_address, role: 'dropoff' },
      ]
      for (const entry of entries) {
        if (!entry.address) continue
        const key = `${entry.role}:${entry.address.toLowerCase()}`
        if (seen.has(key)) continue
        seen.add(key)
        result.push({
          address: entry.address,
          role: entry.role,
          provider: booking.provider_name,
          lastUsedAt: booking.created_at,
        })
      }
    }
    return result
  })

  function stateFor(booking: CustomerBooking) {
    return bookingState(booking)
  }

  function formatMoney(booking: CustomerBooking) {
    return formatBookingMoney(booking.delivery_price_minor, booking.currency)
  }

  function isTerminal(booking: CustomerBooking) {
    return terminalStatuses.includes(booking.status)
  }

  return {
    bookings,
    loading,
    error,
    loaded,
    load,
    activeBookings,
    pendingBookings,
    cancelledBookings,
    recentBookings,
    inTransitCount,
    deliveredBookings,
    providers,
    addresses,
    stateFor,
    formatMoney,
    isTerminal,
  }
}
