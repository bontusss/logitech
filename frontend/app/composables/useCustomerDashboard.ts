import type { CustomerDashboardData, ShipmentStatus } from '~/types/customer'

/**
 * Illustrative marketplace data.
 *
 * The backend does not expose customer shipment or booking endpoints yet (only the
 * rate engine exists), so the dashboard is wired through this composable to keep
 * the shape of the eventual API in one place. Replace `sampleDashboard` with a
 * fetch to the customer dashboard endpoint once it ships.
 */
const sampleDashboard: CustomerDashboardData = {
  shipments: [
    {
      id: 'shp_2481',
      reference: 'BOL-2481',
      provider: 'Metroline Express',
      service: 'Express',
      origin: 'Aba',
      destination: 'Lagos',
      status: 'in_transit',
      currentLocation: 'Benin City transit hub',
      eta: 'Tomorrow, 4:00 PM',
      weightKg: 3,
      priceKobo: 450000,
      events: [
        { label: 'Booking confirmed', location: 'Aba', at: 'Nov 16, 9:12 AM', state: 'done' },
        { label: 'Picked up from sender', location: 'Aba', at: 'Nov 16, 2:40 PM', state: 'done' },
        { label: 'Departed origin hub', location: 'Aba', at: 'Nov 16, 7:05 PM', state: 'done' },
        { label: 'In transit', location: 'Benin City', at: 'Nov 17, 6:30 AM', state: 'current' },
        { label: 'Arrive at destination hub', location: 'Lagos', at: 'Expected Nov 18', state: 'pending' },
        { label: 'Out for delivery', location: 'Lagos', at: 'Expected Nov 19', state: 'pending' },
      ],
    },
    {
      id: 'shp_2479',
      reference: 'BOL-2479',
      provider: 'Northstar Cargo',
      service: 'Same-day',
      origin: 'Onitsha',
      destination: 'Abuja',
      status: 'out_for_delivery',
      currentLocation: 'Wuse delivery hub, Abuja',
      eta: 'Today, before 6:00 PM',
      weightKg: 5,
      priceKobo: 620000,
      events: [
        { label: 'Booking confirmed', location: 'Onitsha', at: 'Nov 16, 11:20 AM', state: 'done' },
        { label: 'Picked up from sender', location: 'Onitsha', at: 'Nov 16, 4:15 PM', state: 'done' },
        { label: 'In transit', location: 'Lokoja', at: 'Nov 17, 3:05 AM', state: 'done' },
        { label: 'Arrived at destination hub', location: 'Abuja', at: 'Nov 18, 7:40 AM', state: 'done' },
        { label: 'Out for delivery', location: 'Wuse, Abuja', at: 'Today, 9:10 AM', state: 'current' },
        { label: 'Delivered to recipient', location: 'Abuja', at: 'Expected today', state: 'pending' },
      ],
    },
    {
      id: 'shp_2464',
      reference: 'BOL-2464',
      provider: 'Swiftline Logistics',
      service: 'Standard',
      origin: 'Port Harcourt',
      destination: 'Enugu',
      status: 'delayed',
      currentLocation: 'Port Harcourt sorting centre',
      eta: 'Delayed — new ETA within 24 hours',
      weightKg: 2,
      priceKobo: 310000,
      events: [
        { label: 'Booking confirmed', location: 'Port Harcourt', at: 'Nov 15, 8:05 AM', state: 'done' },
        { label: 'Picked up from sender', location: 'Port Harcourt', at: 'Nov 15, 1:30 PM', state: 'done' },
        { label: 'Held at sorting centre', location: 'Port Harcourt', at: 'Nov 16, 10:15 AM', state: 'current' },
        { label: 'In transit', location: 'Aba', at: 'Awaiting update', state: 'pending' },
        { label: 'Delivered to recipient', location: 'Enugu', at: 'Awaiting update', state: 'pending' },
      ],
    },
    {
      id: 'shp_2451',
      reference: 'BOL-2451',
      provider: 'Metroline Express',
      service: 'Economy',
      origin: 'Lagos',
      destination: 'Ibadan',
      status: 'delivered',
      currentLocation: 'Delivered to recipient',
      eta: 'Delivered Nov 12',
      weightKg: 1.5,
      priceKobo: 240000,
      events: [
        { label: 'Booking confirmed', location: 'Lagos', at: 'Nov 10, 10:00 AM', state: 'done' },
        { label: 'Picked up from sender', location: 'Lagos', at: 'Nov 10, 3:20 PM', state: 'done' },
        { label: 'In transit', location: 'Ibadan', at: 'Nov 11, 8:45 PM', state: 'done' },
        { label: 'Delivered to recipient', location: 'Ibadan', at: 'Nov 12, 11:05 AM', state: 'done' },
      ],
    },
  ],
  bookings: [
    {
      id: 'bkg_1094',
      reference: 'BK-1094',
      provider: 'Swiftline Logistics',
      service: 'Standard',
      origin: 'Lekki',
      destination: 'Abuja',
      status: 'awaiting_payment',
      priceKobo: 420000,
      note: 'Quoted price is held until 6:00 PM today.',
      createdAt: 'Created today, 9:24 AM',
    },
    {
      id: 'bkg_1091',
      reference: 'BK-1091',
      provider: 'Northstar Cargo',
      service: 'Express',
      origin: 'Ikeja',
      destination: 'Port Harcourt',
      status: 'awaiting_provider',
      priceKobo: 575000,
      note: 'Provider is confirming pickup availability.',
      createdAt: 'Created yesterday, 4:10 PM',
    },
  ],
  spend: {
    thisMonthKobo: 1840000,
    lastMonthKobo: 1190000,
    lifetimeKobo: 8625000,
    deliveredCount: 27,
    monthly: [
      { month: 'Jun', amountKobo: 920000 },
      { month: 'Jul', amountKobo: 1240000 },
      { month: 'Aug', amountKobo: 780000 },
      { month: 'Sep', amountKobo: 1530000 },
      { month: 'Oct', amountKobo: 1190000 },
      { month: 'Nov', amountKobo: 1840000 },
    ],
  },
}

const naira = new Intl.NumberFormat('en-NG', {
  style: 'currency',
  currency: 'NGN',
  maximumFractionDigits: 0,
})

export function formatNaira(kobo: number) {
  return naira.format(kobo / 100)
}

export const shipmentStatusMeta: Record<
  ShipmentStatus,
  { label: string; tone: string; dot: string }
> = {
  confirmed: {
    label: 'Confirmed',
    tone: 'border-slate-200 bg-slate-50 text-slate-600',
    dot: 'bg-slate-400',
  },
  picked_up: {
    label: 'Picked up',
    tone: 'border-sky-200 bg-sky-50 text-sky-700',
    dot: 'bg-sky-500',
  },
  in_transit: {
    label: 'In transit',
    tone: 'border-brand-200 bg-brand-50 text-brand-700',
    dot: 'bg-brand-500',
  },
  out_for_delivery: {
    label: 'Out for delivery',
    tone: 'border-amber-200 bg-amber-50 text-amber-700',
    dot: 'bg-amber-500',
  },
  delivered: {
    label: 'Delivered',
    tone: 'border-emerald-200 bg-emerald-50 text-emerald-700',
    dot: 'bg-emerald-500',
  },
  delayed: {
    label: 'Delayed',
    tone: 'border-red-200 bg-red-50 text-red-700',
    dot: 'bg-red-500',
  },
}

export function useCustomerDashboard() {
  const data = sampleDashboard

  const activeShipments = computed(() =>
    data.shipments.filter((shipment) => shipment.status !== 'delivered'),
  )

  const inTransitCount = computed(
    () =>
      data.shipments.filter(
        (shipment) => shipment.status === 'in_transit' || shipment.status === 'out_for_delivery',
      ).length,
  )

  const pendingBookings = computed(() => data.bookings)

  const maxMonthlySpend = computed(() =>
    Math.max(...data.spend.monthly.map((point) => point.amountKobo), 1),
  )

  return {
    data,
    activeShipments,
    inTransitCount,
    pendingBookings,
    maxMonthlySpend,
    formatNaira,
    shipmentStatusMeta,
    /** Dashboard data is illustrative until customer booking/tracking endpoints exist. */
    isSample: true,
  }
}
