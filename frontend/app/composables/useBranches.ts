import type { Branch } from '~/types/api'

/** Branch list, fetched once and shared across screens. */
export function useBranches() {
  const { token } = useAuth()
  const branches = useState<Branch[]>('branches', () => [])
  const loaded = useState<boolean>('branches:loaded', () => false)
  const loading = useState<boolean>('branches:loading', () => false)

  async function load(force = false): Promise<Branch[]> {
    if (!token.value) return branches.value
    if (loaded.value && !force) return branches.value
    if (loading.value) return branches.value
    loading.value = true
    try {
      const rows = await apiRequest<Branch[]>('/api/branches', { token: token.value })
      branches.value = rows
      loaded.value = true
    } catch {
      // Screens still render; selects just show no options.
    } finally {
      loading.value = false
    }
    return branches.value
  }

  const activeBranches = computed(() => branches.value.filter((b) => b.status === 'active'))

  const options = computed(() => activeBranches.value.map((b) => ({ value: b.id, label: `${b.name} (${b.code})` })))

  function name(branchId?: string): string {
    if (!branchId) return ''
    return branches.value.find((b) => b.id === branchId)?.name ?? ''
  }

  function code(branchId?: string): string {
    if (!branchId) return ''
    return branches.value.find((b) => b.id === branchId)?.code ?? ''
  }

  return { branches, activeBranches, options, loaded, loading, load, name, code }
}
