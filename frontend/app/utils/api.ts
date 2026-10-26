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
    return await $fetch<T>(path, {
      baseURL: config.public.apiBase,
      method: options.method ?? 'GET',
      body: options.body as Record<string, unknown> | undefined,
      headers: options.token ? { Authorization: `Bearer ${options.token}` } : undefined,
    })
  } catch (error) {
    const err = error as { status?: number; statusCode?: number; data?: ApiErrorBody }
    const status = err.status ?? err.statusCode ?? 0
    const message = err.data?.error ?? 'Something went wrong. Please try again.'
    throw new ApiError(status, message, err.data)
  }
}
