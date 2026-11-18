export interface AuthUser {
  id: string
  email: string
  full_name: string
  phone: string | null
  whatsapp_verified: boolean
  account_type: 'platform' | 'company' | 'customer'
  platform_role: 'super_admin' | 'admin' | 'support' | null
  email_verified: boolean
  created_at: string
  last_login_at: string | null
}

export interface Membership {
  company_id: string
  company_name: string
  slug: string
  status: CompanyStatus
  onboarding_status: OnboardingStatus
  member_id: string
  member_status: string
  roles: string[]
}

export interface SessionResponse {
  user: AuthUser
  companies: Membership[]
  access_token: string
  expires_in: number
  dev_verification_token?: string
}

export interface LoginResponse {
  user: AuthUser
  companies: Membership[]
  access_token: string
  expires_in: number
}

export interface OperatingHoursDay {
  open: string
  close: string
  closed?: boolean
}

export type OperatingHours = Record<string, OperatingHoursDay>

export interface Company {
  id: string
  name: string
  slug: string
  business_email: string
  phone: string
  business_type: string
  cac_business_name: string | null
  country: string
  state: string | null
  city: string | null
  address: string | null
  status: CompanyStatus
  onboarding_status: OnboardingStatus
  verified_at: string | null
  created_at: string
  description: string | null
  cover_image_url: string | null
  operating_hours: OperatingHours
}

export type CompanyStatus = 'pending' | 'under_review' | 'verified' | 'rejected' | 'suspended'
export type OnboardingStatus = 'incomplete' | 'pending_review' | 'complete'

export interface CompanySettings {
  company_id: string
  timezone: string
  currency: string
  weight_unit: 'kg' | 'lb'
  auto_accept_bookings: boolean
  notify_email: boolean
}

export interface CompanyBranding {
  company_id: string
  logo_url: string | null
  primary_color: string
  secondary_color: string
  favicon_url: string | null
  cover_image_url: string | null
  public_name: string | null
  description: string | null
  pwa_short_name: string | null
}

export interface CompanyDetail {
  company: Company
  membership: {
    id: string
    roles: string[]
    permissions: string[]
  }
  settings: CompanySettings
  branding: CompanyBranding
  subdomain: string
  platform_access: boolean
}

export interface Branch {
  id: string
  company_id: string
  name: string
  code: string
  address: string
  city: string
  state: string
  country: string
  phone: string | null
  email: string | null
  manager_member_id: string | null
  operating_hours: OperatingHours
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
}

export interface BranchInput {
  name: string
  code: string
  address: string
  city: string
  state: string
  country: string
  phone: string
  email: string
  manager_member_id: string | null
  operating_hours: OperatingHours
}

export interface BranchStaff {
  id: string
  user_id: string
  full_name: string
  email: string
  title: string | null
  status: string
  is_manager: boolean
  roles: string[]
}

export interface StaffMember {
  id: string
  user_id: string
  full_name: string
  email: string
  status: string
  title: string | null
  joined_at: string | null
  roles: string[]
  branch_ids: string[]
}

export interface CompanyInvitation {
  id: string
  email: string
  expires_at: string
  created_at: string
  role_key: string
  branch_ids: string[]
}

export interface Role {
  id: string
  company_id: string | null
  key: string
  name: string
  description: string | null
  is_system: boolean
  permissions: string[]
}

export interface PermissionInfo {
  key: string
  description: string | null
}

export interface RoleSavePayload {
  id?: string
  name: string
  description: string
  permissions: string[]
}

export interface PublicInvitation {
  invitation: {
    email: string
    expires_at: string
    role_key: string
    role_name: string
    branch_ids: string[]
    branches: { id: string; name: string; code: string }[]
    account_exists: boolean
    requires_sign_in: boolean
  }
  company: { id: string; name: string; slug: string }
}

export interface PublicBranding {
  company: {
    id: string
    name: string
    slug: string
    public_name: string
    business_type: string
    description: string | null
    city: string | null
    state: string | null
    country: string
    status: CompanyStatus
    verified_at: string | null
  } | null
  branding: {
    logo_url: string | null
    primary_color: string
    secondary_color: string
    favicon_url: string | null
    cover_image_url: string | null
    description: string | null
  } | null
  subdomain: string
  manifest: Record<string, unknown>
}
