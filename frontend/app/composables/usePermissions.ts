import type { Shipment } from '~/types/api'

/**
 * One place for the access rules the API enforces:
 * - the owner sees every branch, every record and every metric;
 * - a branch manager runs their own branch: sending, receiving, and the
 *   staff accounts of that branch;
 * - operations and dispatch staff record and send items from their branch;
 * - receiving staff confirm items destined for their branch.
 */
export function usePermissions() {
  const { user } = useAuth()

  const role = computed(() => user.value?.role ?? '')
  const jobRole = computed(() => (user.value?.job_role ?? '').toLowerCase())
  const isOwner = computed(() => role.value === 'owner')
  const isManager = computed(() => role.value === 'manager')
  const isStaff = computed(() => role.value === 'staff')
  /** Operations staff work the drop-off counter: they send parcels and never receive them. */
  const isOperations = computed(() => isStaff.value && jobRole.value === 'operations')
  const isSystemAdmin = computed(() => role.value === 'system_admin')
  const branchId = computed(() => user.value?.branch_id ?? '')

  const canViewAllBranches = computed(() => isOwner.value || isSystemAdmin.value)
  const canViewReports = computed(() => isOwner.value || isManager.value)
  const canManageStaff = computed(() => isOwner.value || isManager.value)
  const canManageBranches = computed(() => isOwner.value || isSystemAdmin.value)
  const canManageFleet = computed(() => isOwner.value || isSystemAdmin.value)

  /** Sending is the front-desk job: operations/dispatch staff and managers (owners do not send). */
  const canSend = computed(
    () =>
      !isOwner.value &&
      (isManager.value || (isStaff.value && jobRole.value !== 'receiving')),
  )

  /** Receiving is the destination job: receiving staff and managers (owners do not receive). */
  const canReceiveGoods = computed(
    () =>
      !isOwner.value &&
      (isManager.value ||
        (isStaff.value && jobRole.value !== 'operations' && jobRole.value !== 'dispatch')),
  )

  function touchesBranch(record: Shipment, branch: string): boolean {
    return Boolean(branch) && (record.origin_branch_id === branch || record.destination_branch_id === branch)
  }

  function canView(record: Shipment): boolean {
    if (canViewAllBranches.value) return true
    return touchesBranch(record, branchId.value)
  }

  /** Receiving happens at the destination branch, by a receiving-capable user. */
  function canReceive(record: Shipment): boolean {
    if (isOwner.value) return false
    if (!canReceiveGoods.value || !branchId.value) return false
    return record.destination_branch_id === branchId.value
  }

  /** Sending happens at the origin branch. */
  function canDispatch(record: Shipment): boolean {
    if (isOwner.value) return false
    if (!canSend.value || !branchId.value) return false
    return record.origin_branch_id === branchId.value
  }

  /** Editing a record before it leaves: its recorder, the branch manager or the owner. */
  function canEdit(record: Shipment): boolean {
    if (isOwner.value) return true
    if (isManager.value) return record.origin_branch_id === branchId.value
    return isStaff.value && record.recorded_by === user.value?.id && record.status === 'pending'
  }

  /** Cancelling is a manager/owner decision. */
  function canCancel(record: Shipment): boolean {
    if (isOwner.value) return record.status !== 'received' && record.status !== 'cancelled'
    return isManager.value && canDispatch(record) && record.status !== 'received' && record.status !== 'cancelled'
  }

  /** A record can be edited while it is still at the origin branch. */
  function isEditable(record: Shipment): boolean {
    return record.status === 'pending' && canEdit(record)
  }

  return {
    role,
    jobRole,
    branchId,
    isOwner,
    isManager,
    isStaff,
    isOperations,
    isSystemAdmin,
    canViewAllBranches,
    canViewReports,
    canManageStaff,
    canManageBranches,
    canManageFleet,
    canSend,
    canReceiveGoods,
    canView,
    canReceive,
    canDispatch,
    canEdit,
    canCancel,
    isEditable,
  }
}
