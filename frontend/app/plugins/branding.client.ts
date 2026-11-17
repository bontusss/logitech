import type { PublicBranding } from '~/types/auth'

export default defineNuxtPlugin(async () => {
  const config = useRuntimeConfig()
  const domain = ((config.public.appDomain as string) || 'localhost').toLowerCase()
  const host = window.location.hostname.toLowerCase()

  let subdomain = ''
  if (host.endsWith(`.${domain}`)) {
    const label = host.slice(0, -(domain.length + 1))
    if (label && !label.includes('.') && label !== 'www') subdomain = label
  }
  if (!subdomain) return

  const baseURL = (config.public.apiBase as string) || '/api/v1'
  try {
    const data = await $fetch<PublicBranding>('/public/branding', {
      baseURL,
      query: { subdomain },
      credentials: 'omit',
    })
    const company = data.company
    const branding = data.branding
    if (!company || !branding) return

    const primary = branding.primary_color || '#2b4bff'
    const secondary = branding.secondary_color || '#0a1220'
    const root = document.documentElement
    root.style.setProperty('--color-brand-500', primary)
    root.style.setProperty('--color-brand-600', `color-mix(in srgb, ${primary} 85%, black)`)
    root.style.setProperty('--color-brand-700', `color-mix(in srgb, ${primary} 72%, black)`)
    root.style.setProperty('--color-brand-50', `color-mix(in srgb, ${primary} 8%, white)`)
    root.style.setProperty('--color-brand-100', `color-mix(in srgb, ${primary} 16%, white)`)
    root.style.setProperty('--color-brand-200', `color-mix(in srgb, ${primary} 32%, white)`)
    root.style.setProperty('--color-ink-900', secondary)

    const name = company.public_name || company.name
    document.title = `${name} · BOL`

    const theme = document.querySelector('meta[name="theme-color"]')
    if (theme) theme.setAttribute('content', primary)
    const appTitle = document.querySelector('meta[name="apple-mobile-web-app-title"]')
    if (appTitle) appTitle.setAttribute('content', name)

    let manifest = document.querySelector('link[rel="manifest"]') as HTMLLinkElement | null
    if (!manifest) {
      manifest = document.createElement('link')
      manifest.rel = 'manifest'
      document.head.appendChild(manifest)
    }
    const separator = baseURL.includes('?') ? '&' : '?'
    const manifestBase = baseURL.endsWith('/') ? baseURL.slice(0, -1) : baseURL
    manifest.href = `${manifestBase}/public/manifest${separator}subdomain=${encodeURIComponent(subdomain)}`

    if (branding.favicon_url) {
      let icon = document.querySelector('link[rel="icon"]') as HTMLLinkElement | null
      if (!icon) {
        icon = document.createElement('link')
        icon.rel = 'icon'
        document.head.appendChild(icon)
      }
      icon.href = branding.favicon_url
    }
  } catch {
    // Branding is progressive enhancement; the platform theme remains usable.
  }
})
