import tailwindcss from '@tailwindcss/vite'

// Public origin used for absolute social-share URLs.
const SITE_URL = process.env.NUXT_PUBLIC_SITE_URL ?? 'https://basedonlogistics.com'

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: false },
  css: ['@fontsource-variable/urbanist/index.css', '~/assets/css/main.css'],
  vite: {
    plugins: [tailwindcss()],
  },
  app: {
    head: {
      script: [
        {
          innerHTML:
            "try{const t=localStorage.getItem('bol-theme');if(t==='dark'||(t!=='light'&&matchMedia('(prefers-color-scheme: dark)').matches))document.documentElement.classList.add('dark')}catch{}",
        },
      ],
      htmlAttrs: { lang: 'en' },
      title: 'based on logistics (BOL)',
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'theme-color', content: '#fbfbf9', id: 'theme-color' },
        { name: 'mobile-web-app-capable', content: 'yes' },
        { name: 'apple-mobile-web-app-capable', content: 'yes' },
        { name: 'apple-mobile-web-app-status-bar-style', content: 'black-translucent' },
        { name: 'apple-mobile-web-app-title', content: 'BOL' },
        {
          name: 'description',
          content:
            'Find a better way to send packages across Nigeria. Compare intercity delivery options by price and speed with BOL.',
        },
        { property: 'og:type', content: 'website' },
        { property: 'og:site_name', content: 'based on logistics (BOL)' },
        { property: 'og:url', content: `${SITE_URL}/` },
        { property: 'og:title', content: 'Send it right. Know your options.' },
        {
          property: 'og:description',
          content: 'A simpler way to compare intercity delivery options across Nigeria.',
        },
        { name: 'twitter:card', content: 'summary_large_image' },
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon-32x32.png' },
        { rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-touch-icon.png' },
        { rel: 'manifest', href: '/site.webmanifest' },
      ],
    },
  },
  runtimeConfig: {
    public: {
      // Booking page or contact form for logistics businesses.
      demoUrl: process.env.NUXT_PUBLIC_DEMO_URL ?? '/book-demo',
      // Full click-to-chat URL (https://wa.me/<number>) once a number is confirmed.
      whatsappUrl: process.env.NUXT_PUBLIC_WHATSAPP_URL ?? '',
      // Gin API base. Use /api/v1 when the app and API share an origin.
      apiBase: process.env.NUXT_PUBLIC_API_BASE ?? '/api/v1',
      // Base domain used to resolve company branding from the left-most host label.
      appDomain: process.env.NUXT_PUBLIC_APP_DOMAIN ?? 'localhost',
    },
  },
  routeRules: {
    '/sign-in': { redirect: '/login' },
  },
})
