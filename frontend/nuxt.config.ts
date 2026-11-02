import tailwindcss from '@tailwindcss/vite'

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
      title: 'Logitech',
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
        { name: 'apple-mobile-web-app-title', content: 'Logitech' },
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
      name: 'Logitech',
      short_name: 'Logitech',
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
