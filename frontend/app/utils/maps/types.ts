// Provider-neutral map contracts. UI components depend only on this module so a
// new provider (for example TomTom) can be added as another adapter without
// touching booking, address or fleet screens.
export type MapProviderId = 'geoapify'

export interface MapRuntimeConfig {
  provider: string
  apiKey: string
  style: string
}

export interface Coordinates {
  latitude: number
  longitude: number
}

export interface MapPlace extends Coordinates {
  address: string
  city: string
  state: string
  countryCode?: string
}

export interface PlaceSearchOptions {
  limit?: number
  countryCode?: string
  language?: string
  signal?: AbortSignal
}

export interface RouteResult {
  distanceMeters: number
  durationSeconds: number
  geometry: {
    type: string
    coordinates: unknown
  } | null
}

export interface MapProvider {
  readonly id: MapProviderId
  /** MapLibre style URL for the provider's vector/raster base map. */
  styleUrl(): string
  autocomplete(query: string, options?: PlaceSearchOptions): Promise<MapPlace[]>
  geocode(query: string, options?: PlaceSearchOptions): Promise<MapPlace | null>
  reverse(coordinates: Coordinates, options?: PlaceSearchOptions): Promise<MapPlace | null>
  route(from: Coordinates, to: Coordinates, options?: { signal?: AbortSignal }): Promise<RouteResult | null>
}
