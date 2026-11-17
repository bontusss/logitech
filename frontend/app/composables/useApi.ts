interface ApiErrorBody {
  error?: {
    code?: string
    message?: string
  }
}

export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = (config.public.apiBase as string) || '/api/v1'
  const { currentCompanyId } = useCompany()

  async function api<T>(path: string, options: Record<string, any> = {}): Promise<T> {
    const headers = new Headers((options.headers as HeadersInit | undefined) ?? {})

    const isFormData = typeof FormData !== 'undefined' && options.body instanceof FormData
    if (options.body && !isFormData && !headers.has('Content-Type')) {
      headers.set('Content-Type', 'application/json')
    }
    if (currentCompanyId.value) {
      headers.set('X-Company-Id', currentCompanyId.value)
    }

    return await $fetch<T>(path, {
      baseURL,
      credentials: 'include',
      ...options,
      headers,
    })
  }

  return { api, baseURL }
}

export function apiErrorMessage(error: unknown, fallback = 'Something went wrong. Please try again.') {
  const candidate = error as { data?: ApiErrorBody }
  return candidate?.data?.error?.message || fallback
}
