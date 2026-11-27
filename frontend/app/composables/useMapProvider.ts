import { createMapProvider, type MapProvider, type MapRuntimeConfig } from '~/utils/maps'

interface PublicConfigResponse {
  map?: {
    provider?: string
    api_key?: string
    style?: string
  }
}

export function useMapProvider() {
  const config = useRuntimeConfig()
  const runtime = useState<MapRuntimeConfig | null>('public-map-config', () => null)
  const loadError = useState('public-map-config-error', () => '')
  const loading = useState('public-map-config-loading', () => false)
  const apiBase = (config.public.apiBase as string) || '/api/v1'

  const provider = computed<MapProvider | null>(() => createMapProvider(runtime.value))
  const styleUrl = computed(() => provider.value?.styleUrl() ?? '')

  async function load(force = false) {
    if (runtime.value && !force) return
    if (loading.value) return
    loading.value = true
    loadError.value = ''
    try {
      const result = await $fetch<PublicConfigResponse>('/public/config', { baseURL: apiBase })
      const map = result.map
      runtime.value = map?.api_key
        ? {
            provider: map.provider || 'geoapify',
            apiKey: map.api_key,
            style: map.style || 'osm-bright',
          }
        : null
    } catch (cause) {
      loadError.value = cause instanceof Error
        ? cause.message
        : 'Could not load map configuration.'
    } finally {
      loading.value = false
    }
  }

  return { config: runtime, provider, styleUrl, loadError, loading, load }
}
