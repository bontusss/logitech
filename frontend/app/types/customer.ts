export type ShipmentStatus =
  | 'confirmed'
  | 'picked_up'
  | 'in_transit'
  | 'out_for_delivery'
  | 'delivered'
  | 'delayed'

export type ShipmentEventState = 'done' | 'current' | 'pending'

export interface ShipmentEvent {
  label: string
  location: string
  at: string
  state: ShipmentEventState
}

export interface CustomerShipment {
  id: string
  reference: string
  provider: string
  service: string
  origin: string
  destination: string
  status: ShipmentStatus
  currentLocation: string
  eta: string
  weightKg: number
  priceKobo: number
  events: ShipmentEvent[]
}

export type BookingStatus = 'awaiting_payment' | 'awaiting_provider' | 'awaiting_pickup'

export interface CustomerBooking {
  id: string
  reference: string
  provider: string
  service: string
  origin: string
  destination: string
  status: BookingStatus
  priceKobo: number
  note: string
  createdAt: string
}

export interface SpendPoint {
  month: string
  amountKobo: number
}

export interface CustomerSpend {
  thisMonthKobo: number
  lastMonthKobo: number
  lifetimeKobo: number
  deliveredCount: number
  monthly: SpendPoint[]
}

export interface CustomerDashboardData {
  shipments: CustomerShipment[]
  bookings: CustomerBooking[]
  spend: CustomerSpend
}
