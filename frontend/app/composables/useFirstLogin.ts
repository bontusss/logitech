export interface PendingFirstLogin {
  staffId: string
  oneTimePassword: string
}

/**
 * The API only reveals that a PIN must be created in the response to the
 * login call, so the accepted staff_id + one-time password are carried over
 * to the "Welcome" step in memory.
 */
export function useFirstLogin() {
  return useState<PendingFirstLogin | null>('auth:first-login', () => null)
}
