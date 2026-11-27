import { createGeoapifyProvider } from './geoapify'
import type { MapProvider, MapRuntimeConfig } from './types'

export * from './types'

/**
 * Build the active map provider from runtime configuration. Adding TomTom (or
 * any other service) means adding an adapter with the same `MapProvider`
 * interface and one case here — components stay unchanged.
 */
export function createMapProvider(config: MapRuntimeConfig | null | undefined): MapProvider | null {
  if (!config?.apiKey?.trim()) return null
  switch (config.provider) {
    case 'geoapify':
      return createGeoapifyProvider(config)
    default:
      return null
  }
}
