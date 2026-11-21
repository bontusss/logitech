export type RequestStatus = 'pending' | 'accepted' | 'declined'

export type DeliveryStatus =
  | 'requested'
  | 'created'
  | 'assigned'
  | 'picked_up'
  | 'in_transit'
  | 'delivered'
  | 'cancelled'

export interface CustomerBooking {
  id: string
  status: DeliveryStatus
  request_status: RequestStatus
  tracking_code: string | null
  tracking_url: string | null
  pickup_address: string
  dropoff_address: string
  package_category: string
  item_description: string
  weight_kg: number
  delivery_price_minor: number
  declared_value_minor: number
  price_is_estimate: boolean
  delivery_speed: 'standard' | 'express' | 'same_day'
  distance_km: number
  provider_id: string
  provider_name: string
  provider_logo_url: string | null
  currency: string
  created_at: string
  updated_at: string
}

export interface CustomerAddress {
  address: string
  role: 'pickup' | 'dropoff'
  provider: string
  lastUsedAt: string
}
