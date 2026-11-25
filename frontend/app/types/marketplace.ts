export interface MarketplacePackage {
  weight_kg: number
  length_cm?: number
  width_cm?: number
  height_cm?: number
  category: string
}

export interface MarketplaceSearchPayload {
  origin_zone_id: string
  destination_zone_id: string
  origin_label?: string
  destination_label?: string
  pickup_address?: string
  dropoff_address?: string
  weight_kg: number
  length_cm?: number
  width_cm?: number
  height_cm?: number
  category?: string
  pickup_required?: boolean
  delivery_required?: boolean
  sort?: 'cheapest' | 'fastest' | 'recommended' | 'rating'
}

export interface MarketplaceQuote {
  quote_reference: string
  provider_id: string
  provider_name: string
  provider_logo: string
  provider_slug: string
  booking_slug?: string
  provider_rating: number
  service_key: 'zone_route'
  service_name: string
  origin_state: string
  origin_city: string
  destination_state: string
  destination_city: string
  base_price_minor: number
  chargeable_weight_kg: number
  additional_charges_minor: number
  discount_minor: number
  total_price_minor: number
  currency: string
  estimated_delivery?: string
  estimated_duration_days: number
  pickup_available: boolean
  delivery_available: boolean
  delivery_method: string
  price_confirmation_type: 'confirmed' | 'estimated'
  price_confirmation_note?: string
  status: 'active' | 'selected' | 'expired'
  expires_at: string
  conditions?: string
  breakdown: { name: string; description?: string; amount_minor: number }[]
  distance_km: number
  departure_schedule?: string
}

export interface MarketplaceFilters {
  service: string
  confirmation: string
  maxPriceNaira: string
  pickup: boolean
  doorDelivery: boolean
}

export interface MarketplaceSearchMeta {
  providers_considered: number
  providers_quoted: number
  result_count: number
  failures: { provider: string; code: string }[]
  partial: boolean
  sort: string
  expires_at: string
}

export interface MarketplaceSearchResponse {
  search_id: string
  quotes: MarketplaceQuote[]
  meta: MarketplaceSearchMeta
}

export interface MarketplaceServiceSummary {
  key: string
  name: string
  route_count: number
  from_minimum_charge_minor: number
  estimated_delivery: string
  price_is_estimate: boolean
}

export interface MarketplaceLane {
  id: string
  origin_state: string
  origin_city: string
  destination_state: string
  destination_city: string
  distance_km: number
  minimum_charge_minor: number
  transit?: string
  departure_schedule?: string
  pickup_available: boolean
  delivery_available: boolean
  price_is_estimate: boolean
  speeds: { key: string; price_per_km_minor: number }[]
}

export interface MarketplaceProviderProfile {
  id: string
  name: string
  slug: string
  logo_url: string
  cover_image_url: string
  primary_color: string
  secondary_color: string
  description: string
  state?: string
  city?: string
  country: string
  email: string
  phone: string
  subdomain?: string
  verified: boolean
  verified_at?: string
  published_at?: string
  currency: string
}

export interface MarketplaceProviderResponse {
  provider: MarketplaceProviderProfile
  services: MarketplaceServiceSummary[]
  coverage: { state: string; city: string }[]
  lanes: MarketplaceLane[]
}

export interface SavedQuote {
  saved_quote_id: string
  selection_status: string
  selected_at: string
  expired: boolean
  quote: MarketplaceQuote
}
