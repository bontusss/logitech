/**
 * Thin wrapper around $fetch for the Go API.
 * Turns non-2xx responses into an ApiError carrying the backend's message.
 */
export interface ApiErrorBody {
  error?: string
}

export class ApiError extends Error {
  readonly status: number
  readonly body: ApiErrorBody | undefined

  constructor(status: number, message: string, body?: ApiErrorBody) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.body = body
  }

  /** The backend accepts the one-time password but needs a new PIN in the same call. */
  get needsNewPin(): boolean {
    return this.status === 400 && /new_pin/i.test(this.body?.error ?? '')
  }

  /**
   * The account was locked after too many failed attempts. The attempt that
   * triggers the lock still answers with the generic credentials message, so
   * every 403 counts as locked except an inactive account.
   */
  get isLocked(): boolean {
    return this.status === 403 && !/inactive/i.test(this.body?.error ?? '')
  }
}

export async function apiRequest<T>(
  path: string,
  options: { method?: 'GET' | 'POST' | 'PUT' | 'DELETE'; body?: unknown; token?: string | null } = {},
): Promise<T> {
  const config = useRuntimeConfig()

  try {
    const response = await $fetch.raw<T>(path, {
      baseURL: config.public.apiBase,
      method: options.method ?? 'GET',
      body: options.body as Record<string, unknown> | undefined,
      headers: options.token ? { Authorization: `Bearer ${options.token}` } : undefined,
    })

    // A 2xx that is not JSON means the request never reached the API (for
    // example the app shell answering every path). Fail loudly instead of
    // handing the caller an HTML string that silently has no token in it.
    const contentType = response.headers.get('content-type') ?? ''
    if (response.status !== 204 && !contentType.includes('application/json')) {
      throw new ApiError(
        response.status,
        'The server did not answer with JSON. Check that the API is reachable.',
      )
    }

    return response._data as T
  } catch (error) {
    if (error instanceof ApiError) throw error
    const err = error as { status?: number; statusCode?: number; data?: ApiErrorBody }
    const status = err.status ?? err.statusCode ?? 0
    const message = err.data?.error ?? 'Something went wrong. Please try again.'
    throw new ApiError(status, message, err.data)
  }
}
