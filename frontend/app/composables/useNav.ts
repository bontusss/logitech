/** Navigation sections, filtered by what the signed-in role may reach. */
export interface NavItem {
  label: string
  to: string
  icon: string
}

export function useNav() {
  const { canViewReports, canManageStaff, canManageBranches, canManageFleet } = usePermissions()

  const sections = computed<Array<{ title: string; items: NavItem[] }>>(() => {
    const operations: NavItem[] = [
      { label: 'Dashboard', to: '/', icon: 'home' },
      { label: 'Records', to: '/records', icon: 'package' },
      { label: 'Exceptions', to: '/exceptions', icon: 'alert' },
    ]
    if (canViewReports.value) {
      operations.push({ label: 'Reports', to: '/reports', icon: 'chart' })
    }

    const manage: NavItem[] = []
    if (canManageStaff.value) manage.push({ label: 'Staff', to: '/staff', icon: 'users' })
    if (canManageBranches.value) manage.push({ label: 'Branches', to: '/branches', icon: 'building' })
    if (canManageFleet.value) {
      manage.push({ label: 'Drivers', to: '/drivers', icon: 'user' })
      manage.push({ label: 'Vehicles', to: '/vehicles', icon: 'truck' })
    }

    return [
      { title: 'Operations', items: operations },
      ...(manage.length ? [{ title: 'Manage', items: manage }] : []),
      { title: 'Account', items: [{ label: 'Settings', to: '/settings', icon: 'settings' }] },
    ]
  })

  const mobileItems = computed<NavItem[]>(() => {
    const items: NavItem[] = [
      { label: 'Home', to: '/', icon: 'home' },
      { label: 'Records', to: '/records', icon: 'package' },
      { label: 'Exceptions', to: '/exceptions', icon: 'alert' },
    ]
    items.push(
      canViewReports.value
        ? { label: 'Reports', to: '/reports', icon: 'chart' }
        : { label: 'Settings', to: '/settings', icon: 'settings' },
    )
    return items
  })

  return { sections, mobileItems }
}
