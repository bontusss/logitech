import { defineNuxtRouteMiddleware } from '#app'

/**
 * The "Welcome" step only makes sense straight after the API asked for a new
 * PIN, so it bounces back to sign in when it is opened directly.
 */
export default defineNuxtRouteMiddleware(() => {
  if (!useFirstLogin().value) {
    return navigateTo('/sign-in')
  }
})
