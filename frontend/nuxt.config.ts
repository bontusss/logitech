import tailwindcss from '@tailwindcss/vite'

// Public origin of the deployed app. Share previews need absolute URLs, and the
// generated shell is what a crawler reads before any JavaScript runs, so the
// Open Graph / Twitter tags are declared here rather than only on the page.
const SITE_URL = process.env.NUXT_PUBLIC_SITE_URL ?? 'https://logitech-v3un.onrender.com'

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  ssr: false,
  devtools: { enabled: false },
  modules: ['@vite-pwa/nuxt'],
  // Post build outputs to .output/public for the Go embed (make build).
  nitro: { output: { publicDir: '.output/public' } },
  css: ['@fontsource-variable/urbanist/index.css', '~/assets/css/main.css'],
  vite: {
    plugins: [tailwindcss()],
  },
  app: {
    head: {
      title: 'Galopin',
      htmlAttrs: { lang: 'en' },
      // The manifest link is set per theme in app.vue (light and dark
      // manifests carry different launch colours).
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' },
      ],
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'theme-color', content: '#f5f5f2' },
        { name: 'apple-mobile-web-app-title', content: 'Galopin' },

        // Link previews (Open Graph + Twitter). The copy matches the landing
        // hero, because this is what people see when the link is shared.
        { property: 'og:type', content: 'website' },
        { property: 'og:site_name', content: 'based on logistics (BOL)' },
        { property: 'og:url', content: `${SITE_URL}/` },
        { property: 'og:title', content: 'Every parcel. Every movement. Accounted for.' },
        {
          property: 'og:description',
          content:
            'The logistics operating system for courier, haulage and consolidation businesses — waybills, bags and containers, custody between branches, and the money that comes with all of it.',
        },
        { property: 'og:image', content: `${SITE_URL}/og-image.png` },
        { property: 'og:image:type', content: 'image/png' },
        { property: 'og:image:width', content: '1200' },
        { property: 'og:image:height', content: '630' },
        {
          property: 'og:image:alt',
          content: 'based on logistics (BOL) — every parcel, every movement, accounted for.',
        },
        { name: 'twitter:card', content: 'summary_large_image' },
        { name: 'twitter:title', content: 'Every parcel. Every movement. Accounted for.' },
        {
          name: 'twitter:description',
          content: 'The logistics operating system for courier, haulage and consolidation businesses.',
        },
        { name: 'twitter:image', content: `${SITE_URL}/og-image.png` },
      ],
      // Runs while the page is still parsing, before anything is painted, so a
      // dark device never flashes the light canvas. Mirrors useTheme():
      // 'dark' wins, 'light' blocks the OS preference, anything else follows it.
      script: [
        {
          innerHTML: `(function(){try{var s=localStorage.getItem('logitech:theme');var d=s==='dark'||(s!=='light'&&matchMedia('(prefers-color-scheme: dark)').matches);document.documentElement.classList.toggle('dark',d);var m=document.querySelector('meta[name="theme-color"]');if(m)m.setAttribute('content',d?'#121211':'#f5f5f2')}catch(e){}})();`,
          tagPosition: 'head',
        },
      ],
    },
  },
  pwa: {
    registerType: 'autoUpdate',
    includeAssets: ['favicon.svg', 'apple-touch-icon.png'],
    // Light launch colours. The dark twin lives in public/manifest.dark.webmanifest
    // and app.vue links whichever matches the theme; keep the two in sync.
    manifest: {
      id: '/',
      name: 'Galopin',
      short_name: 'Galopin',
      description: 'Record waybills, receive parcels and track collections across branches.',
      start_url: '/',
      scope: '/',
      display: 'standalone',
      background_color: '#f5f5f2',
      theme_color: '#f5f5f2',
      lang: 'en',
      categories: ['business', 'productivity'],
      icons: [
        { src: '/pwa-192x192.png', sizes: '192x192', type: 'image/png' },
        { src: '/pwa-512x512.png', sizes: '512x512', type: 'image/png' },
        { src: '/maskable-512x512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
      ],
    },
    workbox: {
      // Precache the generated app shell so installed launches work offline.
      globPatterns: ['**/*.{js,css,html,svg,png,ico,woff2}'],
      navigateFallback: '/index.html',
      navigateFallbackDenylist: [/^\/api\//, /^\/health/, /^\/swagger/],
      cleanupOutdatedCaches: true,
      clientsClaim: true,
      skipWaiting: true,
    },
    // Service workers are build-only; dev keeps HMR clean.
    devOptions: { enabled: false },
  },
  runtimeConfig: {
    public: {
      // Base URL of the Go server; override with NUXT_PUBLIC_API_BASE.
      // In `nuxt dev` the app runs on :3000 while the API runs on :8080, so
      // requests go cross-origin (the Go server's CORS allows it). The built
      // app is served by the Go server itself, so it calls its own origin.
      apiBase:
        process.env.NUXT_PUBLIC_API_BASE ??
        (process.env.NODE_ENV === 'development' ? 'http://localhost:8080' : ''),
    },
  },
})
