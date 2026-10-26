import { ApiError, apiRequest } from '~/utils/api'

export interface AuthUser {
  id: string
  staff_id: string
  full_name: string
  role: string
  job_role?: string
  branch_id?: string
  phone?: string
  email?: string
  status: string
  first_login_done: boolean
  locked: boolean
  created_at?: string
}

export interface AuthResponse {
  token: string
  token_type: string
  expires_in: number
  user: AuthUser
}

export interface LoginPayload {
  staffId: string
  /** The personal PIN, or the one-time password on a first login / after a reset. */
  pin: string
  /** Required by the API until the user has chosen a PIN. */
  newPin?: string
}

export function useAuth() {
  // Readable by the browser on purpose: the API is called directly with a
  // Bearer token, and the cookie keeps the session across reloads.
  const token = useCookie<string | null>('logitech_token', {
    maxAge: 60 * 60 * 24,
    sameSite: 'lax',
    default: () => null,
  })
  const user = useState<AuthUser | null>('auth:user', () => null)
  const isAuthenticated = computed(() => Boolean(token.value))

  async function login(payload: LoginPayload): Promise<AuthUser> {
    const response = await apiRequest<AuthResponse>('/api/auth/login', {
      method: 'POST',
      body: {
        staff_id: payload.staffId,
        pin: payload.pin,
        ...(payload.newPin ? { new_pin: payload.newPin } : {}),
      },
    })

    token.value = response.token
    user.value = response.user
    return response.user
  }

  async function fetchMe(): Promise<AuthUser> {
    const me = await apiRequest<AuthUser>('/api/auth/me', { token: token.value })
    user.value = me
    return me
  }

  async function logout(): Promise<void> {
    try {
      if (token.value) {
        await apiRequest('/api/auth/logout', { method: 'POST', token: token.value })
      }
    } catch (error) {
      // The session is dropped locally even if revocation failed.
      if (!(error instanceof ApiError)) throw error
    } finally {
      token.value = null
      user.value = null
    }
  }

  return { token, user, isAuthenticated, login, fetchMe, logout }
}
