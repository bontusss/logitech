import type {
  Coordinates,
  MapPlace,
  MapProvider,
  MapRuntimeConfig,
  PlaceSearchOptions,
  RouteResult,
} from './types'

const GEOCODE_API = 'https://api.geoapify.com/v1/geocode'
const ROUTING_API = 'https://api.geoapify.com/v1/routing'
const STYLE_API = 'https://maps.geoapify.com/v1/styles'

interface GeoapifyResult {
  formatted?: string
  address_line1?: string
  address_line2?: string
  city?: string
  state?: string
  country_code?: string
  lat?: number
  lon?: number
}

interface GeoapifyResponse {
  results?: GeoapifyResult[]
}

interface GeoapifyRouteFeature {
  properties?: {
    distance?: number
    time?: number
  }
  geometry?: {
    type: string
    coordinates: unknown
  } | null
}

interface GeoapifyRouteResponse {
  features?: GeoapifyRouteFeature[]
}

function placeFromResult(result: GeoapifyResult | undefined): MapPlace | null {
  if (!result || typeof result.lat !== 'number' || typeof result.lon !== 'number') return null
  const address = result.formatted
    || [result.address_line1, result.address_line2].filter(Boolean).join(', ')
  if (!address) return null
  return {
    address,
    latitude: result.lat,
    longitude: result.lon,
    city: result.city ?? '',
    state: result.state ?? '',
    countryCode: result.country_code,
  }
}

export function createGeoapifyProvider(config: MapRuntimeConfig): MapProvider {
  const apiKey = config.apiKey.trim()
  const mapStyle = config.style.trim() || 'osm-bright'

  async function request<T>(url: URL, signal?: AbortSignal): Promise<T> {
    const response = await fetch(url.toString(), { signal })
    if (!response.ok) {
      let detail = ''
      try {
        const body = await response.json() as { message?: string }
        detail = body?.message ?? ''
      } catch {
        // Keep the generic message when the provider does not return JSON.
      }
      throw new Error(detail || 'Address lookup is unavailable right now.')
    }
    return await response.json() as T
  }

  function searchUrl(path: string, query: string, options: PlaceSearchOptions) {
    const url = new URL(`${GEOCODE_API}/${path}`)
    url.searchParams.set('text', query)
    url.searchParams.set('filter', `countrycode:${options.countryCode ?? 'ng'}`)
    url.searchParams.set('lang', options.language ?? 'en')
    url.searchParams.set('limit', String(options.limit ?? 5))
    url.searchParams.set('format', 'json')
    url.searchParams.set('apiKey', apiKey)
    return url
  }

  return {
    id: 'geoapify',

    styleUrl() {
      return `${STYLE_API}/${encodeURIComponent(mapStyle)}/style.json?apiKey=${encodeURIComponent(apiKey)}`
    },

    async autocomplete(query: string, options: PlaceSearchOptions = {}) {
      const url = searchUrl('autocomplete', query, options)
      const result = await request<GeoapifyResponse>(url, options.signal)
      return (result.results ?? [])
        .map((item) => placeFromResult(item))
        .filter((item): item is MapPlace => item !== null)
    },

    async geocode(query: string, options: PlaceSearchOptions = {}) {
      const url = searchUrl('search', query, { ...options, limit: 1 })
      const result = await request<GeoapifyResponse>(url, options.signal)
      return placeFromResult(result.results?.[0])
    },

    async reverse(coordinates: Coordinates, options: PlaceSearchOptions = {}) {
      const url = new URL(`${GEOCODE_API}/reverse`)
      url.searchParams.set('lat', String(coordinates.latitude))
      url.searchParams.set('lon', String(coordinates.longitude))
      url.searchParams.set('lang', options.language ?? 'en')
      url.searchParams.set('format', 'json')
      url.searchParams.set('apiKey', apiKey)
      const result = await request<GeoapifyResponse>(url, options.signal)
      return placeFromResult(result.results?.[0])
    },

    async route(from: Coordinates, to: Coordinates, options: { signal?: AbortSignal } = {}): Promise<RouteResult | null> {
      const url = new URL(ROUTING_API)
      url.searchParams.set(
        'waypoints',
        `${from.latitude},${from.longitude}|${to.latitude},${to.longitude}`,
      )
      url.searchParams.set('mode', 'drive')
      url.searchParams.set('apiKey', apiKey)
      const result = await request<GeoapifyRouteResponse>(url, options.signal)
      const feature = result.features?.[0]
      if (!feature) return null
      return {
        distanceMeters: feature.properties?.distance ?? 0,
        durationSeconds: feature.properties?.time ?? 0,
        geometry: feature.geometry ?? null,
      }
    },
  }
}
