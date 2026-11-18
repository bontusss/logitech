const THEME_STORAGE_KEY = 'bol-theme'

export function useTheme() {
  const isDark = useState<boolean>('theme.isDark', () => false)

  onMounted(() => {
    let storedTheme: string | null = null
    try {
      storedTheme = localStorage.getItem(THEME_STORAGE_KEY)
    } catch {
      storedTheme = null
    }
    isDark.value =
      storedTheme === 'dark' ||
      (storedTheme !== 'light' && window.matchMedia('(prefers-color-scheme: dark)').matches)
    applyTheme()
  })

  function applyTheme() {
    document.documentElement.classList.toggle('dark', isDark.value)
    document.querySelector<HTMLMetaElement>('#theme-color')?.setAttribute(
      'content',
      isDark.value ? '#0b1120' : '#fbfbf9',
    )
  }

  function toggleTheme() {
    isDark.value = !isDark.value
    applyTheme()
    try {
      localStorage.setItem(THEME_STORAGE_KEY, isDark.value ? 'dark' : 'light')
    } catch {
      // Keep the selected theme for this page even if storage is unavailable.
    }
  }

  return { isDark, toggleTheme }
}
