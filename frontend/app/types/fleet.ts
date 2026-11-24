export type DeliveryRoadStatus = 'assigned' | 'picked_up' | 'in_transit'

export interface LiveDriverLocation {
  latitude: number
  longitude: number
  updated_at: string
}

export interface LiveDriver {
  driver_id: string
  full_name: string
  vehicle_plate: string
  delivery_id: string
  delivery_status: DeliveryRoadStatus | string
  tracking_code: string | null
  pickup_address: string
  dropoff_address: string
  location: LiveDriverLocation | null
}
