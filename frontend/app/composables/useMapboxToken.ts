export function useMapboxToken() {
  const config = useRuntimeConfig()
  const token = useState('public-mapbox-token', () => '')
  const loadError = useState('public-mapbox-token-error', () => '')
  const apiBase = (config.public.apiBase as string) || '/api/v1'

  async function load() {
    loadError.value = ''
    try {
      const result = await $fetch<{ mapbox_access_token: string }>('/public/config', {
        baseURL: apiBase,
      })
      token.value = result.mapbox_access_token || ''
    } catch (cause) {
      loadError.value = cause instanceof Error
        ? cause.message
        : 'Could not load location search configuration.'
    }
  }

  return { token, loadError, load }
}
