import type { CompanyDetail } from '~/types/auth'

export interface ToastMessage {
  id: number
  type: 'success' | 'error' | 'info'
  message: string
}

export interface ConfirmOptions {
  title: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  danger?: boolean
}

export interface ConfirmState extends ConfirmOptions {
  open: boolean
}

let toastID = 0
let confirmResolver: ((value: boolean) => void) | null = null

export function useWorkspace() {
  const { api } = useApi()
  const { currentCompany } = useCompany()

  const detail = useState<CompanyDetail | null>('workspace.detail', () => null)
  const loading = useState<boolean>('workspace.loading', () => false)
  const error = useState<string>('workspace.error', () => '')
  const toasts = useState<ToastMessage[]>('workspace.toasts', () => [])
  const confirmState = useState<ConfirmState | null>('workspace.confirm', () => null)

  const permissions = computed<string[]>(() => detail.value?.membership.permissions ?? [])
  const can = (permission: string) =>
    Boolean(detail.value?.platform_access || permissions.value.includes(permission))

  async function load(force = false) {
    if (detail.value && !force) return
    if (!currentCompany.value) {
      detail.value = null
      return
    }
    loading.value = true
    error.value = ''
    try {
      detail.value = await api<CompanyDetail>('/companies/me')
    } catch (err) {
      error.value = apiErrorMessage(err, 'Could not load this company workspace.')
      detail.value = null
    } finally {
      loading.value = false
    }
  }

  async function refresh() {
    await load(true)
  }

  function notify(message: string, type: ToastMessage['type'] = 'success') {
    const id = ++toastID
    toasts.value = [...toasts.value, { id, type, message }]
    if (import.meta.client) {
      window.setTimeout(() => dismissToast(id), 5000)
    }
  }

  function dismissToast(id: number) {
    toasts.value = toasts.value.filter((toast) => toast.id !== id)
  }

  function confirm(options: ConfirmOptions): Promise<boolean> {
    confirmState.value = {
      open: true,
      title: options.title,
      message: options.message,
      confirmLabel: options.confirmLabel ?? 'Confirm',
      cancelLabel: options.cancelLabel ?? 'Cancel',
      danger: options.danger ?? false,
    }
    return new Promise<boolean>((resolve) => {
      confirmResolver = resolve
    })
  }

  function resolveConfirm(value: boolean) {
    confirmState.value = null
    confirmResolver?.(value)
    confirmResolver = null
  }

  return {
    detail,
    loading,
    error,
    toasts,
    confirmState,
    permissions,
    can,
    load,
    refresh,
    notify,
    dismissToast,
    confirm,
    resolveConfirm,
  }
}
