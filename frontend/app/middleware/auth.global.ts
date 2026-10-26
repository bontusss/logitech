import { ApiError } from '~/utils/api'

// Reachable without a session.
const PUBLIC_ROUTES = new Set(['/sign-in', '/sign-in/admin', '/first-sign-in', '/locked'])

export default defineNuxtRouteMiddleware(async (to) => {
  const { token, user, fetchMe } = useAuth()
  const isPublic = PUBLIC_ROUTES.has(to.path)

  if (!token.value) {
    if (isPublic) return
    return navigateTo(
      to.fullPath === '/' ? '/sign-in' : { path: '/sign-in', query: { redirect: to.fullPath } },
    )
  }

  // The token lives in a cookie, so the profile is re-fetched after a reload.
  if (!user.value) {
    try {
      await fetchMe()
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        token.value = null
        user.value = null
        return navigateTo('/sign-in')
      }
    }
  }

  if (isPublic) {
    return navigateTo('/')
  }
})
