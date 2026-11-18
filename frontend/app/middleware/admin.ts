export default defineNuxtRouteMiddleware(async (to) => {
  // Auth is cookie-based and client-only; prerendering must not attempt it.
  if (import.meta.server) return

  const { user, loaded, fetchMe } = useAuth()

  if (!loaded.value) {
    await fetchMe()
  }

  if (!user.value) {
    return navigateTo({ path: '/admin/login', query: { next: to.fullPath } })
  }

  if (!user.value.platform_role) {
    return navigateTo('/dashboard')
  }
})
