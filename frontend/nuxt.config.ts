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
      htmlAttrs: { lang: 'en' },
      title: 'based on logistics (BOL)',
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'theme-color', content: '#fbfbf9' },
        {
          name: 'description',
          content:
            'Discover logistics companies, compare shipping rates and book shipments across Nigeria — or run your logistics business on one connected operations platform.',
        },
        { property: 'og:type', content: 'website' },
        { property: 'og:site_name', content: 'based on logistics (BOL)' },
        { property: 'og:url', content: `${SITE_URL}/` },
        { property: 'og:title', content: 'One platform. Every shipment. More possibilities.' },
        {
          property: 'og:description',
          content:
            'A logistics marketplace and business management platform for African intercity logistics.',
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
    },
  },
})
