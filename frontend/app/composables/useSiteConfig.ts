/**
 * Public destinations configured at runtime:
 * NUXT_PUBLIC_DEMO_URL and NUXT_PUBLIC_WHATSAPP_URL.
 */
export function useSiteConfig() {
  const config = useRuntimeConfig()

  const demoUrl = computed(() => (config.public.demoUrl as string) || '/book-demo')
  const whatsappUrl = computed(() => (config.public.whatsappUrl as string) || '')

  // Until a verified number is configured, the secondary CTA routes to the
  // contact block instead of inventing a phone number.
  const whatsappHref = computed(() => whatsappUrl.value || '#contact')

  return { demoUrl, whatsappUrl, whatsappHref }
}
