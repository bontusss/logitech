interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>
}

const DISMISS_KEY = 'bol-install-banner-dismissed'

// Module-level state keeps the banner and the top-bar action in sync, and
// starts listening for the install prompt as soon as the layout loads.
const deferredPrompt = shallowRef<BeforeInstallPromptEvent | null>(null)
const installed = ref(false)
const dismissed = ref(false)
const isMobile = ref(false)
const ready = ref(false)

const bannerVisible = computed(
  () => ready.value && isMobile.value && !installed.value && !dismissed.value,
)
const canInstall = computed(() => ready.value && isMobile.value && !installed.value)

let listenersAttached = false
let initialized = false

function readDismissed() {
  try {
    return localStorage.getItem(DISMISS_KEY) === '1'
  } catch {
    return false
  }
}

function onBeforeInstall(event: Event) {
  event.preventDefault()
  deferredPrompt.value = event as BeforeInstallPromptEvent
}

function onInstalled() {
  installed.value = true
  deferredPrompt.value = null
}

function attachListeners() {
  if (listenersAttached || !import.meta.client) return
  listenersAttached = true
  window.addEventListener('beforeinstallprompt', onBeforeInstall)
  window.addEventListener('appinstalled', onInstalled)
}

attachListeners()

function initialize() {
  if (initialized || !import.meta.client) return
  initialized = true

  installed.value =
    window.matchMedia('(display-mode: standalone)').matches ||
    (navigator as Navigator & { standalone?: boolean }).standalone === true
  dismissed.value = readDismissed()

  const mobileQuery = window.matchMedia('(max-width: 1023px)')
  isMobile.value = mobileQuery.matches
  mobileQuery.addEventListener('change', (event) => {
    isMobile.value = event.matches
  })

  ready.value = true
}

async function install(): Promise<'prompted' | 'unavailable'> {
  const prompt = deferredPrompt.value
  if (!prompt) return 'unavailable'

  await prompt.prompt()
  const choice = await prompt.userChoice
  deferredPrompt.value = null
  if (choice.outcome === 'accepted') {
    installed.value = true
  }
  return 'prompted'
}

function dismissBanner() {
  dismissed.value = true
  try {
    localStorage.setItem(DISMISS_KEY, '1')
  } catch {
    // Dismiss for this visit only when storage is unavailable.
  }
}

export function usePwaInstall() {
  return {
    bannerVisible,
    canInstall,
    install,
    dismissBanner,
    initialize,
  }
}
