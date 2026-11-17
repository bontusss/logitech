/**
 * Public destinations configured at runtime:
 * NUXT_PUBLIC_DEMO_URL and NUXT_PUBLIC_WHATSAPP_URL.
 */
export function useSiteConfig() {
  const config = useRuntimeConfig()

  const demoUrl = computed(() => (config.public.demoUrl as string) || '/book-demo')
  const whatsappUrl = computed(() => (config.public.whatsappUrl as string) || '')

  const whatsappHref = computed(() => whatsappUrl.value)

  return { demoUrl, whatsappUrl, whatsappHref }
}
