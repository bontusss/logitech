import type { Membership } from '~/types/auth'

export function useCompany() {
  // The backend re-validates this id against the user's memberships on every
  // request; the cookie is a selection, never an authorization claim.
  const currentCompanyId = useCookie<string | null>('bol_company_id', {
    default: () => null,
    maxAge: 60 * 60 * 24 * 30,
    sameSite: 'lax',
  })

  const companies = useState<Membership[]>('auth.companies', () => [])

  const currentCompany = computed(
    () => companies.value.find((company) => company.company_id === currentCompanyId.value) ?? null,
  )

  function selectCompany(companyId: string | null) {
    currentCompanyId.value = companyId
  }

  function syncCompanies(list: Membership[]) {
    companies.value = list
    if (!list.some((company) => company.company_id === currentCompanyId.value)) {
      currentCompanyId.value = list[0]?.company_id ?? null
    }
  }

  return {
    companies,
    currentCompanyId,
    currentCompany,
    selectCompany,
    syncCompanies,
  }
}
