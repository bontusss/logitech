import type { AuthUser, LoginResponse, Membership, SessionResponse } from '~/types/auth'

export interface CompanyRegistrationPayload {
  company_name: string
  business_email: string
  phone: string
  business_type: string
  country: string
  state: string
  city: string
  address: string
  contact_name: string
  password: string
}

export interface CustomerRegistrationPayload {
  full_name: string
  email: string
  phone?: string
  password: string
}

const DEV_VERIFY_KEY = 'bol_dev_verify_token'

export function useAuth() {
  const user = useState<AuthUser | null>('auth.user', () => null)
  const loaded = useState<boolean>('auth.loaded', () => false)
  const { api } = useApi()
  const { syncCompanies, companies } = useCompany()

  function applySession(session: LoginResponse) {
    user.value = session.user
    syncCompanies(session.companies)
    loaded.value = true
  }

  function rememberDevToken(token?: string) {
    if (!token || !import.meta.client) return
    try {
      localStorage.setItem(DEV_VERIFY_KEY, token)
    } catch {
      // Storage can be unavailable in private contexts; the flow still works
      // through the development log link.
    }
  }

  async function fetchMe(): Promise<AuthUser | null> {
    try {
      const data = await api<{ user: AuthUser; companies: Membership[] }>('/users/me')
      user.value = data.user
      syncCompanies(data.companies)
    } catch {
      user.value = null
      syncCompanies([])
    } finally {
      loaded.value = true
    }
    return user.value
  }

  async function login(email: string, password: string) {
    const session = await api<LoginResponse>('/auth/login', {
      method: 'POST',
      body: { email, password },
    })
    applySession(session)
    return session
  }

  async function registerCustomer(payload: CustomerRegistrationPayload) {
    const session = await api<SessionResponse>('/auth/register', {
      method: 'POST',
      body: payload,
    })
    applySession(session)
    rememberDevToken(session.dev_verification_token)
    return session
  }

  async function registerCompany(payload: CompanyRegistrationPayload) {
    const session = await api<SessionResponse>('/companies', {
      method: 'POST',
      body: payload,
    })
    applySession(session)
    rememberDevToken(session.dev_verification_token)
    return session
  }

  async function logout() {
    try {
      await api('/auth/logout', { method: 'POST', body: {} })
    } catch {
      // The local session is cleared even if the network call fails.
    } finally {
      user.value = null
      syncCompanies([])
      loaded.value = true
      const { selectCompany } = useCompany()
      selectCompany(null)
    }
  }

  return {
    user,
    companies,
    loaded,
    fetchMe,
    login,
    registerCustomer,
    registerCompany,
    logout,
  }
}
