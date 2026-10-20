export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: false },
  // Post build outputs to .output/public for the Go embed (make build).
  nitro: { output: { publicDir: '.output/public' } },
  runtimeConfig: {
    public: {
      // Base URL of the Go server; override with NUXT_PUBLIC_API_BASE.
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080',
    },
  },
})
