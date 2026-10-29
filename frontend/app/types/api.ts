/** Shapes returned by the Logitech API. */

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

export interface Branch {
  id: string
  name: string
  code: string
  address?: string
  state?: string
  phone?: string
  status: string
  created_at?: string
}

export interface Staff {
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
  created_at?: string
}

export interface CreateUserResponse {
  id: string
  branch_id: string
  full_name: string
  staff_id: string
  one_time_password: string
  role: string
  job_role?: string
  status: string
  first_login_done: boolean
  created_at: string
}

export interface Vehicle {
  id: string
  branch_id: string
  registration_number: string
  vehicle_type: string
  make_model?: string
  capacity_kg?: string
  status: string
  created_at?: string
}

export interface Driver {
  id: string
  branch_id: string
  full_name: string
  phone: string
  license_number?: string
  license_expiry?: string
  status: string
  created_at?: string
}

export interface Party {
  full_name: string
  phone?: string
}

export interface Payment {
  id: string
  amount: number
  currency: string
  method: string
  reference?: string
  recorded_by: string
  recorded_by_name: string
  paid_at: string
  notes?: string
}

export interface ShipmentException {
  id: string
  exception_number: string
  type: string
  severity: string
  title: string
  description?: string
  manifest_item_id: string
  tracking_number: string
  branch_id: string
  branch_name: string
  status: string
  reported_by: string
  reported_by_name: string
  assigned_to?: string
  assigned_to_name?: string
  resolution_notes?: string
  resolved_by?: string
  resolved_by_name?: string
  resolved_at?: string
  created_at: string
  updated_at: string
  updates?: ExceptionUpdate[]
}

export interface ExceptionUpdate {
  id: string
  comment: string
  previous_status?: string
  new_status?: string
  updated_by: string
  updated_by_name: string
  created_at: string
}

/** One shipment: the item a customer hands in, its money and its delivery. */
export interface Shipment {
  id: string
  tracking_number: string
  serial_number?: string
  description: string
  quantity: number
  sender: Party
  receiver: Party
  origin_branch_id: string
  origin_branch: string
  destination_branch_id: string
  destination_branch: string
  amount: number
  amount_paid: number
  balance: number
  currency: string
  payment_status: string
  status: string
  recorded_by: string
  recorded_by_name: string
  photo_url?: string
  notes?: string
  dispatched_at?: string
  received_by?: string
  received_by_name?: string
  received_at?: string
  received_quantity?: number
  condition?: string
  created_at: string
  updated_at: string
  payments?: Payment[]
  exceptions?: ShipmentException[]
}

/** The signed-in account's own waybill counts (dashboard). */
export interface RecordSummary {
  /** Waybills they recorded that have already been dispatched. */
  sent: number
  /** Waybills they recorded that are still waiting to be dispatched. */
  pending: number
}

export interface RecordItemPayload {
  tracking_number?: string
  serial_number?: string
  description: string
  quantity: number
  sender: Party
  receiver: Party
  destination_branch_id: string
  amount: number
  currency?: string
  amount_paid?: number
  payment_method?: string
  payment_reference?: string
  notes?: string
}

export interface CreateRecordPayload {
  origin_branch_id?: string
  destination_branch_id?: string
  items: RecordItemPayload[]
}

export interface ReportPeriod {
  from: string
  to: string
}

export interface ReportOverview {
  period: ReportPeriod
  waybills?: {
    total: number
    quantity: number
    received: number
    by_status: Record<string, number>
    by_payment_status: Record<string, number>
  }
  shipments: {
    total: number
    quantity: number
    received: number
    by_status: Record<string, number>
    by_payment_status: Record<string, number>
  }
  money: {
    billed: number
    collected: number
    outstanding: number
    collection_rate: number
    payments_count: number
    collected_today: number
  }
  payments_by_method: Array<{ method: string; count: number; amount: number }>
  exceptions: {
    open: number
    investigating: number
    resolved: number
    closed: number
    total: number
  }
}

export interface DailyReportRow {
  date: string
  shipments: number
  billed: number
  collected: number
  payments: number
}

export interface BranchReportRow {
  branch_id: string
  branch_name: string
  branch_code: string
  shipments_sent: number
  shipments_received: number
  billed: number
  collected: number
  outstanding: number
  open_exceptions: number
}
