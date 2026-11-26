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
  sender_name: string
  sender_phone: string
  receiver_name: string
  receiver_phone: string
  payer: 'sender' | 'receiver'
  is_paid: boolean
  payment_method: 'cash' | 'pos' | 'bank_transfer' | null
  additional_instructions: string
  package_photo_url: string | null
  booking_source: 'dashboard' | 'booking_page'
  pickup_latitude: number
  pickup_longitude: number
  dropoff_latitude: number
  dropoff_longitude: number
  provider_id: string
  provider_slug: string
  provider_phone: string
  provider_name: string
  provider_logo_url: string | null
  currency: string
  can_cancel: boolean
  created_at: string
  updated_at: string
}

/** @deprecated Use CustomerSavedAddress for address-book entries. */
export interface CustomerRecentAddress {
  address: string
  role: 'pickup' | 'dropoff'
  provider: string
  lastUsedAt: string
}

export type CustomerAddressType = 'residential' | 'business' | 'warehouse' | 'other'

export interface CustomerSavedAddress {
  id: string
  label: string
  address_type: CustomerAddressType
  contact_name: string
  phone: string
  state: string
  city: string
  area: string
  street_address: string
  directions: string
  latitude: number | null
  longitude: number | null
  default_pickup: boolean
  default_delivery: boolean
  created_at: string
  updated_at: string
}

export interface CustomerAddressInput {
  label: string
  address_type: CustomerAddressType
  contact_name: string
  phone: string
  state: string
  city: string
  area: string
  street_address: string
  directions: string
  latitude: number | null
  longitude: number | null
  default_pickup: boolean
  default_delivery: boolean
}

export interface CustomerFavouriteProvider {
  id: string
  name: string
  slug: string
  logo_url: string
  description: string
  rating: number
  route_count: number
  saved_at: string
}

export type CustomerIssueType = 'delayed' | 'missing' | 'damaged' | 'other'

export interface CustomerIssue {
  id: string
  delivery_id: string
  tracking_code: string | null
  pickup_address: string
  dropoff_address: string
  provider_name: string
  issue_type: CustomerIssueType
  description: string
  status: 'open' | 'resolved'
  created_at: string
}

export interface TrackingEvent {
  status: string
  note: string | null
  created_at: string
}

export interface TrackingResult {
  tracking_code: string
  status: DeliveryStatus
  request_status: RequestStatus
  pickup_address: string
  dropoff_address: string
  updated_at: string
  timeline: TrackingEvent[]
}
