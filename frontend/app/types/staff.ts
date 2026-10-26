export interface StaffItem {
  id: string
  staff_id: string
  full_name: string
  phone?: string
  email?: string
  role: string
  job_role?: string
  branch_id: string
  status: string
  first_login_done: boolean
  locked: boolean
  failed_login_attempts: number
  locked_at?: string
  last_sign_in?: string
  inactive_since?: string
  manager_name?: string | null
  branch_name?: string
  subtext?: string
  badge?: StatusBadge
  manager?: string | null
}

export interface StatusBadge {
  label: string
  tone: 'danger' | 'info' | 'neutral'
}

export interface BranchItem {
  id: string
  name: string
}

export interface ResetOtpResponse {
  staff_id: string
  one_time_password: string
}

export interface ConfirmModal {
  mode: 'off' | 'on' | 'remove'
  staffId: string
  staffName: string
  currentStatus: string
}

export interface OtpModal {
  staffId: string
  oneTimePassword: string
  name: string
}

export interface BranchDetails {
  id: string
  name: string
  code: string
  address?: string
  state?: string
  phone?: string
  status: string
}

export interface FleetVehicle {
  id: string
  branch_id: string
  registration_number: string
  vehicle_type: string
  make_model?: string
  capacity_kg?: string
  status: string
}

export interface BranchOptions {
  value: string
  label: string
}
