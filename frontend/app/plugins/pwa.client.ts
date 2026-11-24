export default defineNuxtPlugin(() => {
  if (!import.meta.env.PROD) return
  if (!('serviceWorker' in navigator)) return

  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js', { scope: '/' }).catch(() => {
      // Installation and offline support stay unavailable; the app still works online.
    })
  })
})
