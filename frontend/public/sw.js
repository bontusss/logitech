/*
 * BOL service worker.
 *
 * Caching policy:
 *  - Hashed build assets, fonts and PWA icons are cached cache-first (safe and immutable).
 *  - Public marketing/auth navigations are cached network-first so they still open offline.
 *  - Dashboard and booking navigations are never cached (they can contain private data);
 *    when offline they fall back to a static offline page.
 *  - API requests are never intercepted.
 *
 * Bump SW_VERSION when the caching strategy changes.
 */
const SW_VERSION = 'bol-v1'
const STATIC_CACHE = `${SW_VERSION}-static`
const PAGE_CACHE = `${SW_VERSION}-pages`
const OFFLINE_URL = '/offline'

const PRECACHE_URLS = [
  OFFLINE_URL,
  '/site.webmanifest',
  '/favicon.svg',
  '/favicon-32x32.png',
  '/apple-touch-icon.png',
  '/icon-192.png',
  '/icon-512.png',
]

const CACHEABLE_PAGE_PATHS = new Set([
  '/',
  '/for-logistics-businesses',
  '/book-demo',
  '/login',
  '/register',
  '/forgot-password',
])

const STATIC_PREFIXES = ['/_nuxt/', '/fonts/']

self.addEventListener('install', (event) => {
  event.waitUntil(
    (async () => {
      const cache = await caches.open(STATIC_CACHE)

      for (const url of PRECACHE_URLS) {
        try {
          const response = await fetch(url, { cache: 'reload' })
          if (response.ok) await cache.put(url, response)
        } catch {
          // A missing optional asset must not block the install.
        }
      }

      // Precache the offline page together with the build assets it references,
      // so it still renders correctly with no network at all.
      try {
        const response = await fetch(OFFLINE_URL, { cache: 'reload' })
        if (response.ok) {
          const html = await response.clone().text()
          await cache.put(OFFLINE_URL, response)
          const assetUrls = new Set(
            Array.from(
              html.matchAll(/(?:href|src)="(\/(?:_nuxt|fonts)\/[^"]+)"/g),
              (match) => match[1],
            ),
          )
          await Promise.all(
            Array.from(assetUrls).map(async (url) => {
              try {
                const asset = await fetch(url)
                if (asset.ok) await cache.put(url, asset)
              } catch {
                // Ignore individual asset failures.
              }
            }),
          )
        }
      } catch {
        // The offline page is retried on the next visit if unavailable now.
      }

      await self.skipWaiting()
    })(),
  )
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    (async () => {
      const keys = await caches.keys()
      await Promise.all(
        keys.filter((key) => !key.startsWith(SW_VERSION)).map((key) => caches.delete(key)),
      )
      await self.clients.claim()
    })(),
  )
})

function isStaticAsset(url) {
  if (url.pathname.includes('_payload')) return false
  if (url.pathname.endsWith('.json') && !url.pathname.endsWith('site.webmanifest')) return false
  return (
    STATIC_PREFIXES.some((prefix) => url.pathname.startsWith(prefix)) ||
    PRECACHE_URLS.includes(url.pathname)
  )
}

async function cacheFirst(request) {
  const cache = await caches.open(STATIC_CACHE)
  const cached = await cache.match(request)
  if (cached) return cached

  const response = await fetch(request)
  if (response.ok && response.type === 'basic') {
    await cache.put(request, response.clone())
  }
  return response
}

async function navigationResponse(request) {
  const url = new URL(request.url)
  try {
    const response = await fetch(request)
    if (response.ok && CACHEABLE_PAGE_PATHS.has(url.pathname)) {
      const cache = await caches.open(PAGE_CACHE)
      await cache.put(request, response.clone())
    }
    return response
  } catch {
    const cachedPage = await caches.match(request)
    if (cachedPage) return cachedPage

    // Navigate to the cached offline page so the document URL matches the
    // route it renders. Returning offline HTML for e.g. /dashboard would make
    // the client router hydrate the wrong route.
    if (url.pathname !== OFFLINE_URL) {
      return Response.redirect(new URL(OFFLINE_URL, self.location.origin).href, 302)
    }

    return new Response('You are offline.', {
      status: 503,
      headers: { 'Content-Type': 'text/plain; charset=utf-8' },
    })
  }
}

self.addEventListener('fetch', (event) => {
  const { request } = event
  if (request.method !== 'GET') return

  const url = new URL(request.url)
  if (url.origin !== self.location.origin) return
  if (url.pathname === '/sw.js' || url.pathname.startsWith('/api/')) return

  if (request.mode === 'navigate') {
    event.respondWith(navigationResponse(request))
    return
  }

  if (isStaticAsset(url)) {
    event.respondWith(cacheFirst(request))
  }
})
