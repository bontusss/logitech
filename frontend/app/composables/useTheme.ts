export type ThemeMode = 'light' | 'dark' | 'system'

const STORAGE_KEY = 'logitech:theme'
const THEME_COLOR = { light: '#f5f5f2', dark: '#121211' } as const

/** The theme saved on this device, or null when nothing usable is stored. */
function storedMode(): ThemeMode | null {
  if (!import.meta.client) return null
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY)
    if (stored === 'light' || stored === 'dark' || stored === 'system') return stored
  } catch {
    // Blocked storage falls back to the OS preference.
  }
  return null
}

function systemPrefersDark(): boolean {
  return import.meta.client && window.matchMedia('(prefers-color-scheme: dark)').matches
}

/** Shared light/dark preference, persisted on this device. */
export function useTheme() {
  // Seeded synchronously from storage and the OS so the first render agrees
  // with the class the boot script in nuxt.config already put on <html>;
  // mounting must not flip a dark page back to light for a frame.
  const mode = useState<ThemeMode>('theme:mode', () => storedMode() ?? 'system')
  const systemDark = useState<boolean>('theme:system-dark', systemPrefersDark)
  const isDark = computed(() => mode.value === 'dark' || (mode.value === 'system' && systemDark.value))
  const themeColor = computed(() => (isDark.value ? THEME_COLOR.dark : THEME_COLOR.light))
  const label = computed(() => (isDark.value ? 'Switch to light mode' : 'Switch to dark mode'))

  function toggle() {
    const next: ThemeMode = isDark.value ? 'light' : 'dark'
    mode.value = next
    if (import.meta.client) {
      try {
        window.localStorage.setItem(STORAGE_KEY, next)
      } catch {
        // A blocked storage quota still lets the theme change for this session.
      }
    }
  }

  function init() {
    if (!import.meta.client) return
    const stored = storedMode()
    if (stored) mode.value = stored
    systemDark.value = systemPrefersDark()
    const media = window.matchMedia('(prefers-color-scheme: dark)')
    if (!listening) {
      listening = true
      media.addEventListener('change', (event) => {
        systemDark.value = event.matches
      })
    }
  }

  return { mode, isDark, themeColor, label, toggle, init }
}

// The matchMedia listener is shared: initialise the composable from any screen.
let listening = false
